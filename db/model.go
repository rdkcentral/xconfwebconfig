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

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gocql/gocql"
	"github.com/rdkcentral/xconfwebconfig/util"
	log "github.com/sirupsen/logrus"
)

// OperationType enum
type OperationType string

const (
	CREATE_OPERATION   OperationType = "CREATE"
	UPDATE_OPERATION   OperationType = "UPDATE"
	DELETE_OPERATION   OperationType = "DELETE"
	TRUNCATE_OPERATION OperationType = "TRUNCATE_CF"
)

// Interface to be implemented by all objects stored in DB that has updated timestamp field
type Updatable interface {
	GetUpdated() int64
	SetUpdated(int64)
}

// ChangedData change_events table
type ChangedData struct {
	ColumnName     gocql.UUID    `json:"columnName"`
	CfName         string        `json:"cfName"`
	ChangedKey     string        `json:"changedKey"`
	Operation      OperationType `json:"operation"`
	ValidCacheSize int32         `json:"validCacheSize"`
	UserName       string        `json:"userName"`
	ServerOriginId string        `json:"serverOriginId"`
	TenantId       string        `json:"tenantId"`
}

func NewChangedDataInf() any {
	return &ChangedData{}
}

// Tenant tenants table
type Tenant struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Updated int64  `json:"updated"`
}

func (obj *Tenant) GetUpdated() int64 {
	return obj.Updated
}

func (obj *Tenant) SetUpdated(ts int64) {
	obj.Updated = ts
}

func (obj *Tenant) Validate() error {
	if len(strings.TrimSpace(obj.ID)) > 0 {
		match, _ := regexp.MatchString("^[-a-zA-Z0-9_.' ]+$", obj.ID)
		if match {
			return nil
		}
	}

	return errors.New("Id is invalid")
}

func NewTenant(id string, name string) *Tenant {
	if id != "" {
		id = strings.ToUpper(strings.TrimSpace(id))
	}
	if util.IsBlank(name) {
		name = id
	}

	return &Tenant{
		ID:   id,
		Name: strings.TrimSpace(name),
	}
}

func CreateTenant(id string, name string) (*Tenant, error) {
	// Attempt to retrieve the tenant from the database, if it does not exist, create a new tenant
	tenant, err := GetDatabaseClient().GetTenant(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tenant %s: %v", id, err)
	}
	if tenant == nil {
		tenant = NewTenant(id, name)
		tenant.Updated = util.GetTimestamp()
		if err := tenant.Validate(); err != nil {
			return nil, err
		}
		log.WithFields(log.Fields{"tenantId": id}).Infof("Creating new tenant...")
		if err := GetDatabaseClient().SetTenant(tenant); err != nil {
			return nil, fmt.Errorf("failed to create tenant %s: %v", tenant.ID, err)
		}
	}

	if strings.EqualFold(tenant.ID, id) == false {
		return nil, fmt.Errorf("tenant ID mismatch: expected %s, got %s", id, tenant.ID)
	}

	return tenant, nil
}

// EnsureDefaultTenantExists ensures that the default tenant exists, creating it if necessary
func EnsureDefaultTenantExists() (*Tenant, error) {
	tenantId := GetDefaultTenantId()
	return CreateTenant(tenantId, tenantId)
}
