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
	"net/http"
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

func GetClientProtocolHeaderValue(r *http.Request) string {
	return r.Header.Get(common.XCONF_HTTP_HEADER)
}

func AddClientProtocolToContextMap(contextMap map[string]string, clientProtocolHeader string) {
	switch clientProtocolHeader {
	case common.XCONF_HTTPS_VALUE:
		contextMap[common.CLIENT_PROTOCOL] = common.HTTPS_CLIENT_PROTOCOL
	case common.XCONF_MTLS_VALUE:
		contextMap[common.CLIENT_PROTOCOL] = common.MTLS_CLIENT_PROTOCOL
	case common.XCONF_MTLS_RECOVERY_VALUE:
		contextMap[common.CLIENT_PROTOCOL] = common.MTLS_RECOVERY_CLIENT_PROTOCOL
	default:
		contextMap[common.CLIENT_PROTOCOL] = common.HTTP_CLIENT_PROTOCOL
	}
}

func IsSecureConnection(clientProtocolHeader string) bool {
	if clientProtocolHeader == common.XCONF_HTTPS_VALUE || clientProtocolHeader == common.XCONF_MTLS_VALUE || clientProtocolHeader == common.XCONF_MTLS_RECOVERY_VALUE {
		return true
	}
	return false
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

func AddGroupServiceContext(ws *xhttp.XconfServer, contextMap map[string]string, macKeyName string, fields log.Fields) {
	if Xc.EnableGroupService {
		if Xc.GroupServiceModelSet.Contains(strings.ToUpper(contextMap[common.MODEL])) {
			log.WithFields(fields).Debugf("Getting groups from Group Service for mac=%s, model=%s.", contextMap[macKeyName], contextMap[common.MODEL])
			// remove colons before calling Group Service
			macAddress := util.AlphaNumericMacAddress(contextMap[macKeyName])
			cpeGroups, err := ws.GroupServiceConnector.GetCpeGroups(macAddress, fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Debugf("Error getting groups from Group Service for mac=%s, model=%s.", contextMap[macKeyName], contextMap[common.MODEL])
				return
			}
			for _, group := range cpeGroups {
				contextMap[group] = ""
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
			partnerId = strings.ToUpper(accountObject.DeviceData.Partner)
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
