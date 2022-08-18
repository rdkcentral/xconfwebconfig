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

type Condition struct {
	FreeArg   *FreeArg  `json:"freeArg"`
	Operation string    `json:"operation"`
	FixedArg  *FixedArg `json:"fixedArg"`
}

func NewCondition(freeArg *FreeArg, operation string, fixedArg *FixedArg) *Condition {
	return &Condition{
		FreeArg:   freeArg,
		Operation: operation,
		FixedArg:  fixedArg,
	}
}

func (c *Condition) GetFreeArg() *FreeArg {
	return c.FreeArg
}

func (c *Condition) SetFreeArg(freeArg *FreeArg) {
	c.FreeArg = freeArg
}

func (c *Condition) GetFixedArg() *FixedArg {
	return c.FixedArg
}

func (c *Condition) SetFixedArg(fixedArg *FixedArg) {
	c.FixedArg = fixedArg
}

func (c *Condition) GetOperation() string {
	return c.Operation
}

func (c *Condition) SetOperation(operation string) {
	c.Operation = operation
}

func (c *Condition) String() string {
	// return fmt.Sprintf("Condition(FreeArg=%v, FixedArg=%v, Operation='%v')", c.FreeArg, c.FixedArg, c.Operation)
	if c == nil {
		return "Condition(nil)"
	}
	return fmt.Sprintf("Condition(%v %v %v)", c.FreeArg, c.Operation, c.FixedArg)
}

func (c *Condition) Copy() *Condition {
	return NewCondition(c.GetFreeArg().Copy(), c.GetOperation(), c.GetFixedArg().Copy())
}

func (c *Condition) Equals(x *Condition) bool {
	if c.GetOperation() != x.GetOperation() {
		return false
	}

	if c.GetFreeArg() == nil && x.GetFreeArg() != nil {
		return false
	}
	if c.GetFreeArg() != nil && x.GetFreeArg() == nil {
		return false
	}
	if c.GetFreeArg() != nil && x.GetFreeArg() != nil {
		if !c.GetFreeArg().Equals(x.GetFreeArg()) {
			return false
		}
	}

	if c.GetFixedArg() == nil && x.GetFixedArg() != nil {
		return false
	}
	if c.GetFixedArg() != nil && x.GetFixedArg() == nil {
		return false
	}
	if c.GetFixedArg() != nil && x.GetFixedArg() != nil {
		if !c.GetFixedArg().Equals(x.GetFixedArg()) {
			return false
		}
	}

	return true
}

// ConditionInfo is ...
type ConditionInfo struct {
	FreeArg   FreeArg
	Operation string
}

// NewConditionInfo create a new instance
func NewConditionInfo(freeArg FreeArg, operation string) *ConditionInfo {
	return &ConditionInfo{
		FreeArg:   freeArg,
		Operation: operation,
	}
}
