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
	"strconv"
	"strings"
	"time"

	common "github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/util"

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

func GetClientCertExpiryHeaderValue(r *http.Request) string {
	return r.Header.Get(common.CLIENT_CERT_EXPIRY_HEADER)
}

func AddClientProtocolToContextMap(contextMap map[string]string, clientProtocolHeader string) {
	switch clientProtocolHeader {
	case common.XCONF_HTTPS_VALUE:
		contextMap[common.CLIENT_PROTOCOL] = common.HTTPS_CLIENT_PROTOCOL
	case common.XCONF_MTLS_VALUE:
		contextMap[common.CLIENT_PROTOCOL] = common.MTLS_CLIENT_PROTOCOL
	case common.XCONF_MTLS_OPTIONAL_VALUE:
		contextMap[common.CLIENT_PROTOCOL] = common.MTLS_OPTIONAL_CLIENT_PROTOCOL
	case common.XCONF_MTLS_RECOVERY_VALUE:
		contextMap[common.CLIENT_PROTOCOL] = common.MTLS_RECOVERY_CLIENT_PROTOCOL
	default:
		contextMap[common.CLIENT_PROTOCOL] = common.HTTP_CLIENT_PROTOCOL
	}
}

func AddCertExpiryToContextMap(contextMap map[string]string, clientCertExpiry string) {
	if !util.IsBlank(clientCertExpiry) {
		if contextMap[common.CLIENT_PROTOCOL] == common.MTLS_RECOVERY_CLIENT_PROTOCOL {
			contextMap[common.RECOVERY_CERT_EXPIRY] = clientCertExpiry
		} else {
			contextMap[common.CLIENT_CERT_EXPIRY] = clientCertExpiry
		}
	}
}

func IsSecureConnection(clientProtocolHeader string) bool {
	return clientProtocolHeader == common.XCONF_HTTPS_VALUE || clientProtocolHeader == common.XCONF_MTLS_VALUE || clientProtocolHeader == common.XCONF_MTLS_RECOVERY_VALUE || clientProtocolHeader == common.XCONF_MTLS_OPTIONAL_VALUE
}

func AddClientCertDurationToContext(contextMap map[string]string, clientCertExpiryStr string) {
	if util.IsBlank(clientCertExpiryStr) {
		return
	}
	certExpiryDate, err := time.Parse(common.ClientCertExpiryDateFormat, clientCertExpiryStr)
	if err != nil {
		log.Debugf("wrong data format: %s for string value: %s", common.ClientCertExpiryDateFormat, clientCertExpiryStr)
		return
	}
	nowDate := time.Now().UTC()
	duration := certExpiryDate.Sub(nowDate)
	durationDays := int(duration.Hours() / 24)
	contextMap[common.CERT_EXPIRY_DAYS] = strconv.Itoa(durationDays)
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

func AddContextFromTaggingService(ws *xhttp.XconfServer, contextMap map[string]string, satToken string, configSetHash string, isRfcApi bool, vargs ...log.Fields) []string {
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
			return nil
		} else {
			for _, tag := range tags {
				contextMap[tag] = ""
			}
			return tags
		}
	}
	return nil
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

func RemovePrefix(key string, prefixSet []string) (string, bool) {
	for _, prefix := range prefixSet {
		if strings.HasPrefix(key, prefix) {
			keyWithoutPrefix, _ := strings.CutPrefix(key, prefix)
			if !util.IsBlank(keyWithoutPrefix) {
				return keyWithoutPrefix, true
			}
		}
	}
	return key, false
}

func AddGroupServiceContext(ws *xhttp.XconfServer, contextMap map[string]string, macKeyName string, fields log.Fields) {
	if Xc.EnableGroupService {
		macAddress := util.AlphaNumericMacAddress(contextMap[macKeyName])
		if Xc.GroupServiceModelSet.Contains(strings.ToUpper(contextMap[common.MODEL])) && !util.IsBlank(macAddress) {
			log.WithFields(common.FilterLogFields(fields)).Debugf("Getting groups from Group Service  for mac=%s, model=%s.", contextMap[macKeyName], contextMap[common.MODEL])
			// remove colons before calling GroupService
			CpeGroups, err := ws.GroupServiceConnector.GetCpeGroups(macAddress, fields)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Debugf("Error getting groups from Group Service for mac=%s, model=%s.", contextMap[macKeyName], contextMap[common.MODEL])
				return
			}
			for _, group := range CpeGroups {
				contextMap[group] = ""
			}
		}
	}
}

