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
	"errors"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"

	"gotest.tools/assert"
)

// MockCachedSimpleDao implements db.CachedSimpleDao for testing
type MockCachedSimpleDao struct {
	data map[string]interface{}
	err  error
}

func (m *MockCachedSimpleDao) GetOne(tableName string, rowKey string) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data[rowKey], nil
}

func (m *MockCachedSimpleDao) GetOneFromCacheOnly(tableName string, rowKey string) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data[rowKey], nil
}

func (m *MockCachedSimpleDao) SetOne(tableName string, rowKey string, entity interface{}) error {
	if m.err != nil {
		return m.err
	}
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[rowKey] = entity
	return nil
}

func (m *MockCachedSimpleDao) DeleteOne(tableName string, rowKey string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.data, rowKey)
	return nil
}

func (m *MockCachedSimpleDao) GetAllByKeys(tableName string, rowKeys []string) ([]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []interface{}
	for _, key := range rowKeys {
		if value, exists := m.data[key]; exists {
			result = append(result, value)
		}
	}
	return result, nil
}

func (m *MockCachedSimpleDao) GetAllAsList(tableName string, maxResults int) ([]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []interface{}
	count := 0
	for _, value := range m.data {
		if maxResults > 0 && count >= maxResults {
			break
		}
		result = append(result, value)
		count++
	}
	return result, nil
}

func (m *MockCachedSimpleDao) GetAllAsMap(tableName string) (map[interface{}]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[interface{}]interface{})
	for key, value := range m.data {
		result[key] = value
	}
	return result, nil
}

func (m *MockCachedSimpleDao) GetKeys(tableName string) ([]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []interface{}
	for key := range m.data {
		result = append(result, key)
	}
	return result, nil
}

func (m *MockCachedSimpleDao) RefreshAll(tableName string) error {
	return m.err
}

func (m *MockCachedSimpleDao) RefreshOne(tableName string, rowKey string) error {
	return m.err
}

func TestNewIpAddressEvaluator(t *testing.T) {
	mockDao := &MockCachedSimpleDao{}
	
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	assert.Assert(t, evaluator != nil, "NewIpAddressEvaluator should return non-nil evaluator")
	assert.Equal(t, StandardFreeArgTypeString, evaluator.FreeArgType(), "FreeArgType should match")
	assert.Equal(t, StandardOperationIn, evaluator.Operation(), "Operation should match")
	assert.Assert(t, evaluator.nsListDao != nil, "nsListDao should be set")
}

func TestIpAddressEvaluator_FreeArgType(t *testing.T) {
	mockDao := &MockCachedSimpleDao{}
	
	testCases := []struct {
		name        string
		freeArgType string
	}{
		{"String type", StandardFreeArgTypeString},
		{"Any type", StandardFreeArgTypeAny},
		{"Custom type", "CUSTOM_IP_TYPE"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator := NewIpAddressEvaluator(tc.freeArgType, StandardOperationIn, mockDao)
			assert.Equal(t, tc.freeArgType, evaluator.FreeArgType(), "FreeArgType should match expected value")
		})
	}
}

func TestIpAddressEvaluator_Operation(t *testing.T) {
	mockDao := &MockCachedSimpleDao{}
	
	testCases := []struct {
		name      string
		operation string
	}{
		{"IN operation", StandardOperationIn},
		{"IS operation", StandardOperationIs},
		{"Custom operation", "CUSTOM_IP_OP"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, tc.operation, mockDao)
			assert.Equal(t, tc.operation, evaluator.Operation(), "Operation should match expected value")
		})
	}
}

func TestIpAddressEvaluator_Evaluate_MissingContextKey(t *testing.T) {
	mockDao := &MockCachedSimpleDao{}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("test-list"),
	)
	
	// Context missing the key
	context := map[string]string{
		"otherKey": "value",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when context key is missing")
}

func TestIpAddressEvaluator_Evaluate_EmptyContextValue(t *testing.T) {
	mockDao := &MockCachedSimpleDao{}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("test-list"),
	)
	
	// Context with empty value
	context := map[string]string{
		"ipAddress": "",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when context value is empty")
}

