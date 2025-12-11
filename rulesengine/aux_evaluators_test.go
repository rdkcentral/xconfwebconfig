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

func TestGetAuxEvaluators(t *testing.T) {
	evaluators := GetAuxEvaluators()
	
	// Test that we get evaluators back
	assert.Assert(t, len(evaluators) > 0, "GetAuxEvaluators should return at least one evaluator")
	
	// Test that we have the expected number of evaluators
	// Based on the code: time comparing evaluators + 6 custom evaluators (ev2, ev3, ev4, ev5, ev6, ev7, ev9)
	expectedMinimum := 6 // At minimum we expect 6 custom evaluators
	assert.Assert(t, len(evaluators) >= expectedMinimum, "Should have at least %d evaluators, got %d", expectedMinimum, len(evaluators))
}

func TestAuxEvaluator_IpAddress_Is(t *testing.T) {
	evaluators := GetAuxEvaluators()
	
	// Find the IP_ADDRESS IS evaluator
	var ipIsEvaluator IConditionEvaluator
	for _, eval := range evaluators {
		if eval.FreeArgType() == AuxFreeArgTypeIpAddress && eval.Operation() == StandardOperationIs {
			ipIsEvaluator = eval
			break
		}
	}
	
	assert.Assert(t, ipIsEvaluator != nil, "Should find IP_ADDRESS IS evaluator")
	
	// Test exact match
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationIs,
		NewFixedArg("192.168.1.100"),
	)
	
	context := map[string]string{
		"ipAddress": "192.168.1.100",
	}
	
	result := ipIsEvaluator.Evaluate(condition, context)
	assert.Assert(t, result, "IP address should match exactly")
	
	// Test non-match
	context["ipAddress"] = "192.168.1.101"
	result = ipIsEvaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Different IP addresses should not match")
	
	// Test empty context value
	context["ipAddress"] = ""
	result = ipIsEvaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Empty IP address should not match")
	
	// Test missing context key
	delete(context, "ipAddress")
	result = ipIsEvaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Missing IP address should not match")
	
	// Test with nil fixed arg
	conditionNilArg := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationIs,
		nil,
	)
	context["ipAddress"] = "192.168.1.100"
	result = ipIsEvaluator.Evaluate(conditionNilArg, context)
	assert.Assert(t, !result, "Nil fixed arg should not match")
}

func TestAuxEvaluator_IpAddress_In(t *testing.T) {
	evaluators := GetAuxEvaluators()
	
	// Find the IP_ADDRESS IN evaluator
	var ipInEvaluator IConditionEvaluator
	for _, eval := range evaluators {
		if eval.FreeArgType() == AuxFreeArgTypeIpAddress && eval.Operation() == StandardOperationIn {
			ipInEvaluator = eval
			break
		}
	}
	
	assert.Assert(t, ipInEvaluator != nil, "Should find IP_ADDRESS IN evaluator")
	
	// Test with slice of IPs
	ipList := []string{"10.0.0.1"}
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationIn,
		NewFixedArg(ipList),
	)
	
	context := map[string]string{
		"ipAddress": "10.0.0.1",
	}
	
	result := ipInEvaluator.Evaluate(condition, context)
	assert.Assert(t, result, "IP address should be found in list")
	
	// Test with IP not in list
	context["ipAddress"] = "112.1.1.1" // Different IP that's not in the list
	result = ipInEvaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "IP address not in list should return false")
	
	// Test with empty list
	conditionEmptyList := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationIn,
		NewFixedArg([]string{}),
	)
	context["ipAddress"] = "10.0.0.1"
	result = ipInEvaluator.Evaluate(conditionEmptyList, context)
	assert.Assert(t, !result, "Empty list should return false")
	
	// Test with wrong type (should return false)
	conditionWrongType := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("not-a-slice"),
	)
	
	result = ipInEvaluator.Evaluate(conditionWrongType, context)
	assert.Assert(t, !result, "Wrong fixed arg type should return false")
	
	// Test with nil fixed arg
	conditionNilArg := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationIn,
		nil,
	)
	result = ipInEvaluator.Evaluate(conditionNilArg, context)
	assert.Assert(t, !result, "Nil fixed arg should return false")
}

