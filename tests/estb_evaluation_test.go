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
	"fmt"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

var (
	modelPercentFilterTestCase                *shared.Model
	environmentPercentFilterTestCase          *shared.Environment
	macListPercentFilterTestCase              *shared.GenericNamespacedList
	firmwareConfigPercentFilterTestCase       *coreef.FirmwareConfig
	contextConfigPercentFilterTestCase        *coreef.FirmwareConfig
	envModelRuleTemplatePercentFilterTestCase *corefw.FirmwareRuleTemplate
	envModelFirmwareRulePercentFilterTestCase *corefw.FirmwareRule
	initDone                                  = false
)

func TestDownloadLocationRoundRobinFilterSetLocationByConnectionType(t *testing.T) {
	firmwareConfig := GetFirmwareConfig1()
	assert.Assert(t, firmwareConfig.ID != "")
	assert.Assert(t, firmwareConfig.ID == FirmwareConfigId1)

	ff := coreef.NewFirmwareConfigFacade(firmwareConfig)
	assert.Assert(t, ff != nil)
	estbfirmware.DownloadLocationRoundRobinFilterSetLocationByConnectionType(false, ff, "https://www.fool.com")
	fullHttpLocation := ff.GetStringValue(coreef.FIRMWARE_LOCATION)
	assert.Equal(t, fullHttpLocation, "http://www.fool.com")
}

func initPercentConditions(t *testing.T) {
	//todo if not return, other tests would failure
	if initDone {
		return
	}
	modelPercentFilterTestCase = CreateAndSaveModel("ENV_MODEL_RULE_MODEL_ID")

	assert.Assert(t, modelPercentFilterTestCase != nil)

	environmentPercentFilterTestCase = CreateAndSaveEnvironment("ENV_MODEL_RULE_ENVIRONMENT_ID")

	assert.Assert(t, environmentPercentFilterTestCase != nil)

	macListPercentFilterTestCase = CreateAndSaveGenericNamespacedList("envModelRuleMacListId", "MAC_LIST", "AA:BB:CC:AA:BB:CC")

	assert.Assert(t, macListPercentFilterTestCase != nil)

	firmwareConfigPercentFilterTestCase = CreateAndSaveFirmwareConfig("version", modelPercentFilterTestCase.ID, "http", "stb")

	assert.Assert(t, firmwareConfigPercentFilterTestCase != nil)

	contextConfigPercentFilterTestCase = CreateAndSaveFirmwareConfig("contextVersion", modelPercentFilterTestCase.ID, "http", "stb")

	assert.Assert(t, contextConfigPercentFilterTestCase != nil)

	envModelRuleTemplatePercentFilterTestCase = CreateAndSaveFirmwareRuleTemplate("ENV_MODEL_RULE", CreateEnvModelRule(environmentPercentFilterTestCase.ID, modelPercentFilterTestCase.ID, macListPercentFilterTestCase.ID), CreateTemplateRuleAction(corefw.RuleActionClass, corefw.RULE_TEMPLATE, firmwareConfigPercentFilterTestCase.ID))

	assert.Assert(t, envModelRuleTemplatePercentFilterTestCase != nil)

	envModelFirmwareRulePercentFilterTestCase = CreateAndSaveEnvModelFirmwareRule("envModelRuleName", firmwareConfigPercentFilterTestCase.ID, environmentPercentFilterTestCase.ID, modelPercentFilterTestCase.ID, macListPercentFilterTestCase.ID)

	assert.Assert(t, envModelFirmwareRulePercentFilterTestCase != nil)

	// test rule tempte or run in db
	dbTemplaeRule, err := corefw.GetFirmwareRuleTemplateOneDBWithId(envModelRuleTemplatePercentFilterTestCase.ID)
	assert.NilError(t, err)
	assert.Assert(t, dbTemplaeRule != nil)
	assert.Assert(t, dbTemplaeRule.ApplicableAction != nil)
	assert.Equal(t, dbTemplaeRule.ApplicableAction.ActionType, corefw.RULE_TEMPLATE, fmt.Sprintf("Rule Template actioneType %v", dbTemplaeRule.ApplicableAction.ActionType))

	dbPercRule, err1 := corefw.GetFirmwareRuleOneDB(envModelFirmwareRulePercentFilterTestCase.ID)
	assert.NilError(t, err1)
	assert.Assert(t, dbPercRule != nil)
	assert.Assert(t, dbPercRule.ApplicableAction != nil)
	assert.Equal(t, corefw.ApplicableActionTypeToString(dbPercRule.ApplicableAction.ActionType), corefw.ApplicableActionTypeToString(corefw.RULE), fmt.Sprintf("Rule actioneType %v", dbPercRule.ApplicableAction.ActionType))

	initDone = true

	assert.Assert(t, initDone)
}

