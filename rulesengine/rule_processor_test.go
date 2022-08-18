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
	"encoding/json"
	"testing"

	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func TestBasicRule01(t *testing.T) {
	context := map[string]string{
		"day":      "Friday",
		"age":      "2",
		"vacation": "",
		"time":     "13:30",
		"ip":       "192.168.0.1",
	}
	processor := NewRuleProcessor()

	// fixed arg type string
	day := NewFreeArg(StandardFreeArgTypeString, "day")

	fridayRule := Rule{}
	fridayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Friday")))
	fridayRule.SetId("fridayRule")

	saturdayRule := Rule{}
	saturdayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))
	saturdayRule.SetId("saturdayRule")

	sundayRule := Rule{}
	sundayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Sunday")))
	sundayRule.SetId("sundayRule")

	b1 := processor.Evaluate(&fridayRule, context, log.Fields{})
	assert.Assert(t, b1)

	b2 := processor.Evaluate(&saturdayRule, context, log.Fields{})
	assert.Assert(t, !b2)

	// ----------------------
	twoDayRule := Rule{}
	twoDayRule.SetCompoundParts(
		[]Rule{
			fridayRule,
			Or(saturdayRule),
		},
	)
	twoDayRule.SetId("twoDayRule")

	context3 := map[string]string{
		"day": "Friday",
	}
	b3 := processor.Evaluate(&twoDayRule, context3, log.Fields{})
	assert.Assert(t, b3)

	context4 := map[string]string{
		"day": "Saturday",
	}
	b4 := processor.Evaluate(&twoDayRule, context4, log.Fields{})
	assert.Assert(t, b4)

	context5 := map[string]string{
		"day": "Sunday",
	}
	b5 := processor.Evaluate(&twoDayRule, context5, log.Fields{})
	assert.Assert(t, !b5)
	// ----------------------

	weekendRule := Rule{}
	weekendRule.SetCompoundParts(
		[]Rule{
			saturdayRule,
			Or(sundayRule),
		},
	)
	weekendRule.SetId("weekendRule")
	assert.Assert(t, !processor.Evaluate(&weekendRule, context3, log.Fields{}))
	assert.Assert(t, processor.Evaluate(&weekendRule, context4, log.Fields{}))
	assert.Assert(t, processor.Evaluate(&weekendRule, context5, log.Fields{}))

	weekdayRule := Not(weekendRule)
	weekdayRule.SetId("weekdayRule")
	assert.Assert(t, processor.Evaluate(&weekdayRule, context3, log.Fields{}))
	assert.Assert(t, !processor.Evaluate(&weekdayRule, context4, log.Fields{}))
	assert.Assert(t, !processor.Evaluate(&weekdayRule, context5, log.Fields{}))
}

func TestBasicRule02(t *testing.T) {
	processor := NewRuleProcessor()

	context := map[string]string{
		"day":      "Friday",
		"age":      "2",
		"vacation": "",
		"time":     "13:30",
		"ip":       "192.168.0.1",
	}

	age := NewFreeArg(StandardFreeArgTypeLong, "age")

	// fixed arg type float64
	babyRule := Rule{}
	babyRule.SetCondition(NewCondition(age, StandardOperationLt, NewFixedArg(5.0)))
	babyRule.SetId("babyRule")
	b5 := processor.Evaluate(&babyRule, context, log.Fields{})
	assert.Assert(t, b5)

	kidLowerRule := Rule{}
	kidLowerRule.SetCondition(NewCondition(age, StandardOperationGt, NewFixedArg(5.0)))
	kidLowerRule.SetId("kidLowerRule")

	kidUpperRule := Rule{}
	kidUpperRule.SetCondition(NewCondition(age, StandardOperationLt, NewFixedArg(18.0)))
	kidUpperRule.SetId("kidUpperRule")

	kidRule := Rule{}
	kidRule.SetCompoundParts(
		[]Rule{
			kidLowerRule,
			And(kidUpperRule),
		},
	)
	assert.Assert(t, !processor.Evaluate(&kidRule, context, log.Fields{}))

	context1 := map[string]string{
		"age": "8",
	}
	assert.Assert(t, processor.Evaluate(&kidRule, context1, log.Fields{}))

	context2 := map[string]string{
		"age": "19",
	}
	assert.Assert(t, !processor.Evaluate(&kidRule, context2, log.Fields{}))
}

