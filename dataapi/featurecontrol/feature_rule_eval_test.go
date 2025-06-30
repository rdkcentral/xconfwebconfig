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
	"testing"

	"xconfwebconfig/shared"
	"xconfwebconfig/shared/rfc"

	"gotest.tools/assert"
)

func TestSortCaseInsensitive(t *testing.T) {
	var list []string
	sortedList := SortCaseInsensitive(list)
	assert.Equal(t, sortedList == nil, true)
	assert.Equal(t, list == nil, true)

	list = []string{}
	sortedList = SortCaseInsensitive(list)
	assert.Equal(t, len(sortedList), 0)
	assert.Equal(t, len(list), 0)

	list = []string{"abc"}
	sortedList = SortCaseInsensitive(list)
	assert.Equal(t, len(sortedList), 1)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, sortedList[0], "abc")
	assert.Equal(t, list[0], "abc")

	list = []string{"ABC", "abc", "def", "123", "DEF"}
	sortedList = SortCaseInsensitive(list)
	assert.Equal(t, len(sortedList), 5)
	assert.Equal(t, len(list), 5)
	assert.Equal(t, sortedList[0], "123")
	assert.Equal(t, sortedList[1], "ABC")
	assert.Equal(t, sortedList[2], "abc")
	assert.Equal(t, sortedList[3], "DEF")
	assert.Equal(t, sortedList[4], "def")
	assert.Equal(t, list[0], "ABC")
	assert.Equal(t, list[1], "abc")
	assert.Equal(t, list[2], "def")
	assert.Equal(t, list[3], "123")
	assert.Equal(t, list[4], "DEF")
}

func TestGetSortedFeatureRules(t *testing.T) {
	rfc.GetFeatureListFunc = func() []*rfc.FeatureRule {
		all := []*rfc.FeatureRule{}
		featureIds := []string{"1", "2", "3"}
		featureRule1 := rfc.FeatureRule{
			Id:              "id1",
			Name:            "name1",
			Rule:            nil,
			Priority:        1,
			FeatureIds:      featureIds,
			ApplicationType: "applicationType1",
		}
		featureRule2 := rfc.FeatureRule{
			Id:              "id2",
			Name:            "name2",
			Rule:            nil,
			Priority:        2,
			FeatureIds:      featureIds,
			ApplicationType: "applicationType2",
		}
		featureRule3 := rfc.FeatureRule{
			Id:              "id3",
			Name:            "name3",
			Rule:            nil,
			Priority:        3,
			FeatureIds:      featureIds,
			ApplicationType: "applicationType3",
		}
		all = append(all, &featureRule2)
		all = append(all, &featureRule3)
		all = append(all, &featureRule1)
		assert.Equal(t, len(all), 3)
		assert.Equal(t, all[0].Id, "id2")
		assert.Equal(t, all[0].Priority, 2)
		assert.Equal(t, all[1].Id, "id3")
		assert.Equal(t, all[1].Priority, 3)
		assert.Equal(t, all[2].Id, "id1")
		assert.Equal(t, all[2].Priority, 1)
		return all
	}
	// rules []*rfc.FeatureRule
	sortedRules := rfc.GetSortedFeatureRules()
	assert.Equal(t, len(sortedRules), 3)
	assert.Equal(t, sortedRules[0].Id, "id1")
	assert.Equal(t, sortedRules[0].Priority, 1)
	assert.Equal(t, sortedRules[1].Id, "id2")
	assert.Equal(t, sortedRules[1].Priority, 2)
	assert.Equal(t, sortedRules[2].Id, "id3")
	assert.Equal(t, sortedRules[2].Priority, 3)

}

