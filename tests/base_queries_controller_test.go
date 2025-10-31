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
	"encoding/base64"
	"strings"

	estb "github.com/rdkcentral/xconfwebconfig/dataapi/estbfirmware"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"github.com/google/uuid"
)

const (
	defaultModelId                   = "modelId"
	defaultEnvironmentId             = "environmentId"
	defaultEnvModelId                = "envModelId"
	defaultIpFilterId                = "ipFilterId"
	defaultTimeFilterId              = "timeFilterId"
	defaultRebootImmediatelyFilterId = "rebootImmediatelyFilterId"
	defaultFirmwareVersion           = "firmwareVersion"
	contextFirmwareVersion           = "contextFirmwareVersion"
	defaultIpRuleId                  = "ipRuleId"
	defaultMacRuleId                 = "macRuleId"
	defaultDownloadLocationFilterId  = "dowloadLocationFilterId"
	defaultIpListId                  = "ipListId"
	defaultMacListId                 = "macListId"
	defaultIpAddress                 = "1.1.1.1"
	defaultIpv6Address               = "::1"
	defaultMacAddress                = "11:11:11:11:11:11"
	defaultHttpLocation              = "httpLocation.com"
	defaultHttpFullUrlLocation       = "http://fullUrlLocation.com"
	defaultHttpsFullUrlLocation      = "https://fullUrlLocation.com"
	defaultFormulaId                 = "defaultFormulaObject"
	defaultFirmwareConfigId          = "firmwareConfigId"
	defaultPartnerId                 = "defaultpartnerid"
	defaultTimeZone                  = "Australia/Brisbane"
	defaultServiceAccountUri         = "defaultServiceAccountUri"
	defaultAccountId                 = "defaultAccountId"
	defaultFirmwareDownloadProtocol  = "http"
	defaultDeviceSettingName         = "deviceSettingsName"
	defaultLogUploadSettingName      = "logUploadSettingsName"

	API_VERSION            = "2"
	APPLICATION_TYPE_PARAM = "applicationType"
	WRONG_APPLICATION      = "wrongVersion"
)

var customBase64Encoding = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_").WithPadding(base64.NoPadding)

func DeleteAllEntities() {
	cachedTableList := []string{
		ds.TABLE_DCM_RULE,
		ds.TABLE_ENVIRONMENT,
		ds.TABLE_MODEL,
		ds.TABLE_FIRMWARE_CONFIG,
		ds.TABLE_FIRMWARE_RULE,
		ds.TABLE_FIRMWARE_RULE_TEMPLATE,
		ds.TABLE_SINGLETON_FILTER_VALUE,
		ds.TABLE_UPLOAD_REPOSITORY,
		ds.TABLE_LOG_FILE,
		ds.TABLE_LOG_FILE_LIST,
		ds.TABLE_LOG_FILES_GROUPS,
		ds.TABLE_LOG_UPLOAD_SETTINGS,
		ds.TABLE_SETTING_PROFILES,
		ds.TABLE_SETTING_RULES,
		ds.TABLE_DEVICE_SETTINGS,
		ds.TABLE_VOD_SETTINGS,
		ds.TABLE_TELEMETRY,
		ds.TABLE_PERMANENT_TELEMETRY,
		ds.TABLE_TELEMETRY_RULES,
		ds.TABLE_TELEMETRY_TWO_PROFILES,
		ds.TABLE_TELEMETRY_TWO_RULES,
		ds.TABLE_XCONF_FEATURE,
		ds.TABLE_FEATURE_CONTROL_RULE,
		ds.TABLE_LOGS,
		ds.TABLE_GENERIC_NS_LIST,
	}
	for _, table := range cachedTableList {
		tablemap, _ := ds.GetCachedSimpleDao().GetAllAsMap(table)
		for key := range tablemap {
			ds.GetCachedSimpleDao().DeleteOne(table, key.(string))
		}
	}
}

func CreateGenericNamespacedList(name string, ttype string, data string) *shared.GenericNamespacedList {
	namespacedList := shared.NewGenericNamespacedList(name, ttype, strings.Split(data, ","))
	return namespacedList
}

