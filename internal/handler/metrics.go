package handler

import (
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/IgorKilipenko/metrical/internal/service"
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

	// HTML шаблон для отображения метрик
	const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metric-section { margin-bottom: 30px; }
        .metric-item { 
            padding: 8px; 
            margin: 4px 0; 
            background-color: #f5f5f5; 
            border-radius: 4px;
            display: flex;
            justify-content: space-between;
        }
        .metric-name { font-weight: bold; }
        .metric-value { color: #666; }
        h2 { color: #333; border-bottom: 2px solid #ddd; padding-bottom: 10px; }
        .header { text-align: center; margin-bottom: 30px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Metrics Dashboard</h1>
        <p>Current metrics values</p>
    </div>
    
    <div class="metric-section">
        <h2>Gauge Metrics ({{.GaugeCount}})</h2>
        {{range $name, $value := .Gauges}}
        <div class="metric-item">
            <span class="metric-name">{{$name}}</span>
            <span class="metric-value">{{$value}}</span>
        </div>
        {{else}}
        <p><em>No gauge metrics available</em></p>
        {{end}}
    </div>
    
    <div class="metric-section">
        <h2>Counter Metrics ({{.CounterCount}})</h2>
        {{range $name, $value := .Counters}}
        <div class="metric-item">
            <span class="metric-name">{{$name}}</span>
            <span class="metric-value">{{$value}}</span>
        </div>
        {{else}}
        <p><em>No counter metrics available</em></p>
        {{end}}
    </div>
</body>
</html>`

	// Данные для шаблона
	data := struct {
		Gauges       map[string]float64
		Counters     map[string]int64
		GaugeCount   int
		CounterCount int
	}{
		Gauges:       gauges,
		Counters:     counters,
		GaugeCount:   len(gauges),
		CounterCount: len(counters),
	}

	// Парсим и выполняем шаблон
	tmpl, err := template.New("metrics").Parse(htmlTemplate)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
