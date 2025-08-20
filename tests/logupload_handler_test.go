/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
package tests

import (
	"github.com/rdkcentral/xconfwebconfig/dataapi"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/gorilla/mux"
)

const (
	S3_PATH                  = "cgi-bin/s3.cgi"
	RDKB_SNMP                = "cgi-bin/rdkb_snmp.cgi"
	RDKVLOGUPLOAD            = "cgi-bin/rdkvlogupload.cgi"
	UPLOAD_DUMP              = "cgi-bin/upload_dump.cgi"
	RDKB                     = "cgi-bin/rdkb.cgi"
	EXPIRATION_TIME_IN_HOURS = 1
)

type UrlSecurityTokenTest struct {
	name                string
	ssrPath             string
	url                 string
	apiVersion          string
	mac                 string
	ip                  string
	tokenEnabled        bool
	groupServiceEnabled bool
}

func setUpXconfServerWithLoguploaderSecurityConfig(localServer *xwhttp.XconfServer, ssrPath, securityKey string, tokenEnabled bool, groupServiceEnabled bool) *mux.Router {
	localServer.SecurityTokenConfig = createSecurityTokenConfig(securityKey, groupServiceEnabled)
	localServer.LogUploadSecurityTokenConfig = createSecurityPathConfig(ssrPath, tokenEnabled)
	localRouter := localServer.GetRouter(true)
	dataapi.XconfSetup(localServer, localRouter)
	// xwhttp.InitSatTokenManager(server.XconfServer, true)
	return localRouter
}

func createSecurityPathConfig(path string, enabled bool) *xwhttp.SecurityTokenPathConfig {
	configMap := map[string]bool{path: enabled}
	return &xwhttp.SecurityTokenPathConfig{UrlPathMap: configMap}
}

func createSecurityTokenConfig(securityKey string, groupServiceEnabled bool) *xwhttp.SecurityTokenConfig {
	return &xwhttp.SecurityTokenConfig{
		SkipSecurityTokenClientProtocolSet: util.NewSet(),
		SecurityTokenKey:                   securityKey,
		SecurityTokenGroupServiceEnabled:   groupServiceEnabled,
	}
}

func preCreateFormula(id string, priority int, uploadRepoUrl string, rule rulesengine.Rule) {
	uploadRepo := createUploadRepository(id, uploadRepoUrl)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_UPLOAD_REPOSITORY, uploadRepo.ID, uploadRepo)

	deviceSetting := createDeviceSettings(id)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_DEVICE_SETTINGS, deviceSetting.ID, deviceSetting)

	logUploadSettings := createLogUploadSettings(id, uploadRepo.ID)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_LOG_UPLOAD_SETTINGS, logUploadSettings.ID, logUploadSettings)

	vodSettings := createVodSettings(id)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_VOD_SETTINGS, vodSettings.ID, vodSettings)

	dcmRule := createDcmRule(id, priority, rule)
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_DCM_RULE, dcmRule.ID, dcmRule)
}

func createDcmRule(id string, priority int, rule rulesengine.Rule) *logupload.DCMGenericRule {
	return &logupload.DCMGenericRule{ID: id, Name: "name-" + id, Description: "description" + id, ApplicationType: shared.STB, Priority: priority, Rule: rule, Percentage: 100}
}

func createUploadRepository(id, url string) *logupload.UploadRepository {
	return &logupload.UploadRepository{
		ID:              id,
		Name:            "name-" + id,
		Description:     "description" + id,
		URL:             url,
		ApplicationType: shared.STB,
		Protocol:        shared.Http,
	}
}

func createDeviceSettings(id string) *logupload.DeviceSettings {
	return &logupload.DeviceSettings{
		ID:                id,
		Name:              "DeviceSettings-" + id,
		CheckOnReboot:     false,
		SettingsAreActive: true,
		Schedule: logupload.Schedule{
			Type:              "CronExpression",
			Expression:        "15 0 * * *",
			TimeZone:          "Local time",
			TimeWindowMinutes: "10",
		},
		ApplicationType: shared.STB,
	}
}

func createLogUploadSettings(id, uploadRepoId string) *logupload.LogUploadSettings {
	return &logupload.LogUploadSettings{
		ID:                id,
		Name:              "LogUploadSettings-" + id,
		UploadOnReboot:    false,
		NumberOfDays:      10,
		AreSettingsActive: true,
		Schedule: logupload.Schedule{
			Type:              "CronExpression",
			Expression:        "3 0 * * *",
			TimeZone:          "UTC",
			TimeWindowMinutes: "1",
		},
		UploadRepositoryID: uploadRepoId,
		ApplicationType:    shared.STB,
	}
}

func createVodSettings(id string) *logupload.VodSettings {
	return &logupload.VodSettings{
		ID:              id,
		Name:            "VodSettings-" + id,
		LocationsURL:    "https://test.xcal.tv",
		IPNames:         []string{},
		IPList:          []string{},
		ApplicationType: "",
	}
}
