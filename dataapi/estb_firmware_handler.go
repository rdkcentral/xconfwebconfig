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

	"xconfwebconfig/common"
	dataef "xconfwebconfig/dataapi/estbfirmware"
	xhttp "xconfwebconfig/http"
	"xconfwebconfig/shared"
	sharedef "xconfwebconfig/shared/estbfirmware"
	"xconfwebconfig/util"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func GetEstbFirmwareSwuBseHandler(w http.ResponseWriter, r *http.Request) {
	xw, ok := w.(*xhttp.XResponseWriter)
	if !ok {
		xhttp.Error(w, http.StatusInternalServerError, common.NotOK)
		return
	}
	var isIpAddressPresent bool
	var ipAddress string
	var ips []string
	queryParams := r.URL.Query()
	if len(queryParams) > 0 {
		ips, isIpAddressPresent = queryParams[common.IP_ADDRESS]
		ipAddress = ips[0]
	}
	if !isIpAddressPresent {
		if r.ContentLength != 0 {
			contextMap := make(map[string]string)
			body := xw.Body()
			if body != "" {
				parseProcBody(body, contextMap)
			}
			ipAddress, isIpAddressPresent = contextMap[common.IP_ADDRESS]
		}
	}
	if !isIpAddressPresent {
		xhttp.WriteXconfResponseAsText(w, 400, []byte(fmt.Sprintf("Required IpAddress parameter '%s' is not present", common.IP_ADDRESS)))
		return
	}
	ip := shared.NewIpAddress(ipAddress)
	if ip == nil {
		xhttp.WriteXconfResponseAsText(w, 400, []byte(fmt.Sprintf("Required IpAddress value: '%s' is not a valid IpAddress", ipAddress)))
		return
	}
	estbFirmwareRuleBase := dataef.NewEstbFirmwareRuleBaseDefault()
	bseConfiguration, _ := estbFirmwareRuleBase.GetBseConfiguration(ip)
	if bseConfiguration == nil {
		xhttp.WriteXconfResponseAsText(w, 404, []byte("\"<h2>404 NOT FOUND</h2>\""))
		return
	}
	response, _ := util.JSONMarshal(*bseConfiguration)
	xhttp.WriteXconfResponse(w, 200, response)

}