func TestBasicRule03(t *testing.T) {
	processor := NewRuleProcessor()

	context := map[string]string{
		"day":      "Friday",
		"age":      "2",
		"vacation": "",
		"time":     "13:30",
		"ip":       "192.168.0.1",
	}

	vacation := NewFreeArg(StandardFreeArgTypeAny, "vacation")

	// fixed arg type float64
	vacationRule := Rule{}
	vacationRule.SetCondition(NewCondition(vacation, StandardOperationExists, nil))
	vacationRule.SetId("vacation")
	assert.Assert(t, processor.Evaluate(&vacationRule, context, log.Fields{}))
}

func TestBasicRule04(t *testing.T) {
	processor := NewRuleProcessor()

	context := map[string]string{
		"day":      "Friday",
		"age":      "2",
		"vacation": "",
		"time":     "13:30",
		"ip":       "192.168.0.1",
	}

	noArg := NewFreeArg(StandardFreeArgTypeVoid, "")

	// fixed arg type float64
	alwaysTrueRule := Rule{}
	alwaysTrueRule.SetCondition(NewCondition(noArg, StandardOperationPercent, NewFixedArg(100.0)))
	alwaysTrueRule.SetId("alwaysTrueRule")
	assert.Assert(t, processor.Evaluate(&alwaysTrueRule, context, log.Fields{}))
}

func TestBasicRule05(t *testing.T) {
	processor := NewRuleProcessor()

	context := map[string]string{
		"day":      "Friday",
		"age":      "2",
		"vacation": "",
		"time":     "12:30:00",
		"ip":       "192.168.0.1",
	}

	freeArgTime := NewFreeArg(AuxFreeArgTypeTime, "time")

	// fixed arg type float64
	midnightRule := Rule{}
	midnightRule.SetCondition(NewCondition(freeArgTime, StandardOperationIs, NewFixedArg("00:00:00")))
	midnightRule.SetId("midnightRule")
	b0 := processor.Evaluate(&midnightRule, context, log.Fields{})
	assert.Assert(t, !b0)

	after12Rule := Rule{}
	after12Rule.SetCondition(NewCondition(freeArgTime, StandardOperationGt, NewFixedArg("12:00:00")))
	after12Rule.SetId("after12Rule")
	b1 := processor.Evaluate(&after12Rule, context, log.Fields{})
	assert.Assert(t, b1)

	before13Rule := Rule{}
	before13Rule.SetCondition(NewCondition(freeArgTime, StandardOperationLt, NewFixedArg("13:00:00")))
	before13Rule.SetId("before13Rule")
	b2 := processor.Evaluate(&before13Rule, context, log.Fields{})
	assert.Assert(t, b2)

	lunchTimeRule := Rule{}
	lunchTimeRule.SetCompoundParts(
		[]Rule{
			after12Rule,
			And(before13Rule),
		},
	)
	lunchTimeRule.SetId("lunchTimeRule")
	b3 := processor.Evaluate(&lunchTimeRule, context, log.Fields{})
	assert.Assert(t, b3)
}

func TestBasicRule06(t *testing.T) {
	processor := NewRuleProcessor()

	context := map[string]string{
		"day":      "Friday",
		"age":      "2",
		"vacation": "",
		"time":     "12:30:00",
		"ip":       "192.168.0.1",
	}

	freeArgIp := NewFreeArg(AuxFreeArgTypeIpAddress, "ip")

	// fixed arg type float64
	localhostRule := Rule{}
	localhostRule.SetCondition(NewCondition(freeArgIp, StandardOperationIs, NewFixedArg("127.0.0.1")))
	localhostRule.SetId("localhostRule")
	b0 := processor.Evaluate(&localhostRule, context, log.Fields{})
	assert.Assert(t, !b0)
}

