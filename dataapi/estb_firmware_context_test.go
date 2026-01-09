package dataapi

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/dataapi/estbfirmware"
	coreef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/rdkcentral/xconfwebconfig/shared/firmware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsMacPresentAndValid(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   url.Values
		expectedValid bool
		expectedMac   string
		expectedError string
	}{
		{
			name: "Valid MAC address",
			queryParams: url.Values{
				common.MAC: []string{"AA:BB:CC:DD:EE:FF"},
			},
			expectedValid: true,
			expectedMac:   "AA:BB:CC:DD:EE:FF",
			expectedError: "",
		},
		{
			name:          "Missing MAC parameter",
			queryParams:   url.Values{},
			expectedValid: false,
			expectedMac:   "",
			expectedError: "Required String parameter 'mac' is not present",
		},
		{
			name: "Invalid MAC format",
			queryParams: url.Values{
				common.MAC: []string{"invalid-mac"},
			},
			expectedValid: false,
			expectedMac:   "invalid-mac",
			expectedError: "Mac is invalid: invalid-mac",
		},
		{
			name: "Empty MAC value",
			queryParams: url.Values{
				common.MAC: []string{""},
			},
			expectedValid: false,
			expectedMac:   "",
			expectedError: "Required String parameter 'mac' is not present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, mac, errStr := IsMacPresentAndValid(tt.queryParams)
			assert.Equal(t, tt.expectedValid, valid)
			assert.Equal(t, tt.expectedMac, mac)
			if tt.expectedError != "" {
				assert.Contains(t, errStr, tt.expectedError)
			}
		})
	}
}

func TestGetTimeInLocalTimezone(t *testing.T) {
	tests := []struct {
		name           string
		currentTime    time.Time
		timezoneOffset string
		shouldHaveTime bool
	}{
		{
			name:           "Positive timezone offset +05:30",
			currentTime:    time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
			timezoneOffset: "5:30",
			shouldHaveTime: true,
		},
		{
			name:           "Negative timezone offset -08:00",
			currentTime:    time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
			timezoneOffset: "-8:0",
			shouldHaveTime: true,
		},
		{
			name:           "Zero timezone offset",
			currentTime:    time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
			timezoneOffset: "0:0",
			shouldHaveTime: true,
		},
		{
			name:           "No timezone offset",
			currentTime:    time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
			timezoneOffset: "",
			shouldHaveTime: true,
		},
		{
			name:           "Invalid timezone format - single value",
			currentTime:    time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
			timezoneOffset: "5",
			shouldHaveTime: true,
		},
		{
			name:           "Invalid hours - out of range",
			currentTime:    time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
			timezoneOffset: "25:30",
			shouldHaveTime: true,
		},
		{
			name:           "Invalid minutes - out of range",
			currentTime:    time.Date(2025, 10, 27, 12, 0, 0, 0, time.UTC),
			timezoneOffset: "5:65",
			shouldHaveTime: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextMap := make(map[string]string)
			if tt.timezoneOffset != "" {
				contextMap[common.TIME_ZONE_OFFSET] = tt.timezoneOffset
			}

			GetTimeInLocalTimezone(tt.currentTime, contextMap)

			if tt.shouldHaveTime {
				assert.NotEmpty(t, contextMap[common.TIME])
			}
		})
	}
}

