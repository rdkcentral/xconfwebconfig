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
package shared

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
)

func TestIpAddressGroup(t *testing.T) {
	ipaddrs := []IpAddress{
		*NewIpAddress("192.168.1.1"),
		*NewIpAddress("192.168.1.2"),
		*NewIpAddress("192.168.1.3"),
		*NewIpAddress("192.168.2.3"),
		*NewIpAddress("192.168.3.3"),
	}

	g := &IpAddressGroup{}

	s1 := "foo"
	g.Id = s1
	s2 := "bar"
	g.Name = s2
	g.IpAddresses = ipaddrs

	assert.Equal(t, s1, g.Id)
	assert.Equal(t, s2, g.Name)
	s3 := "192.168.1.2"
	assert.Assert(t, g.IsInRange(s3))
	ipaddr3 := NewIpAddress(s3)
	assert.Assert(t, g.IsInRange(*ipaddr3))

	assert.Assert(t, g.IsInRange("10.0.0.1", "192.168.3.3"))
	assert.Assert(t, !g.IsInRange("10.0.0.1", "10.0.0.2"))

	assert.Assert(t, !g.IsInRange(
		*NewIpAddress("192.168.0.3"),
		*NewIpAddress("192.168.0.4"),
		*NewIpAddress("192.168.0.5"),
	))
}

// Test Clone
func TestIpAddressGroup_Clone(t *testing.T) {
	original := &IpAddressGroup{
		Id:             "group-1",
		Name:           "Test Group",
		RawIpAddresses: []string{"192.168.1.1", "192.168.1.2"},
	}
	original.SetIpAddresses(original.RawIpAddresses)

	cloned, err := original.Clone()

	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.Id, cloned.Id)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, len(original.IpAddresses), len(cloned.IpAddresses))

	// Verify it's a deep copy
	cloned.Id = "modified"
	assert.Assert(t, original.Id != cloned.Id)
}

// Test NewIpAddressGroupInf
func TestNewIpAddressGroupInf(t *testing.T) {
	obj := NewIpAddressGroupInf()

	assert.Assert(t, obj != nil)
	group, ok := obj.(*IpAddressGroup)
	assert.Assert(t, ok)
	assert.Assert(t, group != nil)
}

// Test NewEmptyIpAddressGroup
func TestNewEmptyIpAddressGroup(t *testing.T) {
	group := NewEmptyIpAddressGroup()

	assert.Assert(t, group != nil)
	assert.Equal(t, "", group.Id)
	assert.Equal(t, "", group.Name)
	assert.Equal(t, 0, len(group.IpAddresses))
	assert.Equal(t, 0, len(group.RawIpAddresses))
}

// Test NewIpAddressGroupWithAddrStrings
func TestNewIpAddressGroupWithAddrStrings(t *testing.T) {
	addrs := []string{"192.168.1.1", "192.168.1.2", "10.0.0.1"}
	group := NewIpAddressGroupWithAddrStrings("group-1", "Test Group", addrs)

	assert.Assert(t, group != nil)
	assert.Equal(t, "group-1", group.Id)
	assert.Equal(t, "Test Group", group.Name)
	assert.Equal(t, 3, len(group.RawIpAddresses))
	assert.Equal(t, 3, len(group.IpAddresses))
	assert.Assert(t, group.IsInRange("192.168.1.1"))
}

// Test NewIpAddressGroupWithAddrStrings with empty list
func TestNewIpAddressGroupWithAddrStrings_Empty(t *testing.T) {
	group := NewIpAddressGroupWithAddrStrings("empty-group", "Empty", []string{})

	assert.Assert(t, group != nil)
	assert.Equal(t, "empty-group", group.Id)
	assert.Equal(t, 0, len(group.IpAddresses))
}

// Test NewIpAddressGroup (copy constructor)
func TestNewIpAddressGroup_Copy(t *testing.T) {
	original := &IpAddressGroup{
		Id:          "original-id",
		Name:        "Original",
		IpAddresses: []IpAddress{*NewIpAddress("192.168.1.1"), *NewIpAddress("192.168.1.2")},
	}

	copied := NewIpAddressGroup(original)

	assert.Assert(t, copied != nil)
	assert.Equal(t, original.Id, copied.Id)
	assert.Equal(t, original.Name, copied.Name)
	assert.Equal(t, len(original.IpAddresses), len(copied.IpAddresses))

	// Verify it's a copy (modifying copy doesn't affect original)
	copied.Id = "modified"
	assert.Equal(t, "original-id", original.Id)
}

