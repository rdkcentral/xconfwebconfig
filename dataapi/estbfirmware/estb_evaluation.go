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
package estbfirmware

import (
	"math"
	"math/rand"
	"strconv"
	"strings"

	"xconfwebconfig/shared"
	coreef "xconfwebconfig/shared/estbfirmware"
	"xconfwebconfig/shared/firmware"
)

// EvaluationResult ...
type EvaluationResult struct {
	MatchedRule        *firmware.FirmwareRule
	AppliedFilters     []interface{}
	FirmwareConfig     *coreef.FirmwareConfigFacade
	Description        string
	Blocked            bool
	AppliedVersionInfo map[string]string
}

func NewEvaluationResult() *EvaluationResult {
	return &EvaluationResult{
		MatchedRule:        nil,
		AppliedFilters:     []interface{}{},
		FirmwareConfig:     nil,
		Description:        "",
		Blocked:            false,
		AppliedVersionInfo: make(map[string]string),
	}
}

func (e *EvaluationResult) AddAppliedFilters(filter interface{}) {
	if e.AppliedFilters == nil {
		e.AppliedFilters = make([]interface{}, 0)
	}
	e.AppliedFilters = append(e.AppliedFilters, filter)
}

// DownloadLocationRoundRobinFilter no need this class due to there is data member

// DownloadLocationRoundRobinFilterFilter ...
// @return true if filter is applied, false otherwise
func DownloadLocationRoundRobinFilterFilter(firmwareConfig *coreef.FirmwareConfigFacade, filterValue *coreef.DownloadLocationRoundRobinFilterValue, context *coreef.ConvertedContext) bool {
	supportsFullHttpUrl := context.IsSupportsFullHttpUrl()
	firmwareConfig.SetStringValue(coreef.FIRMWARE_DOWNLOAD_PROTOCOL, "http")
	firmwareConfig.SetFirmwareDownloadProtocol("http")
	if len(filterValue.HttpLocation) != 0 && len(filterValue.HttpFullUrlLocation) != 0 {
		secureConnection := false
		if context.IsXconfHttpHeaderSecureConnection() {
			secureConnection = true
		}
		if supportsFullHttpUrl {
			DownloadLocationRoundRobinFilterSetLocationByConnectionType(secureConnection, firmwareConfig, filterValue.HttpFullUrlLocation)
		} else {
			firmwareConfig.SetStringValue(coreef.FIRMWARE_LOCATION, filterValue.HttpLocation)
			firmwareConfig.SetFirmwareLocation(filterValue.HttpLocation)
		}
		return true
	}

	isIPv4LocationApplied := DownloadLocationRoundRobinFilterSetupIPv4Location(firmwareConfig, filterValue)
	isIPv6LocationApplied := DownloadLocationRoundRobinFilterSetupIPv6Location(firmwareConfig, filterValue)
	return isIPv4LocationApplied || isIPv6LocationApplied
}

// DownloadLocationRoundRobinFilterSetLocationByConnectionType ...
func DownloadLocationRoundRobinFilterSetLocationByConnectionType(secureConnection bool, firmwareConfig *coreef.FirmwareConfigFacade, fullHttpLocation string) {
	if !secureConnection {
		fullHttpLocation = strings.ReplaceAll(fullHttpLocation, "https", "http")
	} else if !strings.HasPrefix(fullHttpLocation, "https") {
		fullHttpLocation = strings.ReplaceAll(fullHttpLocation, "http", "https")
	}
	firmwareConfig.SetStringValue(coreef.FIRMWARE_LOCATION, fullHttpLocation)

}

// DownloadLocationRoundRobinFilterSetupIPv6Location ...
func DownloadLocationRoundRobinFilterSetupIPv6Location(firmwareConfig *coreef.FirmwareConfigFacade, filterValue *coreef.DownloadLocationRoundRobinFilterValue) bool {
	rand := rand.Float64()
	limit := 0.0
	isApplied := false
	if filterValue.Ipv6locations != nil && len(filterValue.Ipv6locations) > 0 {
		for _, location := range filterValue.Ipv6locations {
			limit += location.Percentage / 100.00
			if rand < limit {
				firmwareConfig.SetStringValue(coreef.IPV6_FIRMWARE_LOCATION, location.LocationIp)
				isApplied = true
				break
			}
		}
	}
	return isApplied
}

func DownloadLocationRoundRobinFilterSetupIPv4Location(firmwareConfig *coreef.FirmwareConfigFacade, filterValue *coreef.DownloadLocationRoundRobinFilterValue) bool {
	rand := rand.Float64()
	limit := 0.0
	isApplied := false
	if filterValue.Locations != nil && len(filterValue.Locations) != 0 {
		for _, location := range filterValue.Locations {
			limit += location.Percentage / 100.00
			if rand < limit {
				firmwareConfig.SetStringValue(coreef.FIRMWARE_LOCATION, location.LocationIp)
				isApplied = true
				break
			}
		}
	}
	return isApplied
}

