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
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"

	"xconfwebconfig/common"
	"xconfwebconfig/util"

	"github.com/aead/siphash"
)

const (
	voffset = float64(math.MaxInt64 + 1)
	vrange  = float64(math.MaxInt64*2 + 1)
)

func FitsPercent(itf interface{}, percent float64) bool {
	var bbytes []byte

	switch ty := itf.(type) {
	case string:
		bbytes = []byte(ty)
	case int64:
		bbytes = make([]byte, 8)
		binary.LittleEndian.PutUint64(bbytes, uint64(ty))
	default:
		return false
	}

	hashCode := float64(int64(siphash.Sum64(bbytes, &SipHashKey))) + voffset
	limit := percent / 100 * vrange
	return hashCode <= limit
}

func Copy(r Rule) Rule {
	// rp := &r
	result := Rule{}
	result.SetNegated(r.IsNegated())
	result.SetRelation(r.GetRelation())
	result.SetId(r.Id() + "_Copy")
	if !r.IsCompound() {
		result.SetCondition(r.GetCondition().Copy())
		return result
	}

	newCompoundParts := []Rule{}
	for _, cp := range r.GetCompoundParts() {
		newCompoundParts = append(newCompoundParts, Copy(cp))
	}
	result.SetCompoundParts(newCompoundParts)
	return result
}

func Not(r Rule) Rule {
	result := Copy(r)
	result.SetNegated(!r.IsNegated())
	return result
}

func Or(r Rule) Rule {
	result := Copy(r)
	result.SetRelation(RelationOr)
	return result
}

func And(r Rule) Rule {
	result := Copy(r)
	result.SetRelation(RelationAnd)
	return result
}

func AndRules(base Rule, compound Rule) Rule {
	return addRelatedCompound(base, compound, RelationAnd)
}

func OrRules(base Rule, compound Rule) Rule {
	return addRelatedCompound(base, compound, RelationOr)
}

func addRelatedCompound(base Rule, compound Rule, relation string) Rule {
	if base.GetCondition() == nil {
		return compound
	}

	var result Rule

	if !base.IsCompound() {
		result = Rule{}
		result.SetCompoundParts(make([]Rule, 0))
		result.AddCompoundPart(Copy(base))
	} else {
		result = base
	}

	compound.SetRelation(relation)
	result.AddCompoundPart(compound)
	return result
}

func FlattenRule(r Rule) []Rule {
	var result, tmpList []Rule
	tmpList = append(tmpList, r)
	for len(tmpList) > 0 {
		currentRule := tmpList[0]
		tmpList = tmpList[1:]

		if currentRule.Condition != nil {
			result = append(result, currentRule)
		}
		if len(currentRule.CompoundParts) > 0 {
			tmpList = append(tmpList, currentRule.CompoundParts...)
		}
	}

	return result
}

func removeIndex(ruleList []Rule, index int) (r Rule, rList []Rule) {
	r = ruleList[index]
	rList = append(ruleList[:index], ruleList[index+1:]...)
	return r, rList
}

func toConditions(rule Rule) []Condition {
	result := make([]Condition, 0)
	tmpRulesQ := make([]Rule, 0)
	tmpRulesQ = append(tmpRulesQ, rule)
	var curRule Rule
	for len(tmpRulesQ) > 0 {
		curRule, tmpRulesQ = removeIndex(tmpRulesQ, 0)
		if curRule.GetCondition() != nil {
			result = append(result, *curRule.GetCondition())
		}
		compParts := curRule.GetCompoundParts()
		if compParts != nil {
			tmpRulesQ = append(tmpRulesQ, compParts[0:]...)
		}
	}
	return result
}

func IsExistConditionByFreeArgName(rule Rule, freeArgName string) bool {
	for _, condition := range toConditions(rule) {
		if strings.Contains(strings.ToLower(condition.GetFreeArg().GetName()), strings.ToLower(freeArgName)) {
			return true
		}
	}
	return false
}

func IsExistConditionByFixedArgValue(rule Rule, fixedArgValue string) bool {
	for _, condition := range toConditions(rule) {
		if equalFixedArgCondition(condition, fixedArgValue) {
			return true
		}
	}
	return false
}

func IsExistConditionByFreeArgAndFixedArg(rule *Rule, freeArg string, fixedArg string) bool {
	isExist := false
	conditions := ToConditions(rule)
	for _, condition := range conditions {
		if condition == nil || condition.FixedArg == nil || condition.FreeArg == nil || !strings.EqualFold(condition.FreeArg.Name, freeArg) {
			continue
		}
		val := condition.GetFixedArg().GetValue()
		if reflect.ValueOf(val).Kind() == reflect.String {
			isExist = strings.EqualFold(val.(string), fixedArg)
		} else {
			data, ok := val.([]string)
			if ok {
				isExist = util.CaseInsensitiveContains(data, fixedArg)
			}
		}

		if isExist {
			break
		}
	}

	return isExist
}

