package db

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Проверяем значения по умолчанию
	assert.Equal(t, "", config.DSN, "DSN should be empty by default")
	assert.Equal(t, int32(DefaultMaxConns), config.MaxConns)
	assert.Equal(t, int32(DefaultMinConns), config.MinConns)
	assert.Equal(t, DefaultMaxConnLifetime, config.MaxConnLifetime)
	assert.Equal(t, DefaultMaxConnIdleTime, config.MaxConnIdleTime)
	assert.Equal(t, DefaultConnectTimeout, config.ConnectTimeout)
	assert.Equal(t, DefaultPingTimeout, config.PingTimeout)
	assert.Equal(t, DefaultHealthCheckTimeout, config.HealthCheckTimeout)
}

func TestNewConfig(t *testing.T) {
	// Сохраняем оригинальные переменные окружения
	originalEnv := make(map[string]string)
	envVars := []string{
		EnvDatabaseDSN,
		EnvDBMaxConns,
		EnvDBMinConns,
		EnvDBMaxConnLifetime,
		EnvDBMaxConnIdleTime,
		EnvDBConnectTimeout,
		EnvDBPingTimeout,
		EnvDBHealthCheckTimeout,
	}

	for _, envVar := range envVars {
		originalEnv[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}

	// Восстанавливаем переменные окружения после теста
	defer func() {
		for envVar, value := range originalEnv {
			if value != "" {
				os.Setenv(envVar, value)
			} else {
				os.Unsetenv(envVar)
			}
		}
	}()

	t.Run("default values when no env vars", func(t *testing.T) {
		config := NewConfig()

		assert.Equal(t, "", config.DSN)
		assert.Equal(t, int32(DefaultMaxConns), config.MaxConns)
		assert.Equal(t, int32(DefaultMinConns), config.MinConns)
		assert.Equal(t, DefaultMaxConnLifetime, config.MaxConnLifetime)
		assert.Equal(t, DefaultMaxConnIdleTime, config.MaxConnIdleTime)
		assert.Equal(t, DefaultConnectTimeout, config.ConnectTimeout)
		assert.Equal(t, DefaultPingTimeout, config.PingTimeout)
		assert.Equal(t, DefaultHealthCheckTimeout, config.HealthCheckTimeout)
	})

	t.Run("valid environment variables", func(t *testing.T) {
		os.Setenv(EnvDatabaseDSN, "postgres://user:pass@localhost/db")
		os.Setenv(EnvDBMaxConns, "20")
		os.Setenv(EnvDBMinConns, "5")
		os.Setenv(EnvDBMaxConnLifetime, "2h")
		os.Setenv(EnvDBMaxConnIdleTime, "1h")
		os.Setenv(EnvDBConnectTimeout, "30s")
		os.Setenv(EnvDBPingTimeout, "10s")
		os.Setenv(EnvDBHealthCheckTimeout, "15s")

		config := NewConfig()

		assert.Equal(t, "postgres://user:pass@localhost/db", config.DSN)
		assert.Equal(t, int32(20), config.MaxConns)
		assert.Equal(t, int32(5), config.MinConns)
		assert.Equal(t, 2*time.Hour, config.MaxConnLifetime)
		assert.Equal(t, time.Hour, config.MaxConnIdleTime)
		assert.Equal(t, 30*time.Second, config.ConnectTimeout)
		assert.Equal(t, 10*time.Second, config.PingTimeout)
		assert.Equal(t, 15*time.Second, config.HealthCheckTimeout)
	})

	t.Run("invalid environment variables fallback to defaults", func(t *testing.T) {
		os.Setenv(EnvDBMaxConns, "invalid")
		os.Setenv(EnvDBMinConns, "not_a_number")
		os.Setenv(EnvDBMaxConnLifetime, "invalid_duration")
		os.Setenv(EnvDBConnectTimeout, "bad_timeout")

		config := NewConfig()

		// Должны использоваться значения по умолчанию
		assert.Equal(t, int32(DefaultMaxConns), config.MaxConns)
		assert.Equal(t, int32(DefaultMinConns), config.MinConns)
		assert.Equal(t, DefaultMaxConnLifetime, config.MaxConnLifetime)
		assert.Equal(t, DefaultConnectTimeout, config.ConnectTimeout)
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				DSN:                "postgres://user:pass@localhost/db",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: false,
		},
		{
			name: "empty DSN",
			config: Config{
				DSN:                "",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "database DSN is required",
		},
		{
			name: "invalid DSN format",
			config: Config{
				DSN:                "not-a-url",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "DSN must use postgres or postgresql scheme",
		},
		{
			name: "wrong scheme",
			config: Config{
				DSN:                "mysql://user:pass@localhost/db",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "DSN must use postgres or postgresql scheme",
		},
		{
			name: "max connections too low",
			config: Config{
				DSN:                "postgres://user:pass@localhost/db",
				MaxConns:           0,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "max connections must be at least 1",
		},
		{
			name: "negative min connections",
			config: Config{
				DSN:                "postgres://user:pass@localhost/db",
				MaxConns:           10,
				MinConns:           -1,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "min connections must be non-negative",
		},
		{
			name: "min connections greater than max",
			config: Config{
				DSN:                "postgres://user:pass@localhost/db",
				MaxConns:           5,
				MinConns:           10,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "min connections cannot be greater than max connections",
		},
		{
			name: "zero max connection lifetime",
			config: Config{
				DSN:                "postgres://user:pass@localhost/db",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    0,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "max connection lifetime must be positive",
		},
		{
			name: "zero max connection idle time",
			config: Config{
				DSN:                "postgres://user:pass@localhost/db",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    0,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "max connection idle time must be positive",
		},
		{
			name: "zero connect timeout",
			config: Config{
				DSN:                "postgres://user:pass@localhost/db",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     0,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "connect timeout must be positive",
		},
		{
			name: "zero ping timeout",
			config: Config{
				DSN:                "postgres://user:pass@localhost/db",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        0,
				HealthCheckTimeout: time.Second * 5,
			},
			wantErr: true,
			errMsg:  "ping timeout must be positive",
		},
		{
			name: "zero health check timeout",
			config: Config{
				DSN:                "postgres://user:pass@localhost/db",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: 0,
			},
			wantErr: true,
			errMsg:  "health check timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_String(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "empty DSN",
			config: Config{
				DSN:                "",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			expected: "DSN: <empty>",
		},
		{
			name: "valid DSN with password masking",
			config: Config{
				DSN:                "postgres://user:secret@localhost:5432/db",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			expected: "DSN: postgres://user:%2A%2A%2A@localhost:5432/db",
		},
		{
			name: "DSN without password",
			config: Config{
				DSN:                "postgres://user@localhost:5432/db",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			expected: "DSN: postgres://user@localhost:5432/db",
		},
		{
			name: "invalid DSN",
			config: Config{
				DSN:                "://invalid",
				MaxConns:           10,
				MinConns:           2,
				MaxConnLifetime:    time.Hour,
				MaxConnIdleTime:    time.Minute * 30,
				ConnectTimeout:     time.Second * 10,
				PingTimeout:        time.Second * 5,
				HealthCheckTimeout: time.Second * 5,
			},
			expected: "DSN: <invalid>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.String()
			assert.Contains(t, result, tt.expected)
			// Проверяем, что пароль не присутствует в строке
			assert.NotContains(t, result, "secret")
		})
	}
}

func TestConfig_maskDSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		expected string
	}{
		{
			name:     "empty DSN",
			dsn:      "",
			expected: "<empty>",
		},
		{
			name:     "DSN with password",
			dsn:      "postgres://user:secret@localhost:5432/db",
			expected: "postgres://user:%2A%2A%2A@localhost:5432/db",
		},
		{
			name:     "DSN without password",
			dsn:      "postgres://user@localhost:5432/db",
			expected: "postgres://user@localhost:5432/db",
		},
		{
			name:     "invalid DSN",
			dsn:      "://invalid",
			expected: "<invalid>",
		},
		{
			name:     "DSN with complex password",
			dsn:      "postgres://user:pass@word123@localhost:5432/db",
			expected: "postgres://user:%2A%2A%2A@localhost:5432/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{DSN: tt.dsn}
			result := config.maskDSN()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConstants(t *testing.T) {
	// Проверяем, что константы имеют разумные значения
	assert.Greater(t, int(DefaultMaxConns), 0, "DefaultMaxConns should be positive")
	assert.GreaterOrEqual(t, int(DefaultMinConns), 0, "DefaultMinConns should be non-negative")
	assert.LessOrEqual(t, int(DefaultMinConns), int(DefaultMaxConns), "DefaultMinConns should be <= DefaultMaxConns")
	assert.Greater(t, DefaultMaxConnLifetime, time.Duration(0), "DefaultMaxConnLifetime should be positive")
	assert.Greater(t, DefaultMaxConnIdleTime, time.Duration(0), "DefaultMaxConnIdleTime should be positive")
	assert.Greater(t, DefaultConnectTimeout, time.Duration(0), "DefaultConnectTimeout should be positive")
	assert.Greater(t, DefaultPingTimeout, time.Duration(0), "DefaultPingTimeout should be positive")
	assert.Greater(t, DefaultHealthCheckTimeout, time.Duration(0), "DefaultHealthCheckTimeout should be positive")
}

func TestEnvironmentConstants(t *testing.T) {
	// Проверяем, что константы переменных окружения не пустые
	assert.NotEmpty(t, EnvDatabaseDSN, "EnvDatabaseDSN should not be empty")
	assert.NotEmpty(t, EnvDBMaxConns, "EnvDBMaxConns should not be empty")
	assert.NotEmpty(t, EnvDBMinConns, "EnvDBMinConns should not be empty")
	assert.NotEmpty(t, EnvDBMaxConnLifetime, "EnvDBMaxConnLifetime should not be empty")
	assert.NotEmpty(t, EnvDBMaxConnIdleTime, "EnvDBMaxConnIdleTime should not be empty")
	assert.NotEmpty(t, EnvDBConnectTimeout, "EnvDBConnectTimeout should not be empty")
	assert.NotEmpty(t, EnvDBPingTimeout, "EnvDBPingTimeout should not be empty")
	assert.NotEmpty(t, EnvDBHealthCheckTimeout, "EnvDBHealthCheckTimeout should not be empty")

	// Проверяем, что константы имеют правильные значения
	assert.Equal(t, "DATABASE_DSN", EnvDatabaseDSN)
	assert.Equal(t, "DB_MAX_CONNS", EnvDBMaxConns)
	assert.Equal(t, "DB_MIN_CONNS", EnvDBMinConns)
	assert.Equal(t, "DB_MAX_CONN_LIFETIME", EnvDBMaxConnLifetime)
	assert.Equal(t, "DB_MAX_CONN_IDLE_TIME", EnvDBMaxConnIdleTime)
	assert.Equal(t, "DB_CONNECT_TIMEOUT", EnvDBConnectTimeout)
	assert.Equal(t, "DB_PING_TIMEOUT", EnvDBPingTimeout)
	assert.Equal(t, "DB_HEALTH_CHECK_TIMEOUT", EnvDBHealthCheckTimeout)
}

func TestMessageConstants(t *testing.T) {
	// Проверяем, что константа сообщения не пустая
	assert.NotEmpty(t, MsgInvalidValueUsingDefault, "MsgInvalidValueUsingDefault should not be empty")

	// Проверяем, что константа содержит плейсхолдер
	assert.Contains(t, MsgInvalidValueUsingDefault, "%s", "MsgInvalidValueUsingDefault should contain %s placeholder")

	// Проверяем, что форматирование работает правильно
	expected := "invalid DB_MAX_CONNS value, using default"
	actual := fmt.Sprintf(MsgInvalidValueUsingDefault, EnvDBMaxConns)
	assert.Equal(t, expected, actual)
}
