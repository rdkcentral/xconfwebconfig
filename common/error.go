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
package common

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	NotOK             = fmt.Errorf("!ok")
	NotFound          = fmt.Errorf("Not found")
	NotFirmwareConfig = fmt.Errorf("Not FirmwareCofig")
	NotFirmwareRule   = fmt.Errorf("Not FirmwareRule")
	Forbidden         = fmt.Errorf("Forbidden")
)

type Http400Error struct {
	Message string
}

func (e Http400Error) Error() string {
	return e.Message
}

type Http404Error struct {
	Message string
}

func (e Http404Error) Error() string {
	return e.Message
}

type Http500Error struct {
	Message string
}

func (e Http500Error) Error() string {
	return e.Message
}

type RemoteHttpError struct {
	StatusCode int
	Message    string
}

func (e RemoteHttpError) Error() string {
	return fmt.Sprintf("Http%v %v", e.StatusCode, e.Message)
}

func NewRemoteError(status int, message string) error {
	return RemoteHttpError{StatusCode: status, Message: message}
}

var XconfErrorType = &RemoteHttpErrorAS{}

type RemoteHttpErrorAS struct {
	StatusCode int
	Message    string
}

func (e RemoteHttpErrorAS) Error() string {
	return e.Message
}

func NewRemoteErrorAS(status int, message string) error {
	return RemoteHttpErrorAS{StatusCode: status, Message: message}
}

var (
	Http400ErrorType    = &Http400Error{}
	Http404ErrorType    = &Http404Error{}
	Http500ErrorType    = &Http500Error{}
	RemoteHttpErrorType = &RemoteHttpError{}
)

func GetXconfErrorStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.As(err, XconfErrorType) {
		return err.(RemoteHttpErrorAS).StatusCode
	}
	return http.StatusInternalServerError
}

func UnwrapAll(wrappedErr error) error {
	err := wrappedErr
	for i := 0; i < 10; i++ {
		unerr := errors.Unwrap(err)
		if unerr == nil {
			return err
		}
		err = unerr
	}
	return err
}
