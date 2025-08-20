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
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/dataapi"
	"github.com/rdkcentral/xconfwebconfig/db"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func TestTelemetryTwoHandlerJmeter01(t *testing.T) {
	t.Skip("Debug with the real tagging service, no mocking")
	// setup env
	log.SetLevel(log.WarnLevel)

	router := server.GetRouter(true)
	dataapi.XconfSetup(server, router)
	// Walk(router)

	cc, ok := server.DatabaseClient.(*db.CassandraClient)
	assert.Assert(t, ok)
	fmt.Printf("x=%v\n", cc.Keyspace)

	// ==== case 1 build the query params ====
	queryParamString := "estbMacAddress=11:22:11:22:00:01"
	url := fmt.Sprintf("/loguploader/getTelemetryProfiles?%v", queryParamString)
	req, err := http.NewRequest("GET", url, nil)
	assert.NilError(t, err)
	res := dataapi.ExecuteRequest(req, router).Result()
	rbytes, err := ioutil.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	t.Logf("%v\n", string(rbytes))
	assert.Equal(t, res.StatusCode, http.StatusOK)
}
