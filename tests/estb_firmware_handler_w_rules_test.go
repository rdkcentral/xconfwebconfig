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
	"io"
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/dataapi"
	ds "github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"

	"gotest.tools/assert"
)

func TestGetEstbFirmwareSwuHandler(t *testing.T) {
	// t.Skip()
	// setup env
	router := server.GetRouter(true)

	// set up codebig, xbo, tagging mock servers for ok response
	codebigMockServer := dataapi.SetupSatServiceMockServerOkResponse(t, *server)
	defer codebigMockServer.Close()

	AccountServiceMockServer := dataapi.SetupAccountServiceMockServerOkResponse(t, *server, fmt.Sprintf(URL_ACCOUNT_SERVICE_DEVICE_ESTB, mac1))
	defer AccountServiceMockServer.Close()

	taggingMockServer := dataapi.SetupTaggingMockServerOkResponse(t, *server, fmt.Sprintf(URL_TAGS_MAC_ADDRESS, mac1))
	defer taggingMockServer.Close()

	// setup test data
	server.SetXconfData(shared.TableFirmwareConfig, FirmwareConfigId1, firmwareConfig1Bytes, 3600)
	server.SetXconfData(shared.TableFirmwareConfig, FirmwareConfigId2, firmwareConfig2Bytes, 3600)
	server.SetXconfData(shared.TableFirmwareConfig, FirmwareConfigId3, firmwareConfig3Bytes, 3600)

	server.SetXconfData(shared.TableFirmwareRule, firmwareRuleId1, firmwareRule1Bytes, 3600)
	server.SetXconfData(shared.TableFirmwareRule, firmwareRuleId2, firmwareRule2Bytes, 3600)
	server.SetXconfData(shared.TableFirmwareRule, firmwareRuleId3, firmwareRule3Bytes, 3600)

	macs := []string{mac3, "AA:AA:AA:BB:BB:BB", "AA:AA:AA:BB:BB:CC"}
	newList := shared.NewGenericNamespacedList(namespaceListKey, shared.MacList, macs)
	compDao := ds.GetCompressingDataDao()
	bbytes, err := json.Marshal(newList)
	assert.NilError(t, err)
	err = compDao.SetOne(shared.TableGenericNSList, namespaceListKey, bbytes)
	assert.NilError(t, err)

	// no eStbMac and version is greater than or equal to, 400 error
	url := fmt.Sprintf("/xconf/swu/stb?eStbMac=%v", mac1)
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	// assert.Equal(t, res.StatusCode, http.StatusBadRequest)
	rbytes, err := io.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	t.Logf("%v\n", string(rbytes))

	// ok := false
	// assert.Assert(t, ok)
}
