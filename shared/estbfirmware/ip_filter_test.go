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

	"gotest.tools/assert"
)

func TestNewEmptyIpFilter(t *testing.T) {
	filter := NewEmptyIpFilter()
	assert.Assert(t, filter != nil)
	assert.Equal(t, "", filter.Id)
	assert.Equal(t, "", filter.Name)
	assert.Assert(t, filter.IpAddressGroup == nil)
	assert.Equal(t, false, filter.Warehouse)
}

func TestIsLetter(t *testing.T) {
	// Test with all letters
	assert.Assert(t, IsLetter("abc"))
	assert.Assert(t, IsLetter("ABC"))
	assert.Assert(t, IsLetter("aBc"))
	
	// Test with numbers
	assert.Assert(t, !IsLetter("abc123"))
	assert.Assert(t, !IsLetter("123"))
	assert.Assert(t, !IsLetter("a1b"))
	
	// Test with special characters
	assert.Assert(t, !IsLetter("abc-def"))
	assert.Assert(t, !IsLetter("abc_def"))
	assert.Assert(t, !IsLetter("abc def"))
	
	// Test empty string
	assert.Assert(t, IsLetter(""))
	
	// Test single character
	assert.Assert(t, IsLetter("a"))
	assert.Assert(t, IsLetter("Z"))
	assert.Assert(t, !IsLetter("1"))
	assert.Assert(t, !IsLetter("-"))
}

func TestIsLower(t *testing.T) {
	// Test with all lowercase
	assert.Assert(t, IsLower("abc"))
	assert.Assert(t, IsLower("abcdef"))
	
	// Test with uppercase
	assert.Assert(t, !IsLower("ABC"))
	assert.Assert(t, !IsLower("Abc"))
	assert.Assert(t, !IsLower("aBc"))
	assert.Assert(t, !IsLower("abC"))
	
	// Test with numbers (should return true since numbers are not uppercase letters)
	assert.Assert(t, IsLower("abc123"))
	assert.Assert(t, IsLower("123"))
	
	// Test with special characters
	assert.Assert(t, IsLower("abc-def"))
	assert.Assert(t, IsLower("abc_def"))
	
	// Test empty string
	assert.Assert(t, IsLower(""))
	
	// Test single character
	assert.Assert(t, IsLower("a"))
	assert.Assert(t, !IsLower("A"))
}

func TestIpFilter_IsWarehouse(t *testing.T) {
	// Test lowercase letters only - should be warehouse
	filter1 := &IpFilter{Id: "abc"}
	assert.Assert(t, filter1.IsWarehouse())
	
	filter2 := &IpFilter{Id: "xyz"}
	assert.Assert(t, filter2.IsWarehouse())
	
	// Test with uppercase - should not be warehouse
	filter3 := &IpFilter{Id: "ABC"}
	assert.Assert(t, !filter3.IsWarehouse())
	
	filter4 := &IpFilter{Id: "Abc"}
	assert.Assert(t, !filter4.IsWarehouse())
	
	// Test with numbers - should not be warehouse
	filter5 := &IpFilter{Id: "abc123"}
	assert.Assert(t, !filter5.IsWarehouse())
	
	filter6 := &IpFilter{Id: "123"}
	assert.Assert(t, !filter6.IsWarehouse())
	
	// Test with special characters - should not be warehouse
	filter7 := &IpFilter{Id: "abc-def"}
	assert.Assert(t, !filter7.IsWarehouse())
	
	// Test empty - should be warehouse (all letters are lowercase)
	filter8 := &IpFilter{Id: ""}
	assert.Assert(t, filter8.IsWarehouse())
	
	// Test UUID format - should not be warehouse
	filter9 := &IpFilter{Id: "550e8400-e29b-41d4-a716-446655440000"}
	assert.Assert(t, !filter9.IsWarehouse())
}

func TestIpFilter_IsWarehouseEdgeCases(t *testing.T) {
	// Test with mixed case
	filter1 := &IpFilter{Id: "aBc"}
	assert.Assert(t, !filter1.IsWarehouse(), "Mixed case should not be warehouse")
	
	// Test with single lowercase letter
	filter2 := &IpFilter{Id: "a"}
	assert.Assert(t, filter2.IsWarehouse(), "Single lowercase letter should be warehouse")
	
	// Test with single uppercase letter
	filter3 := &IpFilter{Id: "A"}
	assert.Assert(t, !filter3.IsWarehouse(), "Single uppercase letter should not be warehouse")
	
	// Test with lowercase and numbers
	filter4 := &IpFilter{Id: "abc123def"}
	assert.Assert(t, !filter4.IsWarehouse(), "Letters with numbers should not be warehouse")
	
	// Test with lowercase and underscore
	filter5 := &IpFilter{Id: "abc_def"}
	assert.Assert(t, !filter5.IsWarehouse(), "Letters with underscore should not be warehouse")
}
