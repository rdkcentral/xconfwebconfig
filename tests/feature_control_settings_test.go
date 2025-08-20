/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	xwdataapi "github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/dataapi/featurecontrol"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testFile                                     = "../config/sample_xconfwebconfig.conf"
	PARTNER_TAG                                  = "partnerTag"
	MAC_ADDRESS_TAG                              = "macAddressTag"
	ACCOUNT_TAG                                  = "macAddressTag"
	MAC_AND_PARTNER_TAG                          = "macAndPartnerTag"
	ACCOUNT_HASH_TAG                             = "accountHashTag"
	XYZ_PARTNER                                  = "XYZ"
	MAC_ADDRESS                                  = "11:22:33:44:55:66"
	URL_TAGS_MAC_ADDRESS                         = "/getTagsForMacAddress/%s"
	URL_TAGS_PARTNER                             = "/getTagsForPartner/%s"
	URL_TAGS_PARTNER_AND_MAC_ADDRESS             = "/getTagsForPartnerAndMacAddress/partner/%s/macaddress/%s"
	URL_TAGS_MAC_ADDRESS_AND_ACCOUNT             = "/getTagsForMacAddressAndAccount/macaddress/%s/account/%s"
	URL_TAGS_ACCOUNT                             = "/getTagsForAccount/%s"
	URL_TAGS_PARTNER_AND_MAC_ADDRESS_AND_ACCOUNT = "/getTagsForPartnerAndMacAddressAndAccount/partner/%s/macaddress/%s/account/%s"
	URL_TAGS_PARTNER_AND_ACCOUNT                 = "/getTagsForPartnerAndAccount/partner/%s/account/%s"
	URL_DEVICE_SERVICE                           = "/api/v1/operational/mesh-pod/%s/account"
	URL_ACCOUNT_SERVICE_DEVICE_ESTB              = "/devices?hostMac=%s&status=Active"
	URL_ACCOUNT_SERVICE_DEVICE_ECM               = "/devices?ecmMac=%s&status=Active"
)

func TestFeatureSetting(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)

	taggingMockServer := dataapi.SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, XYZ_PARTNER))
	defer taggingMockServer.Close()

	featureIds := []string{}
	features := []rfc.FeatureResponse{}
	for i := 0; i < 5; i++ {
		feature := createAndSaveFeature()
		featureIds = append(featureIds, feature.ID)
		featureResponse := rfc.CreateFeatureResponseObject(*feature)
		features = append(features, featureResponse)
	}

	createAndSaveFeatureRule(featureIds, createRule(CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "model"), re.StandardOperationIs, "X1-1")), "stb")
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, "?model=X1-1", nil, features)
}

func TestFeatureSettingByApplicationType(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)

	taggingMockServer := dataapi.SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, XYZ_PARTNER))
	defer taggingMockServer.Close()

	features := createAndSaveFeatures()
	createAndSaveFeatureRules(features)
	featureResponseStb := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*features["stb"]),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, "?model=X1-1", nil, featureResponseStb)
	featureResponseXhome := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*features["xhome"]),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, "/xhome?model=X1-1", nil, featureResponseXhome)
}

func Test304StatusIfResponseWasNotModified(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)

	taggingMockServer := dataapi.SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, XYZ_PARTNER))
	defer taggingMockServer.Close()

	feature := createAndSaveFeature()
	rule := CreateDefaultEnvModelRule()
	featureRule := createFeatureRule([]string{feature.ID}, rule, "stb")
	setFeatureRule(featureRule)
	featureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	featureControlRuleBase := featurecontrol.NewFeatureControlRuleBase()
	configSetHash := featureControlRuleBase.CalculateHash(featureResponse)
	assertNotMofifiedStatus(t, server, router, configSetHash, nil)
	assertConfigSetHashChange(t, server, router, configSetHash, nil)
}

func TestIfFeatureRuleIsAppliedByRangeOperation(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)

	taggingMockServer := dataapi.SetupTaggingMockServer404Response(t, *server, fmt.Sprintf(URL_TAGS_MAC_ADDRESS, "B4:F2:E8:15:67:46"))
	defer taggingMockServer.Close()

	feature := createAndSaveFeature()
	createAndSaveFeatureRule([]string{feature.ID}, createPercentRangeRule(), "stb")
	featureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	macFits50To100Range := "B4:F2:E8:15:67:46"
	verifyPercentRangeRuleApplying(t, server, router, macFits50To100Range, featureResponse)
}

