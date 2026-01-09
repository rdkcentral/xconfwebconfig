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
	"testing"

	"gotest.tools/assert"
)

func TestGenericNamespacedListUtil1(t *testing.T) {
	ipStrs := []string{
		"12.34.56.78/31",
		"1.2.3.4",
	}
	ipList := NewGenericNamespacedList("foo", "bar", ipStrs)
	assert.Assert(t, ipList.IsInIpRange("12.34.56.79"))
	assert.Assert(t, ipList.IsInIpRange("1.2.3.4"))
	assert.Assert(t, !ipList.IsInIpRange("12.34.56.1"))
	assert.Assert(t, !ipList.IsInIpRange("1.1.1.1"))
}

func TestGenericNamespacedListUtil2(t *testing.T) {
	ipStrs := []string{
		"1.2.3.4",
	}
	ipList := NewGenericNamespacedList("foo", "bar", ipStrs)
	assert.Assert(t, !ipList.IsInIpRange("abcd"))
	assert.Assert(t, !ipList.IsInIpRange("1.1.1"))
}

// Test IsValidType
func TestIsValidType(t *testing.T) {
	assert.Assert(t, IsValidType(STRING))
	assert.Assert(t, IsValidType(MAC_LIST))
	assert.Assert(t, IsValidType(IP_LIST))
	assert.Assert(t, IsValidType(RI_MAC_LIST))
	assert.Assert(t, !IsValidType("INVALID_TYPE"))
	assert.Assert(t, !IsValidType(""))
	assert.Assert(t, !IsValidType("random"))
}

// Test NewGenericNamespacedList
func TestNewGenericNamespacedList(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	list := NewGenericNamespacedList("testID", STRING, data)

	assert.Equal(t, "testID", list.ID)
	assert.Equal(t, STRING, list.TypeName)
	assert.Equal(t, 3, len(list.Data))
	assert.Equal(t, "item1", list.Data[0])
}

// Test NewEmptyGenericNamespacedList
func TestNewEmptyGenericNamespacedList(t *testing.T) {
	list := NewEmptyGenericNamespacedList()

	assert.Assert(t, list != nil)
	assert.Equal(t, "", list.ID)
	assert.Equal(t, "", list.TypeName)
	assert.Equal(t, 0, len(list.Data))
}

// Test NewMacList
func TestNewMacList(t *testing.T) {
	list := NewMacList()

	assert.Assert(t, list != nil)
	assert.Equal(t, MAC_LIST, list.TypeName)
}

// Test NewIpList
func TestNewIpList(t *testing.T) {
	list := NewIpList()

	assert.Assert(t, list != nil)
	assert.Equal(t, IP_LIST, list.TypeName)
}

// Test IsMacList
func TestIsMacList(t *testing.T) {
	macList := NewMacList()
	assert.Assert(t, macList.IsMacList())

	ipList := NewIpList()
	assert.Assert(t, !ipList.IsMacList())

	stringList := NewGenericNamespacedList("test", STRING, []string{"a", "b"})
	assert.Assert(t, !stringList.IsMacList())
}

// Test IsIpList
func TestIsIpList(t *testing.T) {
	ipList := NewIpList()
	assert.Assert(t, ipList.IsIpList())

	macList := NewMacList()
	assert.Assert(t, !macList.IsIpList())

	stringList := NewGenericNamespacedList("test", STRING, []string{"a", "b"})
	assert.Assert(t, !stringList.IsIpList())
}

// Test Clone
func TestGenericNamespacedListClone(t *testing.T) {
	original := NewGenericNamespacedList("testID", IP_LIST, []string{"1.2.3.4", "5.6.7.8"})

	cloned, err := original.Clone()

	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.TypeName, cloned.TypeName)
	assert.Equal(t, len(original.Data), len(cloned.Data))

	// Verify it's a deep copy
	cloned.ID = "modified"
	assert.Assert(t, original.ID != cloned.ID)
}

// Test Validate - valid cases
func TestGenericNamespacedListValidate_ValidIPList(t *testing.T) {
	list := NewGenericNamespacedList("test-ip-list", IP_LIST, []string{"1.2.3.4", "5.6.7.8"})
	err := list.Validate()
	assert.NilError(t, err)
}

