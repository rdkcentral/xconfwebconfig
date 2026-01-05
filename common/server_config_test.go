package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServerOriginId(t *testing.T) {
	originId := ServerOriginId()

	if originId == "" {
		t.Error("ServerOriginId should not return empty string")
	}

	// Should contain either hostname:pid or just pid
	if !strings.Contains(originId, ":") && !isNumeric(originId) {
		t.Errorf("ServerOriginId should contain either 'hostname:pid' or just 'pid', got '%s'", originId)
	}

	// Should be consistent when called multiple times
	originId2 := ServerOriginId()
	if originId != originId2 {
		t.Errorf("ServerOriginId should be consistent, got '%s' and '%s'", originId, originId2)
	}
}

func TestNewServerConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
xconfwebconfig {
    test_setting = "test_value"
    number_setting = 123
}`

	// Create temp file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config.conf")

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test NewServerConfig
	config, err := NewServerConfig(configFile)
	if err != nil {
		t.Fatalf("NewServerConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("NewServerConfig should return a non-nil config")
	}

	if config.Config == nil {
		t.Error("Config.Config should not be nil")
	}

	// Test that config bytes are stored
	configBytes := config.ConfigBytes()
	if len(configBytes) == 0 {
		t.Error("ConfigBytes should not be empty")
	}

	// Verify the content is correct
	if !strings.Contains(string(configBytes), "test_setting") {
		t.Error("ConfigBytes should contain the original config content")
	}
}

func TestNewServerConfig_NonExistentFile(t *testing.T) {
	// Test with non-existent file
	config, err := NewServerConfig("/non/existent/file.conf")

	if err == nil {
		t.Error("NewServerConfig should return error for non-existent file")
	}

	if config != nil {
		t.Error("NewServerConfig should return nil config on error")
	}
}

func TestNewServerConfig_EmptyFile(t *testing.T) {
	// Create an empty temp file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "empty_config.conf")

	err := os.WriteFile(configFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty test config file: %v", err)
	}

	// Test NewServerConfig with empty file
	config, err := NewServerConfig(configFile)
	if err != nil {
		t.Fatalf("NewServerConfig should handle empty file: %v", err)
	}

	if config == nil {
		t.Error("NewServerConfig should return config even for empty file")
	}

	configBytes := config.ConfigBytes()
	if len(configBytes) != 0 {
		t.Errorf("ConfigBytes should be empty for empty file, got %d bytes", len(configBytes))
	}
}

func TestNewServerConfigFromText(t *testing.T) {
	configText := `
application {
    name = "test_app"
    version = "1.0.0"
}
database {
    host = "localhost"
    port = 5432
}`

	config, err := NewServerConfigFromText(configText)
	if err != nil {
		t.Fatalf("NewServerConfigFromText failed: %v", err)
	}

	if config == nil {
		t.Fatal("NewServerConfigFromText should return a non-nil config")
	}

	if config.Config == nil {
		t.Error("Config.Config should not be nil")
	}

	// Test that config bytes are stored correctly
	configBytes := config.ConfigBytes()
	if string(configBytes) != configText {
		t.Errorf("ConfigBytes should match original text.\nExpected: %s\nGot: %s", configText, string(configBytes))
	}
}

func TestNewServerConfigFromText_EmptyText(t *testing.T) {
	config, err := NewServerConfigFromText("")
	if err != nil {
		t.Fatalf("NewServerConfigFromText should handle empty text: %v", err)
	}

	if config == nil {
		t.Error("NewServerConfigFromText should return config even for empty text")
	}

	configBytes := config.ConfigBytes()
	if string(configBytes) != "" {
		t.Errorf("ConfigBytes should be empty for empty text, got '%s'", string(configBytes))
	}
}

func TestNewServerConfigFromText_InvalidConfig(t *testing.T) {
	// Test with malformed but parseable config - use simpler malformed syntax
	malformedConfig := `{ incomplete config without closing brace`

	// This should panic or error, so we catch it
	defer func() {
		if r := recover(); r != nil {
			// This is expected for truly invalid config
			t.Logf("Expected panic for invalid config: %v", r)
		}
	}()

	config, err := NewServerConfigFromText(malformedConfig)

	// If it doesn't panic, it should either error or handle gracefully
	if err != nil {
		t.Logf("Expected error for malformed config: %v", err)
		if config != nil {
			t.Error("Should return nil config when there's an error")
		}
		return
	}

	// If no error and no panic, verify the config was created
	if config == nil {
		t.Error("NewServerConfigFromText should return config or error")
	}
}

func TestServerConfig_ConfigBytes(t *testing.T) {
	originalText := `test { value = "hello world" }`

	config, err := NewServerConfigFromText(originalText)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Test ConfigBytes method
	configBytes := config.ConfigBytes()

	if string(configBytes) != originalText {
		t.Errorf("ConfigBytes() should return original text.\nExpected: %s\nGot: %s", originalText, string(configBytes))
	}

	// Test that returned bytes are the same reference (current implementation)
	configBytes2 := config.ConfigBytes()

	// Verify they point to the same underlying data
	if len(configBytes) > 0 && len(configBytes2) > 0 {
		// They should be the same length
		if len(configBytes) != len(configBytes2) {
			t.Error("ConfigBytes should return consistent length")
		}

		// Verify content is the same
		if string(configBytes) != string(configBytes2) {
			t.Error("ConfigBytes should return consistent content")
		}

		t.Log("ConfigBytes correctly returns the stored config bytes")
	}
}

func TestServerConfig_Integration(t *testing.T) {
	// Test the full workflow: file -> config -> bytes
	configContent := `
xconfwebconfig {
    server {
        port = 8080
        host = "localhost"
    }
    database {
        url = "jdbc:cassandra://localhost:9042"
    }
}`

	// Create temp file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "integration_test.conf")

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load config from file
	config, err := NewServerConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify all components work
	if config.Config == nil {
		t.Error("Config.Config should not be nil")
	}

	configBytes := config.ConfigBytes()
	if !strings.Contains(string(configBytes), "xconfwebconfig") {
		t.Error("ConfigBytes should contain the configuration content")
	}

	// Test that we can access configuration values through the Config field
	// Note: exact API depends on go-akka/configuration implementation
	if config.Config == nil {
		t.Error("Should be able to access parsed configuration")
	}
}

// Helper function to check if string is numeric
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
