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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"

	"github.com/go-akka/configuration"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	defaultConnectTimeout      = 30
	defaultReadTimeout         = 30
	defaultMaxIdleConns        = 0
	defaultMaxIdleConnsPerHost = 100
	defaultKeepaliveTimeout    = 30
	defaultRetries             = 3
	defaultRetriesInMsecs      = 1000
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type HttpClient struct {
	*http.Client
	retries      int
	retryInMsecs int
}

func NewHttpClient(conf *configuration.Config, serviceName string, tlsConfig *tls.Config) *HttpClient {
	confKey := fmt.Sprintf("xconfwebconfig.%v.connect_timeout_in_secs", serviceName)
	connectTimeout := int(conf.GetInt32(confKey, defaultConnectTimeout))

	confKey = fmt.Sprintf("xconfwebconfig.%v.read_timeout_in_secs", serviceName)
	readTimeout := int(conf.GetInt32(confKey, defaultReadTimeout))

	confKey = fmt.Sprintf("xconfwebconfig.%v.max_idle_conns", serviceName)
	maxIdleConns := int(conf.GetInt32(confKey, defaultMaxIdleConns))

	confKey = fmt.Sprintf("xconfwebconfig.%v.max_idle_conns_per_host", serviceName)
	maxIdleConnsPerHost := int(conf.GetInt32(confKey, defaultMaxIdleConnsPerHost))

	confKey = fmt.Sprintf("xconfwebconfig.%v.keepalive_timeout_in_secs", serviceName)
	keepaliveTimeout := int(conf.GetInt32(confKey, defaultKeepaliveTimeout))

	confKey = fmt.Sprintf("xconfwebconfig.%v.retries", serviceName)
	retries := int(conf.GetInt32(confKey, defaultRetries))

	confKey = fmt.Sprintf("xconfwebconfig.%v.retry_in_msecs", serviceName)
	retryInMsecs := int(conf.GetInt32(confKey, defaultRetriesInMsecs))

	return &HttpClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   time.Duration(connectTimeout) * time.Second,
					KeepAlive: time.Duration(keepaliveTimeout) * time.Second,
				}).DialContext,
				MaxIdleConns:          maxIdleConns,
				MaxIdleConnsPerHost:   maxIdleConnsPerHost,
				IdleConnTimeout:       time.Duration(keepaliveTimeout) * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig:       tlsConfig,
			},
			Timeout: time.Duration(readTimeout) * time.Second,
		},
		retries:      retries,
		retryInMsecs: retryInMsecs,
	}
}

// Do is a wrapper around http.Client.Do
// Inputs: method, url, headers, body as bytes (bbytes), fields for logging (baseFields),
//
//	external service being called (loggerName), attempt # (retry)
//
// Returns: response body as bytes, any err, whether a retry is useful or not, and the status code
func (c *HttpClient) Do(method string, url string, headers map[string]string, bbytes []byte, baseFields log.Fields, loggerName string, retry int) ([]byte, error, bool, int) {
	// verify a response is received
	var respMoracideTagsFound bool
	defer func(found *bool) {
		if !*found {
			log.Debugf("http_client: no moracide tags in response")
		}
	}(&respMoracideTagsFound)

	// statusCode is used in metrics
	statusCode := http.StatusInternalServerError // Default status to return

	var req *http.Request
	var err error
	switch method {
	case "GET":
		req, err = http.NewRequest(method, url, nil)
	case "POST", "PATCH", "DELETE":
		if len(bbytes) > 0 {
			req, err = http.NewRequest(method, url, bytes.NewReader(bbytes))
		} else {
			req, err = http.NewRequest(method, url, nil)
		}
	default:
		return nil, fmt.Errorf("method=%v", method), false, statusCode
	}

	if err != nil {
		return nil, err, true, statusCode
	}

	c.addMoracideTags(headers, baseFields)
	logHeaders := map[string]string{}
	for k, v := range headers {
		req.Header.Set(k, v)
		if k == "Authorization" || k == "X-Client-Secret" {
			logHeaders[k] = "****"
		} else {
			logHeaders[k] = v
		}
	}

	tfields := common.FilterLogFields(baseFields)
	tfields["logger"] = loggerName

	urlKey := fmt.Sprintf("%v_url", loggerName)
	tfields[urlKey] = url

	methodKey := fmt.Sprintf("%v_method", loggerName)
	tfields[methodKey] = method

	headersKey := fmt.Sprintf("%v_headers", loggerName)
	tfields[headersKey] = logHeaders

	bodyKey := fmt.Sprintf("%v_body", loggerName)
	if len(bbytes) > 0 {
		tfields[bodyKey] = string(bbytes)
	}
	fields := common.CopyLogFields(tfields)

	var startMessage string
	if retry > 0 {
		startMessage = fmt.Sprintf("%v retry=%v starts", loggerName, retry)
	} else {
		startMessage = fmt.Sprintf("%v starts", loggerName)
	}
	log.WithFields(fields).Debug(startMessage)
	startTime := time.Now()

	res, err := c.Client.Do(req)

	tdiff := time.Now().Sub(startTime)
	duration := tdiff.Nanoseconds() / 1000000
	fields[fmt.Sprintf("%v_duration", loggerName)] = duration

	delete(fields, urlKey)
	delete(fields, methodKey)
	delete(fields, headersKey)
	delete(fields, bodyKey)

	var endMessage string
	if retry > 0 {
		endMessage = fmt.Sprintf("%v retry=%v ends", loggerName, retry)
	} else {
		endMessage = fmt.Sprintf("%v ends", loggerName)
	}

	errorKey := fmt.Sprintf("%v_error", loggerName)

	if res != nil {
		respMoracideTagsFound = c.addMoracideTagsFromResponse(res.Header, baseFields)
	}

	if err != nil {
		fields[errorKey] = err.Error()
		log.WithFields(fields).Info(endMessage)
		return nil, err, true, statusCode
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	// client.Do succeeded, set the status to response's status code
	statusCode = res.StatusCode

	fields[fmt.Sprintf("%v_status", loggerName)] = res.StatusCode
	rbytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fields[errorKey] = err.Error()
		log.WithFields(fields).Info(endMessage)
		return nil, err, false, statusCode
	}

	var rbody string
	if loggerName == Ws.SatServiceConnector.SatServiceName() && res.StatusCode == http.StatusOK {
		rbody = "****"
	} else if loggerName == "xpc" && strings.HasSuffix(url, "/api/v1/common/sat") && res.StatusCode == http.StatusOK {
		rbody = "****"
	} else {
		rbody = string(rbytes)
	}

	fields[fmt.Sprintf("%v_response", loggerName)] = rbody
	log.WithFields(fields).Debugf("%v ends", loggerName)

	if res.StatusCode >= 400 {
		var errorMessage string
		if len(rbody) > 0 {
			var er ErrorResponse
			if err := json.Unmarshal(rbytes, &er); err == nil {
				errorMessage = er.Message
			}
			if len(errorMessage) == 0 {
				errorMessage = rbody
			}
		} else {
			errorMessage = http.StatusText(res.StatusCode)
		}
		err := common.RemoteHttpError{
			Message:    errorMessage,
			StatusCode: res.StatusCode,
		}

		switch res.StatusCode {
		case http.StatusForbidden, http.StatusBadRequest, http.StatusNotFound, 520:
			return rbytes, err, false, statusCode
		}
		return rbytes, err, true, statusCode
	}
	return rbytes, nil, false, statusCode
}

