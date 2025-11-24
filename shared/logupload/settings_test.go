package logupload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Only unique tests - all Clone and constructor tests are already in telemetry_profile_test.go

// TestSettingRule_GetApplicationType tests application type getter
func TestSettingRule_GetApplicationType(t *testing.T) {
	rule := &SettingRule{ApplicationType: "xhome"}
	assert.Equal(t, "xhome", rule.GetApplicationType())

	// Test default value
	emptyRule := &SettingRule{}
	assert.Equal(t, "stb", emptyRule.GetApplicationType())
}

// TestSettingRule_GetId tests ID getter
func TestSettingRule_GetId(t *testing.T) {
	rule := &SettingRule{ID: "test-123"}
	assert.Equal(t, "test-123", rule.GetId())
}

// TestSettingRule_GetName tests name getter
func TestSettingRule_GetName(t *testing.T) {
	rule := &SettingRule{Name: "Test Rule Name"}
	assert.Equal(t, "Test Rule Name", rule.GetName())
}

// TestSettingRule_GetTemplateId tests template ID getter
func TestSettingRule_GetTemplateId(t *testing.T) {
	rule := &SettingRule{}
	assert.Equal(t, "", rule.GetTemplateId())
}

// TestSettingRule_GetRuleType tests rule type getter
func TestSettingRule_GetRuleType(t *testing.T) {
	rule := &SettingRule{}
	assert.Equal(t, "SettingRule", rule.GetRuleType())
}

// TestSettingRule_GetRule tests rule getter
func TestSettingRule_GetRule(t *testing.T) {
	rule := &SettingRule{}
	result := rule.GetRule()
	assert.NotNil(t, result)
}

// TestNewSettings tests Settings constructor
func TestNewSettings(t *testing.T) {
	settings := NewSettings(5)
	assert.NotNil(t, settings)
	assert.NotNil(t, settings.RuleIDs)
	assert.NotNil(t, settings.SrmIPList)
	assert.NotNil(t, settings.EponSettings)
	assert.NotNil(t, settings.PartnerSettings)
	assert.NotNil(t, settings.LusLogFiles)
	assert.Equal(t, 5, len(settings.LusLogFiles))
}

// TestSettings_CopyDeviceSettings tests device settings copy
func TestSettings_CopyDeviceSettings(t *testing.T) {
	source := &Settings{
		GroupName:               "TestGroup",
		CheckOnReboot:           true,
		ConfigurationServiceURL: "http://test.com",
		ScheduleCron:            "0 0 * * *",
		ScheduleDurationMinutes: 60,
		ScheduleStartDate:       "2024-01-01",
		ScheduleEndDate:         "2024-12-31",
		TimeZoneMode:            "UTC",
	}

	dest := &Settings{}
	dest.CopyDeviceSettings(source)

	assert.Equal(t, source.GroupName, dest.GroupName)
	assert.Equal(t, source.CheckOnReboot, dest.CheckOnReboot)
	assert.Equal(t, source.ConfigurationServiceURL, dest.ConfigurationServiceURL)
	assert.Equal(t, source.ScheduleCron, dest.ScheduleCron)
	assert.Equal(t, source.ScheduleDurationMinutes, dest.ScheduleDurationMinutes)
	assert.Equal(t, source.ScheduleStartDate, dest.ScheduleStartDate)
	assert.Equal(t, source.ScheduleEndDate, dest.ScheduleEndDate)
	assert.Equal(t, source.TimeZoneMode, dest.TimeZoneMode)
}

// TestSettings_CopyLusSetting tests LUS settings copy
func TestSettings_CopyLusSetting(t *testing.T) {
	source := &Settings{
		LusName:                   "TestLUS",
		LusNumberOfDay:            7,
		LusUploadRepositoryName:   "TestRepo",
		LusUploadRepositoryURL:    "http://upload.com",
		LusUploadRepositoryURLNew: "https://upload.com/new",
	}

	dest := &Settings{}
	dest.CopyLusSetting(source, true)

	assert.Equal(t, "", dest.LusMessage) // Should be cleared when setLUSSettings=true
	assert.Equal(t, source.LusName, dest.LusName)
	assert.Equal(t, source.LusNumberOfDay, dest.LusNumberOfDay)
	assert.Equal(t, source.LusUploadRepositoryName, dest.LusUploadRepositoryName)
	assert.Equal(t, source.LusUploadRepositoryURL, dest.LusUploadRepositoryURL)

	// Test with setLUSSettings = false
	dest2 := &Settings{LusMessage: "Original"}
	dest2.CopyLusSetting(source, false)
	assert.Equal(t, DEFAULT_LOG_UPLOAD_SETTINGS_MESSAGE, dest2.LusMessage) // Should be set to default message
	assert.Equal(t, "", dest2.LusName)                                     // Should be cleared
	assert.Equal(t, 0, dest2.LusNumberOfDay)                               // Should be cleared
}

// TestSettings_CopyVodSettings tests VOD settings copy
func TestSettings_CopyVodSettings(t *testing.T) {
	source := &Settings{
		VodSettingsName: "TestVOD",
		LocationUrl:     "http://location.com",
		SrmIPList:       map[string]string{"srm1": "10.0.0.1", "srm2": "10.0.0.2"},
	}

	dest := &Settings{}
	dest.CopyVodSettings(source)

	assert.Equal(t, source.VodSettingsName, dest.VodSettingsName)
	assert.Equal(t, source.LocationUrl, dest.LocationUrl)
	assert.NotNil(t, dest.SrmIPList)
	assert.Equal(t, len(source.SrmIPList), len(dest.SrmIPList))
}

