package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/IgorKilipenko/metrical/internal/agent"
	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/spf13/cobra"
)

var (
	// Флаги командной строки
	serverURL      string
	pollInterval   int
	reportInterval int
	verboseLogging bool
)

// rootCmd представляет корневую команду приложения
var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "Metrics collection agent",
	Long: `Metrics collection agent that polls runtime metrics and sends them to a server.

Supported flags:
  -a: HTTP server endpoint address (default: localhost:8080)
  -p: Poll interval in seconds (default: 2)
  -r: Report interval in seconds (default: 10)

Environment variables:
  ADDRESS: HTTP server endpoint address
  POLL_INTERVAL: Poll interval in seconds
  REPORT_INTERVAL: Report interval in seconds`,
	RunE: runAgent,
}

// init инициализирует флаги командной строки
func init() {
	// Получаем значения по умолчанию с учетом переменных окружения
	defaultServerURL := getEnvOrDefault("ADDRESS", agent.DefaultServerURL)
	defaultPollInterval := getEnvIntOrDefault("POLL_INTERVAL", int(agent.DefaultPollInterval.Seconds()))
	defaultReportInterval := getEnvIntOrDefault("REPORT_INTERVAL", int(agent.DefaultReportInterval.Seconds()))

	rootCmd.Flags().StringVarP(&serverURL, "a", "a", defaultServerURL, "HTTP server endpoint address")
	rootCmd.Flags().IntVarP(&pollInterval, "p", "p", defaultPollInterval, "Poll interval in seconds")
	rootCmd.Flags().IntVarP(&reportInterval, "r", "r", defaultReportInterval, "Report interval in seconds")
	rootCmd.Flags().BoolVarP(&verboseLogging, "v", "v", false, "Enable verbose logging")

	// Отключаем автоматическое использование флага help, так как Cobra его добавляет автоматически
	rootCmd.Flags().BoolP("help", "h", false, "Show help")
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault возвращает целочисленное значение переменной окружения или значение по умолчанию
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getFinalValue возвращает финальное значение с учетом приоритета
func getFinalValue(envKey, flagValue, defaultValue string) string {
	// 1. Переменная окружения (высший приоритет)
	if envValue, ok := os.LookupEnv(envKey); ok {
		return envValue
	}
	// 2. Флаг командной строки (средний приоритет)
	if flagValue != "" && flagValue != defaultValue {
		return flagValue
	}
	// 3. Значение по умолчанию (низший приоритет)
	return defaultValue
}

// getFinalIntValue возвращает финальное целочисленное значение с учетом приоритета
func getFinalIntValue(envKey string, flagValue, defaultValue int) int {
	// 1. Переменная окружения (высший приоритет)
	if envValue, ok := os.LookupEnv(envKey); ok {
		if intValue, err := strconv.Atoi(envValue); err == nil {
			return intValue
		}
	}
	// 2. Флаг командной строки (средний приоритет)
	if flagValue != defaultValue {
		return flagValue
	}
	// 3. Значение по умолчанию (низший приоритет)
	return defaultValue
}

// runAgent запускает агент с заданной конфигурацией
func runAgent(cmd *cobra.Command, args []string) error {
	// Проверяем на неизвестные аргументы
	if len(args) > 0 {
		return fmt.Errorf("unknown arguments: %v", args)
	}

	// Получаем финальные значения с учетом приоритета
	finalServerURL := getFinalValue("ADDRESS", serverURL, agent.DefaultServerURL)
	finalPollInterval := getFinalIntValue("POLL_INTERVAL", pollInterval, int(agent.DefaultPollInterval.Seconds()))
	finalReportInterval := getFinalIntValue("REPORT_INTERVAL", reportInterval, int(agent.DefaultReportInterval.Seconds()))

	// Создаем конфигурацию из финальных значений
	config := &agent.Config{
		ServerURL:      finalServerURL,
		PollInterval:   time.Duration(finalPollInterval) * time.Second,
		ReportInterval: time.Duration(finalReportInterval) * time.Second,
		VerboseLogging: verboseLogging,
	}

	// Валидируем конфигурацию
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Логируем конфигурацию при запуске
	log.Printf("Agent configuration: server=%s, poll=%v, report=%v, verbose=%v",
		config.ServerURL, config.PollInterval, config.ReportInterval, config.VerboseLogging)

	// Создаем логгер
	agentLogger := logger.NewSlogLogger()

	// Создаем и запускаем агент
	metricsAgent := agent.NewAgent(config, agentLogger)

	// Создаем контекст с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем агент в горутине
	go func() {
		log.Printf("Starting metrics agent with config: server=%s, poll=%v, report=%v",
			config.ServerURL, config.PollInterval, config.ReportInterval)
		metricsAgent.Run()
	}()

	// Ждем сигнала завершения
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case <-ctx.Done():
		log.Println("Context cancelled")
	}

	// Graceful shutdown
	metricsAgent.Stop()

	// Даем время на завершение горутин
	time.Sleep(1 * time.Second)
	log.Println("Agent shutdown completed")

	return nil
}

// Execute запускает CLI приложение
func Execute() error {
	return rootCmd.Execute()
}
