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
	"strings"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/dataapi"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

type ResponseProfile struct {
	Name        string    `json:"name"`
	VersionHash string    `json:"versionHash"`
	Value       util.Dict `json:"value"`
}

type TelemetryTwoResponse struct {
	Profiles []ResponseProfile `json:"profiles"`
}

func TestTelemetryTwoHandlerSampleData(t *testing.T) {
	// setup env
	log.SetLevel(log.WarnLevel)

	router := server.GetRouter(true)
	dataapi.XconfSetup(server, router)
	stm := xwhttp.GetSatTokenManager()
	stm.SetTestOnly(true)
	// Walk(router)

	// set up codebig mock server for ok response
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	// ==== setup build sample data ====
	// build sample t2rules
	t2Rules := []logupload.TelemetryTwoRule{}
	err := json.Unmarshal([]byte(SampleTelemetryTwoRulesString), &t2Rules)
	assert.NilError(t, err)
	for _, v := range t2Rules {
		t2Rule := v
		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, t2Rule.ID, &t2Rule)
		assert.NilError(t, err)
		itf, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_RULES, t2Rule.ID)
		assert.NilError(t, err)
		fetchedT2Rule, ok := itf.(*logupload.TelemetryTwoRule)
		assert.Assert(t, ok)
		assert.Assert(t, t2Rule.Equals(fetchedT2Rule))
	}

	// build sample t2profiles
	for profileUuid, profileName := range SampleProfileIdNameMap {
		// write a t2profile
		sp1 := fmt.Sprintf(MockTelemetryTwoProfileTemplate1, profileUuid, profileName)
		var srcT2Profile logupload.TelemetryTwoProfile
		err = json.Unmarshal([]byte(sp1), &srcT2Profile)
		assert.NilError(t, err)
		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid, &srcT2Profile)
		assert.NilError(t, err)
		// get a t2profile
		itf, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid)
		assert.NilError(t, err)
		tgtT2Profile, ok := itf.(*logupload.TelemetryTwoProfile)
		assert.Assert(t, ok)
		assert.DeepEqual(t, &srcT2Profile, tgtT2Profile)
	}

	// ==== insert an extra rule/profile for ill-formatted profile testing ====
	err = json.Unmarshal([]byte(ExtraRuleString1), &t2Rules)
	assert.NilError(t, err)
	for _, v := range t2Rules {
		t2Rule := v
		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, t2Rule.ID, &t2Rule)
		assert.NilError(t, err)
		itf, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_RULES, t2Rule.ID)
		assert.NilError(t, err)
		fetchedT2Rule, ok := itf.(*logupload.TelemetryTwoRule)
		assert.Assert(t, ok)
		assert.Assert(t, t2Rule.Equals(fetchedT2Rule))
	}
	// write a t2profile
	sp1 := fmt.Sprintf(MockTelemetryTwoProfileTemplate2, IllFormattedProfileUuid, IllFormattedProfileName)
	var srcT2Profile logupload.TelemetryTwoProfile
	err = json.Unmarshal([]byte(sp1), &srcT2Profile)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, IllFormattedProfileUuid, &srcT2Profile)
	assert.NilError(t, err)
	// get a t2profile
	itf, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, IllFormattedProfileUuid)
	assert.NilError(t, err)
	tgtT2Profile, ok := itf.(*logupload.TelemetryTwoProfile)
	assert.Assert(t, ok)
	assert.DeepEqual(t, &srcT2Profile, tgtT2Profile)

	// ==== case 1 build the query params ====
	params := [][]string{
		{"env", "PROD"},
		{"version", "2.0"},
		{"model", "CGM4140COM"},
		{"partnerId", "comcast"},
		{"accountId", "1234567890"},
		{"firmwareVersion", "CGM4140COM_4.4p1s11_PROD_sey"},
		{"estbMacAddress", "112233445565"},
		{"ecmMacAddress", "112233445567"},
	}
	queryParamString, err := util.GetURLQueryParameterString(params)
	url := fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	rbytes, err := io.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	var telemetryTwoResponse TelemetryTwoResponse
	err = json.Unmarshal(rbytes, &telemetryTwoResponse)
	assert.NilError(t, err)
	assert.Assert(t, len(telemetryTwoResponse.Profiles) == 1)
	firstProfile := telemetryTwoResponse.Profiles[0]
	expectedName := "xpc_test_profile_001"
	assert.Equal(t, firstProfile.Name, expectedName)

	// ==== case 2 build the query params ====
	params = [][]string{
		{"comp", "test"},
	}
	queryParamString, err = util.GetURLQueryParameterString(params)
	url = fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = dataapi.ExecuteRequest(req, router).Result()
	rbytes, err = io.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	telemetryTwoResponse = TelemetryTwoResponse{}
	err = json.Unmarshal(rbytes, &telemetryTwoResponse)
	assert.NilError(t, err)
	assert.Assert(t, len(telemetryTwoResponse.Profiles) == 1)
	firstProfile = telemetryTwoResponse.Profiles[0]
	expectedName = "wsmithT2.0ProfileTest"
	assert.Equal(t, firstProfile.Name, expectedName)

	// ==== case 3 build the query params ====
	params = [][]string{
		{"env", "AA"},
		{"comp", "test"},
	}
	queryParamString, err = util.GetURLQueryParameterString(params)
	url = fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = dataapi.ExecuteRequest(req, router).Result()
	rbytes, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	// ==== case 4 build the query params ====
	params = [][]string{
		{"estbMacAddress", "84:E0:58:57:53:F0"},
	}
	queryParamString, err = util.GetURLQueryParameterString(params)
	url = fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = dataapi.ExecuteRequest(req, router).Result()
	rbytes, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	telemetryTwoResponse = TelemetryTwoResponse{}
	err = json.Unmarshal(rbytes, &telemetryTwoResponse)
	assert.NilError(t, err)

	// ==== case 5 xpc-15452 the matched T2 profile is ill-formatted json ====
	params = [][]string{
		{"estbMacAddress", "33:33:00:00:11:11"},
	}
	queryParamString, err = util.GetURLQueryParameterString(params)
	url = fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = dataapi.ExecuteRequest(req, router).Result()
	rbytes, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusInternalServerError)
	notFoundText := "404 NOT FOUND"
	assert.Assert(t, !strings.Contains(string(rbytes), notFoundText))
}

