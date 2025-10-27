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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rdkcentral/xconfwebconfig/db"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/change"
	sharedef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	fw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"
	"github.com/rdkcentral/xconfwebconfig/tag"
	"github.com/rdkcentral/xconfwebconfig/util"

	cache "github.com/Comcast/goburrow-cache"
	conf "github.com/go-akka/configuration"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type XconfConfigs struct {
	DeriveAppTypeFromPartnerId   bool
	PartnerApplicationTypes      []string // List of partner's application type
	EnableDeviceService          bool
	EnableDeviceDBLookup         bool
	EnableMacAccountServiceCall  bool
	AccountServiceMacPrefix      string
	EnableAccountService         bool
	EnableTaggingService         bool
	EnableTaggingServiceRFC      bool
	IPv4NetworkMaskPrefixLength  int32
	IPv6NetworkMaskPrefixLength  int32
	EnableFwDownloadLogs         bool
	EnableRfcPrecook             bool
	EnableRfcPrecookForOfferedFw bool
	EnableRfcPrecook304          bool
	RfcPrecookStartTime          string
	RfcPrecookEndTime            string
	RfcPrecookTimeZone           *time.Location
	RfcPrecookTimeFormat         string
	EnableGroupService           bool
	EnableFtGroups               bool
	EnableFtMacTags              bool
	EnableFtAccountTags          bool
	EnableFtPartnerTags          bool
	GroupServiceModelSet         util.Set
	MacTagsModelSet              util.Set
	AccountTagsModelSet          util.Set
	PartnerTagsModelSet          util.Set
	MacTagsPrefixList            []string
	AccountTagsPrefixList        []string
	PartnerTagsPrefixList        []string
	ReturnAccountId              bool
	ReturnAccountHash            bool
	EstbRecoveryFirmwareVersions string
	DiagnosticAPIsEnabled        bool
	Account_mgmt                 string
	GroupServiceCacheEnabled     bool
	RfcReturnCountryCode         bool
	RfcCountryCodeModelsSet      util.Set
	RfcCountryCodePartnersSet    util.Set
	AuxiliaryFirmwareList        []AuxiliaryFirmware
	PartnerIdValidationEnabled   bool
	ValidPartnerIdRegex          *regexp.Regexp
	SecurityTokenManagerEnabled  bool
}

// Function to register the table name and the corresponding model/struct constructor
// so the DAO can instantiate the object when unmarshalling JSON data from the DB
var registerOnce sync.Once

func RegisterTables() {
	registerOnce.Do(func() {
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
		db.RegisterTableConfigSimple(db.TABLE_APP_SETTINGS, shared.NewAppSettingInf)
		db.RegisterTableConfigSimple(db.TABLE_TAG, tag.NewTagInf)

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
			TableName:       db.TABLE_XCONF_TELEMETRY_TWO_CHANGE,
			ConstructorFunc: change.NewTelemetryTwoChangeInf,
		})

		db.RegisterTableConfig(&db.TableInfo{
			TableName:       db.TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE,
			ConstructorFunc: change.NewApprovedTelemetryTwoChangeInf,
			TTL:             432000,
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

		db.RegisterTableConfig(&db.TableInfo{
			TableName:       db.TABLE_XCONF_TELEMETRY_TWO_CHANGE,
			ConstructorFunc: change.NewTelemetryTwoChangeInf,
		})

		db.RegisterTableConfig(&db.TableInfo{
			TableName:       db.TABLE_XCONF_APPROVED_TELEMETRY_TWO_CHANGE,
			ConstructorFunc: change.NewApprovedTelemetryTwoChangeInf,
			TTL:             432000,
		})
	})
}

