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
	"fmt"
	"regexp"
	"strconv"

	"github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

func GetStandardEvaluators() []IConditionEvaluator {
	// evaluators := []IConditionEvaluator{}

	// implement/load TimeGTEEvaluator/TimeLTEEvaluator
	//    (extends xconf/evaluatorsBaseTimeEvaluator)
	evaluators := GetComparingEvaluators(StandardFreeArgTypeString)

	ev1 := NewBaseEvaluator(
		StandardFreeArgTypeAny,
		StandardOperationExists,
		func(freeArgValue string, fixedArgValueItf interface{}) bool {
			if fixedArgValueItf != nil {
				return false
			}
			return true
		},
	)
	evaluators = append(evaluators, ev1)

	ev2 := NewBaseEvaluator(
		StandardFreeArgTypeString,
		StandardOperationIs,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			return freeArgValue == fixedArgValue
			//if v, ok := fixedArgValue.(string); ok {
			//return freeArgValue == v
			//}
			//return false
		},
	)
	evaluators = append(evaluators, ev2)

	ev3 := NewBaseEvaluator(
		StandardFreeArgTypeString,
		StandardOperationLike,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			if pattern, ok := fixedArgValue.(string); ok {
				if matched, err := regexp.Match(pattern, []byte(freeArgValue)); err == nil {
					return matched
				}
			}
			return false
		},
	)
	evaluators = append(evaluators, ev3)

	ev4 := NewBaseEvaluator(
		StandardFreeArgTypeString,
		StandardOperationIn,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			if faValues, ok := fixedArgValue.([]string); ok {
				return util.Contains(faValues, freeArgValue)
			}
			return false
		},
	)
	evaluators = append(evaluators, ev4)

	ev5 := NewBaseEvaluator(
		StandardFreeArgTypeString,
		StandardOperationAnyMatched,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			if patterns, ok := fixedArgValue.([]string); ok {
				for _, pattern := range patterns {
					if matched, err := regexp.Match(pattern, []byte(freeArgValue)); err == nil {
						if matched {
							return true
						}
					}
				}
			}
			return false
		},
	)
	evaluators = append(evaluators, ev5)

	// TODO Implement the FitsPercent()
	ev6 := NewBaseEvaluator(
		StandardFreeArgTypeString,
		StandardOperationPercent,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			// sometimes the value comes back in quotes as a string, and sometimes without quotes as a float
			if percentString, ok := fixedArgValue.(string); ok {
				percent, err := strconv.ParseFloat(percentString, 64)
				if err == nil {
					return FitsPercent(freeArgValue, percent)
				}
			}
			if percent, ok := fixedArgValue.(float64); ok {
				return FitsPercent(freeArgValue, percent)
			}
			log.Warn(fmt.Sprintf("Percent value is not a float64: %+v", fixedArgValue))
			return false
		},
	)
	evaluators = append(evaluators, ev6)

	ev7 := GetComparingEvaluators(StandardFreeArgTypeLong)
	evaluators = append(evaluators, ev7...)

	ev8 := NewBaseEvaluator(
		StandardFreeArgTypeLong,
		StandardOperationIn,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			ivalue, err := strconv.Atoi(freeArgValue)
			if err != nil {
				return false
			}

			if collection, ok := fixedArgValue.([]int); ok {
				return util.Contains(collection, ivalue)
			}
			return false
		},
	)
	evaluators = append(evaluators, ev8)

	// TODO ev9 implement the FitsPercent()

	ev10 := NewBaseEvaluator(
		StandardFreeArgTypeVoid,
		StandardOperationIs,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			if bvalue, ok := fixedArgValue.(bool); ok {
				return bvalue
			}
			return false
		},
	)
	evaluators = append(evaluators, ev10)

	ev11 := NewBaseEvaluator(
		StandardFreeArgTypeVoid,
		StandardOperationPercent,
		func(freeArgValue string, fixedArgValueItf interface{}) bool {
			random := util.RandomDouble()

			if fixedArgValue, ok := fixedArgValueItf.(float64); ok {
				return random*100 < fixedArgValue
			}
			return false
		},
	)
	evaluators = append(evaluators, ev11)

	return evaluators
}
