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

	db "xconfwebconfig/db"
	xhttp "xconfwebconfig/http"
	"xconfwebconfig/rulesengine"
	"xconfwebconfig/shared"
	"xconfwebconfig/shared/change"
	sharedef "xconfwebconfig/shared/estbfirmware"
	fw "xconfwebconfig/shared/firmware"
	"xconfwebconfig/shared/logupload"
	"xconfwebconfig/shared/rfc"

	conf "github.com/go-akka/configuration"
	"github.com/gorilla/mux"
)

type XconfConfigs struct {
	DeriveAppTypeFromPartnerId   bool
	PartnerApplicationTypes      []string // List of partner's application type
	EnableDeviceService          bool
	EnableAccountService         bool
	EnableTaggingService         bool
	EnableTaggingServiceRFC      bool
	ReturnAccountId              bool
	ReturnAccountHash            bool
	DiagnosticAPIsEnabled        bool
	EstbRecoveryFirmwareVersions string
}

// Function to register the table name and the corresponding model/struct constructor
// so the DAO can instantiate the object when unmarshalling JSON data from the DB
func registerTables() {
	db.RegisterTableConfigSimple(db.TABLE_DCM_RULE, logupload.NewDCMGenericRuleInf)
	db.RegisterTableConfigSimple(db.TABLE_ENVIRONMENT, shared.NewEnvironmentInf)
	db.RegisterTableConfigSimple(db.TABLE_MODEL, shared.NewModelInf)
	db.RegisterTableConfigSimple(db.TABLE_IP_ADDRESS_GROUP, shared.NewIpAddressGroupInf)
	db.RegisterTableConfigSimple(db.TABLE_FIRMWARE_CONFIG, sharedef.NewFirmwareConfigInf)
	db.RegisterTableConfigSimple(db.TABLE_FIRMWARE_RULE, fw.NewFirmwareRuleInf)
	db.RegisterTableConfigSimple(db.TABLE_FIRMWARE_RULE_TEMPLATE, fw.NewFirmwareRuleTemplateInf)
	db.RegisterTableConfigSimple(db.TABLE_SINGLETON_FILTER_VALUE, sharedef.NewSingletonFilterValueInf)
	db.RegisterTableConfigSimple(db.TABLE_UPLOAD_REPOSITORY, logupload.NewUploadRepositoryInf)
	db.RegisterTableConfigSimple(db.TABLE_LOG_FILE, logupload.NewLogFileInf)
	db.RegisterTableConfigSimple(db.TABLE_LOG_FILE_LIST, logupload.NewLogFileListInf)
	db.RegisterTableConfigSimple(db.TABLE_LOG_FILES_GROUPS, logupload.NewLogFilesGroupsInf)
	db.RegisterTableConfigSimple(db.TABLE_LOG_UPLOAD_SETTINGS, logupload.NewLogUploadSettingsInf)
	db.RegisterTableConfigSimple(db.TABLE_SETTING_PROFILES, logupload.NewSettingProfilesInf)
	db.RegisterTableConfigSimple(db.TABLE_SETTING_RULES, logupload.NewSettingRulesInf)
	db.RegisterTableConfigSimple(db.TABLE_DEVICE_SETTINGS, logupload.NewDeviceSettingsInf)
	db.RegisterTableConfigSimple(db.TABLE_VOD_SETTINGS, logupload.NewVodSettingsInf)
	db.RegisterTableConfigSimple(db.TABLE_TELEMETRY, logupload.NewTelemetryProfileInf)
	db.RegisterTableConfigSimple(db.TABLE_PERMANENT_TELEMETRY, logupload.NewPermanentTelemetryProfileInf)
	db.RegisterTableConfigSimple(db.TABLE_TELEMETRY_RULES, logupload.NewTelemetryRuleInf)
	db.RegisterTableConfigSimple(db.TABLE_TELEMETRY_TWO_PROFILES, logupload.NewTelemetryTwoProfileInf)
	db.RegisterTableConfigSimple(db.TABLE_TELEMETRY_TWO_RULES, logupload.NewTelemetryTwoRuleInf)
	db.RegisterTableConfigSimple(db.TABLE_XCONF_FEATURE, rfc.NewFeatureInf)
	db.RegisterTableConfigSimple(db.TABLE_FEATURE_CONTROL_RULE, rfc.NewFeatureRuleInf)

	db.RegisterTableConfig(&db.TableInfo{
		TableName:       db.TABLE_XCONF_CHANGE,
		ConstructorFunc: change.NewChangeInf,
		CacheData:       false,
	})

	db.RegisterTableConfig(&db.TableInfo{
		TableName:       db.TABLE_XCONF_APPROVED_CHANGE,
		ConstructorFunc: change.NewApprovedChangeInf,
		TTL:             432000,
		CacheData:       false,
	})

	db.RegisterTableConfig(&db.TableInfo{
		TableName:       db.TABLE_LOGS,
		ConstructorFunc: sharedef.NewConfigChangeLogInf,
		Compress:        true,
		TTL:             90 * 24 * 60 * 60,
		Key2FieldName:   db.DefaultKey2FieldName,
	})

	db.RegisterTableConfig(&db.TableInfo{
		TableName:       db.TABLE_NS_LIST,
		ConstructorFunc: shared.NewNamespacedListInf,
		Compress:        true,
		Split:           true,
		CacheData:       true,
	})

	db.RegisterTableConfig(&db.TableInfo{
		TableName:       db.TABLE_GENERIC_NS_LIST,
		ConstructorFunc: shared.NewGenericNamespacedListInf,
		Compress:        true,
		Split:           true,
		CacheData:       true,
	})

	db.RegisterTableConfig(&db.TableInfo{
		TableName:       db.TABLE_XCONF_CHANGED_KEYS,
		ConstructorFunc: db.NewChangedDataInf,
		Key2FieldName:   db.ChangedKeysKey2FieldName,
		TTL:             86400 * 7, // one week
	})
}

