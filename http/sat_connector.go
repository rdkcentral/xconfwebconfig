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
	"fmt"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	satServiceName = "sat_service"
)

type SatServiceConnector struct {
	*HttpClient
	host string
	name string
}

type SatServiceResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"score"`
	TokenType    string `json:"token_type"`
	ResponseCode int    `json:"responseCode"`
	Description  string `json:"description"`
}

func NewSatServiceConnector(conf *configuration.Config, satClientId, satClientSecret string, tlsConfig *tls.Config) *SatServiceConnector {
	confKey := fmt.Sprintf("xconfwebconfig.%v.host", satServiceName)
	host := conf.GetString(confKey)
	return &SatServiceConnector{
		HttpClient: NewHttpClient(conf, satServiceName, tlsConfig),
		host:       host,
		name:       satServiceName,
	}
}

func (c *SatServiceConnector) SatServiceHost() string {
	return c.host
}

func (c *SatServiceConnector) SetSatServiceHost(host string) {
	c.host = host
}

func (c *SatServiceConnector) GetSatTokenFromSatService(fields log.Fields, vargs ...string) (*SatServiceResponse, error) {
	satServiceResponse := &SatServiceResponse{}
	return satServiceResponse, nil
}
