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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"xconfwebconfig/common"
	"xconfwebconfig/db"
	"xconfwebconfig/util"

	"github.com/go-akka/configuration"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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
	*SatServiceConnector
	*DeviceServiceConnector
	*AccountServiceConnector
	*TaggingConnector
	*GroupServiceConnector
	tlsConfig        *tls.Config
	notLoggedHeaders []string
	metricsEnabled   bool
	AppName          string
}

func NewTlsConfig(conf *configuration.Config) (*tls.Config, error) {
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

// testOnly=true ==> running unit test
func NewXconfServer(sc *common.ServerConfig, testOnly bool, dc db.DatabaseClient) *XconfServer {
	conf := sc.Config
	var dbclient db.DatabaseClient
	var err error

	// appname from config
	appName = strings.Split(conf.GetString("xconfwebconfig.code_git_commit", "xconfwebconfig"), "-")[0]

	if dc == nil {
		dbclient, err = db.NewCassandraClient(conf, testOnly)
		if err != nil {
			fmt.Printf("ERROR cassandra db init error=%v\n", err)
			panic(err)
		}
	} else {
		dbclient = dc
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

	// tlsConfig, here we ignore any error
	tlsConfig, _ := NewTlsConfig(conf)

	// load SAT credentials
	satClientId := os.Getenv("SAT_CLIENT_ID")
	if len(satClientId) == 0 {
		panic("No env SAT_CLIENT_ID")
	}

	satClientSecret := os.Getenv("SAT_CLIENT_SECRET")
	if len(satClientSecret) == 0 {
		panic("No env SAT_CLIENT_SECRET")
	}

	return &XconfServer{
		Server: &http.Server{
			Addr:         fmt.Sprintf(":%s", conf.GetString("xconfwebconfig.server.port")),
			ReadTimeout:  time.Duration(conf.GetInt32("xconfwebconfig.server.read_timeout_in_secs", 3)) * time.Second,
			WriteTimeout: time.Duration(conf.GetInt32("xconfwebconfig.server.write_timeout_in_secs", 3)) * time.Second,
		},
		DatabaseClient:          dbclient,
		ServerConfig:            sc,
		SatServiceConnector:     NewSatServiceConnector(conf, satClientId, satClientSecret, tlsConfig),
		AccountServiceConnector: NewAccountServiceConnector(conf, tlsConfig),
		DeviceServiceConnector:  NewDeviceServiceConnector(conf, tlsConfig),
		TaggingConnector:        NewTaggingConnector(conf, tlsConfig),
		GroupServiceConnector:   NewGroupServiceConnector(conf, tlsConfig),
		tlsConfig:               tlsConfig,
		notLoggedHeaders:        notLoggedHeaders,
		metricsEnabled:          metricsEnabled,
		AppName:                 appName,
	}
}

func (s *XconfServer) TestingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		xp := NewXResponseWriter(w)
		xw := *xp

		if r.Method == "POST" {
			if r.Body != nil {
				if rbytes, err := ioutil.ReadAll(r.Body); err == nil {
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

	// extract moneytrace from the header
	traceId := ""
	tracePart := strings.Split(r.Header.Get("X-Moneytrace"), ";")[0]
	if elements := strings.Split(tracePart, "="); len(elements) == 2 {
		if elements[0] == "trace-id" {
			traceId = elements[1]
		}
	}

	// extract auditid from the header
	auditId := r.Header.Get("X-Auditid")
	if len(auditId) == 0 {
		auditId = util.GetAuditId()
	}

	fields := log.Fields{
		"audit_id":  auditId,
		"remote_ip": remoteIp,
		"host_name": host,
		"logger":    "request",
		"trace_id":  traceId,
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
	copyFields := util.CopyLogFields(fields)
	copyFields["path"] = r.URL.String()
	copyFields["method"] = r.Method
	copyFields["header"] = getHeadersForLogAsMap(r, s.notLoggedHeaders)

	if r.Method == "POST" || r.Method == "PUT" {
		var body string
		if r.Body != nil {
			b, err := ioutil.ReadAll(r.Body)
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

	// always add a "num_results" fields if not already in fields
	// splunk shows that only admin APIs like /queries or /firmwarerule carries nonzero num_results
	if _, ok := fields["num_results"]; !ok {
		fields["num_results"] = 0
	}

	log.WithFields(fields).Info("request ends")
	s.updateMetrics(xw, r)
}

func LogError(w http.ResponseWriter, err error) {
	var fields log.Fields
	if xw, ok := w.(*XResponseWriter); ok {
		fields = xw.Audit()
		fields["error"] = err
	} else {
		fields = make(log.Fields)
	}

	log.WithFields(fields).Error("internal error")
}

func (xw *XResponseWriter) logMessage(r *http.Request, logger string, message string, level int) {
	fields := xw.Audit()
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
