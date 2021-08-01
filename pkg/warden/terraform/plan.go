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
	"github.com/hexbee-net/horus/pkg/terraform/plans"
)

type Plan struct {
	*plans.Plan
}

func (p *Plan) FindResource(resourceType string, resourceName string) ([]*ResourceChange, error) {
	resources := make([]*ResourceChange, 0)

	for _, r := range p.Changes.Resources {
		if r.Addr.Resource.Resource.Type == resourceType && r.Addr.Resource.Resource.Name == resourceName {
			rc := &ResourceChange{
				//Address: AbsResourceInstance{
				//	tfAddr: &r.Addr,
				//},
				////ModuleAddress: "foo",
				////Type:         r.Addr.Resource.Resource.Type,
				////Name:         r.Addr.Resource.Resource.Name,
				//Mode:         "foo",
				//Index:        "foo",
				//ProviderName: "foo",
				//Deposed:      "foo",
			}
			resources = append(resources, rc)
		}
	}

	return resources, nil
}