func TestPercentageIs100AndActive(t *testing.T) {
	//t.Skip("")
	initPercentConditions(t)
	CreateAndSavePercentFilter(envModelFirmwareRulePercentFilterTestCase.Name, 100.0, "", "",
		100.0, []string{firmwareConfigPercentFilterTestCase.FirmwareVersion}, true, true, true, "stb")

	verifyResponse(t)
}

func TestPercentageIs0AndInactive(t *testing.T) {
	//t.Skip("")
	initPercentConditions(t)
	CreateAndSavePercentFilter(envModelFirmwareRulePercentFilterTestCase.Name, 100.0, "", "", 0.0, []string{firmwareConfigPercentFilterTestCase.FirmwareVersion}, false, true, true, "stb")

	verifyResponse(t)
}

func TestPercentageIs100AndInactive(t *testing.T) {
	//t.Skip("")
	initPercentConditions(t)
	CreateAndSavePercentFilter(envModelFirmwareRulePercentFilterTestCase.Name, 100.0, "", "", 100.0, []string{firmwareConfigPercentFilterTestCase.FirmwareVersion}, false, true, true, "stb")

	verifyResponse(t)
}

func verifyResponse(t *testing.T) {
	expectedRuleConfig := createAndNullifyFirmwareConfigFacade(firmwareConfigPercentFilterTestCase)
	assert.Assert(t, expectedRuleConfig != nil)
	performAndVerifyRequest(contextConfigPercentFilterTestCase, expectedRuleConfig, t)
}

func createAndNullifyFirmwareConfigFacade(firmwareConfig *coreef.FirmwareConfig) *coreef.FirmwareConfigFacade {
	firmwareConfig.RebootImmediately = true
	firmwareConfigFacade := coreef.NewFirmwareConfigFacade(firmwareConfig)
	delete(firmwareConfigFacade.Properties, common.ID)
	delete(firmwareConfigFacade.Properties, common.DESCRIPTION)
	delete(firmwareConfigFacade.Properties, common.SUPPORTED_MODEL_IDS)
	delete(firmwareConfigFacade.Properties, common.UPDATED)
	return firmwareConfigFacade
}

func performAndVerifyRequest(firmwareConfigForRequest *coreef.FirmwareConfig, expectedConfig *coreef.FirmwareConfigFacade, t *testing.T) {
	contextMap := make(map[string]string)
	contextMap["applicationType"] = "stb"
	contextMap["eStbMac"] = macListPercentFilterTestCase.Data[0]
	contextMap["env"] = environmentPercentFilterTestCase.ID
	contextMap["model"] = modelPercentFilterTestCase.ID
	contextMap["firmwareVersion"] = firmwareConfigForRequest.FirmwareVersion

	convertedContext := coreef.GetContextConverted(contextMap)
	assert.Assert(t, convertedContext != nil)
	estbFirmwareRuleBase := estbfirmware.NewEstbFirmwareRuleBaseDefault()
	assert.Assert(t, estbFirmwareRuleBase != nil)
	evaluationResult, err := estbFirmwareRuleBase.Eval(contextMap, convertedContext, "stb", log.Fields{})
	assert.Assert(t, err == nil)
	assert.Assert(t, evaluationResult != nil)
	//todo
	assert.Assert(t, evaluationResult.FirmwareConfig != nil)
	assert.Equal(t, evaluationResult.FirmwareConfig.GetRebootImmediately(), expectedConfig.GetRebootImmediately())
	assert.Equal(t, evaluationResult.FirmwareConfig.GetFirmwareVersion(), expectedConfig.GetFirmwareVersion())
	assert.Equal(t, evaluationResult.FirmwareConfig.GetFirmwareFilename(), expectedConfig.GetFirmwareFilename())
	assert.Equal(t, evaluationResult.FirmwareConfig.GetFirmwareDownloadProtocol(), expectedConfig.GetFirmwareDownloadProtocol())
	//assert.Equal(t, evaluationResult.FirmwareConfig, expectedConfig)
	assert.Equal(t, evaluationResult.FirmwareConfig.GetFirmwareLocation(), expectedConfig.GetFirmwareLocation())
	assert.Equal(t, evaluationResult.FirmwareConfig.GetUpgradeDelay(), expectedConfig.GetUpgradeDelay())
}

