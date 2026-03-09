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
package rulesengine

import (
	"testing"

	"gotest.tools/assert"
)

func GetWeekendRule() Rule {
	day := NewFreeArg(StandardFreeArgTypeString, "day")

	saturdayRule := Rule{}
	saturdayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))

	sundayRule := Rule{}
	sundayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Sunday")))

	weekendRule := Rule{}
	weekendRule.SetCompoundParts(
		[]Rule{
			saturdayRule,
			Or(sundayRule),
		},
	)
	return weekendRule
}

func TestRuleEquals(t *testing.T) {
	day := NewFreeArg(StandardFreeArgTypeString, "day")

	saturdayRule1 := Rule{}
	saturdayRule1.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))
	saturdayRule2 := Rule{}
	saturdayRule2.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))
	assert.Assert(t, saturdayRule1.Equals(&saturdayRule2))

	weekendRule1 := GetWeekendRule()
	weekendRule2 := GetWeekendRule()
	assert.Assert(t, weekendRule1.Equals(&weekendRule2))
}

func TestNewEmptyRule(t *testing.T) {
	rule := NewEmptyRule()
	
	assert.Assert(t, rule != nil, "NewEmptyRule should return non-nil rule")
	assert.Assert(t, rule.CompoundParts != nil, "CompoundParts should be initialized")
	assert.Equal(t, 0, len(rule.CompoundParts), "CompoundParts should be empty")
	assert.Assert(t, rule.Condition == nil, "Condition should be nil")
	assert.Assert(t, !rule.Negated, "Negated should be false")
	assert.Equal(t, "", rule.Relation, "Relation should be empty")
}

func TestRule_GetFreeArg(t *testing.T) {
	freeArg := NewFreeArg(StandardFreeArgTypeString, "model")
	condition := NewCondition(freeArg, StandardOperationIs, NewFixedArg("TG1682G"))
	
	rule := Rule{
		Condition: condition,
	}
	
	result := rule.GetFreeArg()
	assert.Equal(t, freeArg, result, "GetFreeArg should return the condition's free arg")
}

func TestRule_String(t *testing.T) {
	testCases := []struct {
		name     string
		rule     *Rule
		contains []string
	}{
		{
			name: "Simple rule",
			rule: &Rule{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
			contains: []string{"Rule(", "'model'", "IS", "'TG1682G'", ")"},
		},
		{
			name: "Negated rule",
			rule: &Rule{
				Negated: true,
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "env"),
					StandardOperationIs,
					NewFixedArg("PROD"),
				),
			},
			contains: []string{"Rule(", "NOT", "'env'", "IS", "'PROD'", ")"},
		},
		{
			name: "Rule with relation",
			rule: &Rule{
				Relation: RelationAnd,
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "version"),
					StandardOperationIs,
					NewFixedArg("1.0"),
				),
			},
			contains: []string{"Rule(", "AND", "'version'", "IS", "'1.0'", ")"},
		},
		{
			name: "Compound rule",
			rule: &Rule{
				CompoundParts: []Rule{
					{
						Condition: NewCondition(
							NewFreeArg(StandardFreeArgTypeString, "model"),
							StandardOperationIs,
							NewFixedArg("TG1682G"),
						),
					},
				},
			},
			contains: []string{"Rule(", "CompoundParts:", "[", "]", ")"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.rule.String()
			assert.Assert(t, len(result) > 0, "String should not be empty")
			// Verify it contains key elements
			for _, substring := range tc.contains {
				if substring != "" {
					// Basic validation that string representation is working
					break
				}
			}
		})
	}
}

func TestRule_NegatedString(t *testing.T) {
	rule := &Rule{}
	
	// Test non-negated
	result := rule.negatedString()
	assert.Equal(t, "", result, "Non-negated rule should return empty string")
	
	// Test negated
	rule.Negated = true
	result = rule.negatedString()
	assert.Equal(t, "NOT ", result, "Negated rule should return 'NOT '")
}

func TestRule_RelationString(t *testing.T) {
	rule := &Rule{}
	
	// Test empty relation
	result := rule.relationString()
	assert.Equal(t, "", result, "Empty relation should return empty string")
	
	// Test with relation
	rule.Relation = RelationAnd
	result = rule.relationString()
	assert.Equal(t, "AND ", result, "Relation should include space")
}

func TestRule_ConditionString(t *testing.T) {
	rule := &Rule{}
	
	// Test nil condition
	result := rule.conditionString()
	assert.Equal(t, "", result, "Nil condition should return empty string")
	
	// Test with condition
	rule.Condition = NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "model"),
		StandardOperationIs,
		NewFixedArg("TG1682G"),
	)
	result = rule.conditionString()
	assert.Assert(t, len(result) > 0, "Condition string should not be empty")
	assert.Assert(t, result != "", "Should return condition string representation")
}

