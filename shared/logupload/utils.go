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
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rdkcentral/xconfwebconfig/common"
	util "github.com/rdkcentral/xconfwebconfig/util"

	log "github.com/sirupsen/logrus"
)

const (
	WHOLE_DAY_RANDOMIZED        = "Whole Day Randomized"
	UTC                  string = "UTC"
	LOCAL_TIME           string = "Local time"
)

func CopySettings(output *Settings, settings *Settings, rule *DCMGenericRule, context map[string]string, fields log.Fields) *Settings {
	if len(output.GroupName) < 1 && len(settings.GroupName) > 0 {
		output.CopyDeviceSettings(settings)

		output.LusScheduleCron = ""
		output.LusScheduleCronL1 = ""
		output.LusScheduleCronL2 = ""
		output.LusScheduleCronL3 = ""

		// The randomization below is used for load balancing purposes and to control the gradual rollout
		// of log upload settings to devices. This ensures that not all devices attempt to upload logs
		// simultaneously, which would potentially overwhelm the servers.
		// NOTE: This randomization is NOT suitable for security or cryptographic purposes.
		var lusSettingsCopied = false
		var randomPercentage = util.RandomPercentage()
		if randomPercentage <= rule.Percentage {
			log.Debug("This request has " + strconv.Itoa(randomPercentage) + " percentage number, which is less or equal to " + strconv.Itoa(rule.Percentage) + ". Log upload settings will be returned.")
			output.CopyLusSetting(settings, true)
			output.LusScheduleCron = settings.LusScheduleCron
			lusSettingsCopied = true
		} else {
			log.Debug("This request has " + strconv.Itoa(randomPercentage) + " percentage number, which is greater then " + strconv.Itoa(rule.Percentage) + ". Log upload settings will NOT be returned.")
			output.CopyLusSetting(settings, false)
		}
		output.RuleIDs[rule.ID] = rule.Name
		log.WithFields(common.FilterLogFields(fields)).Info("SettingsUtil Received attributes from device: " + rule.ToStringOnlyBaseProperties() + "  Applied rule for Log Upload Settings: ") //+ this.toString())

		//if timeWindow is 0 then return non-random cron expression.
		//Randomize getSettings request time cron expression.This shall return random cron expression for each request, with in the range of initial cron and time window.
		deviceSettingsCron := randomizeCronIfNecessary(output.ScheduleCron, output.ScheduleDurationMinutes, false, context, output.TimeZoneMode, "deviceSettingsCronExpression", fields)
		if len(deviceSettingsCron) > 0 {
			output.ScheduleCron = deviceSettingsCron
		}
		isDayRandomized := WHOLE_DAY_RANDOMIZED == settings.SchedulerType
		randomCronExp := randomizeCronIfNecessary(output.LusScheduleCron, output.LusScheduleDurationMinutes, isDayRandomized, context, output.LusTimeZoneMode, "logUploadCronTime", fields)
		if len(randomCronExp) > 0 {
			output.LusScheduleCron = randomCronExp
		}
		p1, _ := rule.PercentageL1.Int64()
		p2, _ := rule.PercentageL2.Int64()
		p3, _ := rule.PercentageL3.Int64()
		// Secondary randomization for distributing devices among different log upload schedule levels.
		// This implements a tiered distribution system where different percentages of devices
		// get assigned to different upload schedules (L1, L2, L3).
		// Again, this randomization is for load balancing and NOT for security purposes.
		randomPercentage = util.RandomPercentage()
		if randomPercentage <= int(p1) {
			lusScheduleCron := settings.LusScheduleCronL1
			randomCron := randomizeCronIfNecessary(lusScheduleCron, settings.LusScheduleDurationMinutes, isDayRandomized, context, output.LusTimeZoneMode, "logUploadCronL1", fields)
			if len(randomCron) > 0 {
				output.LusScheduleCronL1 = randomCron
			} else {
				output.LusScheduleCronL1 = lusScheduleCron
			}
		} else if randomPercentage <= int(p1+p2) {
			lusScheduleCron := settings.LusScheduleCronL2
			randomCron := randomizeCronIfNecessary(lusScheduleCron, settings.LusScheduleDurationMinutes, isDayRandomized, context, output.LusTimeZoneMode, "logUploadCronL2", fields)
			if len(randomCron) > 0 {
				output.LusScheduleCronL2 = randomCron
			} else {
				output.LusScheduleCronL2 = lusScheduleCron
			}
		} else if randomPercentage <= int(p1+p2+p3) {
			lusScheduleCron := settings.LusScheduleCronL3
			randomCron := randomizeCronIfNecessary(lusScheduleCron, settings.LusScheduleDurationMinutes, isDayRandomized, context, output.LusTimeZoneMode, "logUploadCronL3", fields)
			if len(randomCron) > 0 {
				output.LusScheduleCronL3 = randomCron
			} else {
				output.LusScheduleCronL3 = lusScheduleCron
			}
		}
		if !lusSettingsCopied && randomPercentage <= int(p1+p2+p3) {
			output.CopyLusSetting(settings, true)
		}
	}
	if len(output.VodSettingsName) < 1 && len(settings.VodSettingsName) > 0 {
		output.CopyVodSettings(settings)
		output.RuleIDs[rule.ID] = rule.Name
		log.WithFields(common.FilterLogFields(fields)).Info("SettingsUtil Received attributes from device: " + rule.ToStringOnlyBaseProperties() + "  Applied rule for VOD settings.")
	}
	return output
}

