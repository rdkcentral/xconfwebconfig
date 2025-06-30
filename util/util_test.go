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