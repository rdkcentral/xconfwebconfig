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
	"bytes"
	"fmt"
)

const (
	LeftParan  = "("
	RightParan = ")"
	SpaceChar  = " "
)

type Rule struct {
	CompoundParts []Rule     `json:"compoundParts,omitempty"`
	Condition     *Condition `json:"condition,omitempty"`
	Negated       bool       `json:"negated"`
	Relation      string     `json:"relation,omitempty"`
	Xxid          string     `json:"xxid,omitempty"` // temp use for testing
}

// XRule is ...
type XRule interface {
	GetId() string
	GetRule() *Rule
	GetName() string
	GetTemplateId() string
	GetRuleType() string
}

func NewEmptyRule() *Rule {
	return &Rule{
		CompoundParts: make([]Rule, 0),
	}
}

func (r *Rule) IsNegated() bool {
	return r.Negated
}

func (r *Rule) SetNegated(negated bool) {
	r.Negated = negated
}

func (r *Rule) Id() string {
	return r.Xxid
}

func (r *Rule) SetId(xxid string) {
	r.Xxid = xxid
}

func (r *Rule) GetRelation() string {
	return r.Relation
}

func (r *Rule) SetRelation(relation string) {
	r.Relation = relation
}

func (r *Rule) GetCondition() *Condition {
	return r.Condition
}

func (r *Rule) SetCondition(condition *Condition) {
	r.Condition = condition
}

func (r *Rule) GetCompoundParts() []Rule {
	return r.CompoundParts
}

func (r *Rule) SetCompoundParts(compoundParts []Rule) {
	r.CompoundParts = compoundParts
}

func (r *Rule) AddCompoundPart(rule Rule) {
	r.CompoundParts = append(r.CompoundParts, rule)
}

func (r *Rule) IsCompound() bool {
	// return r.CompoundParts != nil && len(r.CompoundParts) > 0
	return r.Condition == nil
}

func (r *Rule) GetFreeArg() *FreeArg {
	return r.GetCondition().GetFreeArg()
}

func (r *Rule) negatedString() string {
	if r.Negated {
		return "NOT "
	}
	return ""
}

func (r *Rule) relationString() string {
	if len(r.Relation) > 0 {
		return fmt.Sprintf("%v ", r.Relation)
	}
	return ""
}

func (r *Rule) conditionString() string {
	if r.Condition == nil {
		return ""
	}
	return r.Condition.String()
}

func (r *Rule) String() string {
	// return fmt.Sprintf("Rule(Negated=%v, Relation='%v', Condition='%v', CompoundParts='%v')", r.Negated, r.Relation, r.Condition, r.CompoundParts)
	if len(r.CompoundParts) > 0 {
		start := fmt.Sprintf("Rule(%v%v%v", r.relationString(), r.negatedString(), r.conditionString())
		buffer := bytes.NewBufferString(start)
		buffer.WriteString("  CompoundParts:[\n")
		for _, cp := range r.CompoundParts {
			buffer.WriteString(fmt.Sprintf("    %v\n", cp.String()))
		}
		buffer.WriteString("  ]")
		buffer.WriteString(")")
		return buffer.String()
	}
	return fmt.Sprintf("Rule(%v%v%v)", r.relationString(), r.negatedString(), r.conditionString())
}

// NOTE: my understanding is the ordering of the CompoundParts matters
func (r *Rule) Equals(x *Rule) bool {
	if r.IsNegated() != x.IsNegated() {
		return false
	}

	if r.GetRelation() != x.GetRelation() {
		return false
	}

	if r.Condition == nil && !r.IsCompoundPartsEmpty() && len(r.CompoundParts) == 1 {
		r = &r.CompoundParts[0]
	}
	if x.Condition == nil && !x.IsCompoundPartsEmpty() && len(x.CompoundParts) == 1 {
		x = &x.CompoundParts[0]
	}

	if r.GetCondition() == nil && x.GetCondition() != nil {
		return false
	}
	if r.GetCondition() != nil && x.GetCondition() == nil {
		return false
	}
	if r.GetCondition() != nil && x.GetCondition() != nil {
		if !r.GetCondition().Equals(x.GetCondition()) {
			return false
		}
	}
	if r.IsCompoundPartsEmpty() && !x.IsCompoundPartsEmpty() {
		return false
	}
	if !r.IsCompoundPartsEmpty() && x.IsCompoundPartsEmpty() {
		return false
	}
	if !r.IsCompoundPartsEmpty() && !x.IsCompoundPartsEmpty() {
		if len(r.GetCompoundParts()) != len(x.GetCompoundParts()) {
			return false
		}

		if len(r.GetCompoundParts()) > 0 {
			for i, cp := range r.GetCompoundParts() {
				xcp := x.GetCompoundParts()[i]
				if !cp.Equals(&xcp) {
					return false
				}
			}
		}
	}
	return true
}

func (r *Rule) GetInListNames() []string {
	names := []string{}
	if len(r.CompoundParts) == 0 {
		condition := r.GetCondition()
		if condition != nil {
			if condition.GetOperation() == StandardOperationInList {
				fixedArg := condition.GetFixedArg()
				name := fixedArg.GetValue().(string)
				names = append(names, name)
			}
		}
	}
	for _, cp := range r.CompoundParts {
		cpNames := cp.GetInListNames()
		names = append(names, cpNames...)
	}
	return names
}

func (r *Rule) GetTree() string {
	if r.Condition != nil {
		return r.Xxid
	}
	left := r.CompoundParts[0].GetTree()
	right := r.CompoundParts[1].GetTree()
	relation := "AND"
	if r.Relation == RelationOr {
		relation = "OR"
	}
	return LeftParan + left + SpaceChar + relation + SpaceChar + right + RightParan
}
