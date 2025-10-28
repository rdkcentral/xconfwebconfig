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
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/util"
	"github.com/stretchr/testify/assert"
)

// Test NewContextDataFromContextMap
func TestNewContextDataFromContextMap_AllFieldsPresent(t *testing.T) {
	contextMap := map[string]string{
		common.ESTB_MAC_ADDRESS: "AA:BB:CC:DD:EE:FF",
		common.MODEL:            "MODEL123",
		common.PARTNER_ID:       "partner1",
		common.SERIAL_NUM:       "SN12345",
		common.EXPERIENCE:       "exp1",
		common.ACCOUNT_ID:       "acc123",
	}
	tags := []string{"tag1", "tag2"}

	result := NewContextDataFromContextMap(contextMap, tags)

	assert.Equal(t, "AA:BB:CC:DD:EE:FF", result.Mac)
	assert.Equal(t, "MODEL123", result.Model)
	assert.Equal(t, "partner1", result.Partner)
	assert.Equal(t, "SN12345", result.SerialNum)
	assert.Equal(t, "exp1", result.Experience)
	assert.Equal(t, "acc123", result.AccountId)
	assert.Equal(t, []string{"tag1", "tag2"}, result.Tags)
}

func TestNewContextDataFromContextMap_EmptyContextMap(t *testing.T) {
	contextMap := map[string]string{}
	tags := []string{}

	result := NewContextDataFromContextMap(contextMap, tags)

	assert.Equal(t, "", result.Mac)
	assert.Equal(t, "", result.Model)
	assert.Equal(t, []string{}, result.Tags)
}

func TestNewContextDataFromContextMap_NilTags(t *testing.T) {
	contextMap := map[string]string{
		common.ESTB_MAC_ADDRESS: "AA:BB:CC:DD:EE:FF",
	}

	result := NewContextDataFromContextMap(contextMap, nil)

	assert.Equal(t, "AA:BB:CC:DD:EE:FF", result.Mac)
	assert.Nil(t, result.Tags)
}

