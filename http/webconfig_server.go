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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/tracing"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/go-akka/configuration"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	LevelWarn = iota
	LevelInfo
	LevelDebug
	MetricsEnabledDefault     = true
	responseLoggingLowerBound = 1000
	responseLoggingUpperBound = 5000
)

var (
	appName string
)

// len(response) < lowerBound               ==> convert to json
// lowerBound <= len(response) < upperBound ==> stay string
// upperBound <= len(response)              ==> truncated

type XconfServer struct {
	*http.Server
	db.DatabaseClient
	*common.ServerConfig
	SatServiceConnector
	DeviceServiceConnector
	AccountServiceConnector
	TaggingConnector
	*AppMetricsConfig
	GroupServiceConnector
	GroupServiceSyncConnector
	*tracing.XpcTracer
	SecurityTokenConfig          *SecurityTokenConfig
	LogUploadSecurityTokenConfig *SecurityTokenPathConfig
	FirmwareSecurityTokenConfig  *SecurityTokenPathConfig
	AppMetrics                   IAppMetrics
	tlsConfig                    *tls.Config
	notLoggedHeaders             []string
	metricsEnabled               bool
	AppName                      string
}

type ExternalConnectors struct {
	db.CassandraConnector
	DeviceServiceConnector
	AccountServiceConnector
	TaggingConnector
	SatServiceConnector
	GroupServiceConnector
	GroupServiceSyncConnector
}

type AppMetricsConfig struct {
	modelRequestsCounterEnabled bool
	allowedModelLabelsSet       util.Set
	connectionTypeMap           map[string]string
}

func NewTlsConfig(conf *configuration.Config) (*tls.Config, error) {
	certValidationEnabled := conf.GetBoolean("xconfwebconfig.http_client.enable_cert_validation", true)
	if !certValidationEnabled {
		log.Warn("TLS certificate validation is disabled by config flag")
		return nil, nil
	}

	caCertPEM, err := ioutil.ReadFile(conf.GetString("xconfwebconfig.http_client.ca_comodo_cert_file"))
	if err != nil {
		return nil, fmt.Errorf("unable to read comodo cert file %s with error: %+v",
			conf.GetString("xconfwebconfig.http_client.ca_comodo_cert_file"), err)
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		return nil, fmt.Errorf("unable to append cert from pem: %+v", err)
	}

	certFile := conf.GetString("xconfwebconfig.http_client.cert_file")
	if len(certFile) == 0 {
		return nil, fmt.Errorf("missing file %v", certFile)
	}
	privateKeyFile := conf.GetString("xconfwebconfig.http_client.private_key_file")
	if len(privateKeyFile) == 0 {
		return nil, fmt.Errorf("missing file %v", privateKeyFile)
	}
	cert, err := tls.LoadX509KeyPair(certFile, privateKeyFile)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}, nil
}

func NewExternalConnectors() *ExternalConnectors {
	return &ExternalConnectors{}
}

