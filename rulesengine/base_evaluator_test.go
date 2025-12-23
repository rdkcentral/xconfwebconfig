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

func TestNewBaseEvaluator(t *testing.T) {
	// Test creating a new base evaluator
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		return freeArgValue == fixedArgValue
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeString, StandardOperationIs, evaluateFunc)
	
	assert.Assert(t, evaluator != nil, "NewBaseEvaluator should return non-nil evaluator")
	assert.Equal(t, StandardFreeArgTypeString, evaluator.FreeArgType(), "FreeArgType should match")
	assert.Equal(t, StandardOperationIs, evaluator.Operation(), "Operation should match")
	assert.Assert(t, evaluator.evaluateInternal != nil, "evaluateInternal function should be set")
}

func TestBaseEvaluator_FreeArgType(t *testing.T) {
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		return true
	}
	
	testCases := []struct {
		name        string
		freeArgType string
	}{
		{"String type", StandardFreeArgTypeString},
		{"Long type", StandardFreeArgTypeLong},
		{"Void type", StandardFreeArgTypeVoid},
		{"Any type", StandardFreeArgTypeAny},
		{"Custom type", "CUSTOM_TYPE"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator := NewBaseEvaluator(tc.freeArgType, StandardOperationIs, evaluateFunc)
			assert.Equal(t, tc.freeArgType, evaluator.FreeArgType(), "FreeArgType should match expected value")
		})
	}
}

func TestBaseEvaluator_Operation(t *testing.T) {
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		return true
	}
	
	testCases := []struct {
		name      string
		operation string
	}{
		{"IS operation", StandardOperationIs},
		{"IN operation", StandardOperationIn},
		{"EXISTS operation", StandardOperationExists},
		{"GT operation", StandardOperationGt},
		{"Custom operation", "CUSTOM_OP"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator := NewBaseEvaluator(StandardFreeArgTypeString, tc.operation, evaluateFunc)
			assert.Equal(t, tc.operation, evaluator.Operation(), "Operation should match expected value")
		})
	}
}

func TestBaseEvaluator_Validate(t *testing.T) {
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		return true
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeString, StandardOperationIs, evaluateFunc)
	
	// Test with nil FixedArg
	err := evaluator.Validate(nil)
	assert.NilError(t, err, "Validate should return no error for nil FixedArg")
	
	// Test with valid FixedArg
	fixedArg := NewFixedArg("test")
	err = evaluator.Validate(fixedArg)
	assert.NilError(t, err, "Validate should return no error for valid FixedArg")
}

func TestBaseEvaluator_Evaluate_StandardFreeArgTypeVoid(t *testing.T) {
	// Test evaluator with VOID free arg type
	callCount := 0
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		callCount++
		assert.Equal(t, "", freeArgValue, "freeArgValue should be empty for VOID type")
		return fixedArgValue == "test"
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeVoid, StandardOperationIs, evaluateFunc)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeVoid, "voidArg"),
		StandardOperationIs,
		NewFixedArg("test"),
	)
	
	context := map[string]string{
		"voidArg": "should_be_ignored",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, result, "Should return true when fixed arg matches expected value")
	assert.Equal(t, 1, callCount, "evaluateInternal should be called once")
}

func TestBaseEvaluator_Evaluate_ExistsOperation(t *testing.T) {
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		// This should not be called for EXISTS operation
		t.Fatal("evaluateInternal should not be called for EXISTS operation")
		return false
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeString, StandardOperationExists, evaluateFunc)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "testArg"),
		StandardOperationExists,
		NewFixedArg("test"),
	)
	
	// Test when key exists in context
	context := map[string]string{
		"testArg": "value",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, result, "Should return true when key exists in context")
	
	// Test when key does not exist in context
	contextMissing := map[string]string{
		"otherArg": "value",
	}
	
	result = evaluator.Evaluate(condition, contextMissing)
	assert.Assert(t, !result, "Should return false when key does not exist in context")
}

func TestBaseEvaluator_Evaluate_MissingContextKey(t *testing.T) {
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		// This should not be called when context key is missing
		t.Fatal("evaluateInternal should not be called when context key is missing")
		return false
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeString, StandardOperationIs, evaluateFunc)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "missingArg"),
		StandardOperationIs,
		NewFixedArg("test"),
	)
	
	context := map[string]string{
		"otherArg": "value",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when context key is missing")
}

func TestBaseEvaluator_Evaluate_EmptyContextValue(t *testing.T) {
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		// This should not be called when context value is empty (for non-ANY types)
		t.Fatal("evaluateInternal should not be called when context value is empty")
		return false
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeString, StandardOperationIs, evaluateFunc)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "emptyArg"),
		StandardOperationIs,
		NewFixedArg("test"),
	)
	
	context := map[string]string{
		"emptyArg": "",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when context value is empty for non-ANY type")
}

