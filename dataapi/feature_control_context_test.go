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
}

func TestPostProcessFeatureControl_WithCountryCodeEnabled(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()

	modelsSet := util.NewSet()
	modelsSet.Add("MODEL123")
	partnersSet := util.NewSet()
	partnersSet.Add("partner1")

	Xc = &XconfConfigs{
		ReturnAccountId:           true,
		RfcReturnCountryCode:      true,
		ReturnAccountHash:         false,
		RfcCountryCodeModelsSet:   modelsSet,
		RfcCountryCodePartnersSet: partnersSet,
	}

	contextMap := map[string]string{
		common.ACCOUNT_ID:        "acc123",
		common.MODEL:             "MODEL123",
		common.PARTNER_ID:        "partner1",
		common.COUNTRY_CODE:      "US",
		common.PASSED_PARTNER_ID: "partner1",
	}

	result := PostProcessFeatureControl(nil, contextMap, true, nil)

	// Should include country code feature
	assert.NotEmpty(t, result)
	// Verify country code is in one of the responses
	foundCountryCode := false
	for _, resp := range result {
		if cc, exists := resp[common.COUNTRY_CODE]; exists && cc == "US" {
			foundCountryCode = true
			break
		}
	}
	assert.True(t, foundCountryCode, "Country code should be included in response")
}

func TestPostProcessFeatureControl_PodDataWithInvalidTimezone(t *testing.T) {
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
	}

	podData := &PodData{
		AccountId: "acc123",
		PartnerId: "partner1",
		TimeZone:  "Invalid/Timezone",
	}

	result := PostProcessFeatureControl(nil, contextMap, true, podData)

	assert.NotEmpty(t, result)
	// With invalid timezone, tzUTCOffset should be "unknown"
	for _, resp := range result {
		if tzOffset, exists := resp["tzUTCOffset"]; exists {
			assert.Equal(t, "unknown", tzOffset, "Invalid timezone should result in unknown offset")
		}
	}
}

func TestPostProcessFeatureControl_PodDataWithEmptyTimezone(t *testing.T) {
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
	}

	podData := &PodData{
		AccountId: "acc123",
		PartnerId: "partner1",
		TimeZone:  "",
	}

	result := PostProcessFeatureControl(nil, contextMap, true, podData)

	assert.NotEmpty(t, result)
	// With empty timezone, both timeZone and tzUTCOffset should be "unknown"
	for _, resp := range result {
		if tz, exists := resp["timeZone"]; exists {
			assert.Equal(t, "unknown", tz, "Empty timezone should result in unknown")
		}
		if tzOffset, exists := resp["tzUTCOffset"]; exists {
			assert.Equal(t, "unknown", tzOffset, "Empty timezone should result in unknown offset")
		}
	}
}

func TestPostProcessFeatureControl_PodDataWithEmptyPartnerId(t *testing.T) {
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
	}

	podData := &PodData{
		AccountId: "acc123",
		PartnerId: "",
		TimeZone:  "America/New_York",
	}

	result := PostProcessFeatureControl(nil, contextMap, true, podData)

	assert.NotEmpty(t, result)
	// With empty partnerId, should be set to "unknown"
	for _, resp := range result {
		if pid, exists := resp["partnerId"]; exists {
			assert.Equal(t, "unknown", pid, "Empty partnerId should result in unknown")
		}
	}
}

func TestPostProcessFeatureControl_NonXPCWithUnknownPartner(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{
		ReturnAccountId:           false,
		RfcReturnCountryCode:      false,
		ReturnAccountHash:         false,
		RfcCountryCodeModelsSet:   util.NewSet(),
		RfcCountryCodePartnersSet: util.NewSet(),
	}

	contextMap := map[string]string{
		common.ACCOUNT_MGMT:      "notxpc",
		common.MODEL:             "MODEL123",
		common.PASSED_PARTNER_ID: "unknown",
		common.PARTNER_ID:        "realpartner",
	}

	result := PostProcessFeatureControl(nil, contextMap, false, nil)

	// When passed_partner is unknown but partner_id exists, should create partner feature
	assert.NotEmpty(t, result)
}

func TestPostProcessFeatureControl_XPCWithGRAndPassedPartner(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{
		ReturnAccountId:           false,
		RfcReturnCountryCode:      false,
		ReturnAccountHash:         false,
		RfcCountryCodeModelsSet:   util.NewSet(),
		RfcCountryCodePartnersSet: util.NewSet(),
	}

	contextMap := map[string]string{
		common.ACCOUNT_MGMT:      "xpc",
		common.MODEL:             "GRMODEL123",
		common.PASSED_PARTNER_ID: "passedpartner",
	}

	result := PostProcessFeatureControl(nil, contextMap, false, nil)

	// With xpc and GR prefix, should have partner features
	assert.NotEmpty(t, result)
	// Should have at least 2 features (partner and timezone)
	assert.GreaterOrEqual(t, len(result), 2)
}