// testOnly=true ==> running unit test
func NewXconfServer(sc *common.ServerConfig, testOnly bool, ec *ExternalConnectors) *XconfServer {
	if ec == nil {
		ec = NewExternalConnectors()
	}
	conf := sc.Config
	var dbclient db.DatabaseClient
	var err error
	// appname from config
	appName = strings.Split(conf.GetString("code_git_commit", "xconfwebconfig-xconf"), "-")[0]
	//if we dont have custom cassandraconnector then we will use the default local cassandra connector
	if ec.CassandraConnector == nil {
		ec.CassandraConnector = &db.DefaultCassandraConnection{Connection_type: "local"}
	}
	dbclient, err = ec.NewCassandraClient(conf, testOnly)
	if err != nil {
		fmt.Printf("ERROR cassandra db init error=%v\n", err)
		panic(err)
	}
	db.SetDatabaseClient(dbclient)

	metricsEnabled := conf.GetBoolean("xconfwebconfig.server.metrics_enabled", MetricsEnabledDefault)

	// configure headers that should not be logged
	ignoredHeaders := conf.GetStringList("xconfwebconfig.log.ignored_headers")
	ignoredHeaders = append(common.DefaultIgnoredHeaders, ignoredHeaders...)
	var notLoggedHeaders []string
	for _, x := range ignoredHeaders {
		notLoggedHeaders = append(notLoggedHeaders, strings.ToLower(x))
	}
	securityTokenConfig := NewSecurityTokenConfig(conf)
	loguploadSecurityTokenConfig := NewLogUploaderNonMtlSsrTokenPathConfig(conf)
	firmwareSecurityTokenConfig := NewFirmwareNonMtlSsrTokenPathConfig(conf)

	tlsConfig, err := NewTlsConfig(conf)
	if err != nil && !testOnly {
		panic(err)
	}

	var serviceHostname string
	if conf.GetBoolean("xconfwebconfig.server.localhost_only") {
		serviceHostname = "localhost"
	}

	var appMetricsConfig *AppMetricsConfig
	modelRequestsCounterEnabled := conf.GetBoolean("xconfwebconfig.xconf.metrics_model_requests_counter_enabled", false)
	allowedModelLabels := conf.GetString("xconfwebconfig.xconf.metrics_allowed_model_labels")
	allowedModelLabelsList := strings.Split(allowedModelLabels, ";")
	allowedModelLabelsSet := util.NewSet(allowedModelLabelsList...)

	connectionTypeMap := InitConnectionTypeMap()

	appMetricsConfig = &AppMetricsConfig{
		modelRequestsCounterEnabled: modelRequestsCounterEnabled,
		allowedModelLabelsSet:       allowedModelLabelsSet,
		connectionTypeMap:           connectionTypeMap,
	}

	xpcTracer := tracing.NewXpcTracer(sc.Config)

	return &XconfServer{
		Server: &http.Server{
			Addr:         fmt.Sprintf("%s:%s", serviceHostname, conf.GetString("xconfwebconfig.server.port")),
			ReadTimeout:  time.Duration(conf.GetInt32("xconfwebconfig.server.read_timeout_in_secs", 3)) * time.Second,
			WriteTimeout: time.Duration(conf.GetInt32("xconfwebconfig.server.write_timeout_in_secs", 3)) * time.Second,
		},
		DatabaseClient:               dbclient,
		ServerConfig:                 sc,
		SecurityTokenConfig:          securityTokenConfig,
		LogUploadSecurityTokenConfig: loguploadSecurityTokenConfig,
		FirmwareSecurityTokenConfig:  firmwareSecurityTokenConfig,
		SatServiceConnector:          NewSatServiceConnector(conf, tlsConfig, ec.SatServiceConnector),
		AccountServiceConnector:      NewAccountServiceConnector(conf, tlsConfig, ec.AccountServiceConnector),
		DeviceServiceConnector:       NewDeviceServiceConnector(conf, tlsConfig, ec.DeviceServiceConnector),
		TaggingConnector:             NewTaggingConnector(conf, tlsConfig, ec.TaggingConnector),
		GroupServiceConnector:        NewGroupServiceConnector(conf, tlsConfig, ec.GroupServiceConnector),
		GroupServiceSyncConnector:    NewGroupServiceSyncConnector(conf, tlsConfig, ec.GroupServiceSyncConnector),
		tlsConfig:                    tlsConfig,
		notLoggedHeaders:             notLoggedHeaders,
		metricsEnabled:               metricsEnabled,
		AppName:                      appName,
		AppMetricsConfig:             appMetricsConfig,
		XpcTracer:                    xpcTracer,
	}
}

func (s *XconfServer) TestingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		xp := NewXResponseWriter(w)
		xw := *xp

		if r.Method == "POST" {
			if r.Body != nil {
				if rbytes, err := io.ReadAll(r.Body); err == nil {
					xw.SetBody(string(rbytes))
				}
			} else {
				xw.SetBody("")
			}
		}
		next.ServeHTTP(&xw, r)
	}
	return http.HandlerFunc(fn)
}

func (s *XconfServer) NoAuthMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		xw := s.LogRequestStarts(w, r)
		defer s.LogRequestEnds(&xw, r)
		next.ServeHTTP(&xw, r)
	}
	return http.HandlerFunc(fn)
}

func (s *XconfServer) MetricsEnabled() bool {
	return s.metricsEnabled
}

func (s *XconfServer) TlsConfig() *tls.Config {
	return s.tlsConfig
}

func (s *XconfServer) NotLoggedHeaders() []string {
	return s.notLoggedHeaders
}

