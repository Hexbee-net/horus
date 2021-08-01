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

	lua "github.com/yuin/gopher-lua"
	"golang.org/x/xerrors"

	wlua "github.com/hexbee-net/horus/pkg/warden/lua"
	"github.com/hexbee-net/horus/pkg/warden/terraform"
)

const luaPlanTypeName = "plan"

const (
	luaFunctionPlanFindResource = "findResource"
)

// RegisterPlanType registers the plan type inside the Lua state.
func RegisterPlanType(ls *lua.LState) {
	var methods = map[string]lua.LGFunction{
		luaFunctionPlanFindResource: planFindResource,
	}

	mt := ls.NewTypeMetatable(luaPlanTypeName)
	ls.SetGlobal(luaPlanTypeName, mt)

	// methods
	ls.SetField(mt, "__index", ls.SetFuncs(ls.NewTable(), methods))
}

func LPlan(ls *lua.LState, plan *terraform.Plan) *lua.LUserData {
	ud := ls.NewUserData()
	ud.Value = plan
	ls.SetMetatable(ud, ls.GetTypeMetatable(luaPlanTypeName))

	return ud
}

// CheckPlan checks whether the first lua argument is a *LUserData with *Plan and returns this *Plan.
func CheckPlan(ls *lua.LState) (*terraform.Plan, error) {
	ud := ls.CheckUserData(1)
	if v, ok := ud.Value.(*terraform.Plan); ok {
		return v, nil
	}

	ls.ArgError(1, "plan expected")

	return nil, xerrors.New("not a plan variable")
}

// -----------------------------------------------------------------------------
// Lua Functions

func planFindResource(ls *lua.LState) int {
	const (
		ArgCount           = 3
		ArgPosResourceType = 2
		ArgPosResourceName = 3
	)

	// We use a flag instead of returning immediately to gather as much errors
	// in one run to spare the user from needing to run the script multiple
	// times before catching all the possible errors in the script.
	invalidCall := false

	p, err := CheckPlan(ls)
	if err != nil {
		invalidCall = true
	}

	if top := ls.GetTop(); top != ArgCount {
		if top < ArgCount {
			ls.ArgError(1, fmt.Sprintf("not enough arguments in call to '%s'", luaFunctionPlanFindResource))
		} else {
			ls.ArgError(1, fmt.Sprintf("too many arguments in call to '%s'", luaFunctionPlanFindResource))
		}

		invalidCall = true
	}

	resourceType, err := wlua.CheckString(ls, ArgPosResourceType)
	if err != nil {
		invalidCall = true
	}

	resourceName, err := wlua.CheckString(ls, ArgPosResourceName)
	if err != nil {
		invalidCall = true
	}

	if invalidCall {
		return 0
	}

	resource, err := p.FindResource(resourceType, resourceName)
	if err != nil {
		ls.RaiseError("failed to search for resource %s.%s in plan file", resourceType, resourceName)

		return 0
	}

	if resource != nil {
		ls.Push(lua.LString("resource"))

		return 1
	}

	return 0
}
