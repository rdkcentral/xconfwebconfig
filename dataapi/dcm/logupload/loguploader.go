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
	"math"
	"sort"
	"strings"

	"xconfwebconfig/common"
	"xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared/logupload"

	log "github.com/sirupsen/logrus"
)

type LogUploadRuleBase struct {
	//DcmRuleDAO           ds.CachedSimpleDao
	RuleProcessorFactory re.RuleProcessorFactory
}

func NewLogUploadRuleBase() *LogUploadRuleBase {
	return &LogUploadRuleBase{
		RuleProcessorFactory: *re.NewRuleProcessorFactory(),
	}
}

func (l *LogUploadRuleBase) Eval(context map[string]string, fields log.Fields) *logupload.Settings {
	settings := logupload.NewSettings(1)
	rules := l.getSortedDcmRules()
	for _, rule := range rules {
		if string(rule.ApplicationType) == context[common.APPLICATION_TYPE] && l.RuleProcessorFactory.RuleProcessor().Evaluate(&rule.Rule, context, log.Fields{}) {
			logupload.CopySettings(settings, l.GetSettings(rule.ID), rule, context, fields)
		}
		if settings.AreFull() {
			return settings
		}
	}
	if len(settings.GroupName) > 0 || len(settings.VodSettingsName) > 0 {
		return settings
	}
	return nil
}

func (l *LogUploadRuleBase) getSortedDcmRules() []*logupload.DCMGenericRule {
	cm := db.GetCacheManager()
	cacheKey := "DCMGenericRuleListSorted"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_DCM_RULE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*logupload.DCMGenericRule)
	}

	all := logupload.GetDCMGenericRuleList()
	if len(all) <= 1 {
		return all
	}

	var sortedList []*logupload.DCMGenericRule
	sortedList = append(sortedList, all...)

	sort.Slice(sortedList, func(i, j int) bool {
		return sortedList[i].Priority < sortedList[j].Priority
	})

	cm.ApplicationCacheSet(db.TABLE_DCM_RULE, cacheKey, sortedList)

	return sortedList
}

var GetOneDeviceSettingsFunc = logupload.GetOneDeviceSettings
var GetOneLogUploadSettingsFunc = logupload.GetOneLogUploadSettings
var GetOneUploadRepositoryFunc = logupload.GetOneUploadRepository
var GetLogFileListFunc = logupload.GetLogFileList
var GetOneVodSettingsFunc = logupload.GetOneVodSettings

