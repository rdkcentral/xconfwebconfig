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
	"strings"

	common "xconfwebconfig/common"
	xhttp "xconfwebconfig/http"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

// Ws - webserver object
var (
	Ws *xhttp.XconfServer
	Xc *XconfConfigs
)

// WebServerInjection - dependency injection
func WebServerInjection(ws *xhttp.XconfServer, xc *XconfConfigs) {
	Ws = ws
	if ws == nil {
		common.CacheUpdateWindowSize = 60000
	} else {
		common.CacheUpdateWindowSize = ws.ServerConfig.GetInt64("xconfwebconfig.xconf.cache_update_window_size")
	}
	Xc = xc
}

func NormalizeCommonContext(contextMap map[string]string, estbMacKey string, ecmMacKey string) {
	model := contextMap[common.MODEL]
	if model != "" {
		contextMap[common.MODEL] = strings.ToUpper(model)
	}
	env := contextMap[common.ENV]
	if env != "" {
		contextMap[common.ENV] = strings.ToUpper(env)
	}
	estbMac := contextMap[estbMacKey]
	if estbMac != "" {
		normalizedEstbMac, err := util.MacAddrComplexFormat(estbMac)
		if err == nil {
			contextMap[estbMacKey] = normalizedEstbMac
		}
	}
	ecmMac := contextMap[ecmMacKey]
	if ecmMac != "" {
		normalizedEcmMac, err := util.MacAddrComplexFormat(ecmMac)
		if err == nil {
			contextMap[ecmMacKey] = normalizedEcmMac
		}
	}
	partnerId := contextMap[common.PARTNER_ID]
	if partnerId != "" {
		contextMap[common.PARTNER_ID] = strings.ToUpper(partnerId)
	}
}

func AddContextFromTaggingService(ws *xhttp.XconfServer, contextMap map[string]string, satToken string, configSetHash string, isRfcApi bool, vargs ...log.Fields) {
	var fields log.Fields
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}
	if (isRfcApi && Xc.EnableTaggingServiceRFC) || (!isRfcApi && Xc.EnableTaggingService) {
		tags, err := ws.TaggingConnector.GetTagsForContext(contextMap, satToken, fields)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Debug("Error getting tags from tagging service")
		} else {
			for _, tag := range tags {
				contextMap[tag] = ""
			}
		}
	}
}

func GetPartnerFromAccountServiceByHostMac(ws *xhttp.XconfServer, macAddress string, satToken string, vargs ...log.Fields) string {
	var fields log.Fields
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}

	var partnerId string
	if Xc.EnableAccountService {
		var accountObject xhttp.AccountServiceDevices
		var err error
		if util.IsValidMacAddress(macAddress) {
			accountObject, err = ws.AccountServiceConnector.GetDevices(common.HOST_MAC_PARAM, macAddress, satToken, fields)
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Error getting account information")
		} else {
			partnerId = strings.ToUpper(accountObject.AccountServiceDeviceData.Partner)
		}
	}
	return partnerId
}

func GetApplicationTypeFromPartnerId(id string) string {
	if !util.IsBlank(id) && Xc.DeriveAppTypeFromPartnerId && len(Xc.PartnerApplicationTypes) > 0 {
		id = strings.ToLower(id)
		for _, appType := range Xc.PartnerApplicationTypes {
			if strings.HasPrefix(id, appType) {
				return appType
			}
		}
	}
	return ""
}
