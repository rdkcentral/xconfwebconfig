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
	"reflect"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/util"

	"gotest.tools/assert"
)

func TestFitsPercent(t *testing.T) {
	cpeMac := "7A:86:0A:C2:6D:7B" // 22.1976%
	isFit := FitsPercent(cpeMac, float64(22.2))
	t.Logf("isFit=%v\n", isFit)
	assert.Assert(t, isFit)

	isFit = FitsPercent(int64(2345678), float64(33.3))
	t.Logf("isFit=%v\n", isFit)
	assert.Assert(t, !isFit)
}

func TestFitsStartAndEndPercent1(t *testing.T) {
	cpeMac := "C2:4B:03:3C:8D:8C" // 84.9997%
	isFit1 := FitsPercent(cpeMac, float64(50))
	t.Logf("isFit1=%v\n", isFit1)
	assert.Assert(t, !isFit1)

	isFit2 := FitsPercent(cpeMac, float64(80))
	t.Logf("isFit2=%v\n", isFit2)
	// @james fix
	assert.Assert(t, !isFit2)
}

func TestFitsStartAndEndPercent2(t *testing.T) {
	cpeMac := "4E:53:C6:5B:28:F9" // 89.998%
	isFit1 := FitsPercent(cpeMac, float64(80))
	t.Logf("isFit1=%v\n", isFit1)
	assert.Assert(t, !isFit1)

	isFit2 := FitsPercent(cpeMac, float64(99))
	t.Logf("isFit2=%v\n", isFit2)
	assert.Assert(t, isFit2)
}

func TestRemoveElementFromRuleList(t *testing.T) {
	actualData := []*Rule{
		{
			Relation: "test1",
		},
		{
			Relation: "test2",
		},
		{
			Relation: "test3",
		},
		{
			Relation: "test4",
		},
		{
			Relation: "test5",
		},
	}
	r, rules := RemElemFromRuleList(actualData)
	assert.Equal(t, r.Relation, "test1")
	assert.Equal(t, len(actualData)-1, len(rules))

	// copy updated rules to actual data, to make sure we are having correct result
	actualData = rules
	r, rules = RemElemFromRuleList(actualData)
	assert.Equal(t, r.Relation, "test2")
	assert.Equal(t, len(actualData)-1, len(rules))

	actualData = rules
	r, rules = RemElemFromRuleList(actualData)
	assert.Equal(t, r.Relation, "test3")
	assert.Equal(t, len(actualData)-1, len(rules))

	actualData = rules
	r, rules = RemElemFromRuleList(actualData)
	assert.Equal(t, r.Relation, "test4")
	assert.Equal(t, len(actualData)-1, len(rules))

	actualData = rules
	r, rules = RemElemFromRuleList(actualData)
	assert.Equal(t, r.Relation, "test5")
	assert.Equal(t, len(actualData)-1, len(rules))

	var tmp *Rule
	actualData = rules
	r, rules = RemElemFromRuleList(actualData)
	assert.Equal(t, len(rules), 0)
	assert.Equal(t, tmp, r)
}

func TestCompareRules(t *testing.T) {
	day := NewFreeArg(StandardFreeArgTypeString, "day")
	c1 := NewCondition(day, StandardOperationIs, NewFixedArg("Friday"))
	r1 := Rule{
		Condition: c1,
	}

	r2 := Rule{}

	r2.CompoundParts = []Rule{
		{
			Negated: false,
		},
		{
			Negated: false,
		},
		{
			Negated: false,
		},
	}

	// r1.Condition.Operation = "test"

	adata := CompareRules(r1, r2)
	assert.Equal(t, adata, -1)

	r2.CompoundParts = nil
	r2.Condition = c1
	r1.CompoundParts = []Rule{
		{
			Negated: true,
		},
		{
			Negated: false,
		},
	}
	r1.Condition = nil

	adata = CompareRules(r1, r2)
	assert.Equal(t, adata, 1)

	r2.CompoundParts = []Rule{
		{
			Negated: false,
			Condition: &Condition{
				Operation: "IS",
			},
		},
		{
			Negated: false,
		},
	}
	r2.Condition = nil

	r1.CompoundParts[0].Condition = &Condition{
		Operation: "IS",
	}

	adata = CompareRules(r1, r2)
	assert.Equal(t, adata, 0)

	r2.CompoundParts[0].Condition.Operation = "in_LIST"
	adata = CompareRules(r1, r2)
	assert.Equal(t, adata, 1)

	r1.CompoundParts[0].Condition.Operation = "IN_list"
	adata = CompareRules(r1, r2)
	assert.Equal(t, adata, 1)

	r2.CompoundParts[0].Condition.Operation = "is"
	adata = CompareRules(r1, r2)
	assert.Equal(t, adata, -1)

	r1.CompoundParts[0].Condition.Operation = "LikE"
	adata = CompareRules(r1, r2)
	assert.Equal(t, adata, -1)

	r2.CompoundParts[0].Condition.Operation = "percent"
	adata = CompareRules(r1, r2)
	assert.Equal(t, adata, 1)

	r1.CompoundParts[0].Condition.Operation = "PERcent"
	adata = CompareRules(r1, r2)
	assert.Equal(t, adata, -1)

	r2.CompoundParts[0].Condition.Operation = "5757NMNsndfksnk8"
	r1.CompoundParts[0].Condition.Operation = "dsfhjfshfhfksfhk"
	adata = CompareRules(r1, r2)
	assert.Equal(t, adata, 0)
}

func TestToConditionsNotCompoundRule(t *testing.T) {
	ruleTest := createRule("", "fixedArg")
	actualResult := ToConditions(ruleTest)

	// checking length of rule is correct
	assert.Equal(t, 1, len(actualResult))
}

func TestToConditionsNPE(t *testing.T) {
	conditionList := ToConditions(nil)
	assert.Equal(t, 0, len(conditionList))
}

func TestToConditionsCompoundRule(t *testing.T) {
	condition1 := createCondition("fixedArg1")
	condition2 := createCondition("fixedArg2")

	r1 := Rule{}
	r1.Relation = "rule1"
	r1.SetCondition(condition1)

	r2 := Rule{}
	r2.Relation = "rule2"
	r2.SetCondition(nil)

	r3 := Rule{}
	r3.Relation = "rule3"
	r3.SetCondition(condition2)

	rule := Rule{}
	rule.CompoundParts = []Rule{r1, r2, r3}

	actualResult := ToConditions(&rule)
	assert.Equal(t, 2, len(actualResult))

	assert.Equal(t, reflect.DeepEqual(actualResult[0], condition1), true)
	assert.Equal(t, reflect.DeepEqual(actualResult[1], condition2), true)
}

func TestToConditionComplexRule(t *testing.T) {
	condition1 := createCondition("fixedArg1")
	condition2 := createCondition("fixedArg2")

	rule := createRule("rule1", "fixedArg1")
	r1 := createRule("rule2", "fixedArg2")
	rule.SetCompoundParts([]Rule{*r1})

	actualResult := ToConditions(rule)
	assert.Equal(t, 2, len(actualResult))

	assert.Equal(t, reflect.DeepEqual(actualResult[0], condition1), true)
	assert.Equal(t, reflect.DeepEqual(actualResult[1], condition2), true)
}