func IsExistPartOfSearchValueInFixedArgs(fixedArgs Collection, searchValue string) bool {
	for _, fixedArg := range fixedArgs.Value {
		if strings.Contains(strings.ToLower(fixedArg), strings.ToLower(searchValue)) {
			return true
		}
	}
	return false
}

func equalFixedArgCondition(condition Condition, fixedArgValue string) bool {
	if condition.GetFixedArg() != nil && condition.GetFixedArg().IsCollectionValue() {
		return IsExistPartOfSearchValueInFixedArgs(condition.GetFixedArg().Collection, fixedArgValue)
	}
	if condition.GetFixedArg() != nil {
		return strings.Contains(strings.ToLower(condition.GetFixedArg().String()), strings.ToLower(fixedArgValue))
	}
	return false
}

func ToConditions(rule *Rule) []*Condition {
	result := []*Condition{}
	rulesQueue := []*Rule{rule}
	var currentRule *Rule

	for {
		currentRule, rulesQueue = RemElemFromRuleList(rulesQueue)
		if currentRule != nil {
			if currentRule.GetCondition() != nil {
				result = append(result, currentRule.GetCondition())
			}

			if currentRule.CompoundParts != nil && len(currentRule.CompoundParts) > 0 {
				for i := 0; i < len(currentRule.CompoundParts); i++ {
					//for _, extraRule := range currentRule.GetCompoundParts() {
					rulesQueue = append(rulesQueue, &currentRule.CompoundParts[i])
				}
			}
		}

		if len(rulesQueue) < 1 {
			break
		}
	}

	return result
}

// RemElemFromRuleList ... remove / popup a element from slice
func RemElemFromRuleList(rules []*Rule) (*Rule, []*Rule) {
	if len(rules) == 0 {
		return nil, rules
	}
	elem := rules[0]
	rules = append(rules[:0], rules[1:]...)
	return elem, rules
}

/**
 * provides compareTo implementation compatible with {@code java.util.Comparator<Rule>}
 * that can be used for rule ordering since it takes into account both rule
 * and priority of operations (ascending PERCENT < LIKE < IN < IS)
 *
 * @param r1 first rule to compare
 * @param r2 second rule to compare
 * @return comparison result according to {@link java.util.Comparator#compare(Object, Object)}
 */
func CompareRules(r1 Rule, r2 Rule) int {
	// based on Java code static int compare(boolean a,boolean b
	// a positive number if only a is true, a negative number if only b is true, or zero if a == b

	compoundResult := 0
	if r1.IsCompound() != r2.IsCompound() {
		if r1.IsCompound() {
			compoundResult = 1
		} else {
			compoundResult = -1
		}
	}

	if compoundResult != 0 {
		return compoundResult
	}

	op1 := getFirstChild(r1).Condition.Operation
	op2 := getFirstChild(r2).Condition.Operation

	// do we need using strings comoparison ?
	if op1 == op2 {
		return 0
	}

	switch strings.ToUpper(op1) {
	case "IS":
		return 1
	case "IN_LIST":
		if strings.ToUpper(op2) == "IS" {
			return -1
		}
		return 1

	case "LIKE":
		if strings.ToUpper(op2) == "PERCENT" {
			return 1
		}
		return -1

	case "PERCENT":
		return -1

	default:
		return 0
	}
}

func getFirstChild(rule Rule) Rule {
	if !rule.IsCompound() {
		return rule
	}
	return getFirstChild(rule.GetCompoundParts()[0])
}

func GetFixedArgsFromRuleByOperation(rule *Rule, operation string) []string {
	var result []string
	conditions := ToConditions(rule)
	for _, condition := range conditions {
		value := GetFixedArgFromConditionByOperation(condition, operation)
		if value != nil {
			switch ty := value.(type) {
			case string:
				result = append(result, ty)
			case float64:
				result = append(result, fmt.Sprintf("%v", ty))
			case []string:
				result = append(result, ty...)
			}
		}
	}
	return result
}

func GetFixedArgFromConditionByOperation(condition *Condition, operation string) interface{} {
	if condition != nil && condition.Operation == operation && condition.FixedArg != nil {
		return condition.FixedArg.GetValue()
	}
	return nil
}

//

func NormalizeConditions(rule *Rule) {
	conditions := ToConditions(rule)
	for _, condition := range conditions {
		NormalizeCondition(condition)
	}
}

