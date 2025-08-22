package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// AgentConfig конфигурация агента
type AgentConfig struct {
	ServerURL      string        `mapstructure:"server_url"`
	PollInterval   time.Duration `mapstructure:"poll_interval"`
	ReportInterval time.Duration `mapstructure:"report_interval"`
	VerboseLogging bool          `mapstructure:"verbose_logging"`
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Address string `mapstructure:"address"`
}

// LoadAgentConfig загружает конфигурацию агента с поддержкой переменных окружения
func LoadAgentConfig() (*AgentConfig, error) {
	v := viper.New()

	// Устанавливаем значения по умолчанию
	v.SetDefault("server_url", "http://localhost:8080")
	v.SetDefault("poll_interval", 2*time.Second)
	v.SetDefault("report_interval", 10*time.Second)
	v.SetDefault("verbose_logging", false)

	// Настраиваем переменные окружения
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	// Привязываем переменные окружения к ключам конфигурации
	v.BindEnv("server_url", "ADDRESS")
	v.BindEnv("poll_interval", "POLL_INTERVAL")
	v.BindEnv("report_interval", "REPORT_INTERVAL")

	var config AgentConfig

	// Сначала обрабатываем переменные окружения для duration
	if pollIntervalStr := v.GetString("POLL_INTERVAL"); pollIntervalStr != "" {
		if pollInterval, err := time.ParseDuration(pollIntervalStr + "s"); err == nil {
			v.Set("poll_interval", pollInterval)
		}
	}

	if reportIntervalStr := v.GetString("REPORT_INTERVAL"); reportIntervalStr != "" {
		if reportInterval, err := time.ParseDuration(reportIntervalStr + "s"); err == nil {
			v.Set("report_interval", reportInterval)
		}
	}

	// Теперь парсим конфигурацию
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Валидируем конфигурацию
	if err := validateAgentConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadServerConfig загружает конфигурацию сервера с поддержкой переменных окружения
func LoadServerConfig() (*ServerConfig, error) {
	v := viper.New()

	// Устанавливаем значение по умолчанию
	v.SetDefault("address", "localhost:8080")

	// Настраиваем переменные окружения
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	// Привязываем переменную окружения к ключу конфигурации
	v.BindEnv("address", "ADDRESS")

	var config ServerConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Валидируем конфигурацию
	if err := validateServerConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateAgentConfig валидирует конфигурацию агента
func validateAgentConfig(config *AgentConfig) error {
	if config.ServerURL == "" {
		return fmt.Errorf("server URL cannot be empty")
	}

	if config.PollInterval <= 0 {
		return fmt.Errorf("poll interval must be positive")
	}

	if config.ReportInterval <= 0 {
		return fmt.Errorf("report interval must be positive")
	}

	return nil
}

// validateServerConfig валидирует конфигурацию сервера
func validateServerConfig(config *ServerConfig) error {
	if config.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	return nil
}

// GetString получает строковое значение из конфигурации
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt получает целочисленное значение из конфигурации
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetDuration получает значение duration из конфигурации
func GetDuration(key string) time.Duration {
	return viper.GetDuration(key)
}

// GetBool получает булево значение из конфигурации
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// SetDefault устанавливает значение по умолчанию
func SetDefault(key string, value interface{}) {
	viper.SetDefault(key, value)
}

// BindEnv привязывает переменную окружения к ключу конфигурации
func BindEnv(key string, envVar string) error {
	return viper.BindEnv(key, envVar)
}