func GetXconfConfigs(conf *conf.Config) *XconfConfigs {
	// Convert partner's application types to lowercase
	var appTypes []string
	list := conf.GetStringList("xconfwebconfig.xconf.partner_application_types")
	for _, v := range list {
		appTypes = append(appTypes, strings.ToLower(v))
	}

	xc := &XconfConfigs{
		DeriveAppTypeFromPartnerId:   conf.GetBoolean("xconfwebconfig.xconf.derive_application_type_from_partner_id"),
		PartnerApplicationTypes:      appTypes,
		EnableDeviceService:          conf.GetBoolean("xconfwebconfig.xconf.enable_odp_service"),
		EnableAccountService:         conf.GetBoolean("xconfwebconfig.xconf.enable_titan_service"),
		EnableTaggingService:         conf.GetBoolean("xconfwebconfig.xconf.enable_tagging_service"),
		EnableTaggingServiceRFC:      conf.GetBoolean("xconfwebconfig.xconf.enable_tagging_service_rfc"),
		ReturnAccountId:              conf.GetBoolean("xconfwebconfig.xconf.return_account_id"),
		ReturnAccountHash:            conf.GetBoolean("xconfwebconfig.xconf.return_account_hash"),
		EstbRecoveryFirmwareVersions: conf.GetString("xconfwebconfig.xconf.estb_recovery_firmware_versions"),
		DiagnosticAPIsEnabled:        conf.GetBoolean("xconfwebconfig.xconf.diagnostic_apis_enabled"),
	}
	return xc
}

// Xconf setup
func XconfSetup(server *xhttp.XconfServer, r *mux.Router) {
	xc := GetXconfConfigs(server.ServerConfig.Config)

	WebServerInjection(server, xc)
	db.ConfigInjection(server.ServerConfig.Config)

	registerTables()
	db.GetCacheManager() // Initialize cache manager

	routeXconfDataserviceApis(r, server)

	if xc.DiagnosticAPIsEnabled {
		routeDiagnosticApis(r, server)
	}
}