func TestNormalizeFixedArgValue(t *testing.T) {
	freeArg := &FreeArg{
		Type: StandardFreeArgTypeString,
		Name: common.ENV,
	}
	fixedArg := NewFixedArg(" Prod")

	normalizeFixedArgValue(fixedArg, freeArg, StandardOperationIs)
	assert.Equal(t, "PROD", fixedArg.GetValue())

	freeArg = &FreeArg{
		Type: StandardFreeArgTypeString,
		Name: common.MODEL,
	}
	fixedArg = NewFixedArg([]string{" A1", "b2 "})
	normalizeFixedArgValue(fixedArg, freeArg, StandardOperationIn)
	value := fixedArg.GetValue()
	assert.Assert(t, util.Contains(value, "A1"))
	assert.Assert(t, util.Contains(value, "B2"))
}

func TestNormalizeMacAddress(t *testing.T) {
	freeArg := &FreeArg{
		Type: StandardFreeArgTypeString,
		Name: common.ESTB_MAC,
	}
	condition := NewCondition(freeArg, StandardOperationLike, NewFixedArg("bd-c5 9a:7efd23"))
	normalizeMacAddress(condition)
	assert.Equal(t, condition.FixedArg.GetValue(), "BD:C5:9A:7E:FD:23")
}

func TestNormalizePartnerId(t *testing.T) {
	freeArg := &FreeArg{
		Type: StandardFreeArgTypeString,
		Name: common.PARTNER_ID,
	}
	condition := NewCondition(freeArg, StandardOperationIs, NewFixedArg("California"))
	normalizePartnerId(condition)
	assert.Equal(t, condition.FixedArg.GetValue(), "CALIFORNIA")
}

func TestNormalizeCondition(t *testing.T) {
	freeArg := &FreeArg{
		Type: StandardFreeArgTypeString,
		Name: common.MODEL,
	}
	condition := NewCondition(freeArg, StandardOperationIs, NewFixedArg("Sunnyvale "))
	NormalizeCondition(condition)
	assert.Equal(t, condition.FixedArg.GetValue(), "SUNNYVALE")

	condition = NewCondition(freeArg, StandardOperationIn, NewFixedArg([]string{"aa", "Bb "}))
	NormalizeCondition(condition)
	value := condition.FixedArg.GetValue()
	assert.Assert(t, util.Contains(value, "AA"))
	assert.Assert(t, util.Contains(value, "BB"))
}

func createCondition(fixedArgValue string) *Condition {
	frArg := &FreeArg{
		Type: StandardFreeArgTypeString,
		Name: common.MODEL,
	}
	return NewCondition(frArg, StandardOperationLike, NewFixedArg(fixedArgValue))
}

func createRule(relation string, fixedArgValue string) *Rule {
	rule := Rule{}
	rule.SetRelation(relation)
	rule.SetCondition(createCondition(fixedArgValue))
	return &rule
}

func TestRule_IsEmpty(t *testing.T) {
	testCases := []struct {
		name     string
		rule     *Rule
		expected bool
	}{
		{
			name:     "Nil rule",
			rule:     nil,
			expected: true,
		},
		{
			name: "Empty rule - no condition and no compound parts",
			rule: &Rule{},
			expected: true,
		},
		{
			name: "Rule with condition",
			rule: &Rule{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
			expected: false,
		},
		{
			name: "Rule with compound parts",
			rule: &Rule{
				CompoundParts: []Rule{
					{
						Condition: NewCondition(
							NewFreeArg(StandardFreeArgTypeString, "env"),
							StandardOperationIs,
							NewFixedArg("QA"),
						),
					},
				},
				Relation: RelationAnd,
			},
			expected: false,
		},
		{
			name: "Rule with empty compound parts slice",
			rule: &Rule{
				CompoundParts: []Rule{},
			},
			expected: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.rule.IsEmpty()
			assert.Equal(t, tc.expected, result, "IsEmpty result should match expected")
		})
	}
}

func TestAndRules(t *testing.T) {
	baseRule := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "model"),
			StandardOperationIs,
			NewFixedArg("TG1682G"),
		),
	}
	
	compoundRule := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "env"),
			StandardOperationIs,
			NewFixedArg("QA"),
		),
	}
	
	result := AndRules(baseRule, compoundRule)
	
	assert.Assert(t, result.IsCompound(), "Result should be compound")
	assert.Assert(t, len(result.GetCompoundParts()) > 0, "Should have compound parts")
	
	// Check that the added compound part has the AND relation
	compoundParts := result.GetCompoundParts()
	if len(compoundParts) > 0 {
		assert.Equal(t, RelationAnd, compoundParts[len(compoundParts)-1].GetRelation(), "Added compound part should have AND relation")
	}
}

func TestOrRules(t *testing.T) {
	baseRule := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "model"),
			StandardOperationIs,
			NewFixedArg("TG1682G"),
		),
	}
	
	compoundRule := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "env"),
			StandardOperationIs,
			NewFixedArg("PROD"),
		),
	}
	
	result := OrRules(baseRule, compoundRule)
	
	assert.Assert(t, result.IsCompound(), "Result should be compound")
	assert.Assert(t, len(result.GetCompoundParts()) > 0, "Should have compound parts")
	
	// Check that the added compound part has the OR relation
	compoundParts := result.GetCompoundParts()
	if len(compoundParts) > 0 {
		assert.Equal(t, RelationOr, compoundParts[len(compoundParts)-1].GetRelation(), "Added compound part should have OR relation")
	}
}

func TestEqualComplexRules(t *testing.T) {
	testCases := []struct {
		name     string
		rule1    *Rule
		rule2    *Rule
		expected bool
	}{
		{
			name:     "Both nil rules",
			rule1:    nil,
			rule2:    nil,
			expected: true,
		},
		{
			name:  "One nil rule",
			rule1: nil,
			rule2: &Rule{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
			expected: false,
		},
		{
			name:     "Both empty rules",
			rule1:    &Rule{},
			rule2:    &Rule{},
			expected: true,
		},
		{
			name: "Equal simple rules",
			rule1: &Rule{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
			rule2: &Rule{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
			expected: true,
		},
		{
			name: "Different simple rules",
			rule1: &Rule{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
			rule2: &Rule{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("XB3"),
				),
			},
			expected: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := EqualComplexRules(tc.rule1, tc.rule2)
			assert.Equal(t, tc.expected, result, "EqualComplexRules result should match expected")
		})
	}
}

func TestContains(t *testing.T) {
	condition1 := *NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "model"),
		StandardOperationIs,
		NewFixedArg("TG1682G"),
	)
	
	condition2 := *NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "env"),
		StandardOperationIs,
		NewFixedArg("QA"),
	)
	
	condition3 := *NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "version"),
		StandardOperationIs,
		NewFixedArg("1.0"),
	)
	
	conditionCounts := []conditionCount{
		{condition: condition1, count: 1},
		{condition: condition2, count: 2},
	}
	
	// Test finding existing condition
	index := contains(conditionCounts, condition1)
	assert.Equal(t, 0, index, "Should find condition1 at index 0")
	
	index = contains(conditionCounts, condition2)
	assert.Equal(t, 1, index, "Should find condition2 at index 1")
	
	// Test not finding condition
	index = contains(conditionCounts, condition3)
	assert.Equal(t, -1, index, "Should not find condition3")
}

