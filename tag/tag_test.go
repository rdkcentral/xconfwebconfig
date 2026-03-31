package tag

import (
	"encoding/json"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/util"
)

func TestNewTagInf(t *testing.T) {
	tag := NewTagInf()
	if tag == nil {
		t.Error("NewTagInf should not return nil")
	}

	// Should return a pointer to Tag
	if _, ok := tag.(*Tag); !ok {
		t.Error("NewTagInf should return *Tag type")
	}
}

func TestTag_Clone(t *testing.T) {
	// Create original tag
	original := &Tag{
		Id:      "test-tag-123",
		Members: util.Set{},
		Updated: 1234567890,
	}
	original.Members.Add("member1", "member2", "member3")

	// Clone it
	cloned, err := original.Clone()
	if err != nil {
		t.Errorf("Clone failed: %v", err)
	}

	// Verify clone is not nil
	if cloned == nil {
		t.Fatal("Clone should not return nil")
	}

	// Verify fields are copied correctly
	if cloned.Id != original.Id {
		t.Errorf("Expected Id %s, got %s", original.Id, cloned.Id)
	}

	if cloned.Updated != original.Updated {
		t.Errorf("Expected Updated %d, got %d", original.Updated, cloned.Updated)
	}

	// Verify members are copied
	if !cloned.Members.Contains("member1") || !cloned.Members.Contains("member2") || !cloned.Members.Contains("member3") {
		t.Error("Clone should contain all members from original")
	}

	// Verify it's a deep copy (not sharing the same Members reference)
	cloned.Members.Add("new-member")
	if original.Members.Contains("new-member") {
		t.Error("Clone should be a deep copy - original should not be affected by changes to clone")
	}

	// Verify changes to original don't affect clone
	original.Members.Add("another-new-member")
	if cloned.Members.Contains("another-new-member") {
		t.Error("Clone should be independent - clone should not be affected by changes to original")
	}
}