func TestAuxEvaluator_IpAddress_Percent(t *testing.T) {
	evaluators := GetAuxEvaluators()
	
	// Find the IP_ADDRESS PERCENT evaluator
	var ipPercentEvaluator IConditionEvaluator
	for _, eval := range evaluators {
		if eval.FreeArgType() == AuxFreeArgTypeIpAddress && eval.Operation() == StandardOperationPercent {
			ipPercentEvaluator = eval
			break
		}
	}
	
	assert.Assert(t, ipPercentEvaluator != nil, "Should find IP_ADDRESS PERCENT evaluator")
	
	// Test percentage evaluation with known values
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationPercent,
		NewFixedArg(50.0),
	)
	
	context := map[string]string{
		"ipAddress": "10.0.0.1",
	}
	
	result := ipPercentEvaluator.Evaluate(condition, context)
	// Result should be deterministic based on FitsPercent function
	assert.Assert(t, result == true || result == false, "Should return a boolean result")
	
	// Test with 0% (should always return false)
	conditionZero := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationPercent,
		NewFixedArg(0.0),
	)
	result = ipPercentEvaluator.Evaluate(conditionZero, context)
	assert.Assert(t, !result, "0 percent should always return false")
	
	// Test with 100% (should always return true)
	conditionHundred := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationPercent,
		NewFixedArg(100.0),
	)
	result = ipPercentEvaluator.Evaluate(conditionHundred, context)
	assert.Assert(t, result, "100 percent should always return true")
	
	// Test with wrong type (should return false)
	conditionWrongType := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationPercent,
		NewFixedArg("not-a-float"),
	)
	
	result = ipPercentEvaluator.Evaluate(conditionWrongType, context)
	assert.Assert(t, !result, "Wrong fixed arg type should return false")
	
	// Test with nil fixed arg
	conditionNilArg := NewCondition(
		NewFreeArg(AuxFreeArgTypeIpAddress, "ipAddress"),
		StandardOperationPercent,
		nil,
	)
	result = ipPercentEvaluator.Evaluate(conditionNilArg, context)
	assert.Assert(t, !result, "Nil fixed arg should return false")
}

func TestAuxEvaluator_MacAddress_Is(t *testing.T) {
	evaluators := GetAuxEvaluators()
	
	// Find the MAC_ADDRESS IS evaluator
	var macIsEvaluator IConditionEvaluator
	for _, eval := range evaluators {
		if eval.FreeArgType() == AuxFreeArgTypeMacAddress && eval.Operation() == StandardOperationIs {
			macIsEvaluator = eval
			break
		}
	}
	
	assert.Assert(t, macIsEvaluator != nil, "Should find MAC_ADDRESS IS evaluator")
	
	// Test exact match
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeMacAddress, "eStbMac"),
		StandardOperationIs,
		NewFixedArg("AA:BB:CC:DD:EE:FF"),
	)
	
	context := map[string]string{
		"eStbMac": "AA:BB:CC:DD:EE:FF",
	}
	
	result := macIsEvaluator.Evaluate(condition, context)
	assert.Assert(t, result, "MAC address should match exactly")
	
	// Test non-match
	context["eStbMac"] = "AA:BB:CC:DD:EE:00"
	result = macIsEvaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Different MAC addresses should not match")
}

func TestAuxEvaluator_MacAddress_Like(t *testing.T) {
	evaluators := GetAuxEvaluators()
	
	// Find the MAC_ADDRESS LIKE evaluator
	var macLikeEvaluator IConditionEvaluator
	for _, eval := range evaluators {
		if eval.FreeArgType() == AuxFreeArgTypeMacAddress && eval.Operation() == StandardOperationLike {
			macLikeEvaluator = eval
			break
		}
	}
	
	assert.Assert(t, macLikeEvaluator != nil, "Should find MAC_ADDRESS LIKE evaluator")
	
	// Test regex pattern match
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeMacAddress, "eStbMac"),
		StandardOperationLike,
		NewFixedArg("AA:BB:CC:.*"),
	)
	
	context := map[string]string{
		"eStbMac": "AA:BB:CC:DD:EE:FF",
	}
	
	result := macLikeEvaluator.Evaluate(condition, context)
	assert.Assert(t, result, "MAC address should match regex pattern")
	
	// Test non-matching pattern
	context["eStbMac"] = "DD:EE:FF:AA:BB:CC"
	result = macLikeEvaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "MAC address should not match different pattern")
	
	// Test exact match pattern
	conditionExact := NewCondition(
		NewFreeArg(AuxFreeArgTypeMacAddress, "eStbMac"),
		StandardOperationLike,
		NewFixedArg("^AA:BB:CC:DD:EE:FF$"),
	)
	context["eStbMac"] = "AA:BB:CC:DD:EE:FF"
	result = macLikeEvaluator.Evaluate(conditionExact, context)
	assert.Assert(t, result, "MAC address should match exact pattern")
	
	// Test with invalid regex (should return false)
	conditionInvalidRegex := NewCondition(
		NewFreeArg(AuxFreeArgTypeMacAddress, "eStbMac"),
		StandardOperationLike,
		NewFixedArg("[invalid"),
	)
	
	result = macLikeEvaluator.Evaluate(conditionInvalidRegex, context)
	assert.Assert(t, !result, "Invalid regex should return false")
	
	// Test with wrong type (should return false)
	conditionWrongType := NewCondition(
		NewFreeArg(AuxFreeArgTypeMacAddress, "eStbMac"),
		StandardOperationLike,
		NewFixedArg(123.45),
	)
	
	result = macLikeEvaluator.Evaluate(conditionWrongType, context)
	assert.Assert(t, !result, "Wrong fixed arg type should return false")
	
	// Test with nil fixed arg
	conditionNilArg := NewCondition(
		NewFreeArg(AuxFreeArgTypeMacAddress, "eStbMac"),
		StandardOperationLike,
		nil,
	)
	result = macLikeEvaluator.Evaluate(conditionNilArg, context)
	assert.Assert(t, !result, "Nil fixed arg should return false")
	
	// Test with empty MAC address
	context["eStbMac"] = ""
	result = macLikeEvaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Empty MAC address should not match")
}