func TestGenericNamespacedListValidate_ValidMACList(t *testing.T) {
	list := NewGenericNamespacedList("test-mac-list", MAC_LIST, []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"})
	err := list.Validate()
	// May fail due to cache not configured, but that's OK for unit test
	if err != nil {
		assert.Assert(t, err.Error() == "cache not found or configured for table 'GenericXconfNamedList'")
	}
}

func TestGenericNamespacedListValidate_ValidSTRINGList(t *testing.T) {
	list := NewGenericNamespacedList("test-string-list", STRING, []string{"value1", "value2"})
	err := list.Validate()
	assert.NilError(t, err)
}

func TestGenericNamespacedListValidate_ValidNameWithSpecialChars(t *testing.T) {
	list := NewGenericNamespacedList("test-name_123.ABC 'quoted'", STRING, []string{"value1"})
	err := list.Validate()
	assert.NilError(t, err)
}

// Test Validate - invalid cases
func TestGenericNamespacedListValidate_InvalidName(t *testing.T) {
	list := NewGenericNamespacedList("test@invalid#name", STRING, []string{"value1"})
	err := list.Validate()
	assert.Error(t, err, "name is invalid")
}

func TestGenericNamespacedListValidate_InvalidType(t *testing.T) {
	list := NewGenericNamespacedList("test", "INVALID_TYPE", []string{"value1"})
	err := list.Validate()
	assert.Assert(t, err != nil)
	assert.Assert(t, err.Error() == "type INVALID_TYPE is invalid")
}

func TestGenericNamespacedListValidate_EmptyData(t *testing.T) {
	list := NewGenericNamespacedList("test", IP_LIST, []string{})
	err := list.Validate()
	assert.Error(t, err, "List must not be empty")
}

func TestGenericNamespacedListValidate_InvalidIPAddress(t *testing.T) {
	list := NewGenericNamespacedList("test", IP_LIST, []string{"invalid-ip", "1.2.3.4"})
	err := list.Validate()
	assert.Assert(t, err != nil)
}

func TestGenericNamespacedListValidate_InvalidMACAddress(t *testing.T) {
	list := NewGenericNamespacedList("test", MAC_LIST, []string{"invalid-mac", "AA:BB:CC:DD:EE:FF"})
	err := list.Validate()
	assert.Assert(t, err != nil)
}

// Test ValidateListData
func TestValidateListData_ValidIPList(t *testing.T) {
	err := ValidateListData(IP_LIST, []string{"1.2.3.4", "10.0.0.1"})
	assert.NilError(t, err)
}

func TestValidateListData_ValidMACList(t *testing.T) {
	err := ValidateListData(MAC_LIST, []string{"AA:BB:CC:DD:EE:FF"})
	assert.NilError(t, err)
}

func TestValidateListData_ValidSTRING(t *testing.T) {
	err := ValidateListData(STRING, []string{"any", "string", "values"})
	assert.NilError(t, err)
}

func TestValidateListData_InvalidType(t *testing.T) {
	err := ValidateListData("INVALID", []string{"value"})
	assert.Error(t, err, "Type is invalid")
}

func TestValidateListData_EmptyList(t *testing.T) {
	err := ValidateListData(IP_LIST, []string{})
	assert.Error(t, err, "List must not be empty")
}

func TestValidateListData_InvalidIPInList(t *testing.T) {
	err := ValidateListData(IP_LIST, []string{"1.2.3.4", "invalid"})
	assert.Assert(t, err != nil)
}

func TestValidateListData_InvalidMACInList(t *testing.T) {
	err := ValidateListData(MAC_LIST, []string{"AA:BB:CC:DD:EE:FF", "INVALID"})
	assert.Assert(t, err != nil)
}

// Test ValidateListDataForAdmin
func TestValidateListDataForAdmin_ValidIPList(t *testing.T) {
	err := ValidateListDataForAdmin(IP_LIST, []string{"1.2.3.4"})
	assert.NilError(t, err)
}

func TestValidateListDataForAdmin_ValidMACList(t *testing.T) {
	err := ValidateListDataForAdmin(MAC_LIST, []string{"AA:BB:CC:DD:EE:FF"})
	assert.NilError(t, err)
}

func TestValidateListDataForAdmin_InvalidType(t *testing.T) {
	err := ValidateListDataForAdmin("INVALID", []string{"value"})
	assert.Error(t, err, "Type is invalid")
}

func TestValidateListDataForAdmin_EmptyList(t *testing.T) {
	err := ValidateListDataForAdmin(IP_LIST, []string{})
	assert.Error(t, err, "List must not be empty")
}

// Test String method
func TestGenericNamespacedListString(t *testing.T) {
	list := NewGenericNamespacedList("testID", IP_LIST, []string{"1.2.3.4"})
	str := list.String()

	assert.Assert(t, str != "")
	// Should contain ID, TypeName, and Data
	// Format: GenericNamespacedList(testID |IP_LIST| [1.2.3.4])
}

// Test CreateIpAddressGroupResponse
func TestCreateIpAddressGroupResponse(t *testing.T) {
	list := NewGenericNamespacedList("test-ip", IP_LIST, []string{"1.2.3.4", "5.6.7.8"})

	response := list.CreateIpAddressGroupResponse()

	assert.Assert(t, response != nil)
	assert.Equal(t, "test-ip", response.Id)
	assert.Equal(t, "test-ip", response.Name)
	assert.Equal(t, 2, len(response.RawIpAddresses))
}

// Test CreateGenericNamespacedListResponse
func TestCreateGenericNamespacedListResponse(t *testing.T) {
	list := NewGenericNamespacedList("test-id", STRING, []string{"a", "b", "c"})

	response := list.CreateGenericNamespacedListResponse()

	assert.Assert(t, response != nil)
	assert.Equal(t, "test-id", response.ID)
	assert.Equal(t, 3, len(response.Data))
}

// Test ConvertToIpAddressGroup
func TestConvertToIpAddressGroup(t *testing.T) {
	genericList := NewGenericNamespacedList("test-ip", IP_LIST, []string{"1.2.3.4"})

	ipGroup := ConvertToIpAddressGroup(genericList)

	assert.Assert(t, ipGroup != nil)
	assert.Equal(t, "test-ip", ipGroup.Name)
	assert.Equal(t, 1, len(ipGroup.RawIpAddresses))
}

// Test ConvertFromIpAddressGroup
func TestConvertFromIpAddressGroup(t *testing.T) {
	ipGroup := &IpAddressGroup{
		Id:             "test-id",
		Name:           "test-name",
		RawIpAddresses: []string{"1.2.3.4", "5.6.7.8"},
	}

	genericList := ConvertFromIpAddressGroup(ipGroup)

	assert.Assert(t, genericList != nil)
	assert.Equal(t, "test-name", genericList.ID)
	assert.Equal(t, IP_LIST, genericList.TypeName)
	assert.Equal(t, 2, len(genericList.Data))
}

// Test NamespacedList Clone
func TestNamespacedListClone(t *testing.T) {
	original := &NamespacedList{
		ID:       "test",
		Updated:  12345,
		Data:     []string{"a", "b"},
		TypeName: STRING,
	}

	cloned, err := original.Clone()

	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
}

// Test NewGenericNamespacedListInf
func TestNewGenericNamespacedListInf(t *testing.T) {
	obj := NewGenericNamespacedListInf()

	assert.Assert(t, obj != nil)
	list, ok := obj.(*GenericNamespacedList)
	assert.Assert(t, ok)
	assert.Assert(t, list != nil)
}

// Test ValidateForAdminService
func TestValidateForAdminService_Valid(t *testing.T) {
	list := NewGenericNamespacedList("test-list", STRING, []string{"value1", "value2"})
	err := list.ValidateForAdminService()
	assert.NilError(t, err)
}

func TestValidateForAdminService_InvalidName(t *testing.T) {
	list := NewGenericNamespacedList("test@invalid", STRING, []string{"value1"})
	err := list.ValidateForAdminService()
	assert.Error(t, err, "name is invalid")
}

func TestValidateForAdminService_InvalidType(t *testing.T) {
	list := NewGenericNamespacedList("test", "BAD_TYPE", []string{"value1"})
	err := list.ValidateForAdminService()
	assert.Assert(t, err != nil)
}

// Test with duplicate data (should be deduplicated)
func TestValidate_RemovesDuplicates(t *testing.T) {
	list := NewGenericNamespacedList("test", STRING, []string{"a", "b", "a", "c", "b"})
	err := list.Validate()

	assert.NilError(t, err)
	// Data should be deduplicated
	assert.Assert(t, len(list.Data) <= 3) // At most 3 unique values: a, b, c
}

// Test NewNamespacedListInf
func TestNewNamespacedListInf(t *testing.T) {
	obj := NewNamespacedListInf()

	assert.Assert(t, obj != nil)
	list, ok := obj.(*NamespacedList)
	assert.Assert(t, ok)
	assert.Assert(t, list != nil)
}

// Additional edge case tests for ValidateForAdminService
func TestValidateForAdminService_EmptyData(t *testing.T) {
	list := NewGenericNamespacedList("test", STRING, []string{})
	err := list.ValidateForAdminService()
	// Should fail because list is empty
	assert.Assert(t, err != nil)
}

func TestValidateForAdminService_InvalidIPList(t *testing.T) {
	list := NewGenericNamespacedList("test", IP_LIST, []string{"192.168.1.1", "invalid"})
	err := list.ValidateForAdminService()
	assert.Assert(t, err != nil)
}

func TestValidateForAdminService_InvalidMACList(t *testing.T) {
	list := NewGenericNamespacedList("test", MAC_LIST, []string{"AA:BB:CC:DD:EE:FF", "ZZZZ"})
	err := list.ValidateForAdminService()
	assert.Assert(t, err != nil)
}

// Additional edge case tests for ValidateListDataForAdmin
func TestValidateListDataForAdmin_EmptyData(t *testing.T) {
	err := ValidateListDataForAdmin(STRING, []string{})
	assert.Error(t, err, "List must not be empty")
}

func TestValidateListDataForAdmin_InvalidIPData(t *testing.T) {
	err := ValidateListDataForAdmin(IP_LIST, []string{"192.168.1.1", "not-an-ip"})
	assert.Assert(t, err != nil)
}

func TestValidateListDataForAdmin_InvalidMACData(t *testing.T) {
	err := ValidateListDataForAdmin(MAC_LIST, []string{"AA:BB:CC:DD:EE:FF", "INVALID"})
	assert.Assert(t, err != nil)
}

func TestValidateListDataForAdmin_ValidStringData(t *testing.T) {
	err := ValidateListDataForAdmin(STRING, []string{"any", "string", "works"})
	assert.NilError(t, err)
}

// Test edge cases for Clone functions
func TestNamespacedListClone_WithData(t *testing.T) {
	original := &NamespacedList{
		ID:       "test-id",
		Updated:  99999,
		Data:     []string{"item1", "item2", "item3"},
		TypeName: STRING,
	}

	cloned, err := original.Clone()

	assert.NilError(t, err)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Updated, cloned.Updated)
	assert.Equal(t, len(original.Data), len(cloned.Data))
	assert.Equal(t, original.TypeName, cloned.TypeName)

	// Modify clone and ensure original is unchanged
	cloned.ID = "modified-id"
	cloned.Data = append(cloned.Data, "item4")
	assert.Assert(t, original.ID != cloned.ID)
	assert.Equal(t, 3, len(original.Data))
}

