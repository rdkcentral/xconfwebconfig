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

func TestConditionEquals(t *testing.T) {
	day := NewFreeArg(StandardFreeArgTypeString, "day")
	c1 := NewCondition(day, StandardOperationIs, NewFixedArg("Friday"))
	c2 := NewCondition(day, StandardOperationIs, NewFixedArg("Friday"))
	assert.Assert(t, c1.Equals(c2))
	c3 := NewCondition(day, StandardOperationIs, NewFixedArg("Saturday"))
	assert.Assert(t, !c1.Equals(c3))
}

func TestCondition_SetFreeArg(t *testing.T) {
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "original"),
		StandardOperationIs,
		NewFixedArg("value"),
	)
	
	newFreeArg := NewFreeArg(StandardFreeArgTypeLong, "newArg")
	condition.SetFreeArg(newFreeArg)
	
	assert.Equal(t, newFreeArg, condition.GetFreeArg(), "SetFreeArg should update the free arg")
	assert.Equal(t, "newArg", condition.GetFreeArg().GetName(), "Free arg name should be updated")
	assert.Equal(t, StandardFreeArgTypeLong, condition.GetFreeArg().GetType(), "Free arg type should be updated")
}

func TestCondition_SetFixedArg(t *testing.T) {
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "test"),
		StandardOperationIs,
		NewFixedArg("original"),
	)
	
	newFixedArg := NewFixedArg("newValue")
	condition.SetFixedArg(newFixedArg)
	
	assert.Equal(t, newFixedArg, condition.GetFixedArg(), "SetFixedArg should update the fixed arg")
	assert.Equal(t, "newValue", condition.GetFixedArg().GetValue(), "Fixed arg value should be updated")
}

func TestCondition_SetOperation(t *testing.T) {
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "test"),
		StandardOperationIs,
		NewFixedArg("value"),
	)
	
	condition.SetOperation(StandardOperationIn)
	
	assert.Equal(t, StandardOperationIn, condition.GetOperation(), "SetOperation should update the operation")
}

func TestCondition_String(t *testing.T) {
	testCases := []struct {
		name      string
		condition *Condition
		expected  string
	}{
		{
			name: "Normal condition",
			condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "model"),
				StandardOperationIs,
				NewFixedArg("TG1682G"),
			),
			expected: "Condition('model' IS 'TG1682G')",
		},
		{
			name: "Condition with IN operation",
			condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "environment"),
				StandardOperationIn,
				NewFixedArg([]string{"QA", "PROD"}),
			),
			expected: "Condition('environment' IN '[QA PROD]')",
		},
		{
			name: "Condition with LONG type",
			condition: NewCondition(
				NewFreeArg(StandardFreeArgTypeLong, "version"),
				StandardOperationGt,
				NewFixedArg(float64(123)),
			),
			expected: "Condition('version(LONG)' GT '123')",
		},
		{
			name:      "Nil condition",
			condition: nil,
			expected:  "Condition(nil)",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.condition.String()
			assert.Equal(t, tc.expected, result, "String representation should match expected")
		})
	}
}

func TestNewConditionInfo(t *testing.T) {
	freeArg := *NewFreeArg(StandardFreeArgTypeString, "testArg")
	operation := StandardOperationIs
	
	conditionInfo := NewConditionInfo(freeArg, operation)
	
	assert.Assert(t, conditionInfo != nil, "NewConditionInfo should return non-nil")
	assert.Equal(t, freeArg, conditionInfo.FreeArg, "ConditionInfo should have correct FreeArg")
	assert.Equal(t, operation, conditionInfo.Operation, "ConditionInfo should have correct Operation")
}

func TestCondition_SetterGetterChaining(t *testing.T) {
	// Test that setters can be chained and getters return the correct values
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "initial"),
		StandardOperationIs,
		NewFixedArg("initial"),
	)
	
	// Set new values
	newFreeArg := NewFreeArg(StandardFreeArgTypeLong, "modified")
	newFixedArg := NewFixedArg(float64(123))
	newOperation := StandardOperationGt
	
	condition.SetFreeArg(newFreeArg)
	condition.SetFixedArg(newFixedArg)
	condition.SetOperation(newOperation)
	
	// Verify all values were updated
	assert.Equal(t, newFreeArg, condition.GetFreeArg(), "FreeArg should be updated")
	assert.Equal(t, newFixedArg, condition.GetFixedArg(), "FixedArg should be updated")
	assert.Equal(t, newOperation, condition.GetOperation(), "Operation should be updated")
}

func TestCondition_SetNilValues(t *testing.T) {
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "test"),
		StandardOperationIs,
		NewFixedArg("value"),
	)
	
	// Set nil values
	condition.SetFreeArg(nil)
	condition.SetFixedArg(nil)
	condition.SetOperation("")
	
	// Verify values were set
	assert.Assert(t, condition.GetFreeArg() == nil, "FreeArg should be nil")
	assert.Assert(t, condition.GetFixedArg() == nil, "FixedArg should be nil")
	assert.Equal(t, "", condition.GetOperation(), "Operation should be empty")
}