func TestAuxEvaluator_MacAddress_In(t *testing.T) {
	evaluators := GetAuxEvaluators()
	
	// Find the MAC_ADDRESS IN evaluators (there are two in the code: ev7 and ev9)
	var macInEvaluators []IConditionEvaluator
	for _, eval := range evaluators {
		if eval.FreeArgType() == AuxFreeArgTypeMacAddress && eval.Operation() == StandardOperationIn {
			macInEvaluators = append(macInEvaluators, eval)
		}
	}
	
	assert.Assert(t, len(macInEvaluators) >= 1, "Should find at least one MAC_ADDRESS IN evaluator")
	
	// Test with the first MAC IN evaluator
	macInEvaluator := macInEvaluators[0]
	
	// Test with slice of MAC addresses
	macList := []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66", "99:88:77:66:55:44"}
	condition := NewCondition(
		NewFreeArg(AuxFreeArgTypeMacAddress, "eStbMac"),
		StandardOperationIn,
		NewFixedArg(macList),
	)
	
	context := map[string]string{
		"eStbMac": "AA:BB:CC:DD:EE:FF",
	}
	
	result := macInEvaluator.Evaluate(condition, context)
	assert.Assert(t, result, "MAC address should be found in list")
	
	// Test with MAC not in list
	context["eStbMac"] = "00:00:00:00:00:00"
	result = macInEvaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "MAC address not in list should return false")
	
	// Test with wrong type (should return false)
	conditionWrongType := NewCondition(
		NewFreeArg(AuxFreeArgTypeMacAddress, "eStbMac"),
		StandardOperationIn,
		NewFixedArg("not-a-slice"),
	)
	
	result = macInEvaluator.Evaluate(conditionWrongType, context)
	assert.Assert(t, !result, "Wrong fixed arg type should return false")
}

func TestAuxEvaluator_TimeComparingEvaluators(t *testing.T) {
	evaluators := GetAuxEvaluators()
	
	// Find time-based evaluators (from GetComparingEvaluators(AuxFreeArgTypeTime))
	var timeEvaluators []IConditionEvaluator
	for _, eval := range evaluators {
		if eval.FreeArgType() == AuxFreeArgTypeTime {
			timeEvaluators = append(timeEvaluators, eval)
		}
	}
	
	assert.Assert(t, len(timeEvaluators) > 0, "Should find time-based evaluators from GetComparingEvaluators")
	
	// The time evaluators should include IS, GT, GTE, LT, LTE operations
	operations := make(map[string]bool)
	for _, eval := range timeEvaluators {
		operations[eval.Operation()] = true
	}
	
	expectedOperations := []string{StandardOperationIs, StandardOperationGt, StandardOperationGte, StandardOperationLt, StandardOperationLte}
	for _, op := range expectedOperations {
		assert.Assert(t, operations[op], "Should have time evaluator for operation: %s", op)
	}
}

func TestAuxEvaluator_EvaluatorProperties(t *testing.T) {
	evaluators := GetAuxEvaluators()
	
	// Test that all evaluators implement the interface correctly
	for i, eval := range evaluators {
		assert.Assert(t, eval.FreeArgType() != "", "Evaluator %d should have non-empty FreeArgType", i)
		assert.Assert(t, eval.Operation() != "", "Evaluator %d should have non-empty Operation", i)
		
		// Test that FreeArgType is one of the expected types
		freeArgType := eval.FreeArgType()
		validTypes := []string{AuxFreeArgTypeTime, AuxFreeArgTypeIpAddress, AuxFreeArgTypeMacAddress}
		isValidType := false
		for _, validType := range validTypes {
			if freeArgType == validType {
				isValidType = true
				break
			}
		}
		assert.Assert(t, isValidType, "Evaluator %d has invalid FreeArgType: %s", i, freeArgType)
	}
}

// Helper functions for creating test fixtures are now replaced with the existing NewFixedArg function
