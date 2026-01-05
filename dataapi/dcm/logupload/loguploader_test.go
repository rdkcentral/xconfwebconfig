/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
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
	"encoding/json"
	"testing"

	"github.com/rdkcentral/xconfwebconfig/shared/logupload"
	"github.com/stretchr/testify/assert"
)

// TestNewLogUploadRuleBase tests the constructor
func TestNewLogUploadRuleBase(t *testing.T) {
	ruleBase := NewLogUploadRuleBase()

	assert.NotNil(t, ruleBase)
	assert.NotNil(t, ruleBase.RuleProcessorFactory)
}

// TestGetLogFileList tests the getLogFileList helper function
func TestGetLogFileList(t *testing.T) {
	// Save original function
	originalGetOneLogFileListFunc := GetOneLogFileListFunc
	defer func() { GetOneLogFileListFunc = originalGetOneLogFileListFunc }()

	t.Run("GetLogFileListWithValidData", func(t *testing.T) {
		// Mock the database call
		GetOneLogFileListFunc = func(id string) (*logupload.LogFileList, error) {
			return &logupload.LogFileList{
				Data: []*logupload.LogFile{
					{Name: "logfile1.log"},
					{Name: "logfile2.log"},
					{Name: "logfile3.log"},
				},
			}, nil
		}

		logFiles, err := getLogFileList("test-id", 100)

		assert.NoError(t, err)
		assert.NotNil(t, logFiles)
		assert.Len(t, logFiles, 3)
		assert.Equal(t, "logfile1.log", logFiles[0].Name)
		assert.Equal(t, "logfile2.log", logFiles[1].Name)
		assert.Equal(t, "logfile3.log", logFiles[2].Name)
	})

	t.Run("GetLogFileListWithMaxResultsLimit", func(t *testing.T) {
		// Mock the database call with more files than maxResults
		GetOneLogFileListFunc = func(id string) (*logupload.LogFileList, error) {
			return &logupload.LogFileList{
				Data: []*logupload.LogFile{
					{Name: "file1.log"},
					{Name: "file2.log"},
					{Name: "file3.log"},
					{Name: "file4.log"},
					{Name: "file5.log"},
				},
			}, nil
		}

		// Request only 3 files (maxResults = 3)
		logFiles, err := getLogFileList("test-id", 3)

		assert.NoError(t, err)
		assert.NotNil(t, logFiles)
		// Should return maxResults-1 = 2 files
		assert.Len(t, logFiles, 2)
		assert.Equal(t, "file1.log", logFiles[0].Name)
		assert.Equal(t, "file2.log", logFiles[1].Name)
	})

	t.Run("GetLogFileListWithNilResponse", func(t *testing.T) {
		// Mock the database call to return nil
		GetOneLogFileListFunc = func(id string) (*logupload.LogFileList, error) {
			return nil, nil
		}

		logFiles, err := getLogFileList("test-id", 100)

		assert.NoError(t, err)
		assert.Nil(t, logFiles)
	})

	t.Run("GetLogFileListWithEmptyData", func(t *testing.T) {
		// Mock the database call with empty data
		GetOneLogFileListFunc = func(id string) (*logupload.LogFileList, error) {
			return &logupload.LogFileList{
				Data: []*logupload.LogFile{},
			}, nil
		}

		logFiles, err := getLogFileList("test-id", 100)

		assert.NoError(t, err)
		assert.NotNil(t, logFiles)
		assert.Len(t, logFiles, 0)
	})

	t.Run("GetLogFileListWithExactlyMaxResults", func(t *testing.T) {
		// Mock the database call with exactly maxResults items
		GetOneLogFileListFunc = func(id string) (*logupload.LogFileList, error) {
			return &logupload.LogFileList{
				Data: []*logupload.LogFile{
					{Name: "file1.log"},
					{Name: "file2.log"},
					{Name: "file3.log"},
				},
			}, nil
		}

		// maxResults = 3, same as data length
		logFiles, err := getLogFileList("test-id", 3)

		assert.NoError(t, err)
		assert.NotNil(t, logFiles)
		// Should return maxResults-1 = 2 files
		assert.Len(t, logFiles, 2)
	})

	t.Run("GetLogFileListWithFewerThanMaxResults", func(t *testing.T) {
		// Mock the database call with fewer items than maxResults
		GetOneLogFileListFunc = func(id string) (*logupload.LogFileList, error) {
			return &logupload.LogFileList{
				Data: []*logupload.LogFile{
					{Name: "file1.log"},
					{Name: "file2.log"},
				},
			}, nil
		}

		// maxResults = 10, but only 2 files available
		logFiles, err := getLogFileList("test-id", 10)

		assert.NoError(t, err)
		assert.NotNil(t, logFiles)
		// Should return all 2 files
		assert.Len(t, logFiles, 2)
		assert.Equal(t, "file1.log", logFiles[0].Name)
		assert.Equal(t, "file2.log", logFiles[1].Name)
	})
}

