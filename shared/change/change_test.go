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
package change

import (
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"gotest.tools/assert"
)

// Test Change struct getters
func TestChange_Getters(t *testing.T) {
	profile := &logupload.PermanentTelemetryProfile{
		Name: "test-profile",
	}

	change := &Change{
		ID:              "change-1",
		EntityID:        "entity-1",
		EntityType:      TelemetryProfile,
		ApplicationType: shared.STB,
		NewEntity:       profile,
		OldEntity:       profile,
		Operation:       Update,
		Author:          "john.doe",
		ApprovedUser:    "jane.smith",
	}

	assert.Equal(t, "change-1", change.GetID())
	assert.Equal(t, "entity-1", change.GetEntityID())
	assert.Equal(t, TelemetryProfile, change.GetEntityType())
	assert.Equal(t, shared.STB, change.GetApplicationType())
	assert.Equal(t, profile, change.GetNewEntity())
	assert.Equal(t, profile, change.GetOldEntity())
	assert.Assert(t, change.GetOperation() == Update)
	assert.Equal(t, "john.doe", change.GetAuthor())
	assert.Equal(t, "jane.smith", change.GetApprovedUser())
}

// Test ApprovedChange struct getters
func TestApprovedChange_Getters(t *testing.T) {
	profile := &logupload.PermanentTelemetryProfile{
		Name: "approved-profile",
	}

	approvedChange := &ApprovedChange{
		ID:              "approved-1",
		EntityID:        "entity-2",
		EntityType:      TelemetryProfile,
		ApplicationType: shared.XHOME,
		NewEntity:       profile,
		OldEntity:       nil,
		Operation:       Create,
		Author:          "alice",
		ApprovedUser:    "bob",
	}

	assert.Equal(t, "approved-1", approvedChange.GetID())
	assert.Equal(t, "entity-2", approvedChange.GetEntityID())
	assert.Equal(t, TelemetryProfile, approvedChange.GetEntityType())
	assert.Equal(t, shared.XHOME, approvedChange.GetApplicationType())
	assert.Equal(t, profile, approvedChange.GetNewEntity())
	assert.Assert(t, approvedChange.GetOldEntity() == nil)
	assert.Assert(t, approvedChange.GetOperation() == Create)
	assert.Equal(t, "alice", approvedChange.GetAuthor())
	assert.Equal(t, "bob", approvedChange.GetApprovedUser())
}

// Test TelemetryTwoChange Validate - CREATE operation
func TestTelemetryTwoChange_Validate_Create(t *testing.T) {
	profile := &logupload.TelemetryTwoProfile{
		Name: "profile1",
	}

	change := &TelemetryTwoChange{
		EntityID:  "entity-1",
		Author:    "author1",
		Operation: Create,
		NewEntity: profile,
	}

	err := change.Validate()
	assert.NilError(t, err)
}

// Test TelemetryTwoChange Validate - UPDATE operation
func TestTelemetryTwoChange_Validate_Update(t *testing.T) {
	newProfile := &logupload.TelemetryTwoProfile{Name: "new"}
	oldProfile := &logupload.TelemetryTwoProfile{Name: "old"}

	change := &TelemetryTwoChange{
		EntityID:  "entity-1",
		Author:    "author1",
		Operation: Update,
		NewEntity: newProfile,
		OldEntity: oldProfile,
	}

	err := change.Validate()
	assert.NilError(t, err)
}

// Test TelemetryTwoChange Validate - DELETE operation
func TestTelemetryTwoChange_Validate_Delete(t *testing.T) {
	oldProfile := &logupload.TelemetryTwoProfile{Name: "old"}

	change := &TelemetryTwoChange{
		EntityID:  "entity-1",
		Author:    "author1",
		Operation: Delete,
		OldEntity: oldProfile,
	}

	err := change.Validate()
	assert.NilError(t, err)
}

// Test TelemetryTwoChange Validate - Empty EntityID
func TestTelemetryTwoChange_Validate_EmptyEntityID(t *testing.T) {
	change := &TelemetryTwoChange{
		EntityID:  "",
		Author:    "author1",
		Operation: Create,
	}

	err := change.Validate()
	assert.Error(t, err, "Entity id is empty")
}

