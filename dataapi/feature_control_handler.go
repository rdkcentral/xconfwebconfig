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
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi/featurecontrol"
	"github.com/rdkcentral/xconfwebconfig/db"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func GetFeatureControlSettingsHandler(w http.ResponseWriter, r *http.Request) {
	// ==== log pre-processing ====
	var fields log.Fields
	if xw, ok := w.(*xhttp.XResponseWriter); ok {
		fields = xw.Audit()
	} else {
		xhttp.Error(w, http.StatusInternalServerError, common.NotOK)
		return
	}

	applicationType, found := mux.Vars(r)[common.APPLICATION_TYPE]
	queryParams := r.URL.Query()

	elements := strings.Split(r.URL.String(), "?")
	var rfcQueryParams string
	if len(elements) > 1 {
		rfcQueryParams = elements[1]
	}
	fields["rfc_query_params"] = rfcQueryParams

	var skipPenetrationLogging bool
	hval := r.Header.Get(common.NoPenetrationMetricsHeader)
	if len(hval) > 0 && hval == "true" {
		skipPenetrationLogging = true
	}

	configSetHash := r.Header.Get(common.CONFIG_SET_HASH)
	fields["configsetHashDevice"] = configSetHash
	contextMap := make(map[string]string)
	if !found {
		applicationType = shared.STB
	}
	contextMap[common.APPLICATION_TYPE] = applicationType
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			contextMap[k] = v[0]
		}
	}

	NormalizeFeatureControlContext(Ws, r, contextMap, fields)
	fields["estbMacAddress"] = contextMap[common.ESTB_MAC_ADDRESS]
	podData, tags, AccountServiceData := AddFeatureControlContext(Ws, r, contextMap, configSetHash, fields)
	clientProtocolHeader := GetClientProtocolHeaderValue(r)
	AddClientProtocolToContextMap(contextMap, clientProtocolHeader)
	clientCertExpiry := GetClientCertExpiryHeaderValue(r)
	AddCertExpiryToContextMap(contextMap, clientCertExpiry)
	AddClientCertDurationToContext(contextMap, clientCertExpiry)

	ipInSameNetwork := true
	canPrecookRfcResponse := false
	isRfcPrecook304Enabled := false
	isRfcPrecookForOfferedFwEnabled := Xc.EnableRfcPrecookForOfferedFw
	isPrecookLockdownMode := shared.GetBooleanAppSetting(common.PROP_PRECOOK_LOCKDOWN_ENABLED, false)
	tfields := common.FilterLogFields(fields)
	// only check values of precook flags if not in precook lockdown mode and mac is not in exclusion
	if isPrecookLockdownMode {
		log.WithFields(tfields).Debug("Currently in pre-cook lockdown mode, setting pre-cook flags to false.")
	} else {
		exclusionMacsSet, _ := shared.GetGenericNamedListSetByType(shared.MAC_LIST)
		if exclusionMacsSet.Contains(contextMap[common.ESTB_MAC_ADDRESS]) {
			log.WithFields(tfields).Debugf("Device mac %s is in precook exclusion list, will not deliver precook data.", contextMap[common.ESTB_MAC_ADDRESS])
			xhttp.IncreasePrecookExcludeMacListCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
		} else {
			log.WithFields(tfields).Debug("Currently not in pre-cook lockdown mode and mac not in precook exclusion list, getting current pre-cook flags from config.")
			canPrecookRfcResponse = canPrecookRfcResponses(tfields)
			isRfcPrecook304Enabled = Xc.EnableRfcPrecook304
		}
	}

	var precookData *PrecookData
	// we need to check the current reported firmware version against the ones in precook data.
	isFwVersionMatched := false
	isOfferedFwMatched := false

	// if any of these precook flags are on, we'll need precook data from XDAS
	if canPrecookRfcResponse || isRfcPrecook304Enabled {
		precookData = getPrecookRfcData(Ws, contextMap, fields)
		tfields = common.FilterLogFields(fields)
		if precookData == nil || len(precookData.RfcHash) == 0 {
			xhttp.IncreaseNoPrecookDataCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
			log.WithFields(tfields).Infof("No precook data found")
		} else {
			ctxHash := precookData.CtxHash
			var err error
			contextHashMatched, err := CompareHashWithXDAS(contextMap, ctxHash, tags)
			if err != nil {
				log.WithFields(tfields).Errorf("Error while comparing hashes: %v", err)
				contextHashMatched = false
			}

			if !contextHashMatched {
				xhttp.IncreaseCtxHashMismatchCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
				generatePrecookDataChangedMetrics(contextMap, precookData, fields)
			}

			//check if the incoming ip address is in the same network as the precook data's Ip address based on the subnet mask
			var networkMask net.IPMask
			isIPv4 := util.IsIPv4(net.ParseIP(contextMap[common.ESTB_IP]))
			var networkMaskPrefixLength int32
			if isIPv4 {
				networkMaskPrefixLength = Xc.IPv4NetworkMaskPrefixLength
				networkMask = util.Ipv4NetworkMask(int(networkMaskPrefixLength))
			} else {
				networkMaskPrefixLength = Xc.IPv6NetworkMaskPrefixLength
				networkMask = util.Ipv6NetworkMask(int(networkMaskPrefixLength))
			}
			ipInSameNetwork, err = util.IsInSameNetwork(contextMap[common.ESTB_IP], precookData.EstbIp, networkMask)
			if err != nil {
				log.WithFields(tfields).Errorf("Error while comparing IP addresses: %v", err)
				ipInSameNetwork = false
			}
			if !ipInSameNetwork {
				log.WithFields(tfields).Debugf("IP address %s is not in the same network as precook data's IP address %s", contextMap[common.ESTB_IP], precookData.EstbIp)
				xhttp.IncreaseIpNotInSameNetworkCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
			}

			if contextMap[common.FIRMWARE_VERSION] == precookData.FwVersion {
				isFwVersionMatched = true
			} else if isRfcPrecookForOfferedFwEnabled && (contextMap[common.FIRMWARE_VERSION] == precookData.OfferedFwVersion) {
				isOfferedFwMatched = true
				xhttp.IncreaseOfferedFwVersionMatchCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
			} else if isRfcPrecookForOfferedFwEnabled {
				log.WithFields(tfields).Debugf("Firmware version %s does not match precook data's firmware version %s or offered firmware version %s", contextMap[common.FIRMWARE_VERSION], precookData.FwVersion, precookData.OfferedFwVersion)
				xhttp.IncreaseFirmwareVersionMismatchCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
			}

			if contextHashMatched && ipInSameNetwork && (isFwVersionMatched || (isRfcPrecookForOfferedFwEnabled && isOfferedFwMatched)) {
				log.WithFields(tfields).Debug("ContextHash match with XDAS's ctxHash, IP addresses are in the same network and FW matched, so delivery precook Data.")
			} else {
				log.WithFields(tfields).Debugf("NOT deliver precook data, ContextHashMatched: %v, IP inSameNetwork: %v, isFwVersionMatched: %v, isRfcPrecookForOfferedFwEnabled: %v, isOfferedFwMatched: %v.", contextHashMatched, ipInSameNetwork, isFwVersionMatched, isRfcPrecookForOfferedFwEnabled, isOfferedFwMatched)
				// disable precook flags
				canPrecookRfcResponse = false
				isRfcPrecook304Enabled = false
			}
		}
	}
	if isRfcPrecook304Enabled {
		// if configsetHash from device matches precook, return 304 without running rules engine
		if matchedHash := getMatchedPrecookHash(configSetHash, precookData, isRfcPrecookForOfferedFwEnabled); matchedHash != "" {
			xhttp.IncreaseReturn304FromPrecookCounter(common.PARTNER_ID, contextMap[common.MODEL])
			tfields := common.FilterLogFields(fields)
			tfields["isLiveCalculated"] = false
			log.WithFields(tfields).Info("Returning 304 based on precook data")
			headers := map[string]string{
				common.CONFIG_SET_HASH: matchedHash,
			}
			if Ws.Config.GetBoolean("xconfwebconfig.xconf.enable_rfc_penetration_metrics", false) && !skipPenetrationLogging {
				copyFields := common.CopyLogFields(fields)
				go UpdatePenetrationMetrics(contextMap, AccountServiceData, tags, "", copyFields, true)
			}
			xhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusNotModified, []byte(""))
			return
		}
	}

	var precookRulesEngineResponse, precookPostProcessingResponse *[]rfc.FeatureResponse
	var rfcPostProc string
	featureControlRuleBase := featurecontrol.NewFeatureControlRuleBase()

	isSecuredConnection := IsSecureConnection(clientProtocolHeader)
	// return response from XPC precook table if possible
	if canPrecookRfcResponse && precookData != nil {
		if isFwVersionMatched {
			precookRulesEngineResponse = getPrecookRfcRulesEngineResponse(precookData.RfcRulesEngineHash, fields)
		} else if isRfcPrecookForOfferedFwEnabled && isOfferedFwMatched {
			log.Debugf("Using offered firmware version for precook rules engine response, OfferedFwRfcRulesEngineHash: %v, offeredFwVersion: %v", precookData.OfferedFwRfcRulesEngineHash, precookData.OfferedFwVersion)
			precookRulesEngineResponse = getPrecookRfcRulesEngineResponse(precookData.OfferedFwRfcRulesEngineHash, fields)
		}

		// ensure isSecureConnection is true first, if not, need to build on the fly so we don't expose accountId info
		if isSecuredConnection {
			precookPostProcessingResponse = getPrecookRfcPostProcessResponse(precookData.RfcPostProcessingHash, fields)
		}
	}
	featureControl := &rfc.FeatureControl{}
	appliedFeatureRules := []*rfc.FeatureRule{}
	// if using precook rules engine response, else need to run rules engine
	if precookRulesEngineResponse != nil && precookData != nil {
		xhttp.IncreaseReturn200FromPrecookCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
		if isFwVersionMatched {
			fields["configsetHashRulesEngine"] = precookData.RfcRulesEngineHash
		} else if isRfcPrecookForOfferedFwEnabled && isOfferedFwMatched {
			fields["configsetHashRulesEngine"] = precookData.OfferedFwRfcRulesEngineHash
		}
		log.WithFields(common.FilterLogFields(fields)).Info("Returning precook response")
		precookResponseList := make([]rfc.FeatureResponse, 0, len(*precookRulesEngineResponse))
		precookResponseList = append(precookResponseList, *precookRulesEngineResponse...)
		featureControl.FeatureResponses = precookResponseList
	} else {
		featureControl, appliedFeatureRules = featureControlRuleBase.Eval(contextMap, contextMap[common.APPLICATION_TYPE], fields)
		// calculate hashes on rules engine response
		rulesEngineConfigsetHash := featureControlRuleBase.CalculateHash(featureControl.FeatureResponses)
		fields["configsetHashRulesEngine"] = rulesEngineConfigsetHash
	}
	// if using precook post-processing response,
	if precookPostProcessingResponse != nil && precookData != nil {
		xhttp.IncreaseReturnPostProcessFromPrecookCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
		fields["configsetHashPostProcess"] = precookData.RfcPostProcessingHash
		featureControl.FeatureResponses = append(featureControl.FeatureResponses, *precookPostProcessingResponse...)
	} else {
		// XPC-18973 prepare post processing response for storage
		extraFeatureResponses := PostProcessFeatureControl(Ws, contextMap, isSecuredConnection, podData)
		// calculate hashes on post-processing response
		postProcessConfigsetHash := featureControlRuleBase.CalculateHash(extraFeatureResponses)
		fields["configsetHashPostProcess"] = postProcessConfigsetHash
		// add post-process response to featureControl object
		if len(extraFeatureResponses) > 0 {
			featureControl.FeatureResponses = append(featureControl.FeatureResponses, extraFeatureResponses...)
			if bbytes, err := json.Marshal(extraFeatureResponses); err == nil {
				rfcPostProc = string(bbytes)
			}
		}
	}

	// XPC-12321
	calculatedConfigSetHash := featureControlRuleBase.CalculateHash(featureControl.FeatureResponses)
	fields["configsetHashCalculated"] = calculatedConfigSetHash

	isLiveCalculated := precookData == nil || precookRulesEngineResponse == nil
	featureControlRuleBase.LogFeatureInfo(contextMap, appliedFeatureRules, featureControl.FeatureResponses, isLiveCalculated, fields)

	if Ws.Config.GetBoolean("xconfwebconfig.xconf.enable_rfc_penetration_metrics", false) {
		if !skipPenetrationLogging {
			copyFields := common.CopyLogFields(fields)
			go UpdatePenetrationMetrics(contextMap, AccountServiceData, tags, rfcPostProc, copyFields, false)
		}
	}

	headers := map[string]string{
		common.CONFIG_SET_HASH: calculatedConfigSetHash,
	}
	// if device configsethash matches the one we calculate, return 304 with no body
	if configSetHash != "" && calculatedConfigSetHash == configSetHash {
		xhttp.IncreaseReturn304RulesEngineCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
		xhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusNotModified, []byte(""))
		return
	}
	// if we get to this point, we know we didn't return a 304, but we need to know if we're returning 200 from the rules engine or from precook
	if precookRulesEngineResponse == nil {
		xhttp.IncreaseReturn200RulesEngineCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
		if !ipInSameNetwork {
			xhttp.IncreaseIpNotInSameNetworkIn200Counter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
		}
		if precookData != nil {
			generatePrecookDataChangedIn200Metrics(contextMap, precookData, fields)
		}
	}

	if precookPostProcessingResponse == nil {
		xhttp.IncreaseReturnPostProcessOnTheFlyCounter(contextMap[common.PARTNER_ID], contextMap[common.MODEL])
	}

	featureControlMap := &map[string]rfc.FeatureControl{
		"featureControl": *featureControl,
	}
	response, _ := util.XConfJSONMarshal(featureControlMap, true)
	xhttp.WriteXconfResponseWithHeaders(w, headers, http.StatusOK, []byte(response))
}