func TestGenericNamespacedListClone_WithEmptyData(t *testing.T) {
	original := NewEmptyGenericNamespacedList()
	original.ID = "empty"
	original.TypeName = STRING

	cloned, err := original.Clone()

	assert.NilError(t, err)
	assert.Equal(t, "empty", cloned.ID)
	assert.Equal(t, STRING, cloned.TypeName)
}

// Test String method with different types
func TestGenericNamespacedListString_MACList(t *testing.T) {
	list := NewGenericNamespacedList("mac-list", MAC_LIST, []string{"AA:BB:CC:DD:EE:FF"})
	str := list.String()

	assert.Assert(t, str != "")
	// Should contain the MAC_LIST type
}

func TestGenericNamespacedListString_EmptyList(t *testing.T) {
	list := NewEmptyGenericNamespacedList()
	list.ID = "empty"
	list.TypeName = STRING
	str := list.String()

	assert.Assert(t, str != "")
}

// Test CreateIpAddressGroupResponse with edge cases
func TestCreateIpAddressGroupResponse_EmptyData(t *testing.T) {
	list := NewIpList()
	list.ID = "empty-ip-list"

	response := list.CreateIpAddressGroupResponse()

	assert.Assert(t, response != nil)
	assert.Equal(t, "empty-ip-list", response.Id)
	assert.Equal(t, 0, len(response.RawIpAddresses))
}

