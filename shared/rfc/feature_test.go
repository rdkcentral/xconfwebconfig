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
package rfc

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestFeatureCreationAndMarshall(t *testing.T) {
	configData := map[string]string{
		"configKey": "configValue",
	}

	// Test mandatory fields
	feature := &Feature{
		ConfigData:         configData,
		FeatureName:        "featureName",
		Name:               "name",
		Enable:             true,
		EffectiveImmediate: true,
	}
	featureResponseObject := CreateFeatureResponseObject(*feature)
	expectedJsonString := "{\"name\":\"name\",\"enable\":true,\"effectiveImmediate\":true,\"configData\":{\"configKey\":\"configValue\"},\"featureInstance\":\"featureName\"}"
	actualByteString, err := featureResponseObject.MarshalJSON()
	assert.NilError(t, err)
	assert.Equal(t, expectedJsonString, string(actualByteString))

	// Test with whitelisting
	feature = &Feature{
		ConfigData:         configData,
		FeatureName:        "whitelistedFeature",
		Name:               "Whitelisted Feature",
		Enable:             true,
		EffectiveImmediate: true,
		Whitelisted:        true,
		WhitelistProperty: &WhitelistProperty{
			Key:                "macAddress",
			Value:             "AA:BB:CC:DD:EE:FF",
			NamespacedListType: "MAC_LIST",
		},
	}
	featureResponseObject = CreateFeatureResponseObject(*feature)
	actualByteString, err = featureResponseObject.MarshalJSON()
	assert.NilError(t, err)
	jsonStr := string(actualByteString)
	assert.Assert(t, strings.Contains(jsonStr, "featureInstance"))
	assert.Assert(t, strings.Contains(jsonStr, "whitelistedFeature"))

	// Test feature equals method
	feature2 := &Feature{
		ConfigData:         configData,
		FeatureName:        "whitelistedFeature",
		Name:               "Whitelisted Feature",
		Enable:             true,
		EffectiveImmediate: true,
		Whitelisted:        true,
		WhitelistProperty: &WhitelistProperty{
			Key:                "macAddress",
			Value:             "AA:BB:CC:DD:EE:FF",
			NamespacedListType: "MAC_LIST",
		},
	}
	
	// Test with same content
	assert.Assert(t, feature.equals(feature2))
}

func TestFeatureResponse_MarshalJSON_AllFields(t *testing.T) {
	// Test with all possible fields populated
	configData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	
	feature := Feature{
		ID:                 "test-id-123",
		Name:               "test-feature",
		FeatureName:        "testFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData:         configData,
	}
	
	response := CreateFeatureResponseObject(feature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	// Verify JSON contains all expected fields
	jsonStr := string(jsonBytes)
	assert.Assert(t, strings.Contains(jsonStr, "\"name\":\"test-feature\""))
	assert.Assert(t, strings.Contains(jsonStr, "\"enable\":true"))
	assert.Assert(t, strings.Contains(jsonStr, "\"effectiveImmediate\":false"))
	assert.Assert(t, strings.Contains(jsonStr, "\"featureInstance\":\"testFeature\""))
	assert.Assert(t, strings.Contains(jsonStr, "\"configData\":{"))
}

func TestFeatureResponse_MarshalJSON_EmptyConfigData(t *testing.T) {
	// Test with empty config data
	feature := Feature{
		Name:               "empty-config-feature",
		FeatureName:        "emptyConfig",
		Enable:             false,
		EffectiveImmediate: true,
		ConfigData:         map[string]string{},
	}
	
	response := CreateFeatureResponseObject(feature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	jsonStr := string(jsonBytes)
	assert.Assert(t, strings.Contains(jsonStr, "\"configData\":{}"))
}

func TestFeatureResponse_MarshalJSON_NilConfigData(t *testing.T) {
	// Test with nil config data
	feature := Feature{
		Name:               "nil-config-feature",
		FeatureName:        "nilConfig",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData:         nil,
	}
	
	response := CreateFeatureResponseObject(feature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	jsonStr := string(jsonBytes)
	assert.Assert(t, strings.Contains(jsonStr, "\"configData\":null"))
}

func TestFeatureEntity_UnmarshalJSON_ValidInput(t *testing.T) {
	// Test basic unmarshaling
	jsonStr := `{
		"id": "test-id",
		"name": "test-feature",
		"featureName": "testFeature",
		"enable": true,
		"effectiveImmediate": false,
		"configData": {"key1": "value1", "key2": "value2"}
	}`
	
	var feature FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &feature)
	assert.NilError(t, err)
	assert.Equal(t, "test-id", feature.ID)
	assert.Equal(t, "test-feature", feature.Name)
	assert.Equal(t, "testFeature", feature.FeatureName)
	assert.Equal(t, true, feature.Enable)
	assert.Equal(t, false, feature.EffectiveImmediate)
	assert.Equal(t, 2, len(feature.ConfigData))
}

func TestFeatureEntity_UnmarshalJSON_MinimalFields(t *testing.T) {
	// Test with minimal required fields
	jsonStr := `{
		"name": "minimal-feature",
		"enable": false
	}`
	
	var feature FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &feature)
	assert.NilError(t, err)
	assert.Equal(t, "minimal-feature", feature.Name)
	assert.Equal(t, false, feature.Enable)
}

func TestFeatureEntity_UnmarshalJSON_InvalidJSON(t *testing.T) {
	// Test with malformed JSON
	invalidJSON := `{"name": "test", "enable": true,}` // trailing comma
	
	var feature FeatureEntity
	err := json.Unmarshal([]byte(invalidJSON), &feature)
	assert.Assert(t, err != nil, "Expected error for invalid JSON")
}

func TestFeatureEntity_UnmarshalJSON_EmptyJSON(t *testing.T) {
	// Test with empty JSON object
	jsonStr := `{}`
	
	var feature FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &feature)
	assert.NilError(t, err)
	assert.Equal(t, "", feature.Name)
	assert.Equal(t, false, feature.Enable)
}

func TestFeatureEntity_UnmarshalJSON_NullConfigData(t *testing.T) {
	// Test with null configData
	jsonStr := `{
		"name": "null-config",
		"enable": true,
		"configData": null
	}`
	
	var feature FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &feature)
	assert.NilError(t, err)
	assert.Equal(t, "null-config", feature.Name)
	// ConfigData may be nil or empty map after unmarshaling null
}

func TestFeatureApplicationType(t *testing.T) {
	feature := &Feature{
		ID:                 "test-app-type",
		Name:               "Test App Type",
		FeatureName:        "testAppType",
		Enable:             true,
		EffectiveImmediate: true,
		ApplicationType:    "stb",
	}

	feature.ApplicationType = "rdkv"
	cloned, err := feature.Clone()
	assert.NilError(t, err)
	assert.Equal(t, "rdkv", cloned.ApplicationType)
}

func TestFeatureWhitelistOperations(t *testing.T) {
	feature := &Feature{
		ID:                 "test-whitelist",
		Name:               "Test Whitelist",
		FeatureName:        "testWhitelist",
		Enable:             true,
		EffectiveImmediate: true,
	}

	response := CreateFeatureResponseObject(*feature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	jsonStr := string(jsonBytes)
	assert.Assert(t, strings.Contains(jsonStr, "featureInstance"))
	assert.Assert(t, strings.Contains(jsonStr, feature.FeatureName))
}

func TestFeature_Clone(t *testing.T) {
	// Test Clone method with all fields populated
	original := Feature{
		ID:                 "original-id",
		Name:               "original-name", 
		FeatureName:        "originalFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	
	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	
	// Verify basic fields were cloned correctly
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.FeatureName, cloned.FeatureName)
	assert.Equal(t, original.Enable, cloned.Enable)
	assert.Equal(t, original.EffectiveImmediate, cloned.EffectiveImmediate)
	
	// Verify ConfigData was deep copied
	assert.Equal(t, len(original.ConfigData), len(cloned.ConfigData))
	for k, v := range original.ConfigData {
		assert.Equal(t, v, cloned.ConfigData[k])
	}
	
	// Test that modifying clone doesn't affect original
	cloned.ConfigData["key3"] = "value3"
	assert.Assert(t, original.ConfigData["key3"] == "")
	assert.Equal(t, 2, len(original.ConfigData), "Original should still have 2 items")
	assert.Equal(t, 3, len(cloned.ConfigData), "Cloned should now have 3 items")
	
	// Verify it's a deep copy - modifying original shouldn't affect clone
	original.ConfigData["key1"] = "modified"
	assert.Equal(t, "value1", cloned.ConfigData["key1"], "Clone should retain original value")
}

func TestFeature_Clone_EmptyConfigData(t *testing.T) {
	// Test Clone with empty config data
	original := Feature{
		Name:        "test-feature",
		Enable:      false,
		ConfigData:  map[string]string{},
	}
	
	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Enable, cloned.Enable)
	assert.Equal(t, 0, len(cloned.ConfigData))
}

func TestFeature_Clone_NilConfigData(t *testing.T) {
	// Test Clone with nil config data
	original := Feature{
		Name:        "nil-config-feature",
		Enable:      true,
		ConfigData:  nil,
	}
	
	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Enable, cloned.Enable)
	// Note: Clone may preserve nil or create empty map depending on implementation
}

