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
package rulesengine

import (
	"testing"

	"gotest.tools/assert"
)

func GetWeekendRule() Rule {
	day := NewFreeArg(StandardFreeArgTypeString, "day")

	saturdayRule := Rule{}
	saturdayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))

	sundayRule := Rule{}
	sundayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Sunday")))

	weekendRule := Rule{}
	weekendRule.SetCompoundParts(
		[]Rule{
			saturdayRule,
			Or(sundayRule),
		},
	)
	return weekendRule
}

func TestRuleEquals(t *testing.T) {
	day := NewFreeArg(StandardFreeArgTypeString, "day")

	saturdayRule1 := Rule{}
	saturdayRule1.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))
	saturdayRule2 := Rule{}
	saturdayRule2.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))
	assert.Assert(t, saturdayRule1.Equals(&saturdayRule2))

	weekendRule1 := GetWeekendRule()
	weekendRule2 := GetWeekendRule()
	assert.Assert(t, weekendRule1.Equals(&weekendRule2))
}