func GetEstbFirmwareSwuHandler(w http.ResponseWriter, r *http.Request) {
	// ==== log pre-processing ====
	var fields log.Fields
	xw, ok := w.(*xhttp.XResponseWriter)
	if ok {
		fields = xw.Audit()
	} else {
		xhttp.Error(w, http.StatusInternalServerError, common.NotOK)
		return
	}

	queryParams := r.URL.Query()
	xconfHttp := r.Header.Get(common.XCONF_HTTP_HEADER)
	contextMap := make(map[string]string)
	var version string
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			if k == common.VERSION {
				version = v[0]
			} else {
				contextMap[k] = strings.Join(v, ",")
			}
		}
	}
	if r.ContentLength != 0 {
		// body, _ := fields["body"].(string)
		body := xw.Body()
		if body != "" {
			version = parseProcBody(body, contextMap)
		}
	}
	contextMap[common.APPLICATION_TYPE] = mux.Vars(r)[common.APPLICATION_TYPE]
	contextMap[common.XCONF_HTTP_HEADER] = xconfHttp
	GetFirstElementsInContextMap(contextMap)
	if contextMap[common.ESTB_MAC] == "" {
		if util.IsVersionGreaterOrEqual(version, 2.0) {
			xhttp.WriteXconfResponseAsText(w, 400, []byte("\"eStbMac should be specified\""))
		} else {
			xhttp.WriteXconfResponseAsText(w, 500, []byte("\"eStbMac should be specified\""))
		}
	} else if !IsSecureConnection(Ws, contextMap) {
		xhttp.WriteXconfResponseAsText(w, 403, []byte("FORBIDDEN"))
	} else {
		_, errmac := util.MACAddressValidator(contextMap[common.ESTB_MAC])
		if errmac != nil {
			xhttp.WriteXconfResponseAsText(w, 500, []byte(fmt.Sprintf("\"<h2>500 Internal Server Error</h2><div>invalid mac address: %s</div>\"", contextMap[common.ESTB_MAC])))
			return
		}

		log.Debugf("GetEstbFirmwareSwuHandler call AddEstbFirmwareContext start ... queryParams %v", queryParams)
		AddEstbFirmwareContext(Ws, r, contextMap, true, true, fields)
		log.Debugf("GetEstbFirmwareSwuHandler call AddEstbFirmwareContext  ... end contextMap %v", contextMap)
		estbFirmwareRuleBase := dataef.NewEstbFirmwareRuleBaseDefault()
		convertedContext := sharedef.GetContextConverted(contextMap)
		evaluationResult, _ := estbFirmwareRuleBase.Eval(contextMap, convertedContext, contextMap[common.APPLICATION_TYPE], fields)
		explanation := GetExplanation(contextMap, evaluationResult)
		if evaluationResult == nil || evaluationResult.Blocked || evaluationResult.FirmwareConfig == nil || evaluationResult.FirmwareConfig.Properties == nil {
			LogResponse(contextMap, convertedContext, explanation, evaluationResult, fields)
			xhttp.WriteXconfResponseAsText(w, 404, []byte(fmt.Sprintf("\"<h2>404 NOT FOUND</h2><div>%s<div>\"", explanation)))
		} else {
			LogResponse(contextMap, convertedContext, explanation, evaluationResult, fields)
			firmwareConfigResponse := sharedef.CreateFirmwareConfigFacadeResponse(*evaluationResult.FirmwareConfig)
			response, _ := util.JSONMarshal(firmwareConfigResponse)
			xhttp.WriteXconfResponse(w, 200, response)
		}
	}
}

func GetCheckMinFirmwareHandler(w http.ResponseWriter, r *http.Request) {
	// ==== log pre-processing ====
	var fields log.Fields
	xw, ok := w.(*xhttp.XResponseWriter)
	if ok {
		fields = xw.Audit()
	} else {
		xhttp.Error(w, http.StatusInternalServerError, common.NotOK)
		return
	}

	queryParams := r.URL.Query()
	contextMap := make(map[string]string)
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			contextMap[k] = strings.Join(v, ",")
		}
	}
	if r.ContentLength != 0 {
		// body, _ := fields["body"].(string)
		body := xw.Body()
		if body != "" {
			parseProcBody(body, contextMap)
		}
	}
	GetFirstElementsInContextMap(contextMap)
	missingFields := []string{}
	emptyFields := []string{}
	GetMissingAndEmptyQueryParams(contextMap, &missingFields, &emptyFields)
	if len(missingFields) > 0 {
		xhttp.WriteXconfResponseAsText(w, 400, []byte(fmt.Sprintf("\"Required field(s) are missing: [%s]\"", strings.Join(missingFields, ", "))))
	} else if len(emptyFields) > 0 {
		log.Warn(fmt.Sprintf("Missing fields: %+v, returning hasMinimumFirmware as true.", emptyFields))
		minimumFirmwareCheckBean := &sharedef.MinimumFirmwareCheckBean{
			HasMinimumFirmware: true,
		}
		response, _ := util.JSONMarshal(*minimumFirmwareCheckBean)
		xhttp.WriteXconfResponse(w, 200, response)
	} else {
		AddEstbFirmwareContext(Ws, r, contextMap, false, false, fields)
		estbFirmwareRuleBase := dataef.NewEstbFirmwareRuleBaseDefault()
		hasMinimumFirmware := estbFirmwareRuleBase.HasMinimumFirmware(contextMap)
		minimumFirmwareCheckBean := &sharedef.MinimumFirmwareCheckBean{
			HasMinimumFirmware: hasMinimumFirmware,
		}
		response, _ := util.JSONMarshal(*minimumFirmwareCheckBean)
		xhttp.WriteXconfResponse(w, 200, response)
	}
}

