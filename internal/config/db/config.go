package db

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
)

// Глобальный логгер для конфигурации
var configLogger = logger.NewSlogLogger()

// Константы для имен переменных окружения
const (
	EnvDatabaseDSN          = "DATABASE_DSN"
	EnvDBMaxConns           = "DB_MAX_CONNS"
	EnvDBMinConns           = "DB_MIN_CONNS"
	EnvDBMaxConnLifetime    = "DB_MAX_CONN_LIFETIME"
	EnvDBMaxConnIdleTime    = "DB_MAX_CONN_IDLE_TIME"
	EnvDBConnectTimeout     = "DB_CONNECT_TIMEOUT"
	EnvDBPingTimeout        = "DB_PING_TIMEOUT"
	EnvDBHealthCheckTimeout = "DB_HEALTH_CHECK_TIMEOUT"
)

// Константы для значений по умолчанию
const (
	DefaultMaxConns           = 10
	DefaultMinConns           = 2
	DefaultMaxConnLifetime    = time.Hour
	DefaultMaxConnIdleTime    = time.Minute * 30
	DefaultConnectTimeout     = time.Second * 10
	DefaultPingTimeout        = time.Second * 5
	DefaultHealthCheckTimeout = time.Second * 5
)

// Константы для сообщений об ошибках
const (
	MsgInvalidValueUsingDefault = "invalid %s value, using default"
)

// Config содержит конфигурацию подключения к базе данных
type Config struct {
	DSN                string        // Data Source Name для подключения к БД
	MaxConns           int32         // Максимальное количество соединений в пуле
	MinConns           int32         // Минимальное количество соединений в пуле
	MaxConnLifetime    time.Duration // Максимальное время жизни соединения
	MaxConnIdleTime    time.Duration // Максимальное время простоя соединения
	ConnectTimeout     time.Duration // Таймаут подключения
	PingTimeout        time.Duration // Таймаут для ping операций
	HealthCheckTimeout time.Duration // Таймаут для health check операций
}

// DefaultConfig возвращает конфигурацию по умолчанию
// DSN должен быть установлен через переменную окружения DATABASE_DSN
func DefaultConfig() Config {
	return Config{
		DSN:                "", // DSN должен быть установлен через переменную окружения
		MaxConns:           DefaultMaxConns,
		MinConns:           DefaultMinConns,
		MaxConnLifetime:    DefaultMaxConnLifetime,
		MaxConnIdleTime:    DefaultMaxConnIdleTime,
		ConnectTimeout:     DefaultConnectTimeout,
		PingTimeout:        DefaultPingTimeout,
		HealthCheckTimeout: DefaultHealthCheckTimeout,
	}
}

