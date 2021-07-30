package warden

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/spf13/afero"
	luaCrypto "github.com/vadv/gopher-lua-libs/crypto"
	luaHumanize "github.com/vadv/gopher-lua-libs/humanize"
	luaInspect "github.com/vadv/gopher-lua-libs/inspect"
	luaJson "github.com/vadv/gopher-lua-libs/json"
	luaRegexp "github.com/vadv/gopher-lua-libs/regexp"
	luaTac "github.com/vadv/gopher-lua-libs/tac"
	luaTime "github.com/vadv/gopher-lua-libs/time"
	luaYaml "github.com/vadv/gopher-lua-libs/yaml"
	"github.com/yuin/gopher-lua"
	"golang.org/x/xerrors"
	"layeh.com/gopher-luar"

	"github.com/hexbee-net/horus/pkg/terraform/plans/planfile"
)

type LuaLib struct {
	Name     string
	Function lua.LGFunction
}

type UserModule struct {
	Name   string
	Script string
}

type Options struct {
	StandardLibs     *[]LuaLib
	Modules          *[]LuaLib
	UserModules      *[]UserModule
	Script           string
	PlanVarName      string
	StateVarName     string
	PrevStateVarName string
	ConfigVarName    string
}

type Warden struct {
	options  *Options
	luaState *lua.LState
}

func defaultOptions() *Options {
	return &Options{
		StandardLibs: &[]LuaLib{
			{lua.TabLibName, lua.OpenTable},     // Table manipulation
			{lua.StringLibName, lua.OpenString}, // String manipulation
			{lua.MathLibName, lua.OpenMath},     // Mathematical functions
		},
		Modules: &[]LuaLib{
			{"crypto", luaCrypto.Loader},     // calculate md5, sha256 hash for string
			{"humanize", luaHumanize.Loader}, // humanize times and sizes
			{"inspect", luaInspect.Loader},   // transforms any Lua value into a human-readable representation
			{"json", luaJson.Loader},         // json encoder/decoder
			{"regexp", luaRegexp.Loader},     // regular expressions
			{"tac", luaTac.Loader},           // line-by-line scanner
			{"time", luaTime.Loader},         // time related functions
			{"yaml", luaYaml.Loader},         // yaml encoder/decoder
		},
		UserModules:      nil,
		PlanVarName:      "plan",
		StateVarName:     "state",
		PrevStateVarName: "prevState",
	}
}

// New creates a new Warden instance.
func New(opts ...*Options) (*Warden, error) {
	opt := defaultOptions()
	for _, o := range opts {
		if err := mergo.Merge(opt, o, mergo.WithOverride); err != nil {
			return nil, xerrors.Errorf("failed to merge the configuration: %w", err)
		}
	}

	w := &Warden{
		options:  opt,
		luaState: lua.NewState(lua.Options{SkipOpenLibs: true}),
	}

	w.openLuaLibs(*opt.StandardLibs)
	w.preloadLuaModules(*opt.Modules)

	if opt.UserModules != nil {
		if err := w.preloadUserModules(*opt.UserModules); err != nil {
			return nil, err
		}
	}

	fn, err := w.luaState.LoadString(w.options.Script)
	if err != nil {
		return nil, xerrors.Errorf("invalid validation script: %w", err)
	}

	w.luaState.Push(fn)

	return w, nil
}

// ValidatePlan checks the validity of the specified plan with the configured
// scripts.
func (w *Warden) ValidatePlan(file afero.File) ([]string, error) {
	fi, err := file.Stat()
	if err != nil {
		return nil, xerrors.Errorf("failed to retrieve file information: %w", err)
	}

	planReader, err := planfile.OpenStream(file, fi.Size())
	if err != nil {
		return nil, xerrors.Errorf("failed to open plan file: %w", err)
	}

	plan, err := planReader.ReadPlan()
	if err != nil {
		return nil, xerrors.Errorf("failed to load plan data: %w", err)
	}

	state, err := planReader.ReadStateFile()
	if err != nil {
		return nil, xerrors.Errorf("failed to load state data: %w", err)
	}

	prevState, err := planReader.ReadStateFile()
	if err != nil {
		return nil, xerrors.Errorf("failed to load state data: %w", err)
	}

	config, _ := planReader.ReadConfig()

	w.luaState.SetGlobal(w.options.PlanVarName, luar.New(w.luaState, plan))
	w.luaState.SetGlobal(w.options.StateVarName, luar.New(w.luaState, state))
	w.luaState.SetGlobal(w.options.PrevStateVarName, luar.New(w.luaState, prevState))
	w.luaState.SetGlobal(w.options.ConfigVarName, luar.New(w.luaState, config))

	if err := w.luaState.PCall(0, lua.MultRet, nil); err != nil {
		return nil, xerrors.Errorf("failed to load and parse validation script: %w", err)
	}

	issues := checkResult(w.luaState)
	if len(issues) > 0 {
		return issues, ErrValidationFailed
	}

	return nil, nil
}

func (w *Warden) Close() {
	if w.luaState != nil {
		w.luaState.Close()
	}
}

func checkResult(luaState *lua.LState) (issues []string) {
	ret := luaState.Get(-1)

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

func (w *Warden) openLuaLibs(libs []LuaLib) {
	openLib := func(lib LuaLib) {
		w.luaState.Push(w.luaState.NewFunction(lib.Function))
		w.luaState.Push(lua.LString(lib.Name))
		w.luaState.Call(1, 0)
	}

	// Map iteration order in Go is deliberately randomised, so must open Load/Base prior to iterating.
	openLib(LuaLib{lua.LoadLibName, lua.OpenPackage})
	openLib(LuaLib{lua.BaseLibName, lua.OpenBase})

	for _, lib := range libs {
		openLib(lib)
	}
}

func (w *Warden) preloadLuaModules(modules []LuaLib) {
	for _, module := range modules {
		w.luaState.PreloadModule(module.Name, module.Function)
	}
}

func (w *Warden) preloadUserModules(modules []UserModule) error {
	for _, module := range modules {
		luaEnv := w.luaState.Get(lua.EnvironIndex)
		preload := w.luaState.GetField(w.luaState.GetField(luaEnv, "package"), "preload")

		if _, ok := preload.(*lua.LTable); !ok {
			w.luaState.RaiseError("package.preload must be a table")
		}

		script, err := w.luaState.LoadString(module.Script)
		if err != nil {
			return xerrors.Errorf("failed to load user module '%s': %w", module.Name, err)
		}

		w.luaState.SetField(preload, module.Name, script)
	}

	return nil
}
