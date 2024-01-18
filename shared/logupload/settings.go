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
package logupload

import (
	"fmt"
	"strings"

	"xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	util "xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

// Enum for SettingType
const (
	EPON = iota + 1
	PARTNER_SETTINGS
)

func SettingTypeEnum(s string) int {
	switch strings.ToLower(s) {
	case "epon":
		return EPON
	case "partner_settings", "partnersettings":
		return PARTNER_SETTINGS
	}
	return 0
}

// SettingProfiles table
type SettingProfiles struct {
	ID               string            `json:"id"`
	Updated          int64             `json:"updated"`
	SettingProfileID string            `json:"settingProfileId"`
	SettingType      string            `json:"settingType"`
	Properties       map[string]string `json:"properties"`
	ApplicationType  string            `json:"applicationType"`
}

func (obj *SettingProfiles) Clone() (*SettingProfiles, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*SettingProfiles), nil
}

// NewSettingProfilesInf constructor
func NewSettingProfilesInf() interface{} {
	return &SettingProfiles{}
}

type FormulaWithSettings struct {
	Formula           *DCMGenericRule    `json:"formula"`
	DeviceSettings    *DeviceSettings    `json:"deviceSettings"`
	LogUpLoadSettings *LogUploadSettings `json:"logUploadSettings"`
	VodSettings       *VodSettings       `json:"vodSettings"`
}

// VodSettings table
type VodSettings struct {
	ID              string            `json:"id"`
	Updated         int64             `json:"updated"`
	Name            string            `json:"name"`
	LocationsURL    string            `json:"locationsURL"`
	IPNames         []string          `json:"ipNames"`
	IPList          []string          `json:"ipList"`
	SrmIPList       map[string]string `json:"srmIPList"`
	ApplicationType string            `json:"applicationType"`
}

func (obj *VodSettings) Clone() (*VodSettings, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*VodSettings), nil
}

// NewVodSettingsInf constructor
func NewVodSettingsInf() interface{} {
	return &VodSettings{}
}

// SettingRule SettingRules table
type SettingRule struct {
	ID              string  `json:"id"`
	Updated         int64   `json:"updated"`
	Name            string  `json:"name"`
	Rule            re.Rule `json:"rule"`
	BoundSettingID  string  `json:"boundSettingId"`
	ApplicationType string  `json:"applicationType"`
}

func (obj *SettingRule) Clone() (*SettingRule, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*SettingRule), nil
}

func (r *SettingRule) GetApplicationType() string {
	if len(r.ApplicationType) > 0 {
		return r.ApplicationType
	}
	return "stb"
}

// GetId XRule interface
func (r *SettingRule) GetId() string {
	return r.ID
}

// GetRule XRule interface
func (r *SettingRule) GetRule() *re.Rule {
	return &r.Rule
}

// GetName XRule interface
func (r *SettingRule) GetName() string {
	return r.Name
}

// GetTemplateId XRule interface
func (r *SettingRule) GetTemplateId() string {
	return ""
}

// GetRuleType XRule interface
func (r *SettingRule) GetRuleType() string {
	return "SettingRule"
}

// NewSettingRulesInf constructor
func NewSettingRulesInf() interface{} {
	return &SettingRule{}
}

const (
	DEFAULT_LOG_UPLOAD_SETTINGS_MESSAGE = "Don't upload your logs, but check for updates on this schedule."
)

