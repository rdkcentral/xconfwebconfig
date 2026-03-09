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

func TestGetComparingEvaluators_StringType(t *testing.T) {
	evaluators := GetComparingEvaluators(StandardFreeArgTypeString)
	
	// STRING type should have 4 evaluators (GT, GTE, LT, LTE but not IS)
	assert.Equal(t, 4, len(evaluators), "STRING type should have 4 comparing evaluators")
	
	// Verify the operations
	operations := make(map[string]bool)
	for _, evaluator := range evaluators {
		operations[evaluator.Operation()] = true
		assert.Equal(t, StandardFreeArgTypeString, evaluator.FreeArgType(), "All evaluators should have STRING type")
	}
	
	assert.Assert(t, operations[StandardOperationGt], "Should include GT operation")
	assert.Assert(t, operations[StandardOperationGte], "Should include GTE operation")
	assert.Assert(t, operations[StandardOperationLt], "Should include LT operation")
	assert.Assert(t, operations[StandardOperationLte], "Should include LTE operation")
	assert.Assert(t, !operations[StandardOperationIs], "Should NOT include IS operation for STRING type")
}

func TestGetComparingEvaluators_LongType(t *testing.T) {
	evaluators := GetComparingEvaluators(StandardFreeArgTypeLong)
	
	// LONG type should have 5 evaluators (IS, GT, GTE, LT, LTE)
	assert.Equal(t, 5, len(evaluators), "LONG type should have 5 comparing evaluators")
	
	// Verify the operations
	operations := make(map[string]bool)
	for _, evaluator := range evaluators {
		operations[evaluator.Operation()] = true
		assert.Equal(t, StandardFreeArgTypeLong, evaluator.FreeArgType(), "All evaluators should have LONG type")
	}
	
	assert.Assert(t, operations[StandardOperationIs], "Should include IS operation")
	assert.Assert(t, operations[StandardOperationGt], "Should include GT operation")
	assert.Assert(t, operations[StandardOperationGte], "Should include GTE operation")
	assert.Assert(t, operations[StandardOperationLt], "Should include LT operation")
	assert.Assert(t, operations[StandardOperationLte], "Should include LTE operation")
}

func TestGetComparingEvaluators_VoidType(t *testing.T) {
	evaluators := GetComparingEvaluators(StandardFreeArgTypeVoid)
	
	// VOID type should have 5 evaluators (all operations)
	assert.Equal(t, 5, len(evaluators), "VOID type should have 5 comparing evaluators")
	
	for _, evaluator := range evaluators {
		assert.Equal(t, StandardFreeArgTypeVoid, evaluator.FreeArgType(), "All evaluators should have VOID type")
	}
}

func TestNewComparingEvaluator(t *testing.T) {
	testCases := []struct {
		name        string
		freeArgType string
		operation   string
	}{
		{"String GT", StandardFreeArgTypeString, StandardOperationGt},
		{"Long IS", StandardFreeArgTypeLong, StandardOperationIs},
		{"Void LTE", StandardFreeArgTypeVoid, StandardOperationLte},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := func(i int) bool { return i > 0 }
			evaluator := NewComparingEvaluator(tc.freeArgType, tc.operation, fn)
			
			assert.Assert(t, evaluator != nil, "NewComparingEvaluator should return non-nil evaluator")
			assert.Equal(t, tc.freeArgType, evaluator.FreeArgType(), "FreeArgType should match")
			assert.Equal(t, tc.operation, evaluator.Operation(), "Operation should match")
			assert.Assert(t, evaluator.evaluation != nil, "Evaluation function should be set")
		})
	}
}

func TestComparingEvaluator_FreeArgType(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(StandardFreeArgTypeString, StandardOperationIs, fn)
	
	assert.Equal(t, StandardFreeArgTypeString, evaluator.FreeArgType(), "FreeArgType should return expected value")
}

func TestComparingEvaluator_Operation(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(StandardFreeArgTypeLong, StandardOperationGt, fn)
	
	assert.Equal(t, StandardOperationGt, evaluator.Operation(), "Operation should return expected value")
}

func TestComparingEvaluator_Evaluate_MissingContextKey(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(StandardFreeArgTypeLong, StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeLong, "testValue"),
		StandardOperationIs,
		NewFixedArg(float64(10)),
	)
	
	// Context missing the key
	context := map[string]string{
		"otherKey": "value",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when context key is missing")
}

func TestComparingEvaluator_Evaluate_EmptyContextValue(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(StandardFreeArgTypeLong, StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeLong, "testValue"),
		StandardOperationIs,
		NewFixedArg(float64(10)),
	)
	
	// Context with empty value
	context := map[string]string{
		"testValue": "",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when context value is empty")
}