func TestIsAllowedRequest(t *testing.T) {
	// Setup
	originalXc := Xc
	defer func() { Xc = originalXc }()

	tests := []struct {
		name                     string
		contextMap               map[string]string
		clientProtocolHeader     string
		recoveryFirmwareVersions string
		expectedAllowed          bool
	}{
		{
			name:                     "Secure connection - HTTPS",
			contextMap:               map[string]string{},
			clientProtocolHeader:     common.XCONF_HTTPS_VALUE,
			recoveryFirmwareVersions: "",
			expectedAllowed:          true,
		},
		{
			name:                     "Secure connection - MTLS",
			contextMap:               map[string]string{},
			clientProtocolHeader:     common.XCONF_MTLS_VALUE,
			recoveryFirmwareVersions: "",
			expectedAllowed:          true,
		},
		{
			name: "Recovery firmware version matches",
			contextMap: map[string]string{
				common.FIRMWARE_VERSION: "PROD_2.0",
				common.MODEL:            "TG1682G",
			},
			clientProtocolHeader:     "HTTP",
			recoveryFirmwareVersions: "PROD_.* TG1682G",
			expectedAllowed:          true,
		},
		{
			name: "Recovery firmware version does not match",
			contextMap: map[string]string{
				common.FIRMWARE_VERSION: "PROD_3.0",
				common.MODEL:            "TG1682G",
			},
			clientProtocolHeader:     "HTTP",
			recoveryFirmwareVersions: "PROD_2.* TG1682G",
			expectedAllowed:          false,
		},
		{
			name: "No firmware version in context",
			contextMap: map[string]string{
				common.MODEL: "TG1682G",
			},
			clientProtocolHeader:     "HTTP",
			recoveryFirmwareVersions: "PROD_.* TG1682G",
			expectedAllowed:          false,
		},
		{
			name: "No model in context",
			contextMap: map[string]string{
				common.FIRMWARE_VERSION: "PROD_2.0",
			},
			clientProtocolHeader:     "HTTP",
			recoveryFirmwareVersions: "PROD_.* TG1682G",
			expectedAllowed:          false,
		},
		{
			name: "Multiple recovery combinations - second matches",
			contextMap: map[string]string{
				common.FIRMWARE_VERSION: "DEV_1.0",
				common.MODEL:            "PX051AEI",
			},
			clientProtocolHeader:     "HTTP",
			recoveryFirmwareVersions: "PROD_.* TG1682G;DEV_.* PX051AEI",
			expectedAllowed:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Xc = &XconfConfigs{
				EstbRecoveryFirmwareVersions: tt.recoveryFirmwareVersions,
			}

			result := IsAllowedRequest(tt.contextMap, tt.clientProtocolHeader)
			assert.Equal(t, tt.expectedAllowed, result)
		})
	}
}

func TestGetMissingAndEmptyQueryParams(t *testing.T) {
	tests := []struct {
		name                  string
		contextMap            map[string]string
		expectedMissingFields []string
		expectedEmptyFields   []string
	}{
		{
			name: "All fields present and filled",
			contextMap: map[string]string{
				common.ESTB_MAC:         "AA:BB:CC:DD:EE:FF",
				common.IP_ADDRESS:       "192.168.1.1",
				common.FIRMWARE_VERSION: "1.0.0",
				common.MODEL:            "Model123",
				common.ENV:              "PROD",
			},
			expectedMissingFields: []string{},
			expectedEmptyFields:   []string{},
		},
		{
			name: "Some fields missing",
			contextMap: map[string]string{
				common.ESTB_MAC:   "AA:BB:CC:DD:EE:FF",
				common.IP_ADDRESS: "192.168.1.1",
			},
			expectedMissingFields: []string{common.FIRMWARE_VERSION, common.MODEL, common.ENV},
			expectedEmptyFields:   []string{},
		},
		{
			name: "Some fields empty",
			contextMap: map[string]string{
				common.ESTB_MAC:         "",
				common.IP_ADDRESS:       "192.168.1.1",
				common.FIRMWARE_VERSION: "",
				common.MODEL:            "Model123",
				common.ENV:              "PROD",
			},
			expectedMissingFields: []string{},
			expectedEmptyFields:   []string{common.ESTB_MAC, common.FIRMWARE_VERSION},
		},
		{
			name:                  "All fields missing",
			contextMap:            map[string]string{},
			expectedMissingFields: []string{common.ESTB_MAC, common.IP_ADDRESS, common.FIRMWARE_VERSION, common.MODEL, common.ENV},
			expectedEmptyFields:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var missingFields []string
			var emptyFields []string

			GetMissingAndEmptyQueryParams(tt.contextMap, &missingFields, &emptyFields)

			assert.ElementsMatch(t, tt.expectedMissingFields, missingFields)
			assert.ElementsMatch(t, tt.expectedEmptyFields, emptyFields)
		})
	}
}