// Test UnmarshalJSON
func TestIpAddressGroup_UnmarshalJSON(t *testing.T) {
	jsonStr := `{
		"id": "test-group",
		"name": "Test Group",
		"ipAddresses": ["192.168.1.1", "192.168.1.2", "10.0.0.1"]
	}`

	var group IpAddressGroup
	err := json.Unmarshal([]byte(jsonStr), &group)

	assert.NilError(t, err)
	assert.Equal(t, "test-group", group.Id)
	assert.Equal(t, "Test Group", group.Name)
	assert.Equal(t, 3, len(group.RawIpAddresses))
	assert.Equal(t, 3, len(group.IpAddresses))
	assert.Assert(t, group.IsInRange("192.168.1.1"))
	assert.Assert(t, group.IsInRange("10.0.0.1"))
}

// Test UnmarshalJSON with empty addresses
func TestIpAddressGroup_UnmarshalJSON_EmptyAddresses(t *testing.T) {
	jsonStr := `{
		"id": "empty-group",
		"name": "Empty Group",
		"ipAddresses": []
	}`

	var group IpAddressGroup
	err := json.Unmarshal([]byte(jsonStr), &group)

	assert.NilError(t, err)
	assert.Equal(t, "empty-group", group.Id)
	assert.Equal(t, 0, len(group.IpAddresses))
}

// Test UnmarshalJSON with invalid JSON
func TestIpAddressGroup_UnmarshalJSON_Invalid(t *testing.T) {
	jsonStr := `{invalid json}`

	var group IpAddressGroup
	err := json.Unmarshal([]byte(jsonStr), &group)

	assert.Assert(t, err != nil)
}

// Test SetIpAddresses
func TestIpAddressGroup_SetIpAddresses(t *testing.T) {
	group := NewEmptyIpAddressGroup()
	addrs := []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"}

	group.SetIpAddresses(addrs)

	assert.Equal(t, 3, len(group.RawIpAddresses))
	assert.Equal(t, 3, len(group.IpAddresses))
	assert.Equal(t, "192.168.1.1", group.RawIpAddresses[0])
	assert.Assert(t, group.IsInRange("192.168.1.1"))
}

// Test SetIpAddresses with invalid IP
func TestIpAddressGroup_SetIpAddresses_WithInvalidIP(t *testing.T) {
	group := NewEmptyIpAddressGroup()
	addrs := []string{"192.168.1.1", "invalid-ip", "10.0.0.1"}

	group.SetIpAddresses(addrs)

	// RawIpAddresses should contain all including invalid
	assert.Equal(t, 3, len(group.RawIpAddresses))
	// IpAddresses should only contain valid IPs (invalid-ip skipped)
	assert.Equal(t, 2, len(group.IpAddresses))
}

// Test SetIpAddresses overwrites existing
func TestIpAddressGroup_SetIpAddresses_Overwrites(t *testing.T) {
	group := NewIpAddressGroupWithAddrStrings("g1", "Group", []string{"1.1.1.1"})
	assert.Equal(t, 1, len(group.IpAddresses))

	group.SetIpAddresses([]string{"2.2.2.2", "3.3.3.3"})

	assert.Equal(t, 2, len(group.IpAddresses))
	assert.Assert(t, !group.IsInRange("1.1.1.1")) // Old address removed
	assert.Assert(t, group.IsInRange("2.2.2.2"))  // New addresses present
}

// Test IsInRange with multiple arguments
func TestIpAddressGroup_IsInRange_MultipleArgs(t *testing.T) {
	group := NewIpAddressGroupWithAddrStrings("g1", "Group", []string{"192.168.1.1"})

	// None match
	assert.Assert(t, !group.IsInRange("10.0.0.1", "10.0.0.2", "10.0.0.3"))

	// One matches
	assert.Assert(t, group.IsInRange("10.0.0.1", "192.168.1.1", "10.0.0.3"))
}

// Test IsInRange with empty group
func TestIpAddressGroup_IsInRange_EmptyGroup(t *testing.T) {
	group := NewEmptyIpAddressGroup()

	assert.Assert(t, !group.IsInRange("192.168.1.1"))
	assert.Assert(t, !group.IsInRange("192.168.1.1", "10.0.0.1"))
}

// Test IsInRange with CIDR notation
func TestIpAddressGroup_IsInRange_CIDR(t *testing.T) {
	group := NewIpAddressGroupWithAddrStrings("g1", "Group", []string{"192.168.1.0/24"})

	// Should match IPs in the range
	assert.Assert(t, group.IsInRange("192.168.1.1"))
	assert.Assert(t, group.IsInRange("192.168.1.254"))

	// Should not match IPs outside the range
	assert.Assert(t, !group.IsInRange("192.168.2.1"))
	assert.Assert(t, !group.IsInRange("192.168.0.1"))
}
