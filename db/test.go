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
	"encoding/binary"
	"testing"
)

func TestGetShardIdSupportsIntAndStringTypes(t *testing.T) {
	int32Key := int32(12345)
	int64Key := int64(12345)
	intKey := int(12345)
	stringKey := "12345"

	if got, want := GetShardId(int32Key), GetShardId(int64Key); got != want {
		t.Fatalf("int32 and int64 shard IDs should match, got=%d want=%d", got, want)
	}

	if got, want := GetShardId(intKey), GetShardId(int64Key); got != want {
		t.Fatalf("int and int64 shard IDs should match, got=%d want=%d", got, want)
	}

	if got := GetShardId(stringKey); got < 0 || got >= ScalingFactor {
		t.Fatalf("string shard ID out of range: %d", got)
	}
}

func TestGetShardIdUnsupportedTypeDefaultsToZero(t *testing.T) {
	if got := GetShardId(true); got != 0 {
		t.Fatalf("unsupported type should map to shard 0, got=%d", got)
	}
}

func TestGetShardIdForInt64MatchesComputeShardId(t *testing.T) {
	key := int64(987654321)
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], uint64(key))

	if got, want := getShardIdForInt64(key), ComputeShardId(data[:], ScalingFactor); got != want {
		t.Fatalf("unexpected shard ID for int64 key, got=%d want=%d", got, want)
	}
}
