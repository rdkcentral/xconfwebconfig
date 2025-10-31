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
	"testing"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"gotest.tools/assert"
)

func TestFeatureRuleMarshaling(t *testing.T) {

	src := `{
    "applicationType": "stb",
    "featureIds": [
        "d471efce-b7d6-4419-a40e-5a095e8b6318",
        "7a98f5d9-9652-47a4-9ee9-4814db8aaa24"
    ],
    "id": "8a0dce3d-0f98-4cd5-8d93-cdb9cefb5211",
    "name": "Test_BLE_NS",
    "priority": 1,
    "rule": {
        "compoundParts": [],
        "condition": {
            "fixedArg": {
                "bean": {
                    "value": {
                        "java.lang.String": "AA:AA:AA:AA:AA:AA"
                    }
                }
            },
            "freeArg": {
                "name": "estbMacAddress",
                "type": "STRING"
            },
            "operation": "IS"
        },
        "negated": false
    }
}`

	var featureRule FeatureRule
	err := json.Unmarshal([]byte(src), &featureRule)
	assert.NilError(t, err)

	t.Logf("\n\nfeatureRule = %v\n\n", featureRule)

	t.Logf("\n\nfeatureRule.Rule = %v\n\n", featureRule.Rule)

	t.Logf("\n\nfeatureRule.Rule.Condition = %v\n\n", featureRule.Rule.Condition)

	t.Logf("\n\nfeatureRule.Rule.Condition.FixedArg = %v\n\n", featureRule.Rule.Condition.FixedArg)

	t.Logf("\n\nfeatureRule.Rule.Condition.FreeArg = %v\n\n", featureRule.Rule.Condition.FreeArg)
}

func TestNewFeatureRuleInf(t *testing.T) {
	ruleInf := NewFeatureRuleInf()
	assert.Assert(t, ruleInf != nil, "NewFeatureRuleInf should return non-nil")
	
	// Check that it returns a *FeatureRule type
	rule, ok := ruleInf.(*FeatureRule)
	assert.Assert(t, ok, "NewFeatureRuleInf should return *FeatureRule type")
	assert.Equal(t, "", rule.Id, "Id should be empty by default")
	assert.Equal(t, "", rule.Name, "Name should be empty by default")
	assert.Equal(t, 0, rule.Priority, "Priority should be 0 by default")
	assert.Assert(t, rule.Rule == nil, "Rule should be nil by default")
	assert.Assert(t, rule.FeatureIds == nil, "FeatureIds should be nil by default")
	assert.Equal(t, "", rule.ApplicationType, "ApplicationType should be empty by default")
}

func TestFeatureRule_GetId(t *testing.T) {
	rule := &FeatureRule{
		Id: "test-id-123",
	}
	
	result := rule.GetId()
	assert.Equal(t, "test-id-123", result, "GetId should return the Id field")
}

func TestFeatureRule_GetID(t *testing.T) {
	rule := &FeatureRule{
		Id: "test-id-456",
	}
	
	result := rule.GetID()
	assert.Equal(t, "test-id-456", result, "GetID should return the Id field")
}

func TestFeatureRule_GetPriority(t *testing.T) {
	rule := &FeatureRule{
		Priority: 42,
	}
	
	result := rule.GetPriority()
	assert.Equal(t, 42, result, "GetPriority should return the Priority field")
}

func TestFeatureRule_SetPriority(t *testing.T) {
	rule := &FeatureRule{}
	
	rule.SetPriority(100)
	assert.Equal(t, 100, rule.Priority, "SetPriority should set the Priority field")
	
	rule.SetPriority(-5)
	assert.Equal(t, -5, rule.Priority, "SetPriority should set negative values")
}

func TestFeatureRule_GetRule(t *testing.T) {
	testRule := &re.Rule{
		Xxid: "test-rule",
	}
	
	featureRule := &FeatureRule{
		Rule: testRule,
	}
	
	result := featureRule.GetRule()
	assert.Equal(t, testRule, result, "GetRule should return the Rule field")
	
	// Test with nil rule
	featureRule.Rule = nil
	result = featureRule.GetRule()
	assert.Assert(t, result == nil, "GetRule should return nil when Rule is nil")
}

func TestFeatureRule_GetName(t *testing.T) {
	rule := &FeatureRule{
		Name: "test-feature-rule",
	}
	
	result := rule.GetName()
	assert.Equal(t, "test-feature-rule", result, "GetName should return the Name field")
}

