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

const (
	SampleTelemetryTwoRulesString = `[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"TESTMODEL"}}}},"compoundParts":[],"id":"84a53ad7-016d-4c55-81b1-92ae1f16f2ee","name":"Scout Rule 1","boundTelemetryIds":["4b84ffce-812a-4074-ba56-18982106f2f8"],"applicationType":"stb"},{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"estbMacAddress"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"84:E0:58:57:53:F0"}}}},"compoundParts":[],"id":"3a7ad3cd-44e9-41a0-87cf-1d94803f3db2","name":"Test Rule with address only","boundTelemetryIds":["f8f4e7c7-924a-4a00-8ecf-04941dd6c4a3","234b46be-0d9d-40f4-8b6c-d8e3a94f64d2","9fbf4f56-301b-4a28-8966-090cd38b498e","2370b5b1-6899-44b8-bb0d-f3706f9389b9","cec11c05-ea4d-45cd-84e0-e44cecfafbf6"],"applicationType":"stb"},{"negated":false,"compoundParts":[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"TESTMODEL"}}}},"compoundParts":[]},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"estbMacAddress"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"84:E0:58:57:53:F0"}}}},"compoundParts":[]}],"id":"a59371b4-8365-484f-9c61-286c78e5386e","name":"Test Rule with many Profiles","boundTelemetryIds":["5d298496-b108-4884-9713-1e51c843287b","8c65c89d-dc11-4842-9000-b9fa6f45f34a","e6ddbe95-daec-49db-8e72-123d53dbe630","f8f4e7c7-924a-4a00-8ecf-04941dd6c4a3","cec11c05-ea4d-45cd-84e0-e44cecfafbf6","05d7bb24-e30f-456b-84c1-55d2a20eddec","2370b5b1-6899-44b8-bb0d-f3706f9389b9","3bbf957d-c61b-4137-8800-634b9ef6013f","495f3ead-576c-4b09-9c47-8b85298a7d76","07cd2a04-7083-44f2-a9d4-23823aed9c42","7eec6e18-0937-4a55-b16b-be4ea2219aa1","4397b229-200a-471f-9b46-41a19960ef18","3a3ba25c-febd-40ac-8e38-b302aa428d69"],"applicationType":"stb"},{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"CGM4140COM"}}}},"compoundParts":[],"id":"55b2419a-2595-4c7b-89a3-c861a1b87f79","name":"webconfig_red_rule_CGM4140COM","boundTelemetryIds":["3586d1d0-b3d3-4304-9a26-85d497d3ea3d"],"applicationType":"stb"},{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"CGA4131COM"}}}},"compoundParts":[],"id":"a9566e59-9eb8-4127-8cfb-2398ce0b6605","name":"WHiX","boundTelemetryIds":["8c65c89d-dc11-4842-9000-b9fa6f45f34a"],"applicationType":"stb"},{"negated":false,"compoundParts":[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"estbMacAddress"},"operation":"IN_LIST","fixedArg":{"bean":{"value":{"java.lang.String":"00031"}}}},"compoundParts":[]},{"negated":false,"relation":"OR","condition":{"freeArg":{"type":"STRING","name":"env"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"WTEST"}}}},"compoundParts":[]},{"negated":false,"relation":"OR","condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"WTEST"}}}},"compoundParts":[]},{"negated":false,"relation":"OR","condition":{"freeArg":{"type":"STRING","name":"partnerId"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"WSMITHPART"}}}},"compoundParts":[]},{"negated":false,"relation":"OR","condition":{"freeArg":{"type":"STRING","name":"accountId"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"112233445566"}}}},"compoundParts":[]},{"negated":false,"relation":"OR","condition":{"freeArg":{"type":"STRING","name":"randomParam"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"1234"}}}},"compoundParts":[]}],"id":"a8ae8db8-0cfc-420d-b5a0-2036a7bbc8a7","name":"wsmithT2.0Rule","boundTelemetryIds":["9fbf4f56-301b-4a28-8966-090cd38b498e"],"applicationType":"stb"},{"negated":false,"compoundParts":[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"comp"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"test"}}}},"compoundParts":[]},{"negated":true,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"env"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"AA"}}}},"compoundParts":[]}],"id":"d789b29f-d9e5-41fd-9c81-1e8604f5dd57","name":"wsmithT2.0Rule2","boundTelemetryIds":["9fbf4f56-301b-4a28-8966-090cd38b498e"],"applicationType":"stb"},{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"partnerId"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"WSMITHPART2"}}}},"compoundParts":[],"id":"bb24df8f-44ad-4324-8049-1d84f5293594","name":"wsmithT2.0Rule3","boundTelemetryIds":["8c65c89d-dc11-4842-9000-b9fa6f45f34a","9fbf4f56-301b-4a28-8966-090cd38b498e"],"applicationType":"stb"},{"negated":false,"compoundParts":[{"negated":false,"condition":{"freeArg":{"type":"ANY","name":"wsmithtag"},"operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}},"compoundParts":[]},{"negated":false,"relation":"OR","condition":{"freeArg":{"type":"ANY","name":"wsmithpartner1"},"operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}},"compoundParts":[]},{"negated":false,"relation":"OR","condition":{"freeArg":{"type":"ANY","name":"wrfctag1"},"operation":"EXISTS","fixedArg":{"bean":{"value":{"java.lang.String":""}}}},"compoundParts":[]}],"id":"7ec4a30a-839f-4619-a5cf-82abb219bbf2","name":"wsmithT2.0TagRule","boundTelemetryIds":["234b46be-0d9d-40f4-8b6c-d8e3a94f64d2"],"applicationType":"stb"},{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"estbMacAddress"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"77:88:99:AA:BB:CC"}}}},"compoundParts":[],"id":"8ab5a6be-5e7b-4188-87d5-9bc173679d93","name":"xpc_dev_rule_003","boundTelemetryIds":["24f29c66-b658-40fd-a39d-dda8c670f3eb"],"applicationType":"stb"},{"negated":false,"compoundParts":[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"MINION_FW_201"}}}},"compoundParts":[]},{"negated":false,"relation":"AND","condition":{"freeArg":{"type":"STRING","name":"foo"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"bar"}}}},"compoundParts":[]}],"id":"92147d7a-75fb-46a6-9292-4d2a96c8ab71","name":"xpc_dev_rule_101","boundTelemetryIds":["24f29c66-b658-40fd-a39d-dda8c670f3eb"],"applicationType":"stb"},{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"model"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"TG1682G"}}}},"compoundParts":[],"id":"281e789d-d182-42a8-95ce-4fc1a8814cdd","name":"xpc_test_rule_004","boundTelemetryIds":["3586d1d0-b3d3-4304-9a26-85d497d3ea3d"],"applicationType":"stb"}]`
	ExtraRuleString1              = `[{"negated":false,"condition":{"freeArg":{"type":"STRING","name":"estbMacAddress"},"operation":"IS","fixedArg":{"bean":{"value":{"java.lang.String":"33:33:00:00:11:11"}}}},"compoundParts":[],"id":"69517b66-8514-4c2e-8483-f34a865054a2","name":"xpc_test_rule_301","boundTelemetryIds":["819a1b27-1aad-4663-ba4f-bfefcdf020ef"],"applicationType":"stb"}]`
	IllFormattedProfileUuid       = "819a1b27-1aad-4663-ba4f-bfefcdf020ef"
	IllFormattedProfileName       = "ill_formatted_profile_1"
)

