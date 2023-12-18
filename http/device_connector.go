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
	deviceServiceName = "device_service"
)

type DeviceServiceData struct {
	AccountId string `json:"account_id"`
	CpeMac    string `json:"cpe_mac"`
	TimeZone  string `json:"timezone"`
	PartnerId string `json:"partner_id"`
}

type DeviceServiceObject struct {
	Status            int                `json:"status"`
	Message           string             `json:"message"`
	DeviceServiceData *DeviceServiceData `json:"data"`
}

type DeviceServiceConnector struct {
	*HttpClient
	host string
}

func NewDeviceServiceConnector(conf *configuration.Config, tlsConfig *tls.Config) *DeviceServiceConnector {
	confKey := fmt.Sprintf("xconfwebconfig.%v.host", deviceServiceName)
	host := conf.GetString(confKey)

	return &DeviceServiceConnector{
		HttpClient: NewHttpClient(conf, deviceServiceName, tlsConfig),
		host:       host,
	}
}

func (c *DeviceServiceConnector) DeviceServiceHost() string {
	return c.host
}

func (c *DeviceServiceConnector) SetDeviceServiceHost(host string) {
	c.host = host
}

func (c *DeviceServiceConnector) GetMeshPodAccountBySerialNum(serialNum string, fields log.Fields) (DeviceServiceObject, error) {
	var deviceServiceObject DeviceServiceObject
	return deviceServiceObject, nil
}
