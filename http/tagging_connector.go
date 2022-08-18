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
	taggingServiceName = "tagging_service"
)

type TaggingConnector struct {
	*HttpClient
	host string
}

func NewTaggingConnector(conf *configuration.Config, tlsConfig *tls.Config) *TaggingConnector {
	confKey := fmt.Sprintf("xconfwebconfig.%v.host", taggingServiceName)
	host := conf.GetString(confKey)

	return &TaggingConnector{
		HttpClient: NewHttpClient(conf, taggingServiceName, tlsConfig),
		host:       host,
	}
}

func (c *TaggingConnector) TaggingHost() string {
	return c.host
}

func (c *TaggingConnector) SetTaggingHost(host string) {
	c.host = host
}

func (c *TaggingConnector) GetTagsForContext(contextMap map[string]string, token string, fields log.Fields) ([]string, error) {
	tags := []string{}
	return tags, nil
}
