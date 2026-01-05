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
package featurecontrol

import (
	"errors"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/stretchr/testify/assert"
)

// Test ToRfcResponse when feature is not whitelisted
func TestToRfcResponse_NotWhitelisted(t *testing.T) {
	feature := &rfc.Feature{
		Name:        "TestFeature",
		FeatureName: "test.feature",
		Enable:      true,
		Whitelisted: false,
		ConfigData: map[string]string{
			"key1": "value1",
		},
	}

	result := ToRfcResponse(feature)

	// Should return feature unchanged
	assert.Equal(t, feature, result)
	assert.False(t, result.Whitelisted)
	assert.Equal(t, "TestFeature", result.Name)
}

// Test ToRfcResponse with valid whitelist property
func TestToRfcResponse_WithValidWhitelistProperty(t *testing.T) {
	// Save original function and restore after test
	originalFunc := GetGenericNamedListOneByTypeFunc
	defer func() { GetGenericNamedListOneByTypeFunc = originalFunc }()

	// Mock the GetGenericNamedListOneByTypeFunc
	GetGenericNamedListOneByTypeFunc = func(id string, namespacedListType string) (*shared.GenericNamespacedList, error) {
		return &shared.GenericNamespacedList{
			ID:       "list123",
			TypeName: "MAC_LIST",
			Data:     []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"},
		}, nil
	}

	feature := &rfc.Feature{
		Name:        "WhitelistedFeature",
		FeatureName: "whitelist.feature",
		Enable:      true,
		Whitelisted: true,
		WhitelistProperty: &rfc.WhitelistProperty{
			Key:                "macAddress",
			Value:              "list123",
			NamespacedListType: "MAC_LIST",
			TypeName:           "MAC_LIST",
		},
		ConfigData: map[string]string{
			"param1": "value1",
		},
	}

	result := ToRfcResponse(feature)

	// Should populate Properties with namespaced list data
	assert.NotNil(t, result)
	assert.NotNil(t, result.Properties)
	assert.Contains(t, result.Properties, "list123")
	assert.Equal(t, []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"}, result.Properties["list123"])
	assert.Equal(t, "MAC_LIST", result.ListType)
	assert.Equal(t, 2, result.ListSize)
}

// Test ToRfcResponse with whitelist property but GetGenericNamedList returns error
func TestToRfcResponse_GetGenericNamedListError(t *testing.T) {
	originalFunc := GetGenericNamedListOneByTypeFunc
	defer func() { GetGenericNamedListOneByTypeFunc = originalFunc }()

	// Mock the function to return error
	GetGenericNamedListOneByTypeFunc = func(id string, namespacedListType string) (*shared.GenericNamespacedList, error) {
		return nil, errors.New("database error")
	}

	feature := &rfc.Feature{
		Name:        "WhitelistedFeature",
		FeatureName: "whitelist.feature",
		Enable:      true,
		Whitelisted: true,
		WhitelistProperty: &rfc.WhitelistProperty{
			Value:              "list123",
			NamespacedListType: "MAC_LIST",
			TypeName:           "MAC_LIST",
		},
		ConfigData: map[string]string{},
	}

	result := ToRfcResponse(feature)

	// Should still return the feature even on error
	assert.NotNil(t, result)
	assert.Equal(t, "WhitelistedFeature", result.Name)
	// Properties should not be set due to error
	assert.Nil(t, result.Properties)
}

// Test ToRfcResponse with whitelist property but GetGenericNamedList returns nil
func TestToRfcResponse_GetGenericNamedListReturnsNil(t *testing.T) {
	originalFunc := GetGenericNamedListOneByTypeFunc
	defer func() { GetGenericNamedListOneByTypeFunc = originalFunc }()

	// Mock the function to return nil
	GetGenericNamedListOneByTypeFunc = func(id string, namespacedListType string) (*shared.GenericNamespacedList, error) {
		return nil, nil
	}

	feature := &rfc.Feature{
		Name:        "WhitelistedFeature",
		Enable:      true,
		Whitelisted: true,
		WhitelistProperty: &rfc.WhitelistProperty{
			Value:              "list123",
			NamespacedListType: "MAC_LIST",
			TypeName:           "MAC_LIST",
		},
	}

	result := ToRfcResponse(feature)

	// Should return feature but Properties should not be set
	assert.NotNil(t, result)
	assert.Nil(t, result.Properties)
	assert.Equal(t, "", result.ListType)
	assert.Equal(t, 0, result.ListSize)
}

// Test ToRfcResponse with nil whitelist property
func TestToRfcResponse_NilWhitelistProperty(t *testing.T) {
	feature := &rfc.Feature{
		Name:              "WhitelistedFeature",
		Enable:            true,
		Whitelisted:       true,
		WhitelistProperty: nil,
	}

	result := ToRfcResponse(feature)

	// Should return feature unchanged
	assert.NotNil(t, result)
	assert.Equal(t, "WhitelistedFeature", result.Name)
	assert.Nil(t, result.Properties)
}

// Test ToRfcResponse with empty whitelist property value
func TestToRfcResponse_EmptyWhitelistPropertyValue(t *testing.T) {
	feature := &rfc.Feature{
		Name:        "WhitelistedFeature",
		Enable:      true,
		Whitelisted: true,
		WhitelistProperty: &rfc.WhitelistProperty{
			Value:              "", // Empty value
			NamespacedListType: "MAC_LIST",
			TypeName:           "MAC_LIST",
		},
	}

	result := ToRfcResponse(feature)

	// Should trigger warning log and return feature unchanged
	assert.NotNil(t, result)
	assert.Nil(t, result.Properties)
}

// Test ToRfcResponse with empty NamespacedListType
func TestToRfcResponse_EmptyNamespacedListType(t *testing.T) {
	feature := &rfc.Feature{
		Name:        "WhitelistedFeature",
		Enable:      true,
		Whitelisted: true,
		WhitelistProperty: &rfc.WhitelistProperty{
			Value:              "list123",
			NamespacedListType: "", // Empty type
			TypeName:           "MAC_LIST",
		},
	}

	result := ToRfcResponse(feature)

	// Should trigger warning log and return feature unchanged
	assert.NotNil(t, result)
	assert.Nil(t, result.Properties)
}

// Test ToRfcResponse with valid whitelist and empty data list
func TestToRfcResponse_WithEmptyDataList(t *testing.T) {
	originalFunc := GetGenericNamedListOneByTypeFunc
	defer func() { GetGenericNamedListOneByTypeFunc = originalFunc }()

	// Mock the function to return empty data list
	GetGenericNamedListOneByTypeFunc = func(id string, namespacedListType string) (*shared.GenericNamespacedList, error) {
		return &shared.GenericNamespacedList{
			ID:       "empty_list",
			TypeName: "MAC_LIST",
			Data:     []string{}, // Empty data
		}, nil
	}

	feature := &rfc.Feature{
		Name:        "WhitelistedFeature",
		Enable:      true,
		Whitelisted: true,
		WhitelistProperty: &rfc.WhitelistProperty{
			Value:              "empty_list",
			NamespacedListType: "MAC_LIST",
			TypeName:           "MAC_LIST",
		},
	}

	result := ToRfcResponse(feature)

	// Should set Properties with empty list
	assert.NotNil(t, result)
	assert.NotNil(t, result.Properties)
	assert.Equal(t, []string{}, result.Properties["empty_list"])
	assert.Equal(t, "MAC_LIST", result.ListType)
	assert.Equal(t, 0, result.ListSize)
}

// Test ToRfcResponse with large data list
func TestToRfcResponse_WithLargeDataList(t *testing.T) {
	originalFunc := GetGenericNamedListOneByTypeFunc
	defer func() { GetGenericNamedListOneByTypeFunc = originalFunc }()

	// Create large data list
	largeData := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeData[i] = "item" + string(rune(i))
	}

	GetGenericNamedListOneByTypeFunc = func(id string, namespacedListType string) (*shared.GenericNamespacedList, error) {
		return &shared.GenericNamespacedList{
			ID:       "large_list",
			TypeName: "ITEM_LIST",
			Data:     largeData,
		}, nil
	}

	feature := &rfc.Feature{
		Name:        "LargeListFeature",
		Enable:      true,
		Whitelisted: true,
		WhitelistProperty: &rfc.WhitelistProperty{
			Value:              "large_list",
			NamespacedListType: "ITEM_LIST",
			TypeName:           "ITEM_LIST",
		},
	}

	result := ToRfcResponse(feature)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Properties)
	assert.Equal(t, 100, result.ListSize)
	assert.Equal(t, largeData, result.Properties["large_list"])
}

