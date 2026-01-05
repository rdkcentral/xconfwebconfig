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
	"bytes"
	"testing"

	"gotest.tools/assert"
)

type TestStruct struct {
	TestVar1 string `json:"var_one"`
	TestVar2 bool   `json:"var_two"`
}

func TestJsonMarshal(t *testing.T) {
	tmp := TestStruct{
		TestVar1: "test&string 1>0",
		TestVar2: false,
	}
	testdata := []byte(`{"var_one":"test&string 1>0","var_two":false}`)

	adata, err := JSONMarshal(tmp)
	assert.NilError(t, err)

	res := bytes.Compare(adata, testdata)
	assert.Equal(t, res, 1)

	list := []string{}
	adata, err = JSONMarshal(list)
	assert.NilError(t, err)
}

func TestApiVersionGreaterOrEqual(t *testing.T) {
	ver := "1.3"
	var val float32
	val = 1.4

	res := ApiVersionGreaterOrEqual(ver, val)
	assert.Equal(t, res, false)

	val = 1.1
	res = ApiVersionGreaterOrEqual(ver, val)
	assert.Equal(t, res, true)

	val = 1.3
	res = ApiVersionGreaterOrEqual(ver, val)
	assert.Equal(t, res, true)

	ver = "testversion"
	res = ApiVersionGreaterOrEqual(ver, val)
	assert.Equal(t, res, false)
}

func TestGetCRC32HashValue(t *testing.T) {
	// tests from known input and output combinations
	// Trivial one
	input := ""
	expectedOutput := "00000000"
	actualOutput := GetCRC32HashValue(input)
	assert.Equal(t, expectedOutput, actualOutput)

	// Source: https://rosettacode.org/wiki/CRC-32
	// input = "The quick brown fox jumps over the lazy dog"
	input = "The quick brown fox jumps over the lazy dog"
	expectedOutput = "414fa339"
	actualOutput = GetCRC32HashValue(input)
	assert.Equal(t, expectedOutput, actualOutput)

	// Source: http://cryptomanager.com/tv.html
	input = "various CRC algorithms input data"
	expectedOutput = "9bd366ae"
	actualOutput = GetCRC32HashValue(input)
	assert.Equal(t, expectedOutput, actualOutput)

	// Source: http://www.febooti.com/products/filetweak/members/hash-and-crc/test-vectors/
	input = "Test vector from febooti.com"
	expectedOutput = "0c877f61"
	actualOutput = GetCRC32HashValue(input)
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestIsVersionGreaterOrEqual(t *testing.T) {
	version := ""
	value := 2.5
	isVersionGreaterOrEqual := IsVersionGreaterOrEqual(version, value)
	assert.Equal(t, isVersionGreaterOrEqual, false)

	version = "hello"
	isVersionGreaterOrEqual = IsVersionGreaterOrEqual(version, value)
	assert.Equal(t, isVersionGreaterOrEqual, false)

	version = "2"
	isVersionGreaterOrEqual = IsVersionGreaterOrEqual(version, value)
	assert.Equal(t, isVersionGreaterOrEqual, false)

	version = "2.5"
	isVersionGreaterOrEqual = IsVersionGreaterOrEqual(version, value)
	assert.Equal(t, isVersionGreaterOrEqual, true)

	version = "3.45"
	isVersionGreaterOrEqual = IsVersionGreaterOrEqual(version, value)
	assert.Equal(t, isVersionGreaterOrEqual, true)
}

func TestUtcCurrentTimestamp(t *testing.T) {
	timestamp := UtcCurrentTimestamp()

	// Verify the timezone is UTC
	zone, _ := timestamp.Zone()
	assert.Equal(t, zone, "UTC")

	// Verify it's a valid timestamp (not zero)
	assert.Assert(t, !timestamp.IsZero())
}

func TestUtcOffsetTimestamp(t *testing.T) {
	// Test with positive offset
	future := UtcOffsetTimestamp(10)
	current := UtcCurrentTimestamp()

	// Future should be ahead of current (accounting for execution time)
	assert.Assert(t, future.After(current))

	// Test with negative offset
	past := UtcOffsetTimestamp(-10)
	assert.Assert(t, past.Before(current))
}

func TestUtcTimeInNano(t *testing.T) {
	nanoTime := UtcTimeInNano()

	// Should be a positive number
	assert.Assert(t, nanoTime > 0)

	// Should be a realistic nanosecond timestamp
	assert.Assert(t, nanoTime > 1600000000000000000) // After year 2020
}

func TestCopy(t *testing.T) {
	// Test copying a simple struct
	original := TestStruct{
		TestVar1: "original",
		TestVar2: true,
	}

	copied, err := Copy(original)
	assert.NilError(t, err)

	copiedStruct := copied.(TestStruct)
	assert.Equal(t, copiedStruct.TestVar1, original.TestVar1)
	assert.Equal(t, copiedStruct.TestVar2, original.TestVar2)

	// Test with a simple value that copy can handle
	originalValue := 42
	copiedValue, err := Copy(originalValue)
	assert.NilError(t, err)
	assert.Equal(t, copiedValue.(int), originalValue)
}

func TestGetTimestamp(t *testing.T) {
	// Test with no arguments (current time)
	timestamp1 := GetTimestamp()
	assert.Assert(t, timestamp1 > 0)

	// Test with specific time
	specificTime := UtcCurrentTimestamp()
	timestamp2 := GetTimestamp(specificTime)

	// Should be positive and realistic
	assert.Assert(t, timestamp2 > 0)
	assert.Assert(t, timestamp2 > 1600000000000) // After year 2020 in milliseconds
}

func TestIsValidAppSetting(t *testing.T) {
	// This will test with some common app setting key that should exist
	// We'll test with empty string and some typical keys
	result := IsValidAppSetting("")
	assert.Assert(t, !result) // Empty string should not be valid

	result = IsValidAppSetting("nonexistent_key_12345")
	assert.Assert(t, !result) // Random key should not be valid
}

func TestRemoveNonAlphabeticSymbols(t *testing.T) {
	// Test with MAC address with colons
	result := RemoveNonAlphabeticSymbols("aa:bb:cc:dd:ee:ff")
	assert.Equal(t, result, "AABBCCDDEEFF")

	// Test with MAC address with hyphens
	result = RemoveNonAlphabeticSymbols("aa-bb-cc-dd-ee-ff")
	assert.Equal(t, result, "AABBCCDDEEFF")

	// Test with mixed case and spaces
	result = RemoveNonAlphabeticSymbols("  aA:bB-cC  ")
	assert.Equal(t, result, "AABBCC")

	// Test with empty string
	result = RemoveNonAlphabeticSymbols("")
	assert.Equal(t, result, "")
}
