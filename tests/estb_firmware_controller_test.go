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
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gotest.tools/assert"
)

const (
	X1_SIGN_REDIRECT = "/cgi-bin/x1-sign-redirect.pl"
)

type SecurityTokenTest struct {
	name                string
	mac                 string
	ip                  string
	tokenEnabled        bool
	groupServiceEnabled bool
}

func TestFirmwareConfigParametersAreReturned(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testConfigFile)

	parameters := map[string]string{}
	configKey := "bindingUrl"
	configValue := "http://test.url.com"
	parameters[configKey] = configValue
	firmwareConfig := CreateFirmwareConfig(defaultFirmwareVersion, defaultModelId, "http", "stb")
	firmwareConfig.Properties = parameters
	err := SetFirmwareConfig(firmwareConfig)
	assert.NilError(t, err)

	applicableAction := corefw.NewTemplateApplicableActionAndType(corefw.RuleActionClass, corefw.RULE_TEMPLATE, "")
	rt := CreateAndSaveFirmwareRuleTemplate("ENV_MODEL_RULE", CreateDefaultEnvModelRule(), applicableAction)
	assert.Assert(t, rt != nil)
	bean, err := createAndSaveUseAccountPercentageBean(firmwareConfig)
	assert.Assert(t, bean != nil)
	assert.NilError(t, err)

	context := CreateContext(defaultFirmwareVersion, defaultModelId, defaultEnvironmentId, defaultIpAddress, defaultMacAddress)
	expectedResponse := map[string]interface{}{
		"bindingUrl":               "http://test.url.com",
		"firmwareDownloadProtocol": "http",
		"firmwareFilename":         "FirmwareFilename",
		"firmwareVersion":          defaultFirmwareVersion,
		"rebootImmediately":        false,
	}

	taggingMockServer := dataapi.SetupTaggingMockServerOkResponse(t, *server, fmt.Sprintf(URL_TAGS_MAC_ADDRESS, defaultMacAddress))
	defer taggingMockServer.Close()

	performPostSwuRequestAndValidateBody(t, server, router, map[string]string{}, context, expectedResponse)
}

func TestFirmwareConfigParametersCanNotBeOverriddenByDefinePropertiesRule(t *testing.T) {
	DeleteAllEntities()
	server, router := dataapi.GetTestXconfServer(testConfigFile)

	parameters := map[string]string{}
	configKey := "bindingUrl"
	configValue := "http://test.url.com"
	parameters[configKey] = configValue

	definePropertiesModelId := "DEFINE_PROPERTIES_MODEL_ID"

	firmwareConfig := CreateFirmwareConfig(defaultFirmwareVersion, definePropertiesModelId, "http", "stb")
	firmwareConfig.Properties = parameters
	err := SetFirmwareConfig(firmwareConfig)
	assert.NilError(t, err)

	applicableAction := corefw.NewTemplateApplicableActionAndType(corefw.RuleActionClass, corefw.RULE_TEMPLATE, "")
	rt := CreateAndSaveFirmwareRuleTemplate("ENV_MODEL_RULE", CreateDefaultEnvModelRule(), applicableAction)
	assert.Assert(t, rt != nil)

	percentageBean := CreatePercentageBean("test percentage bean", defaultEnvironmentId, definePropertiesModelId, "", "", defaultFirmwareVersion, "stb")
	percentageBean.LastKnownGood = firmwareConfig.ID
	percentageBean.FirmwareVersions = append(percentageBean.FirmwareVersions, firmwareConfig.FirmwareVersion)
	err = SavePercentageBean(percentageBean)
	assert.NilError(t, err)

	defineProperties := map[string]string{}
	defineProperties[configKey] = "CHANGED VALUE BY DEFINE PROPERTY RULE"
	defineProperties["definePropertyKey"] = "definePropertyValue"
	modelRule := CreateRule("", *estbfirmware.RuleFactoryMODEL, re.StandardOperationIs, definePropertiesModelId)

	definePropertiesApplicableAction := corefw.NewApplicableActionAndType(corefw.DefinePropertiesActionClass, corefw.DEFINE_PROPERTIES_TEMPLATE, "")
	definePropertiesApplicableAction.Properties = defineProperties

	definePropertiesTemplateAction := corefw.NewTemplateApplicableActionAndType(corefw.DefinePropertiesTemplateActionClass, corefw.DEFINE_PROPERTIES_TEMPLATE, "")
	definePropertiesTemplateAction.Properties = buildDefinePropertyTemplateAction(defineProperties, false)

	definePropertiesTemplate := CreateAndSaveFirmwareRuleTemplate("OVERRIDE_FIRMWARE_CONFIG_PARAMETERS", modelRule, definePropertiesTemplateAction)

	fr := CreateAndSaveFirmwareRule(uuid.New().String(), definePropertiesTemplate.ID, "stb", definePropertiesApplicableAction, &definePropertiesTemplate.Rule)
	assert.Assert(t, fr != nil)

	context := CreateContext(defaultFirmwareVersion, definePropertiesModelId, defaultEnvironmentId, defaultIpAddress, defaultMacAddress)
	expectedResponse := map[string]interface{}{
		"bindingUrl":               "http://test.url.com",
		"firmwareDownloadProtocol": "http",
		"firmwareFilename":         "FirmwareFilename",
		"firmwareVersion":          defaultFirmwareVersion,
		"rebootImmediately":        false,
		"definePropertyKey":        "definePropertyValue",
	}

	taggingMockServer := dataapi.SetupTaggingMockServerOkResponse(t, *server, fmt.Sprintf(URL_TAGS_MAC_ADDRESS, defaultMacAddress))
	defer taggingMockServer.Close()

	performPostSwuRequestAndValidateBody(t, server, router, map[string]string{}, context, expectedResponse)
}

