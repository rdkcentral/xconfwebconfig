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

-- xconf KEYSPACE
USE xconf;

CREATE TABLE IF NOT EXISTS application_types (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS models (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS environments (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS generic_named_lists (tenant_id text, shard_id int, key text, column1 text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key, column1));
 
CREATE TABLE IF NOT EXISTS app_settings (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));

CREATE TABLE IF NOT EXISTS feature_control_rules (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
   
CREATE TABLE IF NOT EXISTS features (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS firmware_rules (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));

CREATE TABLE IF NOT EXISTS firmware_configs (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));

CREATE TABLE IF NOT EXISTS firmware_rule_templates (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS singleton_filter_values (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));

CREATE TABLE IF NOT EXISTS dcm_rules (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS upload_repositories (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS log_upload_settings (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS log_files (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS log_file_lists (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
  
CREATE TABLE IF NOT EXISTS device_settings (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS vod_settings (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS setting_rules (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS setting_profiles (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS telemetry_rules (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS telemetry_profiles (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS telemetry_two_rules (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS telemetry_two_profiles (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS permanent_telemetry_profiles (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS telemetry_changes (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS telemetry_approved_changes (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS telemetry_two_changes (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS telemetry_approved_two_changes (tenant_id text, shard_id int, key text, value blob, updated timestamp, PRIMARY KEY ((tenant_id, shard_id), key));
 
CREATE TABLE IF NOT EXISTS recooking_status (tenant_id text, shard_id int, module text, partition_id text, state int, updated_time timestamp, PRIMARY KEY ((tenant_id, shard_id), module, partition_id));
 
CREATE TABLE IF NOT EXISTS locks (tenant_id text, shard_id int, name text, locked_by text, locked_at timestamp, expires_at timestamp, PRIMARY KEY ((tenant_id, shard_id), name));
 
CREATE TABLE IF NOT EXISTS tag_member_metadata (tenant_id text, shard_id int, tag_id text, bucket_id int, PRIMARY KEY ((tenant_id, shard_id), tag_id, bucket_id));
 
CREATE TABLE IF NOT EXISTS tag_members_bucketed (tenant_id text, tag_id text, bucket_id int, member text, PRIMARY KEY ((tag_id, bucket_id), member));

CREATE TABLE IF NOT EXISTS config_change_logs (tenant_id text, shard_id int, key text, column1 text, value blob, PRIMARY KEY ((tenant_id, shard_id), key, column1));

CREATE TABLE IF NOT EXISTS change_events (key bigint, columnName timeuuid, value blob, PRIMARY KEY (key, columnName));

CREATE TABLE IF NOT EXISTS penetration_data (estb_mac text, ecm_mac text, serial_number text, partner text, model text, fw_filename text, fw_version text, fw_reported_version text, fw_additional_version_info text, fw_applied_rule text, rfc_applied_rules text, rfc_features text, rfc_ts timestamp, fw_ts timestamp, time_zone text, rfc_account_hash text, rfc_account_id text, rfc_account_mgmt text, titan_account_id text, rfc_partner text, titan_partner text, rfc_model text, rfc_fw_reported_version text, rfc_env text, rfc_application_type text, rfc_experience text, rfc_time_zone text, precook_rfc_rules text, rfc_configsethash text, precook_configsethash text, precook_rfc_features text, rfc_post_proc text, rfc_query_params text, rfc_tags text, rfc_estb_ip text, client_cert_expiry text, recovery_cert_expiry text, PRIMARY KEY (estb_mac));

-- ApplicationsDiscoveryDataService KEYSPACE (backward compatibility)

USE "ApplicationsDiscoveryDataService";

CREATE TABLE IF NOT EXISTS "Logs2" (key text, column1 text, value blob, PRIMARY KEY ((key), column1));

CREATE TABLE IF NOT EXISTS "PenetrationMetrics" (estb_mac text, ecm_mac text, serial_number text, partner text, model text, fw_filename text, fw_version text, fw_reported_version text, fw_additional_version_info text, fw_applied_rule text, rfc_applied_rules text, rfc_features text, rfc_ts timestamp, fw_ts timestamp, time_zone text, rfc_account_hash text, rfc_account_id text, rfc_account_mgmt text, titan_account_id text, rfc_partner text, titan_partner text, rfc_model text, rfc_fw_reported_version text, rfc_env text, rfc_application_type text, rfc_experience text, rfc_time_zone text, precook_rfc_rules text, rfc_configsethash text, precook_configsethash text, precook_rfc_features text, rfc_post_proc text, rfc_query_params text, rfc_tags text, rfc_estb_ip text, client_cert_expiry text, recovery_cert_expiry text, PRIMARY KEY (estb_mac));
