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

 CREATE TABLE IF NOT EXISTS "AppSettings" (key text,column1 text,value blob,PRIMARY KEY ((key), column1));

CREATE TABLE "DcmRule" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "DeviceSettings2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "Environment" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "FeatureControlRule2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "FirmwareConfig" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "FirmwareRule4" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "FirmwareRuleTemplate" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "GenericXconfNamedList" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "LogFile" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "LogFileList" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "LogFilesGroups" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "LogUploadSettings2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "Logs2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "Model" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "PermanentTelemetry" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "SettingProfiles" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "SettingRules" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "SingletonFilterValue" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "Telemetry" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "TelemetryRules" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "UploadRepository" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "VodSettings2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "XconfApprovedChange" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "XconfChange" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "XconfFeature" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "XconfChangedKeys4" ( key bigint, columnName timeuuid, value blob, PRIMARY KEY (key, columnName) ) WITH COMPACT STORAGE;

CREATE TABLE "TelemetryTwoProfiles" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE "TelemetryTwoRules" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "XconfTelemetryTwoChange" (key text,column1 text,value blob,PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "XconfApprovedTelemetryTwoChange" (key text,column1 text,value blob,PRIMARY KEY ((key), column1));

CREATE TABLE "PenetrationMetrics" (estb_mac text,partner text,model text,fw_version text,fw_reported_version text,fw_additional_version_info text,fw_applied_rule text,rfc_applied_rules text,rfc_features text,rfc_ts timestamp,fw_ts timestamp,time_zone text,account_id text,env text,application_type text,experience text, PRIMARY KEY (estb_mac)) WITH COMPACT STORAGE; 