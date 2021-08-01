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
	"github.com/spf13/afero"
	"golang.org/x/xerrors"

	"github.com/hexbee-net/horus/pkg/terraform/configs"
	"github.com/hexbee-net/horus/pkg/terraform/plans"
	"github.com/hexbee-net/horus/pkg/terraform/plans/planfile"
	"github.com/hexbee-net/horus/pkg/terraform/states/statefile"
)

type PlanFile struct {
	Plan      *plans.Plan
	State     *statefile.File
	PrevState *statefile.File
	Config    *configs.Config
}

func LoadPlanFile(file afero.File) (*PlanFile, error) {
	fi, err := file.Stat()
	if err != nil {
		return nil, xerrors.Errorf("failed to retrieve planFile information: %w", err)
	}

	planReader, err := planfile.OpenStream(file, fi.Size())
	if err != nil {
		return nil, xerrors.Errorf("failed to open plan planFile: %w", err)
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

	config, diags := planReader.ReadConfig()
	if diags.HasErrors() {
		return nil, xerrors.Errorf("failed to load configuration data: %w", diags.Err())
	}

	return &PlanFile{
		Plan:      plan,
		State:     state,
		PrevState: prevState,
		Config:    config,
	}, nil
}
