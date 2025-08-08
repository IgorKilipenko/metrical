package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/IgorKilipenko/metrical/internal/template"
	"github.com/go-chi/chi/v5"
)

// MetricsHandler обрабатывает HTTP запросы для работы с метриками
type MetricsHandler struct {
	service *service.MetricsService
}

// NewMetricsHandler создает новый обработчик метрик
func NewMetricsHandler(service *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		service: service,
	}
}

// UpdateMetric обрабатывает POST запросы для обновления метрик
func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры из chi роутера
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	err := h.service.UpdateMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetMetricValue обрабатывает GET запросы для получения значения метрики
func (h *MetricsHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры из chi роутера
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	var value string
	var found bool

	switch metricType {
	case "gauge":
		var gaugeValue float64
		gaugeValue, found = h.service.GetGauge(metricName)
		if found {
			value = strconv.FormatFloat(gaugeValue, 'f', -1, 64)
		}
	case "counter":
		var counterValue int64
		counterValue, found = h.service.GetCounter(metricName)
		if found {
			value = strconv.FormatInt(counterValue, 10)
		}
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !found {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

// GetAllMetrics обрабатывает GET запросы для отображения всех метрик
func (h *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gauges := h.service.GetAllGauges()
	counters := h.service.GetAllCounters()

	// Создаем шаблон
	mt, err := template.NewMetricsTemplate()
	if err != nil {
		log.Printf("Error creating template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для шаблона
	data := template.MetricsData{
		Gauges:       gauges,
		Counters:     counters,
		GaugeCount:   len(gauges),
		CounterCount: len(counters),
	}

	// Выполняем шаблон
	htmlBytes, err := mt.Execute(data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(htmlBytes)
}
