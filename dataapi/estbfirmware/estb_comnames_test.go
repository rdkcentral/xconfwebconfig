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
package estbfirmware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test DefaultValue constants
func TestDefaultValueConstants(t *testing.T) {
	t.Run("BLOCKED constant", func(t *testing.T) {
		assert.Equal(t, DefaultValue("BLOCKED"), BLOCKED)
		assert.Equal(t, "BLOCKED", string(BLOCKED))
	})

	t.Run("NOMATCH constant", func(t *testing.T) {
		assert.Equal(t, "NOMATCH", string(NOMATCH))
	})

	t.Run("NORULETYPE constant", func(t *testing.T) {
		assert.Equal(t, "NORULETYPE", string(NORULETYPE))
	})

	t.Run("DefaultValue type is string-based", func(t *testing.T) {
		var val DefaultValue = "TEST"
		assert.Equal(t, "TEST", string(val))
	})
}

// Test constant names for field names
func TestConstantFieldNames(t *testing.T) {
	t.Run("PERCENT_FILTER_NAME", func(t *testing.T) {
		assert.Equal(t, "PercentFilter", PERCENT_FILTER_NAME)
	})

	t.Run("FIRMWARE_SOURCE", func(t *testing.T) {
		assert.Equal(t, "firmwareVersionSource", FIRMWARE_SOURCE)
	})
}

// Test private constant field names (testing through references)
func TestPrivateConstantFieldNames(t *testing.T) {
	// These constants are private but can be tested for their values
	t.Run("envModelPercentagesFieldName value", func(t *testing.T) {
		assert.Equal(t, "envModelPercentages", envModelPercentagesFieldName)
	})

	t.Run("intermediateVersionFieldName value", func(t *testing.T) {
		assert.Equal(t, "intermediateVersion", intermediateVersionFieldName)
	})

	t.Run("lastKnownGoodFieldName value", func(t *testing.T) {
		assert.Equal(t, "lastKnownGood", lastKnownGoodFieldName)
	})
}

// Test DefaultValue comparison
func TestDefaultValueComparison(t *testing.T) {
	t.Run("Compare BLOCKED values", func(t *testing.T) {
		val1 := BLOCKED
		val2 := DefaultValue("BLOCKED")
		assert.Equal(t, val1, val2)
	})

	t.Run("Different DefaultValues are not equal", func(t *testing.T) {
		assert.NotEqual(t, BLOCKED, NOMATCH)
		assert.NotEqual(t, NOMATCH, NORULETYPE)
		assert.NotEqual(t, BLOCKED, NORULETYPE)
	})
}

// Test DefaultValue string conversion
func TestDefaultValueStringConversion(t *testing.T) {
	t.Run("Convert to string", func(t *testing.T) {
		values := []struct {
			val      DefaultValue
			expected string
		}{
			{BLOCKED, "BLOCKED"},
			{NOMATCH, "NOMATCH"},
			{NORULETYPE, "NORULETYPE"},
		}

		for _, tc := range values {
			assert.Equal(t, tc.expected, string(tc.val))
		}
	})
}