func TestIfFeatureRuleIsNotAppliedByRangeOperation(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)

	taggingMockServer := dataapi.SetupTaggingMockServer404Response(t, *server, fmt.Sprintf(URL_TAGS_MAC_ADDRESS, "04:02:10:00:00:01"))
	defer taggingMockServer.Close()

	feature := createAndSaveFeature()
	createAndSaveFeatureRule([]string{feature.ID}, createPercentRangeRule(), "stb")
	featureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*feature),
	}
	macDoesntFit50To100Range := "04:02:10:00:00:01"
	verifyPercentRangeRuleApplying(t, server, router, macDoesntFit50To100Range, featureResponse)
}

func TestFeatureInstanceFieldAddedToRFCResponse(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)

	taggingMockServer := dataapi.SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, XYZ_PARTNER))
	defer taggingMockServer.Close()

	feature := createAndSaveFeature()
	rule := CreateRuleKeyValue("model", strings.ToUpper(defaultModelId))
	createAndSaveFeatureRule([]string{feature.ID}, rule, "stb")
	performGetSettingsRequestAndVerifyFeatureControlInstanceName(t, server, router, fmt.Sprintf("?version=%s&applicationType=stb&model=%s", API_VERSION, defaultModelId), feature)
}

func TestFeatureIsNotReturnedForUnknownPartnerTag(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)

	taggingMockServer := dataapi.SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER, XYZ_PARTNER))
	defer taggingMockServer.Close()

	createTagFeatureRule(PARTNER_TAG)
	emptyFeatureResponse := []rfc.FeatureResponse{}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, "?partnerId=unknown", nil, emptyFeatureResponse)
}

func Test200StatusCodeWhenTaggingServiceUnavailableAndEmptyConfigHash(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)

	taggingMockServer := dataapi.SetupTaggingMockServer500Response(t, *server, fmt.Sprintf(URL_TAGS_PARTNER, XYZ_PARTNER))
	defer taggingMockServer.Close()

	emptyFeatureResponse := []rfc.FeatureResponse{}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?partnerId=%s", XYZ_PARTNER), nil, emptyFeatureResponse)
}

func TestGetFeatureSettingByUnknownPartnerId(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)
	// Xc.
	AccountServiceObjectArray := []xwhttp.AccountServiceDevices{
		CreateAccountServicePartnerObject(XYZ_PARTNER),
	}
	expectedResponse, _ := json.Marshal(AccountServiceObjectArray)
	AccountServiceMockServer := dataapi.SetupAccountServiceMockServerOkResponseDynamic(t, *server, expectedResponse, fmt.Sprintf(URL_ACCOUNT_SERVICE_DEVICE_ESTB, MAC_ADDRESS))
	defer AccountServiceMockServer.Close()
	taggingMockServer := dataapi.SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_AND_PARTNER_TAG), fmt.Sprintf(URL_TAGS_PARTNER_AND_MAC_ADDRESS, XYZ_PARTNER, MAC_ADDRESS))
	defer taggingMockServer.Close()
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(xwdataapi.CreatePartnerIdFeature("unknown")),
	}
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?partnerId=unknown&estbMacAddress=%s", MAC_ADDRESS), nil, expectedFeatureResponse)
}

func TestGetAccountHashFeatureIfAccountHashIsPassed(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)
	accountHash := util.CalculateHash(defaultServiceAccountUri)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(*getAccountHashFeature(accountHash)),
	}
	taggingMockServer := dataapi.SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_MAC_ADDRESS, MAC_ADDRESS))
	defer taggingMockServer.Close()
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountHash=%s&estbMacAddress=%s", accountHash, MAC_ADDRESS), nil, expectedFeatureResponse)
}