func CreateCondition(freeArg re.FreeArg, operation string, fixedArgValue string) *re.Condition {
	return re.NewCondition(&freeArg, operation, re.NewFixedArg(fixedArgValue))
}

func CreateRule(relation string, freeArg re.FreeArg, operation string, fixedArgValue string) *re.Rule {
	rule := re.Rule{}
	rule.SetRelation(relation)
	rule.SetCondition(CreateCondition(freeArg, operation, fixedArgValue))
	return &rule
}

func CreateRuleKeyValue(key string, value string) *re.Rule {
	condition := CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeString, key), re.StandardOperationIs, value)
	return &re.Rule{
		Condition: condition,
	}
}

func CreateAndSaveFirmwareRule(id string, templateId string, applicationType string, action *corefw.ApplicableAction, rule *re.Rule) *corefw.FirmwareRule {
	firmwareRule := CreateFirmwareRule(id, templateId, applicationType, action, rule)
	SetFirmwareRule(firmwareRule)
	return firmwareRule
}

func CreateFirmwareRule(id string, templateId string, applicationType string, action *corefw.ApplicableAction, rule *re.Rule) *corefw.FirmwareRule {
	firmwareRule := &corefw.FirmwareRule{
		ID:               id,
		Name:             id,
		Active:           true,
		ApplicableAction: action,
		ApplicationType:  applicationType,
		Type:             templateId,
		Rule:             *rule,
	}
	return firmwareRule
}

func SetFirmwareRule(firmwareRule *corefw.FirmwareRule) {
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_FIRMWARE_RULE, firmwareRule.ID, firmwareRule)
}

func CreateRuleAction(typ string, actiontyp corefw.ApplicableActionType, firmwareConfigId string) *corefw.ApplicableAction {
	ruleAction := corefw.NewApplicableActionAndType(typ, actiontyp, firmwareConfigId)
	return ruleAction
}

func CreateTemplateRuleAction(typ string, actiontyp corefw.ApplicableActionType, firmwareConfigId string) *corefw.TemplateApplicableAction {
	ruleAction := corefw.NewTemplateApplicableActionAndType(typ, actiontyp, firmwareConfigId)
	return ruleAction
}

func CreateDefaultEnvModelRule() *re.Rule {
	envModelRule := re.NewEmptyRule()
	envModelRule.AddCompoundPart(*CreateRule("", *coreef.RuleFactoryENV, re.StandardOperationIs, strings.ToUpper(defaultEnvironmentId)))
	envModelRule.AddCompoundPart(*CreateRule(re.RelationAnd, *coreef.RuleFactoryMODEL, re.StandardOperationIs, strings.ToUpper(defaultModelId)))
	return envModelRule
}

func CreateEnvModelRule(envId string, modelId string, namespacedListId string) *re.Rule {
	envModelRule := re.NewEmptyRule()
	envModelRule.AddCompoundPart(*CreateRule("", *coreef.RuleFactoryENV, re.StandardOperationIs, envId))
	envModelRule.AddCompoundPart(*CreateRule(re.RelationAnd, *coreef.RuleFactoryMODEL, re.StandardOperationIs, modelId))
	envModelRule.AddCompoundPart(*CreateRule(re.RelationAnd, *coreef.RuleFactoryMAC, *&coreef.RuleFactoryIN_LIST, namespacedListId))

	return envModelRule
}

func CreateExistsRule(tagName string) *re.Rule {
	condition := CreateCondition(*re.NewFreeArg(re.StandardFreeArgTypeAny, tagName), re.StandardOperationExists, "")
	rule := &re.Rule{
		Condition: condition,
	}
	return rule
}

func CreateAccountServicePartnerObject(partnerId string) xwhttp.AccountServiceDevices {
	AccountServiceObject := xwhttp.AccountServiceDevices{
		Id: uuid.New().String(),
		DeviceData: xwhttp.DeviceData{
			Partner:           partnerId,
			ServiceAccountUri: defaultServiceAccountUri,
		},
	}
	return AccountServiceObject
}

func CreateODPPartnerObject() xwhttp.DeviceServiceObject {
	odpObject := xwhttp.DeviceServiceObject{
		Status: 200,
		DeviceServiceData: &xwhttp.DeviceServiceData{
			AccountId: defaultServiceAccountUri,
		}}
	return odpObject
}