type Settings struct {
	RuleIDs                           map[string]bool
	SchedulerType                     string
	GroupName                         string
	CheckOnReboot                     bool
	ConfigurationServiceURL           string
	ScheduleCron                      string
	TimeZoneMode                      string
	ScheduleDurationMinutes           int
	ScheduleStartDate                 string
	ScheduleEndDate                   string
	LusMessage                        string
	LusName                           string
	LusNumberOfDay                    int
	LusUploadRepositoryName           string
	LusUploadRepositoryURLNew         string
	LusUploadRepositoryUploadProtocol string
	LusUploadRepositoryURL            string
	LusUploadOnReboot                 bool
	UploadImmediately                 bool
	//Upload flag to indicate if allowed to upload logs or not.
	Upload               bool
	LusLogFiles          []*LogFile
	LusLogFilesStartDate string
	LusLogFilesEndDate   string
	//For level one logging
	LusScheduleCron            string
	LusTimeZoneMode            string
	LusScheduleCronL1          string
	LusScheduleCronL2          string
	LusScheduleCronL3          string
	LusScheduleDurationMinutes int
	LusScheduleStartDate       string
	LusScheduleEndDate         string
	VodSettingsName            string
	LocationUrl                string
	TelemetryProfile           *PermanentTelemetryProfile
	SrmIPList                  map[string]string
	EponSettings               map[string]string
	PartnerSettings            map[string]string
}

func NewSettings(logFileLenth int) *Settings {
	var newSettings *Settings
	newSettings = new(Settings)
	newSettings.RuleIDs = make(map[string]bool)
	newSettings.SrmIPList = make(map[string]string)
	newSettings.EponSettings = make(map[string]string)
	newSettings.PartnerSettings = make(map[string]string)
	newSettings.LusLogFiles = make([]*LogFile, logFileLenth)
	return newSettings
}

func (s *Settings) CopyDeviceSettings(settings *Settings) {
	s.GroupName = settings.GroupName
	s.CheckOnReboot = settings.CheckOnReboot
	s.ConfigurationServiceURL = settings.ConfigurationServiceURL
	s.ScheduleCron = settings.ScheduleCron
	s.ScheduleDurationMinutes = settings.ScheduleDurationMinutes
	s.ScheduleStartDate = settings.ScheduleStartDate
	s.ScheduleEndDate = settings.ScheduleEndDate
	s.TimeZoneMode = settings.TimeZoneMode
}

func (s *Settings) CopyLusSetting(settings *Settings, setLUSSettings bool) {
	if setLUSSettings {
		s.LusMessage = ""
		s.LusName = settings.LusName
		s.LusNumberOfDay = settings.LusNumberOfDay
		s.LusUploadRepositoryName = settings.LusUploadRepositoryName
		s.LusUploadRepositoryURL = settings.LusUploadRepositoryURL
		s.LusUploadRepositoryURLNew = settings.LusUploadRepositoryURLNew
		s.LusUploadRepositoryUploadProtocol = settings.LusUploadRepositoryUploadProtocol
		s.LusUploadOnReboot = settings.LusUploadOnReboot
		s.LusLogFiles = settings.LusLogFiles
		s.LusLogFilesStartDate = settings.LusLogFilesStartDate
		s.LusLogFilesEndDate = settings.LusLogFilesEndDate
		s.LusTimeZoneMode = settings.LusTimeZoneMode
		s.LusScheduleDurationMinutes = settings.LusScheduleDurationMinutes
		s.LusScheduleStartDate = settings.LusScheduleStartDate
		s.LusScheduleEndDate = settings.LusScheduleEndDate
		s.LusTimeZoneMode = settings.LusTimeZoneMode
		s.Upload = true
	} else {
		s.LusMessage = DEFAULT_LOG_UPLOAD_SETTINGS_MESSAGE
		s.LusName = ""
		s.LusNumberOfDay = 0
		s.LusUploadRepositoryName = ""
		s.LusUploadRepositoryURL = ""
		s.LusUploadRepositoryURLNew = ""
		s.LusUploadRepositoryUploadProtocol = ""
		s.LusUploadOnReboot = false
		s.LusLogFiles = nil
		s.LusLogFilesStartDate = ""
		s.LusLogFilesEndDate = ""
		s.LusScheduleDurationMinutes = 0
		s.LusTimeZoneMode = ""
		s.LusScheduleStartDate = ""
		s.LusScheduleEndDate = ""
		s.LusTimeZoneMode = ""
		s.Upload = false
	}
}