func GetXconfConfigs(conf *conf.Config) *XconfConfigs {
	// Convert partner's application types to lowercase
	var appTypes []string
	list := conf.GetStringList("xconfwebconfig.xconf.partner_application_types")
	for _, v := range list {
		appTypes = append(appTypes, strings.ToLower(v))
	}

	GroupServiceModelSet := util.NewSet()
	GroupsModelString := conf.GetString("xconfwebconfig.xconf.group_service_model_list")
	if !util.IsBlank(GroupsModelString) {
		xdpGroupsModelList := strings.Split(GroupsModelString, ";")
		for _, model := range xdpGroupsModelList {
			GroupServiceModelSet.Add(strings.ToUpper(model))
		}
	}
	macTagsModelSet := util.NewSet()
	macTagsModelString := conf.GetString("xconfwebconfig.xconf.mac_tags_model_list")
	if !util.IsBlank(macTagsModelString) {
		macTagsModelList := strings.Split(macTagsModelString, ";")
		for _, model := range macTagsModelList {
			macTagsModelSet.Add(strings.ToUpper(model))
		}
	}
	accountTagsModelSet := util.NewSet()
	accounTagsModelString := conf.GetString("xconfwebconfig.xconf.account_tags_model_list")
	if !util.IsBlank(accounTagsModelString) {
		accountTagsModelList := strings.Split(accounTagsModelString, ";")
		for _, model := range accountTagsModelList {
			accountTagsModelSet.Add(strings.ToUpper(model))
		}
	}
	partnerTagsModelSet := util.NewSet()
	partnerTagsModelString := conf.GetString("xconfwebconfig.xconf.partner_tags_model_list")
	if !util.IsBlank(partnerTagsModelString) {
		partnerTagsModelList := strings.Split(partnerTagsModelString, ";")
		for _, model := range partnerTagsModelList {
			partnerTagsModelSet.Add(strings.ToUpper(model))
		}
	}

	macTagsPrefixList := []string{}
	macTagsPrefixString := conf.GetString("xconfwebconfig.xconf.mac_tags_prefix_list")
	if !util.IsBlank(macTagsPrefixString) {
		macTagsPrefixList = strings.Split(macTagsPrefixString, ";")
	}
	accountTagsPrefixList := []string{}
	accountTagsPrefixString := conf.GetString("xconfwebconfig.xconf.account_tags_prefix_list")
	if !util.IsBlank(accountTagsPrefixString) {
		accountTagsPrefixList = strings.Split(accountTagsPrefixString, ";")
	}
	partnerTagsPrefixList := []string{}
	partnerTagsPrefixString := conf.GetString("xconfwebconfig.xconf.partner_tags_prefix_list")
	if !util.IsBlank(partnerTagsPrefixString) {
		partnerTagsPrefixList = strings.Split(partnerTagsPrefixString, ";")
	}
	rfcPrecookEnabled := conf.GetBoolean("xconfwebconfig.xconf.enable_rfc_precook")
	rfcPrecookForOfferedFwEnabled := conf.GetBoolean("xconfwebconfig.xconf.enable_rfc_precook_for_offered_fw")
	var timezone *time.Location
	var err error
	if rfcPrecookEnabled {
		timezoneStr := conf.GetString("xconfwebconfig.xconf.rfc_precook_time_zone")
		// if timezoneStr is empty, defaults on UTC
		timezone, err = time.LoadLocation(timezoneStr)
		if err != nil {
			log.Errorf("Error loading timezone: %s", timezoneStr)
			panic(err)
		}
	}

	rfcCountryCodeModelsSet := util.NewSet()
	rfcCountryCodeModelsList := conf.GetString("xconfwebconfig.xconf.rfc_country_code_model_list")
	if !util.IsBlank(rfcCountryCodeModelsList) {
		rfcCountryCodeModels := strings.Split(rfcCountryCodeModelsList, ";")
		for _, model := range rfcCountryCodeModels {
			rfcCountryCodeModelsSet.Add(strings.ToUpper(strings.TrimSpace(model)))
		}
	}

	rfcCountryCodePartnersSet := util.NewSet()
	rfcCountryCodePartnersList := conf.GetString("xconfwebconfig.xconf.rfc_country_code_partner_list")
	if !util.IsBlank(rfcCountryCodePartnersList) {
		rfcCountryCodePartners := strings.Split(rfcCountryCodePartnersList, ";")
		for _, partner := range rfcCountryCodePartners {
			rfcCountryCodePartnersSet.Add(strings.ToUpper(strings.TrimSpace(partner)))
		}
	}

	auxFirmwareList := getAuxiliaryFirmwares(conf.GetString("xconfwebconfig.xconf.auxiliary_extensions"))
	partnerIdValidationEnabled := conf.GetBoolean("xconfwebconfig.xconf.partner_id_validation_enabled", false)

	// Partner ID regex config
	const defaultValidPartnerIdRegex = `^[A-Za-z0-9_.\-,:;]{3,32}$`
	validPartnerIdRegexStr := conf.GetString("xconfwebconfig.xconf.valid_partner_id_regex", defaultValidPartnerIdRegex)
	validPartnerIdRegex, err := regexp.Compile(validPartnerIdRegexStr)
	if err != nil {
		log.Warnf("Invalid partner ID regex pattern '%s', using default. Error: %v", validPartnerIdRegexStr, err)
		validPartnerIdRegex = regexp.MustCompile(defaultValidPartnerIdRegex)
	}

	xc := &XconfConfigs{
		DeriveAppTypeFromPartnerId:   conf.GetBoolean("xconfwebconfig.xconf.derive_application_type_from_partner_id"),
		PartnerApplicationTypes:      appTypes,
		EnableDeviceService:          conf.GetBoolean("xconfwebconfig.xconf.enable_device_service"),
		EnableDeviceDBLookup:         conf.GetBoolean("xconfwebconfig.xconf.enable_device_db_lookup"),
		EnableMacAccountServiceCall:  conf.GetBoolean("xconfwebconfig.xconf.enable_mac_accountservice_call"),
		EnableAccountService:         conf.GetBoolean("xconfwebconfig.xconf.enable_account_service"),
		EnableTaggingService:         conf.GetBoolean("xconfwebconfig.xconf.enable_tagging_service"),
		EnableTaggingServiceRFC:      conf.GetBoolean("xconfwebconfig.xconf.enable_tagging_service_rfc"),
		ReturnAccountId:              conf.GetBoolean("xconfwebconfig.xconf.return_account_id"),
		ReturnAccountHash:            conf.GetBoolean("xconfwebconfig.xconf.return_account_hash"),
		EstbRecoveryFirmwareVersions: conf.GetString("xconfwebconfig.xconf.estb_recovery_firmware_versions"),
		DiagnosticAPIsEnabled:        conf.GetBoolean("xconfwebconfig.xconf.diagnostic_apis_enabled"),
		AccountServiceMacPrefix:      conf.GetString("xconfwebconfig.xconf.account_service_mac_prefix"),
		IPv4NetworkMaskPrefixLength:  conf.GetInt32("xconfwebconfig.xconf.ipv4_network_mask_prefix_length"),
		IPv6NetworkMaskPrefixLength:  conf.GetInt32("xconfwebconfig.xconf.ipv6_network_mask_prefix_length"),
		EnableFwDownloadLogs:         conf.GetBoolean("xconfwebconfig.xconf.enable_fw_download_logs"),
		EnableRfcPrecook:             rfcPrecookEnabled,
		EnableRfcPrecookForOfferedFw: rfcPrecookForOfferedFwEnabled,
		EnableRfcPrecook304:          conf.GetBoolean("xconfwebconfig.xconf.enable_rfc_precook_304"),
		RfcPrecookStartTime:          conf.GetString("xconfwebconfig.xconf.rfc_precook_start_time"),
		RfcPrecookEndTime:            conf.GetString("xconfwebconfig.xconf.rfc_precook_end_time"),
		RfcPrecookTimeZone:           timezone,
		RfcPrecookTimeFormat:         conf.GetString("xconfwebconfig.xconf.rfc_precook_time_format"),
		EnableGroupService:           conf.GetBoolean("xconfwebconfig.xconf.enable_group_service"),
		EnableFtMacTags:              conf.GetBoolean("xconfwebconfig.xconf.enable_ft_mac_tags"),
		EnableFtAccountTags:          conf.GetBoolean("xconfwebconfig.xconf.enable_ft_account_tags"),
		EnableFtPartnerTags:          conf.GetBoolean("xconfwebconfig.xconf.enable_ft_partner_tags"),
		EnableFtGroups:               conf.GetBoolean("xconfwebconfig.xconf.enable_ft_xdp_groups"),
		GroupServiceModelSet:         GroupServiceModelSet,
		MacTagsModelSet:              macTagsModelSet,
		AccountTagsModelSet:          accountTagsModelSet,
		PartnerTagsModelSet:          partnerTagsModelSet,
		MacTagsPrefixList:            macTagsPrefixList,
		AccountTagsPrefixList:        accountTagsPrefixList,
		PartnerTagsPrefixList:        partnerTagsPrefixList,
		GroupServiceCacheEnabled:     conf.GetBoolean(fmt.Sprintf("xconfwebconfig.%v.cache_enabled", conf.GetString("xconfwebconfig.xconf.group_service_name"))),
		RfcReturnCountryCode:         conf.GetBoolean("xconfwebconfig.xconf.rfc_return_country_code"),
		RfcCountryCodeModelsSet:      rfcCountryCodeModelsSet,
		RfcCountryCodePartnersSet:    rfcCountryCodePartnersSet,
		AuxiliaryFirmwareList:        auxFirmwareList,
		ValidPartnerIdRegex:          validPartnerIdRegex,
		PartnerIdValidationEnabled:   partnerIdValidationEnabled,
		SecurityTokenManagerEnabled:  conf.GetBoolean("xconfwebconfig.xconf.security_token_manager_enabled"),
	}
	return xc
}

