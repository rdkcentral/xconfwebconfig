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
package rulesengine

import (
	"github.com/rdkcentral/xconfwebconfig/db"
	"github.com/rdkcentral/xconfwebconfig/shared"

	log "github.com/sirupsen/logrus"
)

const (
	ipListTableName = "GenericXconfNamedList"
)

type IpAddressEvaluator struct {
	freeArgType string
	operation   string
	nsListDao   db.CachedSimpleDao
}

func NewIpAddressEvaluator(freeArgType string, operation string, nsListDao db.CachedSimpleDao) *IpAddressEvaluator {
	return &IpAddressEvaluator{
		freeArgType: freeArgType,
		operation:   operation,
		nsListDao:   nsListDao,
	}
}

func (e *IpAddressEvaluator) FreeArgType() string {
	return e.freeArgType
}

func (e *IpAddressEvaluator) Operation() string {
	return e.operation
}

func (e *IpAddressEvaluator) Evaluate(condition *Condition, context map[string]string) bool {
	var freeArgValue string
	var ok bool
	if e.freeArgType != StandardFreeArgTypeVoid {
		freeArgValue, ok = context[condition.GetFreeArg().GetName()]
		if !ok || (e.freeArgType != StandardFreeArgTypeAny && len(freeArgValue) == 0) {
			return false
		}
	}

	fixedArg := condition.GetFixedArg()
	if fixedArg == nil {
		return false
	}
	fixedArgValue := fixedArg.GetValue().(string)
	if len(fixedArgValue) == 0 {
		return false
	}

	// ==== eval core ====
	// Get data from cache only and avoid loading from DB due to performance,
	// When a new record is added to the DB, the cache will be updated via CacheRefreshTask
	var GetOneFunc func(tableName string, rowKey string) (interface{}, error)
	if db.Conf.GetBoolean("xconfwebconfig.xconf.evaluator_nslist_loading_cache_enabled") {
		GetOneFunc = e.nsListDao.GetOne
	} else {
		GetOneFunc = e.nsListDao.GetOneFromCacheOnly
	}

	nsListItf, err := GetOneFunc(ipListTableName, fixedArgValue)
	if err != nil {
		log.Debugf("NsListInEvaluator  Can't evaluate rule because NsList doesn't exist. ID: %v", fixedArgValue)
		return false
	}
	nsList, ok := nsListItf.(*shared.GenericNamespacedList)
	if !ok {
		return false
	}

	// TODO change to a method
	if nsList.TypeName == shared.MacList {
		for _, v := range nsList.Data {
			if v == freeArgValue {
				return true
			}
		}
		return false
	}

	for _, s := range nsList.Data {
		ipAddr := shared.NewIpAddress(s)
		if ipAddr != nil {
			if ipAddr.IsInRange(freeArgValue) {
				return true
			}
		} else {
			if s == freeArgValue {
				return true
			}
		}
	}
	return false
}
