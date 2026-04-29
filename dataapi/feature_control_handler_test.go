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

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/stretchr/testify/assert"
)

func defaultRuleEvalContext() ruleEvalContext {
	return ruleEvalContext{
		isLiveCalculated:                true,
		isPrecook304:                    false,
		canPrecookRfcResponse:           true,
		precookDataFetched:              true,
		isPrecookLockdownMode:           false,
		isMacExcludedFromPrecook:        false,
		contextHashMatched:              true,
		ipInSameNetwork:                 true,
		isFwVersionMatched:              true,
		isRfcPrecookForOfferedFwEnabled: false,
		isOfferedFwMatched:              false,
		contextMap: map[string]string{
			common.ACCOUNT_ID:       "acct-1",
			common.PARTNER_ID:       "partner-a",
			common.FIRMWARE_VERSION: "fw-1",
		},
		precookData: &PreprocessedData{
			RfcHash:   "hash-1",
			AccountId: "acct-1",
			PartnerId: "partner-a",
			FwVersion: "fw-1",
		},
	}
}

func TestDeriveRuleEvalReason_Precook304(t *testing.T) {
	evalCtx := defaultRuleEvalContext()
	evalCtx.isLiveCalculated = false
	evalCtx.isPrecook304 = true

	assert.Equal(t, "precook-304", deriveRuleEvalReason(evalCtx))
}

func TestDeriveRuleEvalReason_OfferedFirmwareHit(t *testing.T) {
	evalCtx := defaultRuleEvalContext()
	evalCtx.isLiveCalculated = false
	evalCtx.isRfcPrecookForOfferedFwEnabled = true
	evalCtx.isOfferedFwMatched = true

	assert.Equal(t, "precook-offered-firmware", deriveRuleEvalReason(evalCtx))
}

func TestDeriveRuleEvalReason_PrecookDisabledVsFirstContact(t *testing.T) {
	t.Run("precook disabled", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.precookDataFetched = false
		evalCtx.precookData = nil

		assert.Equal(t, "precook-off", deriveRuleEvalReason(evalCtx))
	})

	t.Run("first contact", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.precookDataFetched = true
		evalCtx.precookData = nil

		assert.Equal(t, "first-contact", deriveRuleEvalReason(evalCtx))
	})
}

func TestDeriveRuleEvalReason_ContextHashMismatchReasons(t *testing.T) {
	t.Run("account-new", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.contextHashMatched = false
		evalCtx.precookData.AccountId = ""
		evalCtx.contextMap[common.ACCOUNT_ID] = "acct-new"
		assert.Equal(t, "account-new", deriveRuleEvalReason(evalCtx))
	})

	t.Run("account-deleted", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.contextHashMatched = false
		evalCtx.precookData.AccountId = "acct-old"
		evalCtx.contextMap[common.ACCOUNT_ID] = ""
		assert.Equal(t, "account-deleted", deriveRuleEvalReason(evalCtx))
	})

	t.Run("account-change", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.contextHashMatched = false
		evalCtx.precookData.AccountId = "acct-old"
		evalCtx.contextMap[common.ACCOUNT_ID] = "acct-new"
		assert.Equal(t, "account-change", deriveRuleEvalReason(evalCtx))
	})

	t.Run("partner-change", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.contextHashMatched = false
		evalCtx.contextMap[common.ACCOUNT_ID] = evalCtx.precookData.AccountId
		evalCtx.contextMap[common.PARTNER_ID] = "partner-b"
		assert.Equal(t, "partner-change", deriveRuleEvalReason(evalCtx))
	})

	t.Run("firmware-change", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.contextHashMatched = false
		evalCtx.contextMap[common.ACCOUNT_ID] = evalCtx.precookData.AccountId
		evalCtx.contextMap[common.PARTNER_ID] = evalCtx.precookData.PartnerId
		evalCtx.contextMap[common.FIRMWARE_VERSION] = "fw-2"
		assert.Equal(t, "firmware-change", deriveRuleEvalReason(evalCtx))
	})

	t.Run("context-change", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.contextHashMatched = false
		evalCtx.contextMap[common.ACCOUNT_ID] = evalCtx.precookData.AccountId
		evalCtx.contextMap[common.PARTNER_ID] = evalCtx.precookData.PartnerId
		evalCtx.contextMap[common.FIRMWARE_VERSION] = evalCtx.precookData.FwVersion
		assert.Equal(t, "context-change", deriveRuleEvalReason(evalCtx))
	})
}

func TestDeriveRuleEvalReason_IpAndFirmwareChange(t *testing.T) {
	t.Run("ip-change", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.ipInSameNetwork = false
		assert.Equal(t, "ip-change", deriveRuleEvalReason(evalCtx))
	})

	t.Run("firmware-change after hash match", func(t *testing.T) {
		evalCtx := defaultRuleEvalContext()
		evalCtx.isFwVersionMatched = false
		evalCtx.isRfcPrecookForOfferedFwEnabled = false
		evalCtx.isOfferedFwMatched = false
		assert.Equal(t, "firmware-change", deriveRuleEvalReason(evalCtx))
	})
}

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
