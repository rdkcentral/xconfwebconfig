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
package dataapi

import (
	"io"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

// init is called before any tests run, suppressing logs early
func init() {
	// Check if we're in test mode
	for _, arg := range os.Args {
		if arg == "-test.v=true" || arg == "-test.v" {
			// Keep logs visible in verbose mode
			return
		}
	}
	// Suppress logs for normal test runs
	log.SetLevel(log.PanicLevel)
	log.SetOutput(io.Discard)
}

// SuppressLogs sets the log level to Panic to suppress log output during tests
func SuppressLogs() {
	log.SetLevel(log.PanicLevel)
	log.SetOutput(io.Discard)
}

// RestoreLogs restores the default log level
func RestoreLogs() {
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stderr)
}

// TestMain can be used to suppress logs for all tests in the package
func TestMain(m *testing.M) {
	// Suppress logs before running tests
	SuppressLogs()

	// Run all tests
	exitCode := m.Run()

	// Restore logs after tests (though not really needed since process exits)
	RestoreLogs()

	// Exit with the test exit code
	os.Exit(exitCode)
}