func TestCertExpiryDurationFirmwareRuleEvaluation(t *testing.T) {
	_, router := dataapi.GetTestXconfServer(testConfigFile)
	testCases := []struct {
		name                       string
		certExpiryDurationDays     int
		ruleCertExpiryDurationDays int
		durationOffsetMinutes      int
		operation                  string
		expectedStatusCode         int
	}{
		{
			name:                       "LTE operation: cert expiry in 89 days, rule condition LTE 90 days from now, rule evaluated",
			certExpiryDurationDays:     89,
			ruleCertExpiryDurationDays: 90,
			durationOffsetMinutes:      -5,
			operation:                  re.StandardOperationLte,
			expectedStatusCode:         http.StatusOK,
		}, {
			name:                       "LTE operation: cert expiry in 90 days, rule condition LTE 100 days from now, rule evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 100,
			operation:                  re.StandardOperationLte,
			expectedStatusCode:         http.StatusOK,
		}, {
			name:                       "LTE operation: cert expiry in 90 days from now, rule condition LTE 90 days, rule evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 90,
			operation:                  re.StandardOperationLte,
			expectedStatusCode:         http.StatusOK,
		}, {
			name:                       "LTE operation: cert expiry in 90 days from now, rule condition LTE 89 days, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 89,
			durationOffsetMinutes:      5,
			operation:                  re.StandardOperationLte,
			expectedStatusCode:         http.StatusNotFound,
		}, {
			name:                       "LTE operation: cert expiry in 90 days from now, rule condition LTE 80 days, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 80,
			operation:                  re.StandardOperationLte,
			expectedStatusCode:         http.StatusNotFound,
		}, {
			name:                       "GTE operation: cert expiry in 91 days from now, rule condition GTE 90, rule evaluated",
			certExpiryDurationDays:     91,
			ruleCertExpiryDurationDays: 90,
			operation:                  re.StandardOperationGte,
			expectedStatusCode:         http.StatusOK,
		}, {
			name:                       "GTE operation: cert expiry in 90 days from now, rule condition GTE 90, rule evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 90,
			durationOffsetMinutes:      5,
			operation:                  re.StandardOperationGte,
			expectedStatusCode:         http.StatusOK,
		}, {
			name:                       "GTE operation: cert expiry in 100 days from now, rule condition GTE 90, rule evaluated",
			certExpiryDurationDays:     100,
			ruleCertExpiryDurationDays: 90,
			operation:                  re.StandardOperationGte,
			expectedStatusCode:         http.StatusOK,
		}, {
			name:                       "GTE operation: cert expiry in 90 days from now, rule condition GTE 91, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 91,
			durationOffsetMinutes:      5,
			operation:                  re.StandardOperationGte,
			expectedStatusCode:         http.StatusNotFound,
		}, {
			name:                       "GTE operation: cert expiry in 90 days from now, rule condition GTE 100, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 100,
			operation:                  re.StandardOperationGte,
			expectedStatusCode:         http.StatusNotFound,
		}, {
			name:                       "GTE operation: cert expiry in 90 days from now, rule condition GTE 100, rule not evaluated",
			certExpiryDurationDays:     90,
			ruleCertExpiryDurationDays: 100,
			operation:                  re.StandardOperationGte,
			expectedStatusCode:         http.StatusNotFound,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer DeleteAllEntities()

			modelId := "MODEL" + strconv.Itoa(i)
			rule := createCertExpiryDurationRule(modelId, float64(tc.ruleCertExpiryDurationDays), tc.operation)
			_, firmwareConfig := preCreateCertExpiryFirmwareRule(modelId, rule)
			context, _ := util.GetURLQueryParameterString([][]string{
				{common.IP_ADDRESS, "11.11.11.11"},
				{"model", modelId},
				{"firmwareVersion", firmwareConfig.FirmwareVersion},
				{"eStbMac", "AA:AA:AA:AA:AA:AA"},
			})
			url := fmt.Sprintf("/xconf/swu/stb?%v", context)
			r := httptest.NewRequest("GET", url, nil)
			now := time.Now().UTC().Add(time.Duration(tc.durationOffsetMinutes) * time.Minute)
			clientCertExpiry := now.AddDate(0, 0, tc.certExpiryDurationDays).Format("Jan 02, 2006 @ 15:04:05.000")
			r.Header.Add("Client-Cert-Expiry", clientCertExpiry)
			rr := dataapi.ExecuteRequest(r, router)
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			if tc.expectedStatusCode == http.StatusOK {
				var configResp estbfirmware.FirmwareConfigFacadeResponse
				json.Unmarshal(rr.Body.Bytes(), &configResp)

				assert.Equal(t, firmwareConfig.FirmwareVersion, configResp["firmwareVersion"])
				assert.Equal(t, firmwareConfig.FirmwareDownloadProtocol, configResp["firmwareDownloadProtocol"])
				assert.Equal(t, firmwareConfig.FirmwareFilename, configResp["firmwareFilename"])
			}
		})
	}
}

