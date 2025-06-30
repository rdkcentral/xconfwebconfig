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
	"xconfwebconfig/util"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

const (
	getTagsForMacAddressUrlTemplate                     = "%s/getTagsForMacAddress/%s"
	getTagsForPartnerUrlTemplate                        = "%s/getTagsForPartner/%s"
	getTagsForPartnerAndMacAddressUrlTemplate           = "%s/getTagsForPartnerAndMacAddress/partner/%s/macaddress/%s"
	getTagsForMacAddressAndAccountUrlTemplate           = "%s/getTagsForMacAddressAndAccount/macaddress/%s/account/%s"
	getTagsForAccountUrlTemplate                        = "%s/getTagsForAccount/%s"
	getTagsForPartnerAndMacAddressAndAccountUrlTemplate = "%s/getTagsForPartnerAndMacAddressAndAccount/partner/%s/macaddress/%s/account/%s"
	getTagsForPartnerAndAccountUrlTemplate              = "%s/getTagsForPartnerAndAccount/partner/%s/account/%s"
)

type TaggingConnector interface {
	MakeGetTagsRequest(url string, token string, vargs ...log.Fields) ([]string, error)
	GetTagsForContext(contextMap map[string]string, token string, fields log.Fields) ([]string, error)
	TaggingHost() string
	SetTaggingHost(host string)
	GetTagsForMacAddress(macAddress string, token string, fields log.Fields) ([]string, error)
	GetTagsForPartner(partnerId string, token string, fields log.Fields) ([]string, error)
	GetTagsForPartnerAndMacAddress(partnerId string, macAddress string, token string, fields log.Fields) ([]string, error)
	GetTagsForMacAddressAndAccount(macAddress string, accountId string, token string, fields log.Fields) ([]string, error)
	GetTagsForAccount(accountId string, token string, fields log.Fields) ([]string, error)
	GetTagsForPartnerAndMacAddressAndAccount(partnerId string, macAddress string, accountId string, token string, fields log.Fields) ([]string, error)
	GetTagsForPartnerAndAccount(partnerId string, accountId string, token string, fields log.Fields) ([]string, error)
}

type DefaultTaggingService struct {
	*HttpClient
	host string
}

var taggingServiceName string

func NewTaggingConnector(conf *configuration.Config, tlsConfig *tls.Config, externalTagging TaggingConnector) TaggingConnector {
	if externalTagging != nil {
		return externalTagging
	} else {
		taggingServiceName = conf.GetString("xconfwebconfig.xconf.tagging_service_name")
		confKey := fmt.Sprintf("xconfwebconfig.%v.host", taggingServiceName)
		host := conf.GetString(confKey)
		if util.IsBlank(host) {
			panic(fmt.Errorf("%s is required", confKey))
		}

		return &DefaultTaggingService{
			HttpClient: NewHttpClient(conf, taggingServiceName, tlsConfig),
			host:       host,
		}
	}
}

func (c *DefaultTaggingService) TaggingHost() string {
	return c.host
}

func (c *DefaultTaggingService) SetTaggingHost(host string) {
	c.host = host
}

