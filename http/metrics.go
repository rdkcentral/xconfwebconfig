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
	"strings"
	"time"
	"xconfwebconfig/common"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Define the interface for core metrics so other packages can provide their own implementation
type IAppMetrics interface {
	MetricsHandler(next http.Handler) http.Handler
	UpdateAPIMetrics(req *http.Request, statusCode int, startTime time.Time)
	UpdateExternalAPIMetrics(service string, method string, statusCode int, startTime time.Time)
}

const (
	modelQueryParamKey   = "model"
	connectionTypeHeader = "HA-Haproxy-xconf-http"
)

// AppMetrics just collects all the needed metrics
type AppMetrics struct {
	counter                               *prometheus.CounterVec
	duration                              *prometheus.HistogramVec
	extAPICounts                          *prometheus.CounterVec
	extAPIDuration                        *prometheus.HistogramVec
	fwCounts                              *prometheus.CounterVec
	inFlight                              prometheus.Gauge
	responseSize                          *prometheus.HistogramVec
	requestSize                           *prometheus.HistogramVec
	logCounter                            *prometheus.CounterVec
	return304FromPrecookCounter           *prometheus.CounterVec
	return304RulesEngineCounter           *prometheus.CounterVec
	return200FromPrecookCounter           *prometheus.CounterVec
	return200RulesEngineCounter           *prometheus.CounterVec
	returnPostProcessFromPrecookCounter   *prometheus.CounterVec
	returnPostProcessOnTheFlyCounter      *prometheus.CounterVec
	noPrecookDataCounter                  *prometheus.CounterVec
	titanEmptyResponseCounter             *prometheus.CounterVec
	modelRequestsCounter                  *prometheus.CounterVec
	precookExcludeMacListCounter          *prometheus.CounterVec
	precookCtxHashMismatchCounter         *prometheus.CounterVec
	modelChangedCounter                   *prometheus.CounterVec
	modelChangedIn200Counter              *prometheus.CounterVec
	partnerChangedCounter                 *prometheus.CounterVec
	partnerChangedIn200Counter            *prometheus.CounterVec
	fwVersionChangedCounter               *prometheus.CounterVec
	offeredFwVersionMatchedCounter        *prometheus.CounterVec
	fwVersionMismatchCounter              *prometheus.CounterVec
	fwVersionChangedIn200Counter          *prometheus.CounterVec
	experienceChangedCounter              *prometheus.CounterVec
	experienceChangedIn200Counter         *prometheus.CounterVec
	accountIdChangedCounter               *prometheus.CounterVec
	accountIdChangedIn200Counter          *prometheus.CounterVec
	ipAddressNotInSameNetworkCounter      *prometheus.CounterVec
	ipAddressNotInSameNetworkIn200Counter *prometheus.CounterVec
	AccountServiceEmptyResponseCounter    *prometheus.CounterVec
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
		logCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "log_counter",
				Help: "A counter for total number of logs.",
			},
			[]string{"app", "logType"}, // app name, log type
		),
		return304FromPrecookCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "return_304_from_precook_count",
				Help: "A counter for total number of calls where we return 304 using precook results",
			},
			[]string{"app", "partner", "model"},
		),
		noPrecookDataCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "no_precook_data_count",
				Help: "A counter for total number of calls with no precook data",
			},
			[]string{"app", "partner", "model"},
		),
		return200FromPrecookCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "return_200_from_precook_count",
				Help: "A counter for total number of calls where we return 200 using precook data",
			},
			[]string{"app", "partner", "model"},
		),
		return304RulesEngineCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "return_304_rules_engine_count",
				Help: "A counter for total number of calls where we return a 304 after running the rules engine",
			},
			[]string{"app", "partner", "model"},
		),
		return200RulesEngineCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "return_200_rules_engine_count",
				Help: "A counter for total number of calls where we return 200 after running the rules engine",
			},
			[]string{"app", "partner", "model"},
		),
		returnPostProcessFromPrecookCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "return_post_process_from_precook_count",
				Help: "A counter for total number of calls where we return post-processed data using precook data",
			},
			[]string{"app", "partner", "model"},
		),
		returnPostProcessOnTheFlyCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "return_post_process_on_the_fly_count",
				Help: "A counter for total number of calls where we return post-processed data by generating on the fly",
			},
			[]string{"app", "partner", "model"},
		),
		precookExcludeMacListCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "precook_exclude_mac_list_count",
				Help: "A counter for total number of calls where we exclude precook due to mac being in a mac list",
			},
			[]string{"app", "partner", "model"},
		),
		precookCtxHashMismatchCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "precook_ctx_hash_mismatch_count",
				Help: "A counter for total number of calls where the ctx hash in xdas doesn't match the one computed from device call",
			},
			[]string{"app", "partner", "model"},
		),
		modelChangedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "model_change_count",
				Help: "A counter for total number of calls where the model has changed",
			},
			[]string{"app", "partner", "model"},
		),
		modelChangedIn200Counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "model_change_in_200_count",
				Help: "A counter for total number of calls where the model has changed in 200 response",
			},
			[]string{"app", "partner", "model"},
		),
		partnerChangedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "partner_change_count",
				Help: "A counter for total number of calls where the partner has changed",
			},
			[]string{"app", "partner", "model"},
		),
		partnerChangedIn200Counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "partner_change_in_200_count",
				Help: "A counter for total number of calls where the partner has changed in 200 response",
			},
			[]string{"app", "partner", "model"},
		),
		fwVersionChangedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fw_version_change_count",
				Help: "A counter for total number of calls where the firmware version has changed",
			},
			[]string{"app", "partner", "model"},
		),
		fwVersionMismatchCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fw_version_mismatch_count",
				Help: "A counter for total number of calls where the firmware version is not matching the ones in precook data",
			},
			[]string{"app", "partner", "model"},
		),
		offeredFwVersionMatchedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "offered_fw_version_matched_count",
				Help: "A counter for total number of calls where the firmware version matches the offered fw version in precook data",
			},
			[]string{"app", "partner", "model"},
		),
		fwVersionChangedIn200Counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fw_version_change_in_200_count",
				Help: "A counter for total number of calls where the firmware version has changed in 200 response",
			},
			[]string{"app", "partner", "model"},
		),
		experienceChangedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "experience_change_count",
				Help: "A counter for total number of calls where the experience has changed",
			},
			[]string{"app", "partner", "model"},
		),
		experienceChangedIn200Counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "experience_change_in_200_count",
				Help: "A counter for total number of calls where the experience has changed in 200 response",
			},
			[]string{"app", "partner", "model"},
		),
		accountIdChangedCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "account_change_count",
				Help: "A counter for total number of calls where the account has changed",
			},
			[]string{"app", "partner", "model"},
		),
		accountIdChangedIn200Counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "account_change_in_200_count",
				Help: "A counter for total number of calls where the account has changed in 200 response",
			},
			[]string{"app", "partner", "model"},
		),

		ipAddressNotInSameNetworkCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ip_address_not_in_same_network_count",
				Help: "A counter for total number of calls where the IP address is not in the same network",
			},
			[]string{"app", "partner", "model"},
		),
		ipAddressNotInSameNetworkIn200Counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ip_address_not_in_same_network_in_200_count",
				Help: "A counter for total number of calls where the IP address is not in the same network in 200 response",
			},
			[]string{"app", "partner", "model"},
		),
		titanEmptyResponseCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "titan_empty_response_count",
				Help: "A counter for empty 200 responses from titan",
			},
			[]string{"app", "model"},
		),
		modelRequestsCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_model_requests_count",
				Help: "A counter for total number of requests based on code, model and connection type",
			},
			[]string{"app", "code", "model", "connection_type"},
		),
		AccountServiceEmptyResponseCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "AccountService_empty_response_count",
				Help: "A counter for empty 200 responses from AccountService",
			},
			[]string{"app", "model"},
		),
	}
	prometheus.MustRegister(metrics.inFlight, metrics.counter, metrics.duration,
		metrics.extAPICounts, metrics.extAPIDuration,
		metrics.responseSize, metrics.requestSize, metrics.logCounter,
		metrics.return304FromPrecookCounter, metrics.return304RulesEngineCounter,
		metrics.return200FromPrecookCounter, metrics.return200RulesEngineCounter,
		metrics.returnPostProcessFromPrecookCounter, metrics.returnPostProcessOnTheFlyCounter,
		metrics.noPrecookDataCounter, metrics.precookExcludeMacListCounter, metrics.precookCtxHashMismatchCounter,
		metrics.modelChangedCounter, metrics.partnerChangedCounter, metrics.fwVersionChangedCounter, metrics.fwVersionMismatchCounter, metrics.offeredFwVersionMatchedCounter, metrics.experienceChangedCounter, metrics.accountIdChangedCounter, metrics.ipAddressNotInSameNetworkCounter,
		metrics.modelChangedIn200Counter, metrics.partnerChangedIn200Counter, metrics.fwVersionChangedIn200Counter, metrics.experienceChangedIn200Counter, metrics.accountIdChangedIn200Counter, metrics.ipAddressNotInSameNetworkIn200Counter,
		metrics.titanEmptyResponseCounter, metrics.modelRequestsCounter, metrics.AccountServiceEmptyResponseCounter,
	)
	return metrics
}

