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

package warden

import (
	luaCrypto "github.com/vadv/gopher-lua-libs/crypto"
	luaHumanize "github.com/vadv/gopher-lua-libs/humanize"
	luaInspect "github.com/vadv/gopher-lua-libs/inspect"
	luaJson "github.com/vadv/gopher-lua-libs/json"
	luaRegexp "github.com/vadv/gopher-lua-libs/regexp"
	luaTac "github.com/vadv/gopher-lua-libs/tac"
	luaTime "github.com/vadv/gopher-lua-libs/time"
	luaYaml "github.com/vadv/gopher-lua-libs/yaml"

	wlua "github.com/hexbee-net/horus/pkg/warden/lua"
)

type Options struct {
	Libs        []wlua.Module
	Modules     []wlua.Module
	UserModules []wlua.UserModule
	Script      string
}

func DefaultPreloadModules() []wlua.Module {
	return []wlua.Module{
		{Name: "crypto", Function: luaCrypto.Loader},     // calculate md5, sha256 hash for string
		{Name: "humanize", Function: luaHumanize.Loader}, // humanize times and sizes
		{Name: "inspect", Function: luaInspect.Loader},   // transforms any Lua value into a human-readable representation
		{Name: "json", Function: luaJson.Loader},         // json encoder/decoder
		{Name: "regexp", Function: luaRegexp.Loader},     // regular expressions
		{Name: "tac", Function: luaTac.Loader},           // line-by-line scanner
		{Name: "time", Function: luaTime.Loader},         // time related functions
		{Name: "yaml", Function: luaYaml.Loader},         // yaml encoder/decoder
	}
}

// DefaultOptions returns the default options used when creating a Warden
// instance with New.
func DefaultOptions() *Options {
	return &Options{
		Libs:        wlua.StandardLibs(),
		Modules:     DefaultPreloadModules(),
		UserModules: nil,
		Script:      "",
	}
}
