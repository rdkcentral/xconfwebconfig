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
package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func RandomDouble() float64 {
	return rand.Float64()
}

// RandomPercentage returns a random integer between 0 and 99 (inclusive).
// WARNING: This function uses math/rand which is NOT cryptographically secure.
// It should ONLY be used for non-security critical operations like load balancing or
// gradual feature rollouts, never for security purposes, authentication, or encryption.
func RandomPercentage() int {
	return rand.Intn(100)
}