func (s *Settings) CopyVodSettings(settings *Settings) {
	s.VodSettingsName = settings.VodSettingsName
	s.LocationUrl = settings.LocationUrl
	s.SrmIPList = settings.SrmIPList
}

func (s *Settings) AreFull() bool {
	if s.GroupName != "" && s.LusName != "" && s.VodSettingsName != "" {
		return true
	}
	return false
}

func (s *Settings) SetSettingProfiles(settingProfiles []SettingProfiles) {
	if len(settingProfiles) < 1 {
		return
	}
	for _, settingProfile := range settingProfiles {
		properties := settingProfile.Properties
		switch SettingTypeEnum(settingProfile.SettingType) {
		case PARTNER_SETTINGS:
			s.PartnerSettings = properties
		case EPON:
			s.EponSettings = properties
		}

	}
}

type SettingsResponse struct {
	GroupName                         interface{}                `json:"urn:settings:GroupName"`
	CheckOnReboot                     bool                       `json:"urn:settings:CheckOnReboot"`
	TimeZoneMode                      string                     `json:"urn:settings:TimeZoneMode"`
	ScheduleCron                      interface{}                `json:"urn:settings:CheckSchedule:cron"`
	ScheduleDurationMinutes           int                        `json:"urn:settings:CheckSchedule:DurationMinutes"`
	LusMessage                        interface{}                `json:"urn:settings:LogUploadSettings:Message"`
	LusName                           interface{}                `json:"urn:settings:LogUploadSettings:Name"`
	LusNumberOfDay                    int                        `json:"urn:settings:LogUploadSettings:NumberOfDays"`
	LusUploadRepositoryName           interface{}                `json:"urn:settings:LogUploadSettings:UploadRepositoryName"`
	LusUploadRepositoryURLNew         string                     `json:"urn:settings:LogUploadSettings:UploadRepository:URL,omitempty"`
	LusUploadRepositoryUploadProtocol string                     `json:"urn:settings:LogUploadSettings:UploadRepository:uploadProtocol,omitempty"`
	LusUploadRepositoryURL            string                     `json:"urn:settings:LogUploadSettings:RepositoryURL,omitempty"`
	LusUploadOnReboot                 bool                       `json:"urn:settings:LogUploadSettings:UploadOnReboot"`
	UploadImmediately                 bool                       `json:"urn:settings:LogUploadSettings:UploadImmediately"`
	Upload                            bool                       `json:"urn:settings:LogUploadSettings:upload"`
	LusScheduleCron                   interface{}                `json:"urn:settings:LogUploadSettings:UploadSchedule:cron"`
	LusScheduleCronL1                 interface{}                `json:"urn:settings:LogUploadSettings:UploadSchedule:levelone:cron"`
	LusScheduleCronL2                 interface{}                `json:"urn:settings:LogUploadSettings:UploadSchedule:leveltwo:cron"`
	LusScheduleCronL3                 interface{}                `json:"urn:settings:LogUploadSettings:UploadSchedule:levelthree:cron"`
	LusTimeZoneMode                   string                     `json:"urn:settings:LogUploadSettings:UploadSchedule:TimeZoneMode"`
	LusScheduleDurationMinutes        int                        `json:"urn:settings:LogUploadSettings:UploadSchedule:DurationMinutes"`
	VodSettingsName                   interface{}                `json:"urn:settings:VODSettings:Name"`
	LocationUrl                       interface{}                `json:"urn:settings:VODSettings:LocationsURL"`
	SrmIPList                         interface{}                `json:"urn:settings:VODSettings:SRMIPList"`
	EponSettings                      map[string]string          `json:"urn:settings:SettingType:epon,omitempty"`
	TelemetryProfile                  *PermanentTelemetryProfile `json:"urn:settings:TelemetryProfile,omitempty"`
	PartnerSettings                   map[string]string          `json:"urn:settings:SettingType:partnersettings,omitempty"`
}