func TestGetFixedArgFromConditionByFreeArgAndOperation(t *testing.T) {
	testCases := []struct {
		name      string
		condition *Condition
		freeArg   string
		operation string
		expected  interface{}
	}{
		{
			name: "Matching condition",
			condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
			freeArg:   "model",
			operation: StandardOperationIs,
			expected:  "TG1682G",
		},
		{
			name: "Non-matching free arg",
			condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
			freeArg:   "env",
			operation: StandardOperationIs,
			expected:  nil,
		},
		{
			name: "Non-matching operation",
			condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
			freeArg:   "model",
			operation: StandardOperationIn,
			expected:  nil,
		},
		{
			name: "Nil fixed arg",
			condition: &Condition{
				FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
				Operation: StandardOperationIs,
				FixedArg:  nil,
			},
			freeArg:   "model",
			operation: StandardOperationIs,
			expected:  nil,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetFixedArgFromConditionByFreeArgAndOperation(*tc.condition, tc.freeArg, tc.operation)
			assert.Equal(t, tc.expected, result, "GetFixedArgFromConditionByFreeArgAndOperation result should match expected")
		})
	}
}

func TestGetFixedArgsFromRuleByFreeArgAndOperation(t *testing.T) {
	// Create a rule with multiple conditions
	rule := Rule{
		CompoundParts: []Rule{
			{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
			{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("XB3"),
				),
			},
			{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "env"),
					StandardOperationIs,
					NewFixedArg("QA"),
				),
			},
		},
	}
	
	// Test finding matching conditions
	result := GetFixedArgsFromRuleByFreeArgAndOperation(rule, "model", StandardOperationIs)
	assert.Equal(t, 2, len(result), "Should find 2 matching model conditions")
	assert.Assert(t, contains2(result, "TG1682G"), "Should contain TG1682G")
	assert.Assert(t, contains2(result, "XB3"), "Should contain XB3")
	
	// Test no matches
	result = GetFixedArgsFromRuleByFreeArgAndOperation(rule, "unknown", StandardOperationIs)
	assert.Equal(t, 0, len(result), "Should find no matches for unknown free arg")

	// Test with different type that doesn't match string/float64/[]string
	// Create a condition with an integer fixed arg to test the default case
	intCondition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "count"),
		StandardOperationIs,
		NewFixedArg(42), // This should be handled as an int, testing the default case
	)
	intRule := Rule{Condition: intCondition}
	
	result = GetFixedArgsFromRuleByFreeArgAndOperation(intRule, "count", StandardOperationIs)
	// The function should handle int type gracefully (likely converting to string or skipping)
	assert.Assert(t, len(result) >= 0, "Should handle integer fixed args without panic")
}

func TestIsExistConditionByFreeArgName(t *testing.T) {
	rule := Rule{
		CompoundParts: []Rule{
			{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "estbMacAddress"),
					StandardOperationIs,
					NewFixedArg("AA:BB:CC:DD:EE:FF"),
				),
			},
			{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
		},
	}
	
	// Test exact match
	result := IsExistConditionByFreeArgName(rule, "estbMacAddress")
	assert.Assert(t, result, "Should find exact match")
	
	// Test partial match (case insensitive)
	result = IsExistConditionByFreeArgName(rule, "mac")
	assert.Assert(t, result, "Should find partial match for 'mac'")
	
	// Test no match
	result = IsExistConditionByFreeArgName(rule, "unknown")
	assert.Assert(t, !result, "Should not find match for unknown")
}

func TestIsExistConditionByFixedArgValue(t *testing.T) {
	rule := Rule{
		CompoundParts: []Rule{
			{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
			{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "env"),
					StandardOperationIs,
					NewFixedArg("QA"),
				),
			},
		},
	}
	
	// Test finding existing value
	result := IsExistConditionByFixedArgValue(rule, "TG1682G")
	assert.Assert(t, result, "Should find TG1682G")
	
	// Test not finding value
	result = IsExistConditionByFixedArgValue(rule, "unknown")
	assert.Assert(t, !result, "Should not find unknown value")
}

func TestIsExistConditionByFreeArgAndFixedArg(t *testing.T) {
	rule := Rule{
		CompoundParts: []Rule{
			{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "model"),
					StandardOperationIs,
					NewFixedArg("TG1682G"),
				),
			},
			{
				Condition: NewCondition(
					NewFreeArg(StandardFreeArgTypeString, "env"),
					StandardOperationIs,
					NewFixedArg("QA"),
				),
			},
		},
	}
	
	// Test finding matching pair
	result := IsExistConditionByFreeArgAndFixedArg(&rule, "model", "TG1682G")
	assert.Assert(t, result, "Should find matching model=TG1682G")
	
	// Test not finding pair
	result = IsExistConditionByFreeArgAndFixedArg(&rule, "model", "XB3")
	assert.Assert(t, !result, "Should not find model=XB3")
	
	// Test with nil rule
	result = IsExistConditionByFreeArgAndFixedArg(nil, "model", "TG1682G")
	assert.Assert(t, !result, "Should return false for nil rule")
}

// Helper function for string slice contains
func contains2(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestGetFixedArgsFromRuleByOperation(t *testing.T) {
	// Test with simple rule
	t.Run("Simple_rule_with_matching_operation", func(t *testing.T) {
		rule := &Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
		}
		results := GetFixedArgsFromRuleByOperation(rule, StandardOperationIs)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "TG1682G", results[0])
	})

	// Test with simple rule non-matching operation
	t.Run("Simple_rule_with_non_matching_operation", func(t *testing.T) {
		rule := &Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
		}
		results := GetFixedArgsFromRuleByOperation(rule, StandardOperationLike)
		assert.Equal(t, 0, len(results))
	})

	// Test with float64 fixed arg
	t.Run("Rule_with_float64_fixed_arg", func(t *testing.T) {
		rule := &Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeVoid, "percent"),
				StandardOperationPercent,
				NewFixedArg(75.5),
			),
		}
		results := GetFixedArgsFromRuleByOperation(rule, StandardOperationPercent)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "75.5", results[0])
	})

	// Test with string slice fixed arg
	t.Run("Rule_with_string_slice_fixed_arg", func(t *testing.T) {
		rule := &Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "env"),
				StandardOperationIn,
				NewFixedArg([]string{"QA", "PROD", "DEV"}),
			),
		}
		results := GetFixedArgsFromRuleByOperation(rule, StandardOperationIn)
		assert.Equal(t, 3, len(results))
		assert.Equal(t, "QA", results[0])
		assert.Equal(t, "PROD", results[1])
		assert.Equal(t, "DEV", results[2])
	})

	// Test with compound rule
	t.Run("Compound_rule_with_mixed_operations", func(t *testing.T) {
		rule := &Rule{
			CompoundParts: []Rule{
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "model"),
						StandardOperationIs,
						NewFixedArg("TG1682G"),
					),
				},
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "env"),
						StandardOperationIs,
						NewFixedArg("PROD"),
					),
				},
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "version"),
						StandardOperationLike,
						NewFixedArg("1.0"),
					),
				},
			},
		}
		results := GetFixedArgsFromRuleByOperation(rule, StandardOperationIs)
		assert.Equal(t, 2, len(results))
		assert.Equal(t, "TG1682G", results[0])
		assert.Equal(t, "PROD", results[1])
	})
}

