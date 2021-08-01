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

import lua "github.com/yuin/gopher-lua"

// Module is  go code that can be called from Lua.
// It can be either directly opened and made available to the Lus script
// directly, or just preloaded, in which case it has to be included with
// a 'require' directive in the script before it can be used.
type Module struct {
	Name     string
	Function lua.LGFunction
}

// UserModule is a Lua sub-script that can be included in the main script
// through the use of a 'require' directive.
type UserModule struct {
	Name   string
	Script string
}

// StandardLibs returns the list of standard libraries to load for
// sandboxed usage.
func StandardLibs() []Module {
	return []Module{
		{lua.TabLibName, lua.OpenTable},     // Table manipulation
		{lua.StringLibName, lua.OpenString}, // String manipulation
		{lua.MathLibName, lua.OpenMath},     // Mathematical functions
	}
}
