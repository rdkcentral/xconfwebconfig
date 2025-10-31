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
	"fmt"
	"testing"

	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

// TableGenericNSList  = "GenericXconfNamedList"
// TableFirmwareConfig = "FirmwareConfig"
// TableFirmwareRule   = "FirmwareRule4"

func TestRuleEval(t *testing.T) {
	t.Skip()
	if server == nil {
		panic(noServerErr)
	}

	// setup data
	server.SetXconfData(shared.TableFirmwareConfig, FirmwareConfigId1, firmwareConfig1Bytes, 3600)
	server.SetXconfData(shared.TableFirmwareConfig, FirmwareConfigId2, firmwareConfig2Bytes, 3600)
	server.SetXconfData(shared.TableFirmwareConfig, FirmwareConfigId3, firmwareConfig3Bytes, 3600)

	server.SetXconfData(shared.TableFirmwareRule, firmwareRuleId1, firmwareRule1Bytes, 3600)
	server.SetXconfData(shared.TableFirmwareRule, firmwareRuleId2, firmwareRule2Bytes, 3600)
	server.SetXconfData(shared.TableFirmwareRule, firmwareRuleId3, firmwareRule3Bytes, 3600)

	macs := []string{mac3, "AA:AA:AA:BB:BB:BB", "AA:AA:AA:BB:BB:CC"}
	newList := shared.NewGenericNamespacedList(namespaceListKey, shared.MacList, macs)
	err := shared.CreateGenericNamedListOneDB(newList)
	assert.NilError(t, err)

	// load data
	ruleBytes := server.GetAllXconfDataAsList(shared.TableFirmwareRule, 0)
	processor := re.NewRuleProcessorFactory().RuleProcessor()

	// setup test parameters
	// ==== case 1 operation IS ====
	context := map[string]string{
		"eStbMac": mac1,
		// "env":             "VBN",
		// "model":           "PX001ANM",
		// "ipAddress":       "76.26.119.240",
		// "firmwareVersion": "ABC",
	}
	caseId := 1
	matchedRuleIds := []string{}

	for i, rbytes := range ruleBytes {
		var firmwareRule corefw.FirmwareRule

		err := json.Unmarshal(rbytes, &firmwareRule)
		if err != nil {
			panic(err)
		}

		ibytes, err := json.MarshalIndent(firmwareRule, "", "    ")
		if err != nil {
			panic(err)
		}

		// ERROR pointer
		if ok := processor.GetEvaluatorOK(&firmwareRule.Rule); !ok {
			// _ = ibytes
			fmt.Printf("i=%v no evluator", i)
			fmt.Printf("%v\n", string(ibytes))
			fmt.Printf("fwRule=%v\n\n", firmwareRule)
		}

		matched := processor.Evaluate(&firmwareRule.Rule, context, log.Fields{})
		if matched {
			fmt.Printf("#### MATCH caseId=%v, i=%v ####\n", caseId, i)
			fmt.Printf("%v\n", string(rbytes))
			fmt.Printf("%v\n", string(ibytes))
			fmt.Printf("%v\n", &firmwareRule)
			// assert.Assert(t, len(firmwareRule.ConfigId()) > 0)
			fmt.Printf("===========\n\n")
			matchedRuleIds = append(matchedRuleIds, firmwareRule.ID)
		}
	}
	assert.Equal(t, len(matchedRuleIds), 1)
	assert.Equal(t, matchedRuleIds[0], firmwareRuleId1)

	// ==== case 2 operation IN ====
	context = map[string]string{
		"eStbMac": mac2,
	}
	caseId = 2
	matchedRuleIds = []string{}

	for i, rbytes := range ruleBytes {
		var firmwareRule corefw.FirmwareRule

		err := json.Unmarshal(rbytes, &firmwareRule)
		if err != nil {
			panic(err)
		}

		ibytes, err := json.MarshalIndent(firmwareRule, "", "    ")
		if err != nil {
			panic(err)
		}

		matched := processor.Evaluate(&firmwareRule.Rule, context, log.Fields{})
		if matched {
			fmt.Printf("#### MATCH caseId=%v, i=%v ####\n", caseId, i)
			fmt.Printf("%v\n", string(ibytes))
			fmt.Printf("%v\n", &firmwareRule)
			fmt.Printf("===========\n\n")
			matchedRuleIds = append(matchedRuleIds, firmwareRule.ID)
		}
	}
	assert.Equal(t, len(matchedRuleIds), 1)
	assert.Equal(t, matchedRuleIds[0], firmwareRuleId2)

	// ==== case 3 operation IN_LIST ====
	context = map[string]string{
		"eStbMac": mac3,
	}
	caseId = 3
	matchedRuleIds = []string{}

	for i, rbytes := range ruleBytes {
		var firmwareRule corefw.FirmwareRule

		err := json.Unmarshal(rbytes, &firmwareRule)
		if err != nil {
			panic(err)
		}

		ibytes, err := json.MarshalIndent(firmwareRule, "", "    ")
		if err != nil {
			panic(err)
		}

		matched := processor.Evaluate(&firmwareRule.Rule, context, log.Fields{})
		if matched {
			fmt.Printf("#### MATCH caseId=%v, i=%v ####\n", caseId, i)
			fmt.Printf("%v\n", string(ibytes))
			fmt.Printf("%v\n", &firmwareRule)
			fmt.Printf("===========\n\n")
			matchedRuleIds = append(matchedRuleIds, firmwareRule.ID)
		}
	}
	assert.Equal(t, len(matchedRuleIds), 1)
	assert.Equal(t, matchedRuleIds[0], firmwareRuleId3)

	ok := true
	assert.Assert(t, ok)
}