func TestGetAccountIdFeatureIfAccountIdIsPassed(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(xwdataapi.CreateAccountIdFeature(defaultServiceAccountUri)),
	}
	taggingMockServer := dataapi.SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_MAC_ADDRESS_AND_ACCOUNT, "AA:AA:AA:AA:AA:AA", defaultServiceAccountUri))
	defer taggingMockServer.Close()
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s&estbMacAddress=%s", defaultServiceAccountUri, "AA:AA:AA:AA:AA:AA"), nil, expectedFeatureResponse)
}

func TestGetAccountIdAndHashFeaturesIfSpecificConfigIsEnabled(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testFile)
	accountId := "serviceAccountUri"
	accountHash := util.CalculateHash(accountId)
	accountIdFeature := xwdataapi.CreateAccountIdFeature(accountId)
	accountHashFeature := getAccountHashFeature(accountHash)
	expectedFeatureResponse := []rfc.FeatureResponse{
		rfc.CreateFeatureResponseObject(accountIdFeature),
		rfc.CreateFeatureResponseObject(*accountHashFeature),
	}
	taggingMockServer := dataapi.SetupTaggingMockServerOkResponseDynamic(t, *server, fmt.Sprintf(`["%s"]`, MAC_ADDRESS_TAG), fmt.Sprintf(URL_TAGS_MAC_ADDRESS_AND_ACCOUNT, "AA:AA:AA:AA:AA:AA", accountId))
	defer taggingMockServer.Close()
	performGetSettingsRequestAndVerifyFeatureControl(t, server, router, fmt.Sprintf("?accountId=%s&accountHash=%s&estbMacAddress=%s", accountId, accountHash, "AA:AA:AA:AA:AA:AA"), nil, expectedFeatureResponse)
}

func TestCertExpiryDurationFeatureRuleEvaluation(t *testing.T) {
	_, router := dataapi.GetTestXconfServer(testFile)

	testCases := []struct {
		name                       string
		certExpiryDurationDays     int
		ruleCertExpiryDurationDays int
		durationOffsetMinutes      int
		operation                  string
		responseFeatureNumbers     int
	}{
		{
			name:                       "LTE operation: cert expiry in 89 days, rule condition LTE 90 days from now, rule evaluated",
			certExpiryDurationDays:     89,
			ruleCertExpiryDurationDays: 90,
			durationOffsetMinutes:      -5,
			operation:                  re.StandardOperationLte,
			responseFeatureNumbers:     1,
		}, {
			name:                       "LTE operation: cert expiry in 90 days, rule condition LTE 100 days from now, rule evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 100,
			operation:                  re.StandardOperationLte,
			responseFeatureNumbers:     1,
		}, {
			name:                       "LTE operation: cert expiry in 90 days from now, rule condition LTE 90 days, rule evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 90,
			operation:                  re.StandardOperationLte,
			responseFeatureNumbers:     1,
		}, {
			name:                       "LTE operation: cert expiry in 90 days from now, rule condition LTE 89 days, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 89,
			durationOffsetMinutes:      5,
			operation:                  re.StandardOperationLte,
			responseFeatureNumbers:     0,
		}, {
			name:                       "LTE operation: cert expiry in 90 days from now, rule condition LTE 80 days, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 80,
			operation:                  re.StandardOperationLte,
			responseFeatureNumbers:     0,
		}, {
			name:                       "GTE operation: cert expiry in 91 days from now, rule condition GTE 90, rule evaluated",
			certExpiryDurationDays:     91,
			ruleCertExpiryDurationDays: 90,
			operation:                  re.StandardOperationGte,
			responseFeatureNumbers:     1,
		}, {
			name:                       "GTE operation: cert expiry in 90 days from now, rule condition GTE 90, rule evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 90,
			durationOffsetMinutes:      5,
			operation:                  re.StandardOperationGte,
			responseFeatureNumbers:     1,
		}, {
			name:                       "GTE operation: cert expiry in 100 days from now, rule condition GTE 90, rule evaluated",
			certExpiryDurationDays:     100,
			ruleCertExpiryDurationDays: 90,
			operation:                  re.StandardOperationGte,
			responseFeatureNumbers:     1,
		}, {
			name:                       "GTE operation: cert expiry in 90 days from now, rule condition GTE 91, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 91,
			durationOffsetMinutes:      5,
			operation:                  re.StandardOperationGte,
			responseFeatureNumbers:     0,
		}, {
			name:                       "GTE operation: cert expiry in 90 days from now, rule condition GTE 100, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 100,
			operation:                  re.StandardOperationGte,
			responseFeatureNumbers:     0,
		}, {
			name:                       "GTE operation: cert expiry in 90 days from now, rule condition GTE 100, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 100,
			operation:                  re.StandardOperationGte,
			responseFeatureNumbers:     0,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer DeleteAllEntities()

			modelId := "MODEL" + strconv.Itoa(i)
			rule := createCertExpiryDurationRule(modelId, float64(tc.ruleCertExpiryDurationDays), tc.operation)
			feature := createAndSaveFeature()
			featureRule := createFeatureRule([]string{feature.ID}, rule, shared.STB)
			setFeatureRule(featureRule)
			context, _ := util.GetURLQueryParameterString([][]string{
				{"model", modelId},
			})
			url := fmt.Sprintf("/featureControl/getSettings?%v", context)
			r := httptest.NewRequest("GET", url, nil)
			now := time.Now().UTC().Add(time.Duration(tc.durationOffsetMinutes) * time.Minute)
			clientCertExpiry := now.AddDate(0, 0, tc.certExpiryDurationDays).Format("Jan 02, 2006 @ 15:04:05.000")
			r.Header.Add("Client-Cert-Expiry", clientCertExpiry)
			rr := dataapi.ExecuteRequest(r, router)

			assert.Equal(t, http.StatusOK, rr.Code)

			var featureResp map[string]rfc.FeatureControl
			json.Unmarshal(rr.Body.Bytes(), &featureResp)

			assert.Equal(t, tc.responseFeatureNumbers, len(featureResp["featureControl"].FeatureResponses))
			if len(featureResp["featureControl"].FeatureResponses) > 0 {
				expectedResult := []rfc.FeatureResponse{rfc.CreateFeatureResponseObject(*feature)}
				compareFeatureControlResponses(t, rr.Result(), expectedResult)
			}
		})
	}
}

