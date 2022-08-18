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
package logupload

import (
	"strconv"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func TestGetOffset(t *testing.T) {
	macAddressList := []string{"DB:87:06:0D:63:F7", "00:1B:44:11:3A:B7", "00:1B:44:11:3A:B6", "52:26:C6:BA:22:F1", "57:16:E6:CA:85:03", "79:76:B4:5E:11:27", "F4:9C:E2:5F:08:94", "4E:47:F4:67:12:BC"}
	timeWindowList := []int{480, 420, 300, 180, 90, 60}
	offsetMatrixMacList := map[int][]int{}

	for _, timeWindow := range timeWindowList {
		offsetMatrixMacList[timeWindow] = make([]int, 8)
	}

	// test when estbMacAddress is present
	context := map[string]string{}
	for index, macAddress := range macAddressList {
		context["estbMacAddress"] = macAddress
		calculateAndCompareOffsets(t, timeWindowList, offsetMatrixMacList, index, context)
	}
	compareAllOffsets(t, offsetMatrixMacList, macAddressList)

	// test when estbMacAddress is not present
	context["estbMacAddress"] = ""
	for _, timeWindow := range timeWindowList {
		offset := getOffset(timeWindow, context, log.Fields{})
		assert.Equal(t, offset <= timeWindow, true)
		assert.Equal(t, offset >= 0, true)
		areAllSame := true
		for i := 0; i < 10; i++ {
			anotherOffset := getOffset(timeWindow, context, log.Fields{})
			// test offset is within timeWindow
			assert.Equal(t, anotherOffset <= timeWindow, true)
			assert.Equal(t, anotherOffset >= 0, true)
			// check if offset is always the same or not
			if anotherOffset != offset {
				areAllSame = false
			}
		}
		assert.Equal(t, areAllSame, false)
	}

}

func calculateAndCompareOffsets(t *testing.T, timeWindowList []int, offsetMatrixList map[int][]int, index int, context map[string]string) {
	isAllEqual := true
	for i, timeWindow := range timeWindowList {
		offsetMatrixList[timeWindow][index] = getOffset(timeWindow, context, log.Fields{})
		// check offset is within time window
		assert.Equal(t, offsetMatrixList[timeWindow][index] <= timeWindow, true)
		assert.Equal(t, offsetMatrixList[timeWindow][index] >= 0, true)
		// check that multiple calls using same timewindow and value to hash result in same time
		for j := 0; j < 10; j++ {
			assert.Equal(t, offsetMatrixList[timeWindow][index], getOffset(timeWindow, context, log.Fields{}))
		}
		// ensure values for each timeWindow aren't all the same
		if i != 0 && offsetMatrixList[timeWindowList[i]][index] != offsetMatrixList[timeWindowList[i-1]][index] {
			isAllEqual = false
		}
	}
	assert.Equal(t, isAllEqual, false)
}

func compareAllOffsets(t *testing.T, offsetMatrixList map[int][]int, list []string) {
	// ensure all values aren't equal to each other, the time window, or 0
	// (in other words, assert that getOffset returns a distribution of different values)
	for timeWindow, offsetList := range offsetMatrixList {
		isAllZero := true
		isAllEqual := true
		isAllTimeWindow := true

		for index := range offsetList {
			if offsetList[index] != 0 {
				isAllZero = false
			}
			if offsetList[index] != timeWindow {
				isAllTimeWindow = false
			}
			if index != 0 && offsetList[index] != offsetList[index-1] {
				isAllEqual = false
			}
		}
		assert.Equal(t, isAllZero, false)
		assert.Equal(t, isAllEqual, false)
		assert.Equal(t, isAllTimeWindow, false)
	}
}

func TestGetAddedHoursToRandomizedCronByTimeZone(t *testing.T) {
	addedHours := getAddedHoursToRandomizedCronByTimeZone("America/New_York")
	assert.Equal(t, addedHours, 0)
	addedHours = getAddedHoursToRandomizedCronByTimeZone("America/Costa_Rica")
	assert.Equal(t, addedHours, 1)
	addedHours = getAddedHoursToRandomizedCronByTimeZone("America/Los_Angeles")
	assert.Equal(t, addedHours, 3)
}

func TestGetRandomCronEx(t *testing.T) {
	// invalid cron expression
	expression := "hello world"
	timeWindow := 30
	context := map[string]string{}
	context["estbMacAddress"] = "DB:87:06:0D:63:F7"
	cronEx := randomizeCronEx(expression, timeWindow, false, context, "", log.Fields{})
	assert.Equal(t, cronEx, "")

	// original mins + time window not enough to bump up the hour, no timezone info
	expression = "0 0 * * *"
	timeWindow = 30
	cronEx = randomizeCronEx(expression, timeWindow, false, context, "", log.Fields{})
	cronExArray := strings.Split(cronEx, " ")
	min, err := strconv.Atoi(cronExArray[0])
	assert.NilError(t, err)
	assert.Equal(t, min <= 30, true)
	assert.Equal(t, min > 0, true)
	hour, err := strconv.Atoi(cronExArray[1])
	assert.NilError(t, err)
	assert.Equal(t, hour, 0)
	assert.Equal(t, cronExArray[2], "*")
	assert.Equal(t, cronExArray[3], "*")
	assert.Equal(t, cronExArray[4], "*")

	// original mins + time window guaranteed to bump up the hour, no timezone info
	expression = "59 0 * * *"
	timeWindow = 30
	cronEx = randomizeCronEx(expression, timeWindow, false, context, "", log.Fields{})
	cronExArray = strings.Split(cronEx, " ")
	min, err = strconv.Atoi(cronExArray[0])
	assert.NilError(t, err)
	assert.Equal(t, min <= 30, true)
	assert.Equal(t, min > 0, true)
	hour, err = strconv.Atoi(cronExArray[1])
	assert.NilError(t, err)
	assert.Equal(t, hour, 1)
	assert.Equal(t, cronExArray[2], "*")
	assert.Equal(t, cronExArray[3], "*")
	assert.Equal(t, cronExArray[4], "*")

	// bump up past midnight, no timezone info
	expression = "59 23 * * *"
	timeWindow = 30
	cronEx = randomizeCronEx(expression, timeWindow, false, context, "", log.Fields{})
	cronExArray = strings.Split(cronEx, " ")
	min, err = strconv.Atoi(cronExArray[0])
	assert.NilError(t, err)
	assert.Equal(t, min <= 30, true)
	assert.Equal(t, min > 0, true)
	hour, err = strconv.Atoi(cronExArray[1])
	assert.NilError(t, err)
	assert.Equal(t, hour, 0)
	assert.Equal(t, cronExArray[2], "*")
	assert.Equal(t, cronExArray[3], "*")
	assert.Equal(t, cronExArray[4], "*")

	// non-blank but invalid timezone, invalid timezone mode
	expression = "0 0 * * *"
	timeWindow = 30
	context["timezone"] = "asags"
	cronEx = randomizeCronEx(expression, timeWindow, false, context, "segsgd", log.Fields{})
	cronExArray = strings.Split(cronEx, " ")
	min, err = strconv.Atoi(cronExArray[0])
	assert.NilError(t, err)
	assert.Equal(t, min <= 30, true)
	assert.Equal(t, min > 0, true)
	hour, err = strconv.Atoi(cronExArray[1])
	assert.NilError(t, err)
	assert.Equal(t, hour, 0)
	assert.Equal(t, cronExArray[2], "*")
	assert.Equal(t, cronExArray[3], "*")
	assert.Equal(t, cronExArray[4], "*")

	// non-blank and valid timezone, America/Costa_Rica (no DST)
	// "Local time" timezone mode (no change)
	expression = "0 0 * * *"
	timeWindow = 30
	context["timezone"] = "America/Costa_Rica"
	cronEx = randomizeCronEx(expression, timeWindow, false, context, "Local time", log.Fields{})
	cronExArray = strings.Split(cronEx, " ")
	min, err = strconv.Atoi(cronExArray[0])
	assert.NilError(t, err)
	assert.Equal(t, min <= 30, true)
	assert.Equal(t, min > 0, true)
	hour, err = strconv.Atoi(cronExArray[1])
	assert.NilError(t, err)
	assert.Equal(t, hour, 0)
	assert.Equal(t, cronExArray[2], "*")
	assert.Equal(t, cronExArray[3], "*")
	assert.Equal(t, cronExArray[4], "*")

	// non-blank and valid timezone, America/Costa_Rica (no DST)
	// UTC mode (change timezone)
	expression = "0 0 * * *"
	timeWindow = 30
	cronEx = randomizeCronEx(expression, timeWindow, false, context, "UTC", log.Fields{})
	cronExArray = strings.Split(cronEx, " ")
	min, err = strconv.Atoi(cronExArray[0])
	assert.NilError(t, err)
	assert.Equal(t, min <= 30, true)
	assert.Equal(t, min > 0, true)
	hour, err = strconv.Atoi(cronExArray[1])
	assert.NilError(t, err)
	assert.Equal(t, hour, 1)
	assert.Equal(t, cronExArray[2], "*")
	assert.Equal(t, cronExArray[3], "*")
	assert.Equal(t, cronExArray[4], "*")

	// blank timezone, UTC mode, no change
	expression = "0 0 * * *"
	timeWindow = 30
	context["timezone"] = ""
	cronEx = randomizeCronEx(expression, timeWindow, false, context, "UTC", log.Fields{})
	cronExArray = strings.Split(cronEx, " ")
	min, err = strconv.Atoi(cronExArray[0])
	assert.NilError(t, err)
	assert.Equal(t, min <= 30, true)
	assert.Equal(t, min > 0, true)
	hour, err = strconv.Atoi(cronExArray[1])
	assert.NilError(t, err)
	assert.Equal(t, hour, 0)
	assert.Equal(t, cronExArray[2], "*")
	assert.Equal(t, cronExArray[3], "*")
	assert.Equal(t, cronExArray[4], "*")

	// blank timezone, "Local time" mode, no change
	expression = "0 0 * * *"
	timeWindow = 30
	cronEx = randomizeCronEx(expression, timeWindow, false, context, "Local time", log.Fields{})
	cronExArray = strings.Split(cronEx, " ")
	min, err = strconv.Atoi(cronExArray[0])
	assert.NilError(t, err)
	assert.Equal(t, min <= 30, true)
	assert.Equal(t, min > 0, true)
	hour, err = strconv.Atoi(cronExArray[1])
	assert.NilError(t, err)
	assert.Equal(t, hour, 0)
	assert.Equal(t, cronExArray[2], "*")
	assert.Equal(t, cronExArray[3], "*")
	assert.Equal(t, cronExArray[4], "*")

	// isDayRandomized = true
	expression = "0 0 * * *"
	timeWindow = 30
	cronEx = randomizeCronEx(expression, timeWindow, true, context, "", log.Fields{})
	cronExArray = strings.Split(cronEx, " ")
	min, err = strconv.Atoi(cronExArray[0])
	assert.NilError(t, err)
	assert.Equal(t, min <= 1400, true)
	assert.Equal(t, min > 0, true)
	hour, err = strconv.Atoi(cronExArray[1])
	assert.NilError(t, err)
	assert.Equal(t, hour <= 23, true)
	assert.Equal(t, hour >= 0, true)
	assert.Equal(t, cronExArray[2], "*")
	assert.Equal(t, cronExArray[3], "*")
	assert.Equal(t, cronExArray[4], "*")
}