// NewConfig создает конфигурацию из переменных окружения и значений по умолчанию
func NewConfig() Config {
	config := DefaultConfig()

	// Получаем DSN из переменной окружения
	if dsn, exists := os.LookupEnv(EnvDatabaseDSN); exists {
		config.DSN = dsn
	}

	// Получаем настройки пула соединений из переменных окружения
	if maxConns, exists := os.LookupEnv(EnvDBMaxConns); exists {
		if val, err := strconv.ParseInt(maxConns, 10, 32); err == nil {
			config.MaxConns = int32(val)
		} else {
			configLogger.Warn(fmt.Sprintf(MsgInvalidValueUsingDefault, EnvDBMaxConns), "value", maxConns, "error", err, "default", DefaultMaxConns)
		}
	}

	if minConns, exists := os.LookupEnv(EnvDBMinConns); exists {
		if val, err := strconv.ParseInt(minConns, 10, 32); err == nil {
			config.MinConns = int32(val)
		} else {
			configLogger.Warn(fmt.Sprintf(MsgInvalidValueUsingDefault, EnvDBMinConns), "value", minConns, "error", err, "default", DefaultMinConns)
		}
	}

	if maxConnLifetime, exists := os.LookupEnv(EnvDBMaxConnLifetime); exists {
		if val, err := time.ParseDuration(maxConnLifetime); err == nil {
			config.MaxConnLifetime = val
		} else {
			configLogger.Warn(fmt.Sprintf(MsgInvalidValueUsingDefault, EnvDBMaxConnLifetime), "value", maxConnLifetime, "error", err, "default", DefaultMaxConnLifetime)
		}
	}

	if maxConnIdleTime, exists := os.LookupEnv(EnvDBMaxConnIdleTime); exists {
		if val, err := time.ParseDuration(maxConnIdleTime); err == nil {
			config.MaxConnIdleTime = val
		} else {
			configLogger.Warn(fmt.Sprintf(MsgInvalidValueUsingDefault, EnvDBMaxConnIdleTime), "value", maxConnIdleTime, "error", err, "default", DefaultMaxConnIdleTime)
		}
	}

	if connectTimeout, exists := os.LookupEnv(EnvDBConnectTimeout); exists {
		if val, err := time.ParseDuration(connectTimeout); err == nil {
			config.ConnectTimeout = val
		} else {
			configLogger.Warn(fmt.Sprintf(MsgInvalidValueUsingDefault, EnvDBConnectTimeout), "value", connectTimeout, "error", err, "default", DefaultConnectTimeout)
		}
	}

	if pingTimeout, exists := os.LookupEnv(EnvDBPingTimeout); exists {
		if val, err := time.ParseDuration(pingTimeout); err == nil {
			config.PingTimeout = val
		} else {
			configLogger.Warn(fmt.Sprintf(MsgInvalidValueUsingDefault, EnvDBPingTimeout), "value", pingTimeout, "error", err, "default", DefaultPingTimeout)
		}
	}

	if healthCheckTimeout, exists := os.LookupEnv(EnvDBHealthCheckTimeout); exists {
		if val, err := time.ParseDuration(healthCheckTimeout); err == nil {
			config.HealthCheckTimeout = val
		} else {
			configLogger.Warn(fmt.Sprintf(MsgInvalidValueUsingDefault, EnvDBHealthCheckTimeout), "value", healthCheckTimeout, "error", err, "default", DefaultHealthCheckTimeout)
		}
	}

	return config
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	// Проверяем формат DSN
	u, err := url.Parse(c.DSN)
	if err != nil {
		return fmt.Errorf("invalid DSN format: %w", err)
	}

	// Проверяем, что это PostgreSQL DSN
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf("DSN must use postgres or postgresql scheme, got: %s", u.Scheme)
	}

	if c.MaxConns < 1 {
		return fmt.Errorf("max connections must be at least 1")
	}

	if c.MinConns < 0 {
		return fmt.Errorf("min connections must be non-negative")
	}

	if c.MinConns > c.MaxConns {
		return fmt.Errorf("min connections cannot be greater than max connections")
	}

	if c.MaxConnLifetime <= 0 {
		return fmt.Errorf("max connection lifetime must be positive")
	}

	if c.MaxConnIdleTime <= 0 {
		return fmt.Errorf("max connection idle time must be positive")
	}

	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("connect timeout must be positive")
	}

	if c.PingTimeout <= 0 {
		return fmt.Errorf("ping timeout must be positive")
	}

	if c.HealthCheckTimeout <= 0 {
		return fmt.Errorf("health check timeout must be positive")
	}

	return nil
}

// String возвращает строковое представление конфигурации (без пароля)
func (c *Config) String() string {
	// Безопасно маскируем пароль в DSN для логирования
	maskedDSN := c.maskDSN()

	return fmt.Sprintf("Config{DSN: %s, MaxConns: %d, MinConns: %d, MaxConnLifetime: %v, MaxConnIdleTime: %v, ConnectTimeout: %v, PingTimeout: %v, HealthCheckTimeout: %v}",
		maskedDSN, c.MaxConns, c.MinConns, c.MaxConnLifetime, c.MaxConnIdleTime, c.ConnectTimeout, c.PingTimeout, c.HealthCheckTimeout)
}

// maskDSN безопасно маскирует пароль в DSN для логирования
func (c *Config) maskDSN() string {
	if c.DSN == "" {
		return "<empty>"
	}

	// Парсим URL для надежной маскировки
	u, err := url.Parse(c.DSN)
	if err != nil {
		return "<invalid>"
	}

	// Маскируем пароль если он есть
	if u.User != nil {
		username := u.User.Username()
		_, hasPassword := u.User.Password()
		if hasPassword {
			u.User = url.UserPassword(username, "***")
		} else {
			u.User = url.User(username)
		}
	}

	return u.String()
}