func TestTelemetryTwoHandlerMac(t *testing.T) {
	// setup env
	log.SetLevel(log.WarnLevel)

	router := server.GetRouter(true)
	dataapi.XconfSetup(server, router)
	stm := xwhttp.GetSatTokenManager()
	stm.SetTestOnly(true)
	// Walk(router)

	// ==== setup mock data ====
	namedlistKey := fmt.Sprintf("red%v", uuid.New().String()[:4])
	ruleUuid := uuid.New().String()
	profileName := fmt.Sprintf("orange%v", uuid.New().String()[:4])
	profileUuid := uuid.New().String()

	// ---- part 1 namedlist ----
	macList1 := []string{
		"11:11:22:22:33:02",
		"11:11:22:22:33:03",
		"11:11:22:22:33:05",
		"11:11:22:22:33:07",
	}
	srcGnl := shared.NewGenericNamespacedList(namedlistKey, shared.MacList, macList1)
	err := ds.GetCachedSimpleDao().SetOne(shared.TableGenericNSList, srcGnl.ID, srcGnl)
	assert.NilError(t, err)
	itf, err := ds.GetCachedSimpleDao().GetOne(shared.TableGenericNSList, srcGnl.ID)
	assert.NilError(t, err)
	readGnl, ok := itf.(*shared.GenericNamespacedList)
	assert.Assert(t, ok)
	assert.DeepEqual(t, readGnl.Data, macList1)

	// --- part 2 telemetry profile ----
	// write a t2rule
	sr2 := fmt.Sprintf(MockTelemetryTwoRuleTemplate2, namedlistKey, ruleUuid, profileName, profileUuid)
	var srcT2Rule logupload.TelemetryTwoRule
	err = json.Unmarshal([]byte(sr2), &srcT2Rule)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, srcT2Rule.ID, &srcT2Rule)
	assert.NilError(t, err)
	// get a t2rule
	itf, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_RULES, ruleUuid)
	tgtT2Rule, ok := itf.(*logupload.TelemetryTwoRule)
	assert.Assert(t, ok)
	assert.Assert(t, srcT2Rule.Equals(tgtT2Rule))

	// --- part 3 set telemetry rule ----
	sp1 := fmt.Sprintf(MockTelemetryTwoProfileTemplate1, profileUuid, profileName)
	var srcT2Profile logupload.TelemetryTwoProfile
	err = json.Unmarshal([]byte(sp1), &srcT2Profile)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid, &srcT2Profile)
	assert.NilError(t, err)
	// get a t2profile
	itf, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid)
	tgtT2Profile, ok := itf.(*logupload.TelemetryTwoProfile)
	assert.Assert(t, ok)
	assert.DeepEqual(t, &srcT2Profile, tgtT2Profile)

	// ==== case 1 build the query params ====
	params := [][]string{
		{"estbMacAddress", "111122223307"},
	}
	queryParamString, err := util.GetURLQueryParameterString(params)
	url := fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	rbytes, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	var telemetryTwoResponse TelemetryTwoResponse
	err = json.Unmarshal(rbytes, &telemetryTwoResponse)
	assert.NilError(t, err)
	assert.Assert(t, len(telemetryTwoResponse.Profiles) == 1)
	firstProfile := telemetryTwoResponse.Profiles[0]
	assert.Equal(t, firstProfile.Name, profileName)

	// ==== case 1 build the query params ====
	params = [][]string{
		{"estbMacAddress", "111122223304"},
	}
	queryParamString, err = util.GetURLQueryParameterString(params)
	url = fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = dataapi.ExecuteRequest(req, router).Result()
	rbytes, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	// ==== case 1 build the query params ====
	params = [][]string{
		{"estbMacAddress", "111122223305"},
	}
	queryParamString, err = util.GetURLQueryParameterString(params)
	url = fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = dataapi.ExecuteRequest(req, router).Result()
	rbytes, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	telemetryTwoResponse = TelemetryTwoResponse{}
	err = json.Unmarshal(rbytes, &telemetryTwoResponse)
	assert.NilError(t, err)
	assert.Assert(t, len(telemetryTwoResponse.Profiles) == 1)
	firstProfile = telemetryTwoResponse.Profiles[0]
	assert.Equal(t, firstProfile.Name, profileName)
}