func TestAddFeaturesToResult(t *testing.T) {
	GetGenericNamedListOneByTypeFunc = func(string, string) (*shared.GenericNamespacedList, error) {
		genericNamespacedList := shared.GenericNamespacedList{
			ID:       "id",
			Updated:  12345,
			Data:     []string{"1", "2", "3"},
			TypeName: "typeName",
		}
		return &genericNamespacedList, nil
	}

	rfcGetOneFeatureFunc = func(featureId string) *rfc.Feature {
		whitelistProperty1 := rfc.WhitelistProperty{
			Key:                "key1",
			Value:              "value1",
			NamespacedListType: "namespacedListType1",
			TypeName:           "typeName1",
		}
		whitelistProperty2 := rfc.WhitelistProperty{
			Key:                "key2",
			Value:              "value2",
			NamespacedListType: "namespacedListType2",
			TypeName:           "typeName2",
		}
		whitelistProperty3 := rfc.WhitelistProperty{
			Key:                "key3",
			Value:              "value3",
			NamespacedListType: "namespacedListType3",
			TypeName:           "typeName3",
		}
		feature1 := rfc.Feature{
			ID:                 "id1",
			Updated:            12345,
			Name:               "name1",
			FeatureName:        "featureInstance1",
			EffectiveImmediate: true,
			Enable:             true,
			Whitelisted:        true,
			ConfigData:         make(map[string]string),
			WhitelistProperty:  &whitelistProperty1,
			ApplicationType:    "applicationType1",
		}
		feature2 := rfc.Feature{
			ID:                 "id2",
			Updated:            12345,
			Name:               "name2",
			FeatureName:        "featureInstance2",
			EffectiveImmediate: true,
			Enable:             true,
			Whitelisted:        true,
			ConfigData:         make(map[string]string),
			WhitelistProperty:  &whitelistProperty2,
			ApplicationType:    "applicationType2",
		}
		feature3 := rfc.Feature{
			ID:                 "id3",
			Updated:            12345,
			Name:               "name3",
			FeatureName:        "featureInstance3",
			EffectiveImmediate: true,
			Enable:             true,
			Whitelisted:        true,
			ConfigData:         make(map[string]string),
			WhitelistProperty:  &whitelistProperty3,
			ApplicationType:    "applicationType3",
		}
		if featureId == "id2" {
			return &feature2
		}
		if featureId == "id3" {
			return &feature3
		}
		return &feature1
	}
	var featureControlRuleBase *FeatureControlRuleBase
	featureControlRuleBase = &FeatureControlRuleBase{}
	featureMap := make(map[string]*rfc.Feature)
	var featureIDs = []string{"id1", "id2", "id3", "id4"}
	featureControlRuleBase.AddFeaturesToResult(featureMap, featureIDs)
	assert.Equal(t, len(featureMap), 3)
	if _, ok := featureMap["name3"]; ok {
		feature := featureMap["name3"]
		assert.Equal(t, feature.ID, "id3")
		assert.Equal(t, feature.Name, "name3")
		assert.Equal(t, feature.FeatureName, "featureInstance3")
		assert.Equal(t, feature.EffectiveImmediate, true)
		assert.Equal(t, feature.Enable, true)
		assert.Equal(t, feature.Whitelisted, true)
		assert.Equal(t, len(feature.ConfigData), 0)
		assert.Equal(t, feature.WhitelistProperty.Key, "key3")
		assert.Equal(t, feature.WhitelistProperty.Value, "value3")
		assert.Equal(t, feature.WhitelistProperty.NamespacedListType, "namespacedListType3")
		assert.Equal(t, feature.WhitelistProperty.TypeName, "typeName3")
		assert.Equal(t, feature.ApplicationType, "applicationType3")
	} else {
		assert.Equal(t, "fail", "name3 not in featureMap")
	}
}

