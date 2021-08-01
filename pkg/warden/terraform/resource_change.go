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

package terraform

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"

	"github.com/hexbee-net/horus/pkg/terraform/plans"
)

type ResourceChange struct {
	tfResource *plans.ResourceInstanceChange
}

// -----------------------------------------------------------------------------
// Lua Utilities

const luaResourceChangeTypeName = "resourceChange"

const (
	luaFunctionResourceChangeGetAddress       = "address"
	luaFunctionResourceChangeGetModuleAddress = "moduleAddress"
	luaFunctionResourceChangeGetType          = "type"
	luaFunctionResourceChangeGetName          = "name"
	luaFunctionResourceChangeGetMode          = "mode"
	luaFunctionResourceChangeGetIndex         = "index"
	luaFunctionResourceChangeGetProviderName  = "providerName"
	luaFunctionResourceChangeGetDeposed       = "deposed"
	luaFunctionResourceChangeGetChange        = "change"
)

// RegisterResourceChangeType registers the ResourceChange type inside the Lua state.
func RegisterResourceChangeType(ls *lua.LState) {
	var methods = map[string]lua.LGFunction{
		luaFunctionResourceChangeGetAddress:       resourceChangeGetAddress,
		luaFunctionResourceChangeGetModuleAddress: resourceChangeGetModuleAddress,
		luaFunctionResourceChangeGetType:          resourceChangeGetType,
		luaFunctionResourceChangeGetName:          resourceChangeGetName,
		luaFunctionResourceChangeGetMode:          resourceChangeGetMode,
		luaFunctionResourceChangeGetIndex:         resourceChangeGetIndex,
		luaFunctionResourceChangeGetProviderName:  resourceChangeGetProviderName,
		luaFunctionResourceChangeGetDeposed:       resourceChangeGetDeposed,
		luaFunctionResourceChangeGetChange:        resourceChangeGetChange,
	}

	mt := ls.NewTypeMetatable(luaResourceChangeTypeName)
	ls.SetGlobal(luaResourceChangeTypeName, mt)

	// methods
	ls.SetField(mt, "__index", ls.SetFuncs(ls.NewTable(), methods))
}

// CheckResourceChange checks whether the first lua argument is a *LUserData
// with *ResourceChange and returns this *ResourceChange.
func CheckResourceChange(ls *lua.LState) (*ResourceChange, error) {
	ud := ls.CheckUserData(1)
	if v, ok := ud.Value.(*ResourceChange); ok {
		return v, nil
	}

	ls.ArgError(1, fmt.Sprintf("%s expected", luaResourceChangeTypeName))

	return nil, ErrInvalidType
}

// -----------------------------------------------------------------------------
// Lua Functions

func resourceChangeGetAddress(ls *lua.LState) int {
	panic("not implemented")
}

func resourceChangeGetModuleAddress(ls *lua.LState) int {
	panic("not implemented")
}

func resourceChangeGetType(ls *lua.LState) int {
	r, err := CheckResourceChange(ls)
	if err != nil {
		return 0
	}

	ls.Push(lua.LString(r.tfResource.Addr.Resource.Resource.Type))

	return 1
}

func resourceChangeGetName(ls *lua.LState) int {
	r, err := CheckResourceChange(ls)
	if err != nil {
		return 0
	}

	ls.Push(lua.LString(r.tfResource.Addr.Resource.Resource.Name))

	return 1
}

func resourceChangeGetMode(ls *lua.LState) int {
	panic("not implemented")
}

func resourceChangeGetIndex(ls *lua.LState) int {
	panic("not implemented")
}

func resourceChangeGetProviderName(ls *lua.LState) int {
	panic("not implemented")
}

func resourceChangeGetDeposed(ls *lua.LState) int {
	panic("not implemented")
}

func resourceChangeGetChange(ls *lua.LState) int {
	panic("not implemented")
}