func TestTelemetryTwoHandlerIpRange(t *testing.T) {
	// setup env
	log.SetLevel(log.WarnLevel)

	router := server.GetRouter(true)
	dataapi.XconfSetup(server, router)
	stm := xwhttp.GetSatTokenManager()
	stm.SetTestOnly(true)
	// Walk(router)

	// set up codebig mock server for ok response
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	// ==== setup mock data ====
	namedlistKey := fmt.Sprintf("red%v", uuid.New().String()[:4])
	ruleUuid := uuid.New().String()
	profileName := fmt.Sprintf("orange%v", uuid.New().String()[:4])
	profileUuid := uuid.New().String()

	// ---- part 1 namedlist ----
	ipList1 := []string{
		"1.2.3.4",
		"20.30.40.50/24",
		"33.44.55.66/20",
	}
	srcGnl := shared.NewGenericNamespacedList(namedlistKey, shared.IpList, ipList1)
	err := ds.GetCachedSimpleDao().SetOne(shared.TableGenericNSList, srcGnl.ID, srcGnl)
	assert.NilError(t, err)
	itf, err := ds.GetCachedSimpleDao().GetOne(shared.TableGenericNSList, srcGnl.ID)
	assert.NilError(t, err)
	readGnl, ok := itf.(*shared.GenericNamespacedList)
	assert.Assert(t, ok)
	assert.DeepEqual(t, readGnl.Data, ipList1)

	// --- part 2 telemetry profile ----
	// write a t2rule
	sr3 := fmt.Sprintf(MockTelemetryTwoRuleTemplate3, namedlistKey, ruleUuid, profileName, profileUuid)
	var srcT2Rule logupload.TelemetryTwoRule
	err = json.Unmarshal([]byte(sr3), &srcT2Rule)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, srcT2Rule.ID, &srcT2Rule)
	assert.NilError(t, err)
	// get a t2rule
	itf, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_RULES, ruleUuid)
	tgtT2Rule, ok := itf.(*logupload.TelemetryTwoRule)
	assert.Assert(t, ok)
	assert.Assert(t, srcT2Rule.Equals(tgtT2Rule))

	// --- part 3 set telemetry rule ----
	sp1 := fmt.Sprintf(MockTelemetryTwoProfileTemplate1, profileUuid, profileName)
	var srcT2Profile logupload.TelemetryTwoProfile
	err = json.Unmarshal([]byte(sp1), &srcT2Profile)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid, &srcT2Profile)
	assert.NilError(t, err)
	// get a t2profile
	itf, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid)
	tgtT2Profile, ok := itf.(*logupload.TelemetryTwoProfile)
	assert.Assert(t, ok)
	assert.DeepEqual(t, &srcT2Profile, tgtT2Profile)

	// ==== case 1 build the query params ====
	params := [][]string{
		//{"estbMacAddress", "111122223307"},
		{"ipAddress", "1.2.3.4"},
	}
	queryParamString, err := util.GetURLQueryParameterString(params)
	url := fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	rbytes, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	var telemetryTwoResponse TelemetryTwoResponse
	err = json.Unmarshal(rbytes, &telemetryTwoResponse)
	assert.NilError(t, err)
	assert.Assert(t, len(telemetryTwoResponse.Profiles) == 1)
	firstProfile := telemetryTwoResponse.Profiles[0]
	assert.Equal(t, firstProfile.Name, profileName)

	// ==== case 2 build the query params ====
	params = [][]string{
		{"ipAddress", "11.2.3.4"},
	}
	queryParamString, err = util.GetURLQueryParameterString(params)
	url = fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = dataapi.ExecuteRequest(req, router).Result()
	rbytes, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	// ==== case 3 build the query params ====
	params = [][]string{
		{"ipAddress", "20.30.40.100"},
	}
	queryParamString, err = util.GetURLQueryParameterString(params)
	url = fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err = http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res = dataapi.ExecuteRequest(req, router).Result()
	rbytes, err = ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK)
	telemetryTwoResponse = TelemetryTwoResponse{}
	err = json.Unmarshal(rbytes, &telemetryTwoResponse)
	assert.NilError(t, err)
	assert.Assert(t, len(telemetryTwoResponse.Profiles) == 1)
	firstProfile = telemetryTwoResponse.Profiles[0]
	assert.Equal(t, firstProfile.Name, profileName)
}

