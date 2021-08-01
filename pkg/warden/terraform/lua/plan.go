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
	"fmt"

	wlua "github.com/hexbee-net/horus/pkg/warden/lua"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/xerrors"

	"github.com/hexbee-net/horus/pkg/warden/terraform"
)

const luaPlanTypeName = "plan"

const (
	luaFunctionPlanFindResource = "findResource"
)

// RegisterPlanType registers the plan type inside the Lua state.
func RegisterPlanType(L *lua.LState) {
	var methods = map[string]lua.LGFunction{
		luaFunctionPlanFindResource: planFindResource,
	}

	mt := L.NewTypeMetatable(luaPlanTypeName)
	L.SetGlobal(luaPlanTypeName, mt)

	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
}

func LPlan(L *lua.LState, plan *terraform.Plan) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = plan
	L.SetMetatable(ud, L.GetTypeMetatable(luaPlanTypeName))

	return ud
}

// CheckPlan checks whether the first lua argument is a *LUserData with *Plan and returns this *Plan.
func CheckPlan(L *lua.LState) (*terraform.Plan, error) {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*terraform.Plan); ok {
		return v, nil
	}

	L.ArgError(1, "plan expected")

	return nil, xerrors.New("not a plan variable")
}

// -----------------------------------------------------------------------------
// Lua Functions

func planFindResource(L *lua.LState) int {
	// We use a flag instead of returning immediately to gather as much errors
	// in one run to spare the user from needing to run the script multiple
	// times before catching all the possible errors in the script.
	invalidCall := false

	p, err := CheckPlan(L)
	if err != nil {
		invalidCall = true
	}

	if top := L.GetTop(); top != 3 {
		if top < 3 {
			L.ArgError(1, fmt.Sprintf("not enough arguments in call to '%s'", luaFunctionPlanFindResource))
		} else {
			L.ArgError(1, fmt.Sprintf("too many arguments in call to '%s'", luaFunctionPlanFindResource))
		}

		invalidCall = true
	}

	resourceType, err := wlua.CheckString(L, 2)
	if err != nil {
		invalidCall = true
	}

	resourceName, err := wlua.CheckString(L, 3)
	if err != nil {
		invalidCall = true
	}

	if invalidCall {
		return 0
	}

	resource, err := p.FindResource(resourceType, resourceName)

	if resource != nil {
		L.Push(lua.LString("resource"))
		return 1
	}

	return 0
}