// Test ToRfcResponse preserves existing feature properties
func TestToRfcResponse_PreservesExistingProperties(t *testing.T) {
	originalFunc := GetGenericNamedListOneByTypeFunc
	defer func() { GetGenericNamedListOneByTypeFunc = originalFunc }()

	GetGenericNamedListOneByTypeFunc = func(id string, namespacedListType string) (*shared.GenericNamespacedList, error) {
		return &shared.GenericNamespacedList{
			ID:       "list123",
			TypeName: "MAC_LIST",
			Data:     []string{"AA:BB:CC:DD:EE:FF"},
		}, nil
	}

	feature := &rfc.Feature{
		ID:                 "feat123",
		Name:               "TestFeature",
		FeatureName:        "test.feature",
		Enable:             true,
		EffectiveImmediate: true,
		Whitelisted:        true,
		WhitelistProperty: &rfc.WhitelistProperty{
			Value:              "list123",
			NamespacedListType: "MAC_LIST",
			TypeName:           "MAC_LIST",
		},
		ConfigData: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
	}

	result := ToRfcResponse(feature)

	// Should preserve all original properties
	assert.Equal(t, "feat123", result.ID)
	assert.Equal(t, "TestFeature", result.Name)
	assert.Equal(t, "test.feature", result.FeatureName)
	assert.True(t, result.Enable)
	assert.True(t, result.EffectiveImmediate)
	assert.True(t, result.Whitelisted)
	assert.NotNil(t, result.ConfigData)
	assert.Equal(t, "value1", result.ConfigData["param1"])
	assert.Equal(t, "value2", result.ConfigData["param2"])
	// And add new Properties
	assert.NotNil(t, result.Properties)
	assert.Contains(t, result.Properties, "list123")
}