// WebMetrics updates infligh, reqSize and respSize metrics
func (s *XconfServer) SetWebMetrics(m IAppMetrics) IAppMetrics {
	s.AppMetrics = m // inject metrics into server
	return m
}

// MetricsHandler returns the prometheus handler
func (m *AppMetrics) MetricsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promhttp.InstrumentHandlerInFlight(m.inFlight,
			promhttp.InstrumentHandlerRequestSize(m.requestSize.MustCurryWith(prometheus.Labels{"app": Ws.AppName}),
				promhttp.InstrumentHandlerResponseSize(m.responseSize.MustCurryWith(prometheus.Labels{"app": Ws.AppName}), next),
			),
		).ServeHTTP(w, r)
	})
}

// UpdateAPIMetrics updates api_req_total, number of API calls
func (metrics *AppMetrics) UpdateAPIMetrics(r *http.Request, statusCode int, startTime time.Time) {
	if metrics == nil {
		// Metrics may not be initialized in tests, or disabled by a config flag
		return
	}

	route := mux.CurrentRoute(r)
	if route == nil {
		// Paranoia, the code should never come here
		return
	}

	statusStr := strconv.Itoa(statusCode)

	var path string
	var err error
	// Piggyback on mux's regex matching
	if path, err = route.GetPathTemplate(); err != nil {
		log.Debug(fmt.Sprintf("mux GetPathTemplate err in metrics %+v", err.Error()))
		path = "path extraction error"
	}

	if Ws.modelRequestsCounterEnabled {
		connectionType := strings.ToLower(r.Header.Get(connectionTypeHeader))

		queryParams := r.URL.Query()
		modelQueryParam := strings.ToLower(queryParams.Get(modelQueryParamKey))

		var connectionTypeLabel string = common.HTTP_CLIENT_PROTOCOL // considered as http by default

		if val, ok := Ws.AppMetricsConfig.connectionTypeMap[connectionType]; ok {
			connectionTypeLabel = val
		}

		var modelLabel string

		if modelQueryParam == "" {
			modelLabel = "null"
		} else {
			if Ws.allowedModelLabelsSet.Contains(modelQueryParam) {
				modelLabel = modelQueryParam
			} else {
				modelLabel = "others"
			}
		}
		durationLabels := prometheus.Labels{
			"app":             AppName(),
			"code":            statusStr,
			"model":           modelLabel,
			"connection_type": connectionTypeLabel,
		}
		metrics.modelRequestsCounter.With(durationLabels).Inc()
	}
	durationLabels := prometheus.Labels{
		"app":    Ws.AppName,
		"code":   statusStr,
		"method": r.Method,
		"path":   path,
	}
	metrics.counter.With(durationLabels).Inc()
	metrics.duration.With(durationLabels).Observe(time.Since(startTime).Seconds())
}

