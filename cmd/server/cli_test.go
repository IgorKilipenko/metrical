package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags_DefaultAddress(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Устанавливаем тестовые аргументы (только имя программы)
	os.Args = []string{"server"}

	addr, err := parseFlags()
	require.NoError(t, err, "parseFlags should not return error for default address")

	expected := "localhost:8080"
	assert.Equal(t, expected, addr, "Should return default address")
}

func TestParseFlags_CustomAddress(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Устанавливаем тестовые аргументы с кастомным адресом
	os.Args = []string{"server", "-a", "localhost:9090"}

	addr, err := parseFlags()
	require.NoError(t, err, "parseFlags should not return error for valid custom address")

	expected := "localhost:9090"
	assert.Equal(t, expected, addr, "Should return custom address")
}

func TestParseFlags_InvalidAddress(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Устанавливаем тестовые аргументы с некорректным адресом
	os.Args = []string{"server", "-a", "invalid:address:format"}

	_, err := parseFlags()
	require.Error(t, err, "Expected error for invalid address")
	assert.True(t, IsInvalidAddress(err), "Expected InvalidAddressError for invalid address")
}

func TestParseFlags_UnknownArguments(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Устанавливаем тестовые аргументы с неизвестными параметрами
	os.Args = []string{"server", "unknown", "args"}

	_, err := parseFlags()
	require.Error(t, err, "Expected error for unknown arguments")

	expectedMsg := "неизвестные аргументы: [unknown args]"
	assert.Equal(t, expectedMsg, err.Error(), "Error message should match")
}

func TestParseFlags_HelpFlag(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Устанавливаем тестовые аргументы с флагом help
	os.Args = []string{"server", "--help"}

	_, err := parseFlags()
	require.Error(t, err, "Expected error for help flag")
	assert.True(t, IsHelpRequested(err), "Expected HelpRequestedError for help flag")
}

func TestParseFlags_VariousValidAddresses(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	testCases := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "IP address",
			args:     []string{"server", "-a", "127.0.0.1:9090"},
			expected: "127.0.0.1:9090",
		},
		{
			name:     "Port only",
			args:     []string{"server", "-a", "9090"},
			expected: "9090",
		},
		{
			name:     "All interfaces",
			args:     []string{"server", "-a", ":8080"},
			expected: ":8080",
		},
		{
			name:     "Long form flag",
			args:     []string{"server", "--address", "localhost:9090"},
			expected: "localhost:9090",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args

			addr, err := parseFlags()
			require.NoError(t, err, "parseFlags should not return error for valid address")
			assert.Equal(t, tc.expected, addr, "Should return expected address")
		})
	}
}

func TestParseFlags_InvalidFlagValues(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	testCases := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "Empty address flag",
			args:        []string{"server", "-a", ""},
			expectedErr: "адрес не может быть пустым",
		},
		{
			name:        "Whitespace address",
			args:        []string{"server", "-a", "   "},
			expectedErr: "некорректный формат адреса",
		},
		{
			name:        "Invalid port number",
			args:        []string{"server", "-a", "localhost:abc"},
			expectedErr: "некорректный порт",
		},
		{
			name:        "Port out of range",
			args:        []string{"server", "-a", "localhost:99999"},
			expectedErr: "некорректный порт",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args

			_, err := parseFlags()
			require.Error(t, err, "Expected error for invalid flag value")
			assert.True(t, IsInvalidAddress(err), "Should return InvalidAddressError")
			assert.Contains(t, err.Error(), tc.expectedErr, "Error message should contain expected reason")
		})
	}
}

func TestParseFlags_MultipleUnknownArguments(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	testCases := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "Single unknown argument",
			args:        []string{"server", "unknown"},
			expectedErr: "неизвестные аргументы: [unknown]",
		},
		{
			name:        "Multiple unknown arguments",
			args:        []string{"server", "arg1", "arg2", "arg3"},
			expectedErr: "неизвестные аргументы: [arg1 arg2 arg3]",
		},
		{
			name:        "Unknown arguments with flags",
			args:        []string{"server", "-a", "localhost:8080", "unknown"},
			expectedErr: "неизвестные аргументы: [unknown]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args

			_, err := parseFlags()
			require.Error(t, err, "Expected error for unknown arguments")
			assert.Equal(t, tc.expectedErr, err.Error(), "Error message should match")
		})
	}
}

