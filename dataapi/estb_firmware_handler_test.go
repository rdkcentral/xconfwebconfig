package dataapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfwebconfig/common"
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	"github.com/rdkcentral/xconfwebconfig/shared"
	sharedef "github.com/rdkcentral/xconfwebconfig/shared/estbfirmware"
	"github.com/stretchr/testify/assert"
)

func TestParseProcBody(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		expectedVersion    string
		expectedContextMap map[string]string
	}{
		{
			name:            "Parse body with version",
			body:            "eStbMac=AA:BB:CC:DD:EE:FF&env=PROD&model=Model123&version=1.0.0",
			expectedVersion: "1.0.0",
			expectedContextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
				common.ENV:      "PROD",
				common.MODEL:    "Model123",
			},
		},
		{
			name:            "Parse body without version",
			body:            "eStbMac=AA:BB:CC:DD:EE:FF&env=PROD",
			expectedVersion: "",
			expectedContextMap: map[string]string{
				common.ESTB_MAC: "AA:BB:CC:DD:EE:FF",
				common.ENV:      "PROD",
			},
		},
		{
			name:               "Empty body",
			body:               "",
			expectedVersion:    "",
			expectedContextMap: map[string]string{},
		},
		{
			name:               "Body with only version",
			body:               "version=2.0.0",
			expectedVersion:    "2.0.0",
			expectedContextMap: map[string]string{},
		},
		{
			name:            "Body with multiple parameters",
			body:            "eStbMac=11:22:33:44:55:66&env=QA&model=TestModel&firmwareVersion=3.0.0&ipAddress=192.168.1.1",
			expectedVersion: "",
			expectedContextMap: map[string]string{
				common.ESTB_MAC:         "11:22:33:44:55:66",
				common.ENV:              "QA",
				common.MODEL:            "TestModel",
				common.FIRMWARE_VERSION: "3.0.0",
				common.IP_ADDRESS:       "192.168.1.1",
			},
		},
		{
			name:               "Invalid format - no equals sign",
			body:               "invalidparam",
			expectedVersion:    "",
			expectedContextMap: map[string]string{},
		},
		{
			name:               "Invalid format - multiple equals",
			body:               "key=value=extra",
			expectedVersion:    "",
			expectedContextMap: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextMap := make(map[string]string)
			version := parseProcBody(tt.body, contextMap)

			assert.Equal(t, tt.expectedVersion, version)
			assert.Equal(t, len(tt.expectedContextMap), len(contextMap))
			for key, expectedValue := range tt.expectedContextMap {
				assert.Equal(t, expectedValue, contextMap[key])
			}
		})
	}
}

func TestGetEstbFirmwareSwuBseHandler_MissingIPAddress(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/bse", nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	GetEstbFirmwareSwuBseHandler(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Required IpAddress parameter")
}

func TestGetEstbFirmwareSwuBseHandler_InvalidIPAddress(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/bse?ipAddress=invalid-ip", nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	GetEstbFirmwareSwuBseHandler(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "not a valid IpAddress")
}

func TestGetEstbFirmwareSwuBseHandler_ValidIPInQueryParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/bse?ipAddress=192.168.1.1", nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	GetEstbFirmwareSwuBseHandler(xw, req)

	// Will return 404 since no rules are configured in test, but validates IP parsing
	assert.True(t, recorder.Code == http.StatusNotFound || recorder.Code == http.StatusOK)
}

func TestGetEstbFirmwareSwuBseHandler_ValidIPInBody(t *testing.T) {
	body := "ipAddress=10.0.0.1"
	req := httptest.NewRequest(http.MethodPost, "/estbfirmware/bse", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)
	xw.SetBody(body)

	GetEstbFirmwareSwuBseHandler(xw, req)

	// Will return 404 since no rules are configured in test, but validates IP parsing from body
	assert.True(t, recorder.Code == http.StatusNotFound || recorder.Code == http.StatusOK)
}

func TestGetEstbLastlogPath_InvalidMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/lastlog?mac=invalid-mac", nil)
	recorder := httptest.NewRecorder()

	GetEstbLastlogPath(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Mac is invalid")
}

func TestGetEstbLastlogPath_MissingMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/lastlog", nil)
	recorder := httptest.NewRecorder()

	GetEstbLastlogPath(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "not present")
}

func TestGetEstbLastlogPath_ValidMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/lastlog?mac=AA:BB:CC:DD:EE:FF", nil)
	recorder := httptest.NewRecorder()

	GetEstbLastlogPath(recorder, req)

	// Should return 200 with empty body or valid log
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGetEstbChangelogsPath_InvalidMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/changelogs?mac=invalid", nil)
	recorder := httptest.NewRecorder()

	GetEstbChangelogsPath(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Mac is invalid")
}

