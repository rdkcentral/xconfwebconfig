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
package http

import (
	"testing"
)

// Test metrics counter functions by calling them with nil metrics (safe)
// This tests the nil check paths in the functions which had 0% coverage

func TestIncreaseAccountServiceEmptyResponseCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseAccountServiceEmptyResponseCounter("testModel")
	IncreaseAccountServiceEmptyResponseCounter("")

	metrics = savedMetrics
}

func TestIncreaseReturn304FromPrecookCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseReturn304FromPrecookCounter("testPartner", "testModel")
	IncreaseReturn304FromPrecookCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseReturn304RulesEngineCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseReturn304RulesEngineCounter("testPartner", "testModel")
	IncreaseReturn304RulesEngineCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseReturn200FromPrecookCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseReturn200FromPrecookCounter("testPartner", "testModel")
	IncreaseReturn200FromPrecookCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseReturnPostProcessFromPrecookCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseReturnPostProcessFromPrecookCounter("testPartner", "testModel")
	IncreaseReturnPostProcessFromPrecookCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseReturn200RulesEngineCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseReturn200RulesEngineCounter("testPartner", "testModel")
	IncreaseReturn200RulesEngineCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseReturnPostProcessOnTheFlyCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseReturnPostProcessOnTheFlyCounter("testPartner", "testModel")
	IncreaseReturnPostProcessOnTheFlyCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseNoPrecookDataCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseNoPrecookDataCounter("testPartner", "testModel")
	IncreaseNoPrecookDataCounter("", "")

	metrics = savedMetrics
}

func TestIncreasePrecookExcludeMacListCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreasePrecookExcludeMacListCounter("testPartner", "testModel")
	IncreasePrecookExcludeMacListCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseCtxHashMismatchCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseCtxHashMismatchCounter("testPartner", "testModel")
	IncreaseCtxHashMismatchCounter("", "")

	metrics = savedMetrics
}

func TestUpdateLogCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	UpdateLogCounter("testLogType")
	UpdateLogCounter("")

	metrics = savedMetrics
}

func TestUpdateFirmwarePenetrationCounts(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	UpdateFirmwarePenetrationCounts("testPartner", "testModel", "testVersion")
	UpdateFirmwarePenetrationCounts("", "", "")

	metrics = savedMetrics
}

func TestIncreaseModelChangedCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseModelChangedCounter("testPartner", "testModel")
	IncreaseModelChangedCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseModelChangedIn200Counter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseModelChangedIn200Counter("testPartner", "testModel")
	IncreaseModelChangedIn200Counter("", "")

	metrics = savedMetrics
}

func TestIncreasePartnerChangedCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreasePartnerChangedCounter("testPartner", "testModel")
	IncreasePartnerChangedCounter("", "")

	metrics = savedMetrics
}

func TestIncreasePartnerChangedIn200Counter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreasePartnerChangedIn200Counter("testPartner", "testModel")
	IncreasePartnerChangedIn200Counter("", "")

	metrics = savedMetrics
}

func TestIncreaseFwVersionChangedCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseFwVersionChangedCounter("testPartner", "testModel")
	IncreaseFwVersionChangedCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseFirmwareVersionMismatchCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseFirmwareVersionMismatchCounter("testPartner", "testModel")
	IncreaseFirmwareVersionMismatchCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseOfferedFwVersionMatchCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseOfferedFwVersionMatchCounter("testPartner", "testModel")
	IncreaseOfferedFwVersionMatchCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseFwVersionChangedIn200Counter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseFwVersionChangedIn200Counter("testPartner", "testModel")
	IncreaseFwVersionChangedIn200Counter("", "")

	metrics = savedMetrics
}

func TestIncreaseExperienceChangedCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseExperienceChangedCounter("testPartner", "testModel")
	IncreaseExperienceChangedCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseExperienceChangedIn200Counter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseExperienceChangedIn200Counter("testPartner", "testModel")
	IncreaseExperienceChangedIn200Counter("", "")

	metrics = savedMetrics
}

func TestIncreaseAccountIdChangedCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseAccountIdChangedCounter("testPartner", "testModel")
	IncreaseAccountIdChangedCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseAccountIdChangedIn200Counter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseAccountIdChangedIn200Counter("testPartner", "testModel")
	IncreaseAccountIdChangedIn200Counter("", "")

	metrics = savedMetrics
}

func TestIncreaseIpNotInSameNetworkCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseIpNotInSameNetworkCounter("testPartner", "testModel")
	IncreaseIpNotInSameNetworkCounter("", "")

	metrics = savedMetrics
}

func TestIncreaseIpNotInSameNetworkIn200Counter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseIpNotInSameNetworkIn200Counter("testPartner", "testModel")
	IncreaseIpNotInSameNetworkIn200Counter("", "")

	metrics = savedMetrics
}

func TestIncreaseTitanEmptyResponseCounter(t *testing.T) {
	// Test with nil metrics - should not panic
	savedMetrics := metrics
	metrics = nil

	IncreaseTitanEmptyResponseCounter("testModel")
	IncreaseTitanEmptyResponseCounter("")

	metrics = savedMetrics
}