func TestParseFlags_HelpVariations(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "Short help flag",
			args: []string{"server", "-h"},
		},
		{
			name: "Long help flag",
			args: []string{"server", "--help"},
		},
		{
			name: "Help with other flags",
			args: []string{"server", "-a", "localhost:8080", "--help"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args

			_, err := parseFlags()
			require.Error(t, err, "Expected error for help flag")
			assert.True(t, IsHelpRequested(err), "Expected HelpRequestedError for help flag")
		})
	}
}

func TestParseFlags_HelpFlagPanic(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	testCases := []struct {
		name        string
		args        []string
		expectError bool
		errorType   func(error) bool
	}{
		{
			name:        "No help flag",
			args:        []string{"server", "-a", "localhost:8080"},
			expectError: false,
		},
		{
			name:        "Short help flag",
			args:        []string{"server", "-h"},
			expectError: true,
			errorType:   IsHelpRequested,
		},
		{
			name:        "Long help flag",
			args:        []string{"server", "--help"},
			expectError: true,
			errorType:   IsHelpRequested,
		},
		{
			name:        "Help with other flags",
			args:        []string{"server", "-a", "localhost:8080", "--help"},
			expectError: true,
			errorType:   IsHelpRequested,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args

			// Проверяем, что не происходит паника
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("parseFlags вызвал панику: %v", r)
				}
			}()

			addr, err := parseFlags()

			if tc.expectError {
				require.Error(t, err, "Expected error")
				if tc.errorType != nil {
					assert.True(t, tc.errorType(err), "Expected specific error type")
				}
			} else {
				require.NoError(t, err, "Expected no error")
				assert.Equal(t, "localhost:8080", addr, "Should return expected address")
			}
		})
	}
}

func TestParseFlags_VersionFlag(t *testing.T) {
	// Сохраняем оригинальные аргументы
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Устанавливаем тестовые аргументы с флагом версии
	os.Args = []string{"server", "--version"}

	_, err := parseFlags()
	// Версия должна возвращать VersionRequestedError
	assert.Error(t, err, "Version flag should return error")
	assert.True(t, IsVersionRequested(err), "Expected VersionRequestedError for version flag")
}

func TestVersionVariable(t *testing.T) {
	// Проверяем, что переменная Version не пустая
	assert.NotEmpty(t, Version, "Version should not be empty")
	assert.Contains(t, Version, "dev", "Version should contain 'dev' by default")
}

func TestGetEnvOrDefault(t *testing.T) {
	// Тест с установленной переменной окружения
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnvOrDefault("TEST_VAR", "default_value")
	assert.Equal(t, "test_value", result)

	// Тест без установленной переменной окружения
	result = getEnvOrDefault("NONEXISTENT_VAR", "default_value")
	assert.Equal(t, "default_value", result)
}

func TestEnvironmentVariablePriority(t *testing.T) {
	// Сохраняем оригинальное значение переменной окружения
	originalAddress := os.Getenv("ADDRESS")

	// Восстанавливаем оригинальное значение после теста
	defer func() {
		if originalAddress != "" {
			os.Setenv("ADDRESS", originalAddress)
		} else {
			os.Unsetenv("ADDRESS")
		}
	}()

	// Устанавливаем переменную окружения
	os.Setenv("ADDRESS", "env-server:9090")

	// Симулируем значение флага командной строки
	flagValue := "flag-server:8080"

	// Проверяем, что переменная окружения имеет приоритет
	finalAddr := getEnvOrDefault("ADDRESS", flagValue)
	assert.Equal(t, "env-server:9090", finalAddr)

	// Убираем переменную окружения и проверяем, что используется флаг
	os.Unsetenv("ADDRESS")

	finalAddr = getEnvOrDefault("ADDRESS", flagValue)
	assert.Equal(t, "flag-server:8080", finalAddr)
}

func TestDefaultAddressWithEnvironment(t *testing.T) {
	// Сохраняем оригинальное значение переменной окружения
	originalAddress := os.Getenv("ADDRESS")

	// Восстанавливаем оригинальное значение после теста
	defer func() {
		if originalAddress != "" {
			os.Setenv("ADDRESS", originalAddress)
		} else {
			os.Unsetenv("ADDRESS")
		}
	}()

	// Тест с установленной переменной окружения
	os.Setenv("ADDRESS", "custom-server:9090")

	defaultAddr := getEnvOrDefault("ADDRESS", "localhost:8080")
	assert.Equal(t, "custom-server:9090", defaultAddr)

	// Тест без установленной переменной окружения
	os.Unsetenv("ADDRESS")

	defaultAddr = getEnvOrDefault("ADDRESS", "localhost:8080")
	assert.Equal(t, "localhost:8080", defaultAddr)
}