func TestTag_MarshalJSON(t *testing.T) {
	// Test with populated tag
	tag := &Tag{
		Id:      "test-tag-456",
		Members: util.Set{},
		Updated: 1609459200, // 2021-01-01 00:00:00 UTC
	}
	tag.Members.Add("device1", "device2", "device3")

	data, err := json.Marshal(tag)
	if err != nil {
		t.Errorf("MarshalJSON failed: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}

	// Check fields
	if result["id"] != "test-tag-456" {
		t.Errorf("Expected id 'test-tag-456', got %v", result["id"])
	}

	if result["updated"] != float64(1609459200) {
		t.Errorf("Expected updated 1609459200, got %v", result["updated"])
	}

	// Check members is an array
	members, ok := result["members"].([]interface{})
	if !ok {
		t.Error("Members should be marshaled as array")
	}

	// Convert to string slice for easier checking
	memberStrings := make([]string, len(members))
	for i, m := range members {
		memberStrings[i] = m.(string)
	}

	// Check all members are present (order doesn't matter for sets)
	expectedMembers := []string{"device1", "device2", "device3"}
	if len(memberStrings) != len(expectedMembers) {
		t.Errorf("Expected %d members, got %d", len(expectedMembers), len(memberStrings))
	}

	for _, expected := range expectedMembers {
		found := false
		for _, actual := range memberStrings {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected member '%s' not found in marshaled data", expected)
		}
	}
}

func TestTag_MarshalJSON_EmptyMembers(t *testing.T) {
	// Test with empty members
	tag := &Tag{
		Id:      "empty-tag",
		Members: util.Set{},
		Updated: 0,
	}

	data, err := json.Marshal(tag)
	if err != nil {
		t.Errorf("MarshalJSON failed: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}

	// Check members exists and is an array (might be nil/null)
	members, ok := result["members"]
	if !ok {
		t.Error("Members field should exist in marshaled JSON")
	}

	// Members could be nil (marshaled as null) or empty array
	if members != nil {
		if membersArray, isArray := members.([]interface{}); isArray {
			if len(membersArray) != 0 {
				t.Errorf("Expected empty members array, got %d items", len(membersArray))
			}
		} else {
			t.Errorf("Members should be array or null, got %T", members)
		}
	}
}

func TestTag_UnmarshalJSON(t *testing.T) {
	// Test data with all fields
	jsonData := `{
		"id": "test-unmarshal-789",
		"members": ["device-a", "device-b", "device-c"],
		"updated": 1640995200
	}`

	var tag Tag
	err := json.Unmarshal([]byte(jsonData), &tag)
	if err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	// Verify fields
	if tag.Id != "test-unmarshal-789" {
		t.Errorf("Expected Id 'test-unmarshal-789', got '%s'", tag.Id)
	}

	if tag.Updated != 1640995200 {
		t.Errorf("Expected Updated 1640995200, got %d", tag.Updated)
	}

	// Verify members
	expectedMembers := []string{"device-a", "device-b", "device-c"}
	for _, expected := range expectedMembers {
		if !tag.Members.Contains(expected) {
			t.Errorf("Expected member '%s' not found", expected)
		}
	}

	if len(tag.Members) != len(expectedMembers) {
		t.Errorf("Expected %d members, got %d", len(expectedMembers), len(tag.Members))
	}
}

func TestTag_UnmarshalJSON_EmptyMembers(t *testing.T) {
	// Test with empty members array
	jsonData := `{
		"id": "empty-unmarshal",
		"members": [],
		"updated": 123
	}`

	var tag Tag
	err := json.Unmarshal([]byte(jsonData), &tag)
	if err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if len(tag.Members) != 0 {
		t.Errorf("Expected empty members, got %d", len(tag.Members))
	}
}

func TestTag_UnmarshalJSON_InvalidJSON(t *testing.T) {
	// Test with invalid JSON
	invalidJSON := `{"id": "test", "members": [`

	var tag Tag
	err := json.Unmarshal([]byte(invalidJSON), &tag)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestTag_MarshalUnmarshal_RoundTrip(t *testing.T) {
	// Create original tag
	original := &Tag{
		Id:      "roundtrip-test",
		Members: util.Set{},
		Updated: 1234567890,
	}
	original.Members.Add("member1", "member2", "member3")

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}

	// Unmarshal
	var reconstructed Tag
	err = json.Unmarshal(data, &reconstructed)
	if err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}

	// Verify round trip
	if reconstructed.Id != original.Id {
		t.Errorf("Round trip failed for Id: expected %s, got %s", original.Id, reconstructed.Id)
	}

	if reconstructed.Updated != original.Updated {
		t.Errorf("Round trip failed for Updated: expected %d, got %d", original.Updated, reconstructed.Updated)
	}

	if len(reconstructed.Members) != len(original.Members) {
		t.Errorf("Round trip failed for Members size: expected %d, got %d", len(original.Members), len(reconstructed.Members))
	}

	// Check all members are preserved
	for _, member := range original.Members.ToSlice() {
		if !reconstructed.Members.Contains(member) {
			t.Errorf("Member '%s' lost in round trip", member)
		}
	}
}

func TestTag_JSONWithSpecialCharacters(t *testing.T) {
	// Test with special characters in members and id
	tag := &Tag{
		Id:      "special-chars-test-123",
		Members: util.Set{},
		Updated: 1000000000,
	}
	tag.Members.Add("device/with/slashes", "device with spaces", "device-with-dashes", "device_with_underscores")

	// Marshal
	data, err := json.Marshal(tag)
	if err != nil {
		t.Errorf("Marshal with special characters failed: %v", err)
	}

	// Unmarshal
	var reconstructed Tag
	err = json.Unmarshal(data, &reconstructed)
	if err != nil {
		t.Errorf("Unmarshal with special characters failed: %v", err)
	}

	// Verify special characters are preserved
	specialMembers := []string{"device/with/slashes", "device with spaces", "device-with-dashes", "device_with_underscores"}
	for _, member := range specialMembers {
		if !reconstructed.Members.Contains(member) {
			t.Errorf("Special character member '%s' not preserved", member)
		}
	}
}

func TestTag_CloneError(t *testing.T) {
	// Create a tag with a circular reference that might cause Clone to fail
	// This is difficult to achieve with the current Tag struct, but we can test error handling
	tag := &Tag{
		Id:      "test-clone-error",
		Members: util.Set{},
		Updated: 123456789,
	}
	tag.Members.Add("member1", "member2")

	// For this simple struct, Clone should not fail
	cloned, err := tag.Clone()
	if err != nil {
		t.Errorf("Unexpected error in Clone: %v", err)
	}

	if cloned == nil {
		t.Error("Clone returned nil without error")
	}
}
