package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	assert.Equal(t, DefaultServerURL, config.ServerURL, "ServerURL should match default")
	assert.Equal(t, DefaultPollInterval, config.PollInterval, "PollInterval should match default")
	assert.Equal(t, DefaultReportInterval, config.ReportInterval, "ReportInterval should match default")
}

func TestNewConfigWithURL(t *testing.T) {
	customURL := "http://custom-server:9090"
	config := NewConfigWithURL(customURL)

	assert.Equal(t, customURL, config.ServerURL, "ServerURL should match provided URL")
	assert.Equal(t, DefaultPollInterval, config.PollInterval, "PollInterval should match default")
	assert.Equal(t, DefaultReportInterval, config.ReportInterval, "ReportInterval should match default")
	assert.Equal(t, false, config.VerboseLogging, "VerboseLogging should be false by default")
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid config",
			config:      NewConfig(),
			expectError: false,
		},
		{
			name: "Empty server URL",
			config: &Config{
				ServerURL:      "",
				PollInterval:   DefaultPollInterval,
				ReportInterval: DefaultReportInterval,
			},
			expectError: true,
			errorMsg:    "server URL cannot be empty",
		},
		{
			name: "Zero poll interval",
			config: &Config{
				ServerURL:      DefaultServerURL,
				PollInterval:   0,
				ReportInterval: DefaultReportInterval,
			},
			expectError: true,
			errorMsg:    "poll interval must be positive",
		},
		{
			name: "Negative poll interval",
			config: &Config{
				ServerURL:      DefaultServerURL,
				PollInterval:   -1 * time.Second,
				ReportInterval: DefaultReportInterval,
			},
			expectError: true,
			errorMsg:    "poll interval must be positive",
		},
		{
			name: "Zero report interval",
			config: &Config{
				ServerURL:      DefaultServerURL,
				PollInterval:   DefaultPollInterval,
				ReportInterval: 0,
			},
			expectError: true,
			errorMsg:    "report interval must be positive",
		},
		{
			name: "Negative report interval",
			config: &Config{
				ServerURL:      DefaultServerURL,
				PollInterval:   DefaultPollInterval,
				ReportInterval: -1 * time.Second,
			},
			expectError: true,
			errorMsg:    "report interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err, "Expected validation error")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

func TestConfig_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
	}{
		{
			name:        "Valid config",
			config:      NewConfig(),
			expectValid: true,
		},
		{
			name: "Invalid config - empty URL",
			config: &Config{
				ServerURL:      "",
				PollInterval:   DefaultPollInterval,
				ReportInterval: DefaultReportInterval,
			},
			expectValid: false,
		},
		{
			name: "Invalid config - zero poll interval",
			config: &Config{
				ServerURL:      DefaultServerURL,
				PollInterval:   0,
				ReportInterval: DefaultReportInterval,
			},
			expectValid: false,
		},
		{
			name: "Invalid config - zero report interval",
			config: &Config{
				ServerURL:      DefaultServerURL,
				PollInterval:   DefaultPollInterval,
				ReportInterval: 0,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.IsValid()
			assert.Equal(t, tt.expectValid, isValid, "IsValid should return expected result")
		})
	}
}