func CreateSettingsResponseObject(settings *Settings) *SettingsResponse {
	settingsResponse := &SettingsResponse{
		CheckOnReboot:                     settings.CheckOnReboot,
		TimeZoneMode:                      settings.TimeZoneMode,
		ScheduleDurationMinutes:           settings.ScheduleDurationMinutes,
		LusNumberOfDay:                    settings.LusNumberOfDay,
		LusUploadRepositoryURLNew:         settings.LusUploadRepositoryURLNew,
		LusUploadRepositoryUploadProtocol: settings.LusUploadRepositoryUploadProtocol,
		LusUploadRepositoryURL:            settings.LusUploadRepositoryURL,
		LusUploadOnReboot:                 settings.LusUploadOnReboot,
		UploadImmediately:                 settings.UploadImmediately,
		Upload:                            settings.Upload,
		LusScheduleDurationMinutes:        settings.LusScheduleDurationMinutes,
		LusTimeZoneMode:                   settings.LusTimeZoneMode,
		EponSettings:                      settings.EponSettings,
		TelemetryProfile:                  settings.TelemetryProfile,
		PartnerSettings:                   settings.PartnerSettings,
	}

	if settings.GroupName != "" {
		settingsResponse.GroupName = settings.GroupName
	} else {
		settingsResponse.GroupName = nil
	}
	if settings.ScheduleCron != "" {
		settingsResponse.ScheduleCron = settings.ScheduleCron
	} else {
		settingsResponse.ScheduleCron = nil
	}
	if settings.LusMessage != "" {
		settingsResponse.LusMessage = settings.LusMessage
	} else {
		settingsResponse.LusMessage = nil
	}
	if settings.LusName != "" {
		settingsResponse.LusName = settings.LusName
	} else {
		settingsResponse.LusName = nil
	}
	if settings.LusUploadRepositoryName != "" {
		settingsResponse.LusUploadRepositoryName = settings.LusUploadRepositoryName
	} else {
		settingsResponse.LusUploadRepositoryName = nil
	}
	if settings.LusScheduleCron != "" {
		settingsResponse.LusScheduleCron = settings.LusScheduleCron
	} else {
		settingsResponse.LusScheduleCron = nil
	}
	if settings.LusScheduleCronL1 != "" {
		settingsResponse.LusScheduleCronL1 = settings.LusScheduleCronL1
	} else {
		settingsResponse.LusScheduleCronL1 = nil
	}
	if settings.LusScheduleCronL2 != "" {
		settingsResponse.LusScheduleCronL2 = settings.LusScheduleCronL2
	} else {
		settingsResponse.LusScheduleCronL2 = nil
	}
	if settings.LusScheduleCronL3 != "" {
		settingsResponse.LusScheduleCronL3 = settings.LusScheduleCronL3
	} else {
		settingsResponse.LusScheduleCronL3 = nil
	}
	if settings.VodSettingsName != "" {
		settingsResponse.VodSettingsName = settings.VodSettingsName
	} else {
		settingsResponse.VodSettingsName = nil
	}
	if settings.LocationUrl != "" {
		settingsResponse.LocationUrl = settings.LocationUrl
	} else {
		settingsResponse.LocationUrl = nil
	}
	if settings.SrmIPList != nil && len(settings.SrmIPList) > 0 {
		settingsResponse.SrmIPList = settings.SrmIPList
	} else {
		settingsResponse.SrmIPList = nil
	}
	return settingsResponse
}

// DeviceSettings DeviceSettings2 table
type DeviceSettings struct {
	ID                      string                  `json:"id"`
	Updated                 int64                   `json:"updated"`
	Name                    string                  `json:"name"`
	CheckOnReboot           bool                    `json:"checkOnReboot"`
	ConfigurationServiceURL ConfigurationServiceURL `json:"configurationServiceURL"`
	SettingsAreActive       bool                    `json:"settingsAreActive"`
	Schedule                Schedule                `json:"schedule"`
	ApplicationType         string                  `json:"applicationType"`
}

func (obj *DeviceSettings) Clone() (*DeviceSettings, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*DeviceSettings), nil
}

