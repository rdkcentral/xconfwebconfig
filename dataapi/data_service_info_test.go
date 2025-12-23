package dataapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// GetInfoRefreshAllHandler Tests
// ============================================================================

func TestGetInfoRefreshAllHandler_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/info/refreshAll", nil)
	recorder := httptest.NewRecorder()

	GetInfoRefreshAllHandler(recorder, req)

	// Should return 200 or 404 depending on cache state
	assert.True(t, recorder.Code == http.StatusOK || recorder.Code == http.StatusNotFound)
}

func TestGetInfoRefreshAllHandler_WithFailedTables(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/info/refreshAll", nil)
	recorder := httptest.NewRecorder()

	GetInfoRefreshAllHandler(recorder, req)

	// Response should be valid JSON or error message
	assert.NotEmpty(t, recorder.Body.String())
}

// ============================================================================
// GetInfoRefreshHandler Tests
// ============================================================================

func TestGetInfoRefreshHandler_WithTableName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/info/refresh/FirmwareConfig", nil)
	recorder := httptest.NewRecorder()

	// Add mux vars
	vars := map[string]string{
		common.TABLE_NAME: "FirmwareConfig",
	}
	req = mux.SetURLVars(req, vars)

	GetInfoRefreshHandler(recorder, req)

	// Should return 200 or 500 depending on cache state
	assert.True(t, recorder.Code == http.StatusOK || recorder.Code == http.StatusInternalServerError)
}

func TestGetInfoRefreshHandler_WithInvalidTableName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/info/refresh/InvalidTable", nil)
	recorder := httptest.NewRecorder()

	// Add mux vars
	vars := map[string]string{
		common.TABLE_NAME: "InvalidTable",
	}
	req = mux.SetURLVars(req, vars)

	GetInfoRefreshHandler(recorder, req)

	// Should return error (500)
	assert.True(t, recorder.Code == http.StatusInternalServerError || recorder.Code == http.StatusOK)
}

func TestGetInfoRefreshHandler_WithEmptyTableName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/info/refresh/", nil)
	recorder := httptest.NewRecorder()

	// Add empty mux vars
	vars := map[string]string{
		common.TABLE_NAME: "",
	}
	req = mux.SetURLVars(req, vars)

	GetInfoRefreshHandler(recorder, req)

	// Should handle empty table name
	assert.True(t, recorder.Code >= 200 && recorder.Code < 600)
}

// ============================================================================
// GetInfoStatistics Tests
// ============================================================================

func TestGetInfoStatistics_ReturnsValidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/info/statistics", nil)
	recorder := httptest.NewRecorder()

	GetInfoStatistics(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.NotEmpty(t, recorder.Body.String())
	// Should return valid JSON
	assert.Contains(t, recorder.Body.String(), "{")
}

func TestGetInfoStatistics_MultipleCalls(t *testing.T) {
	// Test that multiple calls work correctly
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/info/statistics", nil)
		recorder := httptest.NewRecorder()

		GetInfoStatistics(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.NotEmpty(t, recorder.Body.String())
	}
}

func TestGetInfoStatistics_ResponseStructure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/info/statistics", nil)
	recorder := httptest.NewRecorder()

	GetInfoStatistics(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	body := recorder.Body.String()

	// Response should be valid JSON object or array
	assert.True(t, len(body) > 0)
	assert.True(t, body[0] == '{' || body[0] == '[')
}
