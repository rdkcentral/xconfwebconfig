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
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/util"
	"gotest.tools/assert"
)

func newTestTenant() *db.Tenant {
	return &db.Tenant{
		ID:      "test-tenant-" + uuid.New().String(),
		Name:    "Test Tenant",
		Updated: util.GetTimestamp(),
	}
}

// TestGetTenant_NotFound verifies that GetTenant returns nil when the tenant does not exist.
func TestGetTenant_NotFound(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	dbClient := db.GetDatabaseClient()

	result, err := dbClient.GetTenant("nonexistent-tenant-" + uuid.New().String())
	assert.NilError(t, err)
	assert.Assert(t, result == nil, "expected nil for nonexistent tenant")
}

// TestGetTenant_EmptyId verifies that GetTenant returns an error when tenantId is empty.
func TestGetTenant_EmptyId(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	dbClient := db.GetDatabaseClient()

	result, err := dbClient.GetTenant("")
	assert.Assert(t, err != nil, "expected error for empty tenantId")
	assert.Assert(t, result == nil)
}

// TestSetAndGetTenant verifies that a tenant can be stored and retrieved.
func TestSetAndGetTenant(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	dbClient := db.GetDatabaseClient()
	tenant := newTestTenant()

	err := dbClient.SetTenant(tenant)
	assert.NilError(t, err)
	defer dbClient.DeleteTenant(tenant.ID)

	result, err := dbClient.GetTenant(tenant.ID)
	assert.NilError(t, err)
	assert.Assert(t, result != nil, "expected tenant to be found after insert")
	assert.Equal(t, result.ID, tenant.ID)
	assert.Equal(t, result.Name, tenant.Name)
	assert.Equal(t, result.Updated, tenant.Updated)
}

// TestDeleteTenant verifies that a tenant is no longer retrievable after deletion.
func TestDeleteTenant(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	dbClient := db.GetDatabaseClient()
	tenant := newTestTenant()

	err := dbClient.SetTenant(tenant)
	assert.NilError(t, err)

	err = dbClient.DeleteTenant(tenant.ID)
	assert.NilError(t, err)

	result, err := dbClient.GetTenant(tenant.ID)
	assert.NilError(t, err)
	assert.Assert(t, result == nil, "expected nil after deletion")
}

// TestGetAllTenants verifies that all inserted tenants are returned.
func TestGetAllTenants(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	dbClient := db.GetDatabaseClient()
	truncateTable(db.TABLE_TENANTS)

	tenant1 := newTestTenant()
	tenant2 := newTestTenant()

	assert.NilError(t, dbClient.SetTenant(tenant1))
	assert.NilError(t, dbClient.SetTenant(tenant2))
	defer func() {
		dbClient.DeleteTenant(tenant1.ID)
		dbClient.DeleteTenant(tenant2.ID)
	}()

	tenants := dbClient.GetAllTenants()
	assert.Assert(t, len(tenants) >= 2, "expected at least 2 tenants")

	ids := make(map[string]bool, len(tenants))
	for _, t := range tenants {
		ids[t.ID] = true
	}
	assert.Assert(t, ids[tenant1.ID], "tenant1 should be present in GetAllTenants result")
	assert.Assert(t, ids[tenant2.ID], "tenant2 should be present in GetAllTenants result")
}

// TestCreateDefaultTenant verifies that CreateDefaultTenant creates the default tenant when it does not exist,
// and returns the existing tenant when called again.
func TestCreateDefaultTenant(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	defaultId := db.GetDefaultTenantId()
	tenantId := strings.ToUpper(strings.TrimSpace(defaultId))
	dbClient := db.GetDatabaseClient()

	// Ensure the default tenant does not exist before the test
	dbClient.DeleteTenant(tenantId)

	// Should create the default tenant
	tenant, err := db.EnsureDefaultTenantExists()
	assert.NilError(t, err)
	assert.Assert(t, tenant != nil, "expected default tenant to be created")
	assert.Equal(t, tenant.ID, tenantId)
	err = dbClient.DeleteTenant(tenantId)
	assert.NilError(t, err)
}

// TestSetTenant_UpdateExisting verifies that calling SetTenant again updates the existing record.
func TestSetTenant_UpdateExisting(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	dbClient := db.GetDatabaseClient()
	tenant := newTestTenant()

	assert.NilError(t, dbClient.SetTenant(tenant))
	defer dbClient.DeleteTenant(tenant.ID)

	tenant.Name = "Updated Tenant Name"
	tenant.Updated = time.Now().UnixMilli()
	assert.NilError(t, dbClient.SetTenant(tenant))

	result, err := dbClient.GetTenant(tenant.ID)
	assert.NilError(t, err)
	assert.Assert(t, result != nil)
	assert.Equal(t, result.Name, "Updated Tenant Name")
}