// Test CalculateHashForContextData
func TestCalculateHashForContextData_WithTags(t *testing.T) {
	data := ContextData{
		Mac:     "AA:BB:CC:DD:EE:FF",
		Model:   "MODEL123",
		Partner: "partner1",
		Tags:    []string{"tag2", "tag1", "tag3"}, // Unsorted
	}

	hash, err := CalculateHashForContextData(data)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestCalculateHashForContextData_TagsSorted(t *testing.T) {
	// Test that tags are sorted before hashing
	data1 := ContextData{
		Mac:   "AA:BB:CC:DD:EE:FF",
		Model: "MODEL123",
		Tags:  []string{"tag3", "tag1", "tag2"},
	}

	data2 := ContextData{
		Mac:   "AA:BB:CC:DD:EE:FF",
		Model: "MODEL123",
		Tags:  []string{"tag1", "tag2", "tag3"},
	}

	hash1, err1 := CalculateHashForContextData(data1)
	hash2, err2 := CalculateHashForContextData(data2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, hash1, hash2, "Hashes should be equal regardless of tag order")
}

func TestCalculateHashForContextData_DifferentDataDifferentHash(t *testing.T) {
	data1 := ContextData{Mac: "AA:BB:CC:DD:EE:FF", Model: "MODEL123"}
	data2 := ContextData{Mac: "AA:BB:CC:DD:EE:FF", Model: "MODEL456"}

	hash1, _ := CalculateHashForContextData(data1)
	hash2, _ := CalculateHashForContextData(data2)

	assert.NotEqual(t, hash1, hash2)
}

// Test CompareHashWithXDAS
func TestCompareHashWithXDAS_MatchingHashes(t *testing.T) {
	contextMap := map[string]string{
		common.ESTB_MAC_ADDRESS: "AA:BB:CC:DD:EE:FF",
		common.MODEL:            "MODEL123",
	}
	tags := []string{"tag1"}

	// Calculate expected hash
	contextData := NewContextDataFromContextMap(contextMap, tags)
	expectedHash, _ := CalculateHashForContextData(contextData)

	result, err := CompareHashWithXDAS(contextMap, expectedHash, tags)

	assert.NoError(t, err)
	assert.True(t, result)
}

func TestCompareHashWithXDAS_NonMatchingHashes(t *testing.T) {
	contextMap := map[string]string{
		common.ESTB_MAC_ADDRESS: "AA:BB:CC:DD:EE:FF",
	}

	result, err := CompareHashWithXDAS(contextMap, "differenthash", nil)

	assert.NoError(t, err)
	assert.False(t, result)
}

// Test CreateAccountIdFeature
func TestCreateAccountIdFeature_ValidAccountId(t *testing.T) {
	feature := CreateAccountIdFeature("acc12345")

	assert.Equal(t, "AccountId", feature.Name)
	assert.Equal(t, "AccountId", feature.FeatureName)
	assert.True(t, feature.EffectiveImmediate)
	assert.True(t, feature.Enable)
	assert.Equal(t, "acc12345", feature.ConfigData[common.TR181_DEVICE_TYPE_ACCOUNT_ID])
}

func TestCreateAccountIdFeature_EmptyAccountId(t *testing.T) {
	feature := CreateAccountIdFeature("")

	assert.Equal(t, "AccountId", feature.Name)
	assert.Equal(t, "", feature.ConfigData[common.TR181_DEVICE_TYPE_ACCOUNT_ID])
}

// Test CreatePartnerIdFeature
func TestCreatePartnerIdFeature_ValidPartnerId(t *testing.T) {
	feature := CreatePartnerIdFeature("comcast")

	assert.Equal(t, common.SYNDICATION_PARTNER, feature.Name)
	assert.True(t, feature.EffectiveImmediate)
	assert.Equal(t, "comcast", feature.ConfigData[common.TR181_DEVICE_TYPE_PARTNER_ID])
}

// Test GetTimezoneOffset
func TestGetTimezoneOffset_ReturnsValidFormat(t *testing.T) {
	offset := GetTimezoneOffset()

	assert.NotEmpty(t, offset)
	assert.Contains(t, offset, "UTC")
	assert.Regexp(t, `^UTC[+-]\d{2}:\d{2}$`, offset)
}

// Test createUnknownAccountIdFeature
func TestCreateUnknownAccountIdFeature_WithPassedPartnerId(t *testing.T) {
	contextMap := map[string]string{
		common.PASSED_PARTNER_ID: "test-partner",
	}

	response := createUnknownAccountIdFeature(contextMap)

	assert.Equal(t, "unknown", response["configData"].(map[string]string)[common.TR181_DEVICE_TYPE_ACCOUNT_ID])
	assert.Equal(t, "test-partner", response["partnerId"])
	assert.Equal(t, ATHENS_EUROPE_TZ, response["timeZone"])
	assert.NotEmpty(t, response["tzUTCOffset"])
}

func TestCreateUnknownAccountIdFeature_WithUnknownPartnerId(t *testing.T) {
	contextMap := map[string]string{
		common.PASSED_PARTNER_ID: "unknown",
	}

	response := createUnknownAccountIdFeature(contextMap)

	assert.Equal(t, "dt-gr", response["partnerId"])
	assert.Equal(t, ATHENS_EUROPE_TZ, response["timeZone"])
}

// Test PostProcessFeatureControl
func TestPostProcessFeatureControl_XPCWithGRPrefix(t *testing.T) {
	// Save and restore Xc
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{
		RfcReturnCountryCode:      false,
		ReturnAccountHash:         false,
		ReturnAccountId:           false,
		RfcCountryCodeModelsSet:   util.NewSet(),
		RfcCountryCodePartnersSet: util.NewSet(),
	}

	contextMap := map[string]string{
		common.ACCOUNT_MGMT:      "xpc",
		common.MODEL:             "GRMODEL123",
		common.PASSED_PARTNER_ID: "unknown",
	}

	result := PostProcessFeatureControl(nil, contextMap, false, nil)

	assert.NotEmpty(t, result)
	assert.GreaterOrEqual(t, len(result), 1)
}

func TestPostProcessFeatureControl_WithAccountHash(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{
		ReturnAccountHash:         true,
		RfcReturnCountryCode:      false,
		RfcCountryCodeModelsSet:   util.NewSet(),
		RfcCountryCodePartnersSet: util.NewSet(),
	}

	contextMap := map[string]string{
		common.ACCOUNT_HASH:      "hash123",
		common.PASSED_PARTNER_ID: "partner1",
	}

	result := PostProcessFeatureControl(nil, contextMap, false, nil)

	// With account hash set and ReturnAccountHash enabled, should have at least one feature
	assert.NotEmpty(t, result)
}

func TestPostProcessFeatureControl_SecureConnectionWithPodData(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{
		ReturnAccountId:           true,
		RfcReturnCountryCode:      false,
		ReturnAccountHash:         false,
		RfcCountryCodeModelsSet:   util.NewSet(),
		RfcCountryCodePartnersSet: util.NewSet(),
	}

	contextMap := map[string]string{
		common.ACCOUNT_ID:        "acc123",
		common.PASSED_PARTNER_ID: "partner1",
		common.MODEL:             "MODEL123",
	}

	podData := &PodData{
		AccountId: "acc123",
		PartnerId: "PARTNER1",
		TimeZone:  "America/New_York",
	}

	result := PostProcessFeatureControl(nil, contextMap, true, podData)

	// With secure connection and account ID, should have at least one feature
	assert.NotEmpty(t, result)
} // Test canPrecookRfcResponses
func TestCanPrecookRfcResponses_DisabledPrecook(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{
		EnableRfcPrecook: false,
	}

	result := canPrecookRfcResponses(nil)

	assert.False(t, result)
}

func TestCanPrecookRfcResponses_EnabledWithNoTimeWindow(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{
		EnableRfcPrecook:    true,
		RfcPrecookStartTime: "",
		RfcPrecookEndTime:   "",
		RfcPrecookTimeZone:  time.UTC,
	}

	result := canPrecookRfcResponses(nil)

	assert.True(t, result)
}

// Test struct instantiation
func TestPodData_Structure(t *testing.T) {
	podData := PodData{
		AccountId: "acc123",
		TimeZone:  "America/New_York",
		PartnerId: "partner1",
	}

	assert.Equal(t, "acc123", podData.AccountId)
	assert.Equal(t, "America/New_York", podData.TimeZone)
	assert.Equal(t, "partner1", podData.PartnerId)
}

func TestAccountServiceData_Structure(t *testing.T) {
	accountData := AccountServiceData{
		AccountId: "acc123",
		TimeZone:  "America/New_York",
		PartnerId: "partner1",
	}

	assert.Equal(t, "acc123", accountData.AccountId)
	assert.Equal(t, "America/New_York", accountData.TimeZone)
	assert.Equal(t, "partner1", accountData.PartnerId)
}

func TestPrecookData_Structure(t *testing.T) {
	precookData := PrecookData{
		AccountId:       "acc123",
		PartnerId:       "partner1",
		Model:           "MODEL123",
		ApplicationType: "stb",
		FwVersion:       "1.0.0",
	}

	assert.Equal(t, "acc123", precookData.AccountId)
	assert.Equal(t, "partner1", precookData.PartnerId)
	assert.Equal(t, "MODEL123", precookData.Model)
	assert.Equal(t, "stb", precookData.ApplicationType)
	assert.Equal(t, "1.0.0", precookData.FwVersion)
}