func (c *HttpClient) DoWithRetries(method string, url string, inHeaders map[string]string, bbytes []byte, fields log.Fields, loggerName string) ([]byte, error) {
	var traceId string
	if itf, ok := fields["xmoney_trace_id"]; ok {
		traceId = itf.(string)
	}
	if len(traceId) == 0 {
		traceId = uuid.New().String()
	}

	xmoney := fmt.Sprintf("trace-id=%s;parent-id=0;span-id=0;span-name=%s", traceId, loggerName)
	headers := map[string]string{
		"X-Moneytrace": xmoney,
	}
	for k, v := range inHeaders {
		headers[k] = v
	}

	// var res *http.Response
	var rbytes []byte
	var err error
	var cont bool
	var statusCode int

	startTimeForAllRetries := time.Now()

	extServiceAuditFields := make(map[string]interface{})
	extServiceAuditFields["audit_id"] = fields["audit_id"]
	extServiceAuditFields["trace_id"] = fields["trace_id"]

	i := 0
	// i=0 is NOT considered a retry, so it ends at i=c.webpaRetries
	for i = 0; i <= c.retries; i++ {
		cbytes := make([]byte, len(bbytes))
		copy(cbytes, bbytes)
		if i > 0 {
			time.Sleep(time.Duration(c.retryInMsecs) * time.Millisecond)
		}
		rbytes, err, cont, statusCode = c.Do(method, url, headers, cbytes, extServiceAuditFields, loggerName, i)
		if !cont {
			break
		}
	}

	if Ws.metricsEnabled && Ws.AppMetrics != nil {
		Ws.AppMetrics.UpdateExternalAPIMetrics(loggerName, method, statusCode, startTimeForAllRetries)
	}

	if err != nil {
		return rbytes, err
	}
	return rbytes, nil
}

// addMoracideTags - if ctx has a moracide tag as a header, add it to the headers
// Also add traceparent, tracestate headers
func (c *HttpClient) addMoracideTags(header map[string]string, fields log.Fields) {
	if itf, ok := fields["out_traceparent"]; ok {
		if ss, ok := itf.(string); ok {
			if len(ss) > 0 {
				header[common.HeaderTraceparent] = ss
			}
		}
	}
	if itf, ok := fields["out_tracestate"]; ok {
		if ss, ok := itf.(string); ok {
			if len(ss) > 0 {
				header[common.HeaderTracestate] = ss
			}
		}
	}

	moracide := common.FieldsGetString(fields, "req_moracide_tag")
	if len(moracide) > 0 {
		header[common.HeaderMoracide] = moracide
	}
}

func (c *HttpClient) addMoracideTagsFromResponse(header http.Header, fields log.Fields) bool {
	var respMoracideTagsFound bool
	moracide := header.Get(common.HeaderMoracide)
	if len(moracide) > 0 {
		fields["resp_moracide_tag"] = moracide
		respMoracideTagsFound = true
	}
	return respMoracideTagsFound
}