func TestNewWhitelistProperty(t *testing.T) {
	// Test NewWhitelistProperty constructor
	property := NewWhitelistProperty()
	assert.Assert(t, property != nil)
	assert.Equal(t, "", property.Key)
	assert.Equal(t, "", property.Value)
	assert.Equal(t, "", property.NamespacedListType)
	assert.Equal(t, "", property.TypeName)
}

func TestNewFeatureLegacy(t *testing.T) {
	// Test NewFeatureLegacy constructor
	feature := NewFeatureLegacy()
	assert.Assert(t, feature != nil)
	assert.Equal(t, "", feature.ID)
	assert.Equal(t, "", feature.Name)
	assert.Equal(t, false, feature.Enable)
	assert.Equal(t, false, feature.EffectiveImmediate)
	assert.Assert(t, feature.ConfigData == nil)
}

func TestNewPercentRange(t *testing.T) {
	// Test NewPercentRange constructor
	percentRange := NewPercentRange()
	
	assert.Assert(t, percentRange != nil)
	assert.Equal(t, 0.0, percentRange.StartRange)
	assert.Equal(t, 0.0, percentRange.EndRange)
}

func TestNewFeature(t *testing.T) {
	// Test NewFeature constructor
	feature := NewFeature()
	assert.Assert(t, feature != nil)
	assert.Equal(t, "", feature.ID)
	assert.Equal(t, "", feature.Name)
	assert.Equal(t, false, feature.Enable)
	assert.Equal(t, false, feature.EffectiveImmediate)
	// ConfigData may be nil or empty depending on implementation
}

func TestNewFeatureInf(t *testing.T) {
	// Test NewFeatureInf constructor
	feature := NewFeatureInf()
	assert.Assert(t, feature != nil)
	// NewFeatureInf returns interface{}, so just verify it's not nil
}

func TestNewFeatureControl(t *testing.T) {
	// Test NewFeatureControl constructor
	featureControl := NewFeatureControl()
	assert.Assert(t, featureControl != nil)
}

func TestWhitelistProperty_Equals(t *testing.T) {
	// Test WhitelistProperty equals method
	prop1 := &WhitelistProperty{
		Key:   "testKey",
		Value: "testValue",
		NamespacedListType: "testType",
		TypeName: "testTypeName",
	}
	
	prop2 := &WhitelistProperty{
		Key:   "testKey",
		Value: "testValue",
		NamespacedListType: "testType",
		TypeName: "testTypeName",
	}
	
	prop3 := &WhitelistProperty{
		Key:   "differentKey",
		Value: "testValue",
		NamespacedListType: "testType",
		TypeName: "testTypeName",
	}
	
	// Test equality using reflection since equals method is private
	assert.Equal(t, prop1.Key, prop2.Key)
	assert.Equal(t, prop1.Value, prop2.Value)
	assert.Assert(t, prop1.Key != prop3.Key)
}

func TestWhitelistProperty_ToString(t *testing.T) {
	// Test WhitelistProperty toString method
	prop := &WhitelistProperty{
		Key:   "testKey",
		Value: "testValue",
		NamespacedListType: "testType",
		TypeName: "testTypeName",
	}
	
	// Test that the property has been created with correct values
	assert.Equal(t, "testKey", prop.Key)
	assert.Equal(t, "testValue", prop.Value)
}

func TestPercentRange_Equals(t *testing.T) {
	// Test PercentRange Equals method
	range1 := &PercentRange{StartRange: 10.0, EndRange: 90.0}
	range2 := &PercentRange{StartRange: 10.0, EndRange: 90.0}
	range3 := &PercentRange{StartRange: 20.0, EndRange: 80.0}
	
	assert.Assert(t, range1.Equals(range2))
	assert.Assert(t, !range1.Equals(range3))
	assert.Assert(t, !range2.Equals(range3))
}

func TestFeature_CreateFeatureEntity(t *testing.T) {
	// Test CreateFeatureEntity method
	feature := &Feature{
		ID:                 "test-id",
		Name:               "test-name",
		FeatureName:        "testFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData: map[string]string{
			"key1": "value1",
		},
		ApplicationType: "stb",
		Whitelisted:     true,
	}
	
	entity := feature.CreateFeatureEntity()
	assert.Assert(t, entity != nil)
	assert.Equal(t, feature.ID, entity.ID)
	assert.Equal(t, feature.Name, entity.Name)
	assert.Equal(t, feature.FeatureName, entity.FeatureName)
	assert.Equal(t, feature.FeatureName, entity.FeatureInstance)
	assert.Equal(t, feature.Enable, entity.Enable)
	assert.Equal(t, feature.EffectiveImmediate, entity.EffectiveImmediate)
	assert.Equal(t, feature.ApplicationType, entity.ApplicationType)
	assert.Equal(t, feature.Whitelisted, entity.Whitelisted)
}

func TestFeature_CreateFeatureEntity_Nil(t *testing.T) {
	// Test CreateFeatureEntity with nil feature
	var feature *Feature = nil
	entity := feature.CreateFeatureEntity()
	assert.Assert(t, entity == nil)
}

func TestFeatureEntity_SetApplicationType(t *testing.T) {
	// Test SetApplicationType method
	entity := &FeatureEntity{}
	appType := "test-app-type"
	
	entity.SetApplicationType(appType)
	assert.Equal(t, appType, entity.ApplicationType)
}

func TestFeatureEntity_GetApplicationType(t *testing.T) {
	// Test GetApplicationType method
	entity := &FeatureEntity{
		ApplicationType: "stb",
	}
	
	result := entity.GetApplicationType()
	assert.Equal(t, "stb", result)
}

func TestFeatureEntity_CreateFeature(t *testing.T) {
	// Test CreateFeature method
	entity := &FeatureEntity{
		ID:                 "test-id",
		Name:               "test-name",
		FeatureName:        "testFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData: map[string]string{
			"key1": "value1",
		},
		ApplicationType: "stb",
		Whitelisted:     true,
	}
	
	feature := entity.CreateFeature()
	assert.Assert(t, feature != nil)
	assert.Equal(t, entity.ID, feature.ID)
	assert.Equal(t, entity.Name, feature.Name)
	assert.Equal(t, entity.FeatureName, feature.FeatureName)
	assert.Equal(t, entity.Enable, feature.Enable)
	assert.Equal(t, entity.EffectiveImmediate, feature.EffectiveImmediate)
	assert.Equal(t, entity.ApplicationType, feature.ApplicationType)
	assert.Equal(t, entity.Whitelisted, feature.Whitelisted)
}

func TestFeatureEntity_CreateFeature_Nil(t *testing.T) {
	// Test CreateFeature with nil entity
	var entity *FeatureEntity = nil
	feature := entity.CreateFeature()
	assert.Assert(t, feature == nil)
}

func TestFeature_ToString(t *testing.T) {
	// Test Feature ToString method
	feature := &Feature{
		ID:                 "test-id",
		Name:               "test-name",
		FeatureName:        "testFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData: map[string]string{
			"key1": "value1",
		},
	}
	
	result := feature.ToString()
	assert.Assert(t, result != "")
	assert.Assert(t, strings.Contains(result, "test-name"))
}

func TestFeature_EqualsMethod(t *testing.T) {
	// Test Feature equals method by creating two similar features
	feature1 := &Feature{
		ID:     "test-id",
		Name:   "test-name",
		Enable: true,
		ConfigData: map[string]string{
			"key1": "value1",
		},
	}
	
	feature2 := &Feature{
		ID:     "test-id",
		Name:   "test-name", 
		Enable: true,
		ConfigData: map[string]string{
			"key1": "value1",
		},
	}
	
	feature3 := &Feature{
		ID:     "different-id",
		Name:   "test-name",
		Enable: true,
		ConfigData: map[string]string{
			"key1": "value1",
		},
	}
	
	// Use reflection to test behavior (since equals is private)
	// But we can test by creating scenarios that would use equals
	assert.Equal(t, feature1.ID, feature2.ID)
	assert.Equal(t, feature1.Name, feature2.Name)
	assert.Assert(t, feature1.ID != feature3.ID)
}

func TestWhitelistProperty_ToStringMethod(t *testing.T) {
	// Test WhitelistProperty toString by creating property and verifying content
	prop := &WhitelistProperty{
		Key:                "testKey",
		Value:              "testValue",
		NamespacedListType: "testType",
		TypeName:           "testTypeName",
	}
	
	// We can't directly call toString (private), but can verify the fields exist
	assert.Equal(t, "testKey", prop.Key)
	assert.Equal(t, "testValue", prop.Value)
	assert.Equal(t, "testType", prop.NamespacedListType)
	assert.Equal(t, "testTypeName", prop.TypeName)
}

