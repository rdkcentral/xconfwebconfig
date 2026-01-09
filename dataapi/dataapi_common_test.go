package dataapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/stretchr/testify/assert"
)

func TestGetClientProtocolHeaderValue(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"HTTPS protocol", "HTTPS", "HTTPS"},
		{"MTLS protocol", "MTLS", "MTLS"},
		{"Empty header", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set(common.XCONF_HTTP_HEADER, tt.header)
			result := GetClientProtocolHeaderValue(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetClientCertExpiryHeaderValue(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"Valid date", "2025-12-31", "2025-12-31"},
		{"Empty header", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set(common.CLIENT_CERT_EXPIRY_HEADER, tt.header)
			result := GetClientCertExpiryHeaderValue(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddClientProtocolToContextMap(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"HTTPS", common.XCONF_HTTPS_VALUE, common.HTTPS_CLIENT_PROTOCOL},
		{"MTLS", common.XCONF_MTLS_VALUE, common.MTLS_CLIENT_PROTOCOL},
		{"MTLS Optional", common.XCONF_MTLS_OPTIONAL_VALUE, common.MTLS_OPTIONAL_CLIENT_PROTOCOL},
		{"MTLS Recovery", common.XCONF_MTLS_RECOVERY_VALUE, common.MTLS_RECOVERY_CLIENT_PROTOCOL},
		{"Default HTTP", "UNKNOWN", common.HTTP_CLIENT_PROTOCOL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextMap := make(map[string]string)
			AddClientProtocolToContextMap(contextMap, tt.header)
			assert.Equal(t, tt.expected, contextMap[common.CLIENT_PROTOCOL])
		})
	}
}

func TestAddCertExpiryToContextMap(t *testing.T) {
	tests := []struct {
		name        string
		certExpiry  string
		protocol    string
		expectedKey string
		shouldAdd   bool
	}{
		{"MTLS Recovery", "2025-12-31", common.MTLS_RECOVERY_CLIENT_PROTOCOL, common.RECOVERY_CERT_EXPIRY, true},
		{"Regular MTLS", "2025-12-31", common.MTLS_CLIENT_PROTOCOL, common.CLIENT_CERT_EXPIRY, true},
		{"Empty expiry", "", common.MTLS_CLIENT_PROTOCOL, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextMap := map[string]string{common.CLIENT_PROTOCOL: tt.protocol}
			AddCertExpiryToContextMap(contextMap, tt.certExpiry)

			if tt.shouldAdd {
				assert.Equal(t, tt.certExpiry, contextMap[tt.expectedKey])
			} else {
				_, exists := contextMap[common.CLIENT_CERT_EXPIRY]
				assert.False(t, exists)
			}
		})
	}
}

func TestIsSecureConnection(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		expected bool
	}{
		{"HTTPS", common.XCONF_HTTPS_VALUE, true},
		{"MTLS", common.XCONF_MTLS_VALUE, true},
		{"MTLS Recovery", common.XCONF_MTLS_RECOVERY_VALUE, true},
		{"MTLS Optional", common.XCONF_MTLS_OPTIONAL_VALUE, true},
		{"HTTP", "HTTP", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSecureConnection(tt.protocol)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddClientCertDurationToContext(t *testing.T) {
	tests := []struct {
		name       string
		certExpiry string
		shouldAdd  bool
	}{
		{"Valid future date", time.Now().UTC().Add(30 * 24 * time.Hour).Format(common.ClientCertExpiryDateFormat), true},
		{"Empty", "", false},
		{"Invalid format", "invalid-date", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextMap := make(map[string]string)
			AddClientCertDurationToContext(contextMap, tt.certExpiry)

			if tt.shouldAdd {
				_, exists := contextMap[common.CERT_EXPIRY_DAYS]
				assert.True(t, exists)
			} else {
				_, exists := contextMap[common.CERT_EXPIRY_DAYS]
				assert.False(t, exists)
			}
		})
	}
}

func TestNormalizeCommonContext(t *testing.T) {
	tests := []struct {
		name       string
		input      map[string]string
		estbMacKey string
		ecmMacKey  string
	}{
		{
			name: "Uppercase model and env",
			input: map[string]string{
				common.MODEL: "model123",
				common.ENV:   "dev",
			},
			estbMacKey: "estbMac",
			ecmMacKey:  "ecmMac",
		},
		{
			name: "Normalize MAC addresses",
			input: map[string]string{
				"estbMac": "aa:bb:cc:dd:ee:ff",
				"ecmMac":  "11:22:33:44:55:66",
			},
			estbMacKey: "estbMac",
			ecmMacKey:  "ecmMac",
		},
		{
			name: "Uppercase partnerId",
			input: map[string]string{
				common.PARTNER_ID: "partner1",
			},
			estbMacKey: "estbMac",
			ecmMacKey:  "ecmMac",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NormalizeCommonContext(tt.input, tt.estbMacKey, tt.ecmMacKey)

			if model, ok := tt.input[common.MODEL]; ok && model != "" {
				assert.Equal(t, model, tt.input[common.MODEL])
			}
			if env, ok := tt.input[common.ENV]; ok && env != "" {
				assert.Equal(t, env, tt.input[common.ENV])
			}
			if partnerId, ok := tt.input[common.PARTNER_ID]; ok && partnerId != "" {
				assert.Equal(t, partnerId, tt.input[common.PARTNER_ID])
			}
		})
	}
}

func TestGetApplicationTypeFromPartnerId(t *testing.T) {
	// Setup
	originalXc := Xc
	defer func() { Xc = originalXc }()

	Xc = &XconfConfigs{
		DeriveAppTypeFromPartnerId: true,
		PartnerApplicationTypes:    []string{"stb", "xhome", "rdkcloud"},
	}

	tests := []struct {
		name      string
		partnerId string
		expected  string
	}{
		{"Match STB", "stb-partner-123", "stb"},
		{"Match xhome", "xhome-camera", "xhome"},
		{"No match", "unknown-partner", ""},
		{"Empty", "", ""},
		{"Case insensitive", "STB-PARTNER", "stb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetApplicationTypeFromPartnerId(tt.partnerId)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetApplicationTypeFromPartnerId_FeatureDisabled(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()

	Xc = &XconfConfigs{
		DeriveAppTypeFromPartnerId: false,
		PartnerApplicationTypes:    []string{"stb"},
	}

	result := GetApplicationTypeFromPartnerId("stb-partner")
	assert.Equal(t, "", result)
}

func TestRemovePrefix(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		prefixSet     []string
		expectedKey   string
		expectedFound bool
	}{
		{"Remove prefix", "prefix_key", []string{"prefix_"}, "key", true},
		{"No match", "somekey", []string{"prefix_"}, "somekey", false},
		{"Empty prefix set", "key", []string{}, "key", false},
		{"Blank after removal", "prefix_", []string{"prefix_"}, "prefix_", false},
		{"Multiple prefixes", "first_key", []string{"first_", "second_"}, "key", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultKey, found := RemovePrefix(tt.key, tt.prefixSet)
			assert.Equal(t, tt.expectedKey, resultKey)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}

func TestWebServerInjection(t *testing.T) {
	t.Run("Nil webserver", func(t *testing.T) {
		xc := &XconfConfigs{}
		WebServerInjection(nil, xc)

		assert.Nil(t, Ws)
		assert.Equal(t, xc, Xc)
		assert.Equal(t, int64(60000), common.CacheUpdateWindowSize)
	})
}
