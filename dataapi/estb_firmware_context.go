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
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"xconfwebconfig/dataapi/estbfirmware"
	"xconfwebconfig/shared"

	"xconfwebconfig/common"
	xhttp "xconfwebconfig/http"
	coreef "xconfwebconfig/shared/estbfirmware"
	"xconfwebconfig/shared/firmware"
	"xconfwebconfig/util"

	"github.com/agrison/go-commons-lang/stringUtils"
	log "github.com/sirupsen/logrus"
)

var (
	baseFields = []string{
		common.ESTB_MAC,
		common.ENV,
		common.MODEL,
		common.FIRMWARE_VERSION,
		common.IP_ADDRESS,
		common.TIME,
		common.TIME_ZONE_OFFSET,
		common.CAPABILITIES,
	}
	baseProperties = []string{
		common.ID,
		common.UPDATED,
		common.DESCRIPTION,
		common.SUPPORTED_MODEL_IDS,
		common.FIRMWARE_DOWNLOAD_PROTOCOL,
		common.FIRMWARE_DOWNLOAD_PROTOCOL,
		common.FIRMWARE_FILENAME,
		common.FIRMWARE_LOCATION,
		common.FIRMWARE_VERSION,
		common.IPV6_FIRMWARE_LOCATION,
		common.UPGRADE_DELAY,
		common.REBOOT_IMMEDIATELY,
		common.APPLICATION_TYPE,
	}
)

func IsMacPresentAndValid(queryParams url.Values) (bool, string, string) {
	var mac string
	var errorStr string
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			if k == common.MAC {
				mac = v[0]
			}
		}
	}
	if mac == "" {
		errorStr = fmt.Sprintf("Required String parameter '%s' is not present", common.MAC)
		return false, mac, errorStr
	}
	if !util.IsValidMacAddress(mac) {
		errorStr = fmt.Sprintf("Mac is invalid: %s", mac)
		return false, mac, errorStr
	}
	return true, mac, errorStr
}

func GetTimeInLocalTimezone(currentTime time.Time, contextMap map[string]string) {
	if contextMap[common.TIME_ZONE_OFFSET] != "" {
		offsetList := strings.Split(contextMap[common.TIME_ZONE_OFFSET], ":")
		if len(offsetList) == 2 {
			hours, err1 := strconv.Atoi(offsetList[0])
			mins, err2 := strconv.Atoi(offsetList[1])
			if err1 == nil && err2 == nil && hours >= -23 && hours <= 23 && mins >= 0 && mins <= 59 {
				currentTime = currentTime.Add(time.Hour * time.Duration(hours))
				if hours < 0 {
					mins *= -1
				}
				currentTime = currentTime.Add(time.Minute * time.Duration(mins))
			}
		}
	}
	contextMap[common.TIME] = currentTime.Format(common.DATE_TIME_FORMATTER)
}

func NormalizeEstbFirmwareContext(ws *xhttp.XconfServer, r *http.Request, contextMap map[string]string, usePartnerAppType bool, shouldAddIp bool) {
	NormalizeCommonContext(contextMap, common.ESTB_MAC, common.ECM_MAC)
	if contextMap[common.TIME] == "" {
		GetTimeInLocalTimezone(time.Now().UTC(), contextMap)
	}
	bypassFilters := contextMap[common.BYPASS_FILTERS]
	if strings.Contains(bypassFilters, estbfirmware.PERCENT_FILTER_NAME) {
		contextMap[common.BYPASS_FILTERS] = fmt.Sprintf("%s,%s", bypassFilters, firmware.GLOBAL_PERCENT)
	}
	if shouldAddIp {
		ipAddress := util.GetIpAddress(r, contextMap[common.IP_ADDRESS])
		contextMap[common.IP_ADDRESS] = ipAddress
	}
	// check if request is for partner
	if usePartnerAppType && contextMap[common.APPLICATION_TYPE] == shared.STB {
		if appType := GetApplicationTypeFromPartnerId(contextMap[common.PARTNER_ID]); appType != "" {
			contextMap[common.APPLICATION_TYPE] = appType
		}
	}
}