// NewDeviceSettingsInf constructor
func NewDeviceSettingsInf() interface{} {
	return &DeviceSettings{}
}

const (
	MODE_TO_GET_LOG_FILES_0 = "LogFiles"
	MODE_TO_GET_LOG_FILES_1 = "LogFilesGroup"
	MODE_TO_GET_LOG_FILES_2 = "AllLogFiles"
)

// LogUploadSettings LogUploadSettings2 table
type LogUploadSettings struct {
	ID                  string   `json:"id"`
	Updated             int64    `json:"updated"`
	Name                string   `json:"name"`
	UploadOnReboot      bool     `json:"uploadOnReboot"`
	NumberOfDays        int      `json:"numberOfDays"`
	AreSettingsActive   bool     `json:"areSettingsActive"`
	Schedule            Schedule `json:"schedule"`
	LogFileIds          []string `json:"logFileIds"`
	LogFilesGroupID     string   `json:"logFilesGroupId"`
	ModeToGetLogFiles   string   `json:"modeToGetLogFiles"`
	UploadRepositoryID  string   `json:"uploadRepositoryId"`
	ActiveDateTimeRange bool     `json:"activeDateTimeRange"`
	FromDateTime        string   `json:"fromDateTime"`
	ToDateTime          string   `json:"toDateTime"`
	ApplicationType     string   `json:"applicationType"`
}

func (obj *LogUploadSettings) Clone() (*LogUploadSettings, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*LogUploadSettings), nil
}

// NewLogUploadSettingsInf constructor
func NewLogUploadSettingsInf() interface{} {
	return &LogUploadSettings{}
}

func GetOneDeviceSettings(id string) *DeviceSettings {
	var deviceSettings *DeviceSettings
	deviceSettingsInst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_DEVICE_SETTINGS, id)
	if err != nil {
		log.Debug(fmt.Sprintf("no deviceSettings found for Id: %s", id))
		return nil
	}
	deviceSettings = deviceSettingsInst.(*DeviceSettings)
	return deviceSettings
}

func GetOneLogUploadSettings(id string) *LogUploadSettings {
	var logUploadSettings *LogUploadSettings
	logUploadSettingsInst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_LOG_UPLOAD_SETTINGS, id)
	if err != nil {
		log.Debug(fmt.Sprintf("no logUploadSettings found for Id: %s", id))
		return nil
	}
	logUploadSettings = logUploadSettingsInst.(*LogUploadSettings)
	return logUploadSettings
}

func GetOneUploadRepository(id string) *UploadRepository {
	var uploadRepository *UploadRepository
	uploadRepositoryInst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_UPLOAD_REPOSITORY, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no uploadRepository found for Id: %s", id))
		return nil
	}
	uploadRepository = uploadRepositoryInst.(*UploadRepository)
	return uploadRepository
}

func GetLogFileList(size int) []*LogFile {
	var logFiles []*LogFile
	logFileListInst, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_LOG_FILE, size)
	if err != nil {
		log.Warn(fmt.Sprintf("no logFiles found "))
		return nil
	}
	for idx := range logFileListInst {
		logFile := logFileListInst[idx].(*LogFile)
		logFiles = append(logFiles, logFile)
	}
	return logFiles
}

func GetOneVodSettings(id string) *VodSettings {
	var vodSettings *VodSettings
	vodSettingsInst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_VOD_SETTINGS, id)
	if err != nil {
		log.Debug(fmt.Sprintf("no vodSettings found for Id: %s", id))
		return nil
	}
	vodSettings = vodSettingsInst.(*VodSettings)
	return vodSettings
}

func GetOneLogFileList(id string) (*LogFileList, error) {
	var logFileList *LogFileList
	logFileListInst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_LOG_FILE_LIST, id)
	if err != nil {
		log.Warn(fmt.Sprintf("no LogFileList found for Id: %s", id))
		return nil, err
	}
	logFileList = logFileListInst.(*LogFileList)
	return logFileList, nil
}