func TestNormalizeContent(t *testing.T) {
	var featureControlRuleBase *FeatureControlRuleBase
	featureControlRuleBase = &FeatureControlRuleBase{}

	voidCtxt := map[string]string{}
	voidCtxt = featureControlRuleBase.NormalizeContext(voidCtxt)
	assert.Equal(t, len(voidCtxt), 0)

	rawCtxt := map[string]string{
		"model":          "model",
		"env":            "env",
		"partnerId":      "partnerId",
		"estbMacAddress": "147ddaa5837b",
		"ecmMacAddress":  "147ddaa5837b",
	}
	rawCtxt = featureControlRuleBase.NormalizeContext(rawCtxt)
	assert.Equal(t, rawCtxt["model"], "MODEL")
	assert.Equal(t, rawCtxt["env"], "ENV")
	assert.Equal(t, rawCtxt["partnerId"], "PARTNERID")
	assert.Equal(t, rawCtxt["estbMacAddress"], "14:7D:DA:A5:83:7B")
	assert.Equal(t, rawCtxt["ecmMacAddress"], "14:7D:DA:A5:83:7B")

	bakedCtxt := map[string]string{
		"model":          "MODEL",
		"env":            "ENV",
		"partnerId":      "PARTNERID",
		"estbMacAddress": "14:7d:da:a5:83:7b",
		"ecmMacAddress":  "14:7d:da:a5:83:7b",
	}
	bakedCtxt = featureControlRuleBase.NormalizeContext(bakedCtxt)
	assert.Equal(t, bakedCtxt["model"], "MODEL")
	assert.Equal(t, bakedCtxt["env"], "ENV")
	assert.Equal(t, bakedCtxt["partnerId"], "PARTNERID")
	assert.Equal(t, bakedCtxt["estbMacAddress"], "14:7D:DA:A5:83:7B")
	assert.Equal(t, bakedCtxt["ecmMacAddress"], "14:7D:DA:A5:83:7B")

	emptyCtxt := map[string]string{
		"model":          "",
		"env":            "",
		"partnerId":      "",
		"estbMacAddress": "",
		"ecmMacAddress":  "",
	}
	emptyCtxt = featureControlRuleBase.NormalizeContext(emptyCtxt)
	assert.Equal(t, emptyCtxt["model"], "")
	assert.Equal(t, emptyCtxt["env"], "")
	assert.Equal(t, emptyCtxt["partnerId"], "")
	assert.Equal(t, emptyCtxt["estbMacAddress"], "")
	assert.Equal(t, emptyCtxt["ecmMacAddress"], "")

	longCtxt := map[string]string{
		"model":          "abcdefghijklmnopqrstuvwxyz",
		"env":            "abcdefghijklmnopqrstuvwxyz",
		"partnerId":      "abcdefghijklmnopqrstuvwxyz",
		"estbMacAddress": "abcdefghijklmnopqrstuvwxyz",
		"ecmMacAddress":  "abcdefghijklmnopqrstuvwxyz",
	}
	longCtxt = featureControlRuleBase.NormalizeContext(longCtxt)
	assert.Equal(t, longCtxt["model"], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	assert.Equal(t, longCtxt["env"], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	assert.Equal(t, longCtxt["partnerId"], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	assert.Equal(t, longCtxt["estbMacAddress"], "AB:CD:EF:GH:IJ:KL")
	assert.Equal(t, longCtxt["ecmMacAddress"], "AB:CD:EF:GH:IJ:KL")

}

func TestCalculateConfigSetHash(t *testing.T) {
	f := NewFeatureControlRuleBase()
	var features []rfc.FeatureResponse
	// both empty and nil featureResponse lists are the same because both have no data
	configSetHashNil := f.CalculateHash(features)
	features = []rfc.FeatureResponse{}
	configSetHashEmpty := f.CalculateHash(features)
	assert.Equal(t, configSetHashEmpty, configSetHashNil)

	configData := map[string]string{
		"tr181.Device.DeviceInfo.X_RDKCENTRAL-COM_RFC.Feature.OVS.Enable": "false",
	}
	featureResponse := rfc.FeatureResponse{
		"name":               "OVS",
		"enable":             false,
		"effectiveImmediate": true,
		"configData":         configData,
		"featureInstance":    "OVS:E",
	}
	var featureResponseList []rfc.FeatureResponse
	featureResponseList = append(featureResponseList, featureResponse)
	configSetHash1 := f.CalculateHash(featureResponseList)
	// change featureInstance only
	featureResponse["featureInstance"] = "OVS:D"
	configSetHash2 := f.CalculateHash(featureResponseList)
	// check that change in feature instance changes hash
	assert.Equal(t, configSetHash1 != configSetHash2, true)

	// add more features
	configData2 := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "value4",
		"key5": "value5",
	}
	featureResponse2 := rfc.FeatureResponse{
		"name":               "name",
		"enable":             true,
		"effectiveImmediate": true,
		"configData":         configData2,
		"featureInstance":    "featureInstance",
	}

	configData3 := map[string]string{
		"foo":   "bar",
		"hello": "world",
	}
	featureResponse3 := rfc.FeatureResponse{
		"name":               "somevalue",
		"enable":             true,
		"effectiveImmediate": true,
		"configData":         configData3,
		"featureInstance":    "someothervalue",
	}
	// three lists with different order but same features
	featureResponseList1 := []rfc.FeatureResponse{featureResponse, featureResponse2, featureResponse3}
	featureResponseList2 := []rfc.FeatureResponse{featureResponse, featureResponse3, featureResponse2}
	featureResponseList3 := []rfc.FeatureResponse{featureResponse2, featureResponse, featureResponse3}

	// ensure that running same order many times gets same hash
	configSetHash1 = f.CalculateHash(featureResponseList1)
	for i := 0; i < 20; i++ {
		assert.Equal(t, configSetHash1, f.CalculateHash(featureResponseList1))
	}
	configSetHash2 = f.CalculateHash(featureResponseList2)
	for i := 0; i < 20; i++ {
		assert.Equal(t, configSetHash2, f.CalculateHash(featureResponseList2))
	}
	configSetHash3 := f.CalculateHash(featureResponseList3)
	for i := 0; i < 20; i++ {
		assert.Equal(t, configSetHash3, f.CalculateHash(featureResponseList3))
	}
	// ensure running with different order still gets same hash
	assert.Equal(t, configSetHash1, configSetHash2)
	assert.Equal(t, configSetHash1, configSetHash3)
	assert.Equal(t, configSetHash2, configSetHash3)

}
