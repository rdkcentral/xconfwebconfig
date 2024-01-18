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
	"xconfwebconfig/common"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	accountServiceName = "account_service"
	getAccountPath     = "%s/account/%s"
)

type AccountServiceConnector struct {
	*HttpClient
	host string
}

type AccountServiceDevices struct {
	Id         string     `json:"id"`
	DeviceData DeviceData `json:"data"`
}

type DeviceData struct {
	Partner           string `json:"partner"`
	ServiceAccountUri string `json:"serviceAccountId"`
}

type Account struct {
	Id          string      `json:"id"`
	AccountData AccountData `json:"data"`
}

type AccountData struct {
	AccountAttributes AccountAttributes `json:"attributes"`
}

type AccountAttributes struct {
	TimeZone string `json:"timeZone"`
}

func NewAccountServiceConnector(conf *configuration.Config, tlsConfig *tls.Config) *AccountServiceConnector {
	confKey := fmt.Sprintf("xconfwebconfig.%v.host", accountServiceName)
	host := conf.GetString(confKey)

	return &AccountServiceConnector{
		HttpClient: NewHttpClient(conf, accountServiceName, tlsConfig),
		host:       host,
	}
}

func (c *AccountServiceConnector) AccountServiceHost() string {
	return c.host
}

func (c *AccountServiceConnector) SetAccountServiceHost(host string) {
	c.host = host
}

func (c *AccountServiceConnector) GetAccountData(serviceAccountId string, token string, fields log.Fields) (Account, error) {
	url := fmt.Sprintf(getAccountPath, c.AccountServiceHost(), serviceAccountId)
	headers := map[string]string{
		common.HeaderAuthorization: fmt.Sprintf("Bearer %s", token),
		common.HeaderUserAgent:     common.HeaderXconfDataService,
	}
	var account Account
	rbytes, err := c.DoWithRetries("GET", url, headers, nil, fields, accountServiceName)
	if err != nil {
		return account, err
	}
	err = json.Unmarshal(rbytes, &account)
	if err != nil {
		return account, err
	}
	return account, nil
}

func (c *AccountServiceConnector) GetDevices(macKey string, macValue string, token string, fields log.Fields) (AccountServiceDevices, error) {
	var devicesInfo AccountServiceDevices
	return devicesInfo, nil
}
