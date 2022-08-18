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
	"net/http"

	"github.com/gorilla/mux"
)

func (s *XconfServer) AddBaseRoutes(testOnly bool, router *mux.Router) {
	r0 := router.Path("/monitor").Subrouter()
	r0.HandleFunc("", s.MonitorHandler).Methods("HEAD", "GET")

	r1 := router.Path("/healthz").Subrouter()
	r1.HandleFunc("", s.HealthZHandler).Methods("HEAD", "GET")

	r2 := router.Path("/version").Subrouter()
	r2.HandleFunc("", s.VersionHandler).Methods("GET")

	// register the notfound handler
	router.NotFoundHandler = http.HandlerFunc(s.NotFoundHandler)
}

func (s *XconfServer) GetRouter(testOnly bool) *mux.Router {
	router := mux.NewRouter()
	s.AddBaseRoutes(testOnly, router)

	return router
}
