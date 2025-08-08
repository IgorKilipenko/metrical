package app

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name         string
		serverPort   string
		expectedPort string
		cleanupEnv   bool
	}{
		{
			name:         "Environment variable set",
			serverPort:   "9090",
			expectedPort: "9090",
			cleanupEnv:   true,
		},
		{
			name:         "Environment variable not set",
			serverPort:   "",
			expectedPort: "8080",
			cleanupEnv:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменную окружения
			if tt.serverPort != "" {
				os.Setenv("SERVER_PORT", tt.serverPort)
			} else {
				os.Unsetenv("SERVER_PORT")
			}

			// Очищаем переменную после теста
			if tt.cleanupEnv {
				defer os.Unsetenv("SERVER_PORT")
			}

			config := LoadConfig()
			if config.Port != tt.expectedPort {
				t.Errorf("LoadConfig() = %s, want %s", config.Port, tt.expectedPort)
			}
		})
	}
}

func TestNew(t *testing.T) {
	config := Config{Port: "9090"}
	app := New(config)

	if app.GetPort() != "9090" {
		t.Errorf("New() port = %s, want 9090", app.GetPort())
	}

	if app.GetServer() != nil {
		t.Error("New() server should be nil before Run()")
	}
}

func TestApp_GetPort(t *testing.T) {
	config := Config{Port: "8080"}
	app := New(config)

	port := app.GetPort()
	if port != "8080" {
		t.Errorf("GetPort() = %s, want 8080", port)
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "Environment variable set",
			key:          "TEST_PORT",
			defaultValue: "8080",
			envValue:     "9090",
			expected:     "9090",
		},
		{
			name:         "Environment variable not set",
			key:          "NONEXISTENT",
			defaultValue: "8080",
			envValue:     "",
			expected:     "8080",
		},
		{
			name:         "Empty environment variable",
			key:          "EMPTY_VAR",
			defaultValue: "8080",
			envValue:     "",
			expected:     "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменную окружения
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %s, want %s", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