func DownloadLocationRoundRobinFilterContainsVersion(firmwareVersions string, contextVersion string) bool {
	split := strings.Split(firmwareVersions, " ")
	for _, s := range split {
		if contextVersion == s {
			return true
		}
	}
	return false
}

/**
 * @return true if firmware output must be returned, false if must be blocked
 */
func PercentFilterfilter(evaluationResult *EvaluationResult, context *coreef.ConvertedContext) bool {
	filterValue, _ := coreef.GetDefaultPercentFilterValueOneDB()
	matchedEnvModelName := ""
	if evaluationResult.MatchedRule != nil && firmware.ENV_MODEL_RULE == evaluationResult.MatchedRule.Type {
		matchedEnvModelName = evaluationResult.MatchedRule.Name
	}

	envModelPercentages := filterValue.EnvModelPercentages
	if matchedEnvModelName != "" && envModelPercentages != nil {
		envModelPercentage, ok := envModelPercentages[matchedEnvModelName]
		if ok && envModelPercentage.Active {
			whiteList := envModelPercentage.Whitelist
			percentage := envModelPercentage.Percentage
			firmwareVersionIsAbsentInFilter := false
			if envModelPercentage.FirmwareVersions == nil || context.GetFirmwareVersionConverted() == "" {
				firmwareVersionIsAbsentInFilter = true
			} else if !firmware.HasFirmwareVersion(envModelPercentage.FirmwareVersions, context.GetFirmwareVersionConverted()) {
				firmwareVersionIsAbsentInFilter = true
			}
			if envModelPercentage.FirmwareCheckRequired && firmwareVersionIsAbsentInFilter {
				if envModelPercentage.RebootImmediately {
					context.AddForceFiltersConverted(firmware.REBOOT_IMMEDIATELY_FILTER)
				}
				context.AddBypassFiltersConverted(firmware.TIME_FILTER)
				config, _ := coreef.GetFirmwareConfigOneDB(envModelPercentage.IntermediateVersion)
				if config != nil && context.GetFirmwareVersionConverted() != config.FirmwareVersion {
					// return IntermediateVersion firmware config
					evaluationResult.FirmwareConfig = coreef.NewFirmwareConfigFacade(config)
					evaluationResult.AppliedVersionInfo["firmwareVersionSource"] = "IV,doesntMeetMinCheck"
				} else {
					config, _ := coreef.GetFirmwareConfigOneDB(envModelPercentage.LastKnownGood)
					if config != nil {
						// return LKG firmware config
						evaluationResult.FirmwareConfig = coreef.NewFirmwareConfigFacade(config)
						evaluationResult.AppliedVersionInfo["firmwareVersionSource"] = "LKG,doesntMeetMinCheck"
					}
				}
				return true
			}
			result := fitsPercent(evaluationResult, context, whiteList, percentage)
			if !result {
				config, _ := coreef.GetFirmwareConfigOneDB(envModelPercentage.LastKnownGood)
				if config != nil && context.GetFirmwareVersionConverted() != config.FirmwareVersion {
					// return LKG firmware config if versions are different
					evaluationResult.FirmwareConfig = coreef.NewFirmwareConfigFacade(config)
					evaluationResult.AppliedVersionInfo["firmwareVersionSource"] = "LKG,meetMinCheck"
					return true
				}
			}
			return result
		}
	}
	return fitsPercent(evaluationResult, context, filterValue.Whitelist, filterValue.Percentage)
}

func fitsPercent(evaluationResult *EvaluationResult, context *coreef.ConvertedContext, whiteList *shared.IpAddressGroup, percentage float32) bool {
	isInWhiteList := false
	fitsPercent := false

	if whiteList != nil && whiteList.IsInRange(context.GetIpAddressConverted()) {
		evaluationResult.AppliedVersionInfo["inWhiteList"] = strconv.FormatBool(isInWhiteList)
		accountId := ""
		mac := context.GetEcmMacConverted()
		if len(mac) == 0 {
			accountId = context.ToString()
		} else {
			accountId = mac
		}
		fitsPercent := fitsPercentByAccountId(accountId, float64(percentage))
		evaluationResult.AppliedVersionInfo["fitsPercent"] = strconv.FormatBool(fitsPercent)
	}
	return isInWhiteList || fitsPercent
}

func fitsPercentByAccountId(accountId string, percent float64) bool {
	prange := math.MaxFloat32*2 + 1
	//todo
	//hashCode := (double) Hashing.sipHash24().hashString(accountId, Charsets.UTF_8).asLong() + offset // from 0 to (2 * Long.MAX_VALUE + 1)
	limit := percent / 100 * prange // from 0 to (2 * Long.MAX_VALUE + 1)
	hashCode := 1.1
	return hashCode <= limit
}
