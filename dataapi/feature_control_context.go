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
	"strings"
	"time"

	"xconfwebconfig/common"
	xhttp "xconfwebconfig/http"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/rfc"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

func AddContextFromDeviceService(ws *xhttp.XconfServer, contextMap map[string]string, vargs ...log.Fields) *xhttp.DeviceServiceData {
	var fields log.Fields
	var deviceServiceData *xhttp.DeviceServiceData
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}
	if Xc.EnableDeviceService && contextMap[common.SERIAL_NUM] != "" {
		deviceServiceObject, err := ws.DeviceServiceConnector.GetMeshPodAccountBySerialNum(contextMap[common.SERIAL_NUM], fields)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Error getting DeviceService information")
			return deviceServiceData
		}
		if deviceServiceObject.Status == http.StatusOK && util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) && deviceServiceObject.DeviceServiceData.AccountId != "" {
			contextMap[common.ACCOUNT_ID] = deviceServiceObject.DeviceServiceData.AccountId
		}
		if deviceServiceObject.DeviceServiceData != nil {
			deviceServiceData = deviceServiceObject.DeviceServiceData
		}
	}
	return deviceServiceData
}

func AddFeatureControlContextFromAccountService(ws *xhttp.XconfServer, contextMap map[string]string, satToken string, vargs ...log.Fields) {
	var fields log.Fields
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}
	if Xc.EnableAccountService {
		var accountServiceObject xhttp.AccountServiceDevices
		var err error
		if util.IsValidMacAddress(contextMap[common.ESTB_MAC_ADDRESS]) {
			accountServiceObject, err = ws.AccountServiceConnector.GetDevices(common.HOST_MAC_PARAM, contextMap[common.ESTB_MAC_ADDRESS], satToken, fields)
		}
		if accountServiceObject.Id == "" && accountServiceObject.AccountServiceDeviceData.Partner == "" && accountServiceObject.AccountServiceDeviceData.ServiceAccountUri == "" && util.IsValidMacAddress(contextMap[common.ECM_MAC_ADDRESS]) {
			accountServiceObject, err = ws.AccountServiceConnector.GetDevices(common.ECM_MAC_PARAM, contextMap[common.ECM_MAC_ADDRESS], satToken, fields)
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Error getting AccountService information")
		} else {
			if accountServiceObject.AccountServiceDeviceData.Partner != "" {
				contextMap[common.PARTNER_ID] = strings.ToUpper(accountServiceObject.AccountServiceDeviceData.Partner)
			}
			if util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) && accountServiceObject.AccountServiceDeviceData.ServiceAccountUri != "" {
				contextMap[common.ACCOUNT_ID] = accountServiceObject.AccountServiceDeviceData.ServiceAccountUri
			}
			if util.IsUnknownValue(contextMap[common.ACCOUNT_HASH]) && accountServiceObject.AccountServiceDeviceData.ServiceAccountUri != "" {
				contextMap[common.ACCOUNT_HASH] = util.CalculateHash(accountServiceObject.AccountServiceDeviceData.ServiceAccountUri)
			}
		}
	}
}

func NormalizeFeatureControlContext(ws *xhttp.XconfServer, r *http.Request, contextMap map[string]string) {
	NormalizeCommonContext(contextMap, common.ESTB_MAC_ADDRESS, common.ECM_MAC_ADDRESS)
	estbIp := util.GetIpAddress(r, contextMap[common.ESTB_IP])
	contextMap[common.ESTB_IP] = estbIp
	// check if request is for partner
	if contextMap[common.APPLICATION_TYPE] == shared.STB {
		if appType := GetApplicationTypeFromPartnerId(contextMap[common.PARTNER_ID]); appType != "" {
			contextMap[common.APPLICATION_TYPE] = appType
		}
	}
}

