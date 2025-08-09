package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

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

// validatePath проверяет путь на наличие неподдерживаемых символов
func validatePath(path string) error {
	// Проверяем только на управляющие символы
	invalidPathRegex, err := regexp.Compile(`[\x00-\x1F\x7F-\x9F]`)
	if err != nil {
		return err
	}
	if invalidPathRegex.MatchString(path) {
		return fmt.Errorf("invalid path: %s", path)
	}

	return nil
}

// splitPath разбивает путь URL на части, сохраняя пустые сегменты
func splitPath(path string) ([]string, error) {
	// Валидируем путь
	if err := validatePath(path); err != nil {
		return nil, err
	}

	// Убираем начальный слеш и пробелы в конце и разбиваем по слешам
	trimRegex, err := regexp.Compile(`^/|/?\s*$`)
	if err != nil {
		return nil, err
	}
	path = trimRegex.ReplaceAllString(path, "")

	if path == "" {
		return []string{""}, nil
	}

	return strings.Split(path, "/"), nil
}
