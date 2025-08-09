package main

import (
	"log"

	"github.com/IgorKilipenko/metrical/internal/app"
)

func main() {
	// Загружаем конфигурацию
	config := app.LoadConfig()

	// Создаем приложение
	application := app.New(config)

	// Запускаем приложение
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
