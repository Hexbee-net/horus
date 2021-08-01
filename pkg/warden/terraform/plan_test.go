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
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDataPath(t *testing.T, localPath string) string {
	t.Helper()
	return path.Clean(path.Join("../../../testData/", localPath))
}

func TestPlan_FindResource_WIP(t *testing.T) {
	testFs := afero.NewReadOnlyFs(afero.NewOsFs())

	file, err := testFs.Open(getTestDataPath(t, "tf-planfile"))
	require.NoError(t, err)

	planFile, err := LoadPlanFile(file)
	require.NoError(t, err)

	plan := Plan{
		tfPlan: planFile.Plan,
	}

	resource, err := plan.FindResource("aws_instance", "multiple_resource")

	assert.NotNil(t, resource)
	assert.NoError(t, err)
}
