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
	"encoding/json"
	"fmt"
	"os"

	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	satServiceUrlTemplate        = "%s/v2/oauth/token"
	satServicePartnerUrlTemplate = "%s/v2/oauth/token?partners=%s"
)

var satServiceName string

type SatServiceConnector interface {
	SatServiceName() string
	SatServiceHost() string
	ConsumerHost() string
	SetSatServiceName(name string)
	SetSatServiceHost(host string)
	GetSatTokenFromSatService(fields log.Fields, vargs ...string) (*SatServiceResponse, error)
}

type DefaultSatService struct {
	host         string
	consumerHost string
	headers      map[string]string
	name         string
	*HttpClient
}

type SatServiceResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"score"`
	TokenType    string `json:"token_type"`
	ResponseCode int    `json:"responseCode"`
	Description  string `json:"description"`
}

func NewSatServiceConnector(conf *configuration.Config, tlsConfig *tls.Config, externalSatConnector SatServiceConnector) SatServiceConnector {
	if externalSatConnector != nil {
		return externalSatConnector
	} else {
		// load SAT credentials
		satServiceName = conf.GetString("xconfwebconfig.xconf.sat_service_name")

		satClientId := os.Getenv("SAT_CLIENT_ID")
		if util.IsBlank(satClientId) {
			satClientId = conf.GetString("xconfwebconfig.%v.client_id", satServiceName)
			if util.IsBlank(satClientId) {
				panic("No env SAT_CLIENT_ID")
			}
		}

		satClientSecret := os.Getenv("SAT_CLIENT_SECRET")
		if util.IsBlank(satClientSecret) {
			satClientSecret = conf.GetString("xconfwebconfig.%v.client_secret", satServiceName)
			if util.IsBlank(satClientSecret) {
				panic("No env SAT_CLIENT_SECRET")
			}
		}
		confKey := fmt.Sprintf("xconfwebconfig.%v.host", satServiceName)
		host := conf.GetString(confKey)
		if util.IsBlank(host) {
			panic(fmt.Errorf("%s is required", confKey))
		}
		consumerHost := conf.GetString("xconfwebconfig.sat_consumer.consumer_host", host)
		if util.IsBlank(consumerHost) {
			panic(fmt.Errorf("%s is required", consumerHost))
		}

		headers := map[string]string{
			"X-Client-Id":     satClientId,
			"X-Client-Secret": satClientSecret,
		}

		return &DefaultSatService{
			HttpClient:   NewHttpClient(conf, satServiceName, tlsConfig),
			host:         host,
			consumerHost: consumerHost,
			headers:      headers,
			name:         satServiceName,
		}
	}
}

func (c *DefaultSatService) SatServiceName() string {
	return c.name
}

func (c *DefaultSatService) SetSatServiceName(name string) {
	c.name = name
}

func (c *DefaultSatService) SatServiceHost() string {
	return c.host
}

func (c *DefaultSatService) ConsumerHost() string {
	return c.consumerHost
}

func (c *DefaultSatService) SetSatServiceHost(host string) {
	c.host = host
}

func (c *DefaultSatService) GetSatTokenFromSatService(fields log.Fields, vargs ...string) (*SatServiceResponse, error) {
	var cb2Res *SatServiceResponse
	var url string

	if len(vargs) > 0 {
		partnerId := vargs[0]
		url = fmt.Sprintf(satServicePartnerUrlTemplate, c.SatServiceHost(), partnerId)
	} else {
		url = fmt.Sprintf(satServiceUrlTemplate, c.SatServiceHost())
	}
	rbytes, err := c.DoWithRetries("POST", url, c.headers, nil, fields, satServiceName)
	if err != nil {
		return cb2Res, err
	}

	if err := json.Unmarshal(rbytes, &cb2Res); err != nil {
		return cb2Res, err
	}

	if len(cb2Res.AccessToken) == 0 {
		err := fmt.Errorf("%v", cb2Res.Description)
		return nil, err
	}
	return cb2Res, nil
}
