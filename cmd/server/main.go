package main

import (
	"os"

	"github.com/IgorKilipenko/metrical/internal/app"
	"github.com/IgorKilipenko/metrical/internal/logger"
)

// Version приложения (можно установить при сборке через ldflags)
var Version = "dev"

// osExit - переменная для подмены os.Exit в тестах
var osExit = os.Exit

// mainLogger - глобальный логгер для main функции
var mainLogger = logger.NewSlogLogger()

// handleError обрабатывает ошибки и завершает программу с соответствующим кодом выхода
func handleError(err error) {
	if err == nil {
		return
	}

	// Если это help, просто выходим без ошибки
	if IsHelpRequested(err) {
		osExit(0)
	}

	// Если это запрос версии, просто выходим без ошибки
	if IsVersionRequested(err) {
		osExit(0)
	}

	// Если это ошибка валидации адреса, выводим сообщение и выходим с кодом 1
	if IsInvalidAddress(err) {
		mainLogger.Error("configuration error", "error", err)
		osExit(1)
		return
	}

	// Для всех остальных ошибок используем log.Fatal
	mainLogger.Error("fatal error", "error", err)
	osExit(1)
}

func main() {
	mainLogger.Info("starting metrics server", "version", Version)

	config, err := parseFlags()
	handleError(err)

	appConfig, err := app.NewConfig(config.Address, config.StoreInterval, config.FileStoragePath, config.Restore)
	handleError(err)

	application := app.New(appConfig)

	if err := application.Run(); err != nil {
		mainLogger.Error("application error", "error", err)
		osExit(1)
	}

	mainLogger.Info("server shutdown completed successfully")
}