func InitConnectionTypeMap() map[string]string {
	connectionTypeMap := make(map[string]string)

	connectionTypeMap[common.XCONF_HTTP_VALUE] = common.HTTP_CLIENT_PROTOCOL
	connectionTypeMap[common.XCONF_HTTPS_VALUE] = common.HTTPS_CLIENT_PROTOCOL
	connectionTypeMap[common.XCONF_MTLS_VALUE] = common.MTLS_CLIENT_PROTOCOL
	connectionTypeMap[common.XCONF_MTLS_RECOVERY_VALUE] = common.MTLS_RECOVERY_CLIENT_PROTOCOL
	connectionTypeMap[common.XCONF_MTLS_OPTIONAL_VALUE] = common.MTLS_OPTIONAL_CLIENT_PROTOCOL

	return connectionTypeMap
}

func getHeadersForLogAsMap(r *http.Request, notLoggedHeaders []string) map[string]interface{} {
	loggedHeaders := make(map[string]interface{})
	for k, v := range r.Header {
		if util.CaseInsensitiveContains(notLoggedHeaders, k) {
			continue
		}
		loggedHeaders[k] = v
	}
	return loggedHeaders
}

func (s *XconfServer) LogRequestStarts(w http.ResponseWriter, r *http.Request) XResponseWriter {
	remoteIp := r.RemoteAddr
	host := r.Host
	// extract the token from the header
	authorization := r.Header.Get("Authorization")
	elements := strings.Split(authorization, " ")
	token := ""
	if len(elements) == 2 && elements[0] == "Bearer" {
		token = elements[1]
	}

	moneytrace := r.Header.Get("X-Moneytrace")

	var xmTraceId string

	// extract moneytrace from the header
	tracePart := strings.Split(r.Header.Get("X-Moneytrace"), ";")[0]
	if elements := strings.Split(tracePart, "="); len(elements) == 2 {
		if elements[0] == "trace-id" {
			xmTraceId = elements[1]
		}
	}

	// extract auditid from the header
	auditId := r.Header.Get("X-Auditid")
	if len(auditId) == 0 {
		auditId = util.GetAuditId()
	}

	// traceparent handling for E2E tracing
	xpcTrace := tracing.NewXpcTrace(s.XpcTracer, r)
	traceId := xpcTrace.TraceID
	if len(traceId) == 0 {
		traceId = xmTraceId
	}

	fields := log.Fields{
		"path":             r.URL.String(),
		"method":           r.Method,
		"audit_id":         auditId,
		"remote_ip":        remoteIp,
		"host_name":        host,
		"logger":           "request",
		"trace_id":         traceId,
		"xmoney_trace_id":  xmTraceId,
		"moneytrace":       moneytrace,
		"traceparent":      xpcTrace.ReqTraceparent,
		"tracestate":       xpcTrace.ReqTracestate,
		"out_traceparent":  xpcTrace.OutTraceparent,
		"out_tracestate":   xpcTrace.OutTracestate,
		"req_moracide_tag": xpcTrace.ReqMoracideTag,
		"xpc_trace":        xpcTrace,
	}

	// add cpemac or csid in loggings
	params := mux.Vars(r)
	gtype := params["gtype"]
	switch gtype {
	case "cpe":
		mac := params["gid"]
		mac = strings.ToUpper(mac)
		fields["cpemac"] = mac
	case "configset":
		csid := params["gid"]
		csid = strings.ToLower(csid)
		fields["csid"] = csid
	}
	if mac, ok := params["mac"]; ok {
		mac = strings.ToUpper(mac)
		fields["cpemac"] = mac
	}

	xp := NewXResponseWriter(w, time.Now(), token, fields)
	xwriter := *xp

	copyFields := common.FilterLogFields(fields)
	copyFields["path"] = r.URL.String()
	copyFields["method"] = r.Method
	copyFields["header"] = getHeadersForLogAsMap(r, s.notLoggedHeaders)

	if r.Method == "POST" || r.Method == "PUT" {
		var body string
		if r.Body != nil {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				copyFields["error"] = err
				log.WithFields(copyFields).Error("request starts")
				return xwriter
			}
			body = string(b)
		}
		xwriter.SetBody(body)
		copyFields["body"] = body

		contentType := r.Header.Get("Content-type")
		if contentType == "application/msgpack" {
			xwriter.SetBodyObfuscated(true)
			copyFields["body"] = "****"
		}
	}

	log.WithFields(copyFields).Debug("request starts")

	return xwriter
}

