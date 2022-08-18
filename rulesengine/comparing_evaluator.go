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
	"strconv"
	"time"
)

const (
	timeFormat = "15:04:05"
)

type FnEvaluation func(int) bool

var (
	comparableOperationFnMap = map[string]func(int) bool{
		StandardOperationIs:  func(i int) bool { return i == 0 },
		StandardOperationGt:  func(i int) bool { return i > 0 },
		StandardOperationGte: func(i int) bool { return i >= 0 },
		StandardOperationLt:  func(i int) bool { return i < 0 },
		StandardOperationLte: func(i int) bool { return i <= 0 },
	}
)

type ComparingEvaluator struct {
	freeArgType string
	operation   string
	evaluation  FnEvaluation
}

func GetComparingEvaluators(freeArgType string) []IConditionEvaluator {
	// It is assumed (STRING, GTE/LTE) evaluators will ONLY be used to handle "time")
	var baseComparingEvaluators []IConditionEvaluator

	if freeArgType == StandardFreeArgTypeString {
		// (STRING, IS) evaluator is implemented in the GetStandardEvaluators()
		baseComparingEvaluators = []IConditionEvaluator{
			NewComparingEvaluator(freeArgType, StandardOperationGt, func(i int) bool { return i > 0 }),
			NewComparingEvaluator(freeArgType, StandardOperationGte, func(i int) bool { return i >= 0 }),
			NewComparingEvaluator(freeArgType, StandardOperationLt, func(i int) bool { return i < 0 }),
			NewComparingEvaluator(freeArgType, StandardOperationLte, func(i int) bool { return i <= 0 }),
		}
	} else {
		baseComparingEvaluators = []IConditionEvaluator{
			NewComparingEvaluator(freeArgType, StandardOperationIs, func(i int) bool { return i == 0 }),
			NewComparingEvaluator(freeArgType, StandardOperationGt, func(i int) bool { return i > 0 }),
			NewComparingEvaluator(freeArgType, StandardOperationGte, func(i int) bool { return i >= 0 }),
			NewComparingEvaluator(freeArgType, StandardOperationLt, func(i int) bool { return i < 0 }),
			NewComparingEvaluator(freeArgType, StandardOperationLte, func(i int) bool { return i <= 0 }),
		}
	}
	return baseComparingEvaluators
}

func NewComparingEvaluator(freeArgType string, operation string, fn FnEvaluation) *ComparingEvaluator {
	return &ComparingEvaluator{
		freeArgType: freeArgType,
		operation:   operation,
		evaluation:  fn,
	}
}

func (e *ComparingEvaluator) FreeArgType() string {
	return e.freeArgType
}

func (e *ComparingEvaluator) Operation() string {
	return e.operation
}

func (e *ComparingEvaluator) Evaluate(condition *Condition, context map[string]string) bool {
	var freeArgValue string
	var ok bool
	if e.freeArgType != StandardFreeArgTypeVoid {
		freeArgValue, ok = context[condition.GetFreeArg().GetName()]
		if !ok || (e.freeArgType != StandardFreeArgTypeAny && len(freeArgValue) == 0) {
			return false
		}
	}

	var compareResult int

	// return e.evaluateInternal(freeArgValue, condition.GetFixedArg().GetValue())

	conditionArgType := condition.GetFreeArg().GetType()
	if conditionArgType == StandardFreeArgTypeLong {
		if freeArgNum, err := strconv.Atoi(freeArgValue); err == nil {
			if fixedValDouble, ok := condition.GetFixedArg().GetValue().(float64); ok {
				fixedArgNum := int(fixedValDouble)
				if freeArgNum > fixedArgNum {
					compareResult = 1
				} else if freeArgNum < fixedArgNum {
					compareResult = -1
				}
				return e.evaluation(compareResult)
			}
		}
	} else if conditionArgType == AuxFreeArgTypeTime {
		if freeArgTime, err := time.Parse(timeFormat, freeArgValue); err == nil {
			fixedArgItf := condition.GetFixedArg().GetValue()
			if fixedArgStr, ok := fixedArgItf.(string); ok {
				if fixedArgTime, err := time.Parse(timeFormat, fixedArgStr); err == nil {
					if freeArgTime.After(fixedArgTime) {
						compareResult = 1
					} else if freeArgTime.Before(fixedArgTime) {
						compareResult = -1
					}
					return e.evaluation(compareResult)
				}
			}
		}

	} else {
		// TODO eval, this should not happend
	}

	return false
}
