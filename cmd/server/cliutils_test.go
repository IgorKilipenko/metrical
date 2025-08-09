package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionRequestedError(t *testing.T) {
	err := VersionRequestedError{}

	// Проверяем сообщение об ошибке
	assert.Equal(t, "version requested", err.Error(), "Error message should match")

	// Проверяем функцию IsVersionRequested
	assert.True(t, IsVersionRequested(err), "IsVersionRequested should return true for VersionRequestedError")

	// Проверяем с обычной ошибкой
	regularErr := &os.PathError{}
	assert.False(t, IsVersionRequested(regularErr), "IsVersionRequested should return false for regular error")
}

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