// Test TelemetryTwoChange Validate - Empty Author
func TestTelemetryTwoChange_Validate_EmptyAuthor(t *testing.T) {
	change := &TelemetryTwoChange{
		EntityID:  "entity-1",
		Author:    "",
		Operation: Create,
	}

	err := change.Validate()
	assert.Error(t, err, "Author is empty")
}

// Test TelemetryTwoChange Validate - Empty Operation
func TestTelemetryTwoChange_Validate_EmptyOperation(t *testing.T) {
	change := &TelemetryTwoChange{
		EntityID: "entity-1",
		Author:   "author1",
	}

	err := change.Validate()
	assert.Error(t, err, "Operation is empty")
}

// Test TelemetryTwoChange Validate - CREATE missing NewEntity
func TestTelemetryTwoChange_Validate_CreateMissingNewEntity(t *testing.T) {
	change := &TelemetryTwoChange{
		EntityID:  "entity-1",
		Author:    "author1",
		Operation: Create,
		NewEntity: nil,
	}

	err := change.Validate()
	assert.Error(t, err, "New entity is empty")
}

// Test TelemetryTwoChange Validate - UPDATE missing NewEntity
func TestTelemetryTwoChange_Validate_UpdateMissingNewEntity(t *testing.T) {
	oldProfile := &logupload.TelemetryTwoProfile{Name: "old"}

	change := &TelemetryTwoChange{
		EntityID:  "entity-1",
		Author:    "author1",
		Operation: Update,
		NewEntity: nil,
		OldEntity: oldProfile,
	}

	err := change.Validate()
	assert.Error(t, err, "New entity is empty")
}

// Test TelemetryTwoChange Validate - UPDATE missing OldEntity
func TestTelemetryTwoChange_Validate_UpdateMissingOldEntity(t *testing.T) {
	newProfile := &logupload.TelemetryTwoProfile{Name: "new"}

	change := &TelemetryTwoChange{
		EntityID:  "entity-1",
		Author:    "author1",
		Operation: Update,
		NewEntity: newProfile,
		OldEntity: nil,
	}

	err := change.Validate()
	assert.Error(t, err, "Old entity is empty")
}

// Test TelemetryTwoChange Validate - DELETE missing OldEntity
func TestTelemetryTwoChange_Validate_DeleteMissingOldEntity(t *testing.T) {
	change := &TelemetryTwoChange{
		EntityID:  "entity-1",
		Author:    "author1",
		Operation: Delete,
		OldEntity: nil,
	}

	err := change.Validate()
	assert.Error(t, err, "Old entity is empty")
}

// Test ApprovedTelemetryTwoChange Validate - Valid
func TestApprovedTelemetryTwoChange_Validate_Valid(t *testing.T) {
	profile := &logupload.TelemetryTwoProfile{Name: "profile"}

	change := &ApprovedTelemetryTwoChange{
		EntityID:     "entity-1",
		Author:       "author1",
		Operation:    Create,
		NewEntity:    profile,
		ApprovedUser: "approver1",
	}

	err := change.Validate()
	assert.NilError(t, err)
}

// Test ApprovedTelemetryTwoChange Validate - Missing ApprovedUser
func TestApprovedTelemetryTwoChange_Validate_MissingApprovedUser(t *testing.T) {
	profile := &logupload.TelemetryTwoProfile{Name: "profile"}

	change := &ApprovedTelemetryTwoChange{
		EntityID:     "entity-1",
		Author:       "author1",
		Operation:    Create,
		NewEntity:    profile,
		ApprovedUser: "",
	}

	err := change.Validate()
	assert.Error(t, err, "Approved user is empty")
}

// Test constructors
func TestNewChangeInf(t *testing.T) {
	obj := NewChangeInf()

	assert.Assert(t, obj != nil)
	change, ok := obj.(*Change)
	assert.Assert(t, ok)
	assert.Equal(t, shared.STB, change.ApplicationType)
}

func TestNewEmptyChange(t *testing.T) {
	change := NewEmptyChange()

	assert.Assert(t, change != nil)
	assert.Equal(t, shared.STB, change.ApplicationType)
}

