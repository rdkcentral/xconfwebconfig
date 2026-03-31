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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test NewResponseEntity
func TestNewResponseEntity_Success(t *testing.T) {
	data := map[string]string{"key": "value"}
	entity := NewResponseEntity(http.StatusOK, nil, data)

	assert.NotNil(t, entity)
	assert.Equal(t, http.StatusOK, entity.Status)
	assert.Nil(t, entity.Error)
	assert.Equal(t, data, entity.Data)
}

func TestNewResponseEntity_WithError(t *testing.T) {
	err := errors.New("test error")
	entity := NewResponseEntity(http.StatusInternalServerError, err, nil)

	assert.NotNil(t, entity)
	assert.Equal(t, http.StatusInternalServerError, entity.Status)
	assert.Equal(t, err, entity.Error)
	assert.Nil(t, entity.Data)
}

func TestNewResponseEntity_AllFields(t *testing.T) {
	data := "test data"
	err := errors.New("error message")
	entity := NewResponseEntity(http.StatusBadRequest, err, data)

	assert.NotNil(t, entity)
	assert.Equal(t, http.StatusBadRequest, entity.Status)
	assert.Equal(t, err, entity.Error)
	assert.Equal(t, data, entity.Data)
}

// Test WriteOkResponseByTemplate
func TestWriteOkResponseByTemplate_ValidData(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	dataStr := `{"name":"test","value":123}`
	WriteOkResponseByTemplate(recorder, req, dataStr)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-type"))
	assert.Contains(t, recorder.Body.String(), `"status":200`)
	assert.Contains(t, recorder.Body.String(), `"message":"OK"`)
	assert.Contains(t, recorder.Body.String(), dataStr)
}

func TestWriteOkResponseByTemplate_EmptyData(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	WriteOkResponseByTemplate(recorder, req, "{}")

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"data":{}`)
}

func TestWriteOkResponseByTemplate_ComplexData(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	dataStr := `{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}`
	WriteOkResponseByTemplate(recorder, req, dataStr)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Alice")
	assert.Contains(t, recorder.Body.String(), "Bob")
}

// Test WriteTR181Response
func TestWriteTR181Response_ValidParams(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	params := `[{"name":"param1","value":"value1"}]`
	version := "1.0.0"

	WriteTR181Response(recorder, req, params, version)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-type"))
	assert.Equal(t, version, recorder.Header().Get("ETag"))
	assert.Contains(t, recorder.Body.String(), `"parameters":[{"name":"param1","value":"value1"}]`)
	assert.Contains(t, recorder.Body.String(), `"version":"1.0.0"`)
}

func TestWriteTR181Response_EmptyParams(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	WriteTR181Response(recorder, req, "[]", "2.0.0")

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "2.0.0", recorder.Header().Get("ETag"))
	assert.Contains(t, recorder.Body.String(), `"parameters":[]`)
}

func TestWriteTR181Response_MultipleParams(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	params := `[{"name":"Device.WiFi.SSID.1.SSID","value":"MyNetwork"},{"name":"Device.WiFi.Radio.1.Enable","value":"true"}]`
	version := "3.1.4"

	WriteTR181Response(recorder, req, params, version)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Device.WiFi.SSID.1.SSID")
	assert.Contains(t, recorder.Body.String(), "MyNetwork")
	assert.Contains(t, recorder.Body.String(), "3.1.4")
}

// Test WriteContentTypeAndResponse
func TestWriteContentTypeAndResponse_JSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	contentType := "application/json"
	responseBytes := []byte(`{"status":"success"}`)
	version := "1.0.0"

	WriteContentTypeAndResponse(recorder, req, responseBytes, version, contentType)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, contentType, recorder.Header().Get("Content-type"))
	assert.Equal(t, version, recorder.Header().Get("ETag"))
	assert.Equal(t, `{"status":"success"}`, recorder.Body.String())
}

func TestWriteContentTypeAndResponse_XML(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	contentType := "application/xml"
	responseBytes := []byte(`<response><status>success</status></response>`)

	WriteContentTypeAndResponse(recorder, req, responseBytes, "2.0.0", contentType)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, contentType, recorder.Header().Get("Content-type"))
	assert.Contains(t, recorder.Body.String(), "<response>")
}

func TestWriteContentTypeAndResponse_PlainText(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	WriteContentTypeAndResponse(recorder, req, []byte("Hello, World!"), "3.0", "text/plain")

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/plain", recorder.Header().Get("Content-type"))
	assert.Equal(t, "Hello, World!", recorder.Body.String())
}

// Test WriteErrorResponse
func TestWriteErrorResponse_WithError(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := errors.New("something went wrong")
	WriteErrorResponse(recorder, http.StatusBadRequest, err)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-type"))
	assert.Contains(t, recorder.Body.String(), `"status":400`)
	assert.Contains(t, recorder.Body.String(), "something went wrong")
}

func TestWriteErrorResponse_InternalServerError(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := errors.New("database connection failed")
	WriteErrorResponse(recorder, http.StatusInternalServerError, err)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"status":500`)
	assert.Contains(t, recorder.Body.String(), "database connection failed")
}

