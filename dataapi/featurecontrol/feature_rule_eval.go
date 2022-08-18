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
package featurecontrol

import (

	// "crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	common "xconfwebconfig/common"
	"xconfwebconfig/db"
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/shared/rfc"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

type FeatureControlRuleBase struct {
	FeatureDAO           db.CachedSimpleDao
	RuleProcessorFactory re.RuleProcessorFactory
}

func NewFeatureControlRuleBase() *FeatureControlRuleBase {
	return &FeatureControlRuleBase{
		FeatureDAO:           db.GetCachedSimpleDao(),
		RuleProcessorFactory: *re.NewRuleProcessorFactory(),
	}
}

func (f *FeatureControlRuleBase) Eval(context map[string]string, applicationType string, fields log.Fields) *rfc.FeatureControl {
	appliedFeatureRules := f.ProcessFeatureRules(context, applicationType)
	featureMap := map[string]*rfc.Feature{}
	if len(appliedFeatureRules) > 0 {
		for _, featureRule := range appliedFeatureRules {
			f.AddFeaturesToResult(featureMap, featureRule.FeatureIds)
		}
	}
	featureResponseList := make([]rfc.FeatureResponse, 0)
	for _, v := range featureMap {
		featureResponseList = append(featureResponseList, rfc.CreateFeatureResponseObject(*v))
	}
	featureControl := &rfc.FeatureControl{
		FeatureResponses: featureResponseList,
	}

	f.LogFeatureInfo(context, appliedFeatureRules, featureResponseList, fields)
	return featureControl
}

var rfcGetOneFeatureFunc = rfc.GetOneFeature

func (f *FeatureControlRuleBase) AddFeaturesToResult(featureMap map[string]*rfc.Feature, featureIds []string) {
	var feature *rfc.Feature
	for _, featureID := range featureIds {
		if featureID == "" {
			continue // no feature
		}
		feature = rfcGetOneFeatureFunc(featureID)
		if feature == nil {
			log.Debug(fmt.Sprintf("AddFeaturesToResult failed to find feature ID %v", featureID))
			continue // feature not found
		}
		if _, ok := featureMap[feature.Name]; ok {
			continue // feature already exists
		}
		clonedFeature, err := feature.Clone()
		if err != nil {
			log.Error(fmt.Sprintf("AddFeaturesToResult failed to clone %v: %v", feature, err))
			continue // cloning failed
		}
		ToRfcResponse(clonedFeature)
		featureMap[feature.Name] = clonedFeature
	}
}

func (f *FeatureControlRuleBase) ProcessFeatureRules(context map[string]string, applicationType string) []*rfc.FeatureRule {
	featureRules := rfc.GetSortedFeatureRules()
	var filteredfeatureRules []*rfc.FeatureRule
	for _, featureRule := range featureRules {
		// TODO: please add log.Fields to this method
		if applicationType == featureRule.ApplicationType && f.RuleProcessorFactory.RuleProcessor().Evaluate(featureRule.Rule, context, log.Fields{}) {
			filteredfeatureRules = append(filteredfeatureRules, featureRule)
		}
	}
	return filteredfeatureRules
}

func (f *FeatureControlRuleBase) CalculateHash(features []rfc.FeatureResponse) string {
	arrBytes := []byte{}
	arrBytes = append(arrBytes, []byte("[")...)
	sort.SliceStable(features, func(i, j int) bool {
		return features[i]["featureInstance"].(string) < features[j]["featureInstance"].(string)
	})
	for _, feature := range features {
		jsonBytes, _ := json.Marshal(feature)
		arrBytes = append(arrBytes, jsonBytes...)
	}
	arrBytes = append(arrBytes, []byte("]")...)
	return util.CalculateHash(string(arrBytes))
}

func (f *FeatureControlRuleBase) LogFeatureInfo(context map[string]string, appliedRules []*rfc.FeatureRule, features []rfc.FeatureResponse, fields log.Fields) {
	fields["context"] = context
	var ruleNames []string
	for _, rule := range appliedRules {
		ruleNames = append(ruleNames, rule.Name)
	}
	if len(ruleNames) > 0 {
		fields["appliedRules"] = ruleNames
	} else {
		fields["appliedRules"] = "NO MATCH"
	}

	var featureInstances []string
	for _, feature := range features {
		featureInstance, ok := feature["featureInstance"].(string)
		if ok {
			featureInstances = append(featureInstances, featureInstance)
		}
	}
	fields["features"] = featureInstances
	fields["configSetHash"] = f.CalculateHash(features)
	log.WithFields(fields).Info("FeatureControlRuleBase")
}

func (f *FeatureControlRuleBase) NormalizeContext(context map[string]string) map[string]string {
	if len(context[common.MODEL]) > 0 {
		context[common.MODEL] = strings.ToUpper(context[common.MODEL])
	}
	if len(context[common.ENV]) > 0 {
		context[common.ENV] = strings.ToUpper(context[common.ENV])
	}
	if len(context[common.PARTNER_ID]) > 0 {
		context[common.PARTNER_ID] = strings.ToUpper(context[common.PARTNER_ID])
	}
	if len(context[common.ESTB_MAC_ADDRESS]) > 0 {
		context[common.ESTB_MAC_ADDRESS] = util.NormalizeMacAddress(context[common.ESTB_MAC_ADDRESS])
	}
	if len(context[common.ECM_MAC_ADDRESS]) > 0 {
		context[common.ECM_MAC_ADDRESS] = util.NormalizeMacAddress(context[common.ECM_MAC_ADDRESS])
	}
	return context
}