func CreateODPPartnerObjectWithPartnerAndTimezone() xwhttp.DeviceServiceObject {
	odpObject := xwhttp.DeviceServiceObject{
		Status: 200,
		DeviceServiceData: &xwhttp.DeviceServiceData{
			AccountId: defaultServiceAccountUri,
			PartnerId: defaultPartnerId,
			TimeZone:  defaultTimeZone,
		}}
	return odpObject
}

func CreateODPPartnerObjectWithPartnerAndTimezoneInvalid() xwhttp.DeviceServiceObject {
	odpObject := xwhttp.DeviceServiceObject{
		Status: 200,
		DeviceServiceData: &xwhttp.DeviceServiceData{
			AccountId: defaultServiceAccountUri,
			PartnerId: defaultPartnerId,
			TimeZone:  "InvalidTimeZone",
		}}
	return odpObject
}

func CreateAndSaveModel(id string) *shared.Model {
	model := shared.NewModel(id, "ModelDescription")
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_MODEL, model.ID, model)
	if err != nil {
		return nil
	}

	return model
}

func CreateAndSaveEnvironment(id string) *shared.Environment {
	env := shared.NewEnvironment(id, "ENV_MODEL_RULE_ENVIRONMENT_ID")
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_ENVIRONMENT, env.ID, env)
	if err != nil {
		return nil
	}

	return env
}

func CreateAndSaveGenericNamespacedList(name string, ttype string, data string) *shared.GenericNamespacedList {
	namespacedList := CreateGenericNamespacedList(name, ttype, data)
	err := ds.GetCachedSimpleDao().SetOne(ds.TABLE_GENERIC_NS_LIST, namespacedList.ID, namespacedList)
	if err != nil {
		return nil
	}
	return namespacedList
}

func CreateFirmwareConfig(firmwareVersion string, modelId string, firmwareDownloadProtocol string, applicationType string) *coreef.FirmwareConfig {
	firmwareConfig := coreef.NewEmptyFirmwareConfig()
	firmwareConfig.ID = uuid.New().String()
	firmwareConfig.Description = "FirmwareDescription"
	firmwareConfig.FirmwareFilename = "FirmwareFilename"
	firmwareConfig.FirmwareVersion = firmwareVersion
	firmwareConfig.FirmwareDownloadProtocol = firmwareDownloadProtocol
	firmwareConfig.ApplicationType = applicationType
	supportedModels := make([]string, 1)
	model := CreateAndSaveModel(strings.ToUpper(modelId))
	supportedModels[0] = model.ID
	return firmwareConfig
}

func CreateAndSaveFirmwareConfig(firmwareVersion string, modelId string, firmwareDownloadProtocol string, applicationType string) *coreef.FirmwareConfig {
	firmwareConfig := CreateFirmwareConfig(firmwareVersion, modelId, firmwareDownloadProtocol, applicationType)
	err := SetFirmwareConfig(firmwareConfig)
	if err != nil {
		return nil
	}
	return firmwareConfig
}

func SetFirmwareConfig(firmwareConfig *coreef.FirmwareConfig) error {
	err := coreef.CreateFirmwareConfigOneDB(firmwareConfig)
	if err != nil {
		return err
	}
	return nil
}

func CreatePercentageBean(name string, envId string, modelId string, whitelistId string, whitelistData string, firmwareVersion string, applicationType string) *coreef.PercentageBean {
	var whitelist string
	if whitelistId != "" {
		whitelist = CreateAndSaveGenericNamespacedList(whitelistId, "IP_LIST", whitelistData).ID
	}
	firmwareConfig := CreateAndSaveFirmwareConfig(firmwareVersion, modelId, "http", applicationType)
	configEntry := corefw.NewConfigEntry(firmwareConfig.ID, 0.0, 66.0)
	percentageBean := &coreef.PercentageBean{
		ID:                    uuid.New().String(),
		Name:                  name,
		Whitelist:             whitelist,
		Active:                true,
		Environment:           CreateAndSaveEnvironment(envId).ID,
		Model:                 CreateAndSaveModel(modelId).ID,
		FirmwareCheckRequired: true,
		ApplicationType:       applicationType,
		FirmwareVersions:      []string{firmwareConfig.FirmwareVersion},
		LastKnownGood:         firmwareConfig.ID,
		Distributions:         []*corefw.ConfigEntry{configEntry},
		IntermediateVersion:   firmwareConfig.ID,
	}
	return percentageBean
}

