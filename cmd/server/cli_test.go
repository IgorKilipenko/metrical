package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelpRequestedError(t *testing.T) {
	err := HelpRequestedError{}

	// Проверяем сообщение об ошибке
	assert.Equal(t, "help requested", err.Error(), "Error message should match")

	// Проверяем функцию IsHelpRequested
	assert.True(t, IsHelpRequested(err), "IsHelpRequested should return true for HelpRequestedError")

	// Проверяем с обычной ошибкой
	regularErr := &os.PathError{}
	assert.False(t, IsHelpRequested(regularErr), "IsHelpRequested should return false for regular error")
}

func TestInvalidAddressError(t *testing.T) {
	err := InvalidAddressError{
		Address: "invalid:address",
		Reason:  "некорректный формат",
	}

	// Проверяем сообщение об ошибке
	expectedMsg := "некорректный адрес 'invalid:address': некорректный формат"
	assert.Equal(t, expectedMsg, err.Error(), "Error message should match")

	// Проверяем функцию IsInvalidAddress
	assert.True(t, IsInvalidAddress(err), "IsInvalidAddress should return true for InvalidAddressError")

	// Проверяем с обычной ошибкой
	regularErr := &os.PathError{}
	assert.False(t, IsInvalidAddress(regularErr), "IsInvalidAddress should return false for regular error")
}

func TestInvalidAddressError_EmptyFields(t *testing.T) {
	err := InvalidAddressError{
		Address: "",
		Reason:  "",
	}

	expectedMsg := "некорректный адрес '': "
	assert.Equal(t, expectedMsg, err.Error(), "Error message should handle empty fields")
}

func TestInvalidAddressError_SpecialCharacters(t *testing.T) {
	err := InvalidAddressError{
		Address: "test:123:456",
		Reason:  "multiple:colons:in:address",
	}

	expectedMsg := "некорректный адрес 'test:123:456': multiple:colons:in:address"
	assert.Equal(t, expectedMsg, err.Error(), "Error message should handle special characters")
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "Valid full address",
			addr:    "localhost:8080",
			wantErr: false,
		},
		{
			name:    "Valid IP address",
			addr:    "127.0.0.1:9090",
			wantErr: false,
		},
		{
			name:    "Valid port only",
			addr:    "8080",
			wantErr: false,
		},
		{
			name:    "Valid all interfaces",
			addr:    ":8080",
			wantErr: false,
		},
		{
			name:    "Empty address",
			addr:    "",
			wantErr: true,
		},
		{
			name:    "Invalid format",
			addr:    "invalid:address:format",
			wantErr: true,
		},
		{
			name:    "Invalid port",
			addr:    "localhost:invalid",
			wantErr: true,
		},
		{
			name:    "Port out of range",
			addr:    "localhost:99999",
			wantErr: true,
		},
		{
			name:    "Missing port",
			addr:    "localhost",
			wantErr: true,
		},
		{
			name:    "Missing host",
			addr:    ":",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAddress(tt.addr)

			if tt.wantErr {
				assert.Error(t, err, "Expected error for invalid address")
				assert.True(t, IsInvalidAddress(err), "Should return InvalidAddressError")
			} else {
				assert.NoError(t, err, "Expected no error for valid address")
			}
		})
	}
}

func TestValidateAddress_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
		reason  string
	}{
		{
			name:    "Zero port",
			addr:    "localhost:0",
			wantErr: false, // Порт 0 считается валидным в Go
			reason:  "",
		},
		{
			name:    "Negative port",
			addr:    "localhost:-1",
			wantErr: true,
			reason:  "некорректный порт",
		},
		{
			name:    "Very large port",
			addr:    "localhost:70000",
			wantErr: true,
			reason:  "некорректный порт",
		},
		{
			name:    "Only colon",
			addr:    ":",
			wantErr: true,
			reason:  "адрес должен содержать хост или порт",
		},
		{
			name:    "Multiple colons",
			addr:    "localhost:8080:9090",
			wantErr: true,
			reason:  "некорректный формат адреса",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAddress(tt.addr)

			if tt.wantErr {
				require.Error(t, err, "Expected error for edge case")
				require.True(t, IsInvalidAddress(err), "Should return InvalidAddressError")

				// Проверяем, что сообщение об ошибке содержит ожидаемую причину
				if tt.reason != "" {
					assert.Contains(t, err.Error(), tt.reason, "Error message should contain expected reason")
				}
			} else {
				assert.NoError(t, err, "Expected no error for valid edge case")
			}
		})
	}
}

func TestValidateAddress_IPv6(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "Valid IPv6 address",
			addr:    "[::1]:8080",
			wantErr: false,
		},
		{
			name:    "Valid IPv6 with port",
			addr:    "[2001:db8::1]:9090",
			wantErr: false,
		},
		{
			name:    "Invalid IPv6 format",
			addr:    "::1:8080",
			wantErr: true,
		},
		{
			name:    "IPv6 without brackets",
			addr:    "2001:db8::1:8080",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAddress(tt.addr)

			if tt.wantErr {
				assert.Error(t, err, "Expected error for invalid IPv6 address")
				assert.True(t, IsInvalidAddress(err), "Should return InvalidAddressError")
			} else {
				assert.NoError(t, err, "Expected no error for valid IPv6 address")
			}
		})
	}
}

func TestValidateAddress_Whitespace(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "Leading whitespace",
			addr:    " localhost:8080",
			wantErr: false, // net.SplitHostPort обрабатывает это как валидный адрес
		},
		{
			name:    "Trailing whitespace",
			addr:    "localhost:8080 ",
			wantErr: true, // net.LookupPort не принимает порт с пробелом
		},
		{
			name:    "Whitespace around colon",
			addr:    "localhost : 8080",
			wantErr: false, // net.SplitHostPort обрабатывает это как валидный адрес
		},
		{
			name:    "Only whitespace",
			addr:    "   ",
			wantErr: true, // Это действительно некорректный адрес
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAddress(tt.addr)

			if tt.wantErr {
				assert.Error(t, err, "Expected error for address with whitespace")
				assert.True(t, IsInvalidAddress(err), "Should return InvalidAddressError")
			} else {
				assert.NoError(t, err, "Expected no error for valid address")
			}
		})
	}
}

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
