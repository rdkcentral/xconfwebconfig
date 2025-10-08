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
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rdkcentral/xconfwebconfig/db"

	"github.com/gocql/gocql"
	"gotest.tools/assert"
)

func TestAcquireLock(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	dbClient := db.GetDatabaseClient()
	truncateTable(db.TABLE_LOCKS)

	lockName := "testLock-" + uuid.New().String()
	lockUser1 := "testUser1"
	lockUser2 := "testUser2"
	ttlSeconds := 10

	// Acquire a new lock
	err := dbClient.AcquireLock(lockName, lockUser1, ttlSeconds)
	assert.NilError(t, err)

	// Verify lock info
	lockInfo, err := dbClient.GetLockInfo(lockName)
	assert.NilError(t, err)
	assert.Equal(t, lockInfo["locked_by"], lockUser1)

	// Fail to acquire an existing lock
	err = dbClient.AcquireLock(lockName, lockUser2, ttlSeconds)
	assert.ErrorContains(t, err, "failed to acquire lock")

	// Fail to release a lock not owned by the user
	err = dbClient.ReleaseLock(lockName, lockUser2)
	assert.ErrorContains(t, err, "failed to release lock")

	// Release the lock
	err = dbClient.ReleaseLock(lockName, lockUser1)
	assert.NilError(t, err)

	// Verify lock is gone
	lockInfo, err = dbClient.GetLockInfo(lockName)
	assert.Assert(t, errors.Is(err, gocql.ErrNotFound), "lock should not be found after release")
	assert.Assert(t, len(lockInfo) == 0, "lock info should be empty after release")

	// Acquire an expired lock (acquire lock with a short TTL)
	shortTtl := 1
	err = dbClient.AcquireLock(lockName, lockUser1, shortTtl)
	assert.NilError(t, err)

	// Wait for it to expire
	time.Sleep(time.Duration(shortTtl+1) * time.Second)

	// Acquire the now-expired lock
	err = dbClient.AcquireLock(lockName, lockUser2, ttlSeconds)
	assert.NilError(t, err)
}

func TestLockTable(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(db.TABLE_LOCKS)

	lockUser1 := "testUser1"
	lockUser2 := "testUser2"
	modelTableLock := db.NewDistributedLock(db.TABLE_MODEL, 5)
	assert.Assert(t, modelTableLock != nil)
	assert.Equal(t, modelTableLock.Name(), db.TABLE_MODEL)
	assert.Equal(t, modelTableLock.TTL(), 5)
	assert.Equal(t, modelTableLock.Retries(), 0)
	assert.Assert(t, modelTableLock.RetryInMsecs() > 0)

	modelTableLock.SetTTL(2)
	assert.Equal(t, modelTableLock.TTL(), 2)

	// Acquire a new lock
	err := modelTableLock.Lock(lockUser1)
	assert.NilError(t, err)

	// Fail to acquire an existing lock
	err = modelTableLock.Lock(lockUser2)
	assert.ErrorContains(t, err, "failed to acquire lock")

	// Fail to release a lock not owned by the user
	err = modelTableLock.Unlock(lockUser2)
	assert.ErrorContains(t, err, "failed to release lock")

	// Release the lock
	err = modelTableLock.Unlock(lockUser1)
	assert.NilError(t, err)

	// Acquire expired lock
	envLockTable := db.NewDistributedLock(db.TABLE_ENVIRONMENT, 1)
	assert.Assert(t, envLockTable != nil)

	// Acquire a new lock
	envLockTable.SetRetries(2)
	envLockTable.SetRetryInMsecs(100)
	err = envLockTable.Lock(lockUser1)
	assert.NilError(t, err)

	time.Sleep(1 * time.Second)

	// Acquire the now-expired lock
	err = envLockTable.Lock(lockUser2)
	assert.NilError(t, err)

	// Acquire a new lock with retries
	envLockTable.SetRetries(6)
	envLockTable.SetRetryInMsecs(200)
	err = envLockTable.Lock(lockUser1)
	assert.NilError(t, err)
}

func TestTableRowLock(t *testing.T) {
	if !db.IsCassandraClient() {
		t.Skip("Not using Cassandra DB")
	}

	truncateTable(db.TABLE_LOCKS)

	lockUser1 := "testUser1"
	lockUser2 := "testUser2"
	modelTableLock := db.NewDistributedLock(db.TABLE_MODEL, 1)
	assert.Assert(t, modelTableLock != nil)

	// Acquire a new lock
	modelTableLock.SetRetries(2)
	modelTableLock.SetRetryInMsecs(100)
	err := modelTableLock.LockRow(lockUser1, "key1")
	assert.NilError(t, err)

	err = modelTableLock.LockRow(lockUser2, "key2")
	assert.NilError(t, err)

	// Fail to acquire an existing lock on the same row
	err = modelTableLock.LockRow(lockUser1, "key2")
	assert.ErrorContains(t, err, "failed to acquire lock")

	// Release lock for key2
	err = modelTableLock.UnlockRow(lockUser2, "key2")
	assert.NilError(t, err)

	// Acquire lock for key2 again for lockUser1
	err = modelTableLock.LockRow(lockUser1, "key2")
	assert.NilError(t, err)

	time.Sleep(1 * time.Second)

	// Acquire the now-expired lock
	err = modelTableLock.LockRow(lockUser2, "key1")
	assert.NilError(t, err)
}