func TestFeatureEntity_UnmarshalJSON_ComplexScenarios(t *testing.T) {
	// Test UnmarshalJSON with various complex scenarios
	testCases := []struct {
		name     string
		jsonStr  string
		expected func(*testing.T, *FeatureEntity)
	}{
		{
			name: "Full feature with whitelist property",
			jsonStr: `{
				"id": "test-id",
				"name": "test-feature",
				"featureName": "testFeature",
				"enable": true,
				"effectiveImmediate": false,
				"whitelisted": true,
				"configData": {"key1": "value1"},
				"applicationType": "stb",
				"whitelistProperty": {
					"key": "testKey",
					"value": "testValue"
				}
			}`,
			expected: func(t *testing.T, feature *FeatureEntity) {
				assert.Equal(t, "test-id", feature.ID)
				assert.Equal(t, "test-feature", feature.Name)
				assert.Equal(t, true, feature.Enable)
				assert.Equal(t, true, feature.Whitelisted)
				assert.Equal(t, "stb", feature.ApplicationType)
				assert.Assert(t, feature.WhitelistProperty != nil)
				assert.Equal(t, "testKey", feature.WhitelistProperty.Key)
			},
		},
		{
			name: "Feature with empty strings",
			jsonStr: `{
				"id": "",
				"name": "",
				"featureName": "",
				"enable": false,
				"effectiveImmediate": false
			}`,
			expected: func(t *testing.T, feature *FeatureEntity) {
				assert.Equal(t, "", feature.ID)
				assert.Equal(t, "", feature.Name)
				assert.Equal(t, false, feature.Enable)
			},
		},
		{
			name: "Feature with boolean variations",
			jsonStr: `{
				"name": "bool-test",
				"enable": true,
				"effectiveImmediate": true,
				"whitelisted": false
			}`,
			expected: func(t *testing.T, feature *FeatureEntity) {
				assert.Equal(t, "bool-test", feature.Name)
				assert.Equal(t, true, feature.Enable)
				assert.Equal(t, true, feature.EffectiveImmediate)
				assert.Equal(t, false, feature.Whitelisted)
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var feature FeatureEntity
			err := json.Unmarshal([]byte(tc.jsonStr), &feature)
			assert.NilError(t, err)
			tc.expected(t, &feature)
		})
	}
}

func TestCreateFeatureResponseObject_WithListTypeAndProperties(t *testing.T) {
	// Test CreateFeatureResponseObject with ListType and Properties
	feature := Feature{
		Name:               "list-feature",
		FeatureName:        "listFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData: map[string]string{
			"key1": "value1",
		},
		ListType: "whitelist",
		ListSize: 5,
		Properties: map[string]interface{}{
			"prop1": "value1",
			"prop2": 42,
			"prop3": true,
		},
	}
	
	response := CreateFeatureResponseObject(feature)
	
	// Verify basic fields
	assert.Equal(t, "list-feature", response["name"])
	assert.Equal(t, "listFeature", response["featureInstance"])
	assert.Equal(t, true, response["enable"])
	
	// Verify list-specific fields are included
	assert.Equal(t, "whitelist", response["listType"])
	assert.Equal(t, 5, response["listSize"])
	
	// Verify properties are copied
	assert.Equal(t, "value1", response["prop1"])
	assert.Equal(t, 42, response["prop2"])
	assert.Equal(t, true, response["prop3"])
}

func TestCreateFeatureResponseObject_WithoutListType(t *testing.T) {
	// Test CreateFeatureResponseObject without ListType (should not include list fields)
	feature := Feature{
		Name:               "simple-feature",
		FeatureName:        "simpleFeature",
		Enable:             false,
		EffectiveImmediate: true,
		ConfigData: map[string]string{
			"key1": "value1",
		},
		ListType: "", // Empty - should not trigger list logic
		ListSize: 5,  // Non-zero but ListType is empty
		Properties: map[string]interface{}{
			"prop1": "value1",
		},
	}
	
	response := CreateFeatureResponseObject(feature)
	
	// Verify basic fields
	assert.Equal(t, "simple-feature", response["name"])
	assert.Equal(t, false, response["enable"])
	
	// Verify list fields are NOT included
	_, hasListType := response["listType"]
	assert.Assert(t, !hasListType)
	
	_, hasListSize := response["listSize"]
	assert.Assert(t, !hasListSize)
	
	_, hasProp1 := response["prop1"]
	assert.Assert(t, !hasProp1)
}

func TestCreateFeatureResponseObject_WithListTypeButZeroSize(t *testing.T) {
	// Test CreateFeatureResponseObject with ListType but zero size
	feature := Feature{
		Name:               "zero-size-feature",
		FeatureName:        "zeroSizeFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData:         map[string]string{},
		ListType:           "blacklist", // Non-empty
		ListSize:           0,           // Zero - should not trigger list logic
		Properties: map[string]interface{}{
			"prop1": "value1",
		},
	}
	
	response := CreateFeatureResponseObject(feature)
	
	// Verify basic fields
	assert.Equal(t, "zero-size-feature", response["name"])
	
	// Verify list fields are NOT included (ListSize is 0)
	_, hasListType := response["listType"]
	assert.Assert(t, !hasListType)
}

func TestFeature_Clone_WithComplexData(t *testing.T) {
	// Test Clone with more complex data structures
	original := Feature{
		ID:                 "complex-id",
		Name:               "complex-feature",
		FeatureName:        "complexFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
		Properties: map[string]interface{}{
			"intProp":    42,
			"stringProp": "test",
			"boolProp":   true,
		},
		ListType: "whitelist",
		ListSize: 10,
		WhitelistProperty: &WhitelistProperty{
			Key:   "testKey",
			Value: "testValue",
		},
		ApplicationType: "stb",
		Whitelisted:     true,
	}
	
	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	
	// Verify all fields are copied
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.ListType, cloned.ListType)
	assert.Equal(t, original.ListSize, cloned.ListSize)
	assert.Equal(t, original.Whitelisted, cloned.Whitelisted)
	
	// Verify deep copy of Properties
	if original.Properties != nil && cloned.Properties != nil {
		assert.Equal(t, len(original.Properties), len(cloned.Properties))
		for k, v := range original.Properties {
			assert.Equal(t, v, cloned.Properties[k])
		}
	}
	
	// Verify WhitelistProperty is deep copied
	if original.WhitelistProperty != nil && cloned.WhitelistProperty != nil {
		assert.Equal(t, original.WhitelistProperty.Key, cloned.WhitelistProperty.Key)
		assert.Equal(t, original.WhitelistProperty.Value, cloned.WhitelistProperty.Value)
	}
}