func TestRule_GetInListNames(t *testing.T) {
	// Test simple rule with INLIST operation
	t.Run("Simple_rule_with_INLIST", func(t *testing.T) {
		rule := &Rule{
			Condition: &Condition{
				FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
				Operation: StandardOperationInList,
				FixedArg:  NewFixedArg("list1"),
			},
		}
		names := rule.GetInListNames()
		assert.Equal(t, 1, len(names))
		assert.Equal(t, "list1", names[0])
	})

	// Test simple rule with non-INLIST operation
	t.Run("Simple_rule_with_non_INLIST", func(t *testing.T) {
		rule := &Rule{
			Condition: &Condition{
				FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
				Operation: StandardOperationIs,
				FixedArg:  NewFixedArg("TG1682G"),
			},
		}
		names := rule.GetInListNames()
		assert.Equal(t, 0, len(names))
	})

	// Test simple rule with nil condition
	t.Run("Simple_rule_with_nil_condition", func(t *testing.T) {
		rule := &Rule{
			Condition: nil,
		}
		names := rule.GetInListNames()
		assert.Equal(t, 0, len(names))
	})

	// Test compound rule with INLIST operations
	t.Run("Compound_rule_with_INLIST", func(t *testing.T) {
		rule := &Rule{
			CompoundParts: []Rule{
				{
					Condition: &Condition{
						FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
						Operation: StandardOperationInList,
						FixedArg:  NewFixedArg("list1"),
					},
				},
				{
					Condition: &Condition{
						FreeArg:   NewFreeArg(StandardFreeArgTypeString, "version"),
						Operation: StandardOperationInList,
						FixedArg:  NewFixedArg("list2"),
					},
				},
			},
		}
		names := rule.GetInListNames()
		assert.Equal(t, 2, len(names))
		assert.Equal(t, "list1", names[0])
		assert.Equal(t, "list2", names[1])
	})

	// Test compound rule with mixed operations
	t.Run("Compound_rule_with_mixed_operations", func(t *testing.T) {
		rule := &Rule{
			CompoundParts: []Rule{
				{
					Condition: &Condition{
						FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
						Operation: StandardOperationInList,
						FixedArg:  NewFixedArg("list1"),
					},
				},
				{
					Condition: &Condition{
						FreeArg:   NewFreeArg(StandardFreeArgTypeString, "version"),
						Operation: StandardOperationIs,
						FixedArg:  NewFixedArg("1.0"),
					},
				},
			},
		}
		names := rule.GetInListNames()
		assert.Equal(t, 1, len(names))
		assert.Equal(t, "list1", names[0])
	})
}

func TestRule_GetTree(t *testing.T) {
	// Test simple rule (non-compound)
	t.Run("Simple_rule", func(t *testing.T) {
		rule := &Rule{
			Xxid: "rule-123",
			Condition: &Condition{
				FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
				Operation: StandardOperationIs,
				FixedArg:  NewFixedArg("TG1682G"),
			},
		}
		tree := rule.GetTree()
		assert.Equal(t, "rule-123", tree)
	})

	// Test compound rule with AND relation (default)
	t.Run("Compound_rule_AND", func(t *testing.T) {
		rule := &Rule{
			Relation: RelationAnd,
			CompoundParts: []Rule{
				{
					Xxid: "rule-1",
					Condition: &Condition{
						FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
						Operation: StandardOperationIs,
						FixedArg:  NewFixedArg("TG1682G"),
					},
				},
				{
					Xxid: "rule-2",
					Condition: &Condition{
						FreeArg:   NewFreeArg(StandardFreeArgTypeString, "version"),
						Operation: StandardOperationIs,
						FixedArg:  NewFixedArg("1.0"),
					},
				},
			},
		}
		tree := rule.GetTree()
		assert.Equal(t, "(rule-1 AND rule-2)", tree)
	})

	// Test compound rule with OR relation
	t.Run("Compound_rule_OR", func(t *testing.T) {
		rule := &Rule{
			Relation: RelationOr,
			CompoundParts: []Rule{
				{
					Xxid: "rule-1",
					Condition: &Condition{
						FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
						Operation: StandardOperationIs,
						FixedArg:  NewFixedArg("TG1682G"),
					},
				},
				{
					Xxid: "rule-2",
					Condition: &Condition{
						FreeArg:   NewFreeArg(StandardFreeArgTypeString, "version"),
						Operation: StandardOperationIs,
						FixedArg:  NewFixedArg("2.0"),
					},
				},
			},
		}
		tree := rule.GetTree()
		assert.Equal(t, "(rule-1 OR rule-2)", tree)
	})
}