func GetExplanation(contextMap map[string]string, evaluationResult *estbfirmware.EvaluationResult) string {
	var input strings.Builder
	for key, value := range contextMap {
		fmt.Fprintf(&input, "%s=%s\n", key, value)
	}
	var explanation strings.Builder
	if evaluationResult.MatchedRule == nil {
		fmt.Fprintf(&explanation, "Request: %s\\ndid not match any rule.", input.String())
	} else {
		if evaluationResult.FirmwareConfig == nil && evaluationResult.Blocked {
			fmt.Fprintf(&explanation, "Request: %s\\n matched %s %s: %s\\n and blocked by Distribution percent in %s", input.String(), evaluationResult.MatchedRule.Type, evaluationResult.MatchedRule.ID, evaluationResult.MatchedRule.Name, evaluationResult.MatchedRule.ApplicableAction)
		} else if evaluationResult.FirmwareConfig == nil {
			fmt.Fprintf(&explanation, "Request: %s\\n matched NO OP %s %s: %s\\n received NO config.", input.String(), evaluationResult.MatchedRule.Type, evaluationResult.MatchedRule.ID, evaluationResult.MatchedRule.Name)
		} else {
			fmt.Fprintf(&explanation, "Request: %s\\n matched %s %s: %s\\n received config: %+v", input.String(), evaluationResult.MatchedRule.Type, evaluationResult.MatchedRule.ID, evaluationResult.MatchedRule.Name, evaluationResult.FirmwareConfig)
			if len(evaluationResult.AppliedFilters) > 0 {
				filter := evaluationResult.AppliedFilters[len(evaluationResult.AppliedFilters)-1]
				var filterString string
				switch filter.(type) {
				case *firmware.FirmwareRule:
					firmwareRule := filter.(*firmware.FirmwareRule)
					switch firmwareRule.Type {
					case firmware.TIME_FILTER:
						// TODO add in time filter values
						filterString = fmt.Sprintf(" %s[ id=%s name=%s start= end= isLocalTime= ipWhitelist=[] neverBlockRebootDecoupled= neverBlockHttpDownload= envModelWhitelist= ]", firmwareRule.Type, firmwareRule.ID, firmwareRule.Name)
					case firmware.DOWNLOAD_LOCATION_FILTER:
					case firmware.IP_FILTER:
						filterString = fmt.Sprintf(" %s [ %s %s ]", firmwareRule.Type, firmwareRule.ID, firmwareRule.Name)
					default:
						filterString = fmt.Sprintf("%s[ FirmwareRule{id=%s, name=%s, type=%s} ]", firmwareRule.Type, firmwareRule.ID, firmwareRule.Name, firmwareRule.Type)
					}
				case coreef.PercentFilterValue:
					percentFilterValue := filter.(coreef.PercentFilterValue)
					filterString = fmt.Sprintf("com.comcast.xconf.estbfirmware.PercentFilter [ percent=%f , envModelPercentage=%+v ]", percentFilterValue.Percentage, percentFilterValue.EnvModelPercentages)
				case coreef.DownloadLocationRoundRobinFilterValue:
					filterString = fmt.Sprintf("SINGLETON_%s %s", stringUtils.SubstringBeforeLast(filter.(coreef.DownloadLocationRoundRobinFilterValue).ID, "_VALUE"), filter.(coreef.DownloadLocationRoundRobinFilterValue).ID)
				case firmware.RuleAction:
					filterString = fmt.Sprintf("DistributionPercent in %s", filter)
				case coreef.PercentageBean:
					percentageBean := filter.(coreef.PercentageBean)
					filterString = fmt.Sprintf("DistributedEnvModelPercentage{id=%s, name=%s, firmwareCheckRequired=%s, lastKnownGood=%s, intermediateVersion=%s, firmwareVersions=%s}", percentageBean.ID, percentageBean.Name, percentageBean.FirmwareVersions, percentageBean.LastKnownGood, percentageBean.IntermediateVersion, percentageBean.FirmwareVersions)
				default:
					filterString = ""
				}
				fmt.Fprintf(&explanation, "\n was blocked/modified by filter %s", filterString)
			}
		}
	}
	return explanation.String()
}

func IsAllowedRequest(contextMap map[string]string, clientProtocolHeader string) bool {
	if IsSecureConnection(clientProtocolHeader) {
		return true
	}
	recoveryFirmwareVersions := Xc.EstbRecoveryFirmwareVersions
	if recoveryFirmwareVersions != "" {
		combinations := strings.Split(recoveryFirmwareVersions, ";")
		for _, value := range combinations {
			// splits string at whitespace
			parts := strings.Fields(value)
			if len(parts) != 2 {
				log.Warn(fmt.Sprintf("Wrong format for recoveryFirmwareVersions. Each combination should contain 2 parts. Got %s", value))
			} else {
				if contextMap[common.FIRMWARE_VERSION] != "" && contextMap[common.MODEL] != "" {
					matched1, _ := regexp.MatchString(parts[0], contextMap[common.FIRMWARE_VERSION])
					matched2, _ := regexp.MatchString(parts[1], contextMap[common.MODEL])
					if matched1 && matched2 {
						return true
					}
				}
			}
		}
	}
	return false
}

