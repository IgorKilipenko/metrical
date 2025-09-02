package routes

import (
	"net/http"
	"strings"

	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/IgorKilipenko/metrical/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// SetupMetricsRoutes настраивает маршруты для метрик
func SetupMetricsRoutes(handler *handler.MetricsHandler) *chi.Mux {
	r := chi.NewRouter()

	// Добавляем middleware для логирования
	r.Use(middleware.LoggingMiddleware())

	// Добавляем middleware для поддержки gzip
	r.Use(middleware.GzipMiddleware())

	// Настраиваем автоматическую обработку trailing slash
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Убираем trailing slash для всех запросов (кроме корневого пути)
			if r.URL.Path != "/" {
				r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			}
			next.ServeHTTP(w, r)
		})
	})

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