func TestBasicRule07(t *testing.T) {
	processor := NewRuleProcessor()

	context := map[string]string{
		"day":      "Friday",
		"age":      "2",
		"vacation": "",
		"time":     "12:30:00",
		"ip":       "192.168.0.1",
	}

	day := NewFreeArg(StandardFreeArgTypeString, "day")
	age := NewFreeArg(StandardFreeArgTypeLong, "age")
	vacation := NewFreeArg(StandardFreeArgTypeAny, "vacation")
	freeArgTime := NewFreeArg(AuxFreeArgTypeTime, "time")

	vacationRule := Rule{}
	vacationRule.SetCondition(NewCondition(vacation, StandardOperationExists, nil))
	vacationRule.SetId("vacation")

	saturdayRule := Rule{}
	saturdayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))
	saturdayRule.SetId("saturdayRule")

	sundayRule := Rule{}
	sundayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Sunday")))
	sundayRule.SetId("sundayRule")

	weekendRule := Rule{}
	weekendRule.SetCompoundParts(
		[]Rule{
			saturdayRule,
			Or(sundayRule),
		},
	)
	weekendRule.SetId("weekendRule")

	after12Rule := Rule{}
	after12Rule.SetCondition(NewCondition(freeArgTime, StandardOperationGt, NewFixedArg("12:00:00")))
	after12Rule.SetId("after12Rule")

	before13Rule := Rule{}
	before13Rule.SetCondition(NewCondition(freeArgTime, StandardOperationLt, NewFixedArg("13:00:00")))
	before13Rule.SetId("before13Rule")

	lunchTimeRule := Rule{}
	lunchTimeRule.SetCompoundParts(
		[]Rule{
			after12Rule,
			And(before13Rule),
		},
	)
	lunchTimeRule.SetId("lunchTimeRule")

	babyRule := Rule{}
	babyRule.SetCondition(NewCondition(age, StandardOperationLt, NewFixedArg(5.0)))
	babyRule.SetId("babyRule")

	retireeRule := Rule{}
	retireeRule.SetCondition(NewCondition(age, StandardOperationGt, NewFixedArg(65.0)))
	retireeRule.SetId("retireeRule")

	notAtWorkRule := Rule{}
	notAtWorkRule.SetCompoundParts(
		[]Rule{
			vacationRule,
			Or(weekendRule),
			Or(lunchTimeRule),
			Or(babyRule),
			Or(retireeRule),
		},
	)

	assert.Assert(
		t,
		processor.Evaluate(&notAtWorkRule, context, log.Fields{}),
	)

	context1 := map[string]string{
		"day":  "Monday",
		"age":  "24",
		"time": "09:00",
		"ip":   "127.0.0.1",
	}
	assert.Assert(
		t,
		!processor.Evaluate(&notAtWorkRule, context1, log.Fields{}),
	)
}

func getIdSet(rules []Rule) util.Set {
	m := util.Set{}
	for _, r := range rules {
		m.Add(r.Id())
	}
	return m
}