// Test ToRfcResponse with different list types
func TestToRfcResponse_DifferentListTypes(t *testing.T) {
	testCases := []struct {
		name             string
		listType         string
		data             []string
		expectedListType string
	}{
		{
			name:             "IP_LIST",
			listType:         "IP_LIST",
			data:             []string{"192.168.1.1", "192.168.1.2"},
			expectedListType: "IP_LIST",
		},
		{
			name:             "MAC_LIST",
			listType:         "MAC_LIST",
			data:             []string{"AA:BB:CC:DD:EE:FF"},
			expectedListType: "MAC_LIST",
		},
		{
			name:             "STRING_LIST",
			listType:         "STRING_LIST",
			data:             []string{"item1", "item2", "item3"},
			expectedListType: "STRING_LIST",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalFunc := GetGenericNamedListOneByTypeFunc
			defer func() { GetGenericNamedListOneByTypeFunc = originalFunc }()

			GetGenericNamedListOneByTypeFunc = func(id string, namespacedListType string) (*shared.GenericNamespacedList, error) {
				return &shared.GenericNamespacedList{
					ID:       "list_" + tc.listType,
					TypeName: tc.listType,
					Data:     tc.data,
				}, nil
			}

			feature := &rfc.Feature{
				Name:        "TestFeature",
				Enable:      true,
				Whitelisted: true,
				WhitelistProperty: &rfc.WhitelistProperty{
					Value:              "list_" + tc.listType,
					NamespacedListType: tc.listType,
					TypeName:           tc.listType,
				},
			}

			result := ToRfcResponse(feature)

			assert.Equal(t, tc.expectedListType, result.ListType)
			assert.Equal(t, len(tc.data), result.ListSize)
			assert.Equal(t, tc.data, result.Properties["list_"+tc.listType])
		})
	}
}
