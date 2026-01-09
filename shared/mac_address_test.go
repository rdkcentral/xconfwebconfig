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
package shared

import (
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestMacAddress(t *testing.T) {
	expected := "F8:A0:97:1E:D6:74"

	// parse "F8A0971ED674" and lower
	s1 := "F8A0971ED674"
	m1, err := NewMacAddress(s1)
	assert.NilError(t, err)
	assert.Equal(t, m1.String(), expected)

	m1, err = NewMacAddress("f8A0971ed674")
	assert.NilError(t, err)
	assert.Equal(t, m1.String(), expected)

	m2, err := NewMacAddress(strings.ToLower(s1))
	assert.NilError(t, err)
	assert.Equal(t, m2.String(), expected)

	// parse "F8:A0:97:1E:D6:74" lower
	m3, err := NewMacAddress(expected)
	assert.NilError(t, err)
	assert.Equal(t, m3.String(), expected)

	m4, err := NewMacAddress(strings.ToLower(expected))
	assert.NilError(t, err)
	assert.Equal(t, m4.String(), expected)

	// parse "F8-A0-97-1E-D6-74"
	s5 := "F8-A0-97-1E-D6-74"
	m5, err := NewMacAddress(s5)
	assert.NilError(t, err)
	assert.Equal(t, m5.String(), expected)

	m6, err := NewMacAddress(strings.ToLower(s5))
	assert.NilError(t, err)
	assert.Equal(t, m6.String(), expected)

	// error
	_, err = NewMacAddress("F8A0971ED674A")
	assert.Assert(t, err != nil)

	_, err = NewMacAddress("X8A0971ED674")
	assert.Assert(t, err != nil)

	_, err = NewMacAddress("x8:a0971ED674")
	assert.Assert(t, err != nil)
}

// Test additional MAC address formats
func TestMacAddress_VariousFormats(t *testing.T) {
	expected := "AA:BB:CC:DD:EE:FF"

	// Test different valid formats
	formats := []string{
		"AA:BB:CC:DD:EE:FF",
		"aa:bb:cc:dd:ee:ff",
		"AABBCCDDEEFF",
		"aabbccddeeff",
		"AA-BB-CC-DD-EE-FF",
		"aa-bb-cc-dd-ee-ff",
		"Aa:Bb:Cc:Dd:Ee:Ff", // Mixed case
	}

	for _, format := range formats {
		mac, err := NewMacAddress(format)
		assert.NilError(t, err)
		assert.Equal(t, expected, mac.String())
	}
}

// Test invalid MAC addresses
func TestMacAddress_InvalidFormats(t *testing.T) {
	invalidMACs := []string{
		"",                     // Empty
		"AABBCCDDEEF",          // Too short (11 chars)
		"AABBCCDDEEFFFFF",      // Too long (15 chars)
		"GG:BB:CC:DD:EE:FF",    // Invalid hex character
		"AA:BB:CC:DD:EE",       // Missing byte
		"AA:BB:CC:DD:EE:FF:11", // Too many bytes
		"AABBCCDDEE",           // 10 chars without separator
		"AA:BB:CC",             // Only 3 bytes
		"not-a-mac",            // Invalid format
		"12-34-56-78-90",       // 5 bytes with dash
		"12:34:56:78:90",       // 5 bytes with colon
	}

	for _, invalid := range invalidMACs {
		_, err := NewMacAddress(invalid)
		assert.Assert(t, err != nil, "Expected error for: %s", invalid)
	}
}

// Test MAC address string output format
func TestMacAddress_StringOutput(t *testing.T) {
	// Input with lowercase should output uppercase
	mac, err := NewMacAddress("aa:bb:cc:dd:ee:ff")
	assert.NilError(t, err)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", mac.String())

	// Already uppercase
	mac2, err := NewMacAddress("11:22:33:44:55:66")
	assert.NilError(t, err)
	assert.Equal(t, "11:22:33:44:55:66", mac2.String())
}

// Test edge cases
func TestMacAddress_EdgeCases(t *testing.T) {
	// All zeros
	mac1, err := NewMacAddress("00:00:00:00:00:00")
	assert.NilError(t, err)
	assert.Equal(t, "00:00:00:00:00:00", mac1.String())

	// All F's
	mac2, err := NewMacAddress("FF:FF:FF:FF:FF:FF")
	assert.NilError(t, err)
	assert.Equal(t, "FF:FF:FF:FF:FF:FF", mac2.String())

	// Mixed case input
	mac3, err := NewMacAddress("aA:bB:cC:dD:eE:fF")
	assert.NilError(t, err)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", mac3.String())
}

// Test 12-character format without separators
func TestMacAddress_12CharFormat(t *testing.T) {
	// Uppercase 12-char format
	mac1, err := NewMacAddress("AABBCCDDEEFF")
	assert.NilError(t, err)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", mac1.String())

	// Lowercase 12-char format
	mac2, err := NewMacAddress("aabbccddeeff")
	assert.NilError(t, err)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", mac2.String())

	// Mixed case 12-char format
	mac3, err := NewMacAddress("AaBbCcDdEeFf")
	assert.NilError(t, err)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", mac3.String())
}

// Test dash-separated format
func TestMacAddress_DashFormat(t *testing.T) {
	mac, err := NewMacAddress("AA-BB-CC-DD-EE-FF")
	assert.NilError(t, err)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", mac.String())

	// Lowercase with dashes
	mac2, err := NewMacAddress("aa-bb-cc-dd-ee-ff")
	assert.NilError(t, err)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", mac2.String())
}

// Test real-world MAC addresses
func TestMacAddress_RealWorldExamples(t *testing.T) {
	realMACs := map[string]string{
		"00:1A:2B:3C:4D:5E": "00:1A:2B:3C:4D:5E",
		"F8:A0:97:1E:D6:74": "F8:A0:97:1E:D6:74",
		"00:50:56:C0:00:08": "00:50:56:C0:00:08",
		"08:00:27:12:34:56": "08:00:27:12:34:56",
	}

	for input, expected := range realMACs {
		mac, err := NewMacAddress(input)
		assert.NilError(t, err)
		assert.Equal(t, expected, mac.String())
	}
}