// UpdateExternalAPIMetrics updates duration and counts for external API calls to AccountService, sat etc.
func (metrics *AppMetrics) UpdateExternalAPIMetrics(service string, method string, statusCode int, startTime time.Time) {
	if metrics == nil {
		// Metrics may not be initialized in tests, or disabled by a config flag
		return
	}
	statusStr := strconv.Itoa(statusCode)
	vals := prometheus.Labels{
		"app":     Ws.AppName,
		"code":    statusStr,
		"method":  method,
		"service": service}
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

// updateLogCounter updates count for new logs
func UpdateLogCounter(logType string) {
	if metrics == nil {
		// Metrics may not be initialized in tests, or disabled by a config flag
		return
	}
	vals := prometheus.Labels{"app": AppName(), "logType": logType}
	metrics.logCounter.With(vals).Inc()
}

func IncreaseAccountServiceEmptyResponseCounter(model string) {
	if metrics == nil {
		return
	}

	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":   AppName(),
		"model": model,
	}
	metrics.AccountServiceEmptyResponseCounter.With(labels).Inc()
}

func IncreaseReturn304FromPrecookCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.return304FromPrecookCounter.With(labels).Inc()
}

func IncreaseReturn304RulesEngineCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.return304RulesEngineCounter.With(labels).Inc()
}

