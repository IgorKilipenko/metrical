package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/IgorKilipenko/metrical/internal/template"
	"github.com/IgorKilipenko/metrical/internal/validation"
	"github.com/go-chi/chi/v5"
)

// MetricsHandler обработчик HTTP запросов для метрик
type MetricsHandler struct {
	service  *service.MetricsService
	template *template.MetricsTemplate
	logger   logger.Logger
}

// NewMetricsHandler создает новый экземпляр MetricsHandler
func NewMetricsHandler(service *service.MetricsService, logger logger.Logger) *MetricsHandler {
	if service == nil {
		panic("service cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	template, err := template.NewMetricsTemplate()
	if err != nil {
		// В продакшене лучше использовать panic или возвращать ошибку
		// Для простоты используем panic, так как это конструктор
		panic(fmt.Sprintf("failed to create metrics template: %v", err))
	}

	return &MetricsHandler{
		service:  service,
		template: template,
		logger:   logger,
	}
}

// UpdateMetric обновляет метрику
func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	h.logger.Info("processing update metric request",
		"method", r.Method,
		"url", r.URL.Path,
		"type", metricType,
		"name", metricName,
		"value", metricValue,
		"remote_addr", r.RemoteAddr)

	// Валидация через пакет validation
	metricReq, err := validation.ValidateMetricRequest(metricType, metricName, metricValue)
	if err != nil {
		h.logger.Warn("metric validation failed",
			"type", metricType,
			"name", metricName,
			"value", metricValue,
			"error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем контекст с таймаутом для операции после валидации
	// Используем контекст запроса, который может быть отменен в тестах
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Вызов сервиса с валидированными данными и контекстом
	err = h.service.UpdateMetric(ctx, metricReq)
	if err != nil {
		h.logger.Error("failed to update metric",
			"type", metricType,
			"name", metricName,
			"value", metricValue,
			"error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("metric updated successfully",
		"type", metricType,
		"name", metricName,
		"value", metricValue)
	w.WriteHeader(http.StatusOK)
}

// GetMetricValue возвращает значение метрики
func (h *MetricsHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	// Создаем контекст с таймаутом для операции
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	h.logger.Info("processing get metric request",
		"method", r.Method,
		"url", r.URL.Path,
		"type", metricType,
		"name", metricName,
		"remote_addr", r.RemoteAddr)

	// Валидация имени метрики
	if err := validation.ValidateMetricName(metricName); err != nil {
		h.logger.Warn("metric name validation failed",
			"type", metricType,
			"name", metricName,
			"error", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Получение значения из сервиса с контекстом
	var value any
	var err error

	switch metricType {
	case "gauge":
		var gaugeValue float64
		var exists bool
		gaugeValue, exists, err = h.service.GetGauge(ctx, metricName)
		if err != nil {
			h.logger.Error("failed to get gauge metric",
				"name", metricName,
				"error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			h.logger.Debug("gauge metric not found",
				"name", metricName)
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		value = gaugeValue

	case "counter":
		var counterValue int64
		var exists bool
		counterValue, exists, err = h.service.GetCounter(ctx, metricName)
		if err != nil {
			h.logger.Error("failed to get counter metric",
				"name", metricName,
				"error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			h.logger.Debug("counter metric not found",
				"name", metricName)
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		value = counterValue

	default:
		h.logger.Warn("invalid metric type requested",
			"type", metricType,
			"name", metricName)
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	h.logger.Info("metric retrieved successfully",
		"type", metricType,
		"name", metricName,
		"value", value)
	w.Header().Set("Content-Type", "text/plain")
	w.Write(fmt.Appendf(nil, "%v", value))
}

// GetAllMetrics возвращает все метрики в HTML формате
func (h *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	// Создаем контекст с таймаутом для операции
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.logger.Info("processing get all metrics request",
		"method", r.Method,
		"url", r.URL.Path,
		"remote_addr", r.RemoteAddr)

	// Получаем данные метрик через приватный метод
	metricsData, err := h.getAllMetricsData(ctx)
	if err != nil {
		h.logger.Error("failed to get metrics data",
			"error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Выполняем шаблон
	htmlBytes, err := h.template.Execute(*metricsData)
	if err != nil {
		h.logger.Error("failed to execute template",
			"error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("all metrics retrieved successfully",
		"gauge_count", metricsData.GaugeCount,
		"counter_count", metricsData.CounterCount)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(htmlBytes)
}

// getAllMetricsData получает все данные метрик
func (h *MetricsHandler) getAllMetricsData(ctx context.Context) (*template.MetricsData, error) {
	h.logger.Debug("fetching all metrics data")

	gauges, err := h.service.GetAllGauges(ctx)
	if err != nil {
		h.logger.Error("failed to get all gauges", "error", err)
		return nil, err
	}

	counters, err := h.service.GetAllCounters(ctx)
	if err != nil {
		h.logger.Error("failed to get all counters", "error", err)
		return nil, err
	}

	h.logger.Debug("metrics data fetched successfully",
		"gauge_count", len(gauges),
		"counter_count", len(counters))

	return &template.MetricsData{
		Gauges:       gauges,
		Counters:     counters,
		GaugeCount:   len(gauges),
		CounterCount: len(counters),
	}, nil
}