func GetEstbFirmwareVersionInfoPath(w http.ResponseWriter, r *http.Request) {
	// ==== log pre-processing ====
	var fields log.Fields
	xw, ok := w.(*xhttp.XResponseWriter)
	if ok {
		fields = xw.Audit()
	} else {
		xhttp.Error(w, http.StatusInternalServerError, common.NotOK)
		return
	}

	queryParams := r.URL.Query()
	xconfHttp := r.Header.Get(common.XCONF_HTTP_HEADER)
	contextMap := make(map[string]string)
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			contextMap[k] = strings.Join(v, ",")
		}
	}
	if r.ContentLength != 0 {
		// body, _ := fields["body"].(string)
		body := xw.Body()
		if body != "" {
			parseProcBody(body, contextMap)
		}
	}
	contextMap[common.APPLICATION_TYPE] = mux.Vars(r)[common.APPLICATION_TYPE]
	contextMap[common.XCONF_HTTP_HEADER] = xconfHttp
	GetFirstElementsInContextMap(contextMap)
	if contextMap[common.ESTB_MAC] == "" {
		xhttp.WriteXconfResponseAsText(w, 400, []byte("eStbMac should be specified"))
	} else if !IsSecureConnection(Ws, contextMap) {
		xhttp.WriteXconfResponseAsText(w, 403, []byte("FORBIDDEN"))
	} else {
		AddEstbFirmwareContext(Ws, r, contextMap, true, true, fields)
		estbFirmwareRuleBase := dataef.NewEstbFirmwareRuleBaseDefault()
		runningVersionInfo := estbFirmwareRuleBase.GetAppliedActivationVersionType(contextMap, contextMap[common.APPLICATION_TYPE])
		fields["context"] = contextMap
		log.WithFields(fields).Info("EstbFirmwareService ActivationVersion")
		response, _ := util.JSONMarshal(*runningVersionInfo)
		xhttp.WriteXconfResponse(w, 200, response)
	}
}

func GetEstbLastlogPath(w http.ResponseWriter, r *http.Request) {
	isValid, mac, errStr := IsMacPresentAndValid(r.URL.Query())
	if !isValid {
		xhttp.WriteXconfResponseAsText(w, 400, []byte(errStr))
	} else {
		mac := util.NormalizeMacAddress(mac)
		lastConfigLog := sharedef.GetLastConfigLog(mac)
		if lastConfigLog != nil {
			LogPreDisplayCleanup(lastConfigLog)
			response, _ := util.JSONMarshal(*lastConfigLog)
			xhttp.WriteXconfResponse(w, 200, response)
		} else {
			log.Debugf("Last log is not found for mac %s", mac)
			xhttp.WriteXconfResponse(w, 200, []byte(""))
		}
	}
}

func GetEstbChangelogsPath(w http.ResponseWriter, r *http.Request) {
	isValid, mac, errStr := IsMacPresentAndValid(r.URL.Query())
	if !isValid {
		xhttp.WriteXconfResponseAsText(w, 400, []byte(errStr))
	} else {
		mac := util.NormalizeMacAddress(mac)
		configChangeLogs := sharedef.GetConfigChangeLogsOnly(mac)
		if len(configChangeLogs) > 0 {
			for _, log := range configChangeLogs {
				LogPreDisplayCleanup(log)
			}
		} else {
			log.Debugf("Last log is not found for mac %s", mac)
		}
		response, _ := util.JSONMarshal(configChangeLogs)
		xhttp.WriteXconfResponse(w, 200, response)
	}
}

func parseProcBody(body string, contextMap map[string]string) string {
	var version string
	queryParamlist := strings.Split(body, "&")
	for _, kv := range queryParamlist {
		kvlst := strings.Split(kv, "=")
		if kvlst == nil || len(kvlst) != 2 {
			continue
		}
		k := kvlst[0]
		v := kvlst[1]
		if k == common.VERSION {
			version = v
		} else {
			contextMap[k] = v
		}
	}
	return version
}
