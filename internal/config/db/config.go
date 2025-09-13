package db

import (
	"fmt"
	"os"
	"strconv"
	"time"
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
func DefaultConfig() Config {
	return Config{
		DSN:                "postgres://metricaldb:Secret@localhost:5432/metricaldb?sslmode=disable",
		MaxConns:           10,
		MinConns:           2,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute * 30,
		ConnectTimeout:     time.Second * 10,
		PingTimeout:        time.Second * 5,
		HealthCheckTimeout: time.Second * 5,
	}
}

// NewConfig создает конфигурацию из переменных окружения и значений по умолчанию
func NewConfig() Config {
	config := DefaultConfig()

	// Получаем DSN из переменной окружения
	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		config.DSN = dsn
	}

	// Получаем настройки пула соединений из переменных окружения
	if maxConns := os.Getenv("DB_MAX_CONNS"); maxConns != "" {
		if val, err := strconv.ParseInt(maxConns, 10, 32); err == nil {
			config.MaxConns = int32(val)
		}
	}

	if minConns := os.Getenv("DB_MIN_CONNS"); minConns != "" {
		if val, err := strconv.ParseInt(minConns, 10, 32); err == nil {
			config.MinConns = int32(val)
		}
	}

	if maxConnLifetime := os.Getenv("DB_MAX_CONN_LIFETIME"); maxConnLifetime != "" {
		if val, err := time.ParseDuration(maxConnLifetime); err == nil {
			config.MaxConnLifetime = val
		}
	}

	if maxConnIdleTime := os.Getenv("DB_MAX_CONN_IDLE_TIME"); maxConnIdleTime != "" {
		if val, err := time.ParseDuration(maxConnIdleTime); err == nil {
			config.MaxConnIdleTime = val
		}
	}

	if connectTimeout := os.Getenv("DB_CONNECT_TIMEOUT"); connectTimeout != "" {
		if val, err := time.ParseDuration(connectTimeout); err == nil {
			config.ConnectTimeout = val
		}
	}

	if pingTimeout := os.Getenv("DB_PING_TIMEOUT"); pingTimeout != "" {
		if val, err := time.ParseDuration(pingTimeout); err == nil {
			config.PingTimeout = val
		}
	}

	if healthCheckTimeout := os.Getenv("DB_HEALTH_CHECK_TIMEOUT"); healthCheckTimeout != "" {
		if val, err := time.ParseDuration(healthCheckTimeout); err == nil {
			config.HealthCheckTimeout = val
		}
	}

	return config
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("database DSN is required")
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
	// Маскируем пароль в DSN для безопасного логирования
	maskedDSN := c.DSN
	if len(c.DSN) > 0 {
		// Простая маскировка - заменяем пароль на ***
		// В реальном проекте лучше использовать более надежный способ
		if idx := len("postgres://"); idx < len(c.DSN) {
			// Ищем двоеточие после имени пользователя
			start := idx
			for i := start; i < len(c.DSN); i++ {
				if c.DSN[i] == ':' {
					// Ищем @ после пароля
					for j := i + 1; j < len(c.DSN); j++ {
						if c.DSN[j] == '@' {
							maskedDSN = c.DSN[:i+1] + "***" + c.DSN[j:]
							break
						}
					}
					break
				}
			}
		}
	}

	return fmt.Sprintf("Config{DSN: %s, MaxConns: %d, MinConns: %d, MaxConnLifetime: %v, MaxConnIdleTime: %v, ConnectTimeout: %v, PingTimeout: %v, HealthCheckTimeout: %v}",
		maskedDSN, c.MaxConns, c.MinConns, c.MaxConnLifetime, c.MaxConnIdleTime, c.ConnectTimeout, c.PingTimeout, c.HealthCheckTimeout)
}
