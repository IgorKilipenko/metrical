package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IgorKilipenko/metrical/internal/agent"
	"github.com/IgorKilipenko/metrical/internal/config"
	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	// Настраиваем Viper для работы с переменными окружения
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Привязываем переменные окружения к ключам конфигурации
	viper.BindEnv("server_url", "ADDRESS")
	viper.BindEnv("poll_interval", "POLL_INTERVAL")
	viper.BindEnv("report_interval", "REPORT_INTERVAL")

	// Устанавливаем значения по умолчанию
	viper.SetDefault("server_url", agent.DefaultServerURL)
	viper.SetDefault("poll_interval", int(agent.DefaultPollInterval.Seconds()))
	viper.SetDefault("report_interval", int(agent.DefaultReportInterval.Seconds()))

	// Получаем значения по умолчанию с учетом переменных окружения
	defaultServerURL := viper.GetString("server_url")
	defaultPollInterval := viper.GetInt("poll_interval")
	defaultReportInterval := viper.GetInt("report_interval")

	rootCmd.Flags().StringVarP(&serverURL, "a", "a", defaultServerURL, "HTTP server endpoint address")
	rootCmd.Flags().IntVarP(&pollInterval, "p", "p", defaultPollInterval, "Poll interval in seconds")
	rootCmd.Flags().IntVarP(&reportInterval, "r", "r", defaultReportInterval, "Report interval in seconds")
	rootCmd.Flags().BoolVarP(&verboseLogging, "v", "v", false, "Enable verbose logging")

	// Отключаем автоматическое использование флага help, так как Cobra его добавляет автоматически
	rootCmd.Flags().BoolP("help", "h", false, "Show help")

	// Привязываем флаги к Viper
	viper.BindPFlag("server_url", rootCmd.Flags().Lookup("a"))
	viper.BindPFlag("poll_interval", rootCmd.Flags().Lookup("p"))
	viper.BindPFlag("report_interval", rootCmd.Flags().Lookup("r"))
	viper.BindPFlag("verbose_logging", rootCmd.Flags().Lookup("v"))
}

// runAgent запускает агент с заданной конфигурацией
func runAgent(cmd *cobra.Command, args []string) error {
	// Проверяем на неизвестные аргументы
	if len(args) > 0 {
		return fmt.Errorf("unknown arguments: %v", args)
	}

	// Загружаем конфигурацию с помощью Viper
	agentConfig, err := config.LoadAgentConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Создаем конфигурацию агента
	config := &agent.Config{
		ServerURL:      agentConfig.ServerURL,
		PollInterval:   agentConfig.PollInterval,
		ReportInterval: agentConfig.ReportInterval,
		VerboseLogging: agentConfig.VerboseLogging,
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