func TestCertExpiryDurationFeatureRuleEvaluationWhenCertExpiryHeaderIsNotPresent(t *testing.T) {
	testCases := []struct {
		name              string
		certHeaderPresent bool
		header            string
	}{
		{
			"Client-Cert-Expiry header present but empty",
			true,
			"",
		}, {
			"Client-Cert-Expiry header present but does not have correct date format",
			true,
			"Next January 1st",
		}, {
			"Client-Cert-Expiry header not present",
			false,
			"",
		},
	}
	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer DeleteAllEntities()

			modelId := "MODEL" + strconv.Itoa(i)
			rule := createCertExpiryDurationRule(modelId, 100, re.StandardOperationLte)

			feature := createAndSaveFeature()
			featureRule := createFeatureRule([]string{feature.ID}, rule, shared.STB)
			setFeatureRule(featureRule)

			context, _ := util.GetURLQueryParameterString([][]string{
				{common.IP_ADDRESS, "11.11.11.11"},
				{"model", modelId},
			})
			url := fmt.Sprintf("/featureControl/getSettings?%v", context)

			_, router := dataapi.GetTestXconfServer(testConfigFile)

			r := httptest.NewRequest("GET", url, nil)
			if tc.certHeaderPresent {
				r.Header.Add("Client-Cert-Expiry", tc.header)
			}
			rr := dataapi.ExecuteRequest(r, router)
			assert.Equal(t, http.StatusOK, rr.Code)
			var featureResp map[string]rfc.FeatureControl
			json.Unmarshal(rr.Body.Bytes(), &featureResp)

			assert.Equal(t, 0, len(featureResp["featureControl"].FeatureResponses))
		})
	}
}

