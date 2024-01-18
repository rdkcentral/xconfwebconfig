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

	"xconfwebconfig/common"
	owcommon "xconfwebconfig/common"
	dcmlogupload "xconfwebconfig/dataapi/dcm/logupload"
	"xconfwebconfig/dataapi/dcm/settings"
	"xconfwebconfig/dataapi/dcm/telemetry"
	xhttp "xconfwebconfig/http"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/logupload"
	"xconfwebconfig/util"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func GetLogUploaderSettingsHandler(w http.ResponseWriter, r *http.Request) {
	GetLogUploaderSettings(w, r, false)
}

func GetLogUploaderT2SettingsHandler(w http.ResponseWriter, r *http.Request) {
	GetLogUploaderSettings(w, r, true)
}

func GetLogUploaderTelemetryProfilesHandler(w http.ResponseWriter, r *http.Request) {
	// ==== log pre-processing ====
	var fields log.Fields
	if xw, ok := w.(*xhttp.XResponseWriter); ok {
		fields = xw.Audit()
	} else {
		xhttp.Error(w, http.StatusInternalServerError, owcommon.NotOK)
		return
	}

	contextMap, _ := GetContextMapAndSettingTypes(r)
	AddLogUploaderContext(Ws, r, contextMap, false, fields)
	dicts, err := GetTelemetryTwoProfileResponeDicts(contextMap)
	if err != nil {
		xhttp.Error(w, http.StatusInternalServerError, err)
		return
	}
	if dicts == nil {
		xhttp.WriteXconfResponseAsText(w, 404, []byte("\"<h2>404 NOT FOUND</h2>profiles not found\""))
	} else {
		resp := util.Dict{
			"profiles": dicts,
		}
		rbytes, err := util.JSONMarshal(resp)
		if err != nil {
			xhttp.Error(w, http.StatusInternalServerError, err)
			return
		}
		xhttp.WriteResponseBytes(w, rbytes, http.StatusOK)
	}
}

func GetContextMapAndSettingTypes(r *http.Request) (map[string]string, []string) {
	applicationType, found := mux.Vars(r)[common.APPLICATION_TYPE]
	queryParams := r.URL.Query()
	if !found {
		applicationType = shared.STB
	}
	contextMap := make(map[string]string)
	contextMap[common.APPLICATION_TYPE] = applicationType
	var settingTypes []string
	if len(queryParams) > 0 {
		for k, v := range queryParams {
			if k == common.SETTING_TYPE {
				for _, settingType := range v {
					settingTypes = append(settingTypes, settingType)
				}
			} else {
				contextMap[k] = v[0]
			}
		}
	}
	return contextMap, settingTypes
}

func GetLogUploaderSettings(w http.ResponseWriter, r *http.Request, isTelemetry2Settings bool) {
	// ==== log pre-processing ====
	var fields log.Fields
	if xw, ok := w.(*xhttp.XResponseWriter); ok {
		fields = xw.Audit()
	} else {
		xhttp.Error(w, http.StatusInternalServerError, owcommon.NotOK)
		return
	}
	// audit_id-included logging example
	// fields["sample_key"] = "sample_value"
	// log.WithFields(fields).Debug("sample debug message")

	contextMap, settingTypes := GetContextMapAndSettingTypes(r)
	AddLogUploaderContext(Ws, r, contextMap, true, fields)
	checkNow, err := strconv.ParseBool(contextMap[common.CHECK_NOW])
	if err == nil && checkNow {
		telemetryProfileService := telemetry.NewTelemetryProfileService()
		telemetryProfile := telemetryProfileService.GetTelemetryForContext(contextMap)
		if telemetryProfile == nil {
			xhttp.WriteXconfResponseAsText(w, 404, []byte("\"<h2>404 NOT FOUND</h2><div> telemetry profile not found</div>\""))
		} else {
			response, _ := util.JSONMarshal(*telemetryProfile)
			xhttp.WriteXconfResponse(w, 200, response)
		}
	} else {
		logUploadRuleBase := dcmlogupload.NewLogUploadRuleBase()
		result := logUploadRuleBase.Eval(contextMap, fields)
		var telemetryRule *logupload.TelemetryRule
		if result != nil {
			telemetryProfileService := telemetry.NewTelemetryProfileService()
			telemetryRule := telemetryProfileService.GetTelemetryRuleForContext(contextMap)
			permanentTelemetryProfile := telemetryProfileService.GetPermanentProfileByTelemetryRule(telemetryRule)
			if permanentTelemetryProfile != nil {
				cloneObj, err := permanentTelemetryProfile.Clone()
				if err == nil {
					permanentTelemetryProfile = cloneObj
				} else {
					log.Error(fmt.Sprintf("GetLogUploaderSettings failed to clone %v: %v", permanentTelemetryProfile, err))
				}
				if isTelemetry2Settings {
					permanentTelemetryProfile.TelemetryProfile = ToTelemetry2Profile(permanentTelemetryProfile.TelemetryProfile)
				}
				permanentTelemetryProfile = logupload.NullifyUnwantedFieldsPermanentTelemetryProfile(permanentTelemetryProfile)
				result.TelemetryProfile = permanentTelemetryProfile
				uploadImmediately, err := strconv.ParseBool(contextMap[common.UPLOAD_IMMEDIATELY])
				if err == nil {
					result.UploadImmediately = uploadImmediately
				} else {
					result.UploadImmediately = false
				}
			}
			CleanupLusUploadRepository(result, contextMap[common.VERSION])
		}

		var settingRules []*logupload.SettingRule
		if util.IsVersionGreaterOrEqual(contextMap[common.VERSION], 2.1) && len(settingTypes) > 0 {
			var settingProfiles []logupload.SettingProfiles
			for _, settingType := range settingTypes {
				rule := settings.GetSettingsRuleByTypeForContext(settingType, contextMap)
				profile := settings.GetSettingProfileBySettingRule(rule)
				if profile != nil {
					settingProfiles = append(settingProfiles, *profile)
					settingRules = append(settingRules, rule)
				}
			}

			if result == nil && len(settingProfiles) > 0 {
				result = logupload.NewSettings(1)
			}
			if result != nil {
				result.SetSettingProfiles(settingProfiles)
			}
		}

		if result == nil {
			xhttp.WriteXconfResponseAsText(w, 404, []byte("\"<h2>404 NOT FOUND</h2><div>settings not found</div>\""))
		} else {
			LogResultSettings(result, telemetryRule, settingRules, fields)
			settingsResponse := logupload.CreateSettingsResponseObject(result)
			response, _ := util.JSONMarshal(settingsResponse)
			xhttp.WriteXconfResponse(w, 200, response)
		}
	}
}