func UpdatePenetrationMetrics(context map[string]string, AccountServiceData *AccountServiceData, taglist []string, rfcPostProc string, fields log.Fields, is304FromPrecook bool) {
	// if we're using precook data, the following fields will be empty and will write empty string
	// to penetration table: appliedRules, features, configsetHashCalculatedruleNames := []string{}
	ruleNames := []string{}
	if itf, ok := fields["appliedRules"]; ok {
		if items, ok := itf.([]string); ok {
			ruleNames = items
		}
	}

	featureInstances := []string{}
	if itf, ok := fields["features"]; ok {
		if items, ok := itf.([]string); ok {
			featureInstances = items
		}
	}

	var rfcQueryParams, calculatedConfigSetHash string
	if itf, ok := fields["rfc_query_params"]; ok {
		rfcQueryParams = itf.(string)
	}

	if itf, ok := fields["configsetHashCalculated"]; ok {
		calculatedConfigSetHash = itf.(string)
	}

	var tags string
	if taglist != nil {
		tags = strings.Join(taglist, ",")
	}

	estbMac := context[common.ESTB_MAC_ADDRESS]
	ecmMac := context[common.ECM_MAC_ADDRESS]
	if len(strings.TrimSpace(estbMac)) > 0 {
		rfcTs := time.Now().UnixNano() / 1000000
		// sort arrays so same order in Penetration Data Table
		// but create copy first so featureControl response is unchanged
		sortedRules := featurecontrol.SortCaseInsensitive(ruleNames)
		sortedFeatures := featurecontrol.SortCaseInsensitive(featureInstances)
		pTable := &db.RfcPenetrationMetrics{
			EstbMac:              estbMac,
			EcmMac:               ecmMac,
			SerialNum:            context[common.SERIAL_NUM],
			Partner:              context[common.PARTNER_ID],
			Model:                context[common.MODEL],
			RfcPartner:           context[common.PARTNER_ID],
			RfcModel:             context[common.MODEL],
			RfcFwReportedVersion: context[common.FIRMWARE_VERSION],
			RfcAppliedRules:      strings.Join(sortedRules, ","),
			RfcFeatures:          strings.Join(sortedFeatures, ","),
			RfcTs:                rfcTs,
			RfcAccountHash:       context[common.ACCOUNT_HASH],
			RfcAccountId:         context[common.ACCOUNT_ID],
			RfcAccountMgmt:       context[common.ACCOUNT_MGMT],
			RfcEnv:               context[common.ENV],
			RfcApplicationType:   context[common.APPLICATION_TYPE],
			RfcExperience:        context[common.EXPERIENCE],
			RfcConfigsetHash:     calculatedConfigSetHash,
			RfcQueryParams:       rfcQueryParams,
			RfcTags:              tags,
			RfcEstbIp:            context[common.ESTB_IP],
			ClientCertExpiry:     context[common.CLIENT_CERT_EXPIRY],
			RecoveryCertExpiry:   context[common.RECOVERY_CERT_EXPIRY],
			RfcPostProc:          rfcPostProc,
		}
		if AccountServiceData != nil {
			pTable.RfcTimeZone = AccountServiceData.TimeZone
			pTable.TitanPartner = AccountServiceData.PartnerId
			pTable.TitanAccountId = AccountServiceData.AccountId
		}
		err := db.GetDatabaseClient().SetRfcPenetrationMetrics(pTable, is304FromPrecook)
		if err != nil {
			log.Error(fmt.Sprintf("Can't save RFC penetration metric, estbMacAddress=%s, error=%+v", estbMac, err))
		}
	} else {
		log.Debugf("No estbMacAddress provided, NOT writing to xconf penetration metrics table.")
	}
}

func getMatchedPrecookHash(configSetHash string, precookData *PrecookData, isRfcPrecookForOfferedFwEnabled bool) string {
	if precookData == nil || configSetHash == "" {
		return ""
	}
	if string(precookData.RfcHash) == configSetHash {
		return precookData.RfcHash
	}
	if isRfcPrecookForOfferedFwEnabled && string(precookData.OfferedFwRfcHash) == configSetHash {
		return precookData.OfferedFwRfcHash
	}
	return ""
}
