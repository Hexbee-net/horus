package warden

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/spf13/afero"
	"github.com/yuin/gopher-lua"
	"golang.org/x/xerrors"

	wlua "github.com/hexbee-net/horus/pkg/warden/lua"
	"github.com/hexbee-net/horus/pkg/warden/terraform"
	tflua "github.com/hexbee-net/horus/pkg/warden/terraform/lua"
)

type Warden struct {
	options *Options
	lState  *wlua.LState
}

// New creates a new Warden instance.
func New(opts ...*Options) (*Warden, error) {
	opt := DefaultOptions()

	for _, o := range opts {
		if o == nil {
			continue
		}

		if err := mergo.Merge(opt, o, mergo.WithOverride); err != nil {
			return nil, xerrors.Errorf("failed to merge the configuration: %w", err)
		}
	}

	w := &Warden{
		options: opt,
		lState:  wlua.NewState(lua.Options{SkipOpenLibs: true}),
	}

	w.lState.OpenModules(opt.Libs)
	w.lState.PreloadModules(opt.Modules)

	if err := w.lState.PreloadUserModule(opt.UserModules); err != nil {
		return nil, err //nolint:wrapcheck // this error actually comes from one of our own packages.
	}

	fn, err := w.lState.LoadString(w.options.Script)
	if err != nil {
		return nil, xerrors.Errorf("invalid validation script: %w", err)
	}

	w.lState.Push(fn)

	return w, nil
}

// ValidatePlan checks the validity of the specified plan with the configured
// scripts.
func (w *Warden) ValidatePlan(file afero.File) ([]string, error) {
	planFile, err := terraform.LoadPlanFile(file)
	if err != nil {
		return nil, xerrors.Errorf("failed to load plan file: %w", err)
	}

	w.lState.PreloadModule("tf", tflua.GetLoader(planFile))

	if err := w.lState.PCall(0, lua.MultRet, nil); err != nil {
		return nil, xerrors.Errorf("failed to load and parse validation script: %w", err)
	}

	issues := checkResult(w.lState)
	if len(issues) > 0 {
		return issues, ErrValidationFailed
	}

	return nil, nil
}

func (w *Warden) Close() {
	if w.lState != nil {
		w.lState.Close()
	}
}

func checkResult(ls *wlua.LState) (issues []string) {
	ret := ls.Get(-1)

	if ret == lua.LNil || ret == lua.LTrue {
		return nil
	}

	if ret == lua.LFalse {
		return []string{"validation failed"}
	}

	// Check for single error
	if str, ok := ret.(lua.LString); ok {
		return []string{str.String()}
	}

	// Check for multiple errors
	if tbl, ok := ret.(*lua.LTable); ok {
		errs := make([]string, 0, tbl.Len())
		tbl.ForEach(func(_ lua.LValue, v lua.LValue) {
			errs = append(errs, v.String())
		})

		return errs
	}

	// The returned value was neither Nil, a boolean, a string or
	// an array of string.
	// Still, something was returned so assume the validation failed and
	// return whatever we got back.
	return []string{fmt.Sprintf("validation failed (%s)", ret.String())}
}