func randomizeCronIfNecessary(expression string, timeWindow int, isDayRandomized bool, context map[string]string, timeZone string, cronName string, fields log.Fields) string {
	var randomCronExp = ""
	estbMac := context[common.ESTB_MAC_ADDRESS]
	if isDayRandomized || (len(expression) > 0 && timeWindow > 0) {
		randomCronExp = randomizeCronEx(expression, timeWindow, isDayRandomized, context, timeZone, fields)
		if len(randomCronExp) < 1 {
			//log.Error("Invalid %s=%s for estbMac=%s", cronName, expression, estbMac)
			log.Error("Invalid {}={} for estbMac={}", cronName, expression, estbMac)
		} else {
			currentTime := time.Now().Format("2021-03-23 10:11:12")
			log.Debugf("SettingsUtil original {%s}={%s} randomized {%s}={%s} for estbMac={%s} at dcmTime={%s}", cronName, expression, cronName, randomCronExp, estbMac, currentTime)
		}
	}
	return randomCronExp
}

/**
 * Randomize the cron expression between the cron expression and upper bound as timeWindow.
 * Also depending on type random range is fixed.
 * @param expression cron expression.
 * @param timeWindow upper bound.
 * @param isDayRandomized DayRandomized/cron expression.
 * @return String randomized cron expression.
 */
func randomizeCronEx(expression string, timeWindow int, isDayRandomized bool, context map[string]string, timeZoneMode string, fields log.Fields) string {
	expressionArray := []string{"0", "0", "*", "*", "*"}
	var lowerMinutes int
	var lowerHour int
	var randomMinutes int
	if isDayRandomized {
		randomMinutes = getOffset(1440, context, fields)
	} else {
		if !validate(expression) {
			return ""
		}
		expressionArray = strings.Split(expression, " ")
		lowerMinutes, _ = strconv.Atoi(expressionArray[0])
		lowerHour, _ = strconv.Atoi(expressionArray[1])
		randomMinutes = getOffset(timeWindow, context, fields)
	}

	// Get next random hour and random minute
	newMin := lowerMinutes + randomMinutes
	// To tackle midnight boundaries.
	// If Minutes >= 60 extract out hour and add it to new hour value.
	// Being division and mod operators it will take care of while conditions,and remainder value shall
	// always be less than 60 for minutes and less than 24 or 0 for hours.
	newHr := newMin / 60
	newMin = newMin % 60
	// If new hour value is >=24 i.e.  at 00 am or more then convert to AM values i.e. 0,1 etc
	newHr = lowerHour + newHr
	if timeZoneMode == UTC {
		newHr = newHr + getAddedHoursToRandomizedCronByTimeZone(context[common.TIME_ZONE])
	}
	newHr = newHr % 24
	// As per ticket only hour and minutes need to be considered.
	var sb strings.Builder
	sb.WriteString(strconv.Itoa(newMin) + " " + strconv.Itoa(newHr))
	for i := 2; i < len(expressionArray); i++ {
		sb.WriteString(" " + expressionArray[i])
	}
	return sb.String()
}