// AddEstbFirmwareContext ..
func AddEstbFirmwareContext(ws *xhttp.XconfServer, r *http.Request, contextMap map[string]string, usePartnerAppType bool, shouldAddIp bool, vargs ...log.Fields) error {
	var fields log.Fields
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}
	NormalizeEstbFirmwareContext(ws, r, contextMap, usePartnerAppType, shouldAddIp)

	// getting local sat token
	localToken, err := xhttp.GetLocalSatToken(fields)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error getting sat token")
		return err
	}
	satToken := localToken.Token

	if util.IsUnknownValue(contextMap[common.PARTNER_ID]) {
		partnerId := GetPartnerFromAccountServiceByHostMac(ws, contextMap[common.ESTB_MAC], satToken, fields)
		if partnerId != "" {
			contextMap[common.PARTNER_ID] = partnerId
		}
	}
	log.Debug(fmt.Sprintf("AddEstbFirmwareContext start ... contextMap %v", contextMap))
	AddContextFromTaggingService(ws, contextMap, satToken, "", false, fields)
	log.Debug(fmt.Sprintf("AddEstbFirmwareContext ... end contextMap %v", contextMap))
	return nil
}

func GetMissingAndEmptyQueryParams(contextMap map[string]string, missingFields *[]string, emptyFields *[]string) {
	fields := []string{common.ESTB_MAC, common.IP_ADDRESS, common.FIRMWARE_VERSION, common.MODEL, common.ENV}
	for _, field := range fields {
		val, ok := contextMap[field]
		if !ok {
			*missingFields = append(*missingFields, field)
		} else if val == "" {
			*emptyFields = append(*emptyFields, field)
		}
	}
}

func LogPreDisplayCleanup(lastConfigLog *coreef.ConfigChangeLog) {
	if lastConfigLog != nil {
		lastConfigLog.ID = ""
		lastConfigLog.Updated = 0
	}
}

func LogResponse(contextMap map[string]string, convertedContext *coreef.ConvertedContext, explanation string, evaluationResult *estbfirmware.EvaluationResult, fields log.Fields) {
	DoSplunkLog(contextMap, evaluationResult, fields)
	doMetrics(contextMap, evaluationResult, fields)

	if contextMap[common.FIRMWARE_VERSION] == "" {
		log.Debug("No firmware version given, not writing to last log.")
	} else {
		go func() {
			log.Trace("Logging last config request.")
			mac := contextMap[common.ESTB_MAC]
			lastConfigLog := coreef.NewConfigChangeLog(convertedContext, explanation, evaluationResult.FirmwareConfig, evaluationResult.AppliedFilters, evaluationResult.MatchedRule, true)
			err := coreef.SetLastConfigLog(mac, lastConfigLog)
			if err != nil {
				log.Error(fmt.Sprintf("Can't save last log config request: %+v", err))
			}
			if evaluationResult.MatchedRule != nil && !evaluationResult.Blocked && evaluationResult.FirmwareConfig != nil && !strings.EqualFold(contextMap[common.FIRMWARE_VERSION], evaluationResult.FirmwareConfig.GetFirmwareVersion()) {
				log.Trace(fmt.Sprintf("logging config change from %s to %s", evaluationResult.FirmwareConfig.GetFirmwareVersion(), contextMap[common.FIRMWARE_VERSION]))
				configChangeLog := coreef.NewConfigChangeLog(convertedContext, explanation, evaluationResult.FirmwareConfig, evaluationResult.AppliedFilters, evaluationResult.MatchedRule, false)
				err = coreef.SetConfigChangeLog(mac, configChangeLog)
				if err != nil {
					log.Error(fmt.Sprintf("Can't save config change log request: %+v", err))
				}
			}
		}()
	}
}

func IsCustomField(key string) bool {
	return !util.Contains(baseFields, key)
}

func IsAdditionalProperty(key string) bool {
	return !util.Contains(baseProperties, key)
}

