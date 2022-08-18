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
package db

import "github.com/gocql/gocql"

// OperationType enum
type OperationType string

const (
	CREATE_OPERATION   OperationType = "CREATE"
	UPDATE_OPERATION   OperationType = "UPDATE"
	DELETE_OPERATION   OperationType = "DELETE"
	TRUNCATE_OPERATION OperationType = "TRUNCATE_CF"
)

// ChangedData XconfChangedKeys4 table
type ChangedData struct {
	ColumnName     gocql.UUID    `json:"columnName"`
	CfName         string        `json:"cfName"`
	ChangedKey     string        `json:"changedKey"`
	Operation      OperationType `json:"operation"`
	DaoId          int32         `json:"DAOid"`
	ValidCacheSize int32         `json:"validCacheSize"`
	UserName       string        `json:"userName"`
}

func NewChangedDataInf() interface{} {
	return &ChangedData{}
}