func TestCertExpiryDurationFirmwareRuleEvaluationWhenCertExpiryHeaderIsNotPresent(t *testing.T) {
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
			_, firmwareConfig := preCreateCertExpiryFirmwareRule(modelId, rule)
			context, _ := util.GetURLQueryParameterString([][]string{
				{common.IP_ADDRESS, "11.11.11.11"},
				{"model", modelId},
				{"firmwareVersion", firmwareConfig.FirmwareVersion},
				{"eStbMac", "AA:AA:AA:AA:AA:AA"},
			})
			url := fmt.Sprintf("/xconf/swu/stb?%v", context)

			_, router := dataapi.GetTestXconfServer(testConfigFile)

			r := httptest.NewRequest("GET", url, nil)
			if tc.certHeaderPresent {
				r.Header.Add("Client-Cert-Expiry", tc.header)
			}
			rr := dataapi.ExecuteRequest(r, router)

			assert.Equal(t, http.StatusNotFound, rr.Code)
		})
	}
}

func createCertExpiryDurationRule(modelId string, certExpiryDurationDays float64, operation string) *re.Rule {
	rule := re.NewEmptyRule()
	rule.AddCompoundPart(re.Rule{Condition: re.NewCondition(estbfirmware.RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(modelId))})
	rule.AddCompoundPart(re.And(re.Rule{Condition: re.NewCondition(estbfirmware.RuleFactoryCERT_EXPIRY_DURATION, operation, re.NewFixedArg(certExpiryDurationDays))}))
	return rule
}

func createFirmwareRule(firmwareLocationTemplate, ssrPath string) string {
	modelId := strings.ToUpper(fmt.Sprintf("modelId%d", rand.Int()))
	ssrFirmwareLocation := fmt.Sprintf(firmwareLocationTemplate, ssrPath)
	preCreateFirmwareRule(modelId, ssrFirmwareLocation)
	return modelId
}

func createRequestUrl(mac, modelId string) string {
	context, _ := util.GetURLQueryParameterString([][]string{
		{common.IP_ADDRESS, "11.11.11.11"},
		{"model", modelId},
		{"firmwareVersion", defaultFirmwareVersion},
		{"eStbMac", mac},
	})
	return fmt.Sprintf("/xconf/swu/stb?%v", context)
}