func TestLogPreDisplayCleanup(t *testing.T) {
	tests := []struct {
		name           string
		lastConfigLog  *coreef.ConfigChangeLog
		expectedID     string
		expectedUpdate int64
	}{
		{
			name: "Clean up non-nil log",
			lastConfigLog: &coreef.ConfigChangeLog{
				ID:      "test-id-123",
				Updated: 1234567890,
			},
			expectedID:     "",
			expectedUpdate: 0,
		},
		{
			name:           "Nil log does nothing",
			lastConfigLog:  nil,
			expectedID:     "",
			expectedUpdate: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LogPreDisplayCleanup(tt.lastConfigLog)

			if tt.lastConfigLog != nil {
				assert.Equal(t, tt.expectedID, tt.lastConfigLog.ID)
				assert.Equal(t, tt.expectedUpdate, tt.lastConfigLog.Updated)
			}
		})
	}
}

func TestIsCustomField(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"Base field - ESTB_MAC", common.ESTB_MAC, false},
		{"Base field - ENV", common.ENV, false},
		{"Base field - MODEL", common.MODEL, false},
		{"Base field - FIRMWARE_VERSION", common.FIRMWARE_VERSION, false},
		{"Base field - IP_ADDRESS", common.IP_ADDRESS, false},
		{"Base field - TIME", common.TIME, false},
		{"Base field - CAPABILITIES", common.CAPABILITIES, false},
		{"Custom field", "customField", true},
		{"Another custom field", "myCustomProperty", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCustomField(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAdditionalProperty(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"Base property - ID", common.ID, false},
		{"Base property - UPDATED", common.UPDATED, false},
		{"Base property - FIRMWARE_VERSION", common.FIRMWARE_VERSION, false},
		{"Base property - FIRMWARE_FILENAME", common.FIRMWARE_FILENAME, false},
		{"Base property - FIRMWARE_LOCATION", common.FIRMWARE_LOCATION, false},
		{"Base property - APPLICATION_TYPE", common.APPLICATION_TYPE, false},
		{"Additional property", "additionalProp", true},
		{"Another additional property", "myProperty", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAdditionalProperty(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFirstElementsInContextMap(t *testing.T) {
	tests := []struct {
		name               string
		inputContextMap    map[string]string
		expectedContextMap map[string]string
	}{
		{
			name: "Single values remain unchanged",
			inputContextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
				common.ENV:      "PROD",
				common.MODEL:    "Model123",
			},
			expectedContextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
				common.ENV:      "PROD",
				common.MODEL:    "Model123",
			},
		},
		{
			name: "Comma-separated values take first element",
			inputContextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF,11:22:33:44:55:66",
				common.ENV:      "PROD,DEV,QA",
				common.MODEL:    "Model123,Model456",
			},
			expectedContextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
				common.ENV:      "PROD",
				common.MODEL:    "Model123",
			},
		},
		{
			name: "Mixed single and multi-value fields",
			inputContextMap: map[string]string{
				common.ESTB_MAC:         "AA:BB:CC:DD:EE:FF",
				common.ENV:              "PROD,DEV",
				common.FIRMWARE_VERSION: "1.0.0",
				common.PARTNER_ID:       "partner1,partner2,partner3",
			},
			expectedContextMap: map[string]string{
				common.ESTB_MAC:         "AA:BB:CC:DD:EE:FF",
				common.ENV:              "PROD",
				common.FIRMWARE_VERSION: "1.0.0",
				common.PARTNER_ID:       "partner1",
			},
		},
		{
			name: "Empty values remain empty",
			inputContextMap: map[string]string{
				common.ESTB_MAC: "",
				common.ENV:      "",
			},
			expectedContextMap: map[string]string{
				common.ESTB_MAC: "",
				common.ENV:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetFirstElementsInContextMap(tt.inputContextMap)

			for key, expectedValue := range tt.expectedContextMap {
				assert.Equal(t, expectedValue, tt.inputContextMap[key])
			}
		})
	}
}

