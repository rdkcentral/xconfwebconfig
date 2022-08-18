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
package common

const (
	LOGGING_TIME_FORMAT = "2006-01-02 15:04:05.000"
	DATE_TIME_FORMATTER = "1/2/2006 15:04"
	LAST_CONFIG_LOG_ID  = "0"
)

const (
	XCONF_HTTP_HEADER       = "HA-Haproxy-xconf-http"
	XCONF_HTTP_VALUE        = "xconf-http"
	X_FORWARDED_FOR_HEADER  = "X-Forwarded-For"
	HA_FORWARDED_FOR_HEADER = "HA-Forwarded-For"

	HeaderAuthorization    = "Authorization"
	HeaderUserAgent        = "User-Agent"
	HeaderIfNoneMatch      = "If-None-Match"
	HeaderFirmwareVersion  = "X-System-Firmware-Version"
	HeaderModelName        = "X-System-Model-Name"
	HeaderProfileVersion   = "X-System-Telemetry-Profile-Version"
	HeaderPartnerID        = "X-System-PartnerID"
	HeaderAccountID        = "X-System-AccountID"
	HeaderXconfDataService = "XconfDataService"
)

const (
	HOST_MAC_PARAM = "hostMac"
	ECM_MAC_PARAM  = "ecmMac"

	ID                         = "id"
	IP_ADDRESS                 = "ipAddress"
	ESTB_IP                    = "estbIP"
	ESTB_MAC_ADDRESS           = "estbMacAddress"
	ESTB_MAC                   = "eStbMac"
	ECM_MAC_ADDRESS            = "ecmMacAddress"
	ECM_MAC                    = "eCMMac"
	ENV                        = "env"
	MODEL                      = "model"
	MODEL_ID                   = "modelId"
	ACCOUNT_MGMT               = "accountMgmt"
	SERIAL_NUM                 = "serialNum"
	PARTNER_ID                 = "partnerId"
	PASSED_PARTNER_ID          = "passedPartnerId"
	FIRMWARE_VERSION           = "firmwareVersion"
	RECEIVER_ID                = "receiverId"
	CONTROLLER_ID              = "controllerId"
	CHANNEL_MAP_ID             = "channelMapId"
	VOD_ID                     = "vodId"
	BYPASS_FILTERS             = "bypassFilters"
	FORCE_FILTERS              = "forceFilters"
	UPLOAD_IMMEDIATELY         = "uploadImmediately"
	TIME_ZONE                  = "timezone"
	TIME_ZONE_OFFSET           = "timeZoneOffset"
	TIME                       = "time"
	APPLICATION_TYPE           = "applicationType"
	ACCOUNT_ID                 = "accountId"
	ACCOUNT_HASH               = "accountHash"
	CONFIG_SET_HASH            = "configSetHash"
	SYNDICATION_PARTNER        = "SyndicationPartner"
	MAC                        = "mac"
	CHECK_NOW                  = "checkNow"
	VERSION                    = "version"
	SETTING_TYPE               = "settingType"
	TABLE_NAME                 = "tableName"
	FIELD                      = "field"
	NAME                       = "name"
	LIST_ID                    = "listId"
	DOWNLOAD_PROTOCOL          = "firmware_download_protocol"
	REBOOT_DECOUPLED           = "rebootDecoupled"
	MATCHED_RULE_TYPE          = "matchedRuleType"
	CAPABILITIES               = "capabilities"
	UPDATED                    = "updated"
	DESCRIPTION                = "description"
	SUPPORTED_MODEL_IDS        = "supportedModelIds"
	FIRMWARE_DOWNLOAD_PROTOCOL = "firmwareDownloadProtocol"
	FIRMWARE_FILENAME          = "firmwareFilename"
	FIRMWARE_LOCATION          = "firmwareLocation"
	IPV6_FIRMWARE_LOCATION     = "ipv6FirmwareLocation"
	UPGRADE_DELAY              = "upgradeDelay"
	REBOOT_IMMEDIATELY         = "rebootImmediately"
	PROPERTIES                 = "properties"
	MANDATORY_UPDATE           = "mandatoryUpdate"
	FIRMWARE_VERSIONS          = "firmwareVersions"
	REGULAR_EXPRESSIONS        = "regularExpressions"
)

const (
	TR181_DEVICE_TYPE_PARTNER_ID   = "tr181.Device.DeviceInfo.X_RDKCENTRAL-COM_Syndication.PartnerId"
	TR181_DEVICE_TYPE_ACCOUNT_ID   = "tr181.Device.DeviceInfo.X_RDKCENTRAL-COM_RFC.Feature.AccountInfo.AccountID"
	TR181_DEVICE_TYPE_ACCOUNT_HASH = "tr181.Device.DeviceInfo.X_RDKCENTRAL-COM_RFC.Feature.MD5AccountHash"
)

const (
	GenericNamespacedListTypes_STRING      = "STRING"
	GenericNamespacedListTypes_MAC_LIST    = "MAC_LIST"
	GenericNamespacedListTypes_IP_LIST     = "IP_LIST"
	GenericNamespacedListTypes_RI_MAC_LIST = "RI_MAC_LIST"
)

func isValidType(namespacedListType string) bool {
	if namespacedListType == GenericNamespacedListTypes_STRING ||
		namespacedListType == GenericNamespacedListTypes_MAC_LIST ||
		namespacedListType == GenericNamespacedListTypes_IP_LIST ||
		namespacedListType == GenericNamespacedListTypes_RI_MAC_LIST {
		return true
	}
	return false
}

var (
	CacheUpdateWindowSize int64

	BinaryVersion   = ""
	BinaryBranch    = ""
	BinaryBuildTime = ""

	DefaultIgnoredHeaders = []string{
		"Accept",
		"User-Agent",
		"Authorization",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-B3-Sampled",
		"X-B3-Spanid",
		"X-B3-Traceid",
		"X-Envoy-Decorator-Operation",
		"X-Envoy-External-Address",
		"X-Envoy-Peer-Metadata",
		"X-Envoy-Peer-Metadata-Id",
		"X-Forwarded-Proto",
	}
)
