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

	"github.com/rdkcentral/xconfwebconfig/common"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func GetPreprocessFeatureControlSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var fields log.Fields
	if xw, ok := w.(*xhttp.XResponseWriter); ok {
		fields = xw.Audit()
	} else {
		xhttp.Error(w, http.StatusInternalServerError, common.NotOK)
		return
	}
	params := mux.Vars(r)
	mac := params["mac"]
	estbMac := strings.ToUpper(mac)
	contextMap := make(map[string]string)
	if estbMac != "" {
		normalizedEstbMac, err := util.MacAddrComplexFormat(estbMac)
		if err == nil {
			contextMap[common.ESTB_MAC_ADDRESS] = normalizedEstbMac
		} else {
			log.WithFields(fields).Warnf("Invalid MAC address format: %v", estbMac)
			xhttp.WriteXconfResponse(w, http.StatusBadRequest, []byte("invalid MAC address format"))
			return
		}
	}

	clientProtocolHeader := GetClientProtocolHeaderValue(r)

	isSecuredConnection := IsSecureConnection(clientProtocolHeader)
	var precookData *PrecookData
	precookData = getPrecookRfcData(Ws, contextMap, fields)

	featureControl := &rfc.FeatureControl{}

	var preprocessedRulesEngineResponse, preprocessedPostProcessingResponse *[]rfc.FeatureResponse
	if precookData == nil || len(precookData.RfcHash) == 0 {
		log.WithFields(fields).Infof("No precook data found")
		xhttp.WriteXconfResponse(w, http.StatusNotFound, []byte("No preprocessed featureControl data found"))
		return
	} else {
		preprocessedRulesEngineResponse = getPreprocessedRfcRulesEngineResponse(precookData.RfcRulesEngineHash, fields)
		if isSecuredConnection && precookData.RfcPostProcessingHash != "" {
			preprocessedPostProcessingResponse = getPreprocessedRfcPostProcessResponse(precookData.RfcPostProcessingHash, fields)
		}
	}

	if preprocessedRulesEngineResponse != nil {
		preprocessedResponseList := make([]rfc.FeatureResponse, 0, len(*preprocessedRulesEngineResponse))
		preprocessedResponseList = append(preprocessedResponseList, *preprocessedRulesEngineResponse...)
		featureControl.FeatureResponses = preprocessedResponseList
	}

	if preprocessedPostProcessingResponse != nil {
		featureControl.FeatureResponses = append(featureControl.FeatureResponses, *preprocessedPostProcessingResponse...)
	}

	featureControlMap := &map[string]rfc.FeatureControl{
		"featureControl": *featureControl,
	}
	response, _ := util.XConfJSONMarshal(featureControlMap, true)
	xhttp.WriteXconfResponseWithHeaders(w, nil, http.StatusOK, []byte(response))
}
