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

CREATE TABLE IF NOT EXISTS "ApprovedChange" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "Change" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "ConfigurationServiceURLs" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "DcmRule" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "DeviceSettings2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "Environment" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "FeatureControlRule2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "FeatureSet" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "FirmwareConfig" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "FirmwareRule4" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "FirmwareRuleTemplate" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "GenericXconfNamedList" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "IndexedLogFiles" (key text,column1 text,column2 text,value blob,PRIMARY KEY ((key), column1, column2)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "IpAddressGroupExtended" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "LogFile" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "LogFileList" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "LogFilesGroups" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "LogUploadSettings2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "Logs2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "Model" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "PermanentTelemetry" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "SettingProfiles" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "SettingRules" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "SingletonFilterValue" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "Telemetry" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "TelemetryRules" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "UploadRepository" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "VodSettings2" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "XconfApprovedChange" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "XconfChange" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "XconfFeature" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "XconfNamedList" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "XconfChangedKeys4" ( key bigint, columnName timeuuid, value blob, PRIMARY KEY (key, columnName) ) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "TelemetryTwoProfiles" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "TelemetryTwoRules" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "XconfTelemetryTwoChange" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;

CREATE TABLE IF NOT EXISTS "XconfApprovedTelemetryTwoChange" (key text,column1 text,value blob,PRIMARY KEY ((key), column1)) WITH COMPACT STORAGE;