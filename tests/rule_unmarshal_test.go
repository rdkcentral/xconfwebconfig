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
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"gotest.tools/assert"
)

func TestRuleUnmarshal(t *testing.T) {
	// ==== formula ====
	var dcmFormula logupload.DCMFormula

	err := json.Unmarshal(formulaData01, &dcmFormula)
	assert.NilError(t, err)

	fmt.Printf("%v\n\n\n", dcmFormula)
	fmt.Printf("dcmFormula.FormulaRule=%v\n\n\n", dcmFormula.Formula)
	fmt.Printf("dcmFormula.FormulaRule.Rule=%v\n\n\n", dcmFormula.Formula.Rule)
	assert.Equal(t, dcmFormula.Formula.Rule.GetCondition().GetOperation(), "IN_LIST")

	// ==== formula ====
	var dcmRule logupload.DCMGenericRule

	err = json.Unmarshal(ruleData01, &dcmRule)
	assert.NilError(t, err)

	fmt.Printf("%v\n\n\n", dcmRule)
	fmt.Printf("%v\n\n\n", dcmRule.Rule)
	assert.Equal(t, dcmRule.Rule.GetCondition().GetOperation(), "IN_LIST")

	// ==== simple rule ====
	ruleData03 := []byte(`{
		"negated": true,
		"condition": {
		  "freeArg": {
			"type": "STRING",
			"name": "estbMacAddress"
		  },
		  "operation": "IN_LIST",
		  "fixedArg": {
			"bean": {
			  "value": {
				"java.lang.String": "CORE_NW_MAC_LIST"
			  }
			}
		  }
		}
	}`)
	var rule re.Rule
	err = json.Unmarshal(ruleData03, &rule)
	assert.NilError(t, err)
	fmt.Printf("rule=%v\n", rule)
	assert.Equal(t, rule.GetCondition().GetOperation(), "IN_LIST")
}