// Xconf setup
func XconfSetup(server *xhttp.XconfServer, r *mux.Router) {
	xc := GetXconfConfigs(server.ServerConfig.Config)
	WebServerInjection(server, xc)
	db.ConfigInjection(server.ServerConfig.Config)
	db.SetGrpCacheLoadFunc(LoadGroupServiceFeatureTags)
	RegisterTables()
	db.GetCacheManager() // Initialize cache manager

	RouteXconfDataserviceApis(r, server)

	if xc.DiagnosticAPIsEnabled {
		RouteDiagnosticApis(r, server)
	}
}

func RouteXconfDataserviceApis(r *mux.Router, s *xhttp.XconfServer) {
	paths := []*mux.Router{}

	getFeatureSettingsPath := r.Path("/featureControl/getSettings").Subrouter()
	getFeatureSettingsPath.HandleFunc("", GetFeatureControlSettingsHandler).Methods("GET", "HEAD")
	paths = append(paths, getFeatureSettingsPath)

	getPrecookFeatureSettingsPath := r.Path("/preprocess/rfc/{mac}").Subrouter()
	getPrecookFeatureSettingsPath.HandleFunc("", GetPreprocessFeatureControlSettingsHandler).Methods("GET", "HEAD")
	paths = append(paths, getPrecookFeatureSettingsPath)

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
		p.Use(s.SpanMiddleware)
		p.Use(s.NoAuthMiddleware)
	}

	// Hack, set a config var to use a map in rules engine
	rulesengine.UseMap = s.GetBoolean("xconfwebconfig.misc.use_map_for_evaluators", false)
}

