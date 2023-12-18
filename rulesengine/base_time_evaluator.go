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
)

// NOTE:
//
//	The development of this function is postponed because I think this type is not used
//	in the current prod/ci data set.
type FnIntEval func(int) bool

type BaseTimeEvaluator struct {
	freeArgType      string
	operation        string
	evaluateInternal FnIntEval
}

func NewBaseTimeEvaluator(freeArgType string, operation string, fn FnIntEval) *BaseTimeEvaluator {
	return &BaseTimeEvaluator{
		freeArgType:      freeArgType,
		operation:        operation,
		evaluateInternal: fn,
	}
}

func (e *BaseTimeEvaluator) FreeArgType() string {
	return e.freeArgType
}

func (e *BaseTimeEvaluator) Operation() string {
	return e.operation
}

func (e *BaseTimeEvaluator) Evaluate(c *Condition, context map[string]string) bool {
	ok := false
	if !ok {
		panic(fmt.Errorf("BaseTimeEvaluator.Evaluate() is not implemented yet"))
	}

	var freeArgValue string
	if fa := c.GetFreeArg(); fa != nil {
		freeArgValue = context[fa.GetName()]
	}
	if len(freeArgValue) == 0 {
		return false
	}
	_ = freeArgValue

	return false
}
