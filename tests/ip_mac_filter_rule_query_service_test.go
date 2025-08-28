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
	"testing"

	"github.com/rdkcentral/xconfwebconfig/dataapi/estbfirmware"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	corefw "github.com/rdkcentral/xconfwebconfig/shared/firmware"

	"gotest.tools/assert"
)

func TestConvertToIpRuleOrReturnNull(t *testing.T) {
	firmwareRule := GetFirmwareRule1()
	assert.Assert(t, firmwareRule.ID != "")

	firmwareConfig := GetFirmwareConfig1()
	assert.Assert(t, firmwareConfig.ID != "")

	svc := &estbfirmware.IpRuleService{}
	// store into DB
	err := corefw.CreateFirmwareRuleOneDB(firmwareRule)
	assert.NilError(t, err)
	err = coreef.CreateFirmwareConfigOneDB(firmwareConfig)
	assert.NilError(t, err)
	bean := svc.ConvertToIpRuleOrReturnNull(firmwareRule)
	assert.Assert(t, bean != nil)
}