// Potential Todo: Add metrics to these routes as well
func RouteDiagnosticApis(r *mux.Router, s *xhttp.XconfServer) {
	paths := []*mux.Router{}

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

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "404")
}

func LoadGroupServiceFeatureTags(key cache.Key) (cache.Value, error) {
	log.WithFields(log.Fields{"key": key}).Info("loading function for group service cache called")
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered in loadGroupServiceFeatureTags: %v", r)
		}
	}()
	partnerID := key.(string)
	fields := log.Fields{"partnerId": partnerID}
	// Fetch feature tags from XDAS service
	featureTags, err := Ws.GroupServiceConnector.GetFeatureTagsHashedItems(partnerID, fields)
	if err != nil {
		if isNotFoundError(err) {
			log.WithFields(log.Fields{"error": err}).Debugf("No feature tags found for partner=%s", partnerID)

			// Cache the absence of tags by storing nil
			emptyTags := map[string]string{}
			cacheErr := db.GetCacheManager().SetGroupServiceFeatureTags(partnerID, emptyTags)
			if cacheErr != nil {
				log.WithFields(log.Fields{"error": cacheErr, "partnerId": partnerID}).Error("Failed to cache empty tags")
			}
			return emptyTags, nil
		}
		log.WithFields(log.Fields{"error": err}).Debugf("Error getting response from XDAS for partner=%s", partnerID)
		return nil, err
	}
	return featureTags, nil
}

func getAuxiliaryFirmwares(auxExtensionString string) []AuxiliaryFirmware {
	var auxFirmwareList []AuxiliaryFirmware
	if auxExtensionString != "" {
		auxExtensionList := strings.Split(auxExtensionString, ";")
		// create list with length 0 but capacity the num of auxExtensions
		auxFirmwareList = make([]AuxiliaryFirmware, 0, len(auxExtensionList))
		for _, auxExtension := range auxExtensionList {
			auxPairList := strings.Split(auxExtension, ":")
			if len(auxPairList) == 2 {
				auxPair := AuxiliaryFirmware{
					Prefix:    auxPairList[0],
					Extension: auxPairList[1],
				}
				auxFirmwareList = append(auxFirmwareList, auxPair)
			}
		}
	}
	return auxFirmwareList
}
