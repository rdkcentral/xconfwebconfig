/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestSimpleHandler(t *testing.T) {
	server := NewXconfServer(sc, true, nil)
	router := server.GetRouter(true)

	// ==== test version api ====
	req, err := http.NewRequest("GET", "/version", nil)
	assert.NilError(t, err)
	res := ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, 200)

	rbytes, err := io.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	t.Log(string(rbytes))

	// ==== test monitor api ====
	req, err = http.NewRequest("GET", "/monitor", nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, 200)

	rbytes, err = io.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, len(rbytes), 0)

	// ==== test monitor api by HEAD ====
	req, err = http.NewRequest("HEAD", "/monitor", nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, 200)

	rbytes, err = io.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	assert.Equal(t, len(rbytes), 0)

	// ==== test server config api ====
	req, err = http.NewRequest("GET", "/config", nil)
	assert.NilError(t, err)
	res = ExecuteRequest(req, router).Result()
	assert.Equal(t, res.StatusCode, 200)

	rbytes, err = io.ReadAll(res.Body)
	assert.NilError(t, err)
	res.Body.Close()
	t.Log(string(rbytes))

	// get the expected config file
	configBytes, err := os.ReadFile(testConfigFile)
	assert.NilError(t, err)
	assert.DeepEqual(t, rbytes, configBytes)
}

// Additional tests moved from simple_handler_additional_test.go for better organization

func TestXconfServer_VersionHandler_Direct(t *testing.T) {
	server := NewXconfServer(sc, true, nil)

	req := httptest.NewRequest("GET", "/version", nil)
	recorder := httptest.NewRecorder()

	server.VersionHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	body := recorder.Body.String()
	// Just check that we get a valid JSON response
	if body == "" {
		t.Error("Response should not be empty")
	}
}

func TestXconfServer_InfoVersionHandler_Direct(t *testing.T) {
	server := NewXconfServer(sc, true, nil)

	req := httptest.NewRequest("GET", "/info/version", nil)
	recorder := httptest.NewRecorder()

	server.InfoVersionHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "<html>") {
		t.Error("Response should be HTML")
	}
	if !strings.Contains(body, "ServiceInfo") {
		t.Error("Response should contain ServiceInfo")
	}
}

func TestXconfServer_MonitorHandler_Direct(t *testing.T) {
	server := NewXconfServer(sc, true, nil)

	req := httptest.NewRequest("GET", "/monitor", nil)
	recorder := httptest.NewRecorder()

	server.MonitorHandler(recorder, req)

	// Monitor handler just sets Content-length header to 0
	contentLength := recorder.Header().Get("Content-length")
	if contentLength != "0" {
		t.Errorf("Expected Content-length to be '0', got '%s'", contentLength)
	}
}

func TestXconfServer_HealthZHandler_Direct(t *testing.T) {
	server := NewXconfServer(sc, true, nil)

	req := httptest.NewRequest("GET", "/healthz", nil)
	recorder := httptest.NewRecorder()

	server.HealthZHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestXconfServer_NotificationHandler_Direct(t *testing.T) {
	server := NewXconfServer(sc, true, nil)

	req := httptest.NewRequest("POST", "/notification", nil)
	recorder := httptest.NewRecorder()

	server.NotificationHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestXconfServer_ServerConfigHandler_Direct(t *testing.T) {
	server := NewXconfServer(sc, true, nil)

	req := httptest.NewRequest("GET", "/config", nil)
	recorder := httptest.NewRecorder()

	server.ServerConfigHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	// The response should contain the config bytes
	body := recorder.Body.String()
	if body == "" {
		t.Error("Response should not be empty")
	}
}

func TestXconfServer_NotFoundHandler_Direct(t *testing.T) {
	server := NewXconfServer(sc, true, nil)

	req := httptest.NewRequest("GET", "/non-existent-path", nil)
	recorder := httptest.NewRecorder()

	server.NotFoundHandler(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "HTTP ERROR 404") {
		t.Error("Response should contain 404 error message")
	}
	if !strings.Contains(body, "/non-existent-path") {
		t.Error("Response should contain the requested path")
	}
	if !strings.Contains(body, "Not Found") {
		t.Error("Response should contain 'Not Found' text")
	}
	if !strings.Contains(body, "<html>") {
		t.Error("Response should be HTML")
	}
}