// TestSettings_AreFull tests completeness check
func TestSettings_AreFull(t *testing.T) {
	// Test with all fields set
	fullSettings := &Settings{
		GroupName:       "Group1",
		LusName:         "LUS1",
		VodSettingsName: "VOD1",
	}
	assert.True(t, fullSettings.AreFull())

	// Test with missing GroupName
	settings1 := &Settings{
		LusName:         "LUS1",
		VodSettingsName: "VOD1",
	}
	assert.False(t, settings1.AreFull())

	// Test with missing LusName
	settings2 := &Settings{
		GroupName:       "Group1",
		VodSettingsName: "VOD1",
	}
	assert.False(t, settings2.AreFull())

	// Test with missing VodSettingsName
	settings3 := &Settings{
		GroupName: "Group1",
		LusName:   "LUS1",
	}
	assert.False(t, settings3.AreFull())

	// Test with all empty
	emptySettings := &Settings{}
	assert.False(t, emptySettings.AreFull())
}

// TestSettings_SetSettingProfiles tests setting profiles assignment
func TestSettings_SetSettingProfiles(t *testing.T) {
	settings := &Settings{
		EponSettings:    make(map[string]string),
		PartnerSettings: make(map[string]string),
	}

	profiles := []SettingProfiles{
		{
			SettingType: "EPON",
			Properties:  map[string]string{"epon_key1": "epon_value1"},
		},
		{
			SettingType: "PARTNER_SETTINGS",
			Properties:  map[string]string{"partner_key1": "partner_value1"},
		},
	}

	settings.SetSettingProfiles(profiles)

	assert.NotNil(t, settings.EponSettings)
	assert.Equal(t, "epon_value1", settings.EponSettings["epon_key1"])
	assert.NotNil(t, settings.PartnerSettings)
	assert.Equal(t, "partner_value1", settings.PartnerSettings["partner_key1"])

	// Test with empty profiles
	settings2 := &Settings{}
	settings2.SetSettingProfiles([]SettingProfiles{})
	// Should not panic
}

// TestCreateSettingsResponseObject tests settings response creation
func TestCreateSettingsResponseObject(t *testing.T) {
	settings := &Settings{
		GroupName:                         "TestGroup",
		CheckOnReboot:                     true,
		TimeZoneMode:                      "UTC",
		ScheduleCron:                      "0 0 * * *",
		ScheduleDurationMinutes:           60,
		LusMessage:                        "Test message",
		LusName:                           "TestLUS",
		LusNumberOfDay:                    7,
		LusUploadRepositoryName:           "TestRepo",
		LusUploadRepositoryURLNew:         "https://upload.com/new",
		LusUploadRepositoryUploadProtocol: "HTTPS",
		LusUploadRepositoryURL:            "https://upload.com",
		LusUploadOnReboot:                 true,
		UploadImmediately:                 false,
		Upload:                            true,
		LusScheduleCron:                   "0 1 * * *",
		LusScheduleCronL1:                 "0 2 * * *",
		LusScheduleCronL2:                 "0 3 * * *",
		LusScheduleCronL3:                 "0 4 * * *",
		LusTimeZoneMode:                   "EST",
		LusScheduleDurationMinutes:        120,
		VodSettingsName:                   "TestVOD",
		LocationUrl:                       "http://location.com",
		SrmIPList:                         map[string]string{"srm1": "10.0.0.1"},
		EponSettings:                      map[string]string{"epon1": "val1"},
		PartnerSettings:                   map[string]string{"partner1": "val1"},
	}

	response := CreateSettingsResponseObject(settings)

	assert.NotNil(t, response)
	assert.Equal(t, "TestGroup", response.GroupName)
	assert.Equal(t, true, response.CheckOnReboot)
	assert.Equal(t, "UTC", response.TimeZoneMode)
	assert.Equal(t, "0 0 * * *", response.ScheduleCron)
	assert.Equal(t, 60, response.ScheduleDurationMinutes)
	assert.Equal(t, "Test message", response.LusMessage)
	assert.Equal(t, "TestLUS", response.LusName)
	assert.Equal(t, 7, response.LusNumberOfDay)
	assert.Equal(t, true, response.Upload)
	assert.Equal(t, "TestVOD", response.VodSettingsName)
	assert.Equal(t, "http://location.com", response.LocationUrl)
	assert.NotNil(t, response.SrmIPList)
	assert.NotNil(t, response.EponSettings)
	assert.NotNil(t, response.PartnerSettings)
}

// TestCreateSettingsResponseObject_EmptyFields tests response with empty fields
func TestCreateSettingsResponseObject_EmptyFields(t *testing.T) {
	settings := &Settings{
		CheckOnReboot: false,
	}

	response := CreateSettingsResponseObject(settings)

	assert.NotNil(t, response)
	assert.Nil(t, response.GroupName)
	assert.Nil(t, response.ScheduleCron)
	assert.Nil(t, response.LusMessage)
	assert.Nil(t, response.LusName)
	assert.Nil(t, response.VodSettingsName)
	assert.Nil(t, response.LocationUrl)
	assert.Nil(t, response.SrmIPList)
}
