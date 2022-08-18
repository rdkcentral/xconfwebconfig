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
package rfc

import (
	"testing"

	"gotest.tools/assert"
)

func TestFeatureCreationAndMarshall(t *testing.T) {

	configData := map[string]string{
		"configKey": "configValue",
	}

	// only mandatory fields
	feature := &Feature{
		ConfigData:         configData,
		FeatureName:        "featureName",
		Name:               "name",
		Enable:             true,
		EffectiveImmediate: true,
	}
	featureResponseObject := CreateFeatureResponseObject(*feature)
	expectedJsonString := "{\"name\":\"name\",\"enable\":true,\"effectiveImmediate\":true,\"configData\":{\"configKey\":\"configValue\"},\"featureInstance\":\"featureName\"}"
	actualByteString, err := featureResponseObject.MarshalJSON()
	assert.NilError(t, err)
	assert.Equal(t, expectedJsonString, string(actualByteString))
}