func getOffset(timeWindow int, context map[string]string, fields log.Fields) int {
	// hash estbMac and use as seed to get random number between 0 and 1
	// so each device will always get the same cron time and we avoid multiple
	// updates in one day
	var valueToHash string
	estbMac := context[common.ESTB_MAC_ADDRESS]
	if estbMac != "" {
		valueToHash = estbMac
		fields["valueToHash"] = common.ESTB_MAC_ADDRESS
		fields[common.ESTB_MAC_ADDRESS] = estbMac
	} else {
		valueToHash = fmt.Sprint(time.Now().UnixNano())
		fields["valueToHash"] = "unix timestamp"
		fields["unix timestamp"] = valueToHash
	}
	h := md5.New()
	h.Write([]byte(valueToHash))
	hash := binary.BigEndian.Uint64(h.Sum(nil))
	randomMinutes := int(hash % uint64(timeWindow))
	fields["hashValue"] = hash
	fields["timeWindow"] = timeWindow
	fields["randomMinutes"] = randomMinutes
	return randomMinutes
}

/**
 * Validates hours and minutes section of  the cron expression.Ideally at the time of entering these details by the user it should be validated.
 *
 * @param expression  Cron expression.
 * @return  boolean for validation.
 */
func validate(expression string) bool {
	split := strings.Split(expression, " ")
	if len(split) < 2 {
		return false
	}
	minutes, err := strconv.Atoi(split[0])
	if err != nil {
		log.Error("Invalid cron expression:" + expression)
		return false
	}
	hour, err := strconv.Atoi(split[1])
	if err != nil {
		log.Error("Invalid cron expression:" + expression)
		return false
	}
	if minutes < 0 || hour < 0 {
		return false
	}
	return true
}

const (
	DEFAULT_TIME_ZONE  = "US/Eastern"
	ONE_HOUR_SECONDS   = 3600
	DEFAULT_OFFSET_ROW = -5
)

func getAddedHoursToRandomizedCronByTimeZone(timeZoneStr string) int {
	if len(timeZoneStr) < 1 {
		return 0
	}
	loc, err := time.LoadLocation(timeZoneStr)
	if err != nil {
		log.Errorf("unknown time zone(%s): %v", timeZoneStr, err)
		loc, err = time.LoadLocation(DEFAULT_TIME_ZONE)
		if err != nil {
			return 0
		}
	}
	log.Debugf("success find time zone: %s", timeZoneStr)
	now := time.Now().In(loc)
	_, offset := now.Zone() // offset in seconds east of UTC for specified TZ
	timeShift := DEFAULT_OFFSET_ROW - offset/ONE_HOUR_SECONDS

	if isDST(now) {
		// Get the raw value that is not affected by daylight saving time
		timeShift++
	}

	log.Debug("SettingsUtil incomingTimeZone=" + timeZoneStr + " matchedTimeZone=" + loc.String() + " timeShift=" + strconv.Itoa(timeShift))
	return timeShift
}

// IsDST returns true if the time given is in DST, false if not
// DST is defined as when the offset from UTC is increased
// Ref: <https://github.com/golang/go/issues/42102>
func isDST(t time.Time) bool {
	// t
	_, tOffset := t.Zone()

	// January 1
	janYear := t.Year()
	if t.Month() > 6 {
		janYear = janYear + 1
	}
	jan1Location := time.Date(janYear, 1, 1, 0, 0, 0, 0, t.Location())
	_, janOffset := jan1Location.Zone()

	// July 1
	jul1Location := time.Date(t.Year(), 7, 1, 0, 0, 0, 0, t.Location())
	_, julOffset := jul1Location.Zone()

	if tOffset == janOffset {
		return janOffset > julOffset
	}
	return julOffset > janOffset
}
