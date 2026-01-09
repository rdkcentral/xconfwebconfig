package http

import (
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestNewXResponseWriter(t *testing.T) {
	recorder := httptest.NewRecorder()

	// Test with no additional arguments
	writer := NewXResponseWriter(recorder)
	if writer == nil {
		t.Error("NewXResponseWriter should return a non-nil writer")
	}
	if writer.ResponseWriter != recorder {
		t.Error("ResponseWriter should be set correctly")
	}

	// Test with audit fields
	audit := log.Fields{"key": "value"}
	writer = NewXResponseWriter(recorder, audit)
	if writer.Audit()["key"] != "value" {
		t.Error("Audit fields should be set correctly")
	}

	// Test with time
	startTime := time.Now()
	writer = NewXResponseWriter(recorder, startTime)
	if writer.StartTime() != startTime {
		t.Error("Start time should be set correctly")
	}

	// Test with token
	token := "test-token"
	writer = NewXResponseWriter(recorder, token)
	if writer.Token() != token {
		t.Error("Token should be set correctly")
	}
}

func TestXResponseWriter_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	status := 404
	writer.WriteHeader(status)

	if writer.Status() != status {
		t.Errorf("Status should be %d, got %d", status, writer.Status())
	}
	if recorder.Code != status {
		t.Errorf("Underlying recorder status should be %d, got %d", status, recorder.Code)
	}
}

func TestXResponseWriter_Write(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	data := []byte("test response data")
	n, err := writer.Write(data)

	if err != nil {
		t.Errorf("Write should not return error, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Write should return %d bytes written, got %d", len(data), n)
	}
	if recorder.Body.String() != string(data) {
		t.Errorf("Underlying recorder should contain written data")
	}
}

func TestXResponseWriter_Status(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	// Initial status should be 0
	if writer.Status() != 0 {
		t.Errorf("Initial status should be 0, got %d", writer.Status())
	}

	// After writing header
	writer.WriteHeader(200)
	if writer.Status() != 200 {
		t.Errorf("Status should be 200 after WriteHeader, got %d", writer.Status())
	}
}

func TestXResponseWriter_Response(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	// Initially empty
	if writer.Response() != "" {
		t.Errorf("Initial response should be empty, got '%s'", writer.Response())
	}

	// After write
	writer.Write([]byte("test"))
	if writer.Response() != "test" {
		t.Errorf("Response should be 'test', got '%s'", writer.Response())
	}
}

func TestXResponseWriter_StartTime(t *testing.T) {
	recorder := httptest.NewRecorder()
	startTime := time.Now()
	writer := NewXResponseWriter(recorder, startTime)

	if writer.StartTime() != startTime {
		t.Error("StartTime should return the provided start time")
	}
}

func TestXResponseWriter_AuditId(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	// Initially should be empty
	if writer.AuditId() != "" {
		t.Errorf("Initial AuditId should be empty, got '%s'", writer.AuditId())
	}

	testId := "test-audit-123"
	writer.SetAuditData("audit_id", testId)

	if writer.AuditId() != testId {
		t.Errorf("AuditId should be '%s', got '%s'", testId, writer.AuditId())
	}
}

func TestXResponseWriter_Body(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	// Initially empty
	if writer.Body() != "" {
		t.Errorf("Initial body should be empty, got '%s'", writer.Body())
	}

	testBody := "test body content"
	writer.SetBody(testBody)

	if writer.Body() != testBody {
		t.Errorf("Body should be '%s', got '%s'", testBody, writer.Body())
	}
}

func TestXResponseWriter_SetBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	testBody := "custom body"
	writer.SetBody(testBody)

	if writer.Body() != testBody {
		t.Errorf("SetBody should set body to '%s', got '%s'", testBody, writer.Body())
	}
}

func TestXResponseWriter_Token(t *testing.T) {
	recorder := httptest.NewRecorder()
	token := "test-token-123"
	writer := NewXResponseWriter(recorder, token)

	if writer.Token() != token {
		t.Errorf("Token should be '%s', got '%s'", token, writer.Token())
	}
}

func TestXResponseWriter_TraceId(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	// Initially should be empty
	if writer.TraceId() != "" {
		t.Errorf("Initial TraceId should be empty, got '%s'", writer.TraceId())
	}

	testTraceId := "trace-123-456"
	writer.SetAuditData("trace_id", testTraceId)

	traceId := writer.TraceId()
	if traceId != testTraceId {
		t.Errorf("TraceId should be '%s', got '%s'", testTraceId, traceId)
	}
}

func TestXResponseWriter_Audit(t *testing.T) {
	recorder := httptest.NewRecorder()
	audit := log.Fields{"key1": "value1", "key2": "value2"}

	writer := NewXResponseWriter(recorder, audit)

	retrievedAudit := writer.Audit()
	if retrievedAudit["key1"] != "value1" {
		t.Error("Audit should contain correct key1 value")
	}
	if retrievedAudit["key2"] != "value2" {
		t.Error("Audit should contain correct key2 value")
	}
}

func TestXResponseWriter_AuditData(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	testKey := "testKey"
	testValue := "testValue"

	// Test setting audit data
	writer.SetAuditData(testKey, testValue)

	// Test getting audit data
	auditData := writer.AuditData(testKey)
	if auditData != testValue {
		t.Errorf("Audit data should be '%s', got '%s'", testValue, auditData)
	}
}

func TestXResponseWriter_SetAuditData(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	testKey := "testKey"
	testData := "audit data content"
	writer.SetAuditData(testKey, testData)

	if writer.AuditData(testKey) != testData {
		t.Errorf("Audit data should be '%s', got '%s'", testData, writer.AuditData(testKey))
	}
}

func TestXResponseWriter_SetBodyObfuscated(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewXResponseWriter(recorder)

	// This method doesn't have a getter, so we just test it doesn't panic
	writer.SetBodyObfuscated(true)
	writer.SetBodyObfuscated(false)
}

func TestXResponseWriter_String(t *testing.T) {
	recorder := httptest.NewRecorder()
	audit := log.Fields{"test": "value"}
	startTime := time.Now()
	writer := NewXResponseWriter(recorder, audit, startTime)

	writer.WriteHeader(200)
	writer.Write([]byte("test response"))

	str := writer.String()

	// Check that the string contains expected components
	if str == "" {
		t.Error("String() should return a non-empty string")
	}

	// Should contain status, length, response, startTime, audit
	// We can't check exact format, but we can check it's not empty
	t.Logf("String output: %s", str)
}