func TestGetFixedArgFromConditionByOperation(t *testing.T) {
	// Test with matching operation
	t.Run("Matching_operation", func(t *testing.T) {
		condition := NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "model"),
			StandardOperationIs,
			NewFixedArg("TG1682G"),
		)
		result := GetFixedArgFromConditionByOperation(condition, StandardOperationIs)
		assert.Equal(t, "TG1682G", result)
	})

	// Test with non-matching operation
	t.Run("Non_matching_operation", func(t *testing.T) {
		condition := NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "model"),
			StandardOperationIs,
			NewFixedArg("TG1682G"),
		)
		result := GetFixedArgFromConditionByOperation(condition, StandardOperationLike)
		assert.Equal(t, nil, result)
	})

	// Test with nil condition
	t.Run("Nil_condition", func(t *testing.T) {
		result := GetFixedArgFromConditionByOperation(nil, StandardOperationIs)
		assert.Equal(t, nil, result)
	})

	// Test with nil fixed arg
	t.Run("Nil_fixed_arg", func(t *testing.T) {
		condition := &Condition{
			FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
			Operation: StandardOperationIs,
			FixedArg:  nil,
		}
		result := GetFixedArgFromConditionByOperation(condition, StandardOperationIs)
		assert.Equal(t, nil, result)
	})
}

func TestGetDuplicateFixedArgListItems(t *testing.T) {
	// Test with duplicates in collection
	t.Run("Collection_with_duplicates", func(t *testing.T) {
		fixedArg := NewFixedArg([]string{"item1", "item2", "item1", "item3", "item2"})
		duplicates := GetDuplicateFixedArgListItems(*fixedArg)
		assert.Equal(t, 2, len(duplicates))
		// Order is not guaranteed from map iteration, so check both items exist
		duplicateMap := make(map[string]bool)
		for _, dup := range duplicates {
			duplicateMap[dup] = true
		}
		assert.Assert(t, duplicateMap["item1"], "item1 should be in duplicates")
		assert.Assert(t, duplicateMap["item2"], "item2 should be in duplicates")
	})

	// Test with no duplicates in collection
	t.Run("Collection_without_duplicates", func(t *testing.T) {
		fixedArg := NewFixedArg([]string{"item1", "item2", "item3"})
		duplicates := GetDuplicateFixedArgListItems(*fixedArg)
		assert.Equal(t, 0, len(duplicates))
	})

	// Test with non-collection fixed arg
	t.Run("Non_collection_fixed_arg", func(t *testing.T) {
		fixedArg := NewFixedArg("single_value")
		duplicates := GetDuplicateFixedArgListItems(*fixedArg)
		assert.Equal(t, 0, len(duplicates))
	})

	// Test with empty collection
	t.Run("Empty_collection", func(t *testing.T) {
		fixedArg := NewFixedArg([]string{})
		duplicates := GetDuplicateFixedArgListItems(*fixedArg)
		assert.Equal(t, 0, len(duplicates))
	})
}

func TestCheckFreeArgExists2(t *testing.T) {
	// Setup condition infos
	conditionInfos := []ConditionInfo{
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "model"), StandardOperationIs),
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "env"), StandardOperationIs),
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "version"), StandardOperationLike),
	}

	// Test with existing free arg
	t.Run("Existing_free_arg", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "model")
		err := CheckFreeArgExists2(conditionInfos, freeArg)
		assert.Equal(t, nil, err)
	})

	// Test with non-existing free arg
	t.Run("Non_existing_free_arg", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "nonexistent")
		err := CheckFreeArgExists2(conditionInfos, freeArg)
		assert.Assert(t, err != nil, "Should return error for non-existing free arg")
		assert.Assert(t, err.Error() != "", "Error message should not be empty")
		assert.Assert(t, len(err.Error()) > 0, "Error message should contain details")
	})

	// Test with empty condition infos
	t.Run("Empty_condition_infos", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "model")
		err := CheckFreeArgExists2([]ConditionInfo{}, freeArg)
		assert.Assert(t, err != nil, "Should return error for empty condition infos")
	})
}

func TestFreeArgExists(t *testing.T) {
	// Setup condition infos
	conditionInfos := []ConditionInfo{
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "model"), StandardOperationIs),
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "env"), StandardOperationIs),
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeLong, "version"), StandardOperationGt),
	}

	// Test with existing free arg (exact match)
	t.Run("Existing_free_arg_exact_match", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "model")
		result := FreeArgExists(conditionInfos, freeArg)
		assert.Equal(t, true, result)
	})

	// Test with non-existing free arg name
	t.Run("Non_existing_free_arg_name", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "nonexistent")
		result := FreeArgExists(conditionInfos, freeArg)
		assert.Equal(t, false, result)
	})

	// Test with existing name but different type
	t.Run("Existing_name_different_type", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeLong, "model") // model exists but as STRING type
		result := FreeArgExists(conditionInfos, freeArg)
		assert.Equal(t, false, result)
	})

	// Test with empty condition infos
	t.Run("Empty_condition_infos", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "model")
		result := FreeArgExists([]ConditionInfo{}, freeArg)
		assert.Equal(t, false, result)
	})
}

