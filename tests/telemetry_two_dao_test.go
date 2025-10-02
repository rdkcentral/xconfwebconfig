/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
package tests

import (
	"encoding/json"
	"fmt"
	"testing"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	re "github.com/rdkcentral/xconfwebconfig/rulesengine"
	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"gotest.tools/assert"
)

func TestTelemetryTwoDao(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	// ==== setup random variable ====
	// namedlistKey := fmt.Sprintf("red%v", uuid.New().String()[:4])
	ruleUuid := uuid.New().String()
	ruleName := fmt.Sprintf("orange%v", uuid.New().String()[:4])
	profileName := fmt.Sprintf("yellow%v", uuid.New().String()[:4])
	profileUuid := uuid.New().String()

	// write a t2rule
	sr1 := fmt.Sprintf(MockTelemetryTwoRuleTemplate1, ruleUuid, ruleName, profileUuid)
	var srcT2Rule logupload.TelemetryTwoRule
	err := json.Unmarshal([]byte(sr1), &srcT2Rule)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, srcT2Rule.ID, &srcT2Rule)
	assert.NilError(t, err)
	// get a t2profile
	itf, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_RULES, ruleUuid)
	tgtT2Rule, ok := itf.(*logupload.TelemetryTwoRule)
	assert.Assert(t, ok)
	assert.Assert(t, srcT2Rule.Equals(tgtT2Rule))

	// write a t2profile
	sp1 := fmt.Sprintf(MockTelemetryTwoProfileTemplate1, profileName, profileUuid)
	var srcT2Profile logupload.TelemetryTwoProfile
	err = json.Unmarshal([]byte(sp1), &srcT2Profile)
	assert.NilError(t, err)
	err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid, &srcT2Profile)
	assert.NilError(t, err)
	// get a t2profile
	itf, err = ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid)
	tgtT2Profile, ok := itf.(*logupload.TelemetryTwoProfile)
	assert.Assert(t, ok)
	assert.DeepEqual(t, &srcT2Profile, tgtT2Profile)
}

func TestTelemetryTwoDaoSampleData(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	// build sample t2rules
	t2Rules := []logupload.TelemetryTwoRule{}
	err := json.Unmarshal([]byte(SampleTelemetryTwoRulesString), &t2Rules)
	assert.NilError(t, err)
	mykeys := []string{}
	sourceData := util.Dict{}
	for _, v := range t2Rules {
		t2Rule := v
		sourceData[t2Rule.ID] = &t2Rule
		mykeys = append(mykeys, t2Rule.ID)
		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_RULES, t2Rule.ID, &t2Rule)
		assert.NilError(t, err)
		itf, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_RULES, t2Rule.ID)
		assert.NilError(t, err)
		fetchedT2Rule, ok := itf.(*logupload.TelemetryTwoRule)
		assert.Assert(t, ok)
		assert.Assert(t, t2Rule.Equals(fetchedT2Rule))
	}

	fetchedData := util.Dict{}
	for _, x := range mykeys {
		itf, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_RULES, x)
		assert.NilError(t, err)
		fetchedT2Rule, ok := itf.(*logupload.TelemetryTwoRule)
		assert.Assert(t, ok)
		fetchedData[fetchedT2Rule.ID] = itf
	}
	assert.DeepEqual(t, sourceData, fetchedData, cmp.AllowUnexported(re.Rule{}))

	// build sample t2profiles
	for profileUuid, profileName := range SampleProfileIdNameMap {
		// write a t2profile
		sp1 := fmt.Sprintf(MockTelemetryTwoProfileTemplate1, profileName, profileUuid)
		var sourceT2Profile logupload.TelemetryTwoProfile
		err = json.Unmarshal([]byte(sp1), &sourceT2Profile)
		assert.NilError(t, err)
		err = ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid, &sourceT2Profile)
		assert.NilError(t, err)
		// get a t2profile
		itf, err := ds.GetCachedSimpleDao().GetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, profileUuid)
		assert.NilError(t, err)
		fetchedT2Profile, ok := itf.(*logupload.TelemetryTwoProfile)
		assert.Assert(t, ok)
		assert.DeepEqual(t, &sourceT2Profile, fetchedT2Profile)
	}
}

func TestGenericNamedListDaoForMacs(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	namedListKey := fmt.Sprintf("red%v", uuid.New().String()[:4])
	macs := []string{
		"11:11:22:22:33:02",
		"11:11:22:22:33:03",
		"11:11:22:22:33:05",
		"11:11:22:22:33:07",
	}
	sourceNamedlist := shared.NewGenericNamespacedList(namedListKey, shared.MacList, macs)
	bbytes, err := json.Marshal(sourceNamedlist)
	assert.NilError(t, err)
	err = ds.GetCompressingDataDao().SetOne(shared.TableGenericNSList, sourceNamedlist.ID, bbytes)
	assert.NilError(t, err)
	itf, err := ds.GetCompressingDataDao().GetOne(shared.TableGenericNSList, sourceNamedlist.ID)
	assert.NilError(t, err)
	fetchedNamedlist, ok := itf.(*shared.GenericNamespacedList)
	assert.Assert(t, ok)
	assert.DeepEqual(t, fetchedNamedlist.Data, macs)
}

func TestGenericNamedlistDaoForIpAddresses(t *testing.T) {
	if !ds.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	namedListKey := fmt.Sprintf("scarlet%v", uuid.New().String()[:4])
	ips := []string{
		"1.2.3.4",
		"20.30.40.50/24",
		"33.44.55.66/20",
	}
	sourceNamedlist := shared.NewGenericNamespacedList(namedListKey, shared.IpList, ips)
	bbytes, err := json.Marshal(sourceNamedlist)
	assert.NilError(t, err)
	err = ds.GetCompressingDataDao().SetOne(shared.TableGenericNSList, sourceNamedlist.ID, bbytes)
	assert.NilError(t, err)
	itf, err := ds.GetCompressingDataDao().GetOne(shared.TableGenericNSList, sourceNamedlist.ID)
	assert.NilError(t, err)
	fetchedNamedlist, ok := itf.(*shared.GenericNamespacedList)
	assert.Assert(t, ok)
	assert.DeepEqual(t, fetchedNamedlist.Data, ips)
}
