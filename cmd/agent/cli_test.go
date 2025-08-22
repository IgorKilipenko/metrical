package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/agent"
	"github.com/IgorKilipenko/metrical/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmdFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedConfig *agent.Config
		expectError    bool
	}{
		{
			name: "default values",
			args: []string{},
			expectedConfig: &agent.Config{
				ServerURL:      agent.DefaultServerURL,
				PollInterval:   agent.DefaultPollInterval,
				ReportInterval: agent.DefaultReportInterval,
			},
			expectError: false,
		},
		{
			name: "custom server URL",
			args: []string{"-a", "http://example.com:9090"},
			expectedConfig: &agent.Config{
				ServerURL:      "http://example.com:9090",
				PollInterval:   agent.DefaultPollInterval,
				ReportInterval: agent.DefaultReportInterval,
			},
			expectError: false,
		},
		{
			name: "custom poll interval",
			args: []string{"-p", "5"},
			expectedConfig: &agent.Config{
				ServerURL:      agent.DefaultServerURL,
				PollInterval:   5 * time.Second,
				ReportInterval: agent.DefaultReportInterval,
			},
			expectError: false,
		},
		{
			name: "custom report interval",
			args: []string{"-r", "15"},
			expectedConfig: &agent.Config{
				ServerURL:      agent.DefaultServerURL,
				PollInterval:   agent.DefaultPollInterval,
				ReportInterval: 15 * time.Second,
			},
			expectError: false,
		},
		{
			name: "all custom values",
			args: []string{"-a", "http://test.com:8080", "-p", "3", "-r", "20"},
			expectedConfig: &agent.Config{
				ServerURL:      "http://test.com:8080",
				PollInterval:   3 * time.Second,
				ReportInterval: 20 * time.Second,
			},
			expectError: false,
		},
		{
			name: "with verbose logging",
			args: []string{"-v"},
			expectedConfig: &agent.Config{
				ServerURL:      agent.DefaultServerURL,
				PollInterval:   agent.DefaultPollInterval,
				ReportInterval: agent.DefaultReportInterval,
			},
			expectError: false,
		},
		{
			name:        "unknown argument",
			args:        []string{"unknown"},
			expectError: true,
		},
		{
			name:        "invalid poll interval",
			args:        []string{"-p", "-1"},
			expectError: true,
		},
		{
			name:        "invalid report interval",
			args:        []string{"-r", "0"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новую команду для каждого теста, чтобы избежать конфликтов
			cmd := &cobra.Command{
				Use:   "test",
				Short: "Test command",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Проверяем на неизвестные аргументы
					if len(args) > 0 {
						return fmt.Errorf("unknown arguments: %v", args)
					}

					// Создаем конфигурацию из флагов
					config := &agent.Config{
						ServerURL:      serverURL,
						PollInterval:   time.Duration(pollInterval) * time.Second,
						ReportInterval: time.Duration(reportInterval) * time.Second,
					}

					// Валидируем конфигурацию
					if err := config.Validate(); err != nil {
						return fmt.Errorf("invalid configuration: %w", err)
					}

					return nil
				},
			}

			// Добавляем флаги
			cmd.Flags().StringVarP(&serverURL, "a", "a", agent.DefaultServerURL, "HTTP server endpoint address")
			cmd.Flags().IntVarP(&pollInterval, "p", "p", int(agent.DefaultPollInterval.Seconds()), "Poll interval in seconds")
			cmd.Flags().IntVarP(&reportInterval, "r", "r", int(agent.DefaultReportInterval.Seconds()), "Report interval in seconds")
			cmd.Flags().BoolVarP(&verboseLogging, "v", "v", false, "Enable verbose logging")

			// Устанавливаем аргументы
			cmd.SetArgs(tt.args)

			// Выполняем команду
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Проверяем, что флаги установлены правильно
			if tt.expectedConfig != nil {
				// Создаем конфигурацию из установленных флагов
				config := &agent.Config{
					ServerURL:      serverURL,
					PollInterval:   time.Duration(pollInterval) * time.Second,
					ReportInterval: time.Duration(reportInterval) * time.Second,
					VerboseLogging: verboseLogging,
				}

				assert.Equal(t, tt.expectedConfig.ServerURL, config.ServerURL)
				assert.Equal(t, tt.expectedConfig.PollInterval, config.PollInterval)
				assert.Equal(t, tt.expectedConfig.ReportInterval, config.ReportInterval)
				// Проверяем VerboseLogging только для теста с verbose
				if tt.name == "with verbose logging" {
					assert.True(t, config.VerboseLogging)
				}
			}
		})
	}
}

func TestRootCmdHelp(t *testing.T) {
	// Проверяем, что команда help работает
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestVersion(t *testing.T) {
	// Проверяем, что версия установлена
	assert.NotEmpty(t, Version, "Version should not be empty")
	assert.Contains(t, Version, "dev", "Version should contain 'dev' by default")
}

func TestViperIntegration(t *testing.T) {
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

		// Тестируем загрузку конфигурации с значениями по умолчанию
		config, err := config.LoadAgentConfig()
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

		// Тестируем загрузку конфигурации с переменными окружения
		config, err := config.LoadAgentConfig()
		require.NoError(t, err)

		assert.Equal(t, "http://test-server:9090", config.ServerURL)
		assert.Equal(t, 5*time.Second, config.PollInterval)
		assert.Equal(t, 15*time.Second, config.ReportInterval)
		assert.False(t, config.VerboseLogging)
	})
}
