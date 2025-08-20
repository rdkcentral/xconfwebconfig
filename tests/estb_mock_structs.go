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
	"encoding/json"

	"github.com/rdkcentral/xconfwebconfig/shared"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"
)

func GetFirmwareRuleTemplate(i int) *corefw.FirmwareRuleTemplate {
	var templateStr string
	switch i {
	case 1:
		templateStr = GetFirmwareTemplateStr1()
	case 2:
		templateStr = GetFirmwareTemplateStr2()
	case 3:
		templateStr = GetFirmwareTemplateStr3()
	default:
		templateStr = GetFirmwareTemplateStr4()
	}

	inst := corefw.FirmwareRuleTemplate{}
	err := json.Unmarshal([]byte(templateStr), &inst)
	if err != nil {
		panic(err)
	}
	return &inst
}

func GetSetDaoGenericNamespacedList() (*shared.GenericNamespacedList, error) {
	newList := shared.NewGenericNamespacedList(NamespaceIPListKey, shared.IpList, NamespaceIPList)
	err := shared.CreateGenericNamedListOneDB(newList)
	return newList, err
}

func GetGetRDKCDownloadLocationROUNDROBINFILTERVALUE() *coreef.DownloadLocationRoundRobinFilterValue {
	downloadStr := GetRDKCDownloadLocationROUNDROBINFILTERVALUE()
	inst := coreef.DownloadLocationRoundRobinFilterValue{}

	err := json.Unmarshal([]byte(downloadStr), &inst)
	if err != nil {
		panic(err)
	}
	return &inst
}

func GetFirmwareConfig1() *coreef.FirmwareConfig {
	firmwareConfigStr := GetFirmwareConfigStr1()
	firmwareConfig := coreef.FirmwareConfig{}
	err := json.Unmarshal([]byte(firmwareConfigStr), &firmwareConfig)
	if err != nil {
		panic(err)
	}
	return &firmwareConfig
}

func GetFirmwareConfig2() *coreef.FirmwareConfig {
	firmwareConfigStr := GetFirmwareConfigStr2()
	firmwareConfig := coreef.FirmwareConfig{}
	err := json.Unmarshal([]byte(firmwareConfigStr), &firmwareConfig)
	if err != nil {
		panic(err)
	}
	return &firmwareConfig
}

func GetFirmwareRule1() *corefw.FirmwareRule {
	firmwareRuleStr := GetFirmwareRuleStr1()
	firmwareRule := corefw.FirmwareRule{}

	err := json.Unmarshal([]byte(firmwareRuleStr), &firmwareRule)
	if err != nil {
		panic(err)
	}
	return &firmwareRule
}

func GetFirmwareRule2() *corefw.FirmwareRule {
	firmwareRuleStr := GetFirmwareRuleStr2()
	firmwareRule := corefw.FirmwareRule{}

	err := json.Unmarshal([]byte(firmwareRuleStr), &firmwareRule)
	if err != nil {
		panic(err)
	}
	return &firmwareRule
}

func GetFirmwareRule3() *corefw.FirmwareRule {
	firmwareRuleStr := GetFirmwareRuleStr3()
	firmwareRule := corefw.FirmwareRule{}

	err := json.Unmarshal([]byte(firmwareRuleStr), &firmwareRule)
	if err != nil {
		panic(err)
	}
	return &firmwareRule
}

func GetFirmwareRule4() *corefw.FirmwareRule {
	firmwareRuleStr := GetFirmwareRuleStr4()
	firmwareRule := corefw.FirmwareRule{}

	err := json.Unmarshal([]byte(firmwareRuleStr), &firmwareRule)
	if err != nil {
		panic(err)
	}
	return &firmwareRule
}
