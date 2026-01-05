package dataapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	conversion "github.com/rdkcentral/xconfwebconfig/protobuf"
	"github.com/stretchr/testify/assert"
)

func TestGetTestConfig(t *testing.T) {
	result := GetTestConfig()
	assert.Equal(t, "../config/sample_xconfwebconfig.conf", result)
}

func TestExecuteRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := ExecuteRequest(req, handler)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "test response", recorder.Body.String())
}

func TestExecuteRequest_WithDifferentMethods(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{"GET request", http.MethodGet, http.StatusOK},
		{"POST request", http.MethodPost, http.StatusCreated},
		{"PUT request", http.MethodPut, http.StatusAccepted},
		{"DELETE request", http.MethodDelete, http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.expectedStatus)
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			recorder := ExecuteRequest(req, handler)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestSetupGroupServiceMockServerOkResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	cpeGroup := &conversion.CpeGroup{
		StormReadyFw: true,
		Wanfailover:  false,
		Gwfailover:   false,
	}

	path := "/testPath"
	mockServer := SetupGroupServiceMockServerOkResponse(t, *server, path, cpeGroup)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.GroupServiceHost())

	// Test the mock server responds correctly
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupGroupServiceMockServer500Response(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	mockServer := SetupGroupServiceMockServer500Response(t, *server)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.GroupServiceHost())

	// Test the mock server responds with 500
	resp, err := http.Get(mockServer.URL + "/anyPath")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestSetupSatServiceMockServerOkResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	mockServer := SetupSatServiceMockServerOkResponse(t, *server)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.SatServiceHost())

	// Test the mock server responds correctly
	resp, err := http.Get(mockServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupSatServiceMockServerErrorResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	mockServer := SetupSatServiceMockServerErrorResponse(t, *server)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.SatServiceHost())

	// Test the mock server responds with error
	resp, err := http.Get(mockServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSetupDeviceServiceMockServerOkResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	path := "/device/test"
	mockServer := SetupDeviceServiceMockServerOkResponse(t, *server, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.DeviceServiceHost())

	// Test the mock server responds correctly
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupDeviceServiceMockServerOkResponseDynamic(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	customResponse := []byte(`{"custom":"response"}`)
	path := "/device/custom"
	mockServer := SetupDeviceServiceMockServerOkResponseDynamic(t, *server, customResponse, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.DeviceServiceHost())

	// Test the mock server responds with custom response
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupDeviceServiceMockServerErrorResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	path := "/device/error"
	mockServer := SetupDeviceServiceMockServerErrorResponse(t, *server, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.DeviceServiceHost())

	// Test the mock server responds with error
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSetupAccountServiceMockServerOkResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	path := "/account/test"
	mockServer := SetupAccountServiceMockServerOkResponse(t, *server, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.AccountServiceHost())

	// Test the mock server responds correctly
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupAccountServiceMockServerOkResponseDynamic(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	customResponse := []byte(`[{"custom":"account"}]`)
	path := "/account/custom"
	mockServer := SetupAccountServiceMockServerOkResponseDynamic(t, *server, customResponse, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.AccountServiceHost())
}

func TestSetupAccountServiceMockServerOkResponseDynamicTwoCalls(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	response1 := []byte(`[{"first":"account"}]`)
	response2 := []byte(`[{"second":"account"}]`)
	path1 := "/account/first"
	path2 := "/account/second"

	mockServer := SetupAccountServiceMockServerOkResponseDynamicTwoCalls(t, *server, response1, response2, path1, path2)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.AccountServiceHost())

	// Test first path
	resp1, err1 := http.Get(mockServer.URL + path1)
	assert.NoError(t, err1)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	// Test second path
	resp2, err2 := http.Get(mockServer.URL + path2)
	assert.NoError(t, err2)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

func TestSetupAccountServiceMockServerEmptyResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	path := "/account/empty"
	mockServer := SetupAccountServiceMockServerEmptyResponse(t, *server, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.AccountServiceHost())

	// Test the mock server responds with empty array
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupAccountServiceMockServerErrorResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	path := "/account/error"
	mockServer := SetupAccountServiceMockServerErrorResponse(t, *server, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.AccountServiceHost())

	// Test the mock server responds with error
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSetupTaggingMockServerOkResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	path := "/tagging/test"
	mockServer := SetupTaggingMockServerOkResponse(t, *server, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.TaggingHost())

	// Test the mock server responds correctly
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupTaggingMockServerOkResponseDynamic(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	customResponse := `["tag1", "tag2"]`
	path := "/tagging/custom"
	mockServer := SetupTaggingMockServerOkResponseDynamic(t, *server, customResponse, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.TaggingHost())
}

func TestSetupTaggingMockServerEmptyResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	path := "/tagging/empty"
	mockServer := SetupTaggingMockServerEmptyResponse(t, *server, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.TaggingHost())

	// Test the mock server responds with empty array
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupTaggingMockServer404Response(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	path := "/tagging/notfound"
	mockServer := SetupTaggingMockServer404Response(t, *server, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.TaggingHost())

	// Test the mock server responds with 404
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSetupTaggingMockServer500Response(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	path := "/tagging/error"
	mockServer := SetupTaggingMockServer500Response(t, *server, path)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.TaggingHost())

	// Test the mock server responds with 500
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestSetupGroupServiceFTMockServerOkResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	xdasHashes := &conversion.XdasHashes{
		Fields: map[string]string{"key1": "value1"},
	}
	path := "/ft/test"

	mockServer := SetupGroupServiceFTMockServerOkResponse(t, *server, path, xdasHashes)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.GroupServiceHost())

	// Test the mock server responds correctly
	resp, err := http.Get(mockServer.URL + path)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupGroupServiceFTMockServerOkResponseMultipleCalls(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	xdasHashes1 := &conversion.XdasHashes{Fields: map[string]string{"key1": "value1"}}
	xdasHashes2 := &conversion.XdasHashes{Fields: map[string]string{"key2": "value2"}}
	xdasHashes3 := &conversion.XdasHashes{Fields: map[string]string{"key3": "value3"}}

	path1 := "/ft/path1"
	path2 := "/ft/path2"
	path3 := "/ft/path3"

	mockServer := SetupGroupServiceFTMockServerOkResponseMultipleCalls(t, *server, path1, path2, path3, xdasHashes1, xdasHashes2, xdasHashes3)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.GroupServiceHost())

	// Test all three paths
	resp1, _ := http.Get(mockServer.URL + path1)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	resp2, _ := http.Get(mockServer.URL + path2)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	resp3, _ := http.Get(mockServer.URL + path3)
	assert.Equal(t, http.StatusOK, resp3.StatusCode)
}

func TestSetupGroupServiceHashesMockServerOkResponse(t *testing.T) {
	server, _ := GetTestXconfServer(GetTestConfig())

	xdasHashes := &conversion.XdasHashes{
		Fields: map[string]string{"hash1": "value1"},
	}

	mockServer := SetupGroupServiceHashesMockServerOkResponse(t, *server, xdasHashes)
	defer mockServer.Close()

	assert.NotNil(t, mockServer)
	assert.Equal(t, mockServer.URL, server.GroupServiceHost())

	// Test the mock server responds correctly
	resp, err := http.Get(mockServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test environment variables setup
func TestGetTestXconfServer_EnvironmentVariables(t *testing.T) {
	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("XPC_KEY")
		os.Unsetenv("SAT_CLIENT_ID")
		os.Unsetenv("SAT_CLIENT_SECRET")
	}()

	server, router := GetTestXconfServer(GetTestConfig())

	assert.NotNil(t, server)
	assert.NotNil(t, router)

	// Verify environment variables are set
	assert.Equal(t, "testXpcKey", os.Getenv("XPC_KEY"))
	assert.Equal(t, "foo", os.Getenv("SAT_CLIENT_ID"))
	assert.Equal(t, "bar", os.Getenv("SAT_CLIENT_SECRET"))
}

func TestGetTestXconfServer_InvalidConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid config file")
		}
	}()

	GetTestXconfServer("/invalid/path/to/config.conf")
}

// Integration test helper functions
func TestMockServerIntegration(t *testing.T) {
	t.Run("Multiple mock servers can coexist", func(t *testing.T) {
		server, _ := GetTestXconfServer(GetTestConfig())

		// Setup multiple mock servers
		satMock := SetupSatServiceMockServerOkResponse(t, *server)
		defer satMock.Close()

		taggingMock := SetupTaggingMockServerOkResponse(t, *server, "/tags")
		defer taggingMock.Close()

		accountMock := SetupAccountServiceMockServerOkResponse(t, *server, "/accounts")
		defer accountMock.Close()

		// Verify all are different URLs
		assert.NotEqual(t, satMock.URL, taggingMock.URL)
		assert.NotEqual(t, satMock.URL, accountMock.URL)
		assert.NotEqual(t, taggingMock.URL, accountMock.URL)

		// Verify server configurations are updated
		assert.Equal(t, satMock.URL, server.SatServiceHost())
		assert.Equal(t, taggingMock.URL, server.TaggingHost())
		assert.Equal(t, accountMock.URL, server.AccountServiceHost())
	})
}
