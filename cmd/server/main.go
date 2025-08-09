package main

import (
	"log"
	"os"

	"github.com/IgorKilipenko/metrical/internal/app"
)

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

	// Если это ошибка валидации адреса, выводим сообщение и выходим с кодом 1
	if IsInvalidAddress(err) {
		log.Printf("Ошибка конфигурации: %v", err)
		osExit(1)
		return
	}

	// Для всех остальных ошибок используем log.Fatal
	log.Fatal(err)
}

func main() {
	addr, err := parseFlags()
	handleError(err)

	config, err := app.NewConfig(addr)
	handleError(err)

	application := app.New(config)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