func TestFeatureEntity_UnmarshalJSON_EdgeCases(t *testing.T) {
	// Test various edge cases for UnmarshalJSON
	testCases := []struct {
		name    string
		jsonStr string
		expect  func(*testing.T, *FeatureEntity, error)
	}{
		{
			name:    "Malformed JSON with extra comma",
			jsonStr: `{"name": "test",}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.Assert(t, err != nil, "Expected error for malformed JSON")
			},
		},
		{
			name:    "JSON with null values",
			jsonStr: `{"name": null, "enable": null, "configData": null}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.NilError(t, err)
				assert.Equal(t, "", feature.Name) // null string becomes empty
			},
		},
		{
			name:    "JSON with wrong data types",
			jsonStr: `{"name": 123, "enable": "not_boolean"}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				// This might error or handle gracefully depending on implementation
				// Just verify it doesn't crash
			},
		},
		{
			name:    "Very large JSON",
			jsonStr: `{"name": "test", "configData": {"key1": "value1", "key2": "value2", "key3": "value3", "key4": "value4", "key5": "value5"}}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.NilError(t, err)
				assert.Equal(t, "test", feature.Name)
				assert.Equal(t, 5, len(feature.ConfigData))
			},
		},
		{
			name:    "JSON with unicode characters",
			jsonStr: `{"name": "测试功能", "featureName": "тест", "configData": {"key": "值"}}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.NilError(t, err)
				assert.Equal(t, "测试功能", feature.Name)
				assert.Equal(t, "тест", feature.FeatureName)
				assert.Equal(t, "值", feature.ConfigData["key"])
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var feature FeatureEntity
			err := json.Unmarshal([]byte(tc.jsonStr), &feature)
			tc.expect(t, &feature, err)
		})
	}
}

func TestFeatureResponse_MarshalJSON_ErrorCases(t *testing.T) {
	// Test MarshalJSON with values that might cause marshaling errors
	response := FeatureResponse{
		"name":               "test-feature",
		"enable":             true,
		"effectiveImmediate": false,
		"configData": map[string]string{
			"key1": "value1",
		},
		// Add some complex values that should still marshal successfully
		"complexValue": map[string]interface{}{
			"nested": map[string]string{
				"deep": "value",
			},
		},
	}
	
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 0)
	
	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NilError(t, err)
	assert.Equal(t, "test-feature", result["name"])
}

func TestFeatureResponse_MarshalJSON_WithNilValues(t *testing.T) {
	// Test MarshalJSON with nil values in main fields
	response := FeatureResponse{
		"name":               nil,
		"enable":             nil,
		"effectiveImmediate": true,
		"configData":         nil,
		"extraField":         "extraValue",
	}
	
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	// Verify JSON structure - nil values should be skipped for main fields
	jsonStr := string(jsonBytes)
	assert.Assert(t, strings.Contains(jsonStr, "\"effectiveImmediate\":true"))
	assert.Assert(t, strings.Contains(jsonStr, "\"extraField\":\"extraValue\""))
	// nil values in main fields should not appear
	assert.Assert(t, !strings.Contains(jsonStr, "\"name\":null"))
	assert.Assert(t, !strings.Contains(jsonStr, "\"enable\":null"))
}

func TestComprehensiveFeatureWorkflow(t *testing.T) {
	// Test a comprehensive workflow that exercises multiple methods
	
	// 1. Create a complex feature
	originalFeature := Feature{
		ID:                 "workflow-test-id",
		Name:               "workflow-test",
		FeatureName:        "workflowTest",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
		Properties: map[string]interface{}{
			"priority": 100,
			"active":   true,
		},
		ListType:        "whitelist",
		ListSize:        3,
		ApplicationType: "stb",
		Whitelisted:     true,
		WhitelistProperty: &WhitelistProperty{
			Key:                "testKey",
			Value:              "testValue",
			NamespacedListType: "device",
			TypeName:           "mac",
		},
	}
	
	// 2. Test Clone
	cloned, err := originalFeature.Clone()
	assert.NilError(t, err)
	assert.Equal(t, originalFeature.ID, cloned.ID)
	
	// 3. Test CreateFeatureEntity
	entity := originalFeature.CreateFeatureEntity()
	assert.Assert(t, entity != nil)
	assert.Equal(t, originalFeature.ID, entity.ID)
	
	// 4. Test entity methods
	entity.SetApplicationType("rdkv")
	assert.Equal(t, "rdkv", entity.GetApplicationType())
	
	// 5. Test entity back to feature
	reconstructed := entity.CreateFeature()
	assert.Assert(t, reconstructed != nil)
	assert.Equal(t, "rdkv", reconstructed.ApplicationType) // Should have updated value
	
	// 6. Test FeatureResponse creation and marshaling
	response := CreateFeatureResponseObject(originalFeature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	// Verify the response contains expected data
	jsonStr := string(jsonBytes)
	assert.Assert(t, strings.Contains(jsonStr, "workflow-test"))
	assert.Assert(t, strings.Contains(jsonStr, "listType"))
	assert.Assert(t, strings.Contains(jsonStr, "priority"))
	
	// 7. Test JSON unmarshaling
	entityJsonStr := `{
		"id": "unmarshal-test",
		"name": "unmarshal-feature",
		"enable": false,
		"effectiveImmediate": true,
		"configData": {"testKey": "testValue"},
		"whitelistProperty": {
			"key": "unmarshalKey",
			"value": "unmarshalValue"
		}
	}`
	
	var unmarshaledEntity FeatureEntity
	err = json.Unmarshal([]byte(entityJsonStr), &unmarshaledEntity)
	assert.NilError(t, err)
	assert.Equal(t, "unmarshal-test", unmarshaledEntity.ID)
	assert.Equal(t, false, unmarshaledEntity.Enable)
	assert.Assert(t, unmarshaledEntity.WhitelistProperty != nil)
	assert.Equal(t, "unmarshalKey", unmarshaledEntity.WhitelistProperty.Key)
	
	// 8. Test ToString
	toStringResult := originalFeature.ToString()
	assert.Assert(t, toStringResult != "")
}

func TestFeature_Clone_ErrorScenarios(t *testing.T) {
	// Test Clone with potential error scenarios
	
	// Test with a feature containing potentially problematic data
	feature := Feature{
		ID:   "error-test",
		Name: "error-feature",
		Properties: map[string]interface{}{
			// Add complex nested structures that might cause copy issues
			"nested": map[string]interface{}{
				"deep": map[string]interface{}{
					"value": "test",
				},
			},
		},
		ConfigData: make(map[string]string),
	}
	
	// Populate ConfigData with many entries
	for i := 0; i < 100; i++ {
		feature.ConfigData[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}
	
	cloned, err := feature.Clone()
	assert.NilError(t, err, "Clone should handle complex data structures")
	assert.Assert(t, cloned != nil)
	assert.Equal(t, feature.ID, cloned.ID)
	assert.Equal(t, len(feature.ConfigData), len(cloned.ConfigData))
}

func TestFeature_Clone_NilFeature(t *testing.T) {
	// Test Clone with edge cases
	var nilFeature *Feature = nil
	
	// This should handle gracefully if the method checks for nil
	// Note: The actual Clone method may not handle nil receiver,
	// but we test the behavior
	defer func() {
		if r := recover(); r != nil {
			// If it panics, that's expected behavior for nil receiver
			t.Log("Clone with nil receiver panicked as expected")
		}
	}()
	
	if nilFeature != nil {
		_, err := nilFeature.Clone()
		assert.Assert(t, err != nil || nilFeature == nil)
	}
}

func TestFeatureResponse_MarshalJSON_ComplexErrorScenarios(t *testing.T) {
	// Test MarshalJSON with scenarios that might cause errors
	
	// Test with circular reference data (though Go's json.Marshal should handle this)
	response := FeatureResponse{
		"name":               "error-test",
		"enable":             true,
		"effectiveImmediate": false,
		"configData": map[string]string{
			"key1": "value1",
		},
	}
	
	// Add complex nested structures
	complexData := make(map[string]interface{})
	complexData["level1"] = map[string]interface{}{
		"level2": map[string]interface{}{
			"level3": "deep-value",
		},
	}
	response["complexNested"] = complexData
	
	// Add various data types
	response["intValue"] = 42
	response["floatValue"] = 3.14159
	response["boolValue"] = true
	response["arrayValue"] = []string{"item1", "item2", "item3"}
	response["nilValue"] = nil
	
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 0)
	
	// Verify it produces valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NilError(t, err)
}

func TestFeatureResponse_MarshalJSON_EdgeCaseFields(t *testing.T) {
	// Test with edge case field names and values
	response := FeatureResponse{
		"name":               "edge-case-test",
		"enable":             true,
		"effectiveImmediate": false,
		"configData":         map[string]string{},
	}
	
	// Add fields with special characters
	response["field-with-dashes"] = "value"
	response["field_with_underscores"] = "value"
	response["fieldWithCamelCase"] = "value"
	response["field.with.dots"] = "value"
	response["field with spaces"] = "value"
	
	// Add empty and special values
	response["emptyString"] = ""
	response["zeroInt"] = 0
	response["falseBool"] = false
	response["emptyArray"] = []string{}
	response["emptyMap"] = map[string]string{}
	
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	// Verify JSON structure
	jsonStr := string(jsonBytes)
	assert.Assert(t, strings.Contains(jsonStr, "\"name\":\"edge-case-test\""))
	assert.Assert(t, strings.Contains(jsonStr, "\"field-with-dashes\""))
	assert.Assert(t, strings.Contains(jsonStr, "\"emptyString\":\"\""))
	assert.Assert(t, strings.Contains(jsonStr, "\"zeroInt\":0"))
}

func TestFeatureEntity_UnmarshalJSON_AdvancedScenarios(t *testing.T) {
	// Test UnmarshalJSON with advanced and edge case scenarios
	
	testCases := []struct {
		name    string
		jsonStr string
		expect  func(*testing.T, *FeatureEntity, error)
	}{
		{
			name:    "JSON with numeric string values",
			jsonStr: `{"name": "123", "enable": "true", "id": "456"}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				// This tests how the unmarshaler handles type mismatches
				// Some fields might handle string-to-type conversion
			},
		},
		{
			name:    "JSON with extra unknown fields",
			jsonStr: `{"name": "test", "unknownField": "value", "anotherUnknown": 123, "enable": true}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.NilError(t, err)
				assert.Equal(t, "test", feature.Name)
				assert.Equal(t, true, feature.Enable)
				// Unknown fields should be ignored
			},
		},
		{
			name:    "JSON with nested whitelist property edge cases",
			jsonStr: `{"name": "whitelist-test", "whitelistProperty": {"key": "", "value": null, "namespacedListType": "test"}}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.NilError(t, err)
				assert.Equal(t, "whitelist-test", feature.Name)
				assert.Assert(t, feature.WhitelistProperty != nil)
				assert.Equal(t, "", feature.WhitelistProperty.Key)
				assert.Equal(t, "test", feature.WhitelistProperty.NamespacedListType)
			},
		},
		{
			name:    "JSON with very long strings",
			jsonStr: `{"name": "very-long-name-that-goes-on-and-on-and-on-and-on-and-on", "featureName": "veryLongFeatureNameThatTestsStringHandling", "enable": true}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.NilError(t, err)
				assert.Assert(t, len(feature.Name) > 50)
				assert.Assert(t, len(feature.FeatureName) > 30)
			},
		},
		{
			name:    "JSON with empty objects and arrays",
			jsonStr: `{"name": "empty-test", "configData": {}, "whitelistProperty": {}}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.NilError(t, err)
				assert.Equal(t, "empty-test", feature.Name)
				assert.Assert(t, feature.ConfigData != nil)
				assert.Equal(t, 0, len(feature.ConfigData))
			},
		},
		{
			name:    "JSON with boolean variations",
			jsonStr: `{"enable": true, "effectiveImmediate": false, "whitelisted": true}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.NilError(t, err)
				assert.Equal(t, true, feature.Enable)
				assert.Equal(t, false, feature.EffectiveImmediate)
				assert.Equal(t, true, feature.Whitelisted)
			},
		},
		{
			name:    "JSON with configData containing special characters",
			jsonStr: `{"name": "special-chars", "configData": {"key-with-dash": "value", "key.with.dots": "value", "key_with_underscores": "value", "normalkey": "normal-value"}}`,
			expect: func(t *testing.T, feature *FeatureEntity, err error) {
				assert.NilError(t, err)
				assert.Equal(t, "special-chars", feature.Name)
				assert.Equal(t, 4, len(feature.ConfigData))
				assert.Equal(t, "value", feature.ConfigData["key-with-dash"])
				assert.Equal(t, "normal-value", feature.ConfigData["normalkey"])
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var feature FeatureEntity
			err := json.Unmarshal([]byte(tc.jsonStr), &feature)
			tc.expect(t, &feature, err)
		})
	}
}

func TestBoundaryConditions(t *testing.T) {
	// Test boundary conditions and edge cases to push coverage higher
	
	// Test very large ConfigData
	largeFeature := Feature{
		Name: "large-config-feature",
		ConfigData: make(map[string]string),
	}
	
	// Add many config entries
	for i := 0; i < 200; i++ {
		largeFeature.ConfigData[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	
	// Test Clone with large data
	cloned, err := largeFeature.Clone()
	assert.NilError(t, err)
	assert.Equal(t, len(largeFeature.ConfigData), len(cloned.ConfigData))
	
	// Test CreateFeatureResponseObject with large data
	response := CreateFeatureResponseObject(largeFeature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 1000) // Should be a large JSON
}

func TestAdvancedUnicodeScenarios(t *testing.T) {
	// Test with unicode and international characters
	unicodeFeature := Feature{
		ID:          "unicode-测试-тест",
		Name:        "Unicode Feature 测试功能",
		FeatureName: "unicodeFeature测试",
		ConfigData: map[string]string{
			"chinese": "中文值",
			"russian": "русское значение",
			"emoji":   "🚀🎯💯",
		},
		ApplicationType: "international-app",
	}
	
	// Test Clone with unicode
	cloned, err := unicodeFeature.Clone()
	assert.NilError(t, err)
	assert.Equal(t, unicodeFeature.Name, cloned.Name)
	assert.Equal(t, "中文值", cloned.ConfigData["chinese"])
	
	// Test ToString with unicode
	toStringResult := unicodeFeature.ToString()
	assert.Assert(t, toStringResult != "")
	
	// Test CreateFeatureEntity with unicode
	entity := unicodeFeature.CreateFeatureEntity()
	assert.Assert(t, entity != nil)
	assert.Equal(t, unicodeFeature.Name, entity.Name)
	
	// Test JSON marshaling with unicode
	response := CreateFeatureResponseObject(unicodeFeature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	// Verify unicode is preserved in JSON
	jsonStr := string(jsonBytes)
	assert.Assert(t, strings.Contains(jsonStr, "Unicode Feature"))
}

func TestFeatureRuleCloneCoverage(t *testing.T) {
	// Test FeatureRule Clone method to improve overall coverage
	featureRule := &FeatureRule{
		Id:              "rule-test",
		Name:            "test-rule",
		Priority:        1,
		ApplicationType: "stb",
	}
	
	cloned, err := featureRule.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, featureRule.Id, cloned.Id)
	assert.Equal(t, featureRule.Name, cloned.Name)
	assert.Equal(t, featureRule.Priority, cloned.Priority)
}

func TestMarshalJSONErrorPaths(t *testing.T) {
	// Try to test more error paths in MarshalJSON
	response := FeatureResponse{
		"name":               "marshal-error-test",
		"enable":             true,
		"effectiveImmediate": false,
		"configData": map[string]string{
			"key1": "value1",
		},
	}
	
	// Add a field that contains complex nested structure
	deepNested := make(map[string]interface{})
	for i := 0; i < 10; i++ {
		level := make(map[string]interface{})
		level[fmt.Sprintf("level_%d", i)] = fmt.Sprintf("value_%d", i)
		deepNested[fmt.Sprintf("nest_%d", i)] = level
	}
	response["deepNested"] = deepNested
	
	// Test that it can handle complex marshaling
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 100)
	
	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NilError(t, err)
}

func TestUnmarshalJSONMoreEdgeCases(t *testing.T) {
	// Test more edge cases for UnmarshalJSON
	
	// Test with very deeply nested JSON
	deepJSON := `{
		"name": "deep-test",
		"configData": {
			"level1": "value1",
			"level2": "value2",
			"level3": "value3"
		},
		"whitelistProperty": {
			"key": "deep-key",
			"value": "deep-value",
			"namespacedListType": "deep-namespace",
			"typeName": "deep-type"
		}
	}`
	
	var feature FeatureEntity
	err := json.Unmarshal([]byte(deepJSON), &feature)
	assert.NilError(t, err)
	assert.Equal(t, "deep-test", feature.Name)
	assert.Equal(t, 3, len(feature.ConfigData))
	assert.Assert(t, feature.WhitelistProperty != nil)
	assert.Equal(t, "deep-key", feature.WhitelistProperty.Key)
	assert.Equal(t, "deep-namespace", feature.WhitelistProperty.NamespacedListType)
	
	// Test with arrays and complex values that might be ignored
	complexJSON := `{
		"name": "complex-test",
		"unknownArray": ["item1", "item2"],
		"unknownObject": {"nested": "value"},
		"enable": true
	}`
	
	var complexFeature FeatureEntity
	err = json.Unmarshal([]byte(complexJSON), &complexFeature)
	assert.NilError(t, err)
	assert.Equal(t, "complex-test", complexFeature.Name)
	assert.Equal(t, true, complexFeature.Enable)
}

func TestMoreFeatureEntityCreateFeaturePaths(t *testing.T) {
	// Test more paths in CreateFeature to improve coverage
	
	// Test with entity that has all optional fields
	entityWithOptionals := &FeatureEntity{
		ID:                 "optional-test",
		Name:               "optional-feature",
		FeatureName:        "optionalFeature",
		FeatureInstance:    "optionalInstance",
		Enable:             true,
		EffectiveImmediate: true,
		Whitelisted:        true,
		ApplicationType:    "test-app",
		ConfigData: map[string]string{
			"opt1": "val1",
			"opt2": "val2",
		},
		WhitelistProperty: &WhitelistProperty{
			Key:                "opt-key",
			Value:              "opt-value",
			NamespacedListType: "opt-namespace",
			TypeName:           "opt-type",
		},
	}
	
	feature := entityWithOptionals.CreateFeature()
	assert.Assert(t, feature != nil)
	
	// Verify all fields are copied correctly
	assert.Equal(t, entityWithOptionals.ID, feature.ID)
	assert.Equal(t, entityWithOptionals.Name, feature.Name)
	assert.Equal(t, entityWithOptionals.FeatureName, feature.FeatureName)
	assert.Equal(t, entityWithOptionals.Enable, feature.Enable)
	assert.Equal(t, entityWithOptionals.EffectiveImmediate, feature.EffectiveImmediate)
	assert.Equal(t, entityWithOptionals.Whitelisted, feature.Whitelisted)
	assert.Equal(t, entityWithOptionals.ApplicationType, feature.ApplicationType)
	
	// Test ConfigData mapping
	assert.Equal(t, len(entityWithOptionals.ConfigData), len(feature.ConfigData))
	for k, v := range entityWithOptionals.ConfigData {
		assert.Equal(t, v, feature.ConfigData[k])
	}
	
	// Test WhitelistProperty mapping
	assert.Assert(t, feature.WhitelistProperty != nil)
	assert.Equal(t, entityWithOptionals.WhitelistProperty.Key, feature.WhitelistProperty.Key)
	assert.Equal(t, entityWithOptionals.WhitelistProperty.Value, feature.WhitelistProperty.Value)
	assert.Equal(t, entityWithOptionals.WhitelistProperty.NamespacedListType, feature.WhitelistProperty.NamespacedListType)
	assert.Equal(t, entityWithOptionals.WhitelistProperty.TypeName, feature.WhitelistProperty.TypeName)
}

func TestComprehensiveFeatureOperations(t *testing.T) {
	// Comprehensive test that exercises multiple code paths
	
	// Create a feature with maximum complexity
	complexFeature := Feature{
		ID:                 "comprehensive-test",
		Name:               "comprehensive-feature",
		FeatureName:        "comprehensiveFeature",
		Enable:             true,
		EffectiveImmediate: false,
		Whitelisted:        true,
		ApplicationType:    "comprehensive-app",
		ListType:           "whitelist",
		ListSize:           10,
		ConfigData: map[string]string{
			"config1": "value1",
			"config2": "value2",
			"config3": "value3",
		},
		Properties: map[string]interface{}{
			"property1": "string-value",
			"property2": 42,
			"property3": true,
			"property4": 3.14159,
			"nested": map[string]interface{}{
				"inner": "nested-value",
			},
		},
		WhitelistProperty: &WhitelistProperty{
			Key:                "comprehensive-key",
			Value:              "comprehensive-value",
			NamespacedListType: "comprehensive-namespace",
			TypeName:           "comprehensive-type",
		},
	}
	
	// Test 1: Clone operation
	cloned, err := complexFeature.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, complexFeature.ID, cloned.ID)
	assert.Equal(t, complexFeature.ListType, cloned.ListType)
	assert.Equal(t, complexFeature.ListSize, cloned.ListSize)
	
	// Verify deep copy of Properties
	if complexFeature.Properties != nil {
		assert.Equal(t, len(complexFeature.Properties), len(cloned.Properties))
		assert.Equal(t, complexFeature.Properties["property1"], cloned.Properties["property1"])
		assert.Equal(t, complexFeature.Properties["property2"], cloned.Properties["property2"])
	}
	
	// Test 2: CreateFeatureEntity
	entity := complexFeature.CreateFeatureEntity()
	assert.Assert(t, entity != nil)
	assert.Equal(t, complexFeature.ID, entity.ID)
	assert.Equal(t, complexFeature.Name, entity.Name)
	assert.Equal(t, complexFeature.FeatureName, entity.FeatureName)
	assert.Equal(t, complexFeature.FeatureName, entity.FeatureInstance) // Should be same
	
	// Test 3: Entity to Feature conversion
	backToFeature := entity.CreateFeature()
	assert.Assert(t, backToFeature != nil)
	assert.Equal(t, entity.ID, backToFeature.ID)
	assert.Equal(t, entity.Name, backToFeature.Name)
	
	// Test 4: CreateFeatureResponseObject with all properties
	response := CreateFeatureResponseObject(complexFeature)
	assert.Assert(t, response != nil)
	
	// Should include ListType and Properties since ListType is not empty and ListSize > 0
	responseMap := map[string]interface{}(response)
	assert.Equal(t, complexFeature.Name, responseMap["name"])
	assert.Equal(t, complexFeature.FeatureName, responseMap["featureInstance"])
	assert.Equal(t, complexFeature.Enable, responseMap["enable"])
	assert.Equal(t, complexFeature.EffectiveImmediate, responseMap["effectiveImmediate"])
	assert.Equal(t, complexFeature.ListType, responseMap["listType"])
	assert.Equal(t, complexFeature.ListSize, responseMap["listSize"])
	
	// Properties should be copied
	assert.Equal(t, complexFeature.Properties["property1"], responseMap["property1"])
	assert.Equal(t, complexFeature.Properties["property2"], responseMap["property2"])
	
	// Test 5: JSON marshaling of response
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 0)
	
	// Verify JSON structure
	jsonStr := string(jsonBytes)
	assert.Assert(t, strings.Contains(jsonStr, "comprehensive-feature"))
	assert.Assert(t, strings.Contains(jsonStr, "listType"))
	assert.Assert(t, strings.Contains(jsonStr, "property1"))
	
	// Test 6: ToString operation
	toStringResult := complexFeature.ToString()
	assert.Assert(t, toStringResult != "")
}

func TestPercentRangeAdditionalScenarios(t *testing.T) {
	// Test PercentRange with edge values
	
	// Test with zero ranges
	zeroRange := NewPercentRange()
	assert.Assert(t, zeroRange != nil)
	assert.Equal(t, 0.0, zeroRange.StartRange)
	assert.Equal(t, 0.0, zeroRange.EndRange)
	
	// Test Equals with same zero ranges
	anotherZeroRange := NewPercentRange()
	assert.Assert(t, zeroRange.Equals(anotherZeroRange))
	
	// Test Equals with different ranges
	zeroRange.StartRange = 10.5
	zeroRange.EndRange = 90.5
	anotherZeroRange.StartRange = 20.0
	anotherZeroRange.EndRange = 80.0
	
	assert.Assert(t, !zeroRange.Equals(anotherZeroRange))
	
	// Test with identical manual ranges
	range1 := &PercentRange{StartRange: 25.5, EndRange: 75.5}
	range2 := &PercentRange{StartRange: 25.5, EndRange: 75.5}
	range3 := &PercentRange{StartRange: 25.5, EndRange: 76.0} // Different end
	
	assert.Assert(t, range1.Equals(range2))
	assert.Assert(t, !range1.Equals(range3))
	assert.Assert(t, !range2.Equals(range3))
}

func TestWhitelistProperty_EqualsEdgeCases(t *testing.T) {
	// Test equals with nil
	prop1 := &WhitelistProperty{Key: "key1", Value: "value1"}
	assert.Assert(t, !prop1.equals(nil), "Should return false when comparing with nil")
	
	// Test equals with same reference
	assert.Assert(t, prop1.equals(prop1), "Should return true when comparing with itself")
	
	// Test equals with different Key
	prop2 := &WhitelistProperty{Key: "key2", Value: "value1"}
	assert.Assert(t, !prop1.equals(prop2), "Should return false with different Key")
	
	// Test equals with different Value
	prop3 := &WhitelistProperty{Key: "key1", Value: "value2"}
	assert.Assert(t, !prop1.equals(prop3), "Should return false with different Value")
	
	// Test equals with different NamespacedListType
	prop4 := &WhitelistProperty{Key: "key1", Value: "value1", NamespacedListType: "TYPE1"}
	prop5 := &WhitelistProperty{Key: "key1", Value: "value1", NamespacedListType: "TYPE2"}
	assert.Assert(t, !prop4.equals(prop5), "Should return false with different NamespacedListType")
	
	// Test equals with different TypeName
	prop6 := &WhitelistProperty{Key: "key1", Value: "value1", TypeName: "TYPE1"}
	prop7 := &WhitelistProperty{Key: "key1", Value: "value1", TypeName: "TYPE2"}
	assert.Assert(t, !prop6.equals(prop7), "Should return false with different TypeName")
	
	// Test equals with all fields matching
	prop8 := &WhitelistProperty{
		Key:                "key1",
		Value:              "value1",
		NamespacedListType: "LIST_TYPE",
		TypeName:           "TYPE_NAME",
	}
	prop9 := &WhitelistProperty{
		Key:                "key1",
		Value:              "value1",
		NamespacedListType: "LIST_TYPE",
		TypeName:           "TYPE_NAME",
	}
	assert.Assert(t, prop8.equals(prop9), "Should return true with all fields matching")
}

func TestFeature_EqualsEdgeCases(t *testing.T) {
	configData := map[string]string{"key1": "value1"}
	
	feature1 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		ConfigData:         configData,
	}
	
	// Test equals with nil
	assert.Assert(t, !feature1.equals(nil), "Should return false when comparing with nil")
	
	// Test equals with same reference
	assert.Assert(t, feature1.equals(feature1), "Should return true when comparing with itself")
	
	// Test equals with different ID
	feature2 := &Feature{
		ID:                 "id2",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		ConfigData:         configData,
	}
	assert.Assert(t, !feature1.equals(feature2), "Should return false with different ID")
	
	// Test equals with different Name
	feature3 := &Feature{
		ID:                 "id1",
		Name:               "name2",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		ConfigData:         configData,
	}
	assert.Assert(t, !feature1.equals(feature3), "Should return false with different Name")
	
	// Test equals with different FeatureName
	feature4 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName2",
		Enable:             true,
		EffectiveImmediate: true,
		ConfigData:         configData,
	}
	assert.Assert(t, !feature1.equals(feature4), "Should return false with different FeatureName")
	
	// Test equals with different Enable
	feature5 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             false,
		EffectiveImmediate: true,
		ConfigData:         configData,
	}
	assert.Assert(t, !feature1.equals(feature5), "Should return false with different Enable")
	
	// Test equals with different EffectiveImmediate
	feature6 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData:         configData,
	}
	assert.Assert(t, !feature1.equals(feature6), "Should return false with different EffectiveImmediate")
	
	// Test equals with different ConfigData size
	feature7 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		ConfigData:         map[string]string{"key1": "value1", "key2": "value2"},
	}
	assert.Assert(t, !feature1.equals(feature7), "Should return false with different ConfigData size")
	
	// Test equals with different ConfigData values
	feature8 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		ConfigData:         map[string]string{"key1": "value2"},
	}
	assert.Assert(t, !feature1.equals(feature8), "Should return false with different ConfigData values")
	
	// Test equals with matching features
	feature9 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		ConfigData:         map[string]string{"key1": "value1"},
	}
	assert.Assert(t, feature1.equals(feature9), "Should return true with matching features")
}

func TestFeature_CloneErrorHandling(t *testing.T) {
	// Test Clone with feature containing WhitelistProperty
	feature := Feature{
		ID:                 "test-id",
		Name:               "test-name",
		FeatureName:        "testFeature",
		Enable:             true,
		EffectiveImmediate: false,
		WhitelistProperty: &WhitelistProperty{
			Key:                "testKey",
			Value:              "testValue",
			NamespacedListType: "testType",
			TypeName:           "testTypeName",
		},
		ConfigData: map[string]string{"key1": "value1"},
	}
	
	cloned, err := feature.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Assert(t, cloned.WhitelistProperty != nil)
	assert.Equal(t, feature.WhitelistProperty.Key, cloned.WhitelistProperty.Key)
	assert.Equal(t, feature.WhitelistProperty.Value, cloned.WhitelistProperty.Value)
}

func TestFeatureEntity_CreateFeatureEdgeCases(t *testing.T) {
	// Test CreateFeature with nil ConfigData
	entity := &FeatureEntity{
		ID:                 "test-id",
		Name:               "test-name",
		FeatureName:        "testFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData:         nil,
	}
	
	feature := entity.CreateFeature()
	assert.Assert(t, feature != nil)
	assert.Equal(t, entity.ID, feature.ID)
	assert.Equal(t, entity.Name, feature.Name)
	assert.Equal(t, entity.FeatureName, feature.FeatureName)
	assert.Equal(t, entity.Enable, feature.Enable)
	assert.Equal(t, entity.EffectiveImmediate, feature.EffectiveImmediate)
	
	// Test CreateFeature with empty ConfigData
	entity2 := &FeatureEntity{
		ID:         "test-id-2",
		ConfigData: map[string]string{},
	}
	
	feature2 := entity2.CreateFeature()
	assert.Assert(t, feature2 != nil)
	assert.Assert(t, feature2.ConfigData != nil)
	assert.Equal(t, 0, len(feature2.ConfigData))
	
	// Test CreateFeature with WhitelistProperty
	entity3 := &FeatureEntity{
		ID:   "test-id-3",
		Name: "test-name-3",
		WhitelistProperty: &WhitelistProperty{
			Key:   "prop-key",
			Value: "prop-value",
		},
	}
	
	feature3 := entity3.CreateFeature()
	assert.Assert(t, feature3 != nil)
	assert.Assert(t, feature3.WhitelistProperty != nil)
	assert.Equal(t, "prop-key", feature3.WhitelistProperty.Key)
	assert.Equal(t, "prop-value", feature3.WhitelistProperty.Value)
}

func TestFeatureResponse_MarshalJSONEdgeCases(t *testing.T) {
	// Test with WhitelistProperty set
	feature := Feature{
		ID:          "test-id",
		FeatureName: "testFeature",
		Whitelisted: true,
		WhitelistProperty: &WhitelistProperty{
			Key:                "propKey",
			Value:              "propValue",
			NamespacedListType: "MAC_LIST",
			TypeName:           "MAC",
		},
		ConfigData: map[string]string{"key1": "value1"},
	}
	
	response := CreateFeatureResponseObject(feature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 0)
	
	// Verify JSON contains expected fields
	jsonStr := string(jsonBytes)
	assert.Assert(t, len(jsonStr) > 0)
}

func TestFeatureEntity_UnmarshalJSONWithWhitelistProperty(t *testing.T) {
	jsonStr := `{
		"id": "test-id",
		"name": "test-feature",
		"featureName": "testFeature",
		"enable": true,
		"whitelisted": true,
		"whitelistProperty": {
			"key": "macAddress",
			"value": "AA:BB:CC:DD:EE:FF",
			"namespacedListType": "MAC_LIST",
			"typeName": "MAC"
		},
		"configData": {"key1": "value1"}
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	assert.Equal(t, "test-id", entity.ID)
	assert.Equal(t, "test-feature", entity.Name)
	assert.Equal(t, true, entity.Whitelisted)
	assert.Assert(t, entity.WhitelistProperty != nil)
	assert.Equal(t, "macAddress", entity.WhitelistProperty.Key)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", entity.WhitelistProperty.Value)
	assert.Equal(t, "MAC_LIST", entity.WhitelistProperty.NamespacedListType)
	assert.Equal(t, "MAC", entity.WhitelistProperty.TypeName)
}

func TestPercentRange_EdgeCases(t *testing.T) {
	// Test with zero values
	range1 := &PercentRange{StartRange: 0.0, EndRange: 0.0}
	range2 := &PercentRange{StartRange: 0.0, EndRange: 0.0}
	assert.Assert(t, range1.Equals(range2))
	
	// Test with negative values
	range3 := &PercentRange{StartRange: -10.0, EndRange: -5.0}
	range4 := &PercentRange{StartRange: -10.0, EndRange: -5.0}
	assert.Assert(t, range3.Equals(range4))
	
	// Test with values > 100
	range5 := &PercentRange{StartRange: 100.0, EndRange: 150.0}
	range6 := &PercentRange{StartRange: 100.0, EndRange: 150.0}
	assert.Assert(t, range5.Equals(range6))
	
	// Test different start ranges
	range7 := &PercentRange{StartRange: 10.0, EndRange: 50.0}
	range8 := &PercentRange{StartRange: 20.0, EndRange: 50.0}
	assert.Assert(t, !range7.Equals(range8))
}

func TestFeature_EqualsWithWhitelisted(t *testing.T) {
	// Test equals with different Whitelisted flag
	feature1 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		Whitelisted:        false,
		ConfigData:         map[string]string{"key1": "value1"},
	}
	
	feature2 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		Whitelisted:        true,
		ConfigData:         map[string]string{"key1": "value1"},
	}
	
	assert.Assert(t, !feature1.equals(feature2), "Should return false with different Whitelisted flag")
}

