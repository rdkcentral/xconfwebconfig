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
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"xconfwebconfig/db"
	"xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

// WhitelistProperty in WhitelistProperty.java
type WhitelistProperty struct {
	Key                string `json:"key,omitempty"`
	Value              string `json:"value,omitempty"`
	NamespacedListType string `json:"namespacedListType,omitempty"`
	TypeName           string `json:"typeName,omitempty"`
}

// NewWhitelistProperty to create a new WhitelistProperty
func NewWhitelistProperty() *WhitelistProperty {
	return &WhitelistProperty{}
}

func (w *WhitelistProperty) equals(o *WhitelistProperty) bool {
	if o == nil {
		return false
	}
	if w == o {
		return true
	}
	if w.Key != o.Key {
		return false
	}
	if w.Value != o.Value {
		return false
	}
	if w.NamespacedListType != o.NamespacedListType {
		return false
	}
	if w.TypeName != o.TypeName {
		return false
	}
	return true
}

// WhitelistProperty toString ...
func (w *WhitelistProperty) toString() string {
	return "WhitelistProperty{" +
		"Key='" + w.Key + "\\" +
		", Value=" + w.Value + "\\" +
		", NamespacedListType='" + w.NamespacedListType + "\\" +
		", TypeName=''" + w.TypeName + "\\" +
		"}"
}

// PercentRange in PercentRange.java
type PercentRange struct {
	StartRange float64
	EndRange   float64
}

// NewPercentRange to create a new PercentRange
func NewPercentRange() *PercentRange {
	return &PercentRange{}
}

// FeatureLegacy in FeatureLegacy.java
type FeatureLegacy struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	EffectiveImmediate bool              `json:"effectiveImmediate"`
	Enable             bool              `json:"enable"`
	ConfigData         map[string]string `json:"configData"`
}

// NewFeatureLegacy to create a new FeatureLegacy
func NewFeatureLegacy() *FeatureLegacy {
	return &FeatureLegacy{}
}

// Feature XconfFeature table
type Feature struct {
	Properties         map[string]interface{} `json:"properties,omitempty"`
	ListType           string                 `json:"listType,omitempty"`
	ListSize           int                    `json:"listSize,omitempty"`
	ID                 string                 `json:"id,omitempty"`
	Updated            int64                  `json:"updated,omitempty"`
	Name               string                 `json:"name"`
	FeatureName        string                 `json:"featureName,omitempty"`
	EffectiveImmediate bool                   `json:"effectiveImmediate"`
	Enable             bool                   `json:"enable"`
	Whitelisted        bool                   `json:"whitelisted,omitempty"`
	ConfigData         map[string]string      `json:"configData"`
	WhitelistProperty  *WhitelistProperty     `json:"whitelistProperty,omitempty"`
	ApplicationType    string                 `json:"applicationType,omitempty"`
}

func (obj *Feature) Clone() (*Feature, error) {
	cloneObj, err := util.Copy(obj)
	if err != nil {
		return nil, err
	}
	return cloneObj.(*Feature), nil
}

type FeatureResponse map[string]interface{}

func CreateFeatureResponseObject(feature Feature) FeatureResponse {
	featureResponse := FeatureResponse{}
	featureResponse["name"] = feature.Name
	featureResponse["featureInstance"] = feature.FeatureName
	featureResponse["effectiveImmediate"] = feature.EffectiveImmediate
	featureResponse["enable"] = feature.Enable
	featureResponse["configData"] = feature.ConfigData

	if feature.ListType != "" && feature.ListSize > 0 {
		featureResponse["listType"] = feature.ListType
		featureResponse["listSize"] = feature.ListSize
		for key, value := range feature.Properties {
			featureResponse[key] = value
		}
	}

	return featureResponse
}