func TestTimezoneOffsetForGrModel(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	tzGrResp := rfc.CreateFeatureResponseObject(xwdataapi.CreateAccountIdFeature("unknown"))
	tzGrResp["timeZone"] = xwdataapi.ATHENS_EUROPE_TZ
	tzGrResp["tzUTCOffset"] = xwdataapi.GetTimezoneOffset()

	feature := createAndSaveFeature()
	featureResp := rfc.CreateFeatureResponseObject(*feature)

	testCases := []struct {
		name         string
		model        string
		expectedResp []rfc.FeatureResponse
	}{
		{
			"Not GR model",
			"TESTMODEL",
			[]rfc.FeatureResponse{featureResp},
		}, {
			"GR model",
			"GRMODEL",
			[]rfc.FeatureResponse{featureResp, tzGrResp},
		}, {
			"GR model with '-' symbol in model name",
			"GR-MODEL",
			[]rfc.FeatureResponse{featureResp, tzGrResp},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			CreateAndSaveModel(tc.model)
			rule := createRule(CreateCondition(*estbfirmware.RuleFactoryMODEL, re.StandardOperationIs, tc.model))

			featureRule := createFeatureRule([]string{feature.ID}, rule, shared.STB)
			setFeatureRule(featureRule)

			context, _ := util.GetURLQueryParameterString([][]string{
				{common.IP_ADDRESS, "11.11.11.11"},
				{"model", tc.model},
				{common.ACCOUNT_MGMT, "xpc"},
			})
			url := fmt.Sprintf("/featureControl/getSettings?%v", context)

			_, router := dataapi.GetTestXconfServer(testConfigFile)

			r := httptest.NewRequest("GET", url, nil)
			rr := dataapi.ExecuteRequest(r, router)
			assert.Equal(t, http.StatusOK, rr.Code)

			var featureControlResponse map[string]rfc.FeatureControl
			json.Unmarshal(rr.Body.Bytes(), &featureControlResponse)

			assert.Equal(t, len(tc.expectedResp), len(featureControlResponse["featureControl"].FeatureResponses))
			compareFeatureControlResponses(t, rr.Result(), tc.expectedResp)
		})
	}
}

func TestDtgrPartnerIsReturnedForGrModelIfPassedUnknownPartner(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	tzGrResp := rfc.CreateFeatureResponseObject(xwdataapi.CreateAccountIdFeature("unknown"))
	tzGrResp["timeZone"] = xwdataapi.ATHENS_EUROPE_TZ
	tzGrResp["tzUTCOffset"] = xwdataapi.GetTimezoneOffset()

	feature := createAndSaveFeature()

	modelId := "GRTESTMODEL"
	CreateAndSaveModel(modelId)
	rule := createRule(CreateCondition(*estbfirmware.RuleFactoryMODEL, re.StandardOperationIs, modelId))

	featureRule := createFeatureRule([]string{feature.ID}, rule, shared.STB)
	setFeatureRule(featureRule)

	context, _ := util.GetURLQueryParameterString([][]string{
		{common.IP_ADDRESS, "11.11.11.11"},
		{"model", modelId},
		{"partnerId", "unknown"},
		{common.ACCOUNT_MGMT, "xpc"},
	})
	url := fmt.Sprintf("/featureControl/getSettings?%v", context)

	_, router := dataapi.GetTestXconfServer(testConfigFile)

	r := httptest.NewRequest("GET", url, nil)
	rr := dataapi.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	var featureControlResponse map[string]rfc.FeatureControl
	json.Unmarshal(rr.Body.Bytes(), &featureControlResponse)

	assert.Equal(t, 3, len(featureControlResponse["featureControl"].FeatureResponses))

	tzFeature := getTimeZoneFeature(featureControlResponse["featureControl"].FeatureResponses)
	assert.NotEmpty(t, tzFeature)
	assert.Equal(t, strings.ToLower(xwdataapi.DTGR_PARTNER_ID), tzFeature["partnerId"])
	partnerFeature := getPartnerFeature(featureControlResponse["featureControl"].FeatureResponses)
	assert.NotEmpty(t, partnerFeature)
	partnerConfigData := partnerFeature["configData"].(map[string]interface{})
	assert.Equal(t, strings.ToLower(xwdataapi.DTGR_PARTNER_ID), partnerConfigData[common.TR181_DEVICE_TYPE_PARTNER_ID])
}

func getTimeZoneFeature(features []rfc.FeatureResponse) rfc.FeatureResponse {
	for _, featureResp := range features {
		if featureResp["timeZone"] == xwdataapi.ATHENS_EUROPE_TZ {
			return featureResp
		}
	}
	return nil
}

func getPartnerFeature(features []rfc.FeatureResponse) rfc.FeatureResponse {
	for _, featureResp := range features {
		if featureResp["name"] == common.SYNDICATION_PARTNER {
			return featureResp
		}
	}
	return nil
}