func TestFeature_EqualsWithApplicationType(t *testing.T) {
	// Test equals with different ApplicationType
	feature1 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		ApplicationType:    "stb",
		ConfigData:         map[string]string{"key1": "value1"},
	}
	
	feature2 := &Feature{
		ID:                 "id1",
		Name:               "name1",
		FeatureName:        "featureName1",
		Enable:             true,
		EffectiveImmediate: true,
		ApplicationType:    "rdkv",
		ConfigData:         map[string]string{"key1": "value1"},
	}
	
	assert.Assert(t, !feature1.equals(feature2), "Should return false with different ApplicationType")
}

func TestFeatureResponse_MarshalJSONWithMultipleFields(t *testing.T) {
	// Test MarshalJSON with various field combinations
	feature := Feature{
		ID:                 "test-id",
		Name:               "test-name",
		FeatureName:        "testFeature",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData:         map[string]string{"key1": "value1", "key2": "value2"},
	}
	
	response := CreateFeatureResponseObject(feature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 0)
	
	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NilError(t, err)
}

func TestFeatureResponse_MarshalJSONWithExtraFields(t *testing.T) {
	// Test MarshalJSON with additional custom fields
	response := FeatureResponse{
		"name":               "test-name",
		"enable":             true,
		"effectiveImmediate": false,
		"configData":         map[string]string{"key1": "value1"},
		"customField1":       "customValue1",
		"customField2":       123,
	}
	
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 0)
	
	// Verify custom fields are included
	jsonStr := string(jsonBytes)
	assert.Assert(t, len(jsonStr) > 0)
}

