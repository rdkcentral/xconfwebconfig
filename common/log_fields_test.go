/**
* Copyright 2021 Comcast Cable Communications Management, LLC
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
	"testing"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func TestCopyLogFields(t *testing.T) {
	src := log.Fields{
		"red":    "maroon",
		"orange": "auburn",
		"yellow": "amber",
		"green":  "viridian",
		"blue":   "turquoise",
		"indigo": "sapphire",
		"violet": "purple",
	}

	c1 := CopyLogFields(src)
	assert.DeepEqual(t, src, c1)
}

func TestUpdateLogFields(t *testing.T) {
	src := log.Fields{
		"red":    "maroon",
		"orange": "auburn",
		"yellow": "amber",
		"green":  "viridian",
		"blue":   "turquoise",
		"indigo": "sapphire",
		"violet": "purple",
	}
	newfields := log.Fields{
		"pink":   "magenta",
		"silver": "gray",
		"blue":   "azure",
		"indigo": "navy",
	}
	UpdateLogFields(src, newfields)
	expected := log.Fields{
		"red":    "maroon",
		"orange": "auburn",
		"yellow": "amber",
		"green":  "viridian",
		"violet": "purple",
		"pink":   "magenta",
		"silver": "gray",
		"blue":   "azure",
		"indigo": "navy",
	}

	assert.DeepEqual(t, src, expected)
}

func TestFilterLogFields(t *testing.T) {
	fields := log.Fields{
		"key1":            "value1",
		"key2":            "value2",
		"key3":            "value3",
		"key4":            "value4",
		"key5":            "value5",
		"moneytrace":      "value6",
		"out_traceparent": "value7",
	}
	newFields := FilterLogFields(fields)
	assert.Equal(t, 5, len(newFields))

	newFields = FilterLogFields(fields, "key1", "key2")
	assert.Equal(t, 3, len(newFields))
}

func TestFieldsGetString(t *testing.T) {
	fields := log.Fields{
		"foo":   "bar",
		"hello": "world",
	}

	v1 := FieldsGetString(fields, "foo")
	assert.Equal(t, "bar", v1)

	v2 := FieldsGetString(fields, "partner_id", "comcast")
	assert.Equal(t, "comcast", v2)

	v3 := FieldsGetString(fields, "monday")
	assert.Equal(t, "", v3)
}