func routeXconfDataserviceApis(r *mux.Router, s *xhttp.XconfServer) {
	paths := []*mux.Router{}

	getFeatureSettingsPath := r.Path("/featureControl/getSettings").Subrouter()
	getFeatureSettingsPath.HandleFunc("", GetFeatureControlSettingsHandler).Methods("GET", "HEAD")
	paths = append(paths, getFeatureSettingsPath)

	getFeatureSettingsApplicationTypePath := r.Path("/featureControl/getSettings/{applicationType}").Subrouter()
	getFeatureSettingsApplicationTypePath.HandleFunc("", GetFeatureControlSettingsHandler).Methods("GET", "HEAD")
	paths = append(paths, getFeatureSettingsApplicationTypePath)

	getEstbFirmwareSwuBsePath := r.Path("/xconf/swu/bse").Subrouter()
	getEstbFirmwareSwuBsePath.HandleFunc("", GetEstbFirmwareSwuBseHandler)
	paths = append(paths, getEstbFirmwareSwuBsePath)

	// Trailing slash version of below
	// Note that trailing slash has to be specified in both the Path and HandleFunc
	getEstbFirmwareSwuPathWithTrailingSlash := r.Path("/xconf/swu/{applicationType}/").Subrouter()
	getEstbFirmwareSwuPathWithTrailingSlash.HandleFunc("/", GetEstbFirmwareSwuHandler)
	paths = append(paths, getEstbFirmwareSwuPathWithTrailingSlash)

	getEstbFirmwareSwuPath := r.Path("/xconf/swu/{applicationType}").Subrouter()
	getEstbFirmwareSwuPath.HandleFunc("", GetEstbFirmwareSwuHandler)
	paths = append(paths, getEstbFirmwareSwuPath)

	getLogUploaderSettingsPath := r.Path("/loguploader/getSettings").Subrouter()
	getLogUploaderSettingsPath.HandleFunc("", GetLogUploaderSettingsHandler).Methods("GET", "HEAD")
	paths = append(paths, getLogUploaderSettingsPath)

	getLogUploaderSettingsApplicationTypePath := r.Path("/loguploader/getSettings/{applicationType}").Subrouter()
	getLogUploaderSettingsApplicationTypePath.HandleFunc("", GetLogUploaderSettingsHandler).Methods("GET", "HEAD")
	paths = append(paths, getLogUploaderSettingsApplicationTypePath)

	getLogUploaderT2SettingsPath := r.Path("/loguploader/getT2Settings").Subrouter()
	getLogUploaderT2SettingsPath.HandleFunc("", GetLogUploaderT2SettingsHandler).Methods("GET", "HEAD")
	paths = append(paths, getLogUploaderT2SettingsPath)

	getLogUploaderT2SettingsApplicationTypePath := r.Path("/loguploader/getT2Settings/{applicationType}").Subrouter()
	getLogUploaderT2SettingsApplicationTypePath.HandleFunc("", GetLogUploaderT2SettingsHandler).Methods("GET", "HEAD")
	paths = append(paths, getLogUploaderT2SettingsApplicationTypePath)

	getLogUploaderTelemetryProfilesPath := r.Path("/loguploader/getTelemetryProfiles").Subrouter()
	getLogUploaderTelemetryProfilesPath.HandleFunc("", GetLogUploaderTelemetryProfilesHandler).Methods("GET")
	paths = append(paths, getLogUploaderTelemetryProfilesPath)

	getLogUploaderTelemetryProfilesAppTypePath := r.Path("/loguploader/getTelemetryProfiles/{applicationType}").Subrouter()
	getLogUploaderTelemetryProfilesAppTypePath.HandleFunc("", GetLogUploaderTelemetryProfilesHandler).Methods("GET")
	paths = append(paths, getLogUploaderTelemetryProfilesAppTypePath)

	getCheckMinFirmwarePath := r.Path("/estbfirmware/checkMinimumFirmware").Subrouter()
	getCheckMinFirmwarePath.HandleFunc("", GetCheckMinFirmwareHandler)
	paths = append(paths, getCheckMinFirmwarePath)

	getEstbFirmwareVersionInfoPath := r.Path("/xconf/{applicationType}/runningFirmwareVersion/info").Subrouter()
	getEstbFirmwareVersionInfoPath.HandleFunc("", GetEstbFirmwareVersionInfoPath)
	paths = append(paths, getEstbFirmwareVersionInfoPath)

	// r.NotFoundHandler = PathNotFoundHandler()
	for _, p := range paths {
		p.Use(s.NoAuthMiddleware)
	}

	// Hack, set a config var to use a map in rules engine
	rulesengine.UseMap = s.GetBoolean("xconfwebconfig.misc.use_map_for_evaluators", false)
}

// Potential Todo: Add metrics to these routes as well
func routeDiagnosticApis(r *mux.Router, s *xhttp.XconfServer) {
	paths := []*mux.Router{}

	getConfigPath := r.Path("/config").Subrouter()
	getConfigPath.HandleFunc("", s.ServerConfigHandler).Methods("GET")
	paths = append(paths, getConfigPath)

	getInfoRefreshAllPath := r.Path("/info/refreshAll").Subrouter()
	getInfoRefreshAllPath.HandleFunc("", GetInfoRefreshAllHandler).Methods("GET")
	paths = append(paths, getInfoRefreshAllPath)

	getInfoRefreshPath := r.Path("/info/refresh/{tableName}").Subrouter()
	getInfoRefreshPath.HandleFunc("", GetInfoRefreshHandler).Methods("GET")
	paths = append(paths, getInfoRefreshPath)

	getInfoStatisticsPath := r.Path("/info/statistics").Subrouter()
	getInfoStatisticsPath.HandleFunc("", GetInfoStatistics).Methods("GET")
	paths = append(paths, getInfoStatisticsPath)
}

// PathNotFoundHandler - invalid URL should return 404 with message
func PathNotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xhttp.WriteXconfResponse(w, 404, []byte(fmt.Sprintf("Problem accessing %s", r.URL.Path)))

	})
}