func TestWriteErrorResponse_NotFound(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := errors.New("resource not found")
	WriteErrorResponse(recorder, http.StatusNotFound, err)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"status":404`)
}

func TestWriteErrorResponse_NilError(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteErrorResponse(recorder, http.StatusBadRequest, nil)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"status":400`)
}

// Test Error
func TestError_BadRequest(t *testing.T) {
	recorder := httptest.NewRecorder()

	Error(recorder, http.StatusBadRequest, errors.New("invalid request"))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-type"))
	assert.Contains(t, recorder.Body.String(), `"status":400`)
	assert.Contains(t, recorder.Body.String(), "invalid request")
}

func TestError_Unauthorized(t *testing.T) {
	recorder := httptest.NewRecorder()

	Error(recorder, http.StatusUnauthorized, errors.New("authentication required"))

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"status":401`)
	assert.Contains(t, recorder.Body.String(), "authentication required")
}

func TestError_Forbidden(t *testing.T) {
	recorder := httptest.NewRecorder()

	Error(recorder, http.StatusForbidden, errors.New("access denied"))

	// Forbidden is special case - just writes status code
	assert.Equal(t, http.StatusForbidden, recorder.Code)
	// Should NOT contain JSON error response
	assert.Equal(t, "", recorder.Body.String())
}

func TestError_NoContent(t *testing.T) {
	recorder := httptest.NewRecorder()

	Error(recorder, http.StatusNoContent, nil)

	// NoContent is special case - just writes status code
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, "", recorder.Body.String())
}

func TestError_NotModified(t *testing.T) {
	recorder := httptest.NewRecorder()

	Error(recorder, http.StatusNotModified, nil)

	// NotModified is special case - just writes status code
	assert.Equal(t, http.StatusNotModified, recorder.Code)
	assert.Equal(t, "", recorder.Body.String())
}

// Test WriteResponseBytes
func TestWriteResponseBytes_JSONData(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := []byte(`{"result":"success","count":42}`)
	WriteResponseBytes(recorder, data, http.StatusOK, "application/json")

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-type"))
	assert.Equal(t, `{"result":"success","count":42}`, recorder.Body.String())
}

func TestWriteResponseBytes_EmptyData(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteResponseBytes(recorder, []byte{}, http.StatusOK)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "", recorder.Body.String())
}

func TestWriteResponseBytes_LargeData(t *testing.T) {
	recorder := httptest.NewRecorder()

	largeData := make([]byte, 1024)
	for i := range largeData {
		largeData[i] = 'A'
	}

	WriteResponseBytes(recorder, largeData, http.StatusOK)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, 1024, len(recorder.Body.Bytes()))
}

func TestWriteResponseBytes_CustomStatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := []byte(`{"error":"not found"}`)
	WriteResponseBytes(recorder, data, http.StatusNotFound, "application/json")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

// Test WriteXconfResponse
func TestWriteXconfResponse_ValidJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := []byte(`{"firmware":"version1.0","config":"enabled"}`)
	WriteXconfResponse(recorder, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-type"))
	assert.Contains(t, recorder.Body.String(), "firmware")
}

func TestWriteXconfResponse_CustomStatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := []byte(`{"status":"created"}`)
	WriteXconfResponse(recorder, http.StatusCreated, data)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-type"))
}

// Test WriteXconfResponseAsText
func TestWriteXconfResponseAsText_PlainText(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := []byte("This is plain text response")
	WriteXconfResponseAsText(recorder, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/plain", recorder.Header().Get("Content-type"))
	assert.Equal(t, "This is plain text response", recorder.Body.String())
}

func TestWriteXconfResponseAsText_EmptyText(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteXconfResponseAsText(recorder, http.StatusOK, []byte(""))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/plain", recorder.Header().Get("Content-type"))
}

func TestWriteXconfResponseAsText_CustomStatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteXconfResponseAsText(recorder, http.StatusAccepted, []byte("Accepted"))

	assert.Equal(t, http.StatusAccepted, recorder.Code)
}

// Test WriteXconfResponseWithHeaders
func TestWriteXconfResponseWithHeaders_MultipleHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := []byte(`{"data":"test"}`)
	headers := map[string]string{
		"Cache-Control": "no-cache",
	}

	WriteXconfResponseWithHeaders(recorder, headers, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-type"))
	assert.Equal(t, "no-cache", recorder.Header().Get("Cache-Control"))
	assert.Contains(t, recorder.Body.String(), `{"data":"test"}`)
}

func TestWriteXconfResponseWithHeaders_NoHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := []byte(`{"result":"success"}`)
	WriteXconfResponseWithHeaders(recorder, nil, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-type"))
}

func TestWriteXconfResponseWithHeaders_CustomStatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()

	data := []byte(`{"created":true}`)
	headers := map[string]string{
		"Location": "/api/resource/123",
	}

	WriteXconfResponseWithHeaders(recorder, headers, http.StatusCreated, data)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "/api/resource/123", recorder.Header().Get("Location"))
}

// Test WriteXconfResponseHtmlWithHeaders
func TestWriteXconfResponseHtmlWithHeaders_ValidHTML(t *testing.T) {
	recorder := httptest.NewRecorder()

	html := []byte("<html><body><h1>Test Page</h1></body></html>")
	headers := map[string]string{
		"X-Frame-Options": "DENY",
	}

	WriteXconfResponseHtmlWithHeaders(recorder, headers, http.StatusOK, html)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Content-type"), "text/html")
	assert.Equal(t, "DENY", recorder.Header().Get("X-Frame-Options"))
	assert.Contains(t, recorder.Body.String(), "<h1>Test Page</h1>")
}

func TestWriteXconfResponseHtmlWithHeaders_EmptyHTML(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteXconfResponseHtmlWithHeaders(recorder, nil, http.StatusOK, []byte(""))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Content-type"), "text/html")
}

func TestWriteXconfResponseHtmlWithHeaders_MultipleHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()

	html := []byte("<html><body>Content</body></html>")
	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-XSS-Protection":       "1; mode=block",
	}

	WriteXconfResponseHtmlWithHeaders(recorder, headers, http.StatusOK, html)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", recorder.Header().Get("X-XSS-Protection"))
}

func TestWriteXconfResponseHtmlWithHeaders_CustomStatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()

	html := []byte("<html><body>Not Found</body></html>")
	WriteXconfResponseHtmlWithHeaders(recorder, nil, http.StatusNotFound, html)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
