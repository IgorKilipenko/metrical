package handler

import (
	"log"
	"net/http"
	"strconv"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/IgorKilipenko/metrical/internal/template"
	"github.com/IgorKilipenko/metrical/internal/validation"
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

	// Валидация и парсинг в handler
	metricReq, err := validation.ValidateMetricRequest(metricType, metricName, metricValue)
	if err != nil {
		if models.IsValidationError(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			log.Printf("Error validating metric request: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Вызов сервиса с готовыми валидированными данными
	err = h.service.UpdateMetric(metricReq)
	if err != nil {
		log.Printf("Error updating metric: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

	// Валидация параметров
	if err := validation.ValidateMetricName(metricName); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := validation.ValidateMetricType(metricType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем значение метрики
	value, found, err := h.getMetricValue(metricType, metricName)
	if err != nil {
		log.Printf("Error getting metric value: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

// getMetricValue получает значение метрики по типу и имени
func (h *MetricsHandler) getMetricValue(metricType, metricName string) (string, bool, error) {
	switch metricType {
	case models.Gauge:
		value, found, err := h.service.GetGauge(metricName)
		if err != nil {
			return "", false, err
		}
		if found {
			return strconv.FormatFloat(value, 'f', -1, 64), true, nil
		}
		return "", false, nil
	case models.Counter:
		value, found, err := h.service.GetCounter(metricName)
		if err != nil {
			return "", false, err
		}
		if found {
			return strconv.FormatInt(value, 10), true, nil
		}
		return "", false, nil
	default:
		return "", false, nil
	}
}

// GetAllMetrics обрабатывает GET запросы для отображения всех метрик
func (h *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные метрик
	metricsData, err := h.getAllMetricsData()
	if err != nil {
		log.Printf("Error getting metrics data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Создаем шаблон
	mt, err := template.NewMetricsTemplate()
	if err != nil {
		log.Printf("Error creating template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Выполняем шаблон
	htmlBytes, err := mt.Execute(*metricsData)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(htmlBytes)
}

// getAllMetricsData получает все данные метрик
func (h *MetricsHandler) getAllMetricsData() (*template.MetricsData, error) {
	gauges, err := h.service.GetAllGauges()
	if err != nil {
		return nil, err
	}

	counters, err := h.service.GetAllCounters()
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
