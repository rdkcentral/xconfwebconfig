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
package dataapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMatchedPrecookHash_WithMatchingRfcHash(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash:          "hash123",
		OfferedFwRfcHash: "hash456",
	}
	result := getMatchedPrecookHash("hash123", precookData, false)
	assert.Equal(t, "hash123", result)
}

func TestGetMatchedPrecookHash_WithMatchingOfferedFwHash(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash:          "hash123",
		OfferedFwRfcHash: "hash456",
	}
	result := getMatchedPrecookHash("hash456", precookData, true)
	assert.Equal(t, "hash456", result)
}

func TestGetMatchedPrecookHash_WithNoMatch(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash:          "hash123",
		OfferedFwRfcHash: "hash456",
	}
	result := getMatchedPrecookHash("hash789", precookData, true)
	assert.Equal(t, "", result)
}

func TestGetMatchedPrecookHash_WithNilPrecookData(t *testing.T) {
	result := getMatchedPrecookHash("hash123", nil, false)
	assert.Equal(t, "", result)
}

func TestGetMatchedPrecookHash_WithEmptyConfigSetHash(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash: "hash123",
	}
	result := getMatchedPrecookHash("", precookData, false)
	assert.Equal(t, "", result)
}

func TestGetMatchedPrecookHash_OfferedFwDisabled(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash:          "hash123",
		OfferedFwRfcHash: "hash456",
	}
	result := getMatchedPrecookHash("hash456", precookData, false)
	assert.Equal(t, "", result)
}

func TestGetMatchedPrecookHash_BothHashesEmpty(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash:          "",
		OfferedFwRfcHash: "",
	}
	result := getMatchedPrecookHash("hash123", precookData, true)
	assert.Equal(t, "", result)
}

func TestGetMatchedPrecookHash_OnlyRfcHashSet(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash:          "hash123",
		OfferedFwRfcHash: "",
	}
	result := getMatchedPrecookHash("hash123", precookData, false)
	assert.Equal(t, "hash123", result)
}

func TestGetMatchedPrecookHash_OnlyOfferedFwHashSet(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash:          "",
		OfferedFwRfcHash: "hash456",
	}
	result := getMatchedPrecookHash("hash456", precookData, true)
	assert.Equal(t, "hash456", result)
}

func TestGetMatchedPrecookHash_CaseSensitivity(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash:          "HASH123",
		OfferedFwRfcHash: "hash456",
	}
	result := getMatchedPrecookHash("hash123", precookData, false)
	assert.Equal(t, "", result)
}

func TestGetMatchedPrecookHash_MatchRfcHashWhenBothEnabled(t *testing.T) {
	precookData := &PreprocessedData{
		RfcHash:          "hash123",
		OfferedFwRfcHash: "hash456",
	}
	result := getMatchedPrecookHash("hash123", precookData, true)
	assert.Equal(t, "hash123", result)
}
