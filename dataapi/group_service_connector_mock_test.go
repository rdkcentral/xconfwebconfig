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
package dataapi

import (
	xhttp "github.com/rdkcentral/xconfwebconfig/http"
	conversion "github.com/rdkcentral/xconfwebconfig/protobuf"
	log "github.com/sirupsen/logrus"
)

// compile-time interface assertion
var _ xhttp.GroupServiceConnector = (*accountInfoGroupServiceConnector)(nil)

// accountInfoGroupServiceConnector is a small test double used by dataapi tests
// that need controlled account-id/account-products responses.
type accountInfoGroupServiceConnector struct {
	accountData     *conversion.XBOAccount
	accountProducts map[string]string
	accountIDErr    error
	productsErr     error
}

func (m *accountInfoGroupServiceConnector) GroupServiceHost() string {
	return ""
}

func (m *accountInfoGroupServiceConnector) SetGroupServiceHost(host string) {}

func (m *accountInfoGroupServiceConnector) GroupPrefix() string {
	return ""
}

func (m *accountInfoGroupServiceConnector) SetGroupPrefix(prefix string) {}

func (m *accountInfoGroupServiceConnector) GetRfcPrecookDetails(cpeMac string, fields log.Fields) (*conversion.XconfDevice, error) {
	return nil, nil
}

func (m *accountInfoGroupServiceConnector) GetCpeGroups(cpeMac string, fields log.Fields) ([]string, error) {
	return nil, nil
}

func (m *accountInfoGroupServiceConnector) CreateListFromGroupServiceProto(cpeGroup *conversion.CpeGroup) []string {
	return nil
}

func (m *accountInfoGroupServiceConnector) GetFeatureTagsHashedItems(name string, fields log.Fields) (map[string]string, error) {
	return nil, nil
}

func (m *accountInfoGroupServiceConnector) GetAccountIdData(mac string, fields log.Fields) (*conversion.XBOAccount, error) {
	return m.accountData, m.accountIDErr
}

func (m *accountInfoGroupServiceConnector) GetAccountProducts(accountId string, fields log.Fields) (map[string]string, error) {
	return m.accountProducts, m.productsErr
}