// Test canPrecookRfcResponses
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

func TestCanPrecookRfcResponses_EnabledWithTimeWindowInside(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()

	// Get current time and create a window around it
	now := time.Now().In(time.UTC)
	// Use HH:MM format (hours:minutes without seconds)
	nowStr := now.Format("15:04")
	// Add 1 hour to endTime
	endTime := now.Add(1 * time.Hour).Format("15:04")
	// Subtract 1 hour from startTime
	startTime := now.Add(-1 * time.Hour).Format("15:04")

	Xc = &XconfConfigs{
		EnableRfcPrecook:     true,
		RfcPrecookStartTime:  startTime,
		RfcPrecookEndTime:    endTime,
		RfcPrecookTimeZone:   time.UTC,
		RfcPrecookTimeFormat: "15:04",
	}

	result := canPrecookRfcResponses(nil)

	assert.True(t, result, "Current time %s should be within window %s to %s", nowStr, startTime, endTime)
}

func TestCanPrecookRfcResponses_EnabledWithTimeWindowOutside(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()

	// Create a time window in the future
	now := time.Now().In(time.UTC)
	// Start 2 hours in future, end 3 hours in future
	startTime := now.Add(2 * time.Hour).Format("15:04")
	endTime := now.Add(3 * time.Hour).Format("15:04")

	Xc = &XconfConfigs{
		EnableRfcPrecook:     true,
		RfcPrecookStartTime:  startTime,
		RfcPrecookEndTime:    endTime,
		RfcPrecookTimeZone:   time.UTC,
		RfcPrecookTimeFormat: "15:04",
	}

	result := canPrecookRfcResponses(nil)

	assert.False(t, result, "Current time should be outside window %s to %s", startTime, endTime)
}

func TestCanPrecookRfcResponses_EnabledWithReverseTimeWindowInside(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()

	// Reverse window: startTime > endTime (e.g., 22:00 to 02:00 - overnight window)
	// Test when current time is after startTime (e.g., 23:00)
	now := time.Now().In(time.UTC)
	// Set startTime 1 hour before current time
	startTime := now.Add(-1 * time.Hour).Format("15:04")
	// Set endTime far in future (this makes startTime < endTime, not reverse)
	// For a TRUE reverse window, endTime must be LESS than startTime
	// e.g., startTime="23:00", endTime="01:00"
	endTime := now.Add(-2 * time.Hour).Format("15:04")

	// This ensures startTime > endTime (reverse window)
	// If endTime is before startTime, it's overnight: current > start OR current < end

	Xc = &XconfConfigs{
		EnableRfcPrecook:     true,
		RfcPrecookStartTime:  startTime,
		RfcPrecookEndTime:    endTime,
		RfcPrecookTimeZone:   time.UTC,
		RfcPrecookTimeFormat: "15:04",
	}

	result := canPrecookRfcResponses(nil)

	// With reverse window where current > startTime, should be true
	nowStr := now.Format("15:04")
	// startTime > endTime is reverse window
	// current > startTime means inside the window
	assert.True(t, result, "With reverse window (start=%s > end=%s), current time %s should be inside", startTime, endTime, nowStr)
}

func TestCanPrecookRfcResponses_EnabledWithReverseTimeWindowOutside(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()

	// Create reverse window where current time is outside
	// e.g., window is 01:00 to 03:00, but startTime=01:00 < endTime=03:00 actually
	// Let's make: startTime=23:00, endTime=01:00, currentTime around 12:00 (outside)
	now := time.Now().In(time.UTC)
	currentHour := now.Hour()

	// Create a window that doesn't include current hour
	// If current hour is between 2 and 22, use window 23:00-01:00
	if currentHour >= 2 && currentHour <= 22 {
		startTime := "23:00"
		endTime := "01:00"

		Xc = &XconfConfigs{
			EnableRfcPrecook:     true,
			RfcPrecookStartTime:  startTime,
			RfcPrecookEndTime:    endTime,
			RfcPrecookTimeZone:   time.UTC,
			RfcPrecookTimeFormat: "15:04",
		}

		result := canPrecookRfcResponses(nil)

		assert.False(t, result, "Current time should be outside reverse window %s to %s", startTime, endTime)
	} else {
		// If current time is 23:xx, 00:xx, or 01:xx, use window 10:00-14:00
		startTime := "10:00"
		endTime := "14:00"

		Xc = &XconfConfigs{
			EnableRfcPrecook:     true,
			RfcPrecookStartTime:  startTime,
			RfcPrecookEndTime:    endTime,
			RfcPrecookTimeZone:   time.UTC,
			RfcPrecookTimeFormat: "15:04",
		}

		result := canPrecookRfcResponses(nil)

		// In this case startTime < endTime (normal window), current time is outside
		assert.False(t, result, "Current time should be outside window %s to %s", startTime, endTime)
	}
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