func TestFeatureRule_GetTemplateId(t *testing.T) {
	rule := &FeatureRule{}
	
	result := rule.GetTemplateId()
	assert.Equal(t, "", result, "GetTemplateId should always return empty string")
}

func TestFeatureRule_GetRuleType(t *testing.T) {
	rule := &FeatureRule{}
	
	result := rule.GetRuleType()
	assert.Equal(t, "FeatureRule", result, "GetRuleType should always return 'FeatureRule'")
}

func TestFeatureRule_SetApplicationType(t *testing.T) {
	rule := &FeatureRule{}
	
	rule.SetApplicationType("stb")
	assert.Equal(t, "stb", rule.ApplicationType, "SetApplicationType should set the ApplicationType field")
	
	rule.SetApplicationType("rdkv")
	assert.Equal(t, "rdkv", rule.ApplicationType, "SetApplicationType should update the ApplicationType field")
	
	rule.SetApplicationType("")
	assert.Equal(t, "", rule.ApplicationType, "SetApplicationType should accept empty string")
}

func TestFeatureRule_GetApplicationType(t *testing.T) {
	rule := &FeatureRule{
		ApplicationType: "stb",
	}
	
	result := rule.GetApplicationType()
	assert.Equal(t, "stb", result, "GetApplicationType should return the ApplicationType field")
	
	rule.ApplicationType = "rdkv"
	result = rule.GetApplicationType()
	assert.Equal(t, "rdkv", result, "GetApplicationType should return updated value")
}

func TestFeatureRule_Clone(t *testing.T) {
	original := &FeatureRule{
		Id:              "test-id",
		Name:            "test-rule",
		Priority:        5,
		ApplicationType: "stb",
		FeatureIds:      []string{"feature1", "feature2"},
		Rule: &re.Rule{
			Xxid: "rule-id",
		},
	}

	cloned, err := original.Clone()
	assert.NilError(t, err, "Clone should not return error")
	assert.Assert(t, cloned != nil, "Cloned rule should not be nil")
	assert.Assert(t, cloned != original, "Cloned rule should be different instance")
	
	// Check all fields are copied
	assert.Equal(t, original.Id, cloned.Id, "Id should be cloned")
	assert.Equal(t, original.Name, cloned.Name, "Name should be cloned")
	assert.Equal(t, original.Priority, cloned.Priority, "Priority should be cloned")
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType, "ApplicationType should be cloned")
	
	// Check FeatureIds slice is deep copied
	assert.Equal(t, len(original.FeatureIds), len(cloned.FeatureIds), "FeatureIds length should match")
	for i, id := range original.FeatureIds {
		assert.Equal(t, id, cloned.FeatureIds[i], "FeatureIds values should match")
	}
	
	// Verify it's a deep copy by modifying cloned and checking original is unchanged
	cloned.Id = "modified-id"
	assert.Equal(t, "test-id", original.Id, "Original Id should be unchanged after cloning")
	
	cloned.FeatureIds[0] = "modified-feature"
	assert.Equal(t, "feature1", original.FeatureIds[0], "Original FeatureIds should be unchanged after cloning")
}

func TestFeatureRule_CompleteWorkflow(t *testing.T) {
	// Test a complete workflow using all methods
	rule := &FeatureRule{}
	
	// Set values using setter methods
	rule.SetApplicationType("stb")
	rule.SetPriority(10)
	
	// Set other fields directly
	rule.Id = "workflow-test-id"
	rule.Name = "Workflow Test Rule"
	rule.FeatureIds = []string{"feature-a", "feature-b"}
	
	// Test getter methods
	assert.Equal(t, "workflow-test-id", rule.GetId(), "GetId should work")
	assert.Equal(t, "workflow-test-id", rule.GetID(), "GetID should work")
	assert.Equal(t, "Workflow Test Rule", rule.GetName(), "GetName should work")
	assert.Equal(t, 10, rule.GetPriority(), "GetPriority should work")
	assert.Equal(t, "stb", rule.GetApplicationType(), "GetApplicationType should work")
	assert.Equal(t, "", rule.GetTemplateId(), "GetTemplateId should work")
	assert.Equal(t, "FeatureRule", rule.GetRuleType(), "GetRuleType should work")
	
	// Test cloning
	cloned, err := rule.Clone()
	assert.NilError(t, err, "Clone should work in workflow")
	assert.Equal(t, rule.Id, cloned.Id, "Cloned rule should have same Id")
	assert.Equal(t, rule.Name, cloned.Name, "Cloned rule should have same Name")
}

