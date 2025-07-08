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
package rfc

import (
	re "xconfwebconfig/rulesengine"
	"xconfwebconfig/util"
)

// FeatureRule FeatureControlRule2 table
type FeatureRule struct {
	Id              string   `json:"id"`
	Name            string   `json:"name"`
	Rule            *re.Rule `json:"rule"`
	Priority        int      `json:"priority"`
	FeatureIds      []string `json:"featureIds"`
	ApplicationType string   `json:"applicationType"`
}

func (obj *FeatureRule) SetApplicationType(appType string) {
	obj.ApplicationType = appType
}

func (obj *FeatureRule) GetApplicationType() string {
	return obj.ApplicationType
}

func (obj *FeatureRule) Clone() (*FeatureRule, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*FeatureRule), nil
}

func NewFeatureRuleInf() interface{} {
	return &FeatureRule{}
}

// GetId XRule interface
func (r *FeatureRule) GetId() string {
	return r.Id
}

func (r *FeatureRule) GetID() string {
	return r.Id
}

func (r *FeatureRule) GetPriority() int {
	return r.Priority
}

func (r *FeatureRule) SetPriority(priority int) {
	r.Priority = priority
}

// GetRule XRule interface
func (r *FeatureRule) GetRule() *re.Rule {
	return r.Rule
}

// GetName XRule interface
func (r *FeatureRule) GetName() string {
	return r.Name
}

// GetTemplateId XRule interface
func (r *FeatureRule) GetTemplateId() string {
	return ""
}

// GetRuleType XRule interface
func (r *FeatureRule) GetRuleType() string {
	return "FeatureRule"
}
