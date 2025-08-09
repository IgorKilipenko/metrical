package routes

import (
	"net/http"

	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/go-chi/chi/v5"
)

// SetupMetricsRoutes настраивает маршруты для метрик
func SetupMetricsRoutes(handler *handler.MetricsHandler) *chi.Mux {
	r := chi.NewRouter()

	// Основные маршруты метрик
	r.Get("/", handler.GetAllMetrics)
	r.Post("/update/{type}/{name}/{value}", handler.UpdateMetric)
	r.Get("/value/{type}/{name}", handler.GetMetricValue)

	return r
}

// SetupHealthRoutes настраивает маршруты для health check (пример расширения)
func SetupHealthRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