func verifyPercentRangeRuleApplying(t *testing.T, server *xwhttp.XconfServer, router *mux.Router, macAddress string, expectedFeatures []rfc.FeatureResponse) {
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings?estbMacAddress=%s", macAddress)
	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	compareFeatureControlResponses(t, res, expectedFeatures)
}

func assertConfigSetHashChange(t *testing.T, server *xwhttp.XconfServer, router *mux.Router, configSetHash string, expectedFeatures []rfc.FeatureResponse) {
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings?model=%s&env=%s", strings.ToUpper(defaultModelId), strings.ToUpper(defaultEnvironmentId))
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("configSetHash", "")
	assert.Nil(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.Equal(t, res.Header["configSetHash"][0], configSetHash)
	compareFeatureControlResponses(t, res, expectedFeatures)
}

func assertNotMofifiedStatus(t *testing.T, server *xwhttp.XconfServer, router *mux.Router, configSetHash string, expectedFeatures []rfc.FeatureResponse) {
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings?model=%s&env=%s", strings.ToUpper(defaultModelId), strings.ToUpper(defaultEnvironmentId))
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("configSetHash", configSetHash)
	assert.Nil(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusNotModified)
	assert.Equal(t, res.Header["configSetHash"][0], configSetHash)
	if expectedFeatures != nil {
		compareFeatureControlResponses(t, res, expectedFeatures)
	}
}

func performGetSettingsRequestAndVerifyFeatureControlInstanceName(t *testing.T, server *xwhttp.XconfServer, router *mux.Router, extraUrl string, expectedFeature *rfc.Feature) {
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings%s", extraUrl)
	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	actualResponse := map[string]rfc.FeatureControl{}
	err = json.Unmarshal(body, &actualResponse)
	assert.Nil(t, err)
	assert.Equal(t, actualResponse["featureControl"].FeatureResponses[0]["featureInstance"], expectedFeature.FeatureName)
	res.Body.Close()
}

func performGetSettingsRequestAndVerifyFeatureControl(t *testing.T, server *xwhttp.XconfServer, router *mux.Router, extraUrl string, headers map[string]string, expectedFeatures []rfc.FeatureResponse) {
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings%s", extraUrl)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("HA-Haproxy-xconf-http", "xconf-https")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	assert.Nil(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	compareFeatureControlResponses(t, res, expectedFeatures)
}

func performGetSettingsRequestAndVerify500ErrorWithNonEmptyConfigSetHash(t *testing.T, server *xwhttp.XconfServer, router *mux.Router, extraUrl string) {
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	url := fmt.Sprintf("/featureControl/getSettings%s", extraUrl)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("configSetHash", "nonEmptyValue")
	assert.Nil(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	body, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusInternalServerError)
	assert.Equal(t, strings.Contains(string(body), "Error Msg"), true)
	res.Body.Close()
}

func compareFeatureControlResponses(t *testing.T, res *http.Response, expectedFeatures []rfc.FeatureResponse) {
	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	actualResponse := map[string]rfc.FeatureControl{}
	err = json.Unmarshal(body, &actualResponse)
	assert.Nil(t, err)
	actualFeatureControl, ok := actualResponse["featureControl"]
	assert.Equal(t, ok, true)
	actualFeatures := actualFeatureControl.FeatureResponses
	assert.Equal(t, actualFeatures != nil, true)
	sortFeatures(actualFeatures)
	sortFeatures(expectedFeatures)
	for i := range expectedFeatures {
		require.Equal(t, len(expectedFeatures[i]), len(actualFeatures[i]))
		for key, value := range expectedFeatures[i] {
			switch v := value.(type) {
			case int:
				assert.Equal(t, value, actualFeatures[i][key].(int))
			case string:
				assert.Equal(t, value, actualFeatures[i][key].(string))
			case bool:
				assert.Equal(t, value, actualFeatures[i][key].(bool))
			case map[string]string:
				for mapK, mapV := range v {
					assert.Equal(t, mapV, actualFeatures[i][key].(map[string]interface{})[mapK].(string))
				}
			// fail if not one of above types so we don't accidentally miss one
			default:
				assert.Equal(t, true, false)
			}
		}
	}
	res.Body.Close()
}

func sortFeatures(features []rfc.FeatureResponse) {
	sort.SliceStable(features, func(i, j int) bool {
		return fmt.Sprintf("%s", features[i]["name"]) < fmt.Sprintf("%s", features[j]["name"])
	})
}

func getAccountHashFeature(accountHash string) *rfc.Feature {
	accountHashFeature := rfc.Feature{
		Name:               "AccountHash",
		FeatureName:        "AccountHash",
		EffectiveImmediate: true,
		Enable:             true,
		ConfigData: map[string]string{
			common.TR181_DEVICE_TYPE_ACCOUNT_HASH: accountHash,
		},
	}
	return &accountHashFeature
}

func createTagFeatureRule(tagNameForRule string) *rfc.Feature {
	feature := createAndSaveFeature()
	createAndSaveFeatureRule([]string{feature.ID}, CreateExistsRule(tagNameForRule), "stb")
	return feature
}

func setFeatureRule(featureRule *rfc.FeatureRule) {
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_FEATURE_CONTROL_RULE, featureRule.Id, featureRule)
}