func TestDoSplunkLog(t *testing.T) {
	tests := []struct {
		name             string
		contextMap       map[string]string
		evaluationResult *estbfirmware.EvaluationResult
		setupFields      log.Fields
	}{
		{
			name: "Basic context logging",
			contextMap: map[string]string{
				common.ESTB_MAC:         "AA:BB:CC:DD:EE:FF",
				common.ENV:              "PROD",
				common.MODEL:            "Model123",
				common.FIRMWARE_VERSION: "1.0.0",
				common.IP_ADDRESS:       "192.168.1.1",
			},
			evaluationResult: nil,
			setupFields:      log.Fields{},
		},
		{
			name: "With evaluation result and matched rule",
			contextMap: map[string]string{
				common.ESTB_MAC:         "AA:BB:CC:DD:EE:FF",
				common.ENV:              "PROD",
				common.MODEL:            "Model123",
				common.FIRMWARE_VERSION: "1.0.0",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "Test Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: nil,
				Blocked:        false,
			},
			setupFields: log.Fields{},
		},
		{
			name: "With firmware config",
			contextMap: map[string]string{
				common.ESTB_MAC:         "AA:BB:CC:DD:EE:FF",
				common.FIRMWARE_VERSION: "1.0.0",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION:           "2.0.0",
						common.FIRMWARE_FILENAME:          "firmware.bin",
						common.FIRMWARE_DOWNLOAD_PROTOCOL: "http",
						common.FIRMWARE_LOCATION:          "http://example.com",
						common.REBOOT_IMMEDIATELY:         false,
					},
				},
			},
			setupFields: log.Fields{},
		},
		{
			name: "Blocked result",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				Blocked: true,
			},
			setupFields: log.Fields{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function logs to splunk/logrus, so we just ensure it doesn't panic
			assert.NotPanics(t, func() {
				DoSplunkLog(tt.contextMap, tt.evaluationResult, tt.setupFields)
			})
		})
	}
}

func TestGetExplanation(t *testing.T) {
	tests := []struct {
		name             string
		contextMap       map[string]string
		evaluationResult *estbfirmware.EvaluationResult
		shouldContain    []string
	}{
		{
			name: "No matched rule",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
				common.MODEL:    "Model123",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: nil,
			},
			shouldContain: []string{"did not match any rule"},
		},
		{
			name: "Matched rule with config",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "Test Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION: "2.0.0",
					},
				},
			},
			shouldContain: []string{"matched", "Test Rule", "received config"},
		},
		{
			name: "Blocked by distribution percent",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:               "rule-123",
					Name:             "Percent Rule",
					Type:             firmware.ENV_MODEL_RULE,
					ApplicableAction: &firmware.ApplicableAction{},
				},
				FirmwareConfig: nil,
				Blocked:        true,
			},
			shouldContain: []string{"matched", "blocked by Distribution percent"},
		},
		{
			name: "NO OP rule",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "NO OP Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: nil,
				Blocked:        false,
			},
			shouldContain: []string{"matched NO OP", "received NO config"},
		},
		{
			name: "With TIME_FILTER applied",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "Time Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION: "2.0.0",
					},
				},
				AppliedFilters: []interface{}{
					&firmware.FirmwareRule{
						ID:   "time-filter-1",
						Name: "Time Filter",
						Type: firmware.TIME_FILTER,
					},
				},
			},
			shouldContain: []string{"matched", "received config", "was blocked/modified by filter", "TIME_FILTER"},
		},
		{
			name: "With IP_FILTER applied",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "IP Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION: "2.0.0",
					},
				},
				AppliedFilters: []interface{}{
					&firmware.FirmwareRule{
						ID:   "ip-filter-1",
						Name: "IP Filter",
						Type: firmware.IP_FILTER,
					},
				},
			},
			shouldContain: []string{"matched", "IP_FILTER"},
		},
		{
			name: "With DOWNLOAD_LOCATION_FILTER applied",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "Download Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION: "2.0.0",
					},
				},
				AppliedFilters: []interface{}{
					&firmware.FirmwareRule{
						ID:   "dl-filter-1",
						Name: "Download Location Filter",
						Type: firmware.DOWNLOAD_LOCATION_FILTER,
					},
				},
			},
			shouldContain: []string{"matched", "received config"},
		},
		{
			name: "With PercentFilterValue applied",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "Percent Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION: "2.0.0",
					},
				},
				AppliedFilters: []interface{}{
					coreef.PercentFilterValue{
						Percentage:          50.0,
						EnvModelPercentages: map[string]coreef.EnvModelPercentage{},
					},
				},
			},
			shouldContain: []string{"matched", "percent=50"},
		},
		{
			name: "With DownloadLocationRoundRobinFilterValue applied",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "RoundRobin Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION: "2.0.0",
					},
				},
				AppliedFilters: []interface{}{
					coreef.DownloadLocationRoundRobinFilterValue{
						ID: "roundrobin_VALUE",
					},
				},
			},
			shouldContain: []string{"matched", "SINGLETON"},
		},
		{
			name: "With RuleAction applied",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "Action Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION: "2.0.0",
					},
				},
				AppliedFilters: []interface{}{
					firmware.RuleAction{},
				},
			},
			shouldContain: []string{"matched", "DistributionPercent"},
		},
		{
			name: "With PercentageBean applied",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "Percentage Bean Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION: "2.0.0",
					},
				},
				AppliedFilters: []interface{}{
					coreef.PercentageBean{
						ID:   "bean-123",
						Name: "Test Bean",
					},
				},
			},
			shouldContain: []string{"matched", "DistributedEnvModelPercentage"},
		},
		{
			name: "With generic FirmwareRule filter",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			evaluationResult: &estbfirmware.EvaluationResult{
				MatchedRule: &firmware.FirmwareRule{
					ID:   "rule-123",
					Name: "Generic Rule",
					Type: firmware.ENV_MODEL_RULE,
				},
				FirmwareConfig: &coreef.FirmwareConfigFacade{
					Properties: map[string]interface{}{
						common.FIRMWARE_VERSION: "2.0.0",
					},
				},
				AppliedFilters: []interface{}{
					&firmware.FirmwareRule{
						ID:   "generic-filter-1",
						Name: "Generic Filter",
						Type: "CUSTOM_FILTER",
					},
				},
			},
			shouldContain: []string{"matched", "FirmwareRule{id=generic-filter-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			explanation := GetExplanation(tt.contextMap, tt.evaluationResult)

			assert.NotEmpty(t, explanation)
			for _, expected := range tt.shouldContain {
				assert.Contains(t, explanation, expected)
			}
		})
	}
}