func (f *FeatureResponse) MarshalJSON() ([]byte, error) {
	/**
	 * The order is important for STB team as they read json response via bash commands in upgrade script.
	 */
	// fields := []string{
	// 	"name",
	// 	"effectiveImmediate",
	// 	"enable",
	// 	"configData",
	// 	"listType",
	// 	"listSize",
	// 	"featureInstance",
	// }

	fields := []string{
		"name",
		"enable",
		"effectiveImmediate",
		"configData",
	}

	buf := bytes.Buffer{}
	buf.WriteByte('{')

	firstEntry := true
	for _, field := range fields {
		value := (*f)[field]
		if value != nil {
			if firstEntry {
				firstEntry = false
			} else {
				buf.WriteByte(',')
			}

			// write key
			buf.WriteString(fmt.Sprintf("\"%s\":", field))

			// marshal value
			val, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			buf.Write(val)
		}
	}

	for key, val := range *f {
		if util.Contains(fields, key) {
			continue // ignore key that has been processed already
		}

		// write key
		buf.WriteString(fmt.Sprintf(",\"%s\":", key))

		// marshal value
		val, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// NewFeature to create a new Feature
func NewFeature() *Feature {
	return &Feature{}
}

// NewFeatureInf to create a new Feature
func NewFeatureInf() interface{} {
	return &Feature{}
}

func (f *Feature) equals(o *Feature) bool {
	if o == nil {
		return false
	} else if f == o {
		return true
	} else if f.ID != o.ID {
		return false
	} else if f.Name != o.Name {
		return false
	} else if f.EffectiveImmediate != o.EffectiveImmediate {
		return false
	} else if f.Enable != o.Enable {
		return false
	} else if f.Whitelisted != o.Whitelisted {
		return false
	} else if f.ApplicationType != o.ApplicationType {
		return false
	} else if f.FeatureName != o.FeatureName {
		return false
	} else if !reflect.DeepEqual(f.ConfigData, o.ConfigData) {
		return false
	} else {
		return true
	}
}

// Feature ToString ...
func (f *Feature) ToString() string {
	return "Feature{" +
		"Name=" + f.Name + "\\" +
		", FeatureName=" + f.FeatureName + "\\" +
		", EffectiveImmediate=" + strconv.FormatBool(f.EffectiveImmediate) + "\\" +
		", Enable=" + strconv.FormatBool(f.Enable) + "\\" +
		", Whitelisted=" + strconv.FormatBool(f.Whitelisted) + "\\" +
		// todo: ", ConfigData=" + f.ConfigData + "\\" +
		// todo: ", WhitelistProperty=" + f.Whitelisted + "\\" +
		", ApplicationType=" + f.ApplicationType + "\\" +
		"}"
}

/// FeatureControl in FeatureControl.java
type FeatureControl struct {
	//set(FeatureResponse) should be defined as map[FeatureResponse]bool as golang set.
	//but FeatureResponse is not comparable because it has map inside
	FeatureResponses []FeatureResponse `json:"features"`
}

// NewFeatureControl to create a new FeatureControl
func NewFeatureControl() *FeatureControl {
	return &FeatureControl{}
}

func GetOneFeature(featureId string) *Feature {
	cftinst, err := db.GetCachedSimpleDao().GetOne(db.TABLE_XCONF_FEATURE, featureId)
	if err != nil {
		log.Warn(fmt.Sprintf("no feature found for featureId: %s", featureId))
		return nil
	}
	feature := cftinst.(*Feature)
	return feature
}

func GetFeatureList() []*Feature {
	cm := db.GetCacheManager()
	cacheKey := "FeatureList"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_XCONF_FEATURE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*Feature)
	}

	featureList, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_XCONF_FEATURE, 0)
	if err != nil {
		log.Warn(fmt.Sprintf("no feature found"))
		return nil
	}

	all := make([]*Feature, 0, len(featureList))

	for idx := range featureList {
		feature := featureList[idx].(*Feature)
		all = append(all, feature)
	}

	if len(all) > 0 {
		cm.ApplicationCacheSet(db.TABLE_XCONF_FEATURE, cacheKey, all)
	}
	return all
}

func GetFeatureRuleList() []*FeatureRule {
	cm := db.GetCacheManager()
	cacheKey := "FeatureRuleList"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_FEATURE_CONTROL_RULE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*FeatureRule)
	}

	featureRuleList, err := db.GetCachedSimpleDao().GetAllAsList(db.TABLE_FEATURE_CONTROL_RULE, 0)
	if err != nil {
		log.Warn(fmt.Sprintf("no featureRule found"))
		return nil
	}

	all := make([]*FeatureRule, 0, len(featureRuleList))

	for idx := range featureRuleList {
		featureRule := featureRuleList[idx].(*FeatureRule)
		all = append(all, featureRule)
	}

	if len(all) > 0 {
		cm.ApplicationCacheSet(db.TABLE_FEATURE_CONTROL_RULE, cacheKey, all)
	}
	return all
}

var GetFeatureListFunc = GetFeatureRuleList

// GetFeatureControl returns FeatureRule sorted by Priority
func GetSortedFeatureRules() []*FeatureRule {
	cm := db.GetCacheManager()
	cacheKey := "FeatureRuleListSorted"
	cacheInst := cm.ApplicationCacheGet(db.TABLE_FEATURE_CONTROL_RULE, cacheKey)
	if cacheInst != nil {
		return cacheInst.([]*FeatureRule)
	}

	all := GetFeatureListFunc()

	if len(all) <= 1 {
		return all
	}

	var sortedList []*FeatureRule
	sortedList = append(sortedList, all...)

	sort.SliceStable(sortedList, func(i, j int) bool {
		return sortedList[i].Priority < sortedList[j].Priority
	})
	cm.ApplicationCacheSet(db.TABLE_FEATURE_CONTROL_RULE, cacheKey, sortedList)

	return sortedList
}
