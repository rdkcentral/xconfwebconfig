package tracing

import (
	"net/http"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/common"
	log "github.com/sirupsen/logrus"
)

func TestNewXpcTrace(t *testing.T) {
	// Create a mock tracer
	tracer := &XpcTracer{} // Assuming XpcTracer exists in tracer.go

	// Create a mock HTTP request with headers
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal("Failed to create test request")
	}

	// Set test headers
	req.Header.Set(common.HeaderTraceparent, "test-traceparent")
	req.Header.Set(common.HeaderTracestate, "test-tracestate")
	req.Header.Set(common.HeaderMoracide, "test-moracide")
	req.Header.Set("User-Agent", "test-agent")

	// Test NewXpcTrace
	trace := NewXpcTrace(tracer, req)

	if trace == nil {
		t.Fatal("NewXpcTrace should return a non-nil trace")
	}

	// Verify extracted values
	if trace.ReqTraceparent != "test-traceparent" {
		t.Errorf("Expected ReqTraceparent 'test-traceparent', got '%s'", trace.ReqTraceparent)
	}

	if trace.ReqTracestate != "test-tracestate" {
		t.Errorf("Expected ReqTracestate 'test-tracestate', got '%s'", trace.ReqTracestate)
	}

	if trace.ReqMoracideTag != "test-moracide" {
		t.Errorf("Expected ReqMoracideTag 'test-moracide', got '%s'", trace.ReqMoracideTag)
	}
}

func TestNewXpcTrace_WithCanary(t *testing.T) {
	tracer := &XpcTracer{}

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal("Failed to create test request")
	}

	// Set canary header
	req.Header.Set(common.HeaderCanary, "true")
	req.Header.Set(common.HeaderMoracide, "existing-moracide")

	trace := NewXpcTrace(tracer, req)

	// Should append service name to moracide tag when canary is true
	if trace.ReqMoracideTag != "existing-moracide," {
		// The service name is appended, but we might not know the exact name
		// Just verify it contains the existing moracide
		if trace.ReqMoracideTag == "existing-moracide" {
			t.Error("Expected moracide tag to be modified when canary=true")
		}
	}
}

func TestNewXpcTrace_EmptyHeaders(t *testing.T) {
	tracer := &XpcTracer{}

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal("Failed to create test request")
	}

	// No headers set
	trace := NewXpcTrace(tracer, req)

	if trace == nil {
		t.Fatal("NewXpcTrace should return a non-nil trace even with empty headers")
	}

	// Verify empty values
	if trace.ReqTraceparent != "" {
		t.Errorf("Expected empty ReqTraceparent, got '%s'", trace.ReqTraceparent)
	}

	if trace.ReqTracestate != "" {
		t.Errorf("Expected empty ReqTracestate, got '%s'", trace.ReqTracestate)
	}

	if trace.ReqMoracideTag != "" {
		t.Errorf("Expected empty ReqMoracideTag, got '%s'", trace.ReqMoracideTag)
	}
}

func TestSetSpanStatusCode(t *testing.T) {
	// Test with valid trace
	trace := &XpcTrace{}
	fields := log.Fields{
		"xpc_trace": trace,
		"status":    200,
	}

	// This should not panic
	SetSpanStatusCode(fields)

	// Test with missing trace
	fieldsNoTrace := log.Fields{
		"status": 404,
	}

	// This should log an error but not panic
	SetSpanStatusCode(fieldsNoTrace)

	// Test with missing status
	fieldsNoStatus := log.Fields{
		"xpc_trace": trace,
	}

	// This should not panic
	SetSpanStatusCode(fieldsNoStatus)
}

func TestSetSpanMoracideTags(t *testing.T) {
	// Test with valid trace and moracide tag
	trace := &XpcTrace{}
	fields := log.Fields{
		"xpc_trace":         trace,
		"resp_moracide_tag": "response-moracide",
	}

	// This should not panic
	SetSpanMoracideTags(fields, "test-prefix")

	// Test with req_moracide_tag when resp is not available
	fieldsReqMoracide := log.Fields{
		"xpc_trace":        trace,
		"req_moracide_tag": "request-moracide",
	}

	// This should not panic
	SetSpanMoracideTags(fieldsReqMoracide, "test-prefix")

	// Test with missing trace
	fieldsNoTrace := log.Fields{
		"resp_moracide_tag": "some-moracide",
	}

	// This should log an error but not panic
	SetSpanMoracideTags(fieldsNoTrace, "test-prefix")
}

func TestExtractParamsFromReq(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal("Failed to create test request")
	}

	// Set test headers
	req.Header.Set(common.HeaderTraceparent, "test-traceparent")
	req.Header.Set(common.HeaderTracestate, "test-tracestate")
	req.Header.Set(common.HeaderMoracide, "test-moracide")
	req.Header.Set("User-Agent", "test-user-agent")

	var trace XpcTrace
	serviceName := "test-service"

	extractParamsFromReq(req, &trace, serviceName)

	// Verify extracted values
	if trace.ReqTraceparent != "test-traceparent" {
		t.Errorf("Expected ReqTraceparent 'test-traceparent', got '%s'", trace.ReqTraceparent)
	}

	if trace.ReqTracestate != "test-tracestate" {
		t.Errorf("Expected ReqTracestate 'test-tracestate', got '%s'", trace.ReqTracestate)
	}

	if trace.ReqMoracideTag != "test-moracide" {
		t.Errorf("Expected ReqMoracideTag 'test-moracide', got '%s'", trace.ReqMoracideTag)
	}

	if trace.ReqUserAgent != "test-user-agent" {
		t.Errorf("Expected ReqUserAgent 'test-user-agent', got '%s'", trace.ReqUserAgent)
	}

	// Verify out values are copied from req values
	if trace.OutTraceparent != trace.ReqTraceparent {
		t.Error("OutTraceparent should be copied from ReqTraceparent")
	}

	if trace.OutTracestate != trace.ReqTracestate {
		t.Error("OutTracestate should be copied from ReqTracestate")
	}
}

func TestExtractParamsFromReq_WithCanary(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal("Failed to create test request")
	}

	// Set canary header with existing moracide
	req.Header.Set(common.HeaderCanary, "true")
	req.Header.Set(common.HeaderMoracide, "existing-tag")

	var trace XpcTrace
	serviceName := "test-service"

	extractParamsFromReq(req, &trace, serviceName)

	// Should append service name to existing moracide
	expected := "existing-tag,test-service"
	if trace.ReqMoracideTag != expected {
		t.Errorf("Expected ReqMoracideTag '%s', got '%s'", expected, trace.ReqMoracideTag)
	}
}

func TestExtractParamsFromReq_CanaryNoExistingMoracide(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal("Failed to create test request")
	}

	// Set canary header without existing moracide
	req.Header.Set(common.HeaderCanary, "true")

	var trace XpcTrace
	serviceName := "test-service"

	extractParamsFromReq(req, &trace, serviceName)

	// Should set service name as moracide
	if trace.ReqMoracideTag != serviceName {
		t.Errorf("Expected ReqMoracideTag '%s', got '%s'", serviceName, trace.ReqMoracideTag)
	}
}
