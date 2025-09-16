package app

import (
	"context"
	"testing"
	"time"

	models "github.com/IgorKilipenko/metrical/internal/model"
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

func TestApp_validateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid config",
			config: Config{
				Addr: "localhost",
				Port: "8080",
			},
			expectError: false,
		},
		{
			name: "Empty address",
			config: Config{
				Addr: "",
				Port: "8080",
			},
			expectError: true,
			errorMsg:    "address cannot be empty",
		},
		{
			name: "Empty port",
			config: Config{
				Addr: "localhost",
				Port: "",
			},
			expectError: true,
			errorMsg:    "port cannot be empty",
		},
		{
			name: "Invalid port - too high",
			config: Config{
				Addr: "localhost",
				Port: "70000",
			},
			expectError: true,
			errorMsg:    "invalid port: 70000 (must be 1-65535)",
		},
		{
			name: "Invalid port - zero",
			config: Config{
				Addr: "localhost",
				Port: "0",
			},
			expectError: true,
			errorMsg:    "invalid port: 0 (must be 1-65535)",
		},
		{
			name: "Invalid port - negative",
			config: Config{
				Addr: "localhost",
				Port: "-1",
			},
			expectError: true,
			errorMsg:    "invalid port: -1 (must be 1-65535)",
		},
		{
			name: "Invalid port - not a number",
			config: Config{
				Addr: "localhost",
				Port: "abc",
			},
			expectError: true,
			errorMsg:    "invalid port: abc (must be 1-65535)",
		},
		{
			name: "Valid IP address",
			config: Config{
				Addr: "127.0.0.1",
				Port: "8080",
			},
			expectError: false,
		},
		{
			name: "Valid localhost",
			config: Config{
				Addr: "localhost",
				Port: "8080",
			},
			expectError: false,
		},
		{
			name: "Valid 0.0.0.0",
			config: Config{
				Addr: "0.0.0.0",
				Port: "8080",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{config: tt.config}
			err := app.validateConfig()

			if tt.expectError {
				assert.Error(t, err, "Expected error, got nil")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
			}
		})
	}
}

func TestApp_waitForServerReady(t *testing.T) {
	tests := []struct {
		name        string
		addr        string
		timeout     time.Duration
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Invalid address format",
			addr:        "invalid:address",
			timeout:     100 * time.Millisecond,
			expectError: true,
			errorMsg:    "server startup timeout",
		},
		{
			name:        "Unreachable address",
			addr:        "localhost:99999", // Invalid port
			timeout:     100 * time.Millisecond,
			expectError: true,
			errorMsg:    "server startup timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{addr: tt.addr}
			ctx := context.Background()

			err := app.waitForServerReady(ctx, tt.timeout)

			if tt.expectError {
				assert.Error(t, err, "Expected error, got nil")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
			}
		})
	}
}

func TestApp_saveMetrics(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := &mockMetricsRepository{}
	app := &App{}

	// Используем существующий мок логгера из testutils
	mockLogger := testutils.NewMockLogger()

	// Тестируем сохранение с контекстом
	app.saveMetrics(mockRepo, mockLogger, "test_context")

	// Проверяем, что SaveToFile был вызван
	assert.True(t, mockRepo.saveToFileCalled, "SaveToFile should be called")
}

// Мок репозитория для тестирования
type mockMetricsRepository struct {
	saveToFileCalled bool
	saveToFileError  error
}

func (m *mockMetricsRepository) SaveToFile() error {
	m.saveToFileCalled = true
	return m.saveToFileError
}

// Заглушки для остальных методов интерфейса MetricsRepository
func (m *mockMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	return nil
}
func (m *mockMetricsRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	return nil
}
func (m *mockMetricsRepository) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	return 0, false, nil
}
func (m *mockMetricsRepository) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	return 0, false, nil
}
func (m *mockMetricsRepository) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	return nil, nil
}
func (m *mockMetricsRepository) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	return nil, nil
}
func (m *mockMetricsRepository) LoadFromFile() error   { return nil }
func (m *mockMetricsRepository) SetSyncSave(sync bool) {}