var (
	SampleProfileIdNameMap = map[string]string{
		"05d7bb24-e30f-456b-84c1-55d2a20eddec": "jw_t2_wifi_part_1",
		"3bbf957d-c61b-4137-8800-634b9ef6013f": "jw_t2_wifi_part_3",
		"b54624d7-1a2d-461b-a1cc-3591ff893b7e": "TestSchema",
		"495f3ead-576c-4b09-9c47-8b85298a7d76": "jw_t2_wifi_part_4",
		"2a077060-b981-4fb6-ad75-72f18b4de130": "xpc_test_profile_002",
		"2370b5b1-6899-44b8-bb0d-f3706f9389b9": "jw_t2_wifi_part_2",
		"5d298496-b108-4884-9713-1e51c843287b": "T2_HOTSPOT",
		"8c65c89d-dc11-4842-9000-b9fa6f45f34a": "t2_whix_test",
		"24f29c66-b658-40fd-a39d-dda8c670f3eb": "james_test_profile_001",
		"8205d716-8e45-4570-a34b-f1ebe0bdc75e": "Connie_test",
		"f8f4e7c7-924a-4a00-8ecf-04941dd6c4a3": "had_gw_wifi_radio",
		"cec11c05-ea4d-45cd-84e0-e44cecfafbf6": "jw_t2_docsis_part_2",
		"7eec6e18-0937-4a55-b16b-be4ea2219aa1": "jw_t2_wifi_part_6",
		"4397b229-200a-471f-9b46-41a19960ef18": "jw_t2_wifi_part_7",
		"3586d1d0-b3d3-4304-9a26-85d497d3ea3d": "xpc_test_profile_001",
		"824b41c8-210a-4eea-bc65-0f1bdb4c2574": "peter_test_profile_002",
		"07cd2a04-7083-44f2-a9d4-23823aed9c42": "jw_t2_wifi_part_5",
		"234b46be-0d9d-40f4-8b6c-d8e3a94f64d2": "wsmithT2.0TagProfileTest",
		"9fbf4f56-301b-4a28-8966-090cd38b498e": "wsmithT2.0ProfileTest",
		"4b84ffce-812a-4074-ba56-18982106f2f8": "SCOUT_DORY",
		"3a3ba25c-febd-40ac-8e38-b302aa428d69": "jw_t2_wifi_part_8",
		"e6ddbe95-daec-49db-8e72-123d53dbe630": "jw_t2_docsis_part_1",
	}
)
