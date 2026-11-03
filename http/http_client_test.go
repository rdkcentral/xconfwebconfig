package http

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-akka/configuration"
	"github.com/rdkcentral/xconfwebconfig/common"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Tests for http_client.go functions using existing mock infrastructure

func TestNewHttpClient(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	// Test NewHttpClient creation with correct parameters
	client := NewHttpClient(sc.Config, "test_service", nil)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Client) // Should have embedded http.Client
}

func TestNewHttpClientTlsOverrides(t *testing.T) {
	conf := configuration.ParseString(`
xconfwebconfig {
    svc_a { tls_server_name = "a.example.com" }
    svc_b { tls_server_name = "" }
    svc_c { tls_insecure_skip_verify = false }
}`)
	shared := &tls.Config{InsecureSkipVerify: false}
	sharedSkip := &tls.Config{InsecureSkipVerify: true}
	tlsOf := func(c *HttpClient) *tls.Config { return c.Client.Transport.(*http.Transport).TLSClientConfig }

	// server name override is applied on a clone, shared config untouched
	tlsA := tlsOf(NewHttpClient(conf, "svc_a", shared))
	assert.Equal(t, "a.example.com", tlsA.ServerName)
	assert.False(t, tlsA.InsecureSkipVerify)
	assert.NotSame(t, shared, tlsA)
	assert.Equal(t, "", shared.ServerName)

	// no overrides keeps the shared config as-is, including inherited skip verify
	assert.Same(t, shared, tlsOf(NewHttpClient(conf, "svc_b", shared)))
	assert.Same(t, sharedSkip, tlsOf(NewHttpClient(conf, "svc_b", sharedSkip)))

	// per-service tls_insecure_skip_verify is ignored; only http_client.insecure_skip_verify applies
	assert.Same(t, sharedSkip, tlsOf(NewHttpClient(conf, "svc_c", sharedSkip)))

	// nil shared config still gets the override
	assert.Equal(t, "a.example.com", tlsOf(NewHttpClient(conf, "svc_a", nil)).ServerName)
}

func TestAddMoracideTags(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	client := NewHttpClient(sc.Config, "test_service", nil)

	// Test with all moracide fields present
	fields := log.Fields{
		"out_traceparent":  "00-12345678901234567890123456789012-1234567890123456-01",
		"out_tracestate":   "test-state=value",
		"req_moracide_tag": "test-experiment-tag",
	}

	header := make(map[string]string)
	client.addMoracideTags(header, fields)

	assert.Equal(t, "00-12345678901234567890123456789012-1234567890123456-01", header[common.HeaderTraceparent])
	assert.Equal(t, "test-state=value", header[common.HeaderTracestate])
	assert.Equal(t, "test-experiment-tag", header[common.HeaderMoracide])
}

func TestAddMoracideTagsEmptyValues(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	client := NewHttpClient(sc.Config, "test_service", nil)

	// Test with empty values - should not add headers
	fields := log.Fields{
		"out_traceparent":  "",
		"out_tracestate":   "",
		"req_moracide_tag": "",
	}

	header := make(map[string]string)
	client.addMoracideTags(header, fields)

	assert.Empty(t, header[common.HeaderTraceparent])
	assert.Empty(t, header[common.HeaderTracestate])
	assert.Empty(t, header[common.HeaderMoracide])
}

func TestAddMoracideTagsPartialFields(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	client := NewHttpClient(sc.Config, "test_service", nil)

	// Test with only some fields present
	fields := log.Fields{
		"out_traceparent": "00-12345678901234567890123456789012-1234567890123456-01",
		// Missing out_tracestate
		"req_moracide_tag": "experiment-123",
	}

	header := make(map[string]string)
	client.addMoracideTags(header, fields)

	assert.Equal(t, "00-12345678901234567890123456789012-1234567890123456-01", header[common.HeaderTraceparent])
	assert.Empty(t, header[common.HeaderTracestate])
	assert.Equal(t, "experiment-123", header[common.HeaderMoracide])
}

func TestAddMoracideTagsInvalidTypes(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	client := NewHttpClient(sc.Config, "test_service", nil)

	// Test with invalid data types - should not add headers
	fields := log.Fields{
		"out_traceparent":  123,         // Invalid type, should be string
		"out_tracestate":   true,        // Invalid type, should be string
		"req_moracide_tag": "valid-tag", // This one should work
	}

	header := make(map[string]string)
	client.addMoracideTags(header, fields)

	assert.Empty(t, header[common.HeaderTraceparent])
	assert.Empty(t, header[common.HeaderTracestate])
	assert.Equal(t, "valid-tag", header[common.HeaderMoracide])
}

func TestAddMoracideTagsFromResponse(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	client := NewHttpClient(sc.Config, "test_service", nil)

	// Test with moracide header in response
	respHeader := http.Header{}
	respHeader.Set(common.HeaderMoracide, "response-experiment-tag")

	fields := log.Fields{}
	found := client.addMoracideTagsFromResponse(respHeader, fields)

	assert.True(t, found)
	assert.Equal(t, "response-experiment-tag", fields["resp_moracide_tag"])
}

func TestAddMoracideTagsFromResponseEmpty(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	client := NewHttpClient(sc.Config, "test_service", nil)

	// Test with no moracide header in response
	respHeader := http.Header{}

	fields := log.Fields{}
	found := client.addMoracideTagsFromResponse(respHeader, fields)

	assert.False(t, found)
	assert.Nil(t, fields["resp_moracide_tag"])
}

func TestHttpClientIntegrationWithMocks(t *testing.T) {
	sc, _ := common.NewServerConfig("../config/sample_xconfwebconfig.conf")
	server := NewXconfServer(sc, true, nil)
	server.SetupMocks() // Use existing mock infrastructure

	client := NewHttpClient(sc.Config, "integration_test_service", nil)

	// Create a test server that echoes headers
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the moracide header if present
		if moracideHeader := r.Header.Get(common.HeaderMoracide); moracideHeader != "" {
			w.Header().Set(common.HeaderMoracide, "echo-"+moracideHeader)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer testServer.Close()

	// Test HTTP client creation with mocks - doesn't hit real DB
	assert.NotNil(t, client)
	assert.NotNil(t, client.Client)
}