func TestRulesEval(t *testing.T) {
	processor := NewRuleProcessor()

	context := map[string]string{
		"day":      "Friday",
		"age":      "2",
		"vacation": "",
		"time":     "12:30:00",
		"ip":       "192.168.0.1",
	}

	day := NewFreeArg(StandardFreeArgTypeString, "day")
	age := NewFreeArg(StandardFreeArgTypeLong, "age")
	vacation := NewFreeArg(StandardFreeArgTypeAny, "vacation")
	freeArgTime := NewFreeArg(AuxFreeArgTypeTime, "time")
	noArg := NewFreeArg(StandardFreeArgTypeVoid, "")
	freeArgIp := NewFreeArg(AuxFreeArgTypeIpAddress, "ip")

	vacationRule := Rule{}
	vacationRule.SetCondition(NewCondition(vacation, StandardOperationExists, nil))
	vacationRule.SetId("vacationRule")

	saturdayRule := Rule{}
	saturdayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))
	saturdayRule.SetId("saturdayRule")

	sundayRule := Rule{}
	sundayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Sunday")))
	sundayRule.SetId("sundayRule")

	weekendRule := Rule{}
	weekendRule.SetCompoundParts(
		[]Rule{
			saturdayRule,
			Or(sundayRule),
		},
	)
	weekendRule.SetId("weekendRule")

	weekdayRule := Not(weekendRule)
	weekdayRule.SetId("weekdayRule")

	after12Rule := Rule{}
	after12Rule.SetCondition(NewCondition(freeArgTime, StandardOperationGt, NewFixedArg("12:00:00")))
	after12Rule.SetId("after12Rule")

	before13Rule := Rule{}
	before13Rule.SetCondition(NewCondition(freeArgTime, StandardOperationLt, NewFixedArg("13:00:00")))
	before13Rule.SetId("before13Rule")

	lunchTimeRule := Rule{}
	lunchTimeRule.SetCompoundParts(
		[]Rule{
			after12Rule,
			And(before13Rule),
		},
	)
	lunchTimeRule.SetId("lunchTimeRule")

	babyRule := Rule{}
	babyRule.SetCondition(NewCondition(age, StandardOperationLt, NewFixedArg(5.0)))
	babyRule.SetId("babyRule")

	retireeRule := Rule{}
	retireeRule.SetCondition(NewCondition(age, StandardOperationGt, NewFixedArg(65.0)))
	retireeRule.SetId("retireeRule")

	alwaysTrueRule := Rule{}
	alwaysTrueRule.SetCondition(NewCondition(noArg, StandardOperationPercent, NewFixedArg(100.0)))
	alwaysTrueRule.SetId("alwaysTrueRule")

	midnightRule := Rule{}
	midnightRule.SetCondition(NewCondition(freeArgTime, StandardOperationIs, NewFixedArg("00:00:00")))
	midnightRule.SetId("midnightRule")

	localhostRule := Rule{}
	localhostRule.SetCondition(NewCondition(freeArgIp, StandardOperationIs, NewFixedArg("127.0.0.1")))
	localhostRule.SetId("localhostRule")

	notAtWorkRule := Rule{}
	notAtWorkRule.SetCompoundParts(
		[]Rule{
			vacationRule,
			Or(weekendRule),
			Or(lunchTimeRule),
			Or(babyRule),
			Or(retireeRule),
		},
	)
	notAtWorkRule.SetId("notAtWorkRule")

	assert.Assert(
		t,
		processor.Evaluate(&notAtWorkRule, context, log.Fields{}),
	)

	context1 := map[string]string{
		"day":  "Monday",
		"age":  "24",
		"time": "09:00",
		"ip":   "127.0.0.1",
	}
	assert.Assert(
		t,
		!processor.Evaluate(&notAtWorkRule, context1, log.Fields{}),
	)

	// test case 1
	rulesToTest := []Rule{
		weekendRule,
		weekdayRule,
		babyRule,
		retireeRule,
		alwaysTrueRule,
		vacationRule,
		midnightRule,
		lunchTimeRule,
		localhostRule,
		notAtWorkRule,
	}

	matched := processor.Filter(rulesToTest, context)
	assert.Assert(t, len(matched) == 6)
	expectedMatchedIdSet := util.Set{}
	expectedMatchedIdSet.Add(
		"weekdayRule",
		"babyRule",
		"vacationRule",
		"lunchTimeRule",
		"alwaysTrueRule",
		"notAtWorkRule",
	)
	assert.DeepEqual(t, getIdSet(matched), expectedMatchedIdSet)

	// test case 2
	context2 := map[string]string{
		"day":      "Saturday",
		"age":      "24",
		"vacation": "",
		"time":     "00:00:00",
		"ip":       "127.0.0.1",
	}
	matched = processor.Filter(rulesToTest, context2)
	assert.Assert(t, len(matched) == 6)
	expectedMatchedIdSet = util.Set{}
	expectedMatchedIdSet.Add(
		"weekendRule",
		"vacationRule",
		"midnightRule",
		"localhostRule",
		"alwaysTrueRule",
		"notAtWorkRule",
	)
	assert.DeepEqual(t, getIdSet(matched), expectedMatchedIdSet)

	// test case 3
	context3 := map[string]string{
		"day":  "Monday",
		"age":  "24",
		"time": "16:00:00",
		"ip":   "135.5.7.98",
	}
	matched = processor.Filter(rulesToTest, context3)
	assert.Assert(t, len(matched) == 2)
	expectedMatchedIdSet = util.Set{}
	expectedMatchedIdSet.Add(
		"weekdayRule",
		"alwaysTrueRule",
	)
	assert.DeepEqual(t, getIdSet(matched), expectedMatchedIdSet)
}