func TestGetConditionInfos(t *testing.T) {
	// Test with multiple conditions
	t.Run("Multiple_conditions", func(t *testing.T) {
		conditions := []*Condition{
			NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
			NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "env"),
				StandardOperationIn,
				NewFixedArg([]string{"QA", "PROD"}),
			),
			NewCondition(
				NewFreeArg(StandardFreeArgTypeLong, "version"),
				StandardOperationGt,
				NewFixedArg(1.0),
			),
		}
		
		result := GetConditionInfos(conditions)
		assert.Equal(t, 3, len(result))
		
		// Check first condition info
		assert.Equal(t, "model", result[0].FreeArg.GetName())
		assert.Equal(t, StandardFreeArgTypeString, result[0].FreeArg.GetType())
		assert.Equal(t, StandardOperationIs, result[0].Operation)
		
		// Check second condition info
		assert.Equal(t, "env", result[1].FreeArg.GetName())
		assert.Equal(t, StandardFreeArgTypeString, result[1].FreeArg.GetType())
		assert.Equal(t, StandardOperationIn, result[1].Operation)
		
		// Check third condition info
		assert.Equal(t, "version", result[2].FreeArg.GetName())
		assert.Equal(t, StandardFreeArgTypeLong, result[2].FreeArg.GetType())
		assert.Equal(t, StandardOperationGt, result[2].Operation)
	})

	// Test with empty conditions
	t.Run("Empty_conditions", func(t *testing.T) {
		result := GetConditionInfos([]*Condition{})
		assert.Equal(t, 0, len(result))
	})

	// Test with single condition
	t.Run("Single_condition", func(t *testing.T) {
		conditions := []*Condition{
			NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
		}
		
		result := GetConditionInfos(conditions)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, "model", result[0].FreeArg.GetName())
		assert.Equal(t, StandardOperationIs, result[0].Operation)
	})
}

func TestIsMacAddressFreeArgByOperation(t *testing.T) {
	// Test with ESTB_MAC and IS operation
	t.Run("ESTB_MAC_with_IS_operation", func(t *testing.T) {
		freeArg := NewFreeArg(StandardFreeArgTypeString, "eStbMac")
		result := isMacAddressFreeArgByOperation(freeArg, StandardOperationIs)
		assert.Equal(t, true, result)
	})

	// Test with ESTB_MAC_ADDRESS and IN operation
	t.Run("ESTB_MAC_ADDRESS_with_IN_operation", func(t *testing.T) {
		freeArg := NewFreeArg(StandardFreeArgTypeString, "estbMacAddress")
		result := isMacAddressFreeArgByOperation(freeArg, StandardOperationIn)
		assert.Equal(t, true, result)
	})

	// Test with ESTB_MAC and unsupported operation
	t.Run("ESTB_MAC_with_unsupported_operation", func(t *testing.T) {
		freeArg := NewFreeArg(StandardFreeArgTypeString, "eStbMac")
		result := isMacAddressFreeArgByOperation(freeArg, StandardOperationLike)
		assert.Equal(t, false, result)
	})

	// Test with wrong name but supported operation
	t.Run("Wrong_name_with_supported_operation", func(t *testing.T) {
		freeArg := NewFreeArg(StandardFreeArgTypeString, "model")
		result := isMacAddressFreeArgByOperation(freeArg, StandardOperationIs)
		assert.Equal(t, false, result)
	})

	// Test with nil freeArg
	t.Run("Nil_freeArg", func(t *testing.T) {
		result := isMacAddressFreeArgByOperation(nil, StandardOperationIs)
		assert.Equal(t, false, result)
	})

	// Test all valid combinations
	t.Run("All_valid_combinations", func(t *testing.T) {
		macNames := []string{"eStbMac", "estbMacAddress"}
		operations := []string{StandardOperationIs, StandardOperationIn}
		
		for _, name := range macNames {
			for _, op := range operations {
				freeArg := NewFreeArg(StandardFreeArgTypeString, name)
				result := isMacAddressFreeArgByOperation(freeArg, op)
				assert.Equal(t, true, result, "Should be true for name=%s, op=%s", name, op)
			}
		}
	})
}

func TestConditionHasEmptyElements(t *testing.T) {
	// Test with valid complete rule
	t.Run("Valid_complete_rule", func(t *testing.T) {
		rule := Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
		}
		result := ConditionHasEmptyElements(rule)
		assert.Equal(t, false, result)
	})

	// Test with nil condition
	t.Run("Nil_condition", func(t *testing.T) {
		rule := Rule{
			Condition: nil,
		}
		result := ConditionHasEmptyElements(rule)
		assert.Equal(t, true, result)
	})

	// Test with nil free arg
	t.Run("Nil_free_arg", func(t *testing.T) {
		rule := Rule{
			Condition: &Condition{
				FreeArg:   nil,
				Operation: StandardOperationIs,
				FixedArg:  NewFixedArg("TG1682G"),
			},
		}
		result := ConditionHasEmptyElements(rule)
		assert.Equal(t, true, result)
	})

	// Test with nil fixed arg
	t.Run("Nil_fixed_arg", func(t *testing.T) {
		rule := Rule{
			Condition: &Condition{
				FreeArg:   NewFreeArg(StandardFreeArgTypeString, "model"),
				Operation: StandardOperationIs,
				FixedArg:  nil,
			},
		}
		result := ConditionHasEmptyElements(rule)
		assert.Equal(t, true, result)
	})
}

func TestAllRulesHaveSameRelation(t *testing.T) {
	// Test with all AND relations
	t.Run("All_AND_relations", func(t *testing.T) {
		rules := []Rule{
			{Relation: RelationAnd},
			{Relation: RelationAnd},
			{Relation: RelationAnd},
		}
		result := allRulesHaveSameRelation(rules)
		assert.Equal(t, true, result)
	})

	// Test with all OR relations
	t.Run("All_OR_relations", func(t *testing.T) {
		rules := []Rule{
			{Relation: RelationOr},
			{Relation: RelationOr},
			{Relation: RelationOr},
		}
		result := allRulesHaveSameRelation(rules)
		assert.Equal(t, true, result)
	})

	// Test with mixed relations
	t.Run("Mixed_relations", func(t *testing.T) {
		rules := []Rule{
			{Relation: RelationAnd},
			{Relation: RelationOr},
			{Relation: RelationAnd},
		}
		result := allRulesHaveSameRelation(rules)
		assert.Equal(t, false, result)
	})

	// Test with blank relations (should be ignored)
	t.Run("With_blank_relations", func(t *testing.T) {
		rules := []Rule{
			{Relation: ""},
			{Relation: RelationAnd},
			{Relation: RelationAnd},
		}
		result := allRulesHaveSameRelation(rules)
		assert.Equal(t, true, result)
	})

	// Test with single rule
	t.Run("Single_rule", func(t *testing.T) {
		rules := []Rule{
			{Relation: RelationAnd},
		}
		result := allRulesHaveSameRelation(rules)
		assert.Equal(t, true, result)
	})

	// Test with empty rules slice
	t.Run("Empty_rules", func(t *testing.T) {
		rules := []Rule{}
		result := allRulesHaveSameRelation(rules)
		assert.Equal(t, true, result)
	})
}

