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
package estbfirmware

import (
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"

	"gotest.tools/assert"
)

func TestContextConvertedGet(t *testing.T) {
	//t.Skip()
	// setup e
	contextMap := map[string]string{}
	contextMap["eStbMac"] = "00:0a:95:9d:68:16"
	contextMap["eCMMac"] = "00:0a:95:9d:68:17"
	contextMap["partnerId"] = "comcast"
	contextMap["ipAddress"] = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	contextMap["bypassFilters"] = "locationfilter1,bypassFilter2,PercentFilter3"
	contextMap["time"] = "04/15/2021 13:01:28"
	contextMap["applicationType"] = "stb"
	contextMap["timeZone"] = "US/Eastern"
	contextMap["timeZoneOffset"] = "08:00" // offset to UTC
	contextMap[common.FORCE_FILTERS] = "someFilter1,forceFilter2"
	contextMap[common.CAPABILITIES] = "RCDL, RebootDecoupled, SupportsFullHttpUrl"

	convertCtx := GetContextConverted(contextMap)

	assert.Assert(t, convertCtx != nil)
	assert.Assert(t, convertCtx.GetBypassFiltersConverted() != nil)
	assert.Equal(t, len(convertCtx.GetBypassFiltersConverted()), 3)
	assert.Assert(t, convertCtx.GetForceFiltersConverted() != nil)
	assert.Equal(t, len(convertCtx.GetForceFiltersConverted()), 2)
	assert.Equal(t, convertCtx.GetPartnerId(), "comcast")
	assert.Equal(t, convertCtx.GetEcmMacConverted(), "00:0a:95:9d:68:17")
	assert.Equal(t, convertCtx.GetEstbMacConverted(), "00:0a:95:9d:68:16")
	assert.Equal(t, convertCtx.GetIpAddressConverted(), "2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	assert.Assert(t, convertCtx.GetTimeConverted() != nil)
	assert.Equal(t, convertCtx.GetTimeConverted().String()[0:7], "2021-04")
	assert.Assert(t, convertCtx.GetCapabilitiesConverted() != nil)
	assert.Equal(t, len(convertCtx.GetCapabilitiesConverted()), 3)
	assert.Equal(t, convertCtx.IsRcdl(), true)
	assert.Equal(t, convertCtx.IsRebootDecoupled(), true)
	assert.Equal(t, convertCtx.GetRawTimeZoneConverted(), "UTC-8")
	assert.Assert(t, convertCtx.GetTimeZoneConverted() != time.UTC)

	caps := convertCtx.CreateCapabilitiesList()
	assert.Equal(t, len(caps), 3)
}

func TestGetTime(t *testing.T) {
	contextMap := map[string]string{}
	contextMap["time"] = "04/15/2021 00:01:43"
	convertCtx := GetContextConverted(contextMap)
	tm := convertCtx.GetTime()
	tn := time.Now()
	assert.Assert(t, *tm != tn)

	_, err := time.Parse(DATE_TIME_SEC_FORMATTER, "04/15/2021 00:01:23")
	assert.NilError(t, err)

	contextMap["time"] = "05/05/2021 19:46:32"
	convertCtx = GetContextConverted(contextMap)
	tm = convertCtx.GetTime()
	tn = time.Now()
	assert.Assert(t, *tm != tn)

	_, err = time.Parse(DATE_TIME_SEC_FORMATTER, "05/05/2021 19:46:32")
	assert.NilError(t, err)
}

func TestOffsetToTimeZone(t *testing.T) {
	tmoffset := offsetToTimeZone("11:00")
	assert.Assert(t, tmoffset != time.UTC)
}
