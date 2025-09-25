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
package dataapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/gorilla/mux"
)

func GetInfoRefreshAllHandler(w http.ResponseWriter, r *http.Request) {
	failedToRefreshTables := db.GetCacheManager().RefreshAll()
	if len(failedToRefreshTables) == 0 {
		stats := db.GetCacheManager().GetStatistics()
		response, _ := util.JSONMarshal(stats.CacheMap)
		xhttp.WriteXconfResponse(w, http.StatusOK, response)
	} else {
		xhttp.WriteXconfResponse(w, 404, []byte(fmt.Sprintf("\"Couldn't refresh caches for tables: %s\"", strings.Join(failedToRefreshTables, ", "))))
	}
}

func GetInfoRefreshHandler(w http.ResponseWriter, r *http.Request) {
	tableName := mux.Vars(r)[common.TABLE_NAME]
	err := db.GetCacheManager().Refresh(tableName)
	if err == nil {
		if stats, err := db.GetCacheManager().GetCacheStats(tableName); err == nil {
			response, _ := util.JSONMarshal(stats)
			xhttp.WriteXconfResponse(w, http.StatusOK, response)
		} else {
			xhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		}
	} else {
		xhttp.WriteXconfResponse(w, http.StatusInternalServerError, []byte(err.Error()))
	}
}

func GetInfoStatistics(w http.ResponseWriter, r *http.Request) {
	stats := *db.GetCacheManager().GetStatistics()
	response, _ := util.JSONMarshal(stats)
	xhttp.WriteXconfResponse(w, 200, response)
}
