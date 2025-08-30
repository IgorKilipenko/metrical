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

	// Основные маршруты метрик
	r.Get("/", handler.GetAllMetrics)
	r.Post("/update/{type}/{name}/{value}", handler.UpdateMetric)
	r.Get("/value/{type}/{name}", handler.GetMetricValue)

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
