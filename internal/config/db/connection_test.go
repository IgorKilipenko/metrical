package db

import (
	"context"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// Test helpers

// Test constants
const (
	testDSN             = "postgres://test:test@localhost:5432/testdb?sslmode=disable"
	testDSNSimple       = "postgres://test:test@localhost:5432/testdb"
	testDSNWithPassword = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
)

// createTestConfig создает тестовую конфигурацию
func createTestConfig() Config {
	return Config{
		DSN:                testDSN,
		MaxConns:           5,
		MinConns:           1,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute * 30,
		ConnectTimeout:     time.Second * 5,
		PingTimeout:        time.Second * 2,
		HealthCheckTimeout: time.Second * 3,
	}
}

// createTestConfigWithDSN создает тестовую конфигурацию с кастомным DSN
func createTestConfigWithDSN(dsn string) Config {
	config := createTestConfig()
	config.DSN = dsn
	return config
}

// createMinimalTestConfig создает минимальную тестовую конфигурацию
func createMinimalTestConfig() Config {
	return Config{
		DSN:                testDSNSimple,
		MaxConns:           5,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute * 30,
		ConnectTimeout:     time.Second * 5,
		PingTimeout:        time.Second * 3,
		HealthCheckTimeout: time.Second * 7,
	}
}

// TestNewConnection тестирует создание нового подключения
func TestNewConnection(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid config",
			config:      createTestConfig(),
			expectError: false,
		},
		{
			name: "Invalid config - empty DSN",
			config: Config{
				DSN: "",
			},
			expectError: true,
			errorMsg:    "database DSN is required",
		},
		{
			name:        "Invalid config - negative max connections",
			config:      createTestConfigWithDSN(testDSNSimple),
			expectError: true,
			errorMsg:    "max connections must be at least 1",
		},
		{
			name: "Invalid config - min connections greater than max",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.MaxConns = 5
				config.MinConns = 10
				return config
			}(),
			expectError: true,
			errorMsg:    "min connections cannot be greater than max connections",
		},
		{
			name: "Invalid config - negative ping timeout",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.PingTimeout = -1 * time.Second
				return config
			}(),
			expectError: true,
			errorMsg:    "ping timeout must be positive",
		},
		{
			name: "Invalid config - negative health check timeout",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.HealthCheckTimeout = -1 * time.Second
				return config
			}(),
			expectError: true,
			errorMsg:    "health check timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := testutils.NewMockLogger()

			conn, err := NewConnection(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, conn)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// Для валидной конфигурации мы не можем создать реальное подключение в тестах
				// без реальной БД, поэтому проверяем только валидацию
				if err != nil {
					// Ожидаем ошибку подключения к БД, но не ошибку валидации
					assert.NotContains(t, err.Error(), "invalid database config")
				}
			}
		})
	}
}

// TestConnection_InterfaceCompliance тестирует соответствие интерфейсу
func TestConnection_InterfaceCompliance(t *testing.T) {
	config := createTestConfig()
	logger := testutils.NewMockLogger()

	// Создаем Connection только для проверки интерфейса
	conn := &Connection{
		pool:   nil, // Будет nil для тестов интерфейса
		config: config,
		logger: logger,
	}

	// Проверяем, что Connection реализует интерфейс DatabaseConnection
	var _ DatabaseConnection = conn
}

// TestConnection_ConfigIntegration тестирует интеграцию с конфигурацией
func TestConnection_ConfigIntegration(t *testing.T) {
	config := createMinimalTestConfig()

	conn := &Connection{
		pool:   nil, // Будет nil для тестов конфигурации
		config: config,
		logger: testutils.NewMockLogger(),
	}

	// Проверяем, что конфигурация правильно сохраняется
	assert.Equal(t, config.DSN, conn.config.DSN)
	assert.Equal(t, config.MaxConns, conn.config.MaxConns)
	assert.Equal(t, config.MinConns, conn.config.MinConns)
	assert.Equal(t, config.MaxConnLifetime, conn.config.MaxConnLifetime)
	assert.Equal(t, config.MaxConnIdleTime, conn.config.MaxConnIdleTime)
	assert.Equal(t, config.ConnectTimeout, conn.config.ConnectTimeout)
	assert.Equal(t, config.PingTimeout, conn.config.PingTimeout)
	assert.Equal(t, config.HealthCheckTimeout, conn.config.HealthCheckTimeout)
}