func TestBaseEvaluator_Evaluate_EmptyContextValue_AnyType(t *testing.T) {
	callCount := 0
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		callCount++
		assert.Equal(t, "", freeArgValue, "freeArgValue should be empty")
		return true
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeAny, StandardOperationIs, evaluateFunc)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeAny, "emptyArg"),
		StandardOperationIs,
		NewFixedArg("test"),
	)
	
	context := map[string]string{
		"emptyArg": "",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, result, "Should allow empty values for ANY type")
	assert.Equal(t, 1, callCount, "evaluateInternal should be called")
}

func TestBaseEvaluator_Evaluate_WithFixedArg(t *testing.T) {
	callCount := 0
	var lastFreeArgValue string
	var lastFixedArgValue interface{}
	
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		callCount++
		lastFreeArgValue = freeArgValue
		lastFixedArgValue = fixedArgValue
		return freeArgValue == fixedArgValue
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeString, StandardOperationIs, evaluateFunc)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "testArg"),
		StandardOperationIs,
		NewFixedArg("expectedValue"),
	)
	
	context := map[string]string{
		"testArg": "testValue",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when values don't match")
	assert.Equal(t, 1, callCount, "evaluateInternal should be called once")
	assert.Equal(t, "testValue", lastFreeArgValue, "freeArgValue should match context value")
	assert.Equal(t, "expectedValue", lastFixedArgValue, "fixedArgValue should match fixed arg value")
	
	// Test with matching values
	context["testArg"] = "expectedValue"
	result = evaluator.Evaluate(condition, context)
	assert.Assert(t, result, "Should return true when values match")
	assert.Equal(t, 2, callCount, "evaluateInternal should be called twice")
	assert.Equal(t, "expectedValue", lastFreeArgValue, "freeArgValue should match updated context value")
	assert.Equal(t, "expectedValue", lastFixedArgValue, "fixedArgValue should still match fixed arg value")
}

func TestBaseEvaluator_Evaluate_WithNilFixedArg(t *testing.T) {
	callCount := 0
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		callCount++
		assert.Equal(t, "testValue", freeArgValue, "freeArgValue should match context value")
		assert.Assert(t, fixedArgValue == nil, "fixedArgValue should be nil")
		return true
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeString, StandardOperationIs, evaluateFunc)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "testArg"),
		StandardOperationIs,
		nil,
	)
	
	context := map[string]string{
		"testArg": "testValue",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, result, "Should return true when evaluateInternal returns true")
	assert.Equal(t, 1, callCount, "evaluateInternal should be called once")
}

func TestBaseEvaluator_Evaluate_ComplexScenarios(t *testing.T) {
	// Test with different data types in fixed arg
	testCases := []struct {
		name          string
		fixedArgValue interface{}
		contextValue  string
		expected      bool
	}{
		{"String comparison", "test", "test", true},
		{"String mismatch", "test", "other", false},
		{"Float comparison", 42.5, "42.5", true},
		{"Slice comparison", []string{"a", "b", "c"}, "b", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
				switch fv := fixedArgValue.(type) {
				case string:
					return freeArgValue == fv
				case float64:
					return freeArgValue == "42.5" && fv == 42.5
				case []string:
					for _, item := range fv {
						if item == freeArgValue {
							return true
						}
					}
					return false
				}
				return false
			}
			
			evaluator := NewBaseEvaluator(StandardFreeArgTypeString, StandardOperationIs, evaluateFunc)
			
			condition := NewCondition(
				NewFreeArg(StandardFreeArgTypeString, "testArg"),
				StandardOperationIs,
				NewFixedArg(tc.fixedArgValue),
			)
			
			context := map[string]string{
				"testArg": tc.contextValue,
			}
			
			result := evaluator.Evaluate(condition, context)
			assert.Equal(t, tc.expected, result, "Result should match expected value for %s", tc.name)
		})
	}
}

func TestBaseEvaluator_Interface_Compliance(t *testing.T) {
	// Test that BaseEvaluator implements IConditionEvaluator interface
	evaluateFunc := func(freeArgValue string, fixedArgValue interface{}) bool {
		return true
	}
	
	evaluator := NewBaseEvaluator(StandardFreeArgTypeString, StandardOperationIs, evaluateFunc)
	
	// Test that it can be used as IConditionEvaluator
	var iface IConditionEvaluator = evaluator
	
	assert.Equal(t, StandardFreeArgTypeString, iface.FreeArgType(), "Interface should return correct FreeArgType")
	assert.Equal(t, StandardOperationIs, iface.Operation(), "Interface should return correct Operation")
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "testArg"),
		StandardOperationIs,
		NewFixedArg("test"),
	)
	
	context := map[string]string{
		"testArg": "test",
	}
	
	result := iface.Evaluate(condition, context)
	assert.Assert(t, result, "Interface Evaluate should work correctly")
}
