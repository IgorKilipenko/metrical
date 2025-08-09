package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IgorKilipenko/metrical/internal/agent"
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
  -r: Report interval in seconds (default: 10)`,
	RunE: runAgent,
}

// init инициализирует флаги командной строки
func init() {
	rootCmd.Flags().StringVarP(&serverURL, "a", "a", agent.DefaultServerURL, "HTTP server endpoint address")
	rootCmd.Flags().IntVarP(&pollInterval, "p", "p", int(agent.DefaultPollInterval.Seconds()), "Poll interval in seconds")
	rootCmd.Flags().IntVarP(&reportInterval, "r", "r", int(agent.DefaultReportInterval.Seconds()), "Report interval in seconds")
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

	// Создаем конфигурацию из флагов
	config := &agent.Config{
		ServerURL:      serverURL,
		PollInterval:   time.Duration(pollInterval) * time.Second,
		ReportInterval: time.Duration(reportInterval) * time.Second,
		VerboseLogging: verboseLogging,
	}

	// Валидируем конфигурацию
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Логируем конфигурацию при запуске
	log.Printf("Agent configuration: server=%s, poll=%v, report=%v, verbose=%v",
		config.ServerURL, config.PollInterval, config.ReportInterval, config.VerboseLogging)

	// Создаем и запускаем агент
	metricsAgent := agent.NewAgent(config)

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