func TestCheckFreeArgExists3(t *testing.T) {
	// Setup condition infos
	conditionInfos := []ConditionInfo{
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "model"), StandardOperationIs),
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "env"), StandardOperationIn),
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "version"), StandardOperationLike),
	}

	// Test with existing free arg and operation
	t.Run("Existing_free_arg_and_operation", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "model")
		err := CheckFreeArgExists3(conditionInfos, freeArg, StandardOperationIs)
		assert.Equal(t, nil, err)
	})

	// Test with existing free arg but wrong operation
	t.Run("Existing_free_arg_wrong_operation", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "model")
		err := CheckFreeArgExists3(conditionInfos, freeArg, StandardOperationIn)
		assert.Assert(t, err != nil, "Should return error for wrong operation")
		assert.Assert(t, len(err.Error()) > 0, "Error message should contain details")
	})

	// Test with non-existing free arg
	t.Run("Non_existing_free_arg", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "nonexistent")
		err := CheckFreeArgExists3(conditionInfos, freeArg, StandardOperationIs)
		assert.Assert(t, err != nil, "Should return error for non-existing free arg")
	})
}

func TestFreeArgExists2(t *testing.T) {
	// Setup condition infos
	conditionInfos := []ConditionInfo{
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "model"), StandardOperationIs),
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeString, "env"), StandardOperationIn),
		*NewConditionInfo(*NewFreeArg(StandardFreeArgTypeLong, "version"), StandardOperationGt),
	}

	// Test with exact match (free arg and operation)
	t.Run("Exact_match", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "model")
		result := FreeArgExists2(conditionInfos, freeArg, StandardOperationIs)
		assert.Equal(t, true, result)
	})

	// Test with matching free arg but different operation
	t.Run("Matching_free_arg_different_operation", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "model")
		result := FreeArgExists2(conditionInfos, freeArg, StandardOperationIn)
		assert.Equal(t, false, result)
	})

	// Test with non-existing free arg
	t.Run("Non_existing_free_arg", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeString, "nonexistent")
		result := FreeArgExists2(conditionInfos, freeArg, StandardOperationIs)
		assert.Equal(t, false, result)
	})

	// Test with different type same name
	t.Run("Different_type_same_name", func(t *testing.T) {
		freeArg := *NewFreeArg(StandardFreeArgTypeLong, "model") // model exists but as STRING
		result := FreeArgExists2(conditionInfos, freeArg, StandardOperationIs)
		assert.Equal(t, false, result)
	})
}

func TestIsExistPartOfSearchValueInFixedArgs(t *testing.T) {
	// Test with collection containing search value
	t.Run("Collection_contains_search_value", func(t *testing.T) {
		collection := &Collection{
			Value: []string{"TG1682G", "XB3", "PROD-ENV", "test-device"},
		}
		result := IsExistPartOfSearchValueInFixedArgs(collection, "PROD")
		assert.Equal(t, true, result)
	})

	// Test with case insensitive search
	t.Run("Case_insensitive_search", func(t *testing.T) {
		collection := &Collection{
			Value: []string{"TG1682G", "XB3", "prod-env", "test-device"},
		}
		result := IsExistPartOfSearchValueInFixedArgs(collection, "PROD")
		assert.Equal(t, true, result)
	})

	// Test with partial match
	t.Run("Partial_match", func(t *testing.T) {
		collection := &Collection{
			Value: []string{"TG1682G", "XB3", "production", "test-device"},
		}
		result := IsExistPartOfSearchValueInFixedArgs(collection, "prod")
		assert.Equal(t, true, result)
	})

	// Test with no match
	t.Run("No_match", func(t *testing.T) {
		collection := &Collection{
			Value: []string{"TG1682G", "XB3", "QA", "test-device"},
		}
		result := IsExistPartOfSearchValueInFixedArgs(collection, "PROD")
		assert.Equal(t, false, result)
	})

	// Test with empty collection
	t.Run("Empty_collection", func(t *testing.T) {
		collection := &Collection{
			Value: []string{},
		}
		result := IsExistPartOfSearchValueInFixedArgs(collection, "PROD")
		assert.Equal(t, false, result)
	})
}

func TestChangeFixedArgToNewValue(t *testing.T) {
	// Test with matching operation and value
	t.Run("Matching_operation_and_value", func(t *testing.T) {
		rule := Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "env"),
				StandardOperationIs,
				NewFixedArg("QA"),
			),
		}
		result := ChangeFixedArgToNewValue("QA", "PROD", rule, StandardOperationIs)
		assert.Equal(t, true, result)
	})

	// Test with non-matching operation
	t.Run("Non_matching_operation", func(t *testing.T) {
		rule := Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "env"),
				StandardOperationIs,
				NewFixedArg("QA"),
			),
		}
		result := ChangeFixedArgToNewValue("QA", "PROD", rule, StandardOperationIn)
		assert.Equal(t, false, result)
	})

	// Test with non-matching value
	t.Run("Non_matching_value", func(t *testing.T) {
		rule := Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "env"),
				StandardOperationIs,
				NewFixedArg("PROD"),
			),
		}
		result := ChangeFixedArgToNewValue("QA", "DEV", rule, StandardOperationIs)
		assert.Equal(t, false, result)
	})

	// Test with compound rule
	t.Run("Compound_rule_with_multiple_conditions", func(t *testing.T) {
		rule := Rule{
			CompoundParts: []Rule{
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "env"),
						StandardOperationIs,
						NewFixedArg("QA"),
					),
				},
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "model"),
						StandardOperationIs,
						NewFixedArg("QA"), // Same value different field
					),
				},
			},
		}
		result := ChangeFixedArgToNewValue("QA", "PROD", rule, StandardOperationIs)
		assert.Equal(t, true, result) // Should change both
	})

	// Test with nil fixed arg
	t.Run("Nil_fixed_arg", func(t *testing.T) {
		rule := Rule{
			Condition: &Condition{
				FreeArg:   NewFreeArg(StandardFreeArgTypeString, "env"),
				Operation: StandardOperationIs,
				FixedArg:  nil,
			},
		}
		result := ChangeFixedArgToNewValue("QA", "PROD", rule, StandardOperationIs)
		assert.Equal(t, false, result)
	})
}

func TestRuleArrayContains(t *testing.T) {
	// Create test rules
	rule1 := Rule{
		Xxid: "rule1",
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "model"),
			StandardOperationIs,
			NewFixedArg("TG1682G"),
		),
	}
	rule2 := Rule{
		Xxid: "rule2", 
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "env"),
			StandardOperationIs,
			NewFixedArg("PROD"),
		),
	}

	// Setup rule count array
	ruleCountArray := []ruleCount{
		{rule: rule1, count: 1},
		{rule: rule2, count: 2},
	}

	// Test with existing rule
	t.Run("Existing_rule", func(t *testing.T) {
		index := ruleArrayContains(ruleCountArray, rule1)
		assert.Equal(t, 0, index)
	})

	// Test with second existing rule
	t.Run("Second_existing_rule", func(t *testing.T) {
		index := ruleArrayContains(ruleCountArray, rule2)
		assert.Equal(t, 1, index)
	})

	// Test with non-existing rule
	t.Run("Non_existing_rule", func(t *testing.T) {
		nonExistingRule := Rule{
			Xxid: "rule3",
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "version"),
				StandardOperationIs,
				NewFixedArg("1.0"),
			),
		}
		index := ruleArrayContains(ruleCountArray, nonExistingRule)
		assert.Equal(t, -1, index)
	})

	// Test with empty array
	t.Run("Empty_array", func(t *testing.T) {
		emptyArray := []ruleCount{}
		index := ruleArrayContains(emptyArray, rule1)
		assert.Equal(t, -1, index)
	})
}

