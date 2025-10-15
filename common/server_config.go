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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

var serverOriginId string

func init() {
	if hostname, err := os.Hostname(); err != nil {
		log.Errorf("ERROR getting host name: %v", err)
		serverOriginId = fmt.Sprintf("%d", os.Getpid())
	} else {
		serverOriginId = fmt.Sprintf("%s:%d", hostname, os.Getpid())
	}
}

func ServerOriginId() string {
	return serverOriginId
}

type ServerConfig struct {
	*configuration.Config
	configBytes []byte
}

func NewServerConfig(configFile string) (*ServerConfig, error) {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	return NewServerConfigFromText(string(configBytes))
}

func NewServerConfigFromText(text string) (*ServerConfig, error) {
	conf := configuration.ParseString(text)
	return &ServerConfig{
		Config:      conf,
		configBytes: []byte(text),
	}, nil
}

func (c *ServerConfig) ConfigBytes() []byte {
	return c.configBytes
}
