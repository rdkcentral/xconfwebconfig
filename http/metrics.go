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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// AppMetrics just collects all the needed metrics
type AppMetrics struct {
	counter        *prometheus.CounterVec
	duration       *prometheus.HistogramVec
	extAPIDuration *prometheus.HistogramVec
	extAPICounts   *prometheus.CounterVec
	inFlight       prometheus.Gauge
	responseSize   *prometheus.HistogramVec
	requestSize    *prometheus.HistogramVec
	fwCounts       *prometheus.CounterVec
}

var metrics *AppMetrics

// NewMetrics creates all the metrics needed for xconfwebconfig
func NewMetrics() *AppMetrics {
	metrics = &AppMetrics{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_requests_total",
				Help: "A counter for total number of requests.",
			},
			[]string{"app", "code", "method", "path"}, // app name, status code, http method, request URL
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_request_duration_seconds",
				Help:    "A histogram of latencies for requests.",
				Buckets: []float64{.01, .02, .05, 0.1, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"app", "code", "method", "path"},
		),
		extAPICounts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "external_api_count",
				Help: "A counter for external API calls",
			},
			[]string{"app", "code", "method", "service"}, // app name, status code, http method, extService
		),
		// Potential Problem: How many data points am I going to get?
		fwCounts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "xconf_firmware_penetration_counts",
				Help: "A counter for firmware penetration stats",
			},
			[]string{"partner", "model", "fw_version"}, // partner, model, version
		),
		extAPIDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "external_api_request_duration_seconds",
				Help:    "A histogram of latencies for requests.",
				Buckets: []float64{.01, .02, .05, 0.1, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"app", "code", "method", "service"},
		),
		inFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "in_flight_requests",
				Help: "A gauge of requests currently being served.",
			},
		),
		requestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "request_size_bytes",
				Help:    "A histogram of request sizes for requests.",
				Buckets: []float64{200, 500, 1000, 10000, 100000},
			},
			[]string{"app"},
		),
		responseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "response_size_bytes",
				Help:    "A histogram of response sizes for requests.",
				Buckets: []float64{200, 500, 1000, 10000, 100000},
			},
			[]string{"app"},
		),
	}
	prometheus.MustRegister(metrics.inFlight, metrics.counter, metrics.duration,
		metrics.extAPICounts, metrics.extAPIDuration,
		metrics.responseSize, metrics.requestSize,
		metrics.fwCounts)
	return metrics
}

// WebMetrics updates infligh, reqSize and respSize metrics
func (s *XconfServer) WebMetrics(m *AppMetrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promhttp.InstrumentHandlerInFlight(m.inFlight,
			promhttp.InstrumentHandlerRequestSize(m.requestSize.MustCurryWith(prometheus.Labels{"app": s.AppName}),
				promhttp.InstrumentHandlerResponseSize(m.responseSize.MustCurryWith(prometheus.Labels{"app": s.AppName}), next),
			),
		).ServeHTTP(w, r)
	})
}

// updateMetrics updates api_req_total, number of API calls
func (s *XconfServer) updateMetrics(xw *XResponseWriter, r *http.Request) {
	if metrics == nil {
		// Metrics may not be initialized in tests, or disabled by a config flag
		return
	}

	route := mux.CurrentRoute(r)
	if route == nil {
		// Paranoia, the code should never come here
		return
	}

	statusCode := strconv.Itoa(xw.Status())

	var path string
	var err error
	// Piggyback on mux's regex matching
	if path, err = route.GetPathTemplate(); err != nil {
		log.Debug(fmt.Sprintf("mux GetPathTemplate err in metrics %+v", err.Error()))
		path = "path extraction error"
	}

	vals := prometheus.Labels{"app": s.AppName, "code": statusCode, "method": r.Method, "path": path}
	metrics.counter.With(vals).Inc()
	metrics.duration.With(vals).Observe(time.Since(xw.StartTime()).Seconds())
}

// updateExternalAPIMetrics updates duration and counts for external API calls to AccountService, DeviceService, SatService etc.
func updateExternalAPIMetrics(statusCode int, method string, service string, startTime time.Time) {
	if metrics == nil {
		// Metrics may not be initialized in tests, or disabled by a config flag
		return
	}
	statusStr := strconv.Itoa(statusCode)
	vals := prometheus.Labels{"app": AppName(), "code": statusStr, "method": method, "service": service}
	metrics.extAPICounts.With(vals).Inc()

	externalCallDuration := time.Since(startTime).Seconds()
	metrics.extAPIDuration.With(vals).Observe(externalCallDuration)
}

// UpdateFirmwarePenetrationCounts updates the counts for firmware penetration dashboard
func UpdateFirmwarePenetrationCounts(partner string, model string, version string) {
	if metrics == nil {
		// Metrics may not be initialized in tests, or disabled by a config flag
		return
	}
	vals := prometheus.Labels{"partner": partner, "model": model, "fw_version": version}
	metrics.fwCounts.With(vals).Inc()
}