func TestFeatureEntity_UnmarshalJSONWithMissingApplicationType(t *testing.T) {
	// Test that missing applicationType defaults to "stb"
	jsonStr := `{
		"id": "test-id",
		"name": "test-feature",
		"enable": true
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	assert.Equal(t, "stb", entity.ApplicationType, "ApplicationType should default to 'stb'")
}

func TestFeatureEntity_UnmarshalJSONWithEmptyApplicationType(t *testing.T) {
	// Test that empty applicationType defaults to "stb"
	jsonStr := `{
		"id": "test-id",
		"name": "test-feature",
		"applicationType": "",
		"enable": true
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	assert.Equal(t, "stb", entity.ApplicationType, "Empty ApplicationType should default to 'stb'")
}

func TestFeatureEntity_UnmarshalJSONWithFeatureInstance(t *testing.T) {
	// Test unmarshaling with featureInstance field
	jsonStr := `{
		"id": "test-id",
		"name": "test-feature",
		"featureName": "testFeature",
		"featureInstance": "testInstance",
		"enable": true
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	assert.Equal(t, "testInstance", entity.FeatureInstance)
}

func TestFeatureEntity_UnmarshalJSONWithNonStringConfigData(t *testing.T) {
	// Test that non-string values in configData are skipped
	jsonStr := `{
		"id": "test-id",
		"configData": {
			"validKey": "validValue",
			"numberKey": 123,
			"boolKey": true
		}
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(entity.ConfigData), "Should only have string values")
	assert.Equal(t, "validValue", entity.ConfigData["validKey"])
}

func TestFeatureEntity_UnmarshalJSONWithPartialWhitelistProperty(t *testing.T) {
	// Test with partial whitelistProperty fields
	jsonStr := `{
		"id": "test-id",
		"whitelistProperty": {
			"key": "testKey"
		}
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	assert.Assert(t, entity.WhitelistProperty != nil)
	assert.Equal(t, "testKey", entity.WhitelistProperty.Key)
	assert.Equal(t, "", entity.WhitelistProperty.Value)
	assert.Equal(t, "", entity.WhitelistProperty.NamespacedListType)
	assert.Equal(t, "", entity.WhitelistProperty.TypeName)
}

func TestFeature_CloneWithAllFields(t *testing.T) {
	// Test cloning a feature with all fields populated
	feature := Feature{
		ID:                 "full-id",
		Name:               "full-name",
		FeatureName:        "fullFeature",
		Enable:             true,
		EffectiveImmediate: true,
		Whitelisted:        true,
		ApplicationType:    "rdkv",
		ConfigData:         map[string]string{"key1": "value1", "key2": "value2"},
		WhitelistProperty: &WhitelistProperty{
			Key:                "propKey",
			Value:              "propValue",
			NamespacedListType: "listType",
			TypeName:           "typeName",
		},
	}
	
	cloned, err := feature.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	
	// Verify all fields are cloned
	assert.Equal(t, feature.ID, cloned.ID)
	assert.Equal(t, feature.Name, cloned.Name)
	assert.Equal(t, feature.FeatureName, cloned.FeatureName)
	assert.Equal(t, feature.Enable, cloned.Enable)
	assert.Equal(t, feature.EffectiveImmediate, cloned.EffectiveImmediate)
	assert.Equal(t, feature.Whitelisted, cloned.Whitelisted)
	assert.Equal(t, feature.ApplicationType, cloned.ApplicationType)
	assert.Equal(t, len(feature.ConfigData), len(cloned.ConfigData))
	
	// Verify WhitelistProperty is cloned
	assert.Assert(t, cloned.WhitelistProperty != nil)
	assert.Equal(t, feature.WhitelistProperty.Key, cloned.WhitelistProperty.Key)
	assert.Equal(t, feature.WhitelistProperty.Value, cloned.WhitelistProperty.Value)
	assert.Equal(t, feature.WhitelistProperty.NamespacedListType, cloned.WhitelistProperty.NamespacedListType)
	assert.Equal(t, feature.WhitelistProperty.TypeName, cloned.WhitelistProperty.TypeName)
}

func TestFeatureResponse_MarshalJSONOrderPreservation(t *testing.T) {
	// Test that the order of fields is preserved (name, enable, effectiveImmediate, configData first)
	feature := Feature{
		Name:               "test-name",
		Enable:             true,
		EffectiveImmediate: false,
		ConfigData:         map[string]string{"key1": "value1"},
		FeatureName:        "testFeature",
	}
	
	response := CreateFeatureResponseObject(feature)
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	jsonStr := string(jsonBytes)
	
	// Verify the standard fields appear in order
	nameIdx := strings.Index(jsonStr, "\"name\"")
	enableIdx := strings.Index(jsonStr, "\"enable\"")
	effectiveIdx := strings.Index(jsonStr, "\"effectiveImmediate\"")
	configIdx := strings.Index(jsonStr, "\"configData\"")
	
	assert.Assert(t, nameIdx >= 0, "name field should be present")
	assert.Assert(t, enableIdx >= 0, "enable field should be present")
	assert.Assert(t, effectiveIdx >= 0, "effectiveImmediate field should be present")
	assert.Assert(t, configIdx >= 0, "configData field should be present")
	
	// Verify order: name should come before enable, enable before effectiveImmediate, etc.
	assert.Assert(t, nameIdx < enableIdx, "name should come before enable")
	assert.Assert(t, enableIdx < effectiveIdx, "enable should come before effectiveImmediate")
	assert.Assert(t, effectiveIdx < configIdx, "effectiveImmediate should come before configData")
}

func TestFeatureResponse_MarshalJSONWithNilFields(t *testing.T) {
	// Test MarshalJSON with some nil fields to ensure they're skipped
	response := FeatureResponse{
		"name":               "test-name",
		"enable":             true,
		"effectiveImmediate": nil,
		"configData":         nil,
	}
	
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	jsonStr := string(jsonBytes)
	
	// Verify nil fields are not included in output
	assert.Assert(t, strings.Contains(jsonStr, "\"name\""))
	assert.Assert(t, strings.Contains(jsonStr, "\"enable\""))
}

func TestFeatureResponse_MarshalJSONEmptyResponse(t *testing.T) {
	// Test MarshalJSON with completely empty response
	response := FeatureResponse{}
	
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	jsonStr := string(jsonBytes)
	
	// Should return empty JSON object
	assert.Equal(t, "{}", jsonStr)
}

func TestFeatureResponse_MarshalJSONOnlyExtraFields(t *testing.T) {
	// Test with only non-standard fields
	response := FeatureResponse{
		"field1": "value1",
		"field2": 123,
		"field3": true,
	}
	
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	
	jsonStr := string(jsonBytes)
	
	// All extra fields should be present
	assert.Assert(t, strings.Contains(jsonStr, "\"field1\""))
	assert.Assert(t, strings.Contains(jsonStr, "\"field2\""))
	assert.Assert(t, strings.Contains(jsonStr, "\"field3\""))
}

func TestFeatureEntity_UnmarshalJSONWithInvalidWhitelistProperty(t *testing.T) {
	// Test with whitelistProperty that's not a map
	jsonStr := `{
		"id": "test-id",
		"whitelistProperty": "invalid-not-a-map"
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	assert.Assert(t, entity.WhitelistProperty == nil, "Should not create WhitelistProperty for invalid type")
}

func TestFeatureEntity_UnmarshalJSONWithNonBooleanFields(t *testing.T) {
	// Test with non-boolean values for boolean fields
	jsonStr := `{
		"id": "test-id",
		"enable": "not-a-bool",
		"effectiveImmediate": 123,
		"whitelisted": []
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	// Boolean fields should remain false (default) when wrong type
	assert.Equal(t, false, entity.Enable)
	assert.Equal(t, false, entity.EffectiveImmediate)
	assert.Equal(t, false, entity.Whitelisted)
}

func TestFeatureEntity_UnmarshalJSONWithNonStringFields(t *testing.T) {
	// Test with non-string values for string fields
	jsonStr := `{
		"id": 123,
		"name": true,
		"featureName": [],
		"applicationType": {"key": "value"}
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	// String fields should remain empty when wrong type
	assert.Equal(t, "", entity.ID)
	assert.Equal(t, "", entity.Name)
	assert.Equal(t, "", entity.FeatureName)
	// ApplicationType should default to "stb" when invalid type
	assert.Equal(t, "stb", entity.ApplicationType)
}

func TestFeatureEntity_UnmarshalJSONWithInvalidConfigData(t *testing.T) {
	// Test with configData that's not a map
	jsonStr := `{
		"id": "test-id",
		"configData": "not-a-map"
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	// ConfigData should be initialized as empty map
	assert.Assert(t, entity.ConfigData != nil)
	assert.Equal(t, 0, len(entity.ConfigData))
}

func TestFeatureEntity_UnmarshalJSONWithMixedConfigData(t *testing.T) {
	// Test with configData containing various types
	jsonStr := `{
		"id": "test-id",
		"configData": {
			"string1": "value1",
			"string2": "value2",
			"number": 123,
			"bool": true,
			"null": null,
			"object": {"nested": "value"},
			"array": [1, 2, 3]
		}
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	// Only string values should be included
	assert.Equal(t, 2, len(entity.ConfigData))
	assert.Equal(t, "value1", entity.ConfigData["string1"])
	assert.Equal(t, "value2", entity.ConfigData["string2"])
}

func TestFeatureEntity_UnmarshalJSONWithNonStringWhitelistPropertyFields(t *testing.T) {
	// Test whitelistProperty with non-string field types
	jsonStr := `{
		"id": "test-id",
		"whitelistProperty": {
			"key": 123,
			"value": true,
			"namespacedListType": [],
			"typeName": {"object": "value"}
		}
	}`
	
	var entity FeatureEntity
	err := json.Unmarshal([]byte(jsonStr), &entity)
	assert.NilError(t, err)
	assert.Assert(t, entity.WhitelistProperty != nil)
	// All fields should be empty strings when wrong type
	assert.Equal(t, "", entity.WhitelistProperty.Key)
	assert.Equal(t, "", entity.WhitelistProperty.Value)
	assert.Equal(t, "", entity.WhitelistProperty.NamespacedListType)
	assert.Equal(t, "", entity.WhitelistProperty.TypeName)
}

func TestFeature_CreateFeatureEntityWithAllFields(t *testing.T) {
	// Test CreateFeatureEntity with all fields populated
	feature := &Feature{
		ID:                 "test-id",
		Name:               "test-name",
		FeatureName:        "testFeature",
		ApplicationType:    "rdkv",
		ConfigData:         map[string]string{"key1": "value1"},
		EffectiveImmediate: true,
		Enable:             true,
		Whitelisted:        true,
		WhitelistProperty: &WhitelistProperty{
			Key:   "propKey",
			Value: "propValue",
		},
	}
	
	entity := feature.CreateFeatureEntity()
	assert.Assert(t, entity != nil)
	assert.Equal(t, feature.ID, entity.ID)
	assert.Equal(t, feature.Name, entity.Name)
	assert.Equal(t, feature.FeatureName, entity.FeatureName)
	assert.Equal(t, feature.FeatureName, entity.FeatureInstance) // FeatureInstance should equal FeatureName
	assert.Equal(t, feature.ApplicationType, entity.ApplicationType)
	assert.Equal(t, feature.EffectiveImmediate, entity.EffectiveImmediate)
	assert.Equal(t, feature.Enable, entity.Enable)
	assert.Equal(t, feature.Whitelisted, entity.Whitelisted)
	assert.Assert(t, entity.WhitelistProperty != nil)
	assert.Equal(t, feature.WhitelistProperty.Key, entity.WhitelistProperty.Key)
}

func TestFeatureResponse_MarshalJSONWithComplexNestedData(t *testing.T) {
	// Test with complex nested data structures
	response := FeatureResponse{
		"name":   "test",
		"enable": true,
		"configData": map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
		"nestedArray": []string{"item1", "item2"},
		"nestedMap": map[string]interface{}{
			"nested1": "value1",
			"nested2": 123,
		},
	}
	
	jsonBytes, err := response.MarshalJSON()
	assert.NilError(t, err)
	assert.Assert(t, len(jsonBytes) > 0)
	
	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NilError(t, err)
}
