package tracing

import (
	"context"
	"testing"
)

func TestSetContext(t *testing.T) {
	// Test setting string value
	ctx := context.Background()
	ctxName := "user_id"
	ctxValue := "12345"

	newCtx := SetContext(ctx, ctxName, ctxValue)

	if newCtx == nil {
		t.Fatal("SetContext should return a non-nil context")
	}

	// Verify the value was set
	retrievedValue := GetContext(newCtx, ctxName)
	if retrievedValue != ctxValue {
		t.Errorf("Expected value %v, got %v", ctxValue, retrievedValue)
	}
}

func TestSetContext_MultipleValues(t *testing.T) {
	ctx := context.Background()

	// Set multiple values
	ctx = SetContext(ctx, "user_id", "12345")
	ctx = SetContext(ctx, "session_id", "abcdef")
	ctx = SetContext(ctx, "request_count", 42)

	// Verify all values are accessible
	if GetContext(ctx, "user_id") != "12345" {
		t.Error("user_id not set correctly")
	}
	if GetContext(ctx, "session_id") != "abcdef" {
		t.Error("session_id not set correctly")
	}
	if GetContext(ctx, "request_count") != 42 {
		t.Error("request_count not set correctly")
	}
}

func TestSetContext_OverwriteValue(t *testing.T) {
	ctx := context.Background()
	key := "test_key"

	// Set initial value
	ctx = SetContext(ctx, key, "initial")

	// Overwrite with new value
	ctx = SetContext(ctx, key, "updated")

	// Verify the value was updated
	retrievedValue := GetContext(ctx, key)
	if retrievedValue != "updated" {
		t.Errorf("Expected updated value 'updated', got %v", retrievedValue)
	}
}

func TestGetContext(t *testing.T) {
	ctx := context.Background()
	key := "test_key"
	expectedValue := "test_value"

	// Set a value
	ctx = SetContext(ctx, key, expectedValue)

	// Retrieve the value
	actualValue := GetContext(ctx, key)

	if actualValue != expectedValue {
		t.Errorf("Expected %v, got %v", expectedValue, actualValue)
	}
}

func TestGetContext_NonExistentKey(t *testing.T) {
	ctx := context.Background()

	// Try to get a value that doesn't exist
	value := GetContext(ctx, "non_existent_key")

	if value != nil {
		t.Errorf("Expected nil for non-existent key, got %v", value)
	}
}

func TestGetContext_DifferentTypes(t *testing.T) {
	ctx := context.Background()

	// Test with different data types
	testCases := []struct {
		key   string
		value interface{}
	}{
		{"string_val", "hello"},
		{"int_val", 123},
		{"bool_val", true},
		{"nil_val", nil},
	}

	// Set all values
	for _, tc := range testCases {
		ctx = SetContext(ctx, tc.key, tc.value)
	}

	// Verify all values
	for _, tc := range testCases {
		retrievedValue := GetContext(ctx, tc.key)
		if retrievedValue != tc.value {
			t.Errorf("For key %s, expected %v, got %v", tc.key, tc.value, retrievedValue)
		}
	}

	// Test slice separately (can't use == comparison)
	sliceKey := "slice_val"
	sliceValue := []string{"a", "b", "c"}
	ctx = SetContext(ctx, sliceKey, sliceValue)

	retrievedSlice := GetContext(ctx, sliceKey)
	if retrievedSlice == nil {
		t.Error("Slice value should not be nil")
	}

	// Test map separately (can't use == comparison)
	mapKey := "map_val"
	mapValue := map[string]string{"key": "value"}
	ctx = SetContext(ctx, mapKey, mapValue)

	retrievedMap := GetContext(ctx, mapKey)
	if retrievedMap == nil {
		t.Error("Map value should not be nil")
	}
}

func TestContextIsolation(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()

	// Set different values in each context
	ctx1 = SetContext(ctx1, "key", "value1")
	ctx2 = SetContext(ctx2, "key", "value2")

	// Verify they don't interfere with each other
	val1 := GetContext(ctx1, "key")
	val2 := GetContext(ctx2, "key")

	if val1 != "value1" {
		t.Errorf("ctx1 should have value1, got %v", val1)
	}
	if val2 != "value2" {
		t.Errorf("ctx2 should have value2, got %v", val2)
	}
}