func setFeature(feature *rfc.Feature) {
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_XCONF_FEATURE, feature.ID, feature)
}

func createAndSaveFeature() *rfc.Feature {
	feature := createFeature()
	setFeature(feature)
	return feature
}

func createFeature() *rfc.Feature {
	id := uuid.New().String()
	configData := map[string]string{}
	configData[fmt.Sprintf("%s-key", id)] = fmt.Sprintf("%s-value", id)

	feature := &rfc.Feature{
		ID:                 id,
		Name:               fmt.Sprintf("%s-name", id),
		EffectiveImmediate: false,
		Enable:             false,
		ConfigData:         configData,
	}
	return feature
}

func createAndSaveFeatureRule(featureIds []string, rule *re.Rule, applicationType string) *rfc.FeatureRule {
	featureRule := createFeatureRule(featureIds, rule, applicationType)
	setFeatureRule(featureRule)
	return featureRule
}

func createFeatureRule(featureIds []string, rule *re.Rule, applicationType string) *rfc.FeatureRule {
	id := uuid.New().String()
	configData := map[string]string{}
	configData[fmt.Sprintf("%s-key", id)] = fmt.Sprintf("%s-value", id)

	featureRule := &rfc.FeatureRule{
		Id:              id,
		Name:            fmt.Sprintf("%s-name", id),
		ApplicationType: applicationType,
		FeatureIds:      featureIds,
		Rule:            rule,
	}
	return featureRule
}

func createRule(condition *re.Condition) *re.Rule {
	rule := &re.Rule{
		Condition: condition,
	}
	return rule
}

func createPercentRangeRule() *re.Rule {
	condition := CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "estbMacAddress"), re.StandardOperationRange, "50-100")
	return createRule(condition)
}

func createAndSaveFeatureRules(features map[string]*rfc.Feature) map[string]*rfc.FeatureRule {
	stbFeatureIdList := []string{features["stb"].ID}
	stbFeatureRule := createFeatureRule(stbFeatureIdList, createRule(CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "model"), re.StandardOperationIs, "X1-1")), "stb")
	setFeatureRule(stbFeatureRule)
	xhomeFeatureIdList := []string{features["xhome"].ID}
	xhomeFeatureRule := createFeatureRule(xhomeFeatureIdList, createRule(CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, "model"), re.StandardOperationIs, "X1-1")), "xhome")
	setFeatureRule(xhomeFeatureRule)
	featureRules := map[string]*rfc.FeatureRule{
		"stb":   stbFeatureRule,
		"xhome": xhomeFeatureRule,
	}
	return featureRules
}

func createAndSaveFeatures() map[string]*rfc.Feature {
	stbFeature := createFeature()
	stbFeature.ApplicationType = "stb"
	setFeature(stbFeature)

	xhomeFeature := createFeature()
	xhomeFeature.ApplicationType = "xhome"
	setFeature(xhomeFeature)

	features := map[string]*rfc.Feature{
		"stb":   stbFeature,
		"xhome": xhomeFeature,
	}
	return features
}