func TestFeatureRule_CloneWithRule(t *testing.T) {
	// Test cloning a feature rule with a Rule object
	src := `{
		"id": "rule-with-rule-id",
		"name": "RuleWithRule",
		"priority": 10,
		"applicationType": "stb",
		"featureIds": ["feature1", "feature2", "feature3"],
		"rule": {
			"compoundParts": [],
			"condition": {
				"fixedArg": {
					"bean": {
						"value": {
							"java.lang.String": "AA:BB:CC:DD:EE:FF"
						}
					}
				},
				"freeArg": {
					"name": "estbMacAddress",
					"type": "STRING"
				},
				"operation": "IS"
			},
			"negated": false
		}
	}`

	var original FeatureRule
	err := json.Unmarshal([]byte(src), &original)
	assert.NilError(t, err)

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Assert(t, cloned != &original)
	
	// Verify Rule is cloned
	assert.Assert(t, cloned.Rule != nil)
	assert.Assert(t, cloned.Rule != original.Rule, "Rule should be deep copied")
	
	// Verify all fields match
	assert.Equal(t, original.Id, cloned.Id)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Priority, cloned.Priority)
	assert.Equal(t, original.ApplicationType, cloned.ApplicationType)
	assert.Equal(t, len(original.FeatureIds), len(cloned.FeatureIds))
	
	// Modify cloned and ensure original is unchanged
	cloned.Name = "ModifiedName"
	assert.Equal(t, "RuleWithRule", original.Name)
	
	cloned.FeatureIds = append(cloned.FeatureIds, "feature4")
	assert.Equal(t, 3, len(original.FeatureIds))
}

