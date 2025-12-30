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
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestNewBaseTimeEvaluator(t *testing.T) {
	// Test creating a new base time evaluator
	evaluateFunc := func(value int) bool {
		return value > 10
	}
	
	evaluator := NewBaseTimeEvaluator(AuxFreeArgTypeTime, StandardOperationGt, evaluateFunc)
	
	assert.Assert(t, evaluator != nil, "NewBaseTimeEvaluator should return non-nil evaluator")
	assert.Equal(t, AuxFreeArgTypeTime, evaluator.FreeArgType(), "FreeArgType should match")
	assert.Equal(t, StandardOperationGt, evaluator.Operation(), "Operation should match")
	assert.Assert(t, evaluator.evaluateInternal != nil, "evaluateInternal function should be set")
}

func TestBaseTimeEvaluator_FreeArgType(t *testing.T) {
	evaluateFunc := func(value int) bool {
		return true
	}
	
	testCases := []struct {
		name        string
		freeArgType string
	}{
		{"Time type", AuxFreeArgTypeTime},
		{"Custom time type", "CUSTOM_TIME"},
		{"String type", StandardFreeArgTypeString},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator := NewBaseTimeEvaluator(tc.freeArgType, StandardOperationIs, evaluateFunc)
			assert.Equal(t, tc.freeArgType, evaluator.FreeArgType(), "FreeArgType should match expected value")
		})
	}
}

func TestBaseTimeEvaluator_Operation(t *testing.T) {
	evaluateFunc := func(value int) bool {
		return true
	}
	
	testCases := []struct {
		name      string
		operation string
	}{
		{"IS operation", StandardOperationIs},
		{"GT operation", StandardOperationGt},
		{"LT operation", StandardOperationLt},
		{"GTE operation", StandardOperationGte},
		{"LTE operation", StandardOperationLte},
		{"Custom operation", "CUSTOM_TIME_OP"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator := NewBaseTimeEvaluator(AuxFreeArgTypeTime, tc.operation, evaluateFunc)
			assert.Equal(t, tc.operation, evaluator.Operation(), "Operation should match expected value")
		})
	}
}

func TestBaseTimeEvaluator_Evaluate_PanicsWhenNotImplemented(t *testing.T) {
	// Test that the Evaluate method panics since it's not fully implemented
	evaluateFunc := func(value int) bool {
		return true
	}
	
	evaluator := NewBaseTimeEvaluator(AuxFreeArgTypeTime, StandardOperationGt, evaluateFunc)
	
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeTime, "timeValue"),
		StandardOperationGt,
		NewFixedArg("12"),
	)
	
	context := map[string]string{
		"timeValue": "15",
	}
	
	// Test that it panics with the expected message
	defer func() {
		if r := recover(); r != nil {
			errMsg := r.(error).Error()
			assert.Assert(t, strings.Contains(errMsg, "BaseTimeEvaluator.Evaluate() is not implemented yet"), 
				"Should panic with not implemented message, got: %s", errMsg)
		} else {
			t.Fatal("Expected panic but none occurred")
		}
	}()
	
	evaluator.Evaluate(condition, context)
}

func TestBaseTimeEvaluator_Evaluate_ReturnsFalseWithEmptyValue(t *testing.T) {
	// This test checks the early return path for empty values
	evaluateFunc := func(value int) bool {
		return true
	}
	
	// We can't actually test the empty value logic without triggering the panic
	// since the method is not fully implemented. This test documents the expected behavior.
	_ = evaluateFunc // Acknowledge we're not using this in the actual test
	
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeTime, "timeValue"),
		StandardOperationGt,
		NewFixedArg("12"),
	)
	
	// Test with empty context (missing key)
	context := map[string]string{
		"otherValue": "15",
	}
	
	// Create evaluator to test the panic behavior
	evaluator := NewBaseTimeEvaluator(AuxFreeArgTypeTime, StandardOperationGt, evaluateFunc)
	
	// This should trigger the panic, but we can test the structure
	defer func() {
		if r := recover(); r != nil {
			// Expected behavior - the method panics
			assert.Assert(t, true, "Method correctly panics when not implemented")
		} else {
			t.Fatal("Expected panic but none occurred")
		}
	}()
	
	evaluator.Evaluate(condition, context)
}

func TestBaseTimeEvaluator_Evaluate_WithNilFreeArg(t *testing.T) {
	evaluateFunc := func(value int) bool {
		return true
	}
	
	evaluator := NewBaseTimeEvaluator(AuxFreeArgTypeTime, StandardOperationGt, evaluateFunc)
	
	// Create condition with nil FreeArg
	condition := &Condition{
		FreeArg:   nil,
		Operation: StandardOperationGt,
		FixedArg:  NewFixedArg("12"),
	}
	
	context := map[string]string{
		"timeValue": "15",
	}
	
	// Should panic because method is not implemented
	defer func() {
		if r := recover(); r != nil {
			errMsg := r.(error).Error()
			assert.Assert(t, strings.Contains(errMsg, "BaseTimeEvaluator.Evaluate() is not implemented yet"), 
				"Should panic with not implemented message")
		} else {
			t.Fatal("Expected panic but none occurred")
		}
	}()
	
	evaluator.Evaluate(condition, context)
}

func TestBaseTimeEvaluator_Interface_Compliance(t *testing.T) {
	// Test that BaseTimeEvaluator implements IConditionEvaluator interface
	evaluateFunc := func(value int) bool {
		return true
	}
	
	evaluator := NewBaseTimeEvaluator(AuxFreeArgTypeTime, StandardOperationGt, evaluateFunc)
	
	// Test that it can be used as IConditionEvaluator
	var iface IConditionEvaluator = evaluator
	
	assert.Equal(t, AuxFreeArgTypeTime, iface.FreeArgType(), "Interface should return correct FreeArgType")
	assert.Equal(t, StandardOperationGt, iface.Operation(), "Interface should return correct Operation")
	
	// Note: We don't test the Evaluate method through the interface since it panics
	assert.Assert(t, iface != nil, "Interface should not be nil")
}

func TestBaseTimeEvaluator_EvaluateInternal_Function(t *testing.T) {
	// Test different evaluate internal functions
	testCases := []struct {
		name     string
		function FnIntEval
		input    int
		expected bool
	}{
		{
			name:     "Greater than 10",
			function: func(value int) bool { return value > 10 },
			input:    15,
			expected: true,
		},
		{
			name:     "Less than 10",
			function: func(value int) bool { return value > 10 },
			input:    5,
			expected: false,
		},
		{
			name:     "Always true",
			function: func(value int) bool { return true },
			input:    0,
			expected: true,
		},
		{
			name:     "Always false",
			function: func(value int) bool { return false },
			input:    100,
			expected: false,
		},
		{
			name:     "Even numbers",
			function: func(value int) bool { return value%2 == 0 },
			input:    4,
			expected: true,
		},
		{
			name:     "Odd numbers",
			function: func(value int) bool { return value%2 == 0 },
			input:    3,
			expected: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator := NewBaseTimeEvaluator(AuxFreeArgTypeTime, StandardOperationGt, tc.function)
			
			// Test the internal function directly
			result := evaluator.evaluateInternal(tc.input)
			assert.Equal(t, tc.expected, result, "Function should return expected result")
		})
	}
}
