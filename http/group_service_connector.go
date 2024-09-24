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
	"reflect"

	conversion "xconfwebconfig/protobuf"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	defaultGroupServiceHost = "https://test.net"
	groupServiceName        = "groupService"
	getCpeGroupsUrlTemplate = "%s/v2/cg/%s"
)

type GroupServiceConnector struct {
	*HttpClient
	host        string
	groupPrefix string
}

func NewGroupServiceConnector(conf *configuration.Config, tlsConfig *tls.Config) *GroupServiceConnector {
	confKey := fmt.Sprintf("webconfig.%v.host", groupServiceName)
	host := conf.GetString(confKey, defaultGroupServiceHost)
	groupPrefix := conf.GetString("webconfig.xconf.group_prefix")

	return &GroupServiceConnector{
		HttpClient:  NewHttpClient(conf, groupServiceName, tlsConfig),
		host:        host,
		groupPrefix: groupPrefix,
	}
}

func (c *GroupServiceConnector) GroupServiceHost() string {
	return c.host
}

func (c *GroupServiceConnector) SetGroupServiceHost(host string) {
	c.host = host
}

func (c *GroupServiceConnector) GroupPrefix() string {
	return c.groupPrefix
}

func (c *GroupServiceConnector) SetGroupPrefix(prefix string) {
	c.groupPrefix = prefix
}

func (c *GroupServiceConnector) GetCpeGroups(cpeMac string, fields log.Fields) ([]string, error) {
	cpeGroups := []string{}
	return cpeGroups, nil
}

func (c *GroupServiceConnector) CreateListFromGroupServiceProto(cpeGroup *conversion.CpeGroup) []string {
	cpeGroups := []string{}
	r := reflect.ValueOf(cpeGroup).Elem()
	for i := 0; i < r.NumField(); i++ {
		// make sure this is one of the actual exported boolean fields (CpeGroup also contains unexported non-group fields: state, sizeCache, and unknownFields)
		if r.Type().Field(i).IsExported() {
			value := r.Field(i).Interface()
			// check that the type is bool and that the value is true
			v, ok := value.(bool)
			if ok && v {
				name := r.Type().Field(i).Name
				cpeGroups = append(cpeGroups, fmt.Sprintf("%s%s", c.GroupPrefix(), name))
			}
		}
	}
	return cpeGroups
}
