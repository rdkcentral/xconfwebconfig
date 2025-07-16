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
package featurecontrol

import (
	"fmt"

	"github.com/rdkcentral/xconfwebconfig/shared"
	"github.com/rdkcentral/xconfwebconfig/shared/rfc"

	log "github.com/sirupsen/logrus"
)

var GetGenericNamedListOneByTypeFunc = shared.GetGenericNamedListOneByType

func ToRfcResponse(feature *rfc.Feature) *rfc.Feature {
	if !feature.Whitelisted {
		return feature
	}
	whitelistProperty := feature.WhitelistProperty
	if whitelistProperty != nil && whitelistProperty.Value != "" && whitelistProperty.NamespacedListType != "" {
		namespacedList, err := GetGenericNamedListOneByTypeFunc(whitelistProperty.Value, whitelistProperty.NamespacedListType)
		if err != nil {
			log.Error(fmt.Sprintf("Call GetGenericNamedListOneByType error %v", err))
		}
		if namespacedList != nil {
			feature.Properties = make(map[string]interface{})
			feature.Properties[namespacedList.ID] = namespacedList.Data
			feature.ListType = feature.WhitelistProperty.TypeName
			feature.ListSize = len(namespacedList.Data)
		}
	} else {
		log.Warn(fmt.Sprintf("Whitelist property has a wrong value: %+v", *feature))
	}

	return feature
}
