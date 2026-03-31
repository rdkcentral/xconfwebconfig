/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package estbfirmware

import (
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"gotest.tools/assert"
)

func TestRebootImmediatelyFilter_Creation(t *testing.T) {
	filter := &RebootImmediatelyFilter{
		Id:   "test-id-123",
		Name: "Test Reboot Filter",
	}
	
	assert.Equal(t, "test-id-123", filter.Id)
	assert.Equal(t, "Test Reboot Filter", filter.Name)
	assert.Assert(t, filter.IpAddressGroup == nil)
	assert.Assert(t, filter.Environments == nil)
	assert.Assert(t, filter.Models == nil)
	assert.Equal(t, "", filter.MacAddress)
}

func TestRebootImmediatelyFilter_WithIpAddressGroup(t *testing.T) {
	ipGroup := &shared.IpAddressGroup{
		Id:   "ip-group-1",
		Name: "Test IP Group",
	}
	
	filter := &RebootImmediatelyFilter{
		Id:             "filter-1",
		Name:           "Filter with IP",
		IpAddressGroup: []*shared.IpAddressGroup{ipGroup},
	}
	
	assert.Equal(t, 1, len(filter.IpAddressGroup))
	assert.Equal(t, "ip-group-1", filter.IpAddressGroup[0].Id)
	assert.Equal(t, "Test IP Group", filter.IpAddressGroup[0].Name)
}

func TestRebootImmediatelyFilter_WithEnvironments(t *testing.T) {
	filter := &RebootImmediatelyFilter{
		Id:           "filter-2",
		Name:         "Filter with Environments",
		Environments: []string{"DEV", "QA", "PROD"},
	}
	
	assert.Equal(t, 3, len(filter.Environments))
	assert.Equal(t, "DEV", filter.Environments[0])
	assert.Equal(t, "QA", filter.Environments[1])
	assert.Equal(t, "PROD", filter.Environments[2])
}

func TestRebootImmediatelyFilter_WithModels(t *testing.T) {
	filter := &RebootImmediatelyFilter{
		Id:     "filter-3",
		Name:   "Filter with Models",
		Models: []string{"MODEL-X1", "MODEL-X2", "MODEL-X3"},
	}
	
	assert.Equal(t, 3, len(filter.Models))
	assert.Equal(t, "MODEL-X1", filter.Models[0])
	assert.Equal(t, "MODEL-X2", filter.Models[1])
	assert.Equal(t, "MODEL-X3", filter.Models[2])
}

func TestRebootImmediatelyFilter_WithMacAddress(t *testing.T) {
	filter := &RebootImmediatelyFilter{
		Id:         "filter-4",
		Name:       "Filter with MAC",
		MacAddress: "AA:BB:CC:DD:EE:FF",
	}
	
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", filter.MacAddress)
}

func TestRebootImmediatelyFilter_CompleteFilter(t *testing.T) {
	ipGroup1 := &shared.IpAddressGroup{Id: "ip-1", Name: "Group 1"}
	ipGroup2 := &shared.IpAddressGroup{Id: "ip-2", Name: "Group 2"}
	
	filter := &RebootImmediatelyFilter{
		Id:             "complete-filter",
		Name:           "Complete Reboot Filter",
		IpAddressGroup: []*shared.IpAddressGroup{ipGroup1, ipGroup2},
		Environments:   []string{"PROD", "STAGE"},
		Models:         []string{"MODEL-A", "MODEL-B"},
		MacAddress:     "11:22:33:44:55:66",
	}
	
	// Verify all fields
	assert.Equal(t, "complete-filter", filter.Id)
	assert.Equal(t, "Complete Reboot Filter", filter.Name)
	assert.Equal(t, 2, len(filter.IpAddressGroup))
	assert.Equal(t, "ip-1", filter.IpAddressGroup[0].Id)
	assert.Equal(t, "ip-2", filter.IpAddressGroup[1].Id)
	assert.Equal(t, 2, len(filter.Environments))
	assert.Equal(t, "PROD", filter.Environments[0])
	assert.Equal(t, 2, len(filter.Models))
	assert.Equal(t, "MODEL-A", filter.Models[0])
	assert.Equal(t, "11:22:33:44:55:66", filter.MacAddress)
}

func TestRebootImmediatelyFilter_EmptyArrays(t *testing.T) {
	filter := &RebootImmediatelyFilter{
		Id:             "filter-empty",
		Name:           "Empty Arrays Filter",
		IpAddressGroup: []*shared.IpAddressGroup{},
		Environments:   []string{},
		Models:         []string{},
	}
	
	assert.Equal(t, 0, len(filter.IpAddressGroup))
	assert.Equal(t, 0, len(filter.Environments))
	assert.Equal(t, 0, len(filter.Models))
}
