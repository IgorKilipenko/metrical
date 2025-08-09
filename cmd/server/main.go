package main

import (
	"log/slog"
	"os"

	"github.com/IgorKilipenko/metrical/internal/app"
)

// Version приложения (можно установить при сборке через ldflags)
var Version = "dev"

// osExit - переменная для подмены os.Exit в тестах
var osExit = os.Exit

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
		slog.Error("configuration error", "error", err)
		osExit(1)
		return
	}

	// Для всех остальных ошибок используем log.Fatal
	slog.Error("fatal error", "error", err)
	osExit(1)
}

func main() {
	slog.Info("starting metrics server", "version", Version)

	addr, err := parseFlags()
	handleError(err)

	config, err := app.NewConfig(addr)
	handleError(err)

	application := app.New(config)

	if err := application.Run(); err != nil {
		slog.Error("application error", "error", err)
		osExit(1)
	}

	slog.Info("server shutdown completed successfully")
}