func TestPercentageIs0AndRuleIsEqualLkgAndActive(t *testing.T) {
	initPercentConditions(t)

	notInMinChkFirmwareConfig := CreateAndSaveFirmwareConfig("notInMinChkVersion", modelPercentFilterTestCase.ID, "http", "stb")
	CreateAndSavePercentFilter(envModelFirmwareRulePercentFilterTestCase.Name, 100, firmwareConfigPercentFilterTestCase.ID, "", 0, []string{firmwareConfigPercentFilterTestCase.FirmwareVersion}, true, true, true, "stb")

	expectedConfig := createAndNullifyFirmwareConfigFacade(firmwareConfigPercentFilterTestCase)

	performAndVerifyRequest(firmwareConfigPercentFilterTestCase, expectedConfig, t)

	performAndVerifyRequest(notInMinChkFirmwareConfig, expectedConfig, t)
}

func TestPercentageIs100AndRuleIsNotEqualLkgAndActive(t *testing.T) {
	/**
	lkgFirmwareConfig := createAndSaveFirmwareConfig("lkgFirmwareVersion", model.getId(), FirmwareConfig.DownloadProtocol.http);
	notInMinChkConfig := createFirmwareConfig("notInMinChkVersion", model.getId(), FirmwareConfig.DownloadProtocol.http);

	createAndSavePercentFilter(envModelFirmwareRule.getName(), 100, lkgFirmwareConfig.getId(), null,
			100, Sets.newHashSet(firmwareConfig.getFirmwareVersion(), lkgFirmwareConfig.getFirmwareVersion()), true, true, true, ApplicationType.STB);

	FirmwareConfigFacade expectedRuleConfig = createAndNullifyFirmwareConfigFacade(firmwareConfig);
	FirmwareConfigFacade expectedLkgConfig = createAndNullifyFirmwareConfigFacade(lkgFirmwareConfig);

	performAndVerifyRequest(null, HttpStatus.OK, expectedLkgConfig);

	performAndVerifyRequest(notInMinChkConfig, HttpStatus.OK, expectedRuleConfig);

	performAndVerifyRequest(contextConfig, HttpStatus.OK, expectedLkgConfig);

	performAndVerifyRequest(firmwareConfig, HttpStatus.OK, expectedRuleConfig);
	**/
}

/*** original java code using Mock
func performAndVerifyRequest(FirmwareConfig firmwareConfigForRequest, HttpStatus status, FirmwareConfigFacade expectedConfig) throws Exception {
	ResultActions resultActions = null;
	if (firmwareConfigForRequest !=  null) {
		resultActions = mockMvc.perform(get("/xconf/swu/stb")
				.param("eStbMac", macList.getData().iterator().next())
				.param("env", environment.getId())
				.param("model", model.getId())
				.param("firmwareVersion", firmwareConfigForRequest.getFirmwareVersion()))
				.andExpect(status().is(status.value()));
	} else {
		resultActions = mockMvc.perform(get("/xconf/swu/stb")
				.param("eStbMac", macList.getData().iterator().next())
				.param("env", environment.getId())
				.param("model", model.getId()))
				.andExpect(status().is(status.value()));
	}
	verifyResponseContent(resultActions, status, expectedConfig);
}
**/