func TestNormalizeEstbFirmwareContext(t *testing.T) {
	tests := []struct {
		name              string
		contextMap        map[string]string
		usePartnerAppType bool
		shouldAddIp       bool
		expectedChanges   map[string]bool // which fields should be modified
	}{
		{
			name: "Normalize MAC addresses",
			contextMap: map[string]string{
				common.ESTB_MAC: "aa:bb:cc:dd:ee:ff",
				common.ECM_MAC:  "11:22:33:44:55:66",
			},
			usePartnerAppType: false,
			shouldAddIp:       false,
			expectedChanges: map[string]bool{
				common.TIME: true,
			},
		},
		{
			name: "Add time if not present",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
			},
			usePartnerAppType: false,
			shouldAddIp:       false,
			expectedChanges: map[string]bool{
				common.TIME: true,
			},
		},
		{
			name: "Time already present",
			contextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
				common.TIME:     "2025-10-27 12:00:00",
			},
			usePartnerAppType: false,
			shouldAddIp:       false,
			expectedChanges: map[string]bool{
				common.TIME: false, // should not change
			},
		},
		{
			name: "Bypass filters includes percent filter",
			contextMap: map[string]string{
				common.ESTB_MAC:       "AA:BB:CC:DD:EE:FF",
				common.BYPASS_FILTERS: estbfirmware.PERCENT_FILTER_NAME,
			},
			usePartnerAppType: false,
			shouldAddIp:       false,
			expectedChanges: map[string]bool{
				common.BYPASS_FILTERS: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			fields := log.Fields{}

			// Store original values
			originalValues := make(map[string]string)
			for k, v := range tt.contextMap {
				originalValues[k] = v
			}

			NormalizeEstbFirmwareContext(nil, req, tt.contextMap, tt.usePartnerAppType, tt.shouldAddIp, fields)

			// Check expected changes
			for field, shouldChange := range tt.expectedChanges {
				if shouldChange {
					// Field should be different or added
					if originalValue, exists := originalValues[field]; exists {
						if field == common.TIME && originalValue != "" {
							// Time was already present, should remain unchanged
							continue
						}
					}
					// For fields that should be added/changed
					if field == common.TIME {
						assert.NotEmpty(t, tt.contextMap[field], "Time should be set")
					}
					if field == common.BYPASS_FILTERS && strings.Contains(originalValues[field], estbfirmware.PERCENT_FILTER_NAME) {
						assert.Contains(t, tt.contextMap[field], firmware.GLOBAL_PERCENT)
					}
				}
			}
		})
	}
}