func TestTelemetryTwoEvaluationNoOpRule(t *testing.T) {
	DeleteAllEntities()
	defer DeleteAllEntities()

	router := server.GetRouter(true)
	dataapi.XconfSetup(server, router)

	profile := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profile.ID, profile)

	model := CreateAndSaveModel("TEST_PROFILE_MODEL")

	telemetryTwoRule := createTelemetryTwoRule([]string{profile.ID}, model.ID, true)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, telemetryTwoRule.ID, telemetryTwoRule)

	reqUrl := "/loguploader/getTelemetryProfiles?model=" + model.ID

	r := httptest.NewRequest("GET", reqUrl, nil)
	rr := dataapi.ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	var telemetryTwoResponse TelemetryTwoResponse
	json.Unmarshal(rr.Body.Bytes(), &telemetryTwoResponse)
	assert.Equal(t, len(telemetryTwoResponse.Profiles), 0)

	reqUrl = "/loguploader/getTelemetryProfiles?model=" + model.ID + "_MISSING"
	r = httptest.NewRequest("GET", reqUrl, nil)
	rr = dataapi.ExecuteRequest(r, router)
	assert.Equal(t, http.StatusNotFound, rr.Code)

	var notFoundResp string
	json.Unmarshal(rr.Body.Bytes(), &notFoundResp)

	assert.Equal(t, "<h2>404 NOT FOUND</h2>profiles not found", notFoundResp)
}

func createTelemetryTwoProfile() *logupload.TelemetryTwoProfile {
	profile := &logupload.TelemetryTwoProfile{}
	profile.ID = uuid.New().String()
	profile.Name = "Profile Name" + profile.ID[len(profile.ID)-4:]
	profile.ApplicationType = shared.STB
	profile.Jsonconfig = TestTelemetryTwoProfileJsonConfig
	return profile
}

func createTelemetryTwoRule(profileIds []string, modelId string, noOp bool) *logupload.TelemetryTwoRule {
	telemetryTwoRule := &logupload.TelemetryTwoRule{}
	telemetryTwoRule.ID = uuid.New().String()
	telemetryTwoRule.Name = "Rule Name" + telemetryTwoRule.ID[len(telemetryTwoRule.ID)-4:]
	telemetryTwoRule.ApplicationType = shared.STB
	telemetryTwoRule.Rule = *createRule(CreateCondition(*estbfirmware.RuleFactoryMODEL, rulesengine.StandardOperationIs, modelId))
	telemetryTwoRule.NoOp = noOp
	if !noOp {
		telemetryTwoRule.BoundTelemetryIDs = profileIds
	}
	return telemetryTwoRule
}
