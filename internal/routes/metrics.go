package routes

import (
	"net/http"

	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/IgorKilipenko/metrical/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// SetupMetricsRoutes настраивает маршруты для метрик
func SetupMetricsRoutes(handler *handler.MetricsHandler) *chi.Mux {
	r := chi.NewRouter()

	// Добавляем middleware для логирования
	r.Use(middleware.LoggingMiddleware())

	// Простые тестовые маршруты
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Router is working"))
	})

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// Основные маршруты метрик
	r.Get("/", handler.GetAllMetrics)
	r.Post("/update/{type}/{name}/{value}", handler.UpdateMetric)
	r.Get("/value/{type}/{name}", handler.GetMetricValue)

	// JSON API маршруты
	r.Post("/update", handler.UpdateMetricJSON)
	r.Post("/value", handler.GetMetricJSON)

	return r
}

// SetupHealthRoutes настраивает маршруты для health check (пример расширения)
func SetupHealthRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Добавляем middleware для логирования
	r.Use(middleware.LoggingMiddleware())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
