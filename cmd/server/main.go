package main

import (
	"log"
	"os"

	"github.com/IgorKilipenko/metrical/internal/app"
)

func main() {
	// Загружаем конфигурацию
	config, err := app.LoadConfig()
	if err != nil {
		// Если это help, просто выходим без ошибки
		if err.Error() == "help requested" {
			os.Exit(0)
		}
		log.Fatal(err)
	}

	// Создаем приложение
	application := app.New(config)

	// Запускаем приложение
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
