package common

import (
	"testing"
)

func TestConstants(t *testing.T) {
	// Test GenericNamespacedListTypes constants
	if GenericNamespacedListTypes_STRING != "STRING" {
		t.Errorf("Expected GenericNamespacedListTypes_STRING to be 'STRING', got '%s'", GenericNamespacedListTypes_STRING)
	}

	if GenericNamespacedListTypes_MAC_LIST != "MAC_LIST" {
		t.Errorf("Expected GenericNamespacedListTypes_MAC_LIST to be 'MAC_LIST', got '%s'", GenericNamespacedListTypes_MAC_LIST)
	}

	if GenericNamespacedListTypes_IP_LIST != "IP_LIST" {
		t.Errorf("Expected GenericNamespacedListTypes_IP_LIST to be 'IP_LIST', got '%s'", GenericNamespacedListTypes_IP_LIST)
	}

	if GenericNamespacedListTypes_RI_MAC_LIST != "RI_MAC_LIST" {
		t.Errorf("Expected GenericNamespacedListTypes_RI_MAC_LIST to be 'RI_MAC_LIST', got '%s'", GenericNamespacedListTypes_RI_MAC_LIST)
	}

	// Test header constants
	if NoPenetrationMetricsHeader != "X-No-Penetration-Metrics" {
		t.Errorf("Expected NoPenetrationMetricsHeader to be 'X-No-Penetration-Metrics', got '%s'", NoPenetrationMetricsHeader)
	}
}

func TestBinaryVersionVariables(t *testing.T) {
	// Test that binary version variables are defined (they may be empty by default)
	// These are typically set during build time

	// Just verify they exist and are strings (can be empty)
	if BinaryVersion == "" {
		t.Log("BinaryVersion is empty (expected for dev builds)")
	}

	if BinaryBranch == "" {
		t.Log("BinaryBranch is empty (expected for dev builds)")
	}

	if BinaryBuildTime == "" {
		t.Log("BinaryBuildTime is empty (expected for dev builds)")
	}

	// Test that CacheUpdateWindowSize is initialized (may be 0)
	// Just verify it exists
	_ = CacheUpdateWindowSize
}
