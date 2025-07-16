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
	"net/http"
	"reflect"
	"strings"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/aead/siphash"
)

const (
	voffset = float64(math.MaxInt64 + 1)
	vrange  = float64(math.MaxInt64*2 + 1)
)

type ruleCount struct {
	rule  Rule
	count int
}

type conditionCount struct {
	condition Condition
	count     int
}

func contains(carray []conditionCount, cond Condition) int {
	for i, c := range carray {
		if cond.Equals(&c.condition) {
			return i
		}
	}
	return -1
}

func FitsPercent(itf interface{}, percent float64) bool {
	var bbytes []byte

	switch ty := itf.(type) {
	case string:
		bbytes = []byte(fmt.Sprintf(`"%v"`, ty))
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

func GetFixedArgsFromRuleByFreeArgAndOperation(rule Rule, freeArg string, operation string) []string {
	var result []string
	conditions := ToConditions(&rule)
	for _, condition := range conditions {
		value := GetFixedArgFromConditionByFreeArgAndOperation(*condition, freeArg, operation)
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

func GetFixedArgFromConditionByFreeArgAndOperation(condition Condition, freeArg string, operation string) interface{} {
	if operation == condition.Operation && freeArg == condition.FreeArg.Name {
		if condition.FixedArg != nil {
			return condition.FixedArg.GetValue()
		}
	}

	return nil
}

func Not(r Rule) Rule {
	result := Copy(r)
	result.SetNegated(!r.IsNegated())
	return result
}

func (r *Rule) IsCompoundPartsEmpty() bool {
	return r.CompoundParts == nil || len(r.CompoundParts) == 0
}

func (r *Rule) IsEmpty() bool {
	if r == nil {
		return true
	}
	if r.IsCompoundPartsEmpty() && r.Condition == nil {
		return true
	}
	return false
}

func EqualComplexRules(rule1 *Rule, rule2 *Rule) bool {
	if rule1.IsEmpty() && rule2.IsEmpty() {
		return true
	}
	if rule1 == nil || rule2 == nil {
		return false
	}
	result := false
	flattenedRule1 := FlattenRule(*rule1)
	flattenedRule2 := FlattenRule(*rule2)
	if len(flattenedRule1) != len(flattenedRule2) {
		return false
	}
	if len(flattenedRule1) > 1 &&
		allRulesHaveSameRelation(flattenedRule1) &&
		allRulesHaveSameRelation(flattenedRule2) {
		sameRelation := flattenedRule1[len(flattenedRule1)-1].GetRelation()
		flattenedRule1[0].SetRelation(sameRelation)
		flattenedRule2[0].SetRelation(sameRelation)
		result = equalNonCompoundRulesCollections(flattenedRule1, flattenedRule2)
		sameRelation = ""
		flattenedRule1[0].SetRelation(sameRelation)
		flattenedRule2[0].SetRelation(sameRelation)
	} else {
		result = equalNonCompoundRulesCollectionsRegardingTheOrder(flattenedRule1, flattenedRule2)
	}

	return result
}

func equalNonCompoundRulesCollections(list1 []Rule, list2 []Rule) bool {
	if len(list1) == len(list2) {
		return len(intersectionOfNonCompoundRules(list1, list2)) == len(list1)
	}
	return false
}

func intersectionOfNonCompoundRules(rules1 []Rule, rules2 []Rule) (result []Rule) {
	for i := 0; i < len(rules1); i++ {
		for j := 0; j < len(rules2); j++ {
			if equalNonCompoundRules(rules1[i], rules2[j]) {
				result = append(result, rules1[i])
				rules2 = append(rules2[:j], rules2[j+1:]...)
				break
			}
		}
	}
	return result
}

func equalNonCompoundRulesCollectionsRegardingTheOrder(list1 []Rule, list2 []Rule) bool {
	if len(list1) != len(list2) {
		return false
	}

	for i := 0; i < len(list1); i++ {
		if !equalNonCompoundRules(list1[i], list2[i]) {
			return false
		}
	}

	return true
}

func equalNonCompoundRules(rule1 Rule, rule2 Rule) bool {
	if &rule1 == nil && &rule2 == nil {
		return true
	}
	if &rule1 == nil || &rule2 == nil {
		return false
	}
	if rule1.IsNegated() != rule2.IsNegated() {
		return false
	}
	if rule1.Relation != rule2.Relation {
		return false
	}

	return equalConditions(rule1.Condition, rule2.Condition)
}

func equalConditions(condition1 *Condition, condition2 *Condition) bool {
	if condition1 == nil && condition2 == nil {
		return true
	}

	if condition1 == nil || condition2 == nil {
		return false
	}

	if condition1.FreeArg != nil && condition2.FreeArg != nil {
		if condition1.FreeArg.Name != condition2.FreeArg.Name {
			return false
		}
	} else if condition1.FreeArg == nil || condition2.FreeArg == nil {
		return false
	}
	if condition1.Operation != "" && condition2.Operation != "" {
		if condition1.Operation != condition2.Operation {
			return false
		}
	} else if condition1.Operation == "" || condition2.Operation == "" {
		return false
	}

	if condition1.FixedArg != nil && condition2.FixedArg != nil {
		return condition1.FixedArg.Equals(condition2.FixedArg)
	} else if condition1.FixedArg == nil || condition2.FixedArg == nil {
		return false
	}

	return true
}

func allRulesHaveSameRelation(rules []Rule) bool {
	sz := len(rules)
	for i := 0; i < sz; i++ {
		if util.IsBlank(rules[i].GetRelation()) {
			continue
		}
		if rules[sz-1].GetRelation() != rules[i].GetRelation() {
			return false
		}
	}

	return true
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
	// if base.GetCondition() == nil {
	// 	return compound
	// }

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

func IsExistPartOfSearchValueInFixedArgs(fixedArgs *Collection, searchValue string) bool {
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

func GetDuplicateFixedArgListItems(fixedArg FixedArg) (retList []string) {
	if fixedArg.IsCollectionValue() {
		colMap := make(map[string]int)
		for _, val := range fixedArg.Collection.Value {
			colMap[val] = colMap[val] + 1
		}
		for k, v := range colMap {
			if v != 1 {
				retList = append(retList, k)
			}
		}
	}
	return retList
}

func CheckFreeArgExists2(conditionInfos []ConditionInfo, freeArg FreeArg) error {
	if !FreeArgExists(conditionInfos, freeArg) {
		return common.NewRemoteError(http.StatusBadRequest, freeArg.GetName()+" does not exist")
	}
	return nil
}

func FreeArgExists(conditionInfos []ConditionInfo, freeArg FreeArg) bool {
	for _, conditionInfo := range conditionInfos {
		if conditionInfo.FreeArg.Equals(&freeArg) {
			return true
		}
	}
	return false
}

func GetConditionInfos(conditions []*Condition) (result []ConditionInfo) {
	for _, condition := range conditions {
		condInfo := NewConditionInfo(*condition.GetFreeArg(), condition.GetOperation())
		result = append(result, *condInfo)
	}
	return result
}

func CheckFreeArgExists3(conditionInfos []ConditionInfo, freeArg FreeArg, operation string) error {
	if !FreeArgExists2(conditionInfos, freeArg, operation) {
		return common.NewRemoteError(http.StatusBadRequest, freeArg.GetName()+" with "+operation+" operation is required")
	}
	return nil
}

func FreeArgExists2(conditionInfos []ConditionInfo, freeArg FreeArg, operation string) bool {
	for _, conditionInfo := range conditionInfos {
		if conditionInfo.FreeArg.Equals(&freeArg) && conditionInfo.Operation == operation {
			return true
		}
	}
	return false
}

func ChangeFixedArgToNewValue(oldFixedArgValue string, newFixedArgValue string, rule Rule, operation string) bool {
	isChanged := false
	conditions := ToConditions(&rule)
	for _, condition := range conditions {
		if condition.Operation == operation && condition.FixedArg != nil && condition.FixedArg.IsStringValue() {
			fixedArgValue := condition.FixedArg.GetValue().(string)
			if fixedArgValue == oldFixedArgValue {
				condition.FixedArg.Bean.Value.JLString = &newFixedArgValue
				isChanged = true
			}
		}
	}
	return isChanged
}

func GetDuplicateConditionsFromRule(rule Rule) (result []Condition) {
	if &rule == nil {
		return result
	}

	duplicateRules := getDuplicateNonCompoundRules(FlattenRule(rule))

	for _, duplicateRule := range duplicateRules {
		result = append(result, *duplicateRule.GetCondition())
	}

	return result
}

//

func ruleArrayContains(carray []ruleCount, rule Rule) int {
	for i, c := range carray {
		if rule.Equals(&c.rule) {
			return i
		}
	}
	return -1
}

func ConditionHasEmptyElements(rule Rule) bool {
	return &rule == nil || rule.GetCondition() == nil || rule.GetFreeArg() == nil || rule.GetCondition().GetFixedArg() == nil
}

func getDuplicateNonCompoundRules(nonCompoundRules []Rule) (result []Rule) {
	if len(nonCompoundRules) > 1 {
		nonCompoundRules[0].SetRelation(RelationAnd)
		ruleCounts := []ruleCount{}
		for _, rule := range nonCompoundRules {
			pos := ruleArrayContains(ruleCounts, rule)
			if pos != -1 {
				ruleCounts[pos].count++
			} else {
				temp := ruleCount{}
				temp.count = 1
				temp.rule = rule
				ruleCounts = append(ruleCounts, temp)
			}
		}
		for _, c := range ruleCounts {
			if c.count != 1 {
				result = append(result, c.rule)
			}
		}
		nonCompoundRules[0].SetRelation("")
	}
	return result
}

func NormalizeConditions(rule *Rule) error {
	if rule == nil {
		return nil
	}
	conditions := ToConditions(rule)
	for _, condition := range conditions {
		if err := NormalizeCondition(condition); err != nil {
			return err
		}
	}
	return nil
}

func NormalizeCondition(condition *Condition) error {
	if condition == nil {
		return nil
	}
	freeArg := condition.FreeArg
	if freeArg != nil {
		freeArg.Name = strings.Trim(freeArg.Name, " ")
	}

	fixedArg := condition.FixedArg
	if fixedArg != nil {
		normalizeFixedArgValue(fixedArg, freeArg, condition.Operation)
	}

	if err := normalizeMacAddress(condition); err != nil {
		return err
	}
	normalizePartnerId(condition)
	return nil
}

func normalizeFixedArgValue(fixedArg *FixedArg, freeArg *FreeArg, operation string) {
	if fixedArg == nil || freeArg == nil {
		return
	}
	if fixedArg.IsCollectionValue() {
		for i, value := range fixedArg.Collection.Value {
			normalizedValue := strings.Trim(value, " ")
			fixedArg.Collection.Value[i] = modifyFixedArgDependingOnFreeArgAndOperation(normalizedValue, freeArg, operation)
		}
	} else if fixedArg.IsStringValue() {
		normalizedValue := strings.Trim(*fixedArg.Bean.Value.JLString, " ")
		var tmp = modifyFixedArgDependingOnFreeArgAndOperation(normalizedValue, freeArg, operation)
		fixedArg.Bean.Value.JLString = &tmp
	}
}

func normalizeMacAddress(condition *Condition) error {
	if condition == nil {
		return nil
	}
	macAddressNames := []string{common.ESTB_MAC, common.ECM_MAC, common.ECM_MAC_ADDRESS, common.ESTB_MAC_ADDRESS}
	if condition.FixedArg != nil && condition.FixedArg.GetValue() != nil &&
		condition.FreeArg != nil && util.Contains(macAddressNames, condition.FreeArg.Name) {
		if StandardOperationIs == condition.Operation || StandardOperationLike == condition.Operation {
			if condition.FixedArg.IsStringValue() {
				rawVal := *condition.FixedArg.Bean.Value.JLString
				normalizedMac, err := util.ValidateAndNormalizeMacAddress(rawVal)
				if err != nil {
					return common.NewRemoteError(http.StatusBadRequest, "Invalid Mac Address:"+rawVal)
				}
				condition.FixedArg.Bean.Value.JLString = &normalizedMac
			}
		} else if StandardOperationIn == condition.Operation {
			for i, value := range condition.FixedArg.Collection.Value {
				normalizedMac, err := util.ValidateAndNormalizeMacAddress(value)
				if err != nil {
					return common.NewRemoteError(http.StatusBadRequest, "Invalid Mac Address:"+value)
				}
				condition.FixedArg.Collection.Value[i] = normalizedMac
			}
		}
	}
	return nil
}

func normalizePartnerId(condition *Condition) {
	if condition == nil {
		return
	}
	if condition.FixedArg != nil && condition.FixedArg.GetValue() != nil &&
		condition.FreeArg != nil && condition.FreeArg.Name == common.PARTNER_ID {
		if StandardOperationIs == condition.Operation {
			if condition.FixedArg.IsStringValue() {
				var tmp = strings.ToUpper(*condition.FixedArg.Bean.Value.JLString)
				condition.FixedArg.Bean.Value.JLString = &tmp
			}
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

func GetDuplicateConditionsBetweenOR(rule Rule) (result []Condition) {
	if &rule == nil {
		return result
	}
	rules := FlattenRule(rule)
	split := []Condition{}
	for _, one := range rules {
		if RelationOr == one.Relation {
			result = append(result, GetDuplicateConditionsForAdmin(split)...)
			split = []Condition{}
		}
		split = append(split, *one.Condition)
	}
	result = append(result, GetDuplicateConditionsForAdmin(split)...)
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

func GetDuplicateConditionsForAdmin(conditions []Condition) (result []Condition) {
	condCounts := []conditionCount{}
	for _, cond := range conditions {
		pos := contains(condCounts, cond)
		if pos != -1 {
			condCounts[pos].count++
		} else {
			temp := conditionCount{}
			temp.count = 1
			temp.condition = cond
			condCounts = append(condCounts, temp)
		}
	}
	for _, c := range condCounts {
		if c.count != 1 {
			result = append(result, c.condition)
		}
	}
	return result
}