func TestNewApprovedChangeInf(t *testing.T) {
	obj := NewApprovedChangeInf()

	assert.Assert(t, obj != nil)
	change, ok := obj.(*ApprovedChange)
	assert.Assert(t, ok)
	assert.Equal(t, shared.STB, change.ApplicationType)
}

func TestNewTelemetryTwoChangeInf(t *testing.T) {
	obj := NewTelemetryTwoChangeInf()

	assert.Assert(t, obj != nil)
	change, ok := obj.(*TelemetryTwoChange)
	assert.Assert(t, ok)
	assert.Equal(t, shared.STB, change.ApplicationType)
}

func TestNewApprovedTelemetryTwoChangeInf(t *testing.T) {
	obj := NewApprovedTelemetryTwoChangeInf()

	assert.Assert(t, obj != nil)
	change, ok := obj.(*ApprovedTelemetryTwoChange)
	assert.Assert(t, ok)
	assert.Equal(t, shared.STB, change.ApplicationType)
}

// Test EqualChangeData
func TestChange_EqualChangeData_Same(t *testing.T) {
	profile := &logupload.PermanentTelemetryProfile{Name: "test"}

	change1 := &Change{
		EntityType:      TelemetryProfile,
		ApplicationType: shared.STB,
		NewEntity:       profile,
		OldEntity:       nil,
		Operation:       Create,
	}

	change2 := &Change{
		EntityType:      TelemetryProfile,
		ApplicationType: shared.STB,
		NewEntity:       profile,
		OldEntity:       nil,
		Operation:       Create,
	}

	assert.Assert(t, change1.EqualChangeData(change2))
}

func TestChange_EqualChangeData_Different(t *testing.T) {
	profile1 := &logupload.PermanentTelemetryProfile{Name: "test1"}
	profile2 := &logupload.PermanentTelemetryProfile{Name: "test2"}

	change1 := &Change{
		EntityType:      TelemetryProfile,
		ApplicationType: shared.STB,
		NewEntity:       profile1,
		OldEntity:       nil,
		Operation:       Create,
	}

	change2 := &Change{
		EntityType:      TelemetryProfile,
		ApplicationType: shared.STB,
		NewEntity:       profile2,
		OldEntity:       nil,
		Operation:       Create,
	}

	assert.Assert(t, !change1.EqualChangeData(change2))
}

func TestChange_EqualChangeData_SameReference(t *testing.T) {
	change := &Change{
		EntityType:      TelemetryProfile,
		ApplicationType: shared.STB,
	}

	assert.Assert(t, change.EqualChangeData(change))
}

// Test ChangeOperation constants
func TestChangeOperation_Constants(t *testing.T) {
	assert.Equal(t, "CREATE", string(Create))
	assert.Equal(t, "UPDATE", string(Update))
	assert.Equal(t, "DELETE", string(Delete))
}

// Test EntityType constants
func TestEntityType_Constants(t *testing.T) {
	assert.Equal(t, "TELEMETRY_PROFILE", string(TelemetryProfile))
}

// Test byAuthor
func TestChange_byAuthor(t *testing.T) {
	change := &Change{Author: "JohnDoe"}

	assert.Assert(t, change.byAuthor("johndoe"))
	assert.Assert(t, change.byAuthor("JOHNDOE"))
	assert.Assert(t, change.byAuthor("JohnDoe"))
	assert.Assert(t, !change.byAuthor("janedoe"))
}

// Test byTelemetryProfileName
func TestChange_byTelemetryProfileName(t *testing.T) {
	newProfile := &logupload.PermanentTelemetryProfile{Name: "NewProfile"}
	oldProfile := &logupload.PermanentTelemetryProfile{Name: "OldProfile"}

	change := &Change{
		NewEntity: newProfile,
		OldEntity: oldProfile,
	}

	assert.Assert(t, change.byTelemetryProfileName("newprofile"))
	assert.Assert(t, change.byTelemetryProfileName("NEWPROFILE"))
	assert.Assert(t, change.byTelemetryProfileName("oldprofile"))
	assert.Assert(t, change.byTelemetryProfileName("OLDPROFILE"))
	assert.Assert(t, !change.byTelemetryProfileName("different"))
}
