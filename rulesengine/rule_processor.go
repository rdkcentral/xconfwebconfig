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
	"time"

	log "github.com/sirupsen/logrus"
)

var UseMap bool

type RuleProcessor struct {
	evaluators   []IConditionEvaluator
	evaluatorMap map[string]IConditionEvaluator
}

func NewRuleProcessor() *RuleProcessor {
	evaluators := GetStandardEvaluators()
	evaluators = append(evaluators, GetAuxEvaluators()...)
	evaluatorMap := make(map[string]IConditionEvaluator)
	for _, evaluator := range evaluators {
		ttype := evaluator.FreeArgType()
		op := evaluator.Operation()
		evaluatorMap[ttype+"_"+op] = evaluator
	}
	return &RuleProcessor{
		evaluators:   evaluators,
		evaluatorMap: evaluatorMap,
	}
}

func (p *RuleProcessor) AddEvaluators(evs []IConditionEvaluator) {
	p.evaluators = append(p.evaluators, evs...)
	for _, evaluator := range evs {
		ttype := evaluator.FreeArgType()
		op := evaluator.Operation()
		p.evaluatorMap[ttype+"_"+op] = evaluator
	}
}

func (p *RuleProcessor) Size() int {
	return len(p.evaluators)
}

func (p *RuleProcessor) Filter(rules []Rule, context map[string]string) []Rule {
	matched := []Rule{}
	for _, rule := range rules {
		if p.Evaluate(&rule, context, log.Fields{}) {
			matched = append(matched, rule)
		}
	}
	return matched
}

func (p *RuleProcessor) Evaluate(r *Rule, context map[string]string, fields log.Fields, vargNegations ...bool) bool {
	if len(vargNegations) > 0 {
		return p.evaluateWithNegation(r, vargNegations[0], context)
	}

	if !r.IsCompound() {
		return p.evaluateWithNegation(r, r.IsNegated(), context)
	}

	var result bool
	for i, cp := range r.GetCompoundParts() {
		if i == 0 {
			result = p.Evaluate(&cp, context, fields)
		} else {
			relation := cp.GetRelation()
			if result && relation == RelationOr {
				continue
			}
			if !result && relation == RelationAnd {
				break
			}
			result = p.Evaluate(&cp, context, fields)
		}
	}
	if r.IsNegated() {
		result = !result
	}
	return result
}

func (p *RuleProcessor) EvaluateTest(r *Rule, context map[string]string, fields log.Fields, vargNegations ...bool) bool {
	start := time.Now()
	if len(vargNegations) > 0 {
		return p.evaluateWithNegation(r, vargNegations[0], context)
	}
	if (fields["firmware_id"] == "5e7aec2c-c3bb-4cb6-9e62-25b00666e08f") || (fields["firmware_id"] == "8945ffb3-c1c3-4fd7-9958-8491f175b482") || (fields["firmware_id"] == "c0cbf02e-6007-499e-a53e-d03dcd1ce538") {
		log.WithFields(fields).Debugf("Evaluate after vargNegations finished in %v", time.Since(start))
	}

	if r.Condition != nil {
		// if !r.IsCompound() {
		return p.evaluateWithNegation(r, r.IsNegated(), context)
	}
	if (fields["firmware_id"] == "5e7aec2c-c3bb-4cb6-9e62-25b00666e08f") || (fields["firmware_id"] == "8945ffb3-c1c3-4fd7-9958-8491f175b482") || (fields["firmware_id"] == "c0cbf02e-6007-499e-a53e-d03dcd1ce538") {
		log.WithFields(fields).Debugf("Evaluate after IsCompound finished in %v", time.Since(start))
	}

	var result bool
	for i, cp := range r.GetCompoundParts() {
		if i == 0 {
			result = p.Evaluate(&cp, context, fields)
			if (fields["firmware_id"] == "5e7aec2c-c3bb-4cb6-9e62-25b00666e08f") || (fields["firmware_id"] == "8945ffb3-c1c3-4fd7-9958-8491f175b482") || (fields["firmware_id"] == "c0cbf02e-6007-499e-a53e-d03dcd1ce538") {
				log.WithFields(fields).Debugf("Evaluate after i==0 finished in %v", time.Since(start))
			}
		} else {
			relation := cp.GetRelation()
			if result && relation == RelationOr {
				continue
			}
			if !result && relation == RelationAnd {
				break
			}
			result = p.Evaluate(&cp, context, fields)
			if (fields["firmware_id"] == "5e7aec2c-c3bb-4cb6-9e62-25b00666e08f") || (fields["firmware_id"] == "8945ffb3-c1c3-4fd7-9958-8491f175b482") || (fields["firmware_id"] == "c0cbf02e-6007-499e-a53e-d03dcd1ce538") {
				log.WithFields(fields).Debugf("Evaluate after relation finished in %v", time.Since(start))
			}
		}
	}
	if r.IsNegated() {
		result = !result
	}
	if (fields["firmware_id"] == "5e7aec2c-c3bb-4cb6-9e62-25b00666e08f") || (fields["firmware_id"] == "8945ffb3-c1c3-4fd7-9958-8491f175b482") || (fields["firmware_id"] == "c0cbf02e-6007-499e-a53e-d03dcd1ce538") {
		log.WithFields(fields).Debugf("Evaluate finished in %v", time.Since(start))
	}
	return result
}

func (p *RuleProcessor) evaluateWithNegation(r *Rule, negation bool, context map[string]string) bool {
	condition := r.GetCondition()
	evaluator := p.getEvaluator(condition.GetFreeArg().GetType(), condition.GetOperation())
	if evaluator == nil {
		fmt.Printf("type=%v, operation=%v\n", condition.GetFreeArg().GetType(), condition.GetOperation())
	}
	result := evaluator.Evaluate(condition, context)
	if negation {
		result = !result
	}
	return result
}

func (p *RuleProcessor) getEvaluator(ttype string, operation string) IConditionEvaluator {
	if UseMap {
		return p.evaluatorMap[ttype+"_"+operation]
	}
	for _, evaluator := range p.evaluators {
		if evaluator.FreeArgType() == ttype && evaluator.Operation() == operation {
			return evaluator
		}
	}
	return nil
}

func (p *RuleProcessor) GetEvaluatorOK(r *Rule) bool {
	if !r.IsCompound() {
		condition := r.GetCondition()
		if condition == nil {
			return false
		}
		if condition.GetFreeArg() == nil {
			return false
		}

		if condition.GetFixedArg() == nil {
			if condition.GetOperation() != StandardOperationExists {
				return false
			}
		}
		evaluator := p.getEvaluator(condition.GetFreeArg().GetType(), condition.GetOperation())
		if evaluator == nil {
			fmt.Printf("type=%v, operation=%v\n", condition.GetFreeArg().GetType(), condition.GetOperation())
			return false
		}
		return true
	}

	for _, cp := range r.GetCompoundParts() {
		if !p.GetEvaluatorOK(&cp) {
			return false
		}
	}
	return true
}
