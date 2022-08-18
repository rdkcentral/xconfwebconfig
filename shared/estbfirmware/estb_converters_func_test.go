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
package estbfirmware

import (
	"encoding/json"
	"testing"

	"xconfwebconfig/shared/firmware"

	"gotest.tools/assert"
)

func TestConvertFirmwareRuleToIpRuleBean(t *testing.T) {
	// create rule like the following
	ruleStr := []byte(`{
		"id": "2123455666",
		"name": "000ipPerformanceTestRule",
		"rule": {
		  "negated": false,
		  "compoundParts": [
			{
			  "negated": false,
			  "condition": {
				"freeArg": {
				  "type": "STRING",
				  "name": "ipAddress"
				},
				"operation": "IN_LIST",
				"fixedArg": {
				  "bean": {
					"value": {
					  "java.lang.String": "%v"
					}
				  }
				}
			  }
			}
		  ],
		  "condition": {
				"freeArg": {
				  "type": "STRING",
				  "name": "ipAddress"
				},
				"operation": "IN_LIST",
				"fixedArg": {
				  "bean": {
					"value": {
					  "java.lang.String": "127.0.0.1"
					}
				  }
				}
			}
		},
		"applicableAction": {
		  "type": ".RuleAction",
		  "actionType": "RULE",
		  "configId": "234567",
		  "configEntries": [],
		  "active": true,
		  "useAccountPercentage": false,
		  "firmwareCheckRequired": false,
		  "rebootImmediately": false,
		  "properties": {
			  "firmwareLocation": "http://127.0.1.1/app/download",
			  "ipv6FirmwareLocation": "http://127.0.1.1/app/downloadv6",
			  "irmwareDownloadProtocol": "https"
		  }
		},
		"type": "IP_RULE",
		"active": true,
		"applicationType": "stb"
	  }`)
	var rule firmware.FirmwareRule
	err := json.Unmarshal(ruleStr, &rule)
	assert.NilError(t, err)
	bean := ConvertFirmwareRuleToIpRuleBean(&rule)
	assert.Assert(t, bean != nil)
	assert.Assert(t, bean.IpAddressGroup != nil)
}