func CreateAndSaveFirmwareRuleTemplate(id string, rule *re.Rule, applicableAction *corefw.TemplateApplicableAction) *corefw.FirmwareRuleTemplate {
	template := CreateFirmwareRuleTemplate(id, rule, applicableAction)
	err := corefw.CreateFirmwareRuleTemplateOneDB(template)
	if err != nil {
		return nil
	}
	return template
}

func CreateFirmwareRuleTemplate(id string, rule *re.Rule, applicableAction *corefw.TemplateApplicableAction) *corefw.FirmwareRuleTemplate {
	template := corefw.NewEmptyFirmwareRuleTemplate()
	template.ID = id
	template.Rule = *rule
	template.ApplicableAction = applicableAction
	return template
}

func CreateAndSaveEnvModelFirmwareRule(name string, firmwareConfigId string, envId string, modelId string, macListId string) *corefw.FirmwareRule {
	envModelRule := corefw.NewEmptyFirmwareRule()
	envModelRule.ID = uuid.New().String()
	envModelRule.Name = name
	ruleAct := CreateRuleAction(corefw.RuleActionClass, corefw.RULE, firmwareConfigId)
	envModelRule.ApplicableAction = ruleAct
	envModelRule.Type = "ENV_MODEL_RULE"
	envModelRule.Rule = *CreateEnvModelRule(envId, modelId, macListId)
	err := corefw.CreateFirmwareRuleOneDB(envModelRule)
	if err != nil {
		return nil
	}
	return envModelRule
}

func CreateIpAddressGroup(stringIpAddresses []string) *shared.IpAddressGroup {
	return CreateIpAddressGroupWithName(uuid.New().String(), stringIpAddresses)
}

func CreateIpAddressGroupWithName(name string, stringIpAddresses []string) *shared.IpAddressGroup {
	return shared.NewIpAddressGroupWithAddrStrings(name, name, stringIpAddresses)
}

func CreateAndSavePercentFilter(
	envModelRuleName string,
	percentage float32,
	lastKnownGood string,
	intermediateVersion string,
	envModelPercent float32,
	firmwareVersions []string,
	isActive bool,
	isFirmwareCheckRequired bool,
	rebootImmediately bool,
	applicationType string) *coreef.PercentFilterValue {

	percentFilter := coreef.NewEmptyPercentFilterValue()
	whitelist := CreateIpAddressGroup([]string{"127.1.1.1", "127.1.1.2"})

	envModelPercentage := coreef.NewEnvModelPercentage()
	envModelPercentage.Whitelist = whitelist
	envModelPercentage.LastKnownGood = lastKnownGood
	envModelPercentage.IntermediateVersion = intermediateVersion
	envModelPercentage.FirmwareVersions = firmwareVersions
	envModelPercentage.Percentage = envModelPercent
	envModelPercentage.Active = isActive
	envModelPercentage.FirmwareCheckRequired = isFirmwareCheckRequired
	envModelPercentage.RebootImmediately = rebootImmediately

	percentFilter.Percentage = percentage
	percentFilter.Whitelist = whitelist
	mapEnvModes := make(map[string]coreef.EnvModelPercentage)
	mapEnvModes[envModelRuleName] = *envModelPercentage
	percentFilter.EnvModelPercentages = mapEnvModes

	percentFilterService := estb.NewPercentFilterService()
	percentFilterService.Save(percentFilter, applicationType)

	return percentFilter
}

func CreateContext(firmwareVersion string, modelId string, environmentId string, ipAddress string, eStbMac string) *coreef.ConvertedContext {
	contextMap := map[string]string{
		"firmwareVersion": firmwareVersion,
		"model":           modelId,
		"env":             environmentId,
		"ipAddress":       ipAddress,
		"eStbMac":         eStbMac,
	}
	context := coreef.GetContextConverted(contextMap)
	return context
}
