package handler

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/IgorKilipenko/metrical/internal/service"
)

// MetricsHandler обработчик HTTP запросов для метрик
type MetricsHandler struct {
	service *service.MetricsService
}

// NewMetricsHandler создает новый экземпляр MetricsHandler
func NewMetricsHandler(service *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		service: service,
	}
}

// UpdateMetric обрабатывает POST запросы для обновления метрик
// Формат: POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Используем регулярное выражение для парсинга URL
	// Паттерн: /update/([^/]+)/([^/]+)/([^/]+)
	re := regexp.MustCompile(`^/update/([^/]+)/([^/]+)/([^/]+)$`)
	matches := re.FindStringSubmatch(r.URL.Path)

	if len(matches) != 4 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	metricType := matches[1]
	metricName := matches[2]
	metricValue := matches[3]

	// Проверяем, что имя метрики не пустое
	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	// Обновляем метрику
	err := h.service.UpdateMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "OK")
}