func DoSplunkLog(contextMap map[string]string, evaluationResult *estbfirmware.EvaluationResult, fields log.Fields) {
	fields["estbMac"] = contextMap[common.ESTB_MAC]
	fields["env"] = contextMap[common.ENV]
	fields["model"] = contextMap[common.MODEL]
	fields["reportedFirmwareVersion"] = contextMap[common.FIRMWARE_VERSION]
	fields["ipAddress"] = contextMap[common.IP_ADDRESS]
	fields["timeZone"] = contextMap[common.TIME_ZONE]
	fields["time"] = contextMap[common.TIME]
	fields["capabilities"] = contextMap[common.CAPABILITIES]

	for key, value := range contextMap {
		if IsCustomField(key) {
			fields[key] = value
		}
	}

	if evaluationResult != nil {
		if evaluationResult.MatchedRule != nil {
			fields["appliedRule"] = evaluationResult.MatchedRule.Name
			fields["ruleType"] = evaluationResult.MatchedRule.Type
		}
		if evaluationResult.FirmwareConfig != nil {
			fields["firmwareVersion"] = evaluationResult.FirmwareConfig.GetFirmwareVersion()
			fields["firmwareDownloadProtocol"] = evaluationResult.FirmwareConfig.GetFirmwareDownloadProtocol()
			fields["firmwareLocation"] = evaluationResult.FirmwareConfig.GetFirmwareLocation()
			fields["rebootImmediately"] = evaluationResult.FirmwareConfig.GetRebootImmediately()

			for key, value := range evaluationResult.FirmwareConfig.Properties {
				if IsAdditionalProperty(key) {
					fields[key] = value
				}
			}
		} else if evaluationResult.Blocked {
			fields["firmwareVersion"] = estbfirmware.BLOCKED
		} else {
			fields["firmwareVersion"] = estbfirmware.NOMATCH
		}
		for key, value := range evaluationResult.AppliedVersionInfo {
			fields[key] = value
		}

		appliedFilters := []util.Dict{}
		if evaluationResult.AppliedFilters != nil && len(evaluationResult.AppliedFilters) > 0 {
			for _, filter := range evaluationResult.AppliedFilters {
				var d util.Dict
				switch v := filter.(type) {
				case *firmware.FirmwareRule:
					d = util.Dict{
						"type": v.Type,
						"name": v.Name,
					}
				case *coreef.PercentFilterValue:
					d = util.Dict{
						"type": "PercentFilter",
					}
				case *coreef.DownloadLocationRoundRobinFilterValue:
					d = util.Dict{
						"type": "DownloadLocationRoundRobinFilter",
					}
				case *firmware.RuleAction:
					d = util.Dict{
						"type": "DistributionPercentInRuleAction",
					}
				case *coreef.PercentageBean:
					d = util.Dict{
						"type": "PercentageBean",
						"name": v.Name,
					}
				}
				appliedFilters = append(appliedFilters, d)
			}
			fields["appliedFilters"] = appliedFilters
		}
	}
	log.WithFields(fields).Info("EstbFirmwareService XCONF_LOG")
	xhttp.UpdateLogCounter("EstbFirmwareService")
}

// doMetrics updates the fw penetration counts
// Duplicate some code here from DoSplunkLog
func doMetrics(contextMap map[string]string, evaluationResult *estbfirmware.EvaluationResult, fields log.Fields) {
	partner := contextMap[common.PARTNER_ID]
	model := contextMap[common.MODEL]

	fwVersion := contextMap[common.FIRMWARE_VERSION]
	if evaluationResult != nil {
		if evaluationResult.FirmwareConfig != nil {
			fwVersion = evaluationResult.FirmwareConfig.GetFirmwareVersion()
		} else if evaluationResult.Blocked {
			// Force this to a string
			fwVersion = string(estbfirmware.BLOCKED)
		} else {
			fwVersion = string(estbfirmware.NOMATCH)
		}
	}
	xhttp.UpdateFirmwarePenetrationCounts(partner, model, fwVersion)
	fields["partner"] = partner
	fields["model"] = model
	fields["firmware_version"] = fwVersion
	log.WithFields(fields).Debug("Firmware Penetration")
}

func GetFirstElementsInContextMap(contextMap map[string]string) {
	keys := []string{common.ESTB_MAC, common.ENV, common.MODEL, common.FIRMWARE_VERSION, common.ECM_MAC, common.RECEIVER_ID, common.CONTROLLER_ID, common.CHANNEL_MAP_ID, common.VOD_ID, common.ACCOUNT_HASH, common.XCONF_HTTP_HEADER, common.TIME, common.IP_ADDRESS, common.BYPASS_FILTERS, common.FORCE_FILTERS, common.TIME_ZONE_OFFSET, common.PARTNER_ID, common.ACCOUNT_ID}
	for _, key := range keys {
		if contextMap[key] != "" {
			contextMap[key] = strings.Split(contextMap[key], ",")[0]
		}
	}
}
