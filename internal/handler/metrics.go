package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	models "github.com/IgorKilipenko/metrical/internal/model"
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
func NewMetricsHandler(service *service.MetricsService, logger logger.Logger) (*MetricsHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	template, err := template.NewMetricsTemplate()
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics template: %w", err)
	}

	return &MetricsHandler{
		service:  service,
		template: template,
		logger:   logger,
	}, nil
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.logger.Info("metric updated successfully",
		"type", metricType,
		"name", metricType,
		"value", metricValue)
	w.WriteHeader(http.StatusOK)
}

// UpdateMetricJSON обновляет метрику из JSON запроса
func (h *MetricsHandler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("processing update metric JSON request",
		"method", r.Method,
		"url", r.URL.Path,
		"remote_addr", r.RemoteAddr)

	// Проверяем Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		h.logger.Warn("invalid content type", "content_type", r.Header.Get("Content-Type"))
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Декодируем JSON
	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.logger.Warn("failed to decode JSON", "error", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Валидация метрики
	if err := h.validateMetricJSON(&metric); err != nil {
		h.logger.Warn("metric validation failed", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Обновляем метрику через сервис
	err := h.service.UpdateMetricJSON(ctx, &metric)
	if err != nil {
		h.logger.Error("failed to update metric", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.logger.Info("metric updated successfully from JSON",
		"id", metric.ID,
		"type", metric.MType)
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
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

// GetMetricJSON возвращает метрику в JSON формате
func (h *MetricsHandler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("processing get metric JSON request",
		"method", r.Method,
		"url", r.URL.Path,
		"remote_addr", r.RemoteAddr)

	// Проверяем Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		h.logger.Warn("invalid content type", "content_type", r.Header.Get("Content-Type"))
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Декодируем JSON
	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.logger.Warn("failed to decode JSON", "error", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Валидация запроса
	if err := h.validateMetricRequestJSON(&metric); err != nil {
		h.logger.Warn("metric request validation failed", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Получаем метрику через сервис
	result, err := h.service.GetMetricJSON(ctx, &metric)
	if err != nil {
		h.logger.Error("failed to get metric", "error", err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем заголовки
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Кодируем ответ в JSON
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.logger.Info("metric retrieved successfully as JSON",
		"id", result.ID,
		"type", result.MType)
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Выполняем шаблон
	htmlBytes, err := h.template.Execute(*metricsData)
	if err != nil {
		h.logger.Error("failed to execute template",
			"error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

// validateMetricJSON валидирует метрику из JSON
func (h *MetricsHandler) validateMetricJSON(metric *models.Metrics) error {
	if metric.ID == "" {
		return fmt.Errorf("metric ID is required")
	}

	if metric.MType == "" {
		return fmt.Errorf("metric type is required")
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("value is required for gauge metric")
		}
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("delta is required for counter metric")
		}
	default:
		return fmt.Errorf("unsupported metric type: %s", metric.MType)
	}

	return nil
}

// validateMetricRequestJSON валидирует запрос на получение метрики
func (h *MetricsHandler) validateMetricRequestJSON(metric *models.Metrics) error {
	if metric.ID == "" {
		return fmt.Errorf("metric ID is required")
	}

	if metric.MType == "" {
		return fmt.Errorf("metric type is required")
	}

	switch metric.MType {
	case models.Gauge, models.Counter:
		// Тип поддерживается
	default:
		return fmt.Errorf("unsupported metric type: %s", metric.MType)
	}

	return nil
}