func TestComparingEvaluator_Evaluate_VoidType(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(StandardFreeArgTypeVoid, StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeVoid, "testValue"),
		StandardOperationIs,
		NewFixedArg(float64(10)),
	)
	
	// VOID type should ignore context
	context := map[string]string{
		"otherKey": "value",
	}
	
	result := evaluator.Evaluate(condition, context)
	// VOID type will have empty freeArgValue, and should try to parse as number (fail)
	assert.Assert(t, !result, "VOID type with empty freeArgValue should fail to parse as number")
}

func TestComparingEvaluator_Evaluate_LongType_Success(t *testing.T) {
	testCases := []struct {
		name          string
		operation     string
		freeArgValue  string
		fixedArgValue float64
		evalFunc      func(int) bool
		expected      bool
	}{
		{"Equal values IS", StandardOperationIs, "10", 10.0, func(i int) bool { return i == 0 }, true},
		{"Equal values IS false", StandardOperationIs, "10", 11.0, func(i int) bool { return i == 0 }, false},
		{"Greater than GT", StandardOperationGt, "15", 10.0, func(i int) bool { return i > 0 }, true},
		{"Greater than GT false", StandardOperationGt, "5", 10.0, func(i int) bool { return i > 0 }, false},
		{"Greater than or equal GTE", StandardOperationGte, "10", 10.0, func(i int) bool { return i >= 0 }, true},
		{"Greater than or equal GTE true", StandardOperationGte, "15", 10.0, func(i int) bool { return i >= 0 }, true},
		{"Less than LT", StandardOperationLt, "5", 10.0, func(i int) bool { return i < 0 }, true},
		{"Less than LT false", StandardOperationLt, "15", 10.0, func(i int) bool { return i < 0 }, false},
		{"Less than or equal LTE", StandardOperationLte, "10", 10.0, func(i int) bool { return i <= 0 }, true},
		{"Less than or equal LTE true", StandardOperationLte, "5", 10.0, func(i int) bool { return i <= 0 }, true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator := NewComparingEvaluator(StandardFreeArgTypeLong, tc.operation, tc.evalFunc)
			
			condition := NewCondition(
				NewFreeArg(StandardFreeArgTypeLong, "testValue"),
				tc.operation,
				NewFixedArg(tc.fixedArgValue),
			)
			
			context := map[string]string{
				"testValue": tc.freeArgValue,
			}
			
			result := evaluator.Evaluate(condition, context)
			assert.Equal(t, tc.expected, result, "Evaluation result should match expected")
		})
	}
}

func TestComparingEvaluator_Evaluate_LongType_InvalidFreeArg(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(StandardFreeArgTypeLong, StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeLong, "testValue"),
		StandardOperationIs,
		NewFixedArg(float64(10)),
	)
	
	context := map[string]string{
		"testValue": "not-a-number",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when free arg value is not a valid number")
}

func TestComparingEvaluator_Evaluate_LongType_InvalidFixedArg(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(StandardFreeArgTypeLong, StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeLong, "testValue"),
		StandardOperationIs,
		NewFixedArg("not-a-float64"),
	)
	
	context := map[string]string{
		"testValue": "10",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when fixed arg value is not a float64")
}

func TestComparingEvaluator_Evaluate_TimeType_Success(t *testing.T) {
	testCases := []struct {
		name          string
		operation     string
		freeArgValue  string
		fixedArgValue string
		evalFunc      func(int) bool
		expected      bool
	}{
		{"Equal times IS", StandardOperationIs, "10:30:00", "10:30:00", func(i int) bool { return i == 0 }, true},
		{"Equal times IS false", StandardOperationIs, "10:30:00", "11:30:00", func(i int) bool { return i == 0 }, false},
		{"Later time GT", StandardOperationGt, "15:30:00", "10:30:00", func(i int) bool { return i > 0 }, true},
		{"Later time GT false", StandardOperationGt, "05:30:00", "10:30:00", func(i int) bool { return i > 0 }, false},
		{"Later or equal time GTE", StandardOperationGte, "10:30:00", "10:30:00", func(i int) bool { return i >= 0 }, true},
		{"Later or equal time GTE true", StandardOperationGte, "15:30:00", "10:30:00", func(i int) bool { return i >= 0 }, true},
		{"Earlier time LT", StandardOperationLt, "05:30:00", "10:30:00", func(i int) bool { return i < 0 }, true},
		{"Earlier time LT false", StandardOperationLt, "15:30:00", "10:30:00", func(i int) bool { return i < 0 }, false},
		{"Earlier or equal time LTE", StandardOperationLte, "10:30:00", "10:30:00", func(i int) bool { return i <= 0 }, true},
		{"Earlier or equal time LTE true", StandardOperationLte, "05:30:00", "10:30:00", func(i int) bool { return i <= 0 }, true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator := NewComparingEvaluator(AuxFreeArgTypeTime, tc.operation, tc.evalFunc)
			
			condition := NewCondition(
				NewFreeArg(AuxFreeArgTypeTime, "timeValue"),
				tc.operation,
				NewFixedArg(tc.fixedArgValue),
			)
			
			context := map[string]string{
				"timeValue": tc.freeArgValue,
			}
			
			result := evaluator.Evaluate(condition, context)
			assert.Equal(t, tc.expected, result, "Time evaluation result should match expected")
		})
	}
}

