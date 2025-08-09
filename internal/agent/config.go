package agent

import (
	"fmt"
	"time"
)

// Константы конфигурации по умолчанию
const (
	DefaultServerURL      = "http://localhost:8080"
	DefaultPollInterval   = 2 * time.Second
	DefaultReportInterval = 10 * time.Second
	DefaultHTTPTimeout    = 10 * time.Second
)

// Config конфигурация агента.
// Содержит настройки для подключения к серверу и интервалы работы.
type Config struct {
	// ServerURL - URL сервера для отправки метрик
	ServerURL string

	// PollInterval - интервал сбора метрик из runtime
	PollInterval time.Duration

	// ReportInterval - интервал отправки метрик на сервер
	ReportInterval time.Duration

	// VerboseLogging - подробное логирование (включая ошибки отправки метрик)
	VerboseLogging bool
}

// NewConfig создает конфигурацию с значениями по умолчанию.
// Используется для быстрого создания стандартной конфигурации агента.
//
// Возвращает:
//   - *Config: указатель на новую конфигурацию с дефолтными значениями
func NewConfig() *Config {
	return &Config{
		ServerURL:      DefaultServerURL,
		PollInterval:   DefaultPollInterval,
		ReportInterval: DefaultReportInterval,
		VerboseLogging: false,
	}
}

// NewConfigWithURL создает конфигурацию с пользовательским URL сервера.
// Остальные параметры устанавливаются по умолчанию.
//
// Параметры:
//   - serverURL: URL сервера для отправки метрик
//
// Возвращает:
//   - *Config: указатель на новую конфигурацию с пользовательским URL
func NewConfigWithURL(serverURL string) *Config {
	return &Config{
		ServerURL:      serverURL,
		PollInterval:   DefaultPollInterval,
		ReportInterval: DefaultReportInterval,
	}
}

// Validate проверяет корректность конфигурации.
// Возвращает ошибку, если конфигурация некорректна.
//
// Возвращает:
//   - error: описание ошибки валидации или nil если конфигурация корректна
func (c *Config) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("server URL cannot be empty")
	}

	if c.PollInterval <= 0 {
		return fmt.Errorf("poll interval must be positive")
	}

	if c.ReportInterval <= 0 {
		return fmt.Errorf("report interval must be positive")
	}

	return nil
}

// IsValid проверяет, является ли конфигурация корректной.
//
// Возвращает:
//   - bool: true если конфигурация корректна, false в противном случае
func (c *Config) IsValid() bool {
	return c.Validate() == nil
}
