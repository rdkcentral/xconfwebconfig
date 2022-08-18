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

type IConditionEvaluator interface {
	Evaluate(*Condition, map[string]string) bool
	FreeArgType() string
	Operation() string
}

type FnEvaluateInternal func(string, interface{}) bool

type BaseEvaluator struct {
	freeArgType      string
	operation        string
	evaluateInternal FnEvaluateInternal
}

func NewBaseEvaluator(freeArgType string, operation string, evaluateInternal FnEvaluateInternal) *BaseEvaluator {
	return &BaseEvaluator{
		freeArgType:      freeArgType,
		operation:        operation,
		evaluateInternal: evaluateInternal,
	}
}

func (e *BaseEvaluator) FreeArgType() string {
	return e.freeArgType
}

func (e *BaseEvaluator) Operation() string {
	return e.operation
}

// TODO eval if this is necessary
func (e *BaseEvaluator) Validate(fixedArg *FixedArg) error {
	return nil
}

func (e *BaseEvaluator) Evaluate(condition *Condition, context map[string]string) bool {
	var freeArgValue string
	var ok bool
	if e.freeArgType != StandardFreeArgTypeVoid {
		freeArgValue, ok = context[condition.GetFreeArg().GetName()]
		if condition.Operation == StandardOperationExists {
			if ok {
				return true
			}
			return false
		}
		if !ok || (e.freeArgType != StandardFreeArgTypeAny && len(freeArgValue) == 0) {
			return false
		}
	}

	fixedArg := condition.GetFixedArg()
	if fixedArg != nil {
		return e.evaluateInternal(freeArgValue, fixedArg.GetValue())
	}
	return e.evaluateInternal(freeArgValue, nil)
}