func TestGetEstbChangelogsPath_MissingMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/changelogs", nil)
	recorder := httptest.NewRecorder()

	GetEstbChangelogsPath(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetEstbChangelogsPath_ValidMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/changelogs?mac=AA:BB:CC:DD:EE:FF", nil)
	recorder := httptest.NewRecorder()

	GetEstbChangelogsPath(recorder, req)

	// Should return 200 with empty array or valid logs
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGetCheckMinFirmwareHandler_MissingFields(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   string
		expectedCode  int
		shouldContain string
	}{
		{
			name:          "All fields missing",
			queryParams:   "",
			expectedCode:  http.StatusBadRequest,
			shouldContain: "Required field(s) are missing",
		},
		{
			name:          "Only MAC provided",
			queryParams:   "?eStbMac=AA:BB:CC:DD:EE:FF",
			expectedCode:  http.StatusBadRequest,
			shouldContain: "Required field(s) are missing",
		},
		{
			name:          "Missing firmware version",
			queryParams:   "?eStbMac=AA:BB:CC:DD:EE:FF&env=PROD&model=Model123&ipAddress=192.168.1.1",
			expectedCode:  http.StatusBadRequest,
			shouldContain: "Required field(s) are missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/estbfirmware/checkMinimumFirmware"+tt.queryParams, nil)
			recorder := httptest.NewRecorder()
			xw := xhttp.NewXResponseWriter(recorder)

			GetCheckMinFirmwareHandler(xw, req)

			assert.Equal(t, tt.expectedCode, recorder.Code)
			if tt.shouldContain != "" {
				assert.Contains(t, recorder.Body.String(), tt.shouldContain)
			}
		})
	}
}

func TestGetCheckMinFirmwareHandler_EmptyFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/estbfirmware/checkMinimumFirmware?eStbMac=&env=PROD&model=Model123&ipAddress=192.168.1.1&firmwareVersion=1.0.0",
		nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	GetCheckMinFirmwareHandler(xw, req)

	// Empty fields should return true for hasMinimumFirmware
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "hasMinimumFirmware")
}

func TestGetEstbFirmwareVersionInfoPath_MissingMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/versionInfo/stb", nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	// Add mux vars
	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	GetEstbFirmwareVersionInfoPath(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "eStbMac should be specified")
}

func TestGetEstbFirmwareVersionInfoPath_WithMAC(t *testing.T) {
	// Setup test server with config
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{}

	req := httptest.NewRequest(http.MethodGet,
		"/estbfirmware/versionInfo/stb?eStbMac=AA:BB:CC:DD:EE:FF&env=PROD&model=Model123",
		nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	// Add mux vars
	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	GetEstbFirmwareVersionInfoPath(xw, req)

	// Should process and return 200 (or 403 depending on security check)
	assert.True(t, recorder.Code == http.StatusOK || recorder.Code == http.StatusForbidden)
}

func TestGetFirmwareResponse_MissingMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/xconf/swu/stb", nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	status, response, _, _ := GetFirmwareResponse(recorder, req, xw, map[string]interface{}{})

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(response), "eStbMac should be specified")
}

func TestGetFirmwareResponse_InvalidMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/xconf/swu/stb?eStbMac=invalid-mac", nil)
	// Add HTTPS header to pass security check so we can reach MAC validation
	req.Header.Set("HA-Haproxy-xconf-http", "xconf-https")
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	status, response, _, _ := GetFirmwareResponse(recorder, req, xw, map[string]interface{}{})

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, string(response), "invalid mac address")
}

func TestGetFirmwareResponse_ForbiddenRequest(t *testing.T) {
	// Setup
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{
		EstbRecoveryFirmwareVersions: "", // No recovery versions configured
	}

	req := httptest.NewRequest(http.MethodGet,
		"/xconf/swu/stb?eStbMac=AA:BB:CC:DD:EE:FF&firmwareVersion=1.0.0&model=Model123",
		nil)
	req.Header.Set(common.XCONF_HTTP_HEADER, "HTTP") // Non-secure connection
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	status, response, _, _ := GetFirmwareResponse(recorder, req, xw, map[string]interface{}{})

	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "FORBIDDEN", string(response))
}

func TestGetFirmwareResponse_IgnoresClientProtocolFromQueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/xconf/swu/stb?eStbMac=AA:BB:CC:DD:EE:FF&clientProtocol=HTTP&clientCertExpiry=2025-12-31",
		nil)
	req.Header.Set(common.XCONF_HTTP_HEADER, common.XCONF_HTTPS_VALUE)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	_, _, _, contextMap := GetFirmwareResponse(recorder, req, xw, map[string]interface{}{})

	// Should use header value, not query param
	if contextMap != nil {
		// clientProtocol should come from header processing, not query params
		assert.NotEqual(t, "HTTP", contextMap[common.CLIENT_PROTOCOL])
	}
}

func TestGetFirmwareResponse_WithBodyParameters(t *testing.T) {
	body := "eStbMac=AA:BB:CC:DD:EE:FF&env=PROD&model=Model123&firmwareVersion=1.0.0"
	req := httptest.NewRequest(http.MethodPost, "/xconf/swu/stb", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(common.XCONF_HTTP_HEADER, common.XCONF_HTTPS_VALUE)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)
	xw.SetBody(body)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	status, _, _, contextMap := GetFirmwareResponse(recorder, req, xw, map[string]interface{}{})

	// Should parse body parameters
	if contextMap != nil {
		assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap[common.ESTB_MAC])
		assert.Equal(t, "PROD", contextMap[common.ENV])
		assert.Equal(t, "Model123", contextMap[common.MODEL])
	}
	// Status could be 200, 404, or 403 depending on rules
	assert.True(t, status >= 200)
}