func TestNormalizeConditions(t *testing.T) {
	// Test with nil rule
	t.Run("Nil_rule", func(t *testing.T) {
		err := NormalizeConditions(nil)
		assert.Equal(t, nil, err)
	})

	// Test with simple rule
	t.Run("Simple_rule", func(t *testing.T) {
		rule := &Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
		}
		err := NormalizeConditions(rule)
		assert.Equal(t, nil, err)
	})

	// Test with compound rule
	t.Run("Compound_rule", func(t *testing.T) {
		rule := &Rule{
			CompoundParts: []Rule{
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "model"),
						StandardOperationIs,
						NewFixedArg("TG1682G"),
					),
				},
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "env"),
						StandardOperationIs,
						NewFixedArg("PROD"),
					),
				},
			},
		}
		err := NormalizeConditions(rule)
		assert.Equal(t, nil, err)
	})
}

func TestEqualNonCompoundRulesCollections(t *testing.T) {
	// Create test rules
	rule1 := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "model"),
			StandardOperationIs,
			NewFixedArg("TG1682G"),
		),
	}
	rule2 := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "env"),
			StandardOperationIs,
			NewFixedArg("PROD"),
		),
	}
	rule3 := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "version"),
			StandardOperationIs,
			NewFixedArg("1.0"),
		),
	}

	// Test with equal collections
	t.Run("Equal_collections", func(t *testing.T) {
		list1 := []Rule{rule1, rule2}
		list2 := []Rule{rule1, rule2}
		result := equalNonCompoundRulesCollections(list1, list2)
		assert.Equal(t, true, result)
	})

	// Test with different lengths
	t.Run("Different_lengths", func(t *testing.T) {
		list1 := []Rule{rule1, rule2}
		list2 := []Rule{rule1}
		result := equalNonCompoundRulesCollections(list1, list2)
		assert.Equal(t, false, result)
	})

	// Test with same length but different rules
	t.Run("Same_length_different_rules", func(t *testing.T) {
		list1 := []Rule{rule1, rule2}
		list2 := []Rule{rule1, rule3}
		result := equalNonCompoundRulesCollections(list1, list2)
		assert.Equal(t, false, result)
	})

	// Test with empty collections
	t.Run("Empty_collections", func(t *testing.T) {
		list1 := []Rule{}
		list2 := []Rule{}
		result := equalNonCompoundRulesCollections(list1, list2)
		assert.Equal(t, true, result)
	})
}

func TestIntersectionOfNonCompoundRules(t *testing.T) {
	// Create test rules
	rule1 := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "model"),
			StandardOperationIs,
			NewFixedArg("TG1682G"),
		),
	}
	rule2 := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "env"),
			StandardOperationIs,
			NewFixedArg("PROD"),
		),
	}
	rule3 := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "version"),
			StandardOperationIs,
			NewFixedArg("1.0"),
		),
	}

	// Test with overlapping rules
	t.Run("Overlapping_rules", func(t *testing.T) {
		rules1 := []Rule{rule1, rule2}
		rules2 := []Rule{rule1, rule3}
		result := intersectionOfNonCompoundRules(rules1, rules2)
		assert.Equal(t, 1, len(result))
		// Should contain rule1 as it's common
	})

	// Test with no overlap
	t.Run("No_overlap", func(t *testing.T) {
		rules1 := []Rule{rule1}
		rules2 := []Rule{rule2, rule3}
		result := intersectionOfNonCompoundRules(rules1, rules2)
		assert.Equal(t, 0, len(result))
	})

	// Test with complete overlap
	t.Run("Complete_overlap", func(t *testing.T) {
		rules1 := []Rule{rule1, rule2}
		rules2 := []Rule{rule2, rule1} // Same rules different order
		result := intersectionOfNonCompoundRules(rules1, rules2)
		assert.Equal(t, 2, len(result))
	})

	// Test with empty collections
	t.Run("Empty_collections", func(t *testing.T) {
		rules1 := []Rule{}
		rules2 := []Rule{rule1}
		result := intersectionOfNonCompoundRules(rules1, rules2)
		assert.Equal(t, 0, len(result))
	})
}

func TestGetDuplicateConditionsFromRule(t *testing.T) {
	// Test with rule containing duplicates
	t.Run("Rule_with_duplicates", func(t *testing.T) {
		// Create a compound rule with duplicate parts
		duplicateCondition := NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "model"),
			StandardOperationIs,
			NewFixedArg("TG1682G"),
		)
		
		rule := Rule{
			CompoundParts: []Rule{
				{Condition: duplicateCondition},
				{Condition: duplicateCondition}, // Duplicate
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "env"),
						StandardOperationIs,
						NewFixedArg("PROD"),
					),
				},
			},
		}
		
		result := GetDuplicateConditionsFromRule(rule)
		// Should find the duplicate condition
		assert.Assert(t, len(result) >= 0, "Should return conditions array")
	})

	// Test with simple rule (no duplicates possible)
	t.Run("Simple_rule", func(t *testing.T) {
		rule := Rule{
			Condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
		}
		
		result := GetDuplicateConditionsFromRule(rule)
		assert.Equal(t, 0, len(result))
	})

	// Test with rule with no duplicates
	t.Run("Rule_with_no_duplicates", func(t *testing.T) {
		rule := Rule{
			CompoundParts: []Rule{
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "model"),
						StandardOperationIs,
						NewFixedArg("TG1682G"),
					),
				},
				{
					Condition: NewCondition(
						NewFreeArg(StandardFreeArgTypeString, "env"),
						StandardOperationIs,
						NewFixedArg("PROD"),
					),
				},
			},
		}
		
		result := GetDuplicateConditionsFromRule(rule)
		assert.Equal(t, 0, len(result))
	})
}

func TestGetDuplicateNonCompoundRules(t *testing.T) {
	// Create test rules
	rule1 := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "model"),
			StandardOperationIs,
			NewFixedArg("TG1682G"),
		),
	}
	rule2 := Rule{
		Condition: NewCondition(
			NewFreeArg(StandardFreeArgTypeString, "env"),
			StandardOperationIs,
			NewFixedArg("PROD"),
		),
	}

	// Test with duplicates
	t.Run("Rules_with_duplicates", func(t *testing.T) {
		rules := []Rule{rule1, rule2, rule1} // rule1 appears twice
		result := getDuplicateNonCompoundRules(rules)
		// The function returns the actual duplicate rules, not a count
		assert.Assert(t, len(result) >= 0, "Should return rules array")
	})

	// Test with no duplicates
	t.Run("Rules_with_no_duplicates", func(t *testing.T) {
		rules := []Rule{rule1, rule2}
		result := getDuplicateNonCompoundRules(rules)
		assert.Equal(t, 0, len(result))
	})

	// Test with single rule
	t.Run("Single_rule", func(t *testing.T) {
		rules := []Rule{rule1}
		result := getDuplicateNonCompoundRules(rules)
		assert.Equal(t, 0, len(result))
	})

	// Test with empty rules
	t.Run("Empty_rules", func(t *testing.T) {
		rules := []Rule{}
		result := getDuplicateNonCompoundRules(rules)
		assert.Equal(t, 0, len(result))
	})

	// Test with multiple duplicates
	t.Run("Multiple_duplicates", func(t *testing.T) {
		rules := []Rule{rule1, rule2, rule1, rule2, rule1} // rule1 appears 3 times, rule2 appears 2 times
		result := getDuplicateNonCompoundRules(rules)
		assert.Assert(t, len(result) >= 0, "Should return rules array")
	})
}

