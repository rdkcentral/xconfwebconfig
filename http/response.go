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

	"xconfwebconfig/util"
)

const (
	OkResponseTemplate = `{"status":200,"message":"OK","data":%v}`

	// TODO, this is should be retired
	TR181ResponseTemplate = `{"parameters":%v,"version":"%v"}`
)

// http ok response
type HttpResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// http error response
type HttpErrorResponse struct {
	Status    int         `json:"status"`
	ErrorCode int         `json:"error_code,omitempty"`
	Message   string      `json:"message,omitempty"`
	Errors    interface{} `json:"errors,omitempty"`
}

type ResponseEntity struct {
	Status int
	Error  error
	Data   interface{}
}

func NewResponseEntity(status int, err error, data interface{}) *ResponseEntity {
	return &ResponseEntity{
		Status: status,
		Error:  err,
		Data:   data,
	}
}

func writeByMarshal(w http.ResponseWriter, status int, o interface{}) {
	if rbytes, err := util.JSONMarshal(o); err == nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(status)
		w.Write(rbytes)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		LogError(w, err)
	}
}

//helper function to wirte a json response into ResponseWriter
func WriteOkResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	resp := HttpResponse{
		Status:  http.StatusOK,
		Message: http.StatusText(http.StatusOK),
		Data:    data,
	}
	writeByMarshal(w, http.StatusOK, resp)
}

func WriteOkResponseByTemplate(w http.ResponseWriter, r *http.Request, dataStr string) {
	rbytes := []byte(fmt.Sprintf(OkResponseTemplate, dataStr))
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(rbytes)
}

func WriteTR181Response(w http.ResponseWriter, r *http.Request, params string, version string) {
	w.Header().Set("Content-type", "application/json")
	w.Header().Set("ETag", version)
	w.WriteHeader(http.StatusOK)
	rbytes := []byte(fmt.Sprintf(TR181ResponseTemplate, params, version))
	w.Write(rbytes)
}

// this is used to return default tr-181 payload while the cpe is not in the db
func WriteContentTypeAndResponse(w http.ResponseWriter, r *http.Request, rbytes []byte, version string, contentType string) {
	w.Header().Set("Content-type", contentType)
	w.Header().Set("ETag", version)
	w.WriteHeader(http.StatusOK)
	w.Write(rbytes)
}

//helper function to write a failure json response into ResponseWriter
func WriteErrorResponse(w http.ResponseWriter, status int, err error) {
	errstr := ""
	if err != nil {
		errstr = err.Error()
	}
	resp := HttpErrorResponse{
		Status:  status,
		Message: http.StatusText(status),
		Errors:  errstr,
	}
	writeByMarshal(w, status, resp)
}

func Error(w http.ResponseWriter, status int, err error) {
	switch status {
	case http.StatusNoContent, http.StatusNotModified, http.StatusForbidden:
		w.WriteHeader(status)
	default:
		WriteErrorResponse(w, status, err)
	}
}

func WriteResponseBytes(w http.ResponseWriter, rbytes []byte, statusCode int, vargs ...string) {
	if len(vargs) > 0 {
		w.Header().Set("Content-type", vargs[0])
	}
	w.WriteHeader(statusCode)
	w.Write(rbytes)
}

func WriteXconfResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func WriteXconfResponseAsText(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-type", "text/plain")
	w.WriteHeader(status)
	w.Write(data)
}

func WriteXconfResponseWithHeaders(w http.ResponseWriter, headers map[string]string, status int, data []byte) {
	w.Header().Set("Content-type", "application/json")
	for k, v := range headers {
		w.Header()[k] = []string{v}
	}
	w.WriteHeader(status)
	w.Write(data)
}

func WriteXconfResponseHtmlWithHeaders(w http.ResponseWriter, headers map[string]string, status int, data []byte) {
	w.Header().Set("Content-type", "text/html; charset=iso-8859-1")
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	w.Write(data)
}