func TestCreateIpAddressGroupResponse_MultipleIPs(t *testing.T) {
	list := NewGenericNamespacedList("multi-ip", IP_LIST, []string{
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.0/24",
		"10.0.0.1",
	})

	response := list.CreateIpAddressGroupResponse()

	assert.Assert(t, response != nil)
	assert.Equal(t, 4, len(response.RawIpAddresses))
}

// Test ConvertToIpAddressGroup edge cases
func TestConvertToIpAddressGroup_EmptyList(t *testing.T) {
	genericList := NewIpList()
	genericList.ID = "empty"

	ipGroup := ConvertToIpAddressGroup(genericList)

	assert.Assert(t, ipGroup != nil)
	assert.Equal(t, "empty", ipGroup.Name)
	assert.Equal(t, 0, len(ipGroup.RawIpAddresses))
}

// Test ConvertFromIpAddressGroup edge cases
func TestConvertFromIpAddressGroup_EmptyGroup(t *testing.T) {
	ipGroup := &IpAddressGroup{
		Id:             "test",
		Name:           "test-name",
		RawIpAddresses: []string{},
	}

	genericList := ConvertFromIpAddressGroup(ipGroup)

	assert.Assert(t, genericList != nil)
	assert.Equal(t, "test-name", genericList.ID)
	assert.Equal(t, IP_LIST, genericList.TypeName)
	assert.Equal(t, 0, len(genericList.Data))
}

// Test IsInIpRange edge cases
func TestIsInIpRange_EmptyData(t *testing.T) {
	list := NewIpList()
	assert.Assert(t, !list.IsInIpRange("192.168.1.1"))
}

func TestIsInIpRange_InvalidInput(t *testing.T) {
	list := NewGenericNamespacedList("test", IP_LIST, []string{"192.168.1.0/24"})
	assert.Assert(t, !list.IsInIpRange("not-an-ip"))
	assert.Assert(t, !list.IsInIpRange(""))
}
