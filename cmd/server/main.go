package main

import (
	"log"
	"net/http"

	"github.com/IgorKilipenko/metrical/internal/handler"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/service"
)

func main() {
	// Создаем хранилище метрик
	storage := models.NewMemStorage()

	// Создаем сервис для работы с метриками
	metricsService := service.NewMetricsService(storage)

	// Создаем HTTP обработчик
	metricsHandler := handler.NewMetricsHandler(metricsService)

	// Настраиваем маршруты
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", metricsHandler.UpdateMetric)

	// Запускаем сервер
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
