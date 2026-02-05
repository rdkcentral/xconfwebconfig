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
	"net/http"
	"reflect"

	conversion "github.com/rdkcentral/xconfwebconfig/protobuf"
	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	getCpeGroupsUrlTemplate     = "%s/v2/cg/%s"
	getRfcPrecookUrlTemplate    = "%s/v2/xd/%s"
	getHashesUrlTemplate        = "%s/v2/ft/%s"
	getSecurityTokenUrlTemplate = "%s/v2/st/%s"
	getAccountIdTemplate        = "%s/v2/xac/%s"
	getAccountProductsTemplate  = "%s/v2/ada/%s"
)

type GroupServiceConnector interface {
	GroupServiceHost() string
	SetGroupServiceHost(host string)
	GroupPrefix() string
	SetGroupPrefix(prefix string)
	GetRfcPrecookDetails(cpeMac string, fields log.Fields) (*conversion.XconfDevice, error)
	GetCpeGroups(cpeMac string, fields log.Fields) ([]string, error)
	CreateListFromGroupServiceProto(cpeGroup *conversion.CpeGroup) []string
	GetFeatureTagsHashedItems(name string, fields log.Fields) (map[string]string, error)
	GetSecurityTokenInfo(securityIdentifier string, fields log.Fields) (map[string]string, error)
	GetAccountIdData(mac string, fields log.Fields) (*conversion.XBOAccount, error)
	GetAccountProducts(accountId string, fields log.Fields) (map[string]string, error)
}

type DefaultGroupService struct {
	*HttpClient
	host        string
	groupPrefix string
}

var groupServiceName string

func NewGroupServiceConnector(conf *configuration.Config, tlsConfig *tls.Config, externalGroupService GroupServiceConnector) GroupServiceConnector {
	if externalGroupService != nil {
		return externalGroupService
	} else {
		groupServiceName = conf.GetString("xconfwebconfig.xconf.group_service_name")
		confKey := fmt.Sprintf("xconfwebconfig.%v.host", groupServiceName)
		host := conf.GetString(confKey)
		if util.IsBlank(host) {
			panic(fmt.Errorf("%s is required", confKey))
		}
		groupPrefix := conf.GetString("xconfwebconfig.xconf.group_prefix")
		return &DefaultGroupService{
			HttpClient:  NewHttpClient(conf, groupServiceName, tlsConfig),
			host:        host,
			groupPrefix: groupPrefix,
		}
	}
}

func (c *DefaultGroupService) GroupServiceHost() string {
	return c.host
}

func (c *DefaultGroupService) SetGroupServiceHost(host string) {
	c.host = host
}

func (c *DefaultGroupService) GroupPrefix() string {
	return c.groupPrefix
}

func (c *DefaultGroupService) SetGroupPrefix(prefix string) {
	c.groupPrefix = prefix
}

func (c *DefaultGroupService) GetRfcPrecookDetails(cpeMac string, fields log.Fields) (*conversion.XconfDevice, error) {
	url := fmt.Sprintf(getRfcPrecookUrlTemplate, c.GroupServiceHost(), cpeMac)
	rbytes, err := c.DoWithRetries("GET", url, nil, nil, fields, groupServiceName)
	if err != nil {
		return nil, err
	}
	message := conversion.XconfDevice{}
	message.ProtoMessage()
	err = proto.Unmarshal(rbytes, &message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (c *DefaultGroupService) GetCpeGroups(cpeMac string, fields log.Fields) ([]string, error) {
	url := fmt.Sprintf(getCpeGroupsUrlTemplate, c.GroupServiceHost(), cpeMac)
	rbytes, err := c.DoWithRetries("GET", url, nil, nil, fields, groupServiceName)
	if err != nil {
		return nil, err
	}
	message := conversion.CpeGroup{}
	message.ProtoMessage()
	err = proto.Unmarshal(rbytes, &message)
	if err != nil {
		return nil, err
	}

	return c.CreateListFromGroupServiceProto(&message), nil
}

func (c *DefaultGroupService) CreateListFromGroupServiceProto(cpeGroup *conversion.CpeGroup) []string {
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

func (c *DefaultGroupService) GetFeatureTagsHashedItems(name string, fields log.Fields) (map[string]string, error) {
	url := fmt.Sprintf(getHashesUrlTemplate, c.GroupServiceHost(), name)
	rbytes, err := c.DoWithRetries("GET", url, nil, nil, fields, groupServiceName)
	if err != nil {
		return nil, err
	}
	message := conversion.XdasHashes{}
	message.ProtoMessage()
	err = proto.Unmarshal(rbytes, &message)
	if err != nil {
		return nil, err
	}
	return message.Fields, nil
}

func (c *DefaultGroupService) GetSecurityTokenInfo(securityIdentifier string, fields log.Fields) (map[string]string, error) {
	url := fmt.Sprintf(getSecurityTokenUrlTemplate, c.GroupServiceHost(), securityIdentifier)
	rbytes, err := c.DoWithRetries("GET", url, nil, nil, fields, groupServiceName)
	if err != nil {
		return nil, err
	}
	message := conversion.XdasHashes{}
	message.ProtoMessage()
	err = proto.Unmarshal(rbytes, &message)
	if err != nil {
		return nil, err
	}
	return message.Fields, nil
}

func (c *DefaultGroupService) GetAccountIdData(mac string, fields log.Fields) (*conversion.XBOAccount, error) {
	url := fmt.Sprintf(getAccountIdTemplate, c.host, mac)
	rbytes, err := c.DoWithRetries(http.MethodGet, url, nil, nil, fields, groupServiceName)
	if err != nil {
		return nil, err
	}

	var xboAccount conversion.XBOAccount
	err = proto.Unmarshal(rbytes, &xboAccount)
	if err != nil {
		return nil, err
	}

	return &xboAccount, nil
}

func (c *DefaultGroupService) GetAccountProducts(accountId string, fields log.Fields) (map[string]string, error) {
	url := fmt.Sprintf(getAccountProductsTemplate, c.GroupServiceHost(), accountId)
	rbytes, err := c.DoWithRetries(http.MethodGet, url, nil, nil, fields, groupServiceName)
	if err != nil {
		return nil, err
	}
	message := conversion.XdasHashes{}
	message.ProtoMessage()
	err = proto.Unmarshal(rbytes, &message)
	if err != nil {
		return nil, err
	}
	return message.Fields, nil
}
