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

// parseProcBody Tests - 100% Coverage
func TestParseProcBody_WithVersion(t *testing.T) {
	contextMap := make(map[string]string)
	version := parseProcBody("eStbMac=AA:BB:CC:DD:EE:FF&env=PROD&version=1.0.0", contextMap)

	assert.Equal(t, "1.0.0", version)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap[common.ESTB_MAC])
	assert.Equal(t, "PROD", contextMap[common.ENV])
}

func TestParseProcBody_WithoutVersion(t *testing.T) {
	contextMap := make(map[string]string)
	version := parseProcBody("eStbMac=AA:BB:CC:DD:EE:FF&env=PROD", contextMap)

	assert.Equal(t, "", version)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", contextMap[common.ESTB_MAC])
}

func TestParseProcBody_EmptyBody(t *testing.T) {
	contextMap := make(map[string]string)
	version := parseProcBody("", contextMap)

	assert.Equal(t, "", version)
	assert.Equal(t, 0, len(contextMap))
}

func TestParseProcBody_InvalidFormat(t *testing.T) {
	contextMap := make(map[string]string)
	version := parseProcBody("invalidparam&key=value=extra", contextMap)

	assert.Equal(t, "", version)
	assert.Equal(t, 0, len(contextMap))
}

func TestParseProcBody_MultipleParameters(t *testing.T) {
	contextMap := make(map[string]string)
	parseProcBody("eStbMac=11:22:33:44:55:66&env=QA&model=TestModel&firmwareVersion=3.0.0", contextMap)

	assert.Equal(t, "11:22:33:44:55:66", contextMap[common.ESTB_MAC])
	assert.Equal(t, "QA", contextMap[common.ENV])
	assert.Equal(t, "TestModel", contextMap[common.MODEL])
}

// GetEstbFirmwareSwuBseHandler Tests
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

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetEstbFirmwareSwuBseHandler_ValidIPInBody(t *testing.T) {
	body := "ipAddress=10.0.0.1"
	req := httptest.NewRequest(http.MethodPost, "/estbfirmware/bse", strings.NewReader(body))
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)
	xw.SetBody(body)

	GetEstbFirmwareSwuBseHandler(xw, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetEstbFirmwareSwuBseHandler_QueryParamTakesPrecedence(t *testing.T) {
	body := "ipAddress=10.0.0.1"
	req := httptest.NewRequest(http.MethodPost, "/estbfirmware/bse?ipAddress=192.168.1.100", strings.NewReader(body))
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)
	xw.SetBody(body)

	GetEstbFirmwareSwuBseHandler(xw, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetEstbFirmwareSwuBseHandler_WithInvalidResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/bse", nil)
	recorder := httptest.NewRecorder()

	GetEstbFirmwareSwuBseHandler(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// GetEstbLastlogPath Tests
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
}

func TestGetEstbLastlogPath_ValidMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/lastlog?mac=AA:BB:CC:DD:EE:FF", nil)
	recorder := httptest.NewRecorder()

	GetEstbLastlogPath(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

// GetEstbChangelogsPath Tests
func TestGetEstbChangelogsPath_InvalidMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/changelogs?mac=invalid", nil)
	recorder := httptest.NewRecorder()

	GetEstbChangelogsPath(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
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

	assert.Equal(t, http.StatusOK, recorder.Code)
}

// GetCheckMinFirmwareHandler Tests
func TestGetCheckMinFirmwareHandler_MissingFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/checkMinimumFirmware", nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	GetCheckMinFirmwareHandler(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Required field(s) are missing")
}

func TestGetCheckMinFirmwareHandler_EmptyFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/estbfirmware/checkMinimumFirmware?eStbMac=&env=PROD&model=Model123&ipAddress=192.168.1.1&firmwareVersion=1.0.0",
		nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	GetCheckMinFirmwareHandler(xw, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "true")
}

func TestGetCheckMinFirmwareHandler_WithInvalidResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/checkMinimumFirmware", nil)
	recorder := httptest.NewRecorder()

	GetCheckMinFirmwareHandler(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// GetEstbFirmwareVersionInfoPath Tests
func TestGetEstbFirmwareVersionInfoPath_MissingMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/versionInfo/stb", nil)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	GetEstbFirmwareVersionInfoPath(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "eStbMac should be specified")
}

func TestGetEstbFirmwareVersionInfoPath_ForbiddenRequest(t *testing.T) {
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{}

	req := httptest.NewRequest(http.MethodGet,
		"/estbfirmware/versionInfo/stb?eStbMac=AA:BB:CC:DD:EE:FF",
		nil)
	req.Header.Set(common.XCONF_HTTP_HEADER, "HTTP")
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	GetEstbFirmwareVersionInfoPath(xw, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestGetEstbFirmwareVersionInfoPath_WithInvalidResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/estbfirmware/versionInfo/stb", nil)
	recorder := httptest.NewRecorder()

	GetEstbFirmwareVersionInfoPath(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// GetFirmwareResponse Tests
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
	originalXc := Xc
	defer func() { Xc = originalXc }()
	Xc = &XconfConfigs{}

	req := httptest.NewRequest(http.MethodGet,
		"/xconf/swu/stb?eStbMac=AA:BB:CC:DD:EE:FF",
		nil)
	req.Header.Set(common.XCONF_HTTP_HEADER, "HTTP")
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

// GetEstbFirmwareSwuHandler Tests
func TestGetEstbFirmwareSwuHandler_MissingMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/xconf/swu/stb", nil)
	req.Header.Set(common.XCONF_HTTP_HEADER, common.XCONF_HTTPS_VALUE)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	GetEstbFirmwareSwuHandler(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetEstbFirmwareSwuHandler_InvalidMAC(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/xconf/swu/stb?eStbMac=invalid", nil)
	req.Header.Set(common.XCONF_HTTP_HEADER, common.XCONF_HTTPS_VALUE)
	recorder := httptest.NewRecorder()
	xw := xhttp.NewXResponseWriter(recorder)

	vars := map[string]string{
		common.APPLICATION_TYPE: shared.STB,
	}
	req = mux.SetURLVars(req, vars)

	GetEstbFirmwareSwuHandler(xw, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetEstbFirmwareSwuHandler_WithInvalidResponseWriter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/xconf/swu/stb", nil)
	recorder := httptest.NewRecorder()

	GetEstbFirmwareSwuHandler(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// Helper Function Tests
func TestLogPreDisplayCleanup_NilLog(t *testing.T) {
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