func NormalizeCondition(condition *Condition) {
	freeArg := condition.FreeArg
	if freeArg != nil {
		freeArg.Name = strings.Trim(freeArg.Name, " ")
	}

	fixedArg := condition.FixedArg
	if fixedArg != nil {
		normalizeFixedArgValue(fixedArg, freeArg, condition.Operation)
	}

	normalizeMacAddress(condition)
	normalizePartnerId(condition)
}

func normalizeFixedArgValue(fixedArg *FixedArg, freeArg *FreeArg, operation string) {
	if fixedArg.IsCollectionValue() {
		for i, value := range fixedArg.Collection.Value {
			normalizedValue := strings.Trim(value, " ")
			fixedArg.Collection.Value[i] = modifyFixedArgDependingOnFreeArgAndOperation(normalizedValue, freeArg, operation)
		}
	} else if len(fixedArg.Bean.Value.JLString) > 0 {
		normalizedValue := strings.Trim(fixedArg.Bean.Value.JLString, " ")
		fixedArg.Bean.Value.JLString = modifyFixedArgDependingOnFreeArgAndOperation(normalizedValue, freeArg, operation)
	}
}

func normalizeMacAddress(condition *Condition) {
	macAddressNames := []string{common.ESTB_MAC, common.ECM_MAC, common.ECM_MAC_ADDRESS, common.ESTB_MAC_ADDRESS}
	if condition.FixedArg != nil && condition.FixedArg.GetValue() != nil &&
		condition.FreeArg != nil && util.Contains(macAddressNames, condition.FreeArg.Name) {
		if StandardOperationIs == condition.Operation || StandardOperationLike == condition.Operation {
			normalizedMac := util.NormalizeMacAddress(condition.FixedArg.Bean.Value.JLString)
			condition.FixedArg.Bean.Value.JLString = normalizedMac
		} else if StandardOperationIn == condition.Operation {
			for i, value := range condition.FixedArg.Collection.Value {
				normalizedMac := util.NormalizeMacAddress(value)
				condition.FixedArg.Collection.Value[i] = normalizedMac
			}
		}
	}
}

func normalizePartnerId(condition *Condition) {
	if condition.FixedArg != nil && condition.FixedArg.GetValue() != nil &&
		condition.FreeArg != nil && condition.FreeArg.Name == common.PARTNER_ID {
		if StandardOperationIs == condition.Operation {
			condition.FixedArg.Bean.Value.JLString = strings.ToUpper(condition.FixedArg.Bean.Value.JLString)
		} else if StandardOperationIn == condition.Operation {
			for i, value := range condition.FixedArg.Collection.Value {
				condition.FixedArg.Collection.Value[i] = strings.ToUpper(value)
			}
		}
	}
}

func modifyFixedArgDependingOnFreeArgAndOperation(fixedArgValue string, freeArg *FreeArg, operation string) string {
	if isEnvOrModelFreeArgByOperation(freeArg, operation) || isMacAddressFreeArgByOperation(freeArg, operation) {
		return strings.ToUpper(fixedArgValue)
	}
	return fixedArgValue
}

func isEnvOrModelFreeArgByOperation(freeArg *FreeArg, operation string) bool {
	return freeArg != nil && (freeArg.Name == common.MODEL || freeArg.Name == common.ENV) &&
		(operation == StandardOperationIs || operation == StandardOperationIn)
}

func isMacAddressFreeArgByOperation(freeArg *FreeArg, operation string) bool {
	return freeArg != nil && (freeArg.Name == common.ESTB_MAC || freeArg.Name == common.ESTB_MAC_ADDRESS) &&
		(operation == StandardOperationIs || operation == StandardOperationIn)
}

func GetDuplicateConditionsBetweenOR(rule *Rule) (result []Condition) {
	rules := FlattenRule(*rule)
	split := []Condition{}
	for _, one := range rules {
		if RelationOr == one.Relation {
			result = append(result, GetDuplicateConditions(split)...)
			split = []Condition{}
		}
		split = append(split, *one.Condition)
	}
	result = append(result, GetDuplicateConditions(split)...)
	return result
}

func GetDuplicateConditions(conditions []Condition) (result []Condition) {
	tempList := make(map[Condition]int)
	for _, k := range conditions {
		_, ok := tempList[k]
		if ok {
			tempList[k]++
		} else {
			tempList[k] = 1
		}
	}
	for key, _ := range tempList {
		if tempList[key] == 1 {
			delete(tempList, key)
		} else {
			result = append(result, key)
		}
	}

	return result
}
