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

	"xconfwebconfig/common"
	owcommon "xconfwebconfig/common"
	"xconfwebconfig/dataapi/featurecontrol"
	xhttp "xconfwebconfig/http"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/rfc"
	"xconfwebconfig/util"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func GetFeatureControlSettingsHandler(w http.ResponseWriter, r *http.Request) {
	// ==== log pre-processing ====
	var fields log.Fields
	if xw, ok := w.(*xhttp.XResponseWriter); ok {
		fields = xw.Audit()
	} else {
		xhttp.Error(w, http.StatusInternalServerError, owcommon.NotOK)
		return
	}

	applicationType, found := mux.Vars(r)[common.APPLICATION_TYPE]
	queryParams := r.URL.Query()
	xconfHttp := r.Header.Get(common.XCONF_HTTP_HEADER)
	configSetHash := r.Header.Get(common.CONFIG_SET_HASH)
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
	deviceServiceData := AddFeatureControlContext(Ws, r, contextMap, xconfHttp, configSetHash, fields)
	featureControlRuleBase := featurecontrol.NewFeatureControlRuleBase()
	featureControl := featureControlRuleBase.Eval(contextMap, contextMap[common.APPLICATION_TYPE], fields)
	isSecuredConnection := xconfHttp == ""
	PostProcessFeatureControl(Ws, featureControl, contextMap, isSecuredConnection, deviceServiceData)
	featureControlMap := &map[string]rfc.FeatureControl{
		"featureControl": *featureControl,
	}
	calculatedConfigSetHash := featureControlRuleBase.CalculateHash(featureControl.FeatureResponses)
	headers := map[string]string{
		common.CONFIG_SET_HASH: calculatedConfigSetHash,
	}
	status := http.StatusOK
	if configSetHash != "" && calculatedConfigSetHash == configSetHash {
		status = http.StatusNotModified // status code 304
	}
	response, _ := util.XConfJSONMarshal(featureControlMap, true)
	xhttp.WriteXconfResponseWithHeaders(w, headers, status, []byte(response))

}
