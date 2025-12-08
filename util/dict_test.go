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
package util

import (
	"strings"
	"testing"
	"time"
)

func TestUtilPrettyPrint(t *testing.T) {
	line := `{"foo":"bar", "enabled": true, "age": 30}`
	t.Logf(PrettyJson(line))

	a := Dict{
		"broadcast_ssid":  true,
		"radio_index":     10100,
		"ssid_enabled":    true,
		"ssid_index":      10101,
		"ssid_name":       "hello",
		"wifi_security":   4,
		"wifi_passphrase": "password1",
	}
	t.Logf(PrettyJson(a))

	b := []Dict{
		Dict{
			"broadcast_ssid":  true,
			"radio_index":     10000,
			"ssid_enabled":    true,
			"ssid_index":      10001,
			"ssid_name":       "ssid_2g",
			"wifi_security":   4,
			"wifi_passphrase": "password2",
		},
		Dict{
			"broadcast_ssid":  true,
			"radio_index":     10100,
			"ssid_enabled":    true,
			"ssid_index":      10101,
			"ssid_name":       "ssid_5g",
			"wifi_security":   4,
			"wifi_passphrase": "password5",
		},
	}
	t.Logf(PrettyJson(b))
}

func TestDict_TimeToMsecs(t *testing.T) {
	// Test with time.Time value
	now := time.Now()
	d := Dict{
		"timeField": now,
	}

	d.TimeToMsecs("timeField")

	// Should convert to int milliseconds
	val, ok := d["timeField"]
	if !ok {
		t.Error("timeField should exist after TimeToMsecs")
	}

	if _, ok := val.(int); !ok {
		t.Errorf("timeField should be converted to int, got %T", val)
	}

	// Test with zero time (should be deleted)
	d["zeroTime"] = time.Time{}
	d.TimeToMsecs("zeroTime")

	if _, exists := d["zeroTime"]; exists {
		t.Error("zeroTime field should be deleted for zero time")
	}

	// Test with non-time field (should remain unchanged)
	d["stringField"] = "not a time"
	original := d["stringField"]
	d.TimeToMsecs("stringField")

	if d["stringField"] != original {
		t.Error("non-time field should remain unchanged")
	}
}

func TestDict_MsecsToTime(t *testing.T) {
	// Test with int milliseconds
	d := Dict{
		"intField": 1609459200000, // 2021-01-01 00:00:00 UTC
	}

	d.MsecsToTime("intField")

	val, ok := d["intField"]
	if !ok {
		t.Error("intField should exist after MsecsToTime")
	}

	timeVal, ok := val.(time.Time)
	if !ok {
		t.Errorf("intField should be converted to time.Time, got %T", val)
	}

	if timeVal.Year() != 2021 {
		t.Errorf("Expected year 2021, got %d", timeVal.Year())
	}

	// Test with float64 milliseconds
	d["floatField"] = float64(1609459200000)
	d.MsecsToTime("floatField")

	timeVal2, ok := d["floatField"].(time.Time)
	if !ok {
		t.Error("floatField should be converted to time.Time")
	}

	if timeVal2.Year() != 2021 {
		t.Errorf("Expected year 2021 from float64, got %d", timeVal2.Year())
	}
}

func TestDict_ToInt(t *testing.T) {
	// Test with int value (should remain int)
	d := Dict{
		"intField": 42,
	}

	d.ToInt("intField")

	val, ok := d["intField"].(int)
	if !ok {
		t.Error("intField should remain as int")
	}

	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	// Test with float64 value
	d["floatField"] = 42.7
	d.ToInt("floatField")

	val2, ok := d["floatField"].(int)
	if !ok {
		t.Error("floatField should be converted to int")
	}

	if val2 != 42 {
		t.Errorf("Expected 42 from float64, got %d", val2)
	}
}

func TestDict_ToInt64(t *testing.T) {
	// Test with int value
	d := Dict{
		"intField": 42,
	}

	d.ToInt64("intField")

	val, ok := d["intField"].(int64)
	if !ok {
		t.Error("intField should be converted to int64")
	}

	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	// Test with int64 value (should remain int64)
	d["int64Field"] = int64(9223372036854775807)
	d.ToInt64("int64Field")

	val3, ok := d["int64Field"].(int64)
	if !ok {
		t.Error("int64Field should remain as int64")
	}

	if val3 != 9223372036854775807 {
		t.Errorf("Expected max int64, got %d", val3)
	}
}

func TestDict_String(t *testing.T) {
	// Test Dict.String method
	d := Dict{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	str := d.String()
	if str == "" {
		t.Error("Dict.String() should not return empty string")
	}

	// Should contain the keys in JSON format
	if !strings.Contains(str, "key1") {
		t.Errorf("Dict.String() should contain key1, got: %s", str)
	}
}

func TestDict_Copy(t *testing.T) {
	// Test Dict.Copy method
	original := Dict{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	copied := original.Copy()

	// Verify copy has same values
	if copied["key1"] != original["key1"] {
		t.Error("Copied dict should have same values")
	}
	if copied["key2"] != original["key2"] {
		t.Error("Copied dict should have same values")
	}
	if copied["key3"] != original["key3"] {
		t.Error("Copied dict should have same values")
	}

	// Verify modifying copy doesn't affect original
	copied["key1"] = "modified"
	if original["key1"] == "modified" {
		t.Error("Modifying copy should not affect original")
	}
}

func TestDict_SelectByKeys(t *testing.T) {
	// Test with existing keys
	original := Dict{
		"key1": "value1",
		"key2": 42,
		"key3": true,
		"key4": "value4",
	}

	selected := original.SelectByKeys("key1", "key3")

	if len(selected) != 2 {
		t.Errorf("Expected 2 keys in selected dict, got %d", len(selected))
	}

	if selected["key1"] != "value1" {
		t.Error("Selected dict should contain key1")
	}

	if selected["key3"] != true {
		t.Error("Selected dict should contain key3")
	}

	if _, exists := selected["key2"]; exists {
		t.Error("Selected dict should not contain key2")
	}

	// Test with non-existing keys
	selected = original.SelectByKeys("nonexistent")
	if len(selected) != 0 {
		t.Errorf("Expected empty dict for non-existing keys, got %d items", len(selected))
	}
}

func TestToInt(t *testing.T) {
	// Test with int
	result := ToInt(42)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test with float64
	result = ToInt(42.7)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test with int64
	result = ToInt(int64(123))
	if result != 123 {
		t.Errorf("Expected 123, got %d", result)
	}

	// Test with unsupported type (should return 0)
	result = ToInt("not a number")
	if result != 0 {
		t.Errorf("Expected 0 for unsupported type, got %d", result)
	}

	// Test with nil (should return 0)
	result = ToInt(nil)
	if result != 0 {
		t.Errorf("Expected 0 for nil, got %d", result)
	}
}