func IncreaseReturn200FromPrecookCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.return200FromPrecookCounter.With(labels).Inc()
}

func IncreaseReturnPostProcessFromPrecookCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.returnPostProcessFromPrecookCounter.With(labels).Inc()
}

func IncreaseReturn200RulesEngineCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.return200RulesEngineCounter.With(labels).Inc()
}

func IncreaseReturnPostProcessOnTheFlyCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.returnPostProcessOnTheFlyCounter.With(labels).Inc()
}

func IncreaseNoPrecookDataCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.noPrecookDataCounter.With(labels).Inc()
}

func IncreasePrecookExcludeMacListCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.precookExcludeMacListCounter.With(labels).Inc()
}

func IncreaseCtxHashMismatchCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.precookCtxHashMismatchCounter.With(labels).Inc()
}

func IncreaseModelChangedCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.modelChangedCounter.With(labels).Inc()
}

func IncreaseModelChangedIn200Counter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.modelChangedIn200Counter.With(labels).Inc()
}

func IncreasePartnerChangedCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.partnerChangedCounter.With(labels).Inc()
}

func IncreasePartnerChangedIn200Counter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.partnerChangedIn200Counter.With(labels).Inc()
}

func IncreaseFwVersionChangedCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.fwVersionChangedCounter.With(labels).Inc()
}

func IncreaseFirmwareVersionMismatchCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.fwVersionMismatchCounter.With(labels).Inc()
}

func IncreaseOfferedFwVersionMatchCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.offeredFwVersionMatchedCounter.With(labels).Inc()
}

func IncreaseFwVersionChangedIn200Counter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.fwVersionChangedIn200Counter.With(labels).Inc()
}

func IncreaseExperienceChangedCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.experienceChangedCounter.With(labels).Inc()
}

func IncreaseExperienceChangedIn200Counter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.experienceChangedIn200Counter.With(labels).Inc()
}

func IncreaseAccountIdChangedCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.accountIdChangedCounter.With(labels).Inc()
}

func IncreaseAccountIdChangedIn200Counter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.accountIdChangedIn200Counter.With(labels).Inc()
}

func IncreaseIpNotInSameNetworkCounter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.ipAddressNotInSameNetworkCounter.With(labels).Inc()
}

func IncreaseIpNotInSameNetworkIn200Counter(partner, model string) {
	if metrics == nil {
		return
	}
	if len(partner) == 0 {
		partner = "null"
	}
	if len(model) == 0 {
		model = "null"
	}
	labels := prometheus.Labels{
		"app":     AppName(),
		"partner": partner,
		"model":   model,
	}
	metrics.ipAddressNotInSameNetworkIn200Counter.With(labels).Inc()
}

func IncreaseTitanEmptyResponseCounter(model string) {
	if metrics == nil {
		return
	}

	if len(model) == 0 {
		model = "null"
	}

	labels := prometheus.Labels{
		"app":   AppName(),
		"model": model,
	}
	metrics.titanEmptyResponseCounter.With(labels).Inc()
}
