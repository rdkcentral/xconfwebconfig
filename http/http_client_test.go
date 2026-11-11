package http

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

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

	// nil shared config still gets the override
	assert.Equal(t, "a.example.com", tlsOf(NewHttpClient(conf, "svc_a", nil)).ServerName)
}

func TestPartnerTlsConfigDoesNotInheritPrimaryOverride(t *testing.T) {
	conf := configuration.ParseString(`
xconfwebconfig {
    svc_a { tls_server_name = "primary.example.com" }
    partner_services {
        sat { tls_server_name = "partner.example.com" }
        account { host = "https://partner" }
    }
}`)
	shared := &tls.Config{}
	tlsOf := func(c *HttpClient) *tls.Config { return c.Client.Transport.(*http.Transport).TLSClientConfig }

	client := NewHttpClient(conf, "svc_a", shared)
	client.SetTlsConfig(ApplyServiceTlsOverrides(conf, "partner_services.sat", shared))
	assert.Equal(t, "partner.example.com", tlsOf(client).ServerName)

	// no partner override falls back to the shared config, not the primary one
	client = NewHttpClient(conf, "svc_a", shared)
	client.SetTlsConfig(ApplyServiceTlsOverrides(conf, "partner_services.account", shared))
	assert.Same(t, shared, tlsOf(client))
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

// writeTestCertPair generates a self-signed cert/key pair in dir and returns their paths and the parsed cert
func writeTestCertPair(t *testing.T, dir, name string) (string, string, *x509.Certificate) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	assert.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	assert.NoError(t, err)
	keyDer, err := x509.MarshalECPrivateKey(key)
	assert.NoError(t, err)

	certFile := filepath.Join(dir, name+".crt")
	keyFile := filepath.Join(dir, name+".key")
	assert.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600))
	assert.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDer}), 0600))
	return certFile, keyFile, cert
}

func TestNewTlsConfig(t *testing.T) {
	dir := t.TempDir()
	caFile, _, caCert := writeTestCertPair(t, dir, "ca")
	clientCert, clientKey, _ := writeTestCertPair(t, dir, "client")
	_, otherKey, _ := writeTestCertPair(t, dir, "other")
	badPem := filepath.Join(dir, "bad.pem")
	assert.NoError(t, os.WriteFile(badPem, []byte("not a pem"), 0600))
	missing := filepath.Join(dir, "missing.pem")

	expectedPool := x509.NewCertPool()
	expectedPool.AddCert(caCert)

	cases := []struct {
		name         string
		httpClient   string
		wantErr      string
		wantRootCAs  bool
		wantCerts    int
		wantSkipVrfy bool
	}{
		{name: "empty config uses system CAs, no mTLS"},
		{name: "ca_cert_file", httpClient: fmt.Sprintf(`ca_cert_file = %q`, caFile), wantRootCAs: true},
		{name: "legacy ca_comodo_cert_file", httpClient: fmt.Sprintf(`ca_comodo_cert_file = %q`, caFile), wantRootCAs: true},
		{name: "ca_cert_file takes precedence over legacy key", httpClient: fmt.Sprintf(`ca_cert_file = %q, ca_comodo_cert_file = %q`, caFile, missing), wantRootCAs: true},
		{name: "missing CA file", httpClient: fmt.Sprintf(`ca_cert_file = %q`, missing), wantErr: "failed to read CA certificate file"},
		{name: "invalid CA PEM", httpClient: fmt.Sprintf(`ca_cert_file = %q`, badPem), wantErr: "failed to parse CA certificate"},
		{name: "client cert and key enable mTLS", httpClient: fmt.Sprintf(`client_cert_file = %q, client_key_file = %q`, clientCert, clientKey), wantCerts: 1},
		{name: "legacy cert_file and private_key_file", httpClient: fmt.Sprintf(`cert_file = %q, private_key_file = %q`, clientCert, clientKey), wantCerts: 1},
		{name: "only client cert set", httpClient: fmt.Sprintf(`client_cert_file = %q`, clientCert), wantErr: "incomplete mTLS configuration"},
		{name: "only client key set", httpClient: fmt.Sprintf(`client_key_file = %q`, clientKey), wantErr: "incomplete mTLS configuration"},
		{name: "mismatched key pair", httpClient: fmt.Sprintf(`client_cert_file = %q, client_key_file = %q`, clientCert, otherKey), wantErr: "failed to load client key pair"},
		{name: "missing key pair files", httpClient: fmt.Sprintf(`client_cert_file = %q, client_key_file = %q`, missing, missing), wantErr: "failed to load client key pair"},
		{name: "insecure_skip_verify", httpClient: `insecure_skip_verify = true`, wantSkipVrfy: true},
		{
			name:        "CA and mTLS together",
			httpClient:  fmt.Sprintf(`ca_cert_file = %q, client_cert_file = %q, client_key_file = %q`, caFile, clientCert, clientKey),
			wantRootCAs: true,
			wantCerts:   1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conf := configuration.ParseString(fmt.Sprintf(`xconfwebconfig { http_client { %s } }`, tc.httpClient))
			cfg, err := NewTlsConfig(conf)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Nil(t, cfg)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			if tc.wantRootCAs {
				assert.True(t, expectedPool.Equal(cfg.RootCAs))
			} else {
				assert.Nil(t, cfg.RootCAs)
			}
			assert.Len(t, cfg.Certificates, tc.wantCerts)
			assert.Equal(t, tc.wantSkipVrfy, cfg.InsecureSkipVerify)
		})
	}
}
