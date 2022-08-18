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
	"path/filepath"
	"strconv"
	"strings"

	"xconfwebconfig/db"

	log "github.com/sirupsen/logrus"
)

type RuleProcessorFactory struct {
	Processor *RuleProcessor
}

func NewRuleProcessorFactory() *RuleProcessorFactory {
	log.Debug("NewRuleProcessorFactory Call dao GetAllAsList GenericXconfNamedList start ... ")
	processor := NewRuleProcessor()
	customizedEvaluators := []IConditionEvaluator{}

	// ==== build the customizedEvaluators ====
	nsListDao := db.GetCachedSimpleDao()
	// nsListDao := db.GetCompressingDataDao()
	ipEval := NewIpAddressEvaluator(
		AuxFreeArgTypeIpAddress,
		StandardOperationInList,
		nsListDao,
	)
	customizedEvaluators = append(customizedEvaluators, ipEval)

	nsEval := NewIpAddressEvaluator(
		StandardFreeArgTypeString,
		StandardOperationInList,
		nsListDao,
	)
	customizedEvaluators = append(customizedEvaluators, nsEval)

	// ==== TODO eval if MatchOperationEvaluator is needed ====
	matchEval := NewBaseEvaluator(
		StandardFreeArgTypeString,
		StandardOperationMatch,
		func(freeArgValue string, fixedArgValue interface{}) bool {
			if pattern, ok := fixedArgValue.(string); ok {
				matched, err := filepath.Match(pattern, freeArgValue)
				if err == nil {
					return matched
				}
			}
			return false
		},
	)
	customizedEvaluators = append(customizedEvaluators, matchEval)

	// ==== TODO eval if MatchOperationEvaluator is needed ====
	percEval := NewBaseEvaluator(
		StandardFreeArgTypeString,
		StandardOperationRange,
		evalRange,
	)
	customizedEvaluators = append(customizedEvaluators, percEval)

	processor.AddEvaluators(customizedEvaluators)

	return &RuleProcessorFactory{
		Processor: processor,
	}
}

func (f *RuleProcessorFactory) RuleProcessor() *RuleProcessor {
	return f.Processor
}

func evalRange(freeArgValue string, fixedArgValueItf interface{}) bool {
	quotedFreeArgValue := fmt.Sprintf(`"%v"`, freeArgValue)

	fixedArgValue, ok := fixedArgValueItf.(string)
	if !ok {
		return false
	}

	elements := strings.Split(fixedArgValue, "-")
	if len(elements) != 2 {
		return false
	}

	lowRange, err := strconv.ParseFloat(elements[0], 64)
	if err != nil || lowRange < 0 {
		return false
	}
	highRange, err := strconv.ParseFloat(elements[1], 64)
	if err != nil || highRange <= 0 {
		return false
	}
	return !FitsPercent(quotedFreeArgValue, lowRange) && FitsPercent(quotedFreeArgValue, highRange)
}
