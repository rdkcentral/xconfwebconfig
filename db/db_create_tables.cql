/**
 * Copyright 2023 Comcast Cable Communications Management, LLC
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

CREATE TABLE IF NOT EXISTS "DcmRule" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "DeviceSettings2" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "Environment" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "FeatureControlRule2" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "FirmwareConfig" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "FirmwareRule4" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "FirmwareRuleTemplate" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "GenericXconfNamedList" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "LogFile" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "LogFileList" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "LogFilesGroups" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "LogUploadSettings2" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "Logs2" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "Model" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "PermanentTelemetry" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "SettingProfiles" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "SettingRules" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "SingletonFilterValue" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "Telemetry" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "TelemetryRules" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "UploadRepository" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "VodSettings2" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "XconfApprovedChange" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "XconfChange" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "XconfFeature" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "XconfChangedKeys4" (key bigint, columnName timeuuid, value blob, PRIMARY KEY (key, columnName));

CREATE TABLE IF NOT EXISTS "TelemetryTwoProfiles" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "TelemetryTwoRules" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "XconfTelemetryTwoChange" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "XconfApprovedTelemetryTwoChange" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "AppSettings" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "RecookingStatus" (module text, partition_id text, state int, updated_time timestamp, primary key (module, partition_id));

CREATE TABLE IF NOT EXISTS "PenetrationMetrics" (estb_mac text, ecm_mac text, serial_number text, partner text, model text, fw_filename text, fw_version text, fw_reported_version text, fw_additional_version_info text, fw_applied_rule text, rfc_applied_rules text, rfc_features text, rfc_ts timestamp, fw_ts timestamp, time_zone text, rfc_account_hash text, rfc_account_id text, rfc_account_mgmt text, AccountService_account_id text, rfc_partner text, AccountService_partner text, rfc_model text, rfc_fw_reported_version text, rfc_env text, rfc_application_type text, rfc_experience text, rfc_time_zone text, precook_rfc_rules text, rfc_configsethash text, precook_configsethash text, precook_rfc_features text, rfc_post_proc text, rfc_query_params text, rfc_tags text, rfc_estb_ip text, client_cert_expiry text, recovery_cert_expiry text, PRIMARY KEY (estb_mac));

CREATE TABLE IF NOT EXISTS "Tag" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "Locks" (name text, locked_by text, locked_at timestamp, expires_at timestamp, PRIMARY KEY (name));

CREATE TABLE IF NOT EXISTS "TagMembersBucketed" (
    tag_id TEXT,
    bucket_id INT,
    member TEXT,
    created BIGINT,
    PRIMARY KEY ((tag_id, bucket_id), member)
) WITH CLUSTERING ORDER BY (member ASC)
    AND compaction = {'class': 'SizeTieredCompactionStrategy'}
    AND compression = {'sstable_compression': 'LZ4Compressor'};

CREATE TABLE IF NOT EXISTS "TagBucketMetadata" (
    tag_id TEXT,
    bucket_id INT,
    PRIMARY KEY (tag_id, bucket_id)
);