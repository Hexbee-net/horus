// Copyright © 2021 Xavier Basty <xavier@hexbee.net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lua

import (
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/xerrors"
)

// LState is a wrapper around lua.LState to augment it with our own methods.
type LState struct {
	*lua.LState
}

// NewState creates and initialize a new LState.
func NewState(opts ...lua.Options) *LState {
	return &LState{lua.NewState(opts...)}
}

func AsLState(L *lua.LState) *LState {
	return &LState{L}
}

// OpenModules loads the specified modules in the current LState.
// A loaded module is directly usable by the script without having to use
// a 'require' directive first.
func (ls *LState) OpenModules(modules []Module) {
	openModule := func(m Module) {
		ls.Push(ls.NewFunction(m.Function))
		ls.Push(lua.LString(m.Name))
		ls.Call(1, 0)
	}

	// Map iteration order in Go is deliberately randomised, so must open Load/Base prior to iterating.
	openModule(Module{lua.LoadLibName, lua.OpenPackage})
	openModule(Module{lua.BaseLibName, lua.OpenBase})

	for _, m := range modules {
		openModule(m)
	}
}

// PreloadModules preloads the specified modules in the current LState.
// A preloaded module is made available for inclusion but is not readily
// available to the script without the use of a 'require' directive first.
func (ls *LState) PreloadModules(modules []Module) {
	for _, m := range modules {
		ls.PreloadModule(m.Name, m.Function)
	}
}

func (ls *LState) PreloadUserModule(modules []UserModule) error {
	if len(modules) == 0 {
		return nil
	}

	luaEnv := ls.Get(lua.EnvironIndex)
	preload := ls.GetField(ls.GetField(luaEnv, "package"), "preload")

	if _, ok := preload.(*lua.LTable); !ok {
		ls.RaiseError("package.preload must be a table")

		return xerrors.New("invalid value for package.preload.")
	}

	for _, module := range modules {
		script, err := ls.LoadString(module.Script)
		if err != nil {
			return xerrors.Errorf("failed to load user module '%s': %w", module.Name, err)
		}

		ls.SetField(preload, module.Name, script)
	}

	return nil
}

func (ls *LState) CheckType(n int, typ lua.LValueType) error {
	return CheckType(ls.LState, n, typ)
}

func CheckType(L *lua.LState, n int, typ lua.LValueType) error {
	v := L.Get(n)
	if v.Type() != typ {
		L.TypeError(n, typ)

		return xerrors.Errorf("invalid argument #%v (%v expected, got %v)", n, typ.String(), L.Get(n).Type().String())
	}

	return nil
}

func (ls *LState) CheckString(n int) (string, error) {
	return CheckString(ls.LState, n)
}

func CheckString(L *lua.LState, n int) (string, error) {
	v := L.Get(n)
	if lv, ok := v.(lua.LString); ok {
		return string(lv), nil
	} else if lua.LVCanConvToString(v) {
		return L.ToString(n), nil
	}

	L.TypeError(n, lua.LTString)

	return "", xerrors.Errorf("invalid argument #%v (%v expected, got %v)", n, lua.LTString.String(), L.Get(n).Type().String())
}