func createLikeRule(freeArgKey string, fixedArgValue string) Rule {
	rule := Rule{}
	freeArg := NewFreeArg(StandardFreeArgTypeString, freeArgKey)
	rule.SetCondition(NewCondition(freeArg, StandardOperationLike, NewFixedArg(fixedArgValue)))
	return rule
}

func TestRelations(t *testing.T) {
	processor := NewRuleProcessor()

	complexRule := Rule{}
	compoundParts := []Rule{
		createLikeRule("adsqa27", "test1.adsqa27"),
		Or(Not(createLikeRule("adsqa28", "[asd|qwe].*"))),
		And(Not(createLikeRule("adsqa29", ".*ads.*qa.*29"))),
		Or(createLikeRule("adsqa30", "(12|c)*")),
		And(createLikeRule("adsqa31", "[hc]at")),
	}
	complexRule.SetCompoundParts(compoundParts)

	context := map[string]string{
		"adsqa21": "10000",
		"adsqa22": "333334",
		"adsqa23": "666667",
		"adsqa24": "44444445",
		"adsqa25": "23",
	}
	assert.Assert(t, !processor.Evaluate(&complexRule, context, log.Fields{}))

	context = map[string]string{
		"adsqa21": "10000",
		"adsqa22": "333334",
		"adsqa23": "666667",
		"adsqa27": "test1.adsqa27",
		"adsqa31": "hat",
	}

	oneRule := createLikeRule("adsqa27", "test1.adsqa27")
	assert.Assert(t, processor.Evaluate(&oneRule, context, log.Fields{}))

	twoRule := Rule{}
	twoRule.SetCompoundParts(
		[]Rule{
			createLikeRule("adsqa27", "test1.adsqa27"),
		},
	)

	assert.Assert(t, processor.Evaluate(&complexRule, context, log.Fields{}))
}

func TestRuleJsonEncoding(t *testing.T) {
	day := NewFreeArg(StandardFreeArgTypeString, "day")

	saturdayRule := Rule{}
	saturdayRule.SetCondition(NewCondition(day, StandardOperationIs, NewFixedArg("Saturday")))

	bbytes, err := json.Marshal(&saturdayRule)
	assert.NilError(t, err)

	var parsedRule Rule
	err = json.Unmarshal(bbytes, &parsedRule)
	assert.NilError(t, err)

	assert.Assert(t, saturdayRule.Equals(&parsedRule))
}

func TestProcessorMethods(t *testing.T) {
	processor := NewRuleProcessor()
	oldSize := processor.Size()

	evaluators := []IConditionEvaluator{
		NewBaseEvaluator(
			"foo",
			"bar",
			func(freeArgValue string, fixedArgValueItf interface{}) bool {
				if fixedArgValueItf != nil {
					return false
				}
				return true
			},
		),
		NewBaseEvaluator(
			"hello",
			"world",
			func(freeArgValue string, fixedArgValueItf interface{}) bool {
				if fixedArgValueItf != nil {
					return false
				}
				return true
			},
		),
	}

	processor.AddEvaluators(evaluators)
	assert.Equal(t, oldSize+2, processor.Size())
}
