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
package db

import (
	"fmt"
)

const (
	EcmMacColumnName    = "cpe_mac"
	PodSerialColumnName = "pod_id"
)

// GetEcmMacFromPodTable Get ecmMacAdress from table cpe_mac using pod serialNum
func (c *CassandraClient) GetEcmMacFromPodTable(serialNum string) (string, error) {
	c.ConcurrentQueries <- true
	defer func() { <-c.ConcurrentQueries }()

	var ecmMac []byte

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? LIMIT 1", EcmMacColumnName, fmt.Sprintf("%s.%s", c.GetDeviceKeyspace(), c.GetDevicePodTableName()), PodSerialColumnName)
	err := c.Query(stmt, serialNum).Scan(&ecmMac)
	if err != nil {
		return "", err
	}

	return string(ecmMac), nil
}
