package main

import (
	"log"
	"os"

	"github.com/IgorKilipenko/metrical/internal/app"
)

func main() {
	addr, err := parseFlags()
	if err != nil {
		// Если это help, просто выходим без ошибки
		if IsHelpRequested(err) {
			os.Exit(0)
		}
		// Если это ошибка валидации адреса, выводим сообщение и выходим с кодом 1
		if IsInvalidAddress(err) {
			log.Printf("Ошибка конфигурации: %v", err)
			os.Exit(1)
		}
		log.Fatal(err)
	}

	config, err := app.NewConfig(addr)
	if err != nil {
		log.Fatal(err)
	}

	application := app.New(config)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
