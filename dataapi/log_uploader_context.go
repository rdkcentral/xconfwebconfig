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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/rdkcentral/xconfwebconfig/common"
	loguploader "github.com/rdkcentral/xconfwebconfig/dataapi/dcm/logupload"
	"github.com/rdkcentral/xconfwebconfig/dataapi/dcm/telemetry"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	log "github.com/sirupsen/logrus"
)

func NormalizeLogUploaderContext(ws *xhttp.XconfServer, r *http.Request, contextMap map[string]string, usePartnerAppType bool, fields log.Fields) {
	NormalizeCommonContext(contextMap, common.ESTB_MAC_ADDRESS, common.ECM_MAC_ADDRESS)
	estbIp := util.GetIpAddress(r, contextMap[common.ESTB_IP], fields)
	contextMap[common.ESTB_IP] = estbIp
	fields[common.ESTB_IP] = estbIp
	// check if request is for partner
	if usePartnerAppType && contextMap[common.APPLICATION_TYPE] == shared.STB {
		if appType := GetApplicationTypeFromPartnerId(contextMap[common.PARTNER_ID]); appType != "" {
			contextMap[common.APPLICATION_TYPE] = appType
		}
	}
}

// AddLogUploaderContext ..
func AddLogUploaderContext(ws *xhttp.XconfServer, r *http.Request, contextMap map[string]string, usePartnerAppType bool, vargs ...log.Fields) error {
	var fields log.Fields
	var accountId string
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}

	NormalizeLogUploaderContext(ws, r, contextMap, usePartnerAppType, fields)

	localToken, err := xhttp.GetLocalSatToken(fields)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error getting sat token from codebig")
		return err
	}
	satToken := localToken.Token
	if Xc.EnableXacGroupService {
		if util.IsValidMacAddress(contextMap[common.ESTB_MAC_ADDRESS]) || util.IsValidMacAddress(contextMap[common.ECM_MAC_ADDRESS]) {
			var macAddress string
			if contextMap[common.ESTB_MAC_ADDRESS] != "" {
				macAddress = util.GetEcmMacAddress(util.AlphaNumericMacAddress(strings.TrimSpace(contextMap[common.ESTB_MAC_ADDRESS])))
			} else {
				macAddress = contextMap[common.ECM_MAC_ADDRESS]
			}

			macValue := util.RemoveNonAlphabeticSymbols(macAddress)
			xboAccount, err := ws.GroupServiceConnector.GetAccountIdData(macValue, fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Error getting accountId information from Grp Service")
			}
			if xboAccount != nil && xboAccount.GetAccountId() != "" {
				accountId = xboAccount.GetAccountId()
				contextMap[common.ACCOUNT_ID] = accountId
			} else {
				xhttp.IncreaseAccountServiceEmptyResponseCounter(contextMap[common.MODEL])
			}

			accountProducts, err := ws.GroupServiceConnector.GetAccountProducts(accountId, fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Error getting accountProducts information from Grp Service")
			} else {
				if partner, ok := accountProducts["Partner"]; ok && partner != "" {
					contextMap[common.PARTNER_ID] = strings.ToUpper(partner)
				}

				contextMap[common.ACCOUNT_HASH] = util.CalculateHash(accountId)

				if countryCode, ok := accountProducts["CountryCode"]; ok {
					contextMap[common.COUNTRY_CODE] = countryCode
				}

				if raw, ok := accountProducts["AccountProducts"]; ok && raw != "" {
					var ap map[string]string
					err := json.Unmarshal([]byte(accountProducts["AccountProducts"]), &ap)
					if err == nil {
						for key, val := range ap {
							contextMap[key] = val
						}
					} else {
						log.WithFields(fields).Error("Failed to unmarshall AccountProducts")
					}

				}
				xhttp.IncreaseGrpServiceFetchCounter(contextMap[common.MODEL], contextMap[common.PARTNER_ID])
				log.WithFields(fields).Debugf("AddFeatureControlContextFromAccountService AcntId='%s' ,AccntPrd='%s'  retrieved from xac/ada", contextMap[common.ACCOUNT_ID], contextMap[common.ACCOUNT_PRODUCTS])
			}
		}
	}

	if contextMap[common.PARTNER_ID] == "" {
		if util.IsUnknownValue(contextMap[common.PARTNER_ID]) {
			partnerId := GetPartnerFromAccountServiceByHostMac(ws, contextMap[common.ESTB_MAC_ADDRESS], satToken, fields)
			if partnerId != "" {
				contextMap[common.PARTNER_ID] = partnerId
			}
		}
	}
	AddContextFromTaggingService(ws, contextMap, satToken, "", false, fields)
	return nil
}

