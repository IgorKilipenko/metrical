package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAgentConfig(t *testing.T) {
	// Сохраняем оригинальные значения переменных окружения
	originalAddress := os.Getenv("ADDRESS")
	originalPollInterval := os.Getenv("POLL_INTERVAL")
	originalReportInterval := os.Getenv("REPORT_INTERVAL")

	// Восстанавливаем оригинальные значения после теста
	defer func() {
		if originalAddress != "" {
			os.Setenv("ADDRESS", originalAddress)
		} else {
			os.Unsetenv("ADDRESS")
		}
		if originalPollInterval != "" {
			os.Setenv("POLL_INTERVAL", originalPollInterval)
		} else {
			os.Unsetenv("POLL_INTERVAL")
		}
		if originalReportInterval != "" {
			os.Setenv("REPORT_INTERVAL", originalReportInterval)
		} else {
			os.Unsetenv("REPORT_INTERVAL")
		}
	}()

	t.Run("default values", func(t *testing.T) {
		// Очищаем переменные окружения
		os.Unsetenv("ADDRESS")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("REPORT_INTERVAL")

		config, err := LoadAgentConfig()
		require.NoError(t, err)

		assert.Equal(t, "http://localhost:8080", config.ServerURL)
		assert.Equal(t, 2*time.Second, config.PollInterval)
		assert.Equal(t, 10*time.Second, config.ReportInterval)
		assert.False(t, config.VerboseLogging)
	})

	t.Run("with environment variables", func(t *testing.T) {
		// Устанавливаем переменные окружения
		os.Setenv("ADDRESS", "http://test-server:9090")
		os.Setenv("POLL_INTERVAL", "5")
		os.Setenv("REPORT_INTERVAL", "15")

		config, err := LoadAgentConfig()
		require.NoError(t, err)

		assert.Equal(t, "http://test-server:9090", config.ServerURL)
		assert.Equal(t, 5*time.Second, config.PollInterval)
		assert.Equal(t, 15*time.Second, config.ReportInterval)
		assert.False(t, config.VerboseLogging)
	})

	t.Run("invalid poll interval", func(t *testing.T) {
		os.Setenv("POLL_INTERVAL", "invalid")

		_, err := LoadAgentConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid duration")
	})

	t.Run("invalid report interval", func(t *testing.T) {
		os.Setenv("REPORT_INTERVAL", "invalid")

		_, err := LoadAgentConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid duration")
	})

	t.Run("zero poll interval", func(t *testing.T) {
		// Очищаем все переменные окружения
		os.Unsetenv("ADDRESS")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("REPORT_INTERVAL")

		os.Setenv("POLL_INTERVAL", "0s")

		_, err := LoadAgentConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "poll interval must be positive")
	})

	t.Run("zero report interval", func(t *testing.T) {
		// Очищаем все переменные окружения
		os.Unsetenv("ADDRESS")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("REPORT_INTERVAL")

		os.Setenv("REPORT_INTERVAL", "0s")

		_, err := LoadAgentConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "report interval must be positive")
	})
}

func TestLoadServerConfig(t *testing.T) {
	// Сохраняем оригинальное значение переменной окружения
	originalAddress := os.Getenv("ADDRESS")

	// Восстанавливаем оригинальное значение после теста
	defer func() {
		if originalAddress != "" {
			os.Setenv("ADDRESS", originalAddress)
		} else {
			os.Unsetenv("ADDRESS")
		}
	}()

	t.Run("default value", func(t *testing.T) {
		// Очищаем переменную окружения
		os.Unsetenv("ADDRESS")

		config, err := LoadServerConfig()
		require.NoError(t, err)

		assert.Equal(t, "localhost:8080", config.Address)
	})

	t.Run("with environment variable", func(t *testing.T) {
		// Устанавливаем переменную окружения
		os.Setenv("ADDRESS", "test-server:9090")

		config, err := LoadServerConfig()
		require.NoError(t, err)

		assert.Equal(t, "test-server:9090", config.Address)
	})
}

func TestValidateAgentConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &AgentConfig{
			ServerURL:      "http://localhost:8080",
			PollInterval:   2 * time.Second,
			ReportInterval: 10 * time.Second,
			VerboseLogging: false,
		}

		err := validateAgentConfig(config)
		assert.NoError(t, err)
	})

	t.Run("empty server URL", func(t *testing.T) {
		config := &AgentConfig{
			ServerURL:      "",
			PollInterval:   2 * time.Second,
			ReportInterval: 10 * time.Second,
		}

		err := validateAgentConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server URL cannot be empty")
	})

	t.Run("zero poll interval", func(t *testing.T) {
		config := &AgentConfig{
			ServerURL:      "http://localhost:8080",
			PollInterval:   0,
			ReportInterval: 10 * time.Second,
		}

		err := validateAgentConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "poll interval must be positive")
	})

	t.Run("negative poll interval", func(t *testing.T) {
		config := &AgentConfig{
			ServerURL:      "http://localhost:8080",
			PollInterval:   -1 * time.Second,
			ReportInterval: 10 * time.Second,
		}

		err := validateAgentConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "poll interval must be positive")
	})

	t.Run("zero report interval", func(t *testing.T) {
		config := &AgentConfig{
			ServerURL:      "http://localhost:8080",
			PollInterval:   2 * time.Second,
			ReportInterval: 0,
		}

		err := validateAgentConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "report interval must be positive")
	})

	t.Run("negative report interval", func(t *testing.T) {
		config := &AgentConfig{
			ServerURL:      "http://localhost:8080",
			PollInterval:   2 * time.Second,
			ReportInterval: -1 * time.Second,
		}

		err := validateAgentConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "report interval must be positive")
	})
}

func TestValidateServerConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &ServerConfig{
			Address: "localhost:8080",
		}

		err := validateServerConfig(config)
		assert.NoError(t, err)
	})

	t.Run("empty address", func(t *testing.T) {
		config := &ServerConfig{
			Address: "",
		}

		err := validateServerConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address cannot be empty")
	})
}

func TestConfigHelpers(t *testing.T) {
	// Сохраняем оригинальные значения переменных окружения
	originalAddress := os.Getenv("ADDRESS")
	originalPollInterval := os.Getenv("POLL_INTERVAL")

	// Восстанавливаем оригинальные значения после теста
	defer func() {
		if originalAddress != "" {
			os.Setenv("ADDRESS", originalAddress)
		} else {
			os.Unsetenv("ADDRESS")
		}
		if originalPollInterval != "" {
			os.Setenv("POLL_INTERVAL", originalPollInterval)
		} else {
			os.Unsetenv("POLL_INTERVAL")
		}
	}()

	t.Run("GetString", func(t *testing.T) {
		os.Setenv("ADDRESS", "test-server:9090")
		SetDefault("test_key", "default_value")

		value := GetString("test_key")
		assert.Equal(t, "default_value", value)
	})

	t.Run("GetInt", func(t *testing.T) {
		os.Setenv("POLL_INTERVAL", "5")
		SetDefault("test_int", 10)

		value := GetInt("test_int")
		assert.Equal(t, 10, value)
	})

	t.Run("GetDuration", func(t *testing.T) {
		SetDefault("test_duration", 5*time.Second)

		value := GetDuration("test_duration")
		assert.Equal(t, 5*time.Second, value)
	})

	t.Run("GetBool", func(t *testing.T) {
		SetDefault("test_bool", true)

		value := GetBool("test_bool")
		assert.True(t, value)
	})
}