func TestFeatureRule_CloneEmptyFeatureIds(t *testing.T) {
	// Test cloning with empty FeatureIds slice
	original := &FeatureRule{
		Id:              "test-id",
		Name:            "test-rule",
		Priority:        5,
		ApplicationType: "stb",
		FeatureIds:      []string{},
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Equal(t, 0, len(cloned.FeatureIds))
}

func TestFeatureRule_CloneNilFeatureIds(t *testing.T) {
	// Test cloning with nil FeatureIds
	original := &FeatureRule{
		Id:              "test-id",
		Name:            "test-rule",
		Priority:        5,
		ApplicationType: "stb",
		FeatureIds:      nil,
	}

	cloned, err := original.Clone()
	assert.NilError(t, err)
	assert.Assert(t, cloned != nil)
	assert.Assert(t, cloned.FeatureIds == nil)
}

func TestFeatureRule_GettersWithZeroValues(t *testing.T) {
	// Test getters with zero/empty values
	rule := &FeatureRule{}
	
	assert.Equal(t, "", rule.GetId())
	assert.Equal(t, "", rule.GetID())
	assert.Equal(t, "", rule.GetName())
	assert.Equal(t, 0, rule.GetPriority())
	assert.Equal(t, "", rule.GetApplicationType())
	assert.Equal(t, "", rule.GetTemplateId())
	assert.Equal(t, "FeatureRule", rule.GetRuleType())
	assert.Assert(t, rule.GetRule() == nil)
}

func TestFeatureRule_SettersWithVariousValues(t *testing.T) {
	rule := &FeatureRule{}
	
	// Test setting various priority values
	rule.SetPriority(0)
	assert.Equal(t, 0, rule.Priority)
	
	rule.SetPriority(1)
	assert.Equal(t, 1, rule.Priority)
	
	rule.SetPriority(999)
	assert.Equal(t, 999, rule.Priority)
	
	rule.SetPriority(-100)
	assert.Equal(t, -100, rule.Priority)
	
	// Test setting various application types
	rule.SetApplicationType("")
	assert.Equal(t, "", rule.ApplicationType)
	
	rule.SetApplicationType("stb")
	assert.Equal(t, "stb", rule.ApplicationType)
	
	rule.SetApplicationType("rdkv")
	assert.Equal(t, "rdkv", rule.ApplicationType)
	
	rule.SetApplicationType("sky")
	assert.Equal(t, "sky", rule.ApplicationType)
}

func TestFeatureRule_MarshalingWithComplexRule(t *testing.T) {
	// Test marshaling and unmarshaling with a complex rule structure
	src := `{
    "id": "complex-rule-id",
    "name": "ComplexRule",
    "priority": 15,
    "applicationType": "rdkv",
    "featureIds": ["feat1", "feat2"],
    "rule": {
        "compoundParts": [],
        "condition": {
            "fixedArg": {
                "bean": {
                    "value": {
                        "java.lang.String": "TestValue"
                    }
                }
            },
            "freeArg": {
                "name": "model",
                "type": "STRING"
            },
            "operation": "IS"
        },
        "negated": true
    }
}`

	var featureRule FeatureRule
	err := json.Unmarshal([]byte(src), &featureRule)
	assert.NilError(t, err)
	assert.Equal(t, "complex-rule-id", featureRule.Id)
	assert.Equal(t, "ComplexRule", featureRule.Name)
	assert.Equal(t, 15, featureRule.Priority)
	assert.Equal(t, "rdkv", featureRule.ApplicationType)
	assert.Equal(t, 2, len(featureRule.FeatureIds))
	assert.Assert(t, featureRule.Rule != nil)
	assert.Equal(t, true, featureRule.Rule.Negated)
}

func TestFeatureRule_EmptyJSONUnmarshaling(t *testing.T) {
	// Test unmarshaling empty JSON
	src := `{}`
	
	var featureRule FeatureRule
	err := json.Unmarshal([]byte(src), &featureRule)
	assert.NilError(t, err)
	assert.Equal(t, "", featureRule.Id)
	assert.Equal(t, "", featureRule.Name)
	assert.Equal(t, 0, featureRule.Priority)
	assert.Assert(t, featureRule.FeatureIds == nil)
}

func TestFeatureRule_PartialJSONUnmarshaling(t *testing.T) {
	// Test unmarshaling with only some fields
	src := `{
		"id": "partial-id",
		"name": "PartialRule"
	}`
	
	var featureRule FeatureRule
	err := json.Unmarshal([]byte(src), &featureRule)
	assert.NilError(t, err)
	assert.Equal(t, "partial-id", featureRule.Id)
	assert.Equal(t, "PartialRule", featureRule.Name)
	assert.Equal(t, 0, featureRule.Priority)
	assert.Equal(t, "", featureRule.ApplicationType)
	assert.Assert(t, featureRule.FeatureIds == nil)
	assert.Assert(t, featureRule.Rule == nil)
}

func TestFeatureRule_RoundTripMarshaling(t *testing.T) {
	// Test that marshaling and unmarshaling preserves data
	original := &FeatureRule{
		Id:              "roundtrip-id",
		Name:            "RoundTripRule",
		Priority:        20,
		ApplicationType: "stb",
		FeatureIds:      []string{"f1", "f2", "f3"},
	}
	
	// Marshal to JSON
	jsonBytes, err := json.Marshal(original)
	assert.NilError(t, err)
	
	// Unmarshal back
	var recovered FeatureRule
	err = json.Unmarshal(jsonBytes, &recovered)
	assert.NilError(t, err)
	
	// Verify all fields match
	assert.Equal(t, original.Id, recovered.Id)
	assert.Equal(t, original.Name, recovered.Name)
	assert.Equal(t, original.Priority, recovered.Priority)
	assert.Equal(t, original.ApplicationType, recovered.ApplicationType)
	assert.Equal(t, len(original.FeatureIds), len(recovered.FeatureIds))
	for i, id := range original.FeatureIds {
		assert.Equal(t, id, recovered.FeatureIds[i])
	}
}

func TestFeatureRule_CloneIndependence(t *testing.T) {
	// Test that cloned rule is truly independent
	original := &FeatureRule{
		Id:              "independence-test",
		Name:            "OriginalRule",
		Priority:        1,
		ApplicationType: "stb",
		FeatureIds:      []string{"original1", "original2"},
	}
	
	cloned, err := original.Clone()
	assert.NilError(t, err)
	
	// Modify every field in cloned
	cloned.Id = "cloned-id"
	cloned.Name = "ClonedRule"
	cloned.Priority = 999
	cloned.ApplicationType = "rdkv"
	cloned.FeatureIds[0] = "cloned1"
	cloned.FeatureIds = append(cloned.FeatureIds, "cloned3")
	
	// Verify original is unchanged
	assert.Equal(t, "independence-test", original.Id)
	assert.Equal(t, "OriginalRule", original.Name)
	assert.Equal(t, 1, original.Priority)
	assert.Equal(t, "stb", original.ApplicationType)
	assert.Equal(t, 2, len(original.FeatureIds))
	assert.Equal(t, "original1", original.FeatureIds[0])
	assert.Equal(t, "original2", original.FeatureIds[1])
}
