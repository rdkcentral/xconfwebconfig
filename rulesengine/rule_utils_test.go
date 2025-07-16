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