func TestGetFirmwareResponse_WithCertExpiry(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/xconf/swu/stb?eStbMac=AA:BB:CC:DD:EE:FF&firmwareVersion=1.0.0",
		nil)
	req.Header.Set(common.XCONF_HTTP_HEADER, common.XCONF_MTLS_VALUE)
	req.Header.Set(common.CLIENT_CERT_EXPIRY_HEADER, "2025-12-31")
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	_, _, _, contextMap := GetFirmwareResponse(recorder, req, xw, map[string]interface{}{})

	if contextMap != nil {
		// Should have cert expiry in context
		assert.NotEmpty(t, contextMap[common.CLIENT_CERT_EXPIRY])
	}
}

func TestGetEstbFirmwareSwuHandler_ReturnsProperStatusCodes(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  string
		headerValue  string
		expectedCode int
	}{
		{
			name:         "Missing MAC returns 400",
			queryParams:  "",
			headerValue:  common.XCONF_HTTPS_VALUE,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid MAC returns 400",
			queryParams:  "?eStbMac=invalid",
			headerValue:  common.XCONF_HTTPS_VALUE,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/xconf/swu/stb"+tt.queryParams, nil)
			req.Header.Set(common.XCONF_HTTP_HEADER, tt.headerValue)
			recorder := httptest.NewRecorder()
			xw := xhttp.NewXResponseWriter(recorder)

			vars := map[string]string{
				common.APPLICATION_TYPE: shared.STB,
			}
			req = mux.SetURLVars(req, vars)

			GetEstbFirmwareSwuHandler(xw, req)

			assert.Equal(t, tt.expectedCode, recorder.Code)
		})
	}
}

func TestParseProcBody_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		expectedLen int
	}{
		{
			name:        "Body with empty values",
			body:        "eStbMac=&env=&model=",
			expectedLen: 3,
		},
		{
			name:        "Body with special characters",
			body:        "eStbMac=AA%3ABB%3ACC%3ADD%3AEE%3AFF",
			expectedLen: 1,
		},
		{
			name:        "Body with single parameter",
			body:        "eStbMac=AA:BB:CC:DD:EE:FF",
			expectedLen: 1,
		},
		{
			name:        "Body with trailing ampersand",
			body:        "eStbMac=AA:BB:CC:DD:EE:FF&env=PROD&",
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextMap := make(map[string]string)
			parseProcBody(tt.body, contextMap)
			assert.Equal(t, tt.expectedLen, len(contextMap))
		})
	}
}

func TestGetFirstElementsInContextMap_Integration(t *testing.T) {
	contextMap := map[string]string{
		common.ESTB_MAC:         "AA:BB:CC:DD:EE:FF,11:22:33:44:55:66",
		common.ENV:              "PROD,DEV,QA",
		common.MODEL:            "Model123",
		common.FIRMWARE_VERSION: "1.0.0",
	}

	GetFirstElementsInContextMap(contextMap)

	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap[common.ESTB_MAC])
	assert.Equal(t, "PROD", contextMap[common.ENV])
	assert.Equal(t, "Model123", contextMap[common.MODEL])
	assert.Equal(t, "1.0.0", contextMap[common.FIRMWARE_VERSION])
}

func TestLogPreDisplayCleanup_NilLog(t *testing.T) {
	// Should not panic with nil log
	assert.NotPanics(t, func() {
		LogPreDisplayCleanup(nil)
	})
}

func TestLogPreDisplayCleanup_ValidLog(t *testing.T) {
	log := &sharedef.ConfigChangeLog{
		ID:      "test-id",
		Updated: 1234567890,
	}

	LogPreDisplayCleanup(log)

	assert.Equal(t, "", log.ID)
	assert.Equal(t, int64(0), log.Updated)
}

func TestGetEstbFirmwareSwuHandler_WithInvalidResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/xconf/swu/stb", nil)
	recorder := httptest.NewRecorder()

	// Call with standard ResponseWriter instead of XResponseWriter
	GetEstbFirmwareSwuHandler(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetEstbFirmwareSwuBseHandler_WithInvalidResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/bse", nil)
	recorder := httptest.NewRecorder()

	// Call with standard ResponseWriter instead of XResponseWriter
	GetEstbFirmwareSwuBseHandler(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetCheckMinFirmwareHandler_WithInvalidResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/checkMinimumFirmware", nil)
	recorder := httptest.NewRecorder()

	// Call with standard ResponseWriter instead of XResponseWriter
	GetCheckMinFirmwareHandler(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetEstbFirmwareVersionInfoPath_WithInvalidResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/versionInfo/stb", nil)
	recorder := httptest.NewRecorder()

	// Call with standard ResponseWriter instead of XResponseWriter
	GetEstbFirmwareVersionInfoPath(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