// AddFeatureControlContext ..
func AddFeatureControlContext(ws *xhttp.XconfServer, r *http.Request, contextMap map[string]string, xconfHttp string, configSetHash string, vargs ...log.Fields) *xhttp.DeviceServiceData {
	var fields log.Fields
	var deviceServiceData *xhttp.DeviceServiceData
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}
	NormalizeFeatureControlContext(ws, r, contextMap)
	contextMap[common.PASSED_PARTNER_ID] = contextMap[common.PARTNER_ID]

	// getting local sat token
	localToken, err := xhttp.GetLocalSatToken(fields)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error getting sat token from SatService")
		return deviceServiceData
	}
	satToken := localToken.Token

	// if/else statement to check if we should call DeviceService or AccountService
	if strings.EqualFold("XPC", contextMap[common.ACCOUNT_MGMT]) && util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) {
		deviceServiceData = AddContextFromDeviceService(ws, contextMap, fields)

	} else if util.IsUnknownValue(contextMap[common.ACCOUNT_ID]) || util.IsUnknownValue(contextMap[common.PARTNER_ID]) || util.IsUnknownValue(contextMap[common.ACCOUNT_HASH]) {
		AddFeatureControlContextFromAccountService(ws, contextMap, satToken, fields)
	}
	AddContextFromTaggingService(ws, contextMap, satToken, configSetHash, true, fields)
	return deviceServiceData
}

// PostProcessFeatureControl
func PostProcessFeatureControl(ws *xhttp.XconfServer, featureControl *rfc.FeatureControl, contextMap map[string]string, isSecuredConnection bool, deviceServiceData *xhttp.DeviceServiceData) {
	if util.IsUnknownValue(contextMap[common.PASSED_PARTNER_ID]) && contextMap[common.PARTNER_ID] != "" {
		partnerFeature := rfc.Feature{
			Name:               common.SYNDICATION_PARTNER,
			FeatureName:        common.SYNDICATION_PARTNER,
			EffectiveImmediate: true,
			Enable:             true,
			ConfigData: map[string]string{
				common.TR181_DEVICE_TYPE_PARTNER_ID: strings.ToLower(contextMap[common.PARTNER_ID]),
			},
		}
		featureControl.FeatureResponses = append(featureControl.FeatureResponses, rfc.CreateFeatureResponseObject(partnerFeature))
	}

	if isSecuredConnection && Xc.ReturnAccountId && contextMap[common.ACCOUNT_ID] != "" {
		accountIdFeature := rfc.Feature{
			Name:               "AccountId",
			FeatureName:        "AccountId",
			EffectiveImmediate: true,
			Enable:             true,
			ConfigData: map[string]string{
				common.TR181_DEVICE_TYPE_ACCOUNT_ID: contextMap[common.ACCOUNT_ID],
			},
		}
		accountIdFeatureResponse := rfc.CreateFeatureResponseObject(accountIdFeature)
		if deviceServiceData != nil {
			accountIdFeatureResponse["accountId"] = deviceServiceData.AccountId
			if deviceServiceData.PartnerId != "" {
				accountIdFeatureResponse["partnerId"] = deviceServiceData.PartnerId
			} else {
				accountIdFeatureResponse["partnerId"] = "unknown"
			}
			if deviceServiceData.TimeZone != "" {
				accountIdFeatureResponse["timeZone"] = deviceServiceData.TimeZone
				loc, err := time.LoadLocation(deviceServiceData.TimeZone)
				if err == nil {
					t := time.Now().In(loc)
					accountIdFeatureResponse["tzUTCOffset"] = fmt.Sprintf("UTC%s", t.Format("-07:00"))
				} else {
					accountIdFeatureResponse["tzUTCOffset"] = "unknown"
				}

			} else {
				accountIdFeatureResponse["timeZone"] = "unknown"
				accountIdFeatureResponse["tzUTCOffset"] = "unknown"
			}
		}
		featureControl.FeatureResponses = append(featureControl.FeatureResponses, accountIdFeatureResponse)
	}

	if Xc.ReturnAccountHash && contextMap[common.ACCOUNT_HASH] != "" {
		accountHashFeature := rfc.Feature{
			Name:               "AccountHash",
			FeatureName:        "AccountHash",
			EffectiveImmediate: true,
			Enable:             true,
			ConfigData: map[string]string{
				common.TR181_DEVICE_TYPE_ACCOUNT_HASH: contextMap[common.ACCOUNT_HASH],
			},
		}
		featureControl.FeatureResponses = append(featureControl.FeatureResponses, rfc.CreateFeatureResponseObject(accountHashFeature))
	}
}