// TestConnection_ConfigValidation тестирует валидацию конфигурации
func TestConnection_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid config",
			config:      createTestConfig(),
			expectError: false,
		},
		{
			name: "Empty DSN",
			config: Config{
				DSN: "",
			},
			expectError: true,
			errorMsg:    "database DSN is required",
		},
		{
			name: "Zero max connections",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.MaxConns = 0
				return config
			}(),
			expectError: true,
			errorMsg:    "max connections must be at least 1",
		},
		{
			name: "Negative min connections",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.MinConns = -1
				return config
			}(),
			expectError: true,
			errorMsg:    "min connections must be non-negative",
		},
		{
			name: "Min connections greater than max",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.MaxConns = 5
				config.MinConns = 10
				return config
			}(),
			expectError: true,
			errorMsg:    "min connections cannot be greater than max connections",
		},
		{
			name: "Zero max connection lifetime",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.MaxConnLifetime = 0
				return config
			}(),
			expectError: true,
			errorMsg:    "max connection lifetime must be positive",
		},
		{
			name: "Zero max connection idle time",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.MaxConnIdleTime = 0
				return config
			}(),
			expectError: true,
			errorMsg:    "max connection idle time must be positive",
		},
		{
			name: "Zero connect timeout",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.ConnectTimeout = 0
				return config
			}(),
			expectError: true,
			errorMsg:    "connect timeout must be positive",
		},
		{
			name: "Zero ping timeout",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.PingTimeout = 0
				return config
			}(),
			expectError: true,
			errorMsg:    "ping timeout must be positive",
		},
		{
			name: "Zero health check timeout",
			config: func() Config {
				config := createTestConfigWithDSN(testDSNSimple)
				config.HealthCheckTimeout = 0
				return config
			}(),
			expectError: true,
			errorMsg:    "health check timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConnection_ConfigString тестирует строковое представление конфигурации
func TestConnection_ConfigString(t *testing.T) {
	config := func() Config {
		config := createMinimalTestConfig()
		config.DSN = testDSNWithPassword
		return config
	}()

	str := config.String()

	// Проверяем, что пароль замаскирован
	assert.Contains(t, str, "***")
	assert.NotContains(t, str, "password")

	// Проверяем, что остальные параметры присутствуют
	assert.Contains(t, str, "user")
	assert.Contains(t, str, "localhost:5432")
	assert.Contains(t, str, "dbname")
	assert.Contains(t, str, "MaxConns: 5")
	assert.Contains(t, str, "MinConns: 0")
	assert.Contains(t, str, "PingTimeout: 3s")
	assert.Contains(t, str, "HealthCheckTimeout: 7s")
}

// TestConnection_ConfigDefaults тестирует значения по умолчанию
func TestConnection_ConfigDefaults(t *testing.T) {
	config := DefaultConfig()

	// Проверяем значения по умолчанию
	assert.Equal(t, "postgres://metricaldb:Secret@localhost:5432/metricaldb?sslmode=disable", config.DSN)
	assert.Equal(t, int32(10), config.MaxConns)
	assert.Equal(t, int32(2), config.MinConns)
	assert.Equal(t, time.Hour, config.MaxConnLifetime)
	assert.Equal(t, time.Minute*30, config.MaxConnIdleTime)
	assert.Equal(t, time.Second*10, config.ConnectTimeout)
	assert.Equal(t, time.Second*5, config.PingTimeout)
	assert.Equal(t, time.Second*5, config.HealthCheckTimeout)
}

// TestConnection_ConfigFromEnvironment тестирует создание конфигурации из переменных окружения
func TestConnection_ConfigFromEnvironment(t *testing.T) {
	// Этот тест проверяет, что NewConfig() не падает
	// В реальных тестах мы не можем установить переменные окружения
	config := NewConfig()

	// Проверяем, что конфигурация создается
	assert.NotEmpty(t, config.DSN)
	assert.Greater(t, config.MaxConns, int32(0))
	assert.GreaterOrEqual(t, config.MinConns, int32(0))
	assert.Greater(t, config.MaxConnLifetime, time.Duration(0))
	assert.Greater(t, config.MaxConnIdleTime, time.Duration(0))
	assert.Greater(t, config.ConnectTimeout, time.Duration(0))
	assert.Greater(t, config.PingTimeout, time.Duration(0))
	assert.Greater(t, config.HealthCheckTimeout, time.Duration(0))
}

// TestConnection_ContextHandling тестирует обработку контекста
func TestConnection_ContextHandling(t *testing.T) {
	config := createTestConfig()
	logger := testutils.NewMockLogger()

	conn := &Connection{
		pool:   nil, // Будет nil для тестов контекста
		config: config,
		logger: logger,
	}

	// Тестируем, что конфигурация содержит правильные таймауты
	assert.Equal(t, time.Second*2, conn.config.PingTimeout)
	assert.Equal(t, time.Second*3, conn.config.HealthCheckTimeout)

	// Тестируем создание контекста с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), conn.config.PingTimeout)
	defer cancel()

	// Проверяем, что контекст создается корректно
	assert.NotNil(t, ctx)

	// Проверяем, что контекст может быть отменен
	cancel()
	select {
	case <-ctx.Done():
		// Ожидаемо
	default:
		t.Error("Context should be cancelled")
	}
}

// TestConnection_EdgeCases тестирует граничные случаи
func TestConnection_EdgeCases(t *testing.T) {
	t.Run("Very small timeouts", func(t *testing.T) {
		config := func() Config {
			config := createTestConfigWithDSN(testDSNSimple)
			config.MaxConns = 1
			config.MinConns = 1
			config.MaxConnLifetime = time.Millisecond
			config.MaxConnIdleTime = time.Millisecond
			config.ConnectTimeout = time.Millisecond
			config.PingTimeout = time.Millisecond
			config.HealthCheckTimeout = time.Millisecond
			return config
		}()

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("Very large timeouts", func(t *testing.T) {
		config := func() Config {
			config := createTestConfigWithDSN(testDSNSimple)
			config.MaxConns = 1000
			config.MinConns = 100
			config.MaxConnLifetime = time.Hour * 24
			config.MaxConnIdleTime = time.Hour * 12
			config.ConnectTimeout = time.Minute * 5
			config.PingTimeout = time.Minute * 2
			config.HealthCheckTimeout = time.Minute * 3
			return config
		}()

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("Min connections equals max connections", func(t *testing.T) {
		config := func() Config {
			config := createTestConfigWithDSN(testDSNSimple)
			config.MaxConns = 5
			config.MinConns = 5
			return config
		}()

		err := config.Validate()
		assert.NoError(t, err)
	})
}
