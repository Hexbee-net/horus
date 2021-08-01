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
	"reflect"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/hexbee-net/horus/pkg/warden/terraform"
)

const (
	planFieldName      = "plan"
	stateFieldName     = "state"
	prevStateFieldName = "prevState"
	configFieldName    = "config"
)

var exports = map[string]lua.LGFunction{
	//"filter": filter,
}

func GetLoader(planFile *terraform.PlanFile) lua.LGFunction {
	return func(L *lua.LState) int {
		// register user types
		RegisterPlanType(L)

		// register functions
		mod := L.SetFuncs(L.NewTable(), exports)

		// register fields
		L.SetField(mod, planFieldName, LPlan(L, &terraform.Plan{Plan: planFile.Plan}))
		L.SetField(mod, stateFieldName, luar.New(L, planFile.State))
		L.SetField(mod, prevStateFieldName, luar.New(L, planFile.PrevState))
		L.SetField(mod, configFieldName, luar.New(L, planFile.Config))

		// returns the module
		L.Push(mod)

		return 1
	}
}

func newUserDate(L *lua.LState, value interface{}) lua.LValue {
	if value == nil {
		return lua.LNil
	}
	if lval, ok := value.(lua.LValue); ok {
		return lval
	}

	val := reflect.ValueOf(value)

	ud := L.NewUserData()
	ud.Value = val.Interface()
	return ud
}