func TestIpAddressEvaluator_Evaluate_NilFixedArg(t *testing.T) {
	mockDao := &MockCachedSimpleDao{}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := &Condition{
		FreeArg:   NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		Operation: StandardOperationIn,
		FixedArg:  nil,
	}
	
	context := map[string]string{
		"ipAddress": "192.168.1.100",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when fixed arg is nil")
}

func TestIpAddressEvaluator_Evaluate_EmptyFixedArgValue(t *testing.T) {
	mockDao := &MockCachedSimpleDao{}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		StandardOperationIn,
		NewFixedArg(""),
	)
	
	context := map[string]string{
		"ipAddress": "192.168.1.100",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when fixed arg value is empty")
}

func TestIpAddressEvaluator_Evaluate_DatabaseError(t *testing.T) {
	mockDao := &MockCachedSimpleDao{
		err: errors.New("database error"),
	}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("test-list"),
	)
	
	context := map[string]string{
		"ipAddress": "192.168.1.100",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when database error occurs")
}

func TestIpAddressEvaluator_Evaluate_ListNotFound(t *testing.T) {
	mockDao := &MockCachedSimpleDao{
		data: map[string]interface{}{},
	}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("nonexistent-list"),
	)
	
	context := map[string]string{
		"ipAddress": "192.168.1.100",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when list is not found")
}

func TestIpAddressEvaluator_Evaluate_WrongListType(t *testing.T) {
	mockDao := &MockCachedSimpleDao{
		data: map[string]interface{}{
			"test-list": "not-a-generic-namespaced-list",
		},
	}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("test-list"),
	)
	
	context := map[string]string{
		"ipAddress": "192.168.1.100",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when list is wrong type")
}

func TestIpAddressEvaluator_Evaluate_MacList_Match(t *testing.T) {
	macList := &shared.GenericNamespacedList{
		TypeName: shared.MacList,
		Data:     []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66", "99:88:77:66:55:44"},
	}
	
	mockDao := &MockCachedSimpleDao{
		data: map[string]interface{}{
			"mac-list": macList,
		},
	}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "eStbMac"),
		StandardOperationIn,
		NewFixedArg("mac-list"),
	)
	
	context := map[string]string{
		"eStbMac": "AA:BB:CC:DD:EE:FF",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, result, "Should return true when MAC address is found in MAC list")
}

func TestIpAddressEvaluator_Evaluate_MacList_NoMatch(t *testing.T) {
	macList := &shared.GenericNamespacedList{
		TypeName: shared.MacList,
		Data:     []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66", "99:88:77:66:55:44"},
	}
	
	mockDao := &MockCachedSimpleDao{
		data: map[string]interface{}{
			"mac-list": macList,
		},
	}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "eStbMac"),
		StandardOperationIn,
		NewFixedArg("mac-list"),
	)
	
	context := map[string]string{
		"eStbMac": "00:00:00:00:00:00",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when MAC address is not found in MAC list")
}

func TestIpAddressEvaluator_Evaluate_IpList_ExactMatch(t *testing.T) {
	ipList := &shared.GenericNamespacedList{
		TypeName: shared.IpList,
		Data:     []string{"192.168.1.100", "10.0.0.1", "172.16.0.1"},
	}
	
	mockDao := &MockCachedSimpleDao{
		data: map[string]interface{}{
			"ip-list": ipList,
		},
	}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("ip-list"),
	)
	
	context := map[string]string{
		"ipAddress": "192.168.1.100",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, result, "Should return true when IP address is found in IP list (exact match)")
}

func TestIpAddressEvaluator_Evaluate_IpList_RangeMatch(t *testing.T) {
	ipList := &shared.GenericNamespacedList{
		TypeName: shared.IpList,
		Data:     []string{"192.168.1.0/24", "10.0.0.0/8"},
	}
	
	mockDao := &MockCachedSimpleDao{
		data: map[string]interface{}{
			"ip-list": ipList,
		},
	}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("ip-list"),
	)
	
	context := map[string]string{
		"ipAddress": "192.168.1.150",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, result, "Should return true when IP address is in range")
}

func TestIpAddressEvaluator_Evaluate_IpList_NoMatch(t *testing.T) {
	ipList := &shared.GenericNamespacedList{
		TypeName: shared.IpList,
		Data:     []string{"192.168.1.0/24", "10.0.0.0/8"},
	}
	
	mockDao := &MockCachedSimpleDao{
		data: map[string]interface{}{
			"ip-list": ipList,
		},
	}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeString, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("ip-list"),
	)
	
	context := map[string]string{
		"ipAddress": "172.16.0.1",
	}
	
	result := evaluator.Evaluate(condition, context)
	assert.Assert(t, !result, "Should return false when IP address is not in any range")
}

func TestIpAddressEvaluator_Evaluate_VoidType(t *testing.T) {
	ipList := &shared.GenericNamespacedList{
		TypeName: shared.IpList,
		Data:     []string{"192.168.1.100"},
	}
	
	mockDao := &MockCachedSimpleDao{
		data: map[string]interface{}{
			"ip-list": ipList,
		},
	}
	
	// Test with VOID type (should skip context checking)
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeVoid, StandardOperationIn, mockDao)
	
	condition := NewCondition(
		NewFreeArg(StandardFreeArgTypeVoid, "ipAddress"),
		StandardOperationIn,
		NewFixedArg("ip-list"),
	)
	
	// VOID type should ignore context
	context := map[string]string{
		"otherKey": "value",
	}
	
	result := evaluator.Evaluate(condition, context)
	// For VOID type, freeArgValue will be empty string, and it should still try to match
	assert.Assert(t, !result, "VOID type with empty freeArgValue should not match")
}

func TestIpAddressEvaluator_Interface_Compliance(t *testing.T) {
	mockDao := &MockCachedSimpleDao{}
	evaluator := NewIpAddressEvaluator(StandardFreeArgTypeString, StandardOperationIn, mockDao)
	
	// Test that it implements IConditionEvaluator interface
	var iface IConditionEvaluator = evaluator
	
	assert.Equal(t, StandardFreeArgTypeString, iface.FreeArgType(), "Interface should return correct FreeArgType")
	assert.Equal(t, StandardOperationIn, iface.Operation(), "Interface should return correct Operation")
	assert.Assert(t, iface != nil, "Interface should not be nil")
}