func TestComparingEvaluator_Evaluate_TimeType_InvalidFreeArg(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(AuxFreeArgTypeTime, StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeTime, "timeValue"),
		StandardOperationIs,
		NewFixedArg("10:30:00"),
	)
	
	context := map[string]string{
		"timeValue": "invalid-time",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when free arg time is invalid")
}

func TestComparingEvaluator_Evaluate_TimeType_InvalidFixedArg(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(AuxFreeArgTypeTime, StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeTime, "timeValue"),
		StandardOperationIs,
		NewFixedArg(123), // Not a string
	)
	
	context := map[string]string{
		"timeValue": "10:30:00",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when fixed arg is not a string")
}

func TestComparingEvaluator_Evaluate_TimeType_InvalidFixedArgTime(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(AuxFreeArgTypeTime, StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeTime, "timeValue"),
		StandardOperationIs,
		NewFixedArg("invalid-time"),
	)
	
	context := map[string]string{
		"timeValue": "10:30:00",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when fixed arg time is invalid")
}

func TestComparingEvaluator_Evaluate_UnsupportedType(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator("UNSUPPORTED_TYPE", StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg("UNSUPPORTED_TYPE", "testValue"),
		StandardOperationIs,
		NewFixedArg("value"),
	)
	
	context := map[string]string{
		"testValue": "test",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false for unsupported type")
}

func TestComparingEvaluator_Evaluate_AnyType_EmptyValue(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(StandardFreeArgTypeAny, StandardOperationIs, fn)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeAny, "testValue"),
		StandardOperationIs,
		NewFixedArg("value"),
	)
	
	// ANY type allows empty values
	context := map[string]string{
		"testValue": "",
	}
	
	result := evaluator.Evaluate(condition, context)
	// Should proceed to evaluation but fail since empty string is not valid for LONG or TIME
	assert.Assert(t, !result, "ANY type with empty value should fail for unsupported evaluation")
}

func TestComparingEvaluator_Interface_Compliance(t *testing.T) {
	fn := func(i int) bool { return i == 0 }
	evaluator := NewComparingEvaluator(StandardFreeArgTypeLong, StandardOperationIs, fn)
	
	// Test that it implements IConditionEvaluator interface
	var iface IConditionEvaluator = evaluator
	
	assert.Equal(t, StandardFreeArgTypeLong, iface.FreeArgType(), "Interface should return correct FreeArgType")
	assert.Equal(t, StandardOperationIs, iface.Operation(), "Interface should return correct Operation")
	assert.Assert(t, iface != nil, "Interface should not be nil")
}

func TestComparableOperationFnMap(t *testing.T) {
	testCases := []struct {
		operation string
		value     int
		expected  bool
	}{
		{StandardOperationIs, 0, true},
		{StandardOperationIs, 1, false},
		{StandardOperationIs, -1, false},
		{StandardOperationGt, 1, true},
		{StandardOperationGt, 0, false},
		{StandardOperationGt, -1, false},
		{StandardOperationGte, 1, true},
		{StandardOperationGte, 0, true},
		{StandardOperationGte, -1, false},
		{StandardOperationLt, -1, true},
		{StandardOperationLt, 0, false},
		{StandardOperationLt, 1, false},
		{StandardOperationLte, -1, true},
		{StandardOperationLte, 0, true},
		{StandardOperationLte, 1, false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.operation+"_"+string(rune(tc.value+'0')), func(t *testing.T) {
			fn, exists := comparableOperationFnMap[tc.operation]
			assert.Assert(t, exists, "Operation should exist in map")
			
			result := fn(tc.value)
			assert.Equal(t, tc.expected, result, "Function result should match expected")
		})
	}
}