func ToTelemetry2Profile(telemetryProfile []logupload.TelemetryElement) []logupload.TelemetryElement {
	for index, element := range telemetryProfile {
		if element.Component != "" {
			telemetryProfile[index].Content = element.Component
			telemetryProfile[index].Type = "<event>"
			telemetryProfile[index].Component = ""
		}
	}
	return telemetryProfile
}

func NullifyUnwantedFields(permanentTelemetryProfile *logupload.PermanentTelemetryProfile) {
	if permanentTelemetryProfile != nil {
		telemetryProfile := permanentTelemetryProfile.TelemetryProfile
		for index := range telemetryProfile {
			telemetryProfile[index].ID = ""
			telemetryProfile[index].Component = ""
		}
	}
}

func CleanupLusUploadRepository(settings *logupload.Settings, apiVersion string) {
	if settings != nil {
		if util.IsVersionGreaterOrEqual(apiVersion, 2.0) {
			settings.LusUploadRepositoryURL = ""
		} else {
			settings.LusUploadRepositoryUploadProtocol = ""
			settings.LusUploadRepositoryURLNew = ""
		}
	}
}

func LogResultSettings(settings *logupload.Settings, telemetryRule *logupload.TelemetryRule, settingRules []*logupload.SettingRule, fields log.Fields) {
	ruleNames := make([]string, 0, len(settings.RuleIDs))
	for ruleId, _ := range settings.RuleIDs {
		dcmRule := loguploader.GetOneDcmRuleFunc(ruleId)
		if dcmRule != nil && len(dcmRule.Name) > 0 {
			ruleNames = append(ruleNames, dcmRule.Name)
		}
	}
	settingRuleNames := make([]string, 0, len(settingRules))
	if len(settingRules) > 0 {
		for _, settingRule := range settingRules {
			settingRuleNames = append(settingRuleNames, settingRule.Name)
		}
	}
	telemetryRuleName := "NoMatch"
	if telemetryRule != nil {
		telemetryRuleName = telemetryRule.Name
	}

	fields["formulaNames"] = ruleNames
	fields["telemetryRuleName"] = telemetryRuleName
	fields["settingRuleNames"] = settingRuleNames
	log.WithFields(common.FilterLogFields(fields)).Info("LogUploaderService AppliedRules")
}

type TelemetryEvaluationResult struct {
	RulesMatched bool
	ProfilesData []util.Dict
}

func GetTelemetryTwoProfileResponeDicts(contextMap map[string]string, fields log.Fields) (*TelemetryEvaluationResult, error) {
	telemetryProfileService := telemetry.NewTelemetryProfileService()
	matchedRules := telemetryProfileService.ProcessTelemetryTwoRules(contextMap)
	matchedProfiles := telemetryProfileService.GetTelemetryTwoProfileByTelemetryRules(matchedRules, fields)
	dicts := []util.Dict{}
	for _, profile := range matchedProfiles {
		// profile = nil should not happen
		var valueDict util.Dict
		var err = json.Unmarshal([]byte(profile.Jsonconfig), &valueDict)
		if err != nil {
			return nil, err
		}

		profileData := util.Dict{
			"name":        profile.Name,
			"versionHash": util.GetCRC32HashValue(profile.Jsonconfig),
			"value":       valueDict,
		}
		dicts = append(dicts, profileData)
	}
	evaluationResult := &TelemetryEvaluationResult{}
	if len(matchedRules) > 0 {
		evaluationResult.RulesMatched = true
	}
	evaluationResult.ProfilesData = dicts
	return evaluationResult, nil
}
