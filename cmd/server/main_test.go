package main

import (
	"os"
	"testing"
)

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

func TestGetEnv_Integration(t *testing.T) {
	// Тестируем реальную переменную SERVER_PORT
	originalPort := os.Getenv("SERVER_PORT")
	defer func() {
		if originalPort != "" {
			os.Setenv("SERVER_PORT", originalPort)
		} else {
			os.Unsetenv("SERVER_PORT")
		}
	}()

	// Тест с установленной переменной
	os.Setenv("SERVER_PORT", "9090")
	result := getEnv("SERVER_PORT", "8080")
	if result != "9090" {
		t.Errorf("Expected 9090, got %s", result)
	}

	// Тест без переменной
	os.Unsetenv("SERVER_PORT")
	result = getEnv("SERVER_PORT", "8080")
	if result != "8080" {
		t.Errorf("Expected 8080, got %s", result)
	}
}
