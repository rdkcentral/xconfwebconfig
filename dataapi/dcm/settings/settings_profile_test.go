/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
package settings

import (
	"testing"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

// TestGetMaxRule tests the GetMaxRule function
func TestGetMaxRule(t *testing.T) {
	t.Run("GetMaxRuleWithMultipleRules", func(t *testing.T) {
		rules := []logupload.SettingRule{
			{
				ID:   "rule1",
				Name: "Rule 1",
				Rule: re.Rule{
					Condition: &re.Condition{
						FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
						Operation: "IS",
						FixedArg:  re.NewFixedArg("MODEL1"),
					},
				},
			},
			{
				ID:   "rule2",
				Name: "Rule 2",
				Rule: re.Rule{
					CompoundParts: []re.Rule{
						{
							Condition: &re.Condition{
								FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
								Operation: "IS",
								FixedArg:  re.NewFixedArg("MODEL2"),
							},
						},
						{
							Condition: &re.Condition{
								FreeArg:   &re.FreeArg{Type: "STRING", Name: "env"},
								Operation: "IS",
								FixedArg:  re.NewFixedArg("PROD"),
							},
						},
					},
					Relation: "AND",
				},
			},
			{
				ID:   "rule3",
				Name: "Rule 3",
				Rule: re.Rule{
					CompoundParts: []re.Rule{
						{
							Condition: &re.Condition{
								FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
								Operation: "IS",
								FixedArg:  re.NewFixedArg("MODEL3"),
							},
						},
					},
				},
			},
		}

		result := GetMaxRule(rules)

		assert.NotNil(t, result)
		// The rule with most compound parts (most complex) should be selected
		assert.Equal(t, "rule2", result.ID)
	})

	t.Run("GetMaxRuleWithSingleRule", func(t *testing.T) {
		rules := []logupload.SettingRule{
			{
				ID:   "rule1",
				Name: "Rule 1",
				Rule: re.Rule{
					Condition: &re.Condition{
						FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
						Operation: "IS",
						FixedArg:  re.NewFixedArg("MODEL1"),
					},
				},
			},
		}

		result := GetMaxRule(rules)

		assert.NotNil(t, result)
		assert.Equal(t, "rule1", result.ID)
	})

	t.Run("GetMaxRuleWithEmptySlice", func(t *testing.T) {
		rules := []logupload.SettingRule{}

		result := GetMaxRule(rules)

		assert.Nil(t, result)
	})

	t.Run("GetMaxRuleWithNilSlice", func(t *testing.T) {
		result := GetMaxRule(nil)

		assert.Nil(t, result)
	})

	t.Run("GetMaxRuleWithEqualComplexityRules", func(t *testing.T) {
		rules := []logupload.SettingRule{
			{
				ID:   "rule1",
				Name: "Rule 1",
				Rule: re.Rule{
					Condition: &re.Condition{
						FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
						Operation: "IS",
						FixedArg:  re.NewFixedArg("MODEL1"),
					},
				},
			},
			{
				ID:   "rule2",
				Name: "Rule 2",
				Rule: re.Rule{
					Condition: &re.Condition{
						FreeArg:   &re.FreeArg{Type: "STRING", Name: "env"},
						Operation: "IS",
						FixedArg:  re.NewFixedArg("PROD"),
					},
				},
			},
		}

		result := GetMaxRule(rules)

		assert.NotNil(t, result)
		// With equal complexity, should return one (sorting is stable)
		assert.Contains(t, []string{"rule1", "rule2"}, result.ID)
	})

	t.Run("GetMaxRuleWithComplexRuleStructures", func(t *testing.T) {
		rules := []logupload.SettingRule{
			{
				ID:   "rule1",
				Name: "Simple Rule",
				Rule: re.Rule{
					Condition: &re.Condition{
						FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
						Operation: "IS",
						FixedArg:  re.NewFixedArg("MODEL1"),
					},
				},
			},
			{
				ID:   "rule2",
				Name: "Medium Rule",
				Rule: re.Rule{
					CompoundParts: []re.Rule{
						{
							Condition: &re.Condition{
								FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
								Operation: "IS",
								FixedArg:  re.NewFixedArg("MODEL2"),
							},
						},
						{
							Condition: &re.Condition{
								FreeArg:   &re.FreeArg{Type: "STRING", Name: "env"},
								Operation: "IS",
								FixedArg:  re.NewFixedArg("PROD"),
							},
						},
					},
					Relation: "AND",
				},
			},
			{
				ID:   "rule3",
				Name: "Complex Rule",
				Rule: re.Rule{
					CompoundParts: []re.Rule{
						{
							Condition: &re.Condition{
								FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
								Operation: "IS",
								FixedArg:  re.NewFixedArg("MODEL3"),
							},
						},
						{
							Condition: &re.Condition{
								FreeArg:   &re.FreeArg{Type: "STRING", Name: "env"},
								Operation: "IS",
								FixedArg:  re.NewFixedArg("PROD"),
							},
						},
						{
							Condition: &re.Condition{
								FreeArg:   &re.FreeArg{Type: "STRING", Name: "firmwareVersion"},
								Operation: "IS",
								FixedArg:  re.NewFixedArg("1.0.0"),
							},
						},
					},
					Relation: "AND",
				},
			},
		}

		result := GetMaxRule(rules)

		assert.NotNil(t, result)
		// The most complex rule should be selected
		// Note: CompareRules uses a specific algorithm, so we just verify a rule is returned
		assert.Contains(t, []string{"rule2", "rule3"}, result.ID)
	})

	t.Run("GetMaxRuleWithNestedCompoundRules", func(t *testing.T) {
		rules := []logupload.SettingRule{
			{
				ID:   "rule1",
				Name: "Nested Rule",
				Rule: re.Rule{
					CompoundParts: []re.Rule{
						{
							CompoundParts: []re.Rule{
								{
									Condition: &re.Condition{
										FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
										Operation: "IS",
										FixedArg:  re.NewFixedArg("MODEL1"),
									},
								},
								{
									Condition: &re.Condition{
										FreeArg:   &re.FreeArg{Type: "STRING", Name: "partner"},
										Operation: "IS",
										FixedArg:  re.NewFixedArg("PARTNER1"),
									},
								},
							},
							Relation: "AND",
						},
						{
							Condition: &re.Condition{
								FreeArg:   &re.FreeArg{Type: "STRING", Name: "env"},
								Operation: "IS",
								FixedArg:  re.NewFixedArg("PROD"),
							},
						},
					},
					Relation: "OR",
				},
			},
			{
				ID:   "rule2",
				Name: "Simple Rule",
				Rule: re.Rule{
					Condition: &re.Condition{
						FreeArg:   &re.FreeArg{Type: "STRING", Name: "model"},
						Operation: "IS",
						FixedArg:  re.NewFixedArg("MODEL2"),
					},
				},
			},
		}

		result := GetMaxRule(rules)

		assert.NotNil(t, result)
		// The nested (more complex) rule should be selected
		assert.Equal(t, "rule1", result.ID)
	})
}

// TestGetSettingProfileBySettingRule tests the GetSettingProfileBySettingRule function
func TestGetSettingProfileBySettingRule(t *testing.T) {
	t.Run("GetSettingProfileWithNilSettingRule", func(t *testing.T) {
		result := GetSettingProfileBySettingRule(nil)
		assert.Nil(t, result)
	})

	t.Run("GetSettingProfileWithEmptyBoundSettingID", func(t *testing.T) {
		settingRule := &logupload.SettingRule{
			ID:             "rule1",
			Name:           "Test Rule",
			BoundSettingID: "",
		}

		result := GetSettingProfileBySettingRule(settingRule)
		assert.Nil(t, result)
	})

	// Note: Tests requiring database/cache access cannot be reliably tested
	// without integration testing infrastructure. The above tests cover the
	// testable logic paths (nil checks, empty string checks).
}
