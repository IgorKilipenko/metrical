package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/agent"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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
				}

				assert.Equal(t, tt.expectedConfig.ServerURL, config.ServerURL)
				assert.Equal(t, tt.expectedConfig.PollInterval, config.PollInterval)
				assert.Equal(t, tt.expectedConfig.ReportInterval, config.ReportInterval)
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