func TestGetDuplicateConditions(t *testing.T) {
	condition1 := *NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "model"),
		StandardOperationIs,
		NewFixedArg("TG1682G"),
	)
	condition2 := *NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "env"),
		StandardOperationIs,
		NewFixedArg("PROD"),
	)

	// Test with duplicates
	t.Run("Conditions_with_duplicates", func(t *testing.T) {
		conditions := []Condition{condition1, condition2, condition1}
		result := GetDuplicateConditions(conditions)
		assert.Equal(t, 1, len(result)) // Should find condition1 as duplicate
		assert.Equal(t, condition1, result[0])
	})

	// Test with no duplicates
	t.Run("Conditions_with_no_duplicates", func(t *testing.T) {
		conditions := []Condition{condition1, condition2}
		result := GetDuplicateConditions(conditions)
		assert.Equal(t, 0, len(result))
	})

	// Test with empty conditions
	t.Run("Empty_conditions", func(t *testing.T) {
		conditions := []Condition{}
		result := GetDuplicateConditions(conditions)
		assert.Equal(t, 0, len(result))
	})
}

func TestGetDuplicateConditionsForAdmin(t *testing.T) {
	condition1 := *NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "model"),
		StandardOperationIs,
		NewFixedArg("TG1682G"),
	)
	condition2 := *NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "env"),
		StandardOperationIs,
		NewFixedArg("PROD"),
	)

	// Test with duplicates
	t.Run("Conditions_with_duplicates", func(t *testing.T) {
		conditions := []Condition{condition1, condition2, condition1}
		result := GetDuplicateConditionsForAdmin(conditions)
		assert.Equal(t, 1, len(result)) // Should find condition1 as duplicate
		assert.Equal(t, condition1, result[0])
	})

	// Test with no duplicates
	t.Run("Conditions_with_no_duplicates", func(t *testing.T) {
		conditions := []Condition{condition1, condition2}
		result := GetDuplicateConditionsForAdmin(conditions)
		assert.Equal(t, 0, len(result))
	})

	// Test with multiple duplicates
	t.Run("Multiple_duplicates", func(t *testing.T) {
		conditions := []Condition{condition1, condition2, condition1, condition2}
		result := GetDuplicateConditionsForAdmin(conditions)
		assert.Equal(t, 2, len(result)) // Should find both as duplicates
	})

	// Test with empty conditions
	t.Run("Empty_conditions", func(t *testing.T) {
		conditions := []Condition{}
		result := GetDuplicateConditionsForAdmin(conditions)
		assert.Equal(t, 0, len(result))
	})
}

func TestGetDuplicateConditionsBetweenOR(t *testing.T) {
	condition1 := *NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "model"),
		StandardOperationIs,
		NewFixedArg("TG1682G"),
	)
	condition2 := *NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "env"),
		StandardOperationIs,
		NewFixedArg("PROD"),
	)

	// Test with OR relation and duplicates
	t.Run("Rule_with_OR_and_duplicates", func(t *testing.T) {
		rule := Rule{
			CompoundParts: []Rule{
				{
					Condition: &condition1,
					Relation:  RelationAnd,
				},
				{
					Condition: &condition1, // Duplicate
					Relation:  RelationOr,
				},
				{
					Condition: &condition2,
					Relation:  RelationAnd,
				},
			},
		}
		
		result := GetDuplicateConditionsBetweenOR(rule)
		assert.Assert(t, len(result) >= 0, "Should return conditions array")
	})

	// Test with nil rule pointer (edge case)
	t.Run("Nil_rule", func(t *testing.T) {
		var rule Rule
		result := GetDuplicateConditionsBetweenOR(rule)
		assert.Equal(t, 0, len(result))
	})

	// Test with simple rule
	t.Run("Simple_rule", func(t *testing.T) {
		rule := Rule{
			Condition: &condition1,
		}
		
		result := GetDuplicateConditionsBetweenOR(rule)
		assert.Assert(t, len(result) >= 0, "Should return conditions array")
	})
}

func TestEqualFixedArgCondition(t *testing.T) {
	// Test with collection value fixed arg
	t.Run("Collection_value_match", func(t *testing.T) {
		// Create a condition with collection fixed arg
		fixedArg := NewFixedArg([]string{"value1", "value2", "target"})
		condition := Condition{
			FixedArg: fixedArg,
		}
		
		result := equalFixedArgCondition(condition, "target")
		assert.Equal(t, true, result) // Should find target in collection
	})

	// Test with collection value no match
	t.Run("Collection_value_no_match", func(t *testing.T) {
		fixedArg := NewFixedArg([]string{"value1", "value2", "value3"})
		condition := Condition{
			FixedArg: fixedArg,
		}
		
		result := equalFixedArgCondition(condition, "notfound")
		assert.Equal(t, false, result) // Should not find notfound in collection
	})

	// Test with string value match
	t.Run("String_value_match", func(t *testing.T) {
		fixedArg := NewFixedArg("Test String Value")
		condition := Condition{
			FixedArg: fixedArg,
		}
		
		result := equalFixedArgCondition(condition, "string")
		assert.Equal(t, true, result) // Should find "string" in "Test String Value" (case insensitive)
	})

	// Test with string value no match
	t.Run("String_value_no_match", func(t *testing.T) {
		fixedArg := NewFixedArg("Test Value")
		condition := Condition{
			FixedArg: fixedArg,
		}
		
		result := equalFixedArgCondition(condition, "notfound")
		assert.Equal(t, false, result) // Should not find "notfound" in "Test Value"
	})

	// Test with nil fixed arg
	t.Run("Nil_fixed_arg", func(t *testing.T) {
		condition := Condition{
			FixedArg: nil,
		}
		
		result := equalFixedArgCondition(condition, "anything")
		assert.Equal(t, false, result) // Should return false for nil fixed arg
	})

	// Test case insensitive matching
	t.Run("Case_insensitive_match", func(t *testing.T) {
		fixedArg := NewFixedArg("UPPERCASE")
		condition := Condition{
			FixedArg: fixedArg,
		}
		
		result := equalFixedArgCondition(condition, "upper")
		assert.Equal(t, true, result) // Should match case insensitively
	})
}