func (l *LogUploadRuleBase) GetSettings(id string) *logupload.Settings {
	settings := logupload.NewSettings(1)
	var err error
	var deviceSettings logupload.DeviceSettings
	var deviceSettingsOK = false
	deviceSettingsPointer := GetOneDeviceSettingsFunc(id)
	if deviceSettingsPointer != nil {
		deviceSettings = *deviceSettingsPointer
		deviceSettingsOK = true
	}

	var logUploadSettings logupload.LogUploadSettings
	logUploadSettingsPointer := GetOneLogUploadSettingsFunc(id)
	var logUploadSettingsOK = false
	if logUploadSettingsPointer != nil {
		logUploadSettings = *logUploadSettingsPointer
		logUploadSettingsOK = true
	}

	if deviceSettingsOK && logUploadSettingsOK && deviceSettings.SettingsAreActive && logUploadSettings.AreSettingsActive {
		settings.GroupName = deviceSettings.Name
		settings.CheckOnReboot = deviceSettings.CheckOnReboot
		settings.ConfigurationServiceURL = deviceSettings.ConfigurationServiceURL.URL
		checkSchedule := deviceSettings.Schedule
		if checkSchedule.TimeZone == logupload.LOCAL_TIME {
			settings.TimeZoneMode = logupload.LOCAL_TIME
		} else {
			settings.TimeZoneMode = logupload.UTC
		}
		settings.ScheduleCron = checkSchedule.Expression
		settings.ScheduleDurationMinutes = checkSchedule.TimeWindowMinutes
		settings.LusName = logUploadSettings.Name
		settings.LusNumberOfDay = logUploadSettings.NumberOfDays
		uploadRepositoryId := logUploadSettings.UploadRepositoryID
		if len(uploadRepositoryId) > 0 {
			var uploadRepository logupload.UploadRepository
			uploadRepositoryPointer := GetOneUploadRepositoryFunc(uploadRepositoryId)
			if uploadRepositoryPointer != nil {
				uploadRepository = *uploadRepositoryPointer
				settings.LusUploadRepositoryName = uploadRepository.Name
				protocol := uploadRepository.Protocol
				url := uploadRepository.URL
				if len(protocol) < 1 || strings.Contains(url, "://") {
					settings.LusUploadRepositoryURL = url
				} else {
					settings.LusUploadRepositoryURL = strings.ToLower(protocol) + "://" + url
				}
				settings.LusUploadRepositoryURLNew = uploadRepository.URL
				settings.LusUploadRepositoryUploadProtocol = uploadRepository.Protocol
			}
		}
		settings.LusUploadOnReboot = logUploadSettings.UploadOnReboot
		if len(logUploadSettings.ModeToGetLogFiles) > 0 {
			var listLogFilesForLogUplSettings []*logupload.LogFile
			var listLogFilesOK = true
			if logUploadSettings.ModeToGetLogFiles == logupload.MODE_TO_GET_LOG_FILES_0 {
				listLogFilesForLogUplSettings, err = getLogFileList(logUploadSettings.ID, math.MaxInt32/100)
				if err != nil {
					listLogFilesOK = false
				}
			} else if logUploadSettings.ModeToGetLogFiles == logupload.MODE_TO_GET_LOG_FILES_1 {
				keyFileGroup := logUploadSettings.LogFilesGroupID
				if len(keyFileGroup) > 0 {
					listLogFilesForLogUplSettings, err = getLogFileList(keyFileGroup, math.MaxInt32/100)
					if err != nil {
						listLogFilesOK = false
					}
				}
			} else if logUploadSettings.ModeToGetLogFiles == logupload.MODE_TO_GET_LOG_FILES_2 {
				// logFiles []*logupload.LogFile
				logFiles := GetLogFileListFunc(math.MaxInt32 / 100)
				if logFiles == nil {
					log.Warn(fmt.Sprintf("no logFiles found "))
					listLogFilesOK = false
				} else {
					listLogFilesForLogUplSettings = logFiles
				}
			}
			if listLogFilesOK {
				settings.LusLogFiles = listLogFilesForLogUplSettings
			}
		}
		uploadSchedule := logUploadSettings.Schedule
		if uploadSchedule.TimeZone == logupload.LOCAL_TIME {
			settings.LusTimeZoneMode = logupload.LOCAL_TIME
		} else {
			settings.LusTimeZoneMode = logupload.UTC
		}
		settings.LusScheduleCron = uploadSchedule.Expression
		settings.LusScheduleCronL1 = uploadSchedule.ExpressionL1
		settings.LusScheduleCronL2 = uploadSchedule.ExpressionL2
		settings.LusScheduleCronL3 = uploadSchedule.ExpressionL3
		settings.LusScheduleDurationMinutes = uploadSchedule.TimeWindowMinutes
		settings.LusTimeZoneMode = uploadSchedule.TimeZone
		settings.SchedulerType = uploadSchedule.Type
	}
	var vodSettings logupload.VodSettings
	vodSettingsPointer := GetOneVodSettingsFunc(id)
	if vodSettingsPointer != nil {
		vodSettings = *vodSettingsPointer
		settings.VodSettingsName = vodSettings.Name
		settings.LocationUrl = vodSettings.LocationsURL
		settings.SrmIPList = vodSettings.SrmIPList
	}
	return settings
}

var GetOneLogFileListFunc = logupload.GetOneLogFileList

func getLogFileList(id string, maxResults int) ([]*logupload.LogFile, error) {
	var logFileList logupload.LogFileList
	logFileListPointer, err := GetOneLogFileListFunc(id)
	if logFileListPointer == nil {
		return nil, err
	}
	logFileList = *logFileListPointer
	if len(logFileList.Data) < maxResults {
		return logFileList.Data, nil
	}
	return logFileList.Data[:maxResults-1], nil
}

var ruleBase = LogUploadRuleBase{}
var EvalFunc = ruleBase.Eval
var GetOneDcmRuleFunc = logupload.GetOneDCMGenericRule