func AddGroupServiceFTContext(ws *xhttp.XconfServer, macAddressKey string, contextMap map[string]string, checkForGroups bool, fields log.Fields) []string {
	var groupServiceDao = db.GetGroupServiceCacheDao()
	var tags []string
	tfields := common.FilterLogFields(fields)

	if mac, ok := contextMap[macAddressKey]; ok {
		if Xc.EnableFtMacTags && (len(Xc.MacTagsModelSet) == 0 || Xc.MacTagsModelSet.Contains(strings.ToUpper(contextMap[common.MODEL]))) {
			convertedMacAddress := util.GetEcmMacAddress(util.AlphaNumericMacAddress(strings.TrimSpace(mac)))
			log.WithFields(tfields).Debugf("Getting all data from from GroupService /ft keyspace for estbMac=%s, ecmMac=%s", mac, convertedMacAddress)
			ftTags := AddGroupServiceFeatureTags(ws, convertedMacAddress, contextMap, true, true, Xc.MacTagsPrefixList, fields)
			tags = append(tags, ftTags...)
		} else if checkForGroups && Xc.EnableFtGroups && (len(Xc.GroupServiceModelSet) == 0 || Xc.GroupServiceModelSet.Contains(strings.ToUpper(contextMap[common.MODEL]))) {
			convertedMacAddress := util.GetEcmMacAddress(util.AlphaNumericMacAddress(strings.TrimSpace(mac)))
			log.WithFields(tfields).Debugf("Getting group data from GroupService /ft keyspace for estbMac=%s, ecmMac=%s", mac, convertedMacAddress)
			AddGroupServiceFeatureTags(ws, convertedMacAddress, contextMap, true, false, Xc.MacTagsPrefixList, fields)
		}
	}
	if account, ok := contextMap[common.ACCOUNT_ID]; ok {
		if Xc.EnableFtAccountTags && (len(Xc.AccountTagsModelSet) == 0 || Xc.AccountTagsModelSet.Contains(strings.ToUpper(contextMap[common.MODEL]))) {
			account = strings.ToUpper(strings.TrimSpace(account))
			log.WithFields(tfields).Debugf("Getting all data from GroupService /ft keyspace for accountId=%s", account)
			ftTags := AddGroupServiceFeatureTags(ws, account, contextMap, false, true, Xc.AccountTagsPrefixList, fields)
			tags = append(tags, ftTags...)
		}
	}
	if partner, ok := contextMap[common.PARTNER_ID]; ok {
		getPrefixData := true
		if Xc.EnableFtPartnerTags && (len(Xc.PartnerTagsModelSet) == 0 || Xc.PartnerTagsModelSet.Contains(strings.ToUpper(contextMap[common.MODEL]))) {
			partner = strings.TrimSpace(partner)

			if Xc.PartnerIdValidationEnabled && !Xc.ValidPartnerIdRegex.MatchString(partner) {
				log.WithFields(fields).Infof("Skipping AddGroupServiceFTContext for invalid partnerId: %q", partner)
				return tags
			}

			log.WithFields(fields).Debugf("Getting all data from GroupService /ft keyspace for partnerId=%s", partner)

			if Xc.GroupServiceCacheEnabled {
				Tags := groupServiceDao.GetGroupServiceFeatureTags(partner)
				for key, value := range Tags {
					if keyWithoutPrefix, ok := RemovePrefix(key, Xc.PartnerTagsPrefixList); ok {
						if getPrefixData {
							contextMap[keyWithoutPrefix] = value
							tags = append(tags, fmt.Sprintf("%s#%s", keyWithoutPrefix, value))
						}
					}
				}
				log.WithFields(log.Fields{"partnerId": partner, "fields": fields, "contextMap": contextMap, "tags": tags}).Debug("Cache hit")

			} else {
				ftTags := AddGroupServiceFeatureTags(ws, partner, contextMap, false, true, Xc.PartnerTagsPrefixList, fields)
				tags = append(tags, ftTags...)
			}
		}
	}
	return tags
}

// CompareTaggingSources compares COAST tags with XConf tags and logs differences
// to track migration progress from COAST to XConf tagging service.
// Only logs when COAST has tags not present in XConf (indicates incomplete migration).
func CompareTaggingSources(contextMap map[string]string, coastTags []string, xconfTags []string, fields log.Fields) {
	if !Xc.EnableTaggingComparison {
		return
	}
	if len(coastTags) == 0 {
		return
	}

	xconfTagSet := make(map[string]struct{})
	for _, tag := range xconfTags {
		xconfTagSet[tag] = struct{}{}
	}

	var missingTags []string
	for _, coastTag := range coastTags {
		if _, exists := xconfTagSet[coastTag]; !exists {
			missingTags = append(missingTags, coastTag)
		}
	}

	// Log only when missing tags found
	if len(missingTags) > 0 {
		estbMac := contextMap[common.ESTB_MAC_ADDRESS]
		if estbMac == "" {
			estbMac = contextMap[common.ESTB_MAC]
		}

		log.WithFields(log.Fields{
			"estbMac":      estbMac,
			"partner":      contextMap[common.PARTNER_ID],
			"model":        contextMap[common.MODEL],
			"missingTags":  strings.Join(missingTags, ","),
			"missingCount": len(missingTags),
		}).Warn("COAST tags missing in XConf tagging service")
	}
}

func AddGroupServiceFeatureTags(ws *xhttp.XconfServer, groupName string, contextMap map[string]string, getGroupsData bool, getPrefixData bool, prefixList []string, fields log.Fields) []string {
	featureTags, err := ws.GroupServiceConnector.GetFeatureTagsHashedItems(groupName, fields)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debugf("Error getting response from GroupService for /ft keyspace for mac=%s", contextMap[common.ESTB_MAC])
		return nil
	}
	var tags []string
	for key, value := range featureTags {
		if keyWithoutPrefix, ok := RemovePrefix(key, prefixList); ok {
			if getPrefixData {
				contextMap[keyWithoutPrefix] = value
				tags = append(tags, fmt.Sprintf("%s#%s", keyWithoutPrefix, value))
			}
			continue
		}
		if getGroupsData && value == "1" {
			key = fmt.Sprintf("%s%s", ws.GroupServiceConnector.GroupPrefix(), key)
			contextMap[key] = ""
			tags = append(tags, key)
		}
	}
	return tags
}