func createRequestUrlWithCapabilities(mac, modelId string) string {
	context, _ := util.GetURLQueryParameterString([][]string{
		{common.IP_ADDRESS, "11.11.11.11"},
		{"model", modelId},
		{"firmwareVersion", defaultFirmwareVersion},
		{"eStbMac", mac},
		{"capabilities", "supportsFullHttpUrl"},
	})
	return fmt.Sprintf("/xconf/swu/stb?%v", context)
}

func executeRequest(localRouter http.Handler, reqUrl string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", reqUrl, nil)
	return dataapi.ExecuteRequest(r, localRouter)
}

func verifyFirmwareLocation(t *testing.T, rr *httptest.ResponseRecorder, testParameters SecurityTokenTest, securityKey string) {
	var firmwareConfigResp estbfirmware.FirmwareConfigFacadeResponse
	json.Unmarshal(rr.Body.Bytes(), &firmwareConfigResp)

	firmwareLocation := firmwareConfigResp["firmwareLocation"].(string)
	assert.Assert(t, len(firmwareLocation) > 0)

	firmwareLocationUrl, _ := url.Parse(firmwareLocation)

	// since the path starts with a `/`, the first entry in the list will be empty string
	pathList := strings.Split(firmwareLocationUrl.Path, "/")
	xdsKey := pathList[1]
	respToken := pathList[2]
	// if there's no token, the path will just be the original url path
	if !testParameters.tokenEnabled {
		assert.Equal(t, X1_SIGN_REDIRECT, firmwareLocationUrl.Path)
		assert.Assert(t, xdsKey != "xds")
	} else if !testParameters.groupServiceEnabled {
		signingKey := hmac.New(sha1.New, []byte(securityKey))
		signingKey.Write([]byte(testParameters.ip))
		securityToken := customBase64Encoding.EncodeToString(signingKey.Sum(nil))
		assert.Equal(t, securityToken, respToken)
		assert.Equal(t, xdsKey, "xds")
	} else {
		mac := strings.ToUpper(strings.ReplaceAll(testParameters.mac, ":", ""))
		assert.Equal(t, mac, respToken)
		assert.Equal(t, xdsKey, "xds")
	}
}

func preCreateDownlowadLocationRoundRobinFilter(ssrHttpTempl string) *estbfirmware.DownloadLocationRoundRobinFilterValue {
	rrFilter := estbfirmware.NewEmptyDownloadLocationRoundRobinFilterValue()
	rrFilter.ID = estbfirmware.ROUND_ROBIN_FILTER_SINGLETON_ID
	rrFilter.HttpLocation = fmt.Sprintf(ssrHttpTempl, "")
	rrFilter.Locations = []estbfirmware.Location{{"10.10.10.10", 100.0}}
	rrFilter.Ipv6locations = []estbfirmware.Location{{"06a0:ac48:eaa2:bdf4:6943:f310:39c3:ec1b", 100.0}}
	rrFilter.HttpFullUrlLocation = fmt.Sprintf(ssrHttpTempl, X1_SIGN_REDIRECT)

	ds.GetCachedSimpleDao().SetOne(ds.TABLE_SINGLETON_FILTER_VALUE, rrFilter.ID, rrFilter)
	return rrFilter
}

func preCreateFirmwareRule(modelId, firmwareLocation string) {
	model := CreateAndSaveModel(modelId)

	firmwareConfig := CreateFirmwareConfig(defaultFirmwareVersion+modelId, modelId, "http", "stb")
	firmwareConfig.FirmwareLocation = firmwareLocation
	SetFirmwareConfig(firmwareConfig)

	templateAction := corefw.NewTemplateApplicableActionAndType(corefw.RuleActionClass, corefw.RULE_TEMPLATE, "")
	rule := re.Rule{
		Condition: re.NewCondition(estbfirmware.RuleFactoryMODEL, re.StandardOperationIs, re.NewFixedArg(model.ID)),
	}
	template := CreateAndSaveFirmwareRuleTemplate("MODEL_RULE", &rule, templateAction)

	ruleAction := CreateRuleAction(corefw.RuleActionClass, corefw.RULE, firmwareConfig.ID)
	CreateAndSaveFirmwareRule(uuid.NewString(), template.ID, shared.STB, ruleAction, &rule)
}

