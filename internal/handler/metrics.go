package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/IgorKilipenko/metrical/internal/template"
	"github.com/IgorKilipenko/metrical/internal/validation"
	"github.com/go-chi/chi/v5"
)

// MetricsHandler обработчик HTTP запросов для метрик
type MetricsHandler struct {
	service  *service.MetricsService
	template *template.MetricsTemplate
}

// NewMetricsHandler создает новый экземпляр MetricsHandler
func NewMetricsHandler(service *service.MetricsService) *MetricsHandler {
	template, err := template.NewMetricsTemplate()
	if err != nil {
		// В продакшене лучше использовать panic или возвращать ошибку
		// Для простоты используем panic, так как это конструктор
		panic(fmt.Sprintf("failed to create metrics template: %v", err))
	}

	return &MetricsHandler{
		service:  service,
		template: template,
	}
}

// UpdateMetric обновляет метрику
func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	// Валидация через пакет validation
	metricReq, err := validation.ValidateMetricRequest(metricType, metricName, metricValue)
	if err != nil {
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
		log.Printf("Error updating metric: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetMetricValue возвращает значение метрики
func (h *MetricsHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	// Создаем контекст с таймаутом для операции
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	// Валидация имени метрики
	if err := validation.ValidateMetricName(metricName); err != nil {
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
			log.Printf("Error getting gauge metric: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		value = gaugeValue

	case "counter":
		var counterValue int64
		var exists bool
		counterValue, exists, err = h.service.GetCounter(ctx, metricName)
		if err != nil {
			log.Printf("Error getting counter metric: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		value = counterValue

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(fmt.Appendf(nil, "%v", value))
}

// GetAllMetrics возвращает все метрики в HTML формате
func (h *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	// Создаем контекст с таймаутом для операции
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Получаем данные метрик через приватный метод
	metricsData, err := h.getAllMetricsData(ctx)
	if err != nil {
		log.Printf("Error getting metrics data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Выполняем шаблон
	htmlBytes, err := h.template.Execute(*metricsData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(htmlBytes)
}

// getAllMetricsData получает все данные метрик
func (h *MetricsHandler) getAllMetricsData(ctx context.Context) (*template.MetricsData, error) {
	gauges, err := h.service.GetAllGauges(ctx)
	if err != nil {
		return nil, err
	}

	counters, err := h.service.GetAllCounters(ctx)
	if err != nil {
		return nil, err
	}

	return &template.MetricsData{
		Gauges:       gauges,
		Counters:     counters,
		GaugeCount:   len(gauges),
		CounterCount: len(counters),
	}, nil
}
