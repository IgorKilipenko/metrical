package app

import (
	"testing"

	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedAddr string
		expectedPort string
		expectError  bool
	}{
		{
			name:         "Default address",
			input:        "localhost:8080",
			expectedAddr: "localhost",
			expectedPort: "8080",
			expectError:  false,
		},
		{
			name:         "Custom address",
			input:        "localhost:9090",
			expectedAddr: "localhost",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "Only port",
			input:        "9090",
			expectedAddr: "localhost",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "Custom host and port",
			input:        "127.0.0.1:9090",
			expectedAddr: "127.0.0.1",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "Invalid address format",
			input:        "invalid:address:format",
			expectedAddr: "invalid",
			expectedPort: "address:format",
			expectError:  false,
		},
		{
			name:         "Empty string",
			input:        "",
			expectedAddr: "localhost",
			expectedPort: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.input, 300, testutils.TestMetricsFile, true, "")

			if tt.expectError {
				assert.Error(t, err, "Expected error, got nil")
			} else {
				require.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.expectedAddr, config.Addr, "Address should match")
				assert.Equal(t, tt.expectedPort, config.Port, "Port should match")
			}
		})
	}
}

func TestParseAddr(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedAddr string
		expectedPort string
		expectError  bool
	}{
		{
			name:         "Full address",
			input:        "localhost:8080",
			expectedAddr: "localhost",
			expectedPort: "8080",
			expectError:  false,
		},
		{
			name:         "Only port",
			input:        "9090",
			expectedAddr: "localhost",
			expectedPort: "9090",
			expectError:  false,
		},
		{
			name:         "IP address",
			input:        "127.0.0.1:8080",
			expectedAddr: "127.0.0.1",
			expectedPort: "8080",
			expectError:  false,
		},
		{
			name:         "Empty string",
			input:        "",
			expectedAddr: "localhost",
			expectedPort: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, port, err := parseAddr(tt.input)

			if tt.expectError {
				assert.Error(t, err, "Expected error, got nil")
			} else {
				require.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.expectedAddr, addr, "Address should match")
				assert.Equal(t, tt.expectedPort, port, "Port should match")
			}
		})
	}
}

func TestNew(t *testing.T) {
	config := Config{Addr: "localhost", Port: "9090"}
	app := New(config)

	assert.Equal(t, "localhost:9090", app.GetPort(), "Port should be correctly formatted")
	assert.Nil(t, app.GetServer(), "Server should be nil before Run()")
}

func TestApp_GetPort(t *testing.T) {
	config := Config{Addr: "localhost", Port: "8080"}
	app := New(config)

	addr := app.GetPort()
	assert.Equal(t, "localhost:8080", addr, "GetPort() should return correctly formatted address")
}

func TestConfig_GetStorageType(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		expectedType string
		description  string
	}{
		{
			name: "PostgreSQL storage type",
			config: Config{
				DatabaseDSN: "postgres://user:pass@localhost:5432/db",
			},
			expectedType: StorageTypePostgres,
			description:  "Должен возвращать postgresql когда указан DATABASE_DSN",
		},
		{
			name: "File storage type",
			config: Config{
				FileStoragePath: "/tmp/metrics.json",
			},
			expectedType: StorageTypeFile,
			description:  "Должен возвращать file когда указан FileStoragePath",
		},
		{
			name:   "Memory storage type (default)",
			config: Config{
				// Пустая конфигурация
			},
			expectedType: StorageTypeMemory,
			description:  "Должен возвращать memory по умолчанию",
		},
		{
			name: "PostgreSQL priority over file",
			config: Config{
				DatabaseDSN:     "postgres://user:pass@localhost:5432/db",
				FileStoragePath: "/tmp/metrics.json",
			},
			expectedType: StorageTypePostgres,
			description:  "PostgreSQL должен иметь приоритет над файловым хранилищем",
		},
		{
			name: "File priority over memory",
			config: Config{
				FileStoragePath: "/tmp/metrics.json",
				// DatabaseDSN не указан
			},
			expectedType: StorageTypeFile,
			description:  "Файловое хранилище должно иметь приоритет над памятью",
		},
		{
			name: "Empty database DSN should not trigger PostgreSQL",
			config: Config{
				DatabaseDSN: "",
			},
			expectedType: StorageTypeMemory,
			description:  "Пустая строка DATABASE_DSN не должна активировать PostgreSQL",
		},
		{
			name: "Empty file path should not trigger file storage",
			config: Config{
				FileStoragePath: "",
			},
			expectedType: StorageTypeMemory,
			description:  "Пустая строка FileStoragePath не должна активировать файловое хранилище",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetStorageType()
			assert.Equal(t, tt.expectedType, result, tt.description)
		})
	}
}