func (c *DefaultTaggingService) MakeGetTagsRequest(url string, token string, vargs ...log.Fields) ([]string, error) {
	var fields log.Fields
	if len(vargs) > 0 {
		fields = vargs[0]
	} else {
		fields = log.Fields{}
	}

	headers := map[string]string{
		common.HeaderAuthorization: fmt.Sprintf("Bearer %s", token),
		common.HeaderUserAgent:     common.HeaderXconfDataService,
	}
	var response []string
	rbytes, err := c.DoWithRetries("GET", url, headers, nil, fields, taggingServiceName)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(rbytes, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (c *DefaultTaggingService) GetTagsForMacAddress(macAddress string, token string, fields log.Fields) ([]string, error) {
	url := fmt.Sprintf(getTagsForMacAddressUrlTemplate, c.TaggingHost(), macAddress)
	return c.MakeGetTagsRequest(url, token, fields)
}

func (c *DefaultTaggingService) GetTagsForPartner(partnerId string, token string, fields log.Fields) ([]string, error) {
	url := fmt.Sprintf(getTagsForPartnerUrlTemplate, c.TaggingHost(), partnerId)
	return c.MakeGetTagsRequest(url, token, fields)
}

func (c *DefaultTaggingService) GetTagsForPartnerAndMacAddress(partnerId string, macAddress string, token string, fields log.Fields) ([]string, error) {
	url := fmt.Sprintf(getTagsForPartnerAndMacAddressUrlTemplate, c.TaggingHost(), partnerId, macAddress)
	return c.MakeGetTagsRequest(url, token, fields)
}

func (c *DefaultTaggingService) GetTagsForMacAddressAndAccount(macAddress string, accountId string, token string, fields log.Fields) ([]string, error) {
	url := fmt.Sprintf(getTagsForMacAddressAndAccountUrlTemplate, c.TaggingHost(), macAddress, accountId)
	return c.MakeGetTagsRequest(url, token, fields)
}

func (c *DefaultTaggingService) GetTagsForAccount(accountId string, token string, fields log.Fields) ([]string, error) {
	url := fmt.Sprintf(getTagsForAccountUrlTemplate, c.TaggingHost(), accountId)
	return c.MakeGetTagsRequest(url, token, fields)
}

func (c *DefaultTaggingService) GetTagsForPartnerAndMacAddressAndAccount(partnerId string, macAddress string, accountId string, token string, fields log.Fields) ([]string, error) {
	url := fmt.Sprintf(getTagsForPartnerAndMacAddressAndAccountUrlTemplate, c.TaggingHost(), partnerId, macAddress, accountId)
	return c.MakeGetTagsRequest(url, token, fields)
}

func (c *DefaultTaggingService) GetTagsForPartnerAndAccount(partnerId string, accountId string, token string, fields log.Fields) ([]string, error) {
	url := fmt.Sprintf(getTagsForPartnerAndAccountUrlTemplate, c.TaggingHost(), partnerId, accountId)
	return c.MakeGetTagsRequest(url, token, fields)
}

func (c *DefaultTaggingService) GetTagsForContext(contextMap map[string]string, token string, fields log.Fields) ([]string, error) {
	var macAddress string
	if contextMap[common.ESTB_MAC_ADDRESS] != "" {
		macAddress = contextMap[common.ESTB_MAC_ADDRESS]
	} else {
		macAddress = contextMap[common.ESTB_MAC]
	}
	partnerId := contextMap[common.PARTNER_ID]
	accountId := contextMap[common.ACCOUNT_ID]
	hasMacAddress := util.IsValidMacAddress(macAddress)
	hasPartnerId := partnerId != "" && !util.IsUnknownValue(partnerId)
	hasAccountId := accountId != "" && !util.IsUnknownValue(accountId)

	if hasAccountId && hasMacAddress && hasPartnerId {
		return c.GetTagsForPartnerAndMacAddressAndAccount(partnerId, macAddress, accountId, token, fields)
	}
	if hasAccountId && hasMacAddress {
		return c.GetTagsForMacAddressAndAccount(macAddress, accountId, token, fields)
	}
	if hasAccountId && hasPartnerId {
		return c.GetTagsForPartnerAndAccount(partnerId, accountId, token, fields)
	}
	if hasMacAddress && hasPartnerId {
		return c.GetTagsForPartnerAndMacAddress(partnerId, macAddress, token, fields)
	}
	if hasAccountId {
		return c.GetTagsForAccount(accountId, token, fields)
	}
	if hasPartnerId {
		return c.GetTagsForPartner(partnerId, token, fields)
	}
	if hasMacAddress {
		return c.GetTagsForMacAddress(macAddress, token, fields)
	}
	return []string{}, nil
}