func preCreateCertExpiryFirmwareRule(modelId string, rule *re.Rule) (*corefw.FirmwareRule, *estbfirmware.FirmwareConfig) {
	CreateAndSaveModel(modelId)

	firmwareConfig := CreateFirmwareConfig(defaultFirmwareVersion+modelId, modelId, "http", "stb")
	SetFirmwareConfig(firmwareConfig)

	templateAction := corefw.NewTemplateApplicableActionAndType(corefw.RuleActionClass, corefw.RULE_TEMPLATE, "")
	template := CreateAndSaveFirmwareRuleTemplate("MODEL_RULE", rule, templateAction)

	ruleAction := CreateRuleAction(corefw.RuleActionClass, corefw.RULE, firmwareConfig.ID)
	firmwareRule := CreateAndSaveFirmwareRule(uuid.NewString(), template.ID, shared.STB, ruleAction, rule)
	return firmwareRule, firmwareConfig
}

func setUpXconfServerWithFirmwareSecurityConfig(localServer *xwhttp.XconfServer, ssrPath, securityToken string, tokenEnabled bool, groupServiceEnabled bool) *mux.Router {
	localServer.SecurityTokenConfig = createSecurityTokenConfig(securityToken, groupServiceEnabled)
	localServer.FirmwareSecurityTokenConfig = createSecurityPathConfig(ssrPath, tokenEnabled)
	localRouter := localServer.GetRouter(true)
	dataapi.XconfSetup(localServer, localRouter)
	// xwhttp.InitSatTokenManager(server.XconfServer, true)
	return localRouter
}

func createAndSaveUseAccountPercentageBean(lkgConfig *estbfirmware.FirmwareConfig) (*estbfirmware.PercentageBean, error) {
	useAccountBean := CreatePercentageBean("useAccountName", defaultEnvironmentId, defaultModelId, "", "", defaultFirmwareVersion, "stb")
	useAccountBean.UseAccountIdPercentage = true
	useAccountBean.LastKnownGood = lkgConfig.ID
	firmwareVersions := useAccountBean.FirmwareVersions
	firmwareVersions = append(firmwareVersions, lkgConfig.FirmwareVersion)
	useAccountBean.FirmwareVersions = firmwareVersions
	err := SavePercentageBean(useAccountBean)
	return useAccountBean, err
}

func buildDefinePropertyTemplateAction(parameters map[string]string, requiredAll bool) map[string]corefw.PropertyValue {
	propertyValues := map[string]corefw.PropertyValue{}
	for k, v := range parameters {
		propertyValue := corefw.PropertyValue{
			Value:           v,
			Optional:        requiredAll,
			ValidationTypes: []corefw.ValidationType{"STRING"},
		}
		propertyValues[k] = propertyValue
	}
	return propertyValues
}

func SavePercentageBean(percentageBean *estbfirmware.PercentageBean) error {
	firmwareRule := estbfirmware.ConvertPercentageBeanToFirmwareRule(*percentageBean)
	return corefw.CreateFirmwareRuleOneDB(firmwareRule)
}

func performPostSwuRequestAndValidateBody(t *testing.T, server *xwhttp.XconfServer, router *mux.Router, headers map[string]string, context *estbfirmware.ConvertedContext, expectedResponse estbfirmware.FirmwareConfigFacadeResponse) {
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	url := postContext("/xconf/swu/stb", context)
	req, err := http.NewRequest("POST", url, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	assert.NilError(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	body, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	actualResponse := map[string]interface{}{}
	err = json.Unmarshal(body, &actualResponse)
	assert.NilError(t, err)
	for k, v := range expectedResponse {
		switch v.(type) {
		case string:
			assert.Equal(t, v, actualResponse[k].(string))
		case bool:
			assert.Equal(t, v, actualResponse[k].(bool))
		// fail if not one of above types so we don't accidentally miss one
		default:
			assert.Equal(t, true, false)
		}
	}
}

func postContext(url string, context *estbfirmware.ConvertedContext) string {
	contextMap := context.Context
	if len(contextMap) == 0 {
		return url
	}
	var sb strings.Builder
	for k, v := range contextMap {
		sb.Write([]byte(fmt.Sprintf("%s=%s&", k, v)))
	}
	queryParamString := sb.String()
	return fmt.Sprintf("%s?%s", url, queryParamString[0:len(queryParamString)-1])
}
