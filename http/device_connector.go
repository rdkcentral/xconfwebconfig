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

	"github.com/rdkcentral/xconfwebconfig/common"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	getMeshPodUrlTemplate = "%s/api/v1/operational/mesh-pod/%s/account"
)

type DeviceServiceConnector interface {
	DeviceServiceHost() string
	SetDeviceServiceHost(host string)
	GetMeshPodAccountBySerialNum(serialNum string, fields log.Fields) (DeviceServiceObject, error)
}

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

type DefaultDeviceService struct {
	*HttpClient
	host string
}

var deviceServiceName string

func NewDeviceServiceConnector(conf *configuration.Config, tlsConfig *tls.Config, externalDeviceConnector DeviceServiceConnector) DeviceServiceConnector {

	if externalDeviceConnector != nil {
		return externalDeviceConnector
	} else {
		deviceServiceName := conf.GetString("xconfwebconfig.xconf.device_service_name")
		confKey := fmt.Sprintf("xconfwebconfig.%v.host", deviceServiceName)
		host := conf.GetString(confKey)
		if util.IsBlank(host) {
			panic(fmt.Errorf("%s is required", confKey))
		}

		return &DefaultDeviceService{
			HttpClient: NewHttpClient(conf, deviceServiceName, tlsConfig),
			host:       host,
		}
	}
}

func (c *DefaultDeviceService) DeviceServiceHost() string {
	return c.host
}

func (c *DefaultDeviceService) SetDeviceServiceHost(host string) {
	c.host = host
}

func (c *DefaultDeviceService) GetMeshPodAccountBySerialNum(serialNum string, fields log.Fields) (DeviceServiceObject, error) {
	url := fmt.Sprintf(getMeshPodUrlTemplate, c.DeviceServiceHost(), serialNum)
	headers := map[string]string{
		common.HeaderUserAgent: common.HeaderXconfDataService,
	}
	var deviceServiceObject DeviceServiceObject
	rrbytes, err := c.DoWithRetries("GET", url, headers, nil, fields, deviceServiceName)
	if err != nil {
		return deviceServiceObject, err
	}
	err = json.Unmarshal(rrbytes, &deviceServiceObject)
	if err != nil {
		return deviceServiceObject, err
	}
	return deviceServiceObject, nil
}
