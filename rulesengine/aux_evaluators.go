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
	"regexp"

	"github.com/rdkcentral/xconfwebconfig/util"
)

func GetAuxEvaluators() []IConditionEvaluator {
	evaluators := GetComparingEvaluators(AuxFreeArgTypeTime)

	ev2 := NewBaseEvaluator(
		AuxFreeArgTypeIpAddress,
		StandardOperationIs,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			return freeArgValue == fixedArgValue
		},
	)
	evaluators = append(evaluators, ev2)

	// TODO for now, i will use the []string style to eval
	//      if it causes errors, I can come back and fix the problem
	//      ASSUME the fixedArgValue is a []string and all normalized
	ev3 := NewBaseEvaluator(
		AuxFreeArgTypeIpAddress,
		StandardOperationIn,
		func(freeArgValue string, fixedArgValueItf interface{}) bool {
			if values, ok := fixedArgValueItf.([]string); ok {
				return util.Contains(values, freeArgValue)
			}
			return false
		},
	)
	evaluators = append(evaluators, ev3)

	ev4 := NewBaseEvaluator(
		AuxFreeArgTypeIpAddress,
		StandardOperationPercent,
		func(freeArgValue string, fixedArgValueItf interface{}) bool {
			if percent, ok := fixedArgValueItf.(float64); ok {
				return FitsPercent(freeArgValue, percent)
			}
			return false
		},
	)
	evaluators = append(evaluators, ev4)

	ev5 := NewBaseEvaluator(
		AuxFreeArgTypeMacAddress,
		StandardOperationIs,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			return freeArgValue == fixedArgValue
		},
	)
	evaluators = append(evaluators, ev5)

	ev6 := NewBaseEvaluator(
		AuxFreeArgTypeMacAddress,
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
	evaluators = append(evaluators, ev6)

	ev7 := NewBaseEvaluator(
		AuxFreeArgTypeMacAddress,
		StandardOperationIn,
		func(freeArgValue string, fixedArgValueItf interface{}) bool {
			if values, ok := fixedArgValueItf.([]string); ok {
				return util.Contains(values, freeArgValue)
			}
			return false
		},
	)
	evaluators = append(evaluators, ev7)

	// ev8

	// ev9
	ev9 := NewBaseEvaluator(
		AuxFreeArgTypeMacAddress,
		StandardOperationIn,
		func(freeArgValue string, fixedArgValueItf interface{}) bool {
			if values, ok := fixedArgValueItf.([]string); ok {
				return util.Contains(values, freeArgValue)
			}
			return false
		},
	)
	evaluators = append(evaluators, ev9)

	return evaluators
}