func (s *XconfServer) LogRequestEnds(xw *XResponseWriter, r *http.Request) {
	tdiff := time.Since(xw.StartTime())
	duration := tdiff.Nanoseconds() / 1000000

	url := r.URL.String()
	response := xw.Response()
	if strings.Contains(url, "/config") || (strings.Contains(url, "/document") && r.Method == "GET") || (url == "/api/v1/token" && r.Method == "POST") {
		response = "****"
	}

	fields := xw.Audit()
	statusCode := xw.Status()

	// Log the response only for failures, don't log the response for happy path
	// Log response also for a special path, "xconf/swu/{applicationType"}
	pathTemplate, _ := mux.CurrentRoute(r).GetPathTemplate()
	splPath := false
	if strings.Contains(pathTemplate, "xconf/swu/{applicationType}") {
		splPath = true
		fields["body_text"] = xw.Body()
	}
	if splPath || statusCode >= http.StatusBadRequest {
		fields["response"] = response
		if len(response) < responseLoggingLowerBound {
			dict := util.Dict{}
			err := json.Unmarshal([]byte(response), &dict)
			if err == nil && len(dict) > 0 {
				if _, ok := dict["password"]; ok {
					dict["password"] = "****"
				}
				fields["response"] = dict
			}
		} else if len(response) > responseLoggingUpperBound {
			fields["response"] = fmt.Sprintf("%v...TRUNCATED", response[:responseLoggingUpperBound])
		}
	}

	fields["path"] = r.URL.String()
	fields["method"] = r.Method
	fields["status"] = statusCode
	fields["duration"] = duration
	fields["logger"] = "request"
	fields["response_header"] = xw.Header()

	s.XpcTracer.SetSpan(fields, s.XpcTracer.MoracideTagPrefix())

	// always add a "num_results" fields if not already in fields
	// splunk shows that only admin APIs like /queries or /firmwarerule carries nonzero num_results
	if _, ok := fields["num_results"]; !ok {
		fields["num_results"] = 0
	}

	tfields := common.FilterLogFields(fields)

	log.WithFields(tfields).Info("request ends")

	if s.metricsEnabled && s.AppMetrics != nil {
		s.AppMetrics.UpdateAPIMetrics(r, xw.Status(), xw.StartTime())
	}
}

func LogError(w http.ResponseWriter, err error) {
	var fields log.Fields
	if xw, ok := w.(*XResponseWriter); ok {
		xfields := xw.Audit()
		fields = common.FilterLogFields(xfields)
	} else {
		fields = make(log.Fields)
	}
	fields["error"] = err

	log.WithFields(fields).Error("internal error")
}

func (xw *XResponseWriter) logMessage(r *http.Request, logger string, message string, level int) {
	xfields := xw.Audit()
	fields := common.FilterLogFields(xfields)
	fields["logger"] = logger

	switch level {
	case LevelWarn:
		log.WithFields(fields).Warn(message)
	case LevelInfo:
		log.WithFields(fields).Info(message)
	case LevelDebug:
		log.WithFields(fields).Debug(message)
	}
}

func (xw *XResponseWriter) LogDebug(r *http.Request, logger string, message string) {
	xw.logMessage(r, logger, message, LevelDebug)
}

func (xw *XResponseWriter) LogInfo(r *http.Request, logger string, message string) {
	xw.logMessage(r, logger, message, LevelInfo)
}

func (xw *XResponseWriter) LogWarn(r *http.Request, logger string, message string) {
	xw.logMessage(r, logger, message, LevelWarn)
}

// AppName is just a convenience func that returns the AppName, used in metrics
func AppName() string {
	return appName
}

func (m *AppMetricsConfig) GetAllowedModelLabelsSet() util.Set {
	return m.allowedModelLabelsSet
}

func (m *AppMetricsConfig) GetConnectionTypeMap() map[string]string {
	return m.connectionTypeMap
}

func (s *XconfServer) StopXpcTracer() {
	sdkTraceProvider, ok := s.XpcTracer.OtelTracerProvider().(*sdktrace.TracerProvider)
	if ok && sdkTraceProvider != nil {
		sdkTraceProvider.Shutdown(context.TODO())
	}
}

func (s *XconfServer) SpanMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For DD, we rely on default instrumentation
		// For Otel, create the span explicitly
		if s.XpcTracer.OtelEnabled {
			ctx, otelSpan := tracing.NewOtelSpan(s.XpcTracer, r)
			r = r.WithContext(ctx)
			defer tracing.EndOtelSpan(s.XpcTracer, otelSpan)
		}
		next.ServeHTTP(w, r)
	})
}
