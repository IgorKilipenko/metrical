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

// getEnvOrDefault получает значение из переменной окружения или возвращает значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault получает целочисленное значение из переменной окружения или возвращает значение по умолчанию
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getDefaultValues получает значения по умолчанию с учетом переменных окружения
func getDefaultValues() (string, int, int) {
	// Получаем значения из переменных окружения или используем дефолтные
	envServerURL := getEnvOrDefault("ADDRESS", agent.DefaultServerURL)
	envPollInterval := getEnvIntOrDefault("POLL_INTERVAL", int(agent.DefaultPollInterval.Seconds()))
	envReportInterval := getEnvIntOrDefault("REPORT_INTERVAL", int(agent.DefaultReportInterval.Seconds()))

	return envServerURL, envPollInterval, envReportInterval
}

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
	defaultServerURL, defaultPollInterval, defaultReportInterval := getDefaultValues()

	rootCmd.Flags().StringVarP(&serverURL, "a", "a", defaultServerURL, "HTTP server endpoint address")
	rootCmd.Flags().IntVarP(&pollInterval, "p", "p", defaultPollInterval, "Poll interval in seconds")
	rootCmd.Flags().IntVarP(&reportInterval, "r", "r", defaultReportInterval, "Report interval in seconds")
	rootCmd.Flags().BoolVarP(&verboseLogging, "v", "v", false, "Enable verbose logging")

	// Отключаем автоматическое использование флага help, так как Cobra его добавляет автоматически
	rootCmd.Flags().BoolP("help", "h", false, "Show help")
}

// runAgent запускает агент с заданной конфигурацией
func runAgent(cmd *cobra.Command, args []string) error {
	// Проверяем на неизвестные аргументы
	if len(args) > 0 {
		return fmt.Errorf("unknown arguments: %v", args)
	}

	// Определяем финальные значения с учетом приоритета:
	// 1. Переменные окружения
	// 2. Флаги командной строки
	// 3. Значения по умолчанию

	finalServerURL := getEnvOrDefault("ADDRESS", serverURL)
	finalPollInterval := getEnvIntOrDefault("POLL_INTERVAL", pollInterval)
	finalReportInterval := getEnvIntOrDefault("REPORT_INTERVAL", reportInterval)

	// Создаем конфигурацию с учетом приоритета параметров
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
