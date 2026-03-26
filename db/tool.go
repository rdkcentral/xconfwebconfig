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
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/spaolacci/murmur3"
)

func GetValuesStr(length int) string {
	buffer := bytes.NewBufferString("?")
	for i := 0; i < length-1; i++ {
		buffer.WriteString(",?")
	}
	return buffer.String()
}

func GetColumnsStr(columns []string) string {
	buffer := bytes.NewBuffer([]byte{})
	for i, v := range columns {
		if i > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(v)
	}
	return buffer.String()
}

func GetSetColumnsStr(columns []string) string {
	buffer := bytes.NewBuffer([]byte{})
	for i, c := range columns {
		if i > 0 {
			buffer.WriteString(",")
		}
		s := fmt.Sprintf("%v=?", c)
		buffer.WriteString(s)
	}
	return buffer.String()
}

// GetShardId returns the shard ID for the given key which can be
// used in a partition key to distribute data across multiple nodes
func GetShardId(key interface{}) int {
	switch t := key.(type) {
	case int:
		return getShardIdForInt64(int64(t))
	case int32:
		return getShardIdForInt64(int64(t))
	case int64:
		return getShardIdForInt64(t)
	case string:
		return ComputeShardId([]byte(t), ScalingFactor)
	default:
		return 0 // default to shard 0 for unsupported key types
	}
}

// getShardIdForInt64 computes the shard ID for an int64 key by converting it to a byte slice and hashing it
func getShardIdForInt64(key int64) int {
	var data [8]byte

	// Convert the int64 to a byte slice using LittleEndian byte order.
	// It's important to be consistent in order to reproduce the same hash across different systems.
	binary.LittleEndian.PutUint64(data[:], uint64(key))
	return ComputeShardId(data[:], ScalingFactor)
}

// ComputeShardId calculates a deterministic bucket between 0 and n-1
func ComputeShardId(data []byte, n int) int {
	if n <= 1 {
		return 0
	}
	hash := murmur3.Sum32(data)

	// Standard modulo: results in 0 to N-1
	return int(hash % uint32(n))
}