// TestGetSettings tests the GetSettings method with various scenarios
func TestGetSettings(t *testing.T) {
	// Save original functions
	originalGetOneDeviceSettingsFunc := GetOneDeviceSettingsFunc
	originalGetOneLogUploadSettingsFunc := GetOneLogUploadSettingsFunc
	originalGetOneUploadRepositoryFunc := GetOneUploadRepositoryFunc
	originalGetLogFileListFunc := GetLogFileListFunc
	originalGetOneVodSettingsFunc := GetOneVodSettingsFunc
	defer func() {
		GetOneDeviceSettingsFunc = originalGetOneDeviceSettingsFunc
		GetOneLogUploadSettingsFunc = originalGetOneLogUploadSettingsFunc
		GetOneUploadRepositoryFunc = originalGetOneUploadRepositoryFunc
		GetLogFileListFunc = originalGetLogFileListFunc
		GetOneVodSettingsFunc = originalGetOneVodSettingsFunc
	}()

	t.Run("GetSettingsWithNoData", func(t *testing.T) {
		// Mock all database calls to return nil
		GetOneDeviceSettingsFunc = func(id string) *logupload.DeviceSettings { return nil }
		GetOneLogUploadSettingsFunc = func(id string) *logupload.LogUploadSettings { return nil }
		GetOneUploadRepositoryFunc = func(id string) *logupload.UploadRepository { return nil }
		GetLogFileListFunc = func(maxResults int) []*logupload.LogFile { return nil }
		GetOneVodSettingsFunc = func(id string) *logupload.VodSettings { return nil }

		ruleBase := NewLogUploadRuleBase()
		settings := ruleBase.GetSettings("test-id")

		assert.NotNil(t, settings)
		// Should have empty/default values
		assert.Equal(t, "", settings.GroupName)
		assert.Equal(t, "", settings.LusName)
		assert.Equal(t, "", settings.VodSettingsName)
	})

	t.Run("GetSettingsWithInactiveDeviceSettings", func(t *testing.T) {
		// Mock device settings as inactive
		GetOneDeviceSettingsFunc = func(id string) *logupload.DeviceSettings {
			return &logupload.DeviceSettings{
				ID:                id,
				Name:              "Test Device",
				SettingsAreActive: false, // Inactive
			}
		}
		GetOneLogUploadSettingsFunc = func(id string) *logupload.LogUploadSettings {
			return &logupload.LogUploadSettings{
				ID:                id,
				Name:              "Test Log Upload",
				AreSettingsActive: true,
			}
		}
		GetOneUploadRepositoryFunc = func(id string) *logupload.UploadRepository { return nil }
		GetLogFileListFunc = func(maxResults int) []*logupload.LogFile { return nil }
		GetOneVodSettingsFunc = func(id string) *logupload.VodSettings { return nil }

		ruleBase := NewLogUploadRuleBase()
		settings := ruleBase.GetSettings("test-id")

		assert.NotNil(t, settings)
		// Should not populate GroupName since device settings are inactive
		assert.Equal(t, "", settings.GroupName)
	})

	t.Run("GetSettingsWithInactiveLogUploadSettings", func(t *testing.T) {
		// Mock log upload settings as inactive
		GetOneDeviceSettingsFunc = func(id string) *logupload.DeviceSettings {
			return &logupload.DeviceSettings{
				ID:                id,
				Name:              "Test Device",
				SettingsAreActive: true,
			}
		}
		GetOneLogUploadSettingsFunc = func(id string) *logupload.LogUploadSettings {
			return &logupload.LogUploadSettings{
				ID:                id,
				Name:              "Test Log Upload",
				AreSettingsActive: false, // Inactive
			}
		}
		GetOneUploadRepositoryFunc = func(id string) *logupload.UploadRepository { return nil }
		GetLogFileListFunc = func(maxResults int) []*logupload.LogFile { return nil }
		GetOneVodSettingsFunc = func(id string) *logupload.VodSettings { return nil }

		ruleBase := NewLogUploadRuleBase()
		settings := ruleBase.GetSettings("test-id")

		assert.NotNil(t, settings)
		// Should not populate LusName since log upload settings are inactive
		assert.Equal(t, "", settings.LusName)
	})

	t.Run("GetSettingsWithVodSettingsOnly", func(t *testing.T) {
		// Mock only VOD settings
		GetOneDeviceSettingsFunc = func(id string) *logupload.DeviceSettings { return nil }
		GetOneLogUploadSettingsFunc = func(id string) *logupload.LogUploadSettings { return nil }
		GetOneUploadRepositoryFunc = func(id string) *logupload.UploadRepository { return nil }
		GetLogFileListFunc = func(maxResults int) []*logupload.LogFile { return nil }
		GetOneVodSettingsFunc = func(id string) *logupload.VodSettings {
			return &logupload.VodSettings{
				ID:           id,
				Name:         "Test VOD",
				LocationsURL: "http://vod.example.com",
				SrmIPList:    map[string]string{"ip1": "192.168.1.1", "ip2": "192.168.1.2"},
			}
		}

		ruleBase := NewLogUploadRuleBase()
		settings := ruleBase.GetSettings("test-id")

		assert.NotNil(t, settings)
		assert.Equal(t, "Test VOD", settings.VodSettingsName)
		assert.Equal(t, "http://vod.example.com", settings.LocationUrl)
		assert.Len(t, settings.SrmIPList, 2)
		assert.Equal(t, "192.168.1.1", settings.SrmIPList["ip1"])
		assert.Equal(t, "192.168.1.2", settings.SrmIPList["ip2"])
	})

	t.Run("GetSettingsWithAllActiveSettings", func(t *testing.T) {
		// Mock all settings as active with valid data
		GetOneDeviceSettingsFunc = func(id string) *logupload.DeviceSettings {
			return &logupload.DeviceSettings{
				ID:                id,
				Name:              "Test Device Group",
				SettingsAreActive: true,
				CheckOnReboot:     true,
				ConfigurationServiceURL: logupload.ConfigurationServiceURL{
					URL: "http://config.example.com",
				},
				Schedule: logupload.Schedule{
					Type:              "CronExpression",
					Expression:        "0 0 * * *",
					TimeZone:          "UTC",
					TimeWindowMinutes: json.Number("60"),
				},
			}
		}
		GetOneLogUploadSettingsFunc = func(id string) *logupload.LogUploadSettings {
			return &logupload.LogUploadSettings{
				ID:                 id,
				Name:               "Test Log Upload",
				AreSettingsActive:  true,
				NumberOfDays:       7,
				UploadRepositoryID: "repo-123",
				UploadOnReboot:     true,
				Schedule: logupload.Schedule{
					Type:              "CronExpression",
					Expression:        "0 0 * * *",
					ExpressionL1:      "0 1 * * *",
					ExpressionL2:      "0 2 * * *",
					ExpressionL3:      "0 3 * * *",
					TimeZone:          "LOCAL_TIME",
					TimeWindowMinutes: json.Number("30"),
				},
			}
		}
		GetOneUploadRepositoryFunc = func(id string) *logupload.UploadRepository {
			return &logupload.UploadRepository{
				ID:       id,
				Name:     "Test Repo",
				URL:      "upload.example.com",
				Protocol: "HTTPS",
			}
		}
		GetLogFileListFunc = func(maxResults int) []*logupload.LogFile { return nil }
		GetOneVodSettingsFunc = func(id string) *logupload.VodSettings { return nil }

		ruleBase := NewLogUploadRuleBase()
		settings := ruleBase.GetSettings("test-id")

		assert.NotNil(t, settings)

		// Device settings assertions
		assert.Equal(t, "Test Device Group", settings.GroupName)
		assert.True(t, settings.CheckOnReboot)
		assert.Equal(t, "http://config.example.com", settings.ConfigurationServiceURL)
		assert.Equal(t, "UTC", settings.TimeZoneMode)
		assert.Equal(t, "0 0 * * *", settings.ScheduleCron)
		assert.Equal(t, 60, settings.ScheduleDurationMinutes)

		// Log upload settings assertions
		assert.Equal(t, "Test Log Upload", settings.LusName)
		assert.Equal(t, 7, settings.LusNumberOfDay)
		assert.True(t, settings.LusUploadOnReboot)
		assert.Equal(t, "LOCAL_TIME", settings.LusTimeZoneMode)
		assert.Equal(t, "0 0 * * *", settings.LusScheduleCron)
		assert.Equal(t, "0 1 * * *", settings.LusScheduleCronL1)
		assert.Equal(t, "0 2 * * *", settings.LusScheduleCronL2)
		assert.Equal(t, "0 3 * * *", settings.LusScheduleCronL3)
		assert.Equal(t, 30, settings.LusScheduleDurationMinutes)

		// Upload repository assertions
		assert.Equal(t, "Test Repo", settings.LusUploadRepositoryName)
		assert.Equal(t, "https://upload.example.com", settings.LusUploadRepositoryURL)
		assert.Equal(t, "upload.example.com", settings.LusUploadRepositoryURLNew)
		assert.Equal(t, "HTTPS", settings.LusUploadRepositoryUploadProtocol)
	})

	t.Run("GetSettingsWithUploadRepositoryContainingProtocol", func(t *testing.T) {
		// Test URL that already contains protocol
		GetOneDeviceSettingsFunc = func(id string) *logupload.DeviceSettings {
			return &logupload.DeviceSettings{
				ID:                id,
				Name:              "Device",
				SettingsAreActive: true,
				Schedule: logupload.Schedule{
					Type:              "CronExpression",
					Expression:        "0 0 * * *",
					TimeWindowMinutes: json.Number("60"),
				},
			}
		}
		GetOneLogUploadSettingsFunc = func(id string) *logupload.LogUploadSettings {
			return &logupload.LogUploadSettings{
				ID:                 id,
				Name:               "Log Upload",
				AreSettingsActive:  true,
				UploadRepositoryID: "repo-123",
				Schedule: logupload.Schedule{
					Type:              "CronExpression",
					Expression:        "0 0 * * *",
					TimeWindowMinutes: json.Number("30"),
				},
			}
		}
		GetOneUploadRepositoryFunc = func(id string) *logupload.UploadRepository {
			return &logupload.UploadRepository{
				ID:       id,
				Name:     "Repo",
				URL:      "https://upload.example.com/path",
				Protocol: "HTTPS",
			}
		}
		GetLogFileListFunc = func(maxResults int) []*logupload.LogFile { return nil }
		GetOneVodSettingsFunc = func(id string) *logupload.VodSettings { return nil }

		ruleBase := NewLogUploadRuleBase()
		settings := ruleBase.GetSettings("test-id")

		assert.NotNil(t, settings)
		// URL already contains "://", should use as-is
		assert.Equal(t, "https://upload.example.com/path", settings.LusUploadRepositoryURL)
	})
}
