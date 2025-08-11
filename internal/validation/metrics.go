package validation

import (
	"strconv"

	models "github.com/IgorKilipenko/metrical/internal/model"
)

// MetricRequest представляет валидированный запрос на обновление метрики
type MetricRequest struct {
	Type  string
	Name  string
	Value any // float64 для gauge, int64 для counter
}

// ValidateMetricRequest валидирует и парсит запрос на обновление метрики
func ValidateMetricRequest(metricType, name, value string) (*MetricRequest, error) {
	// Валидация типа метрики
	if metricType != models.Gauge && metricType != models.Counter {
		return nil, models.ValidationError{
			Field:   "type",
			Value:   metricType,
			Message: "must be 'gauge' or 'counter'",
		}
	}

	// Валидация имени метрики
	if name == "" {
		return nil, models.ValidationError{
			Field:   "name",
			Value:   name,
			Message: "cannot be empty",
		}
	}

	// Валидация и парсинг значения в зависимости от типа
	var parsedValue any
	switch metricType {
	case models.Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, models.ValidationError{
				Field:   "value",
				Value:   value,
				Message: "must be a valid float number",
			}
		}
		parsedValue = val
	case models.Counter:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, models.ValidationError{
				Field:   "value",
				Value:   value,
				Message: "must be a valid integer number",
			}
		}
		parsedValue = val
	}

	return &MetricRequest{
		Type:  metricType,
		Name:  name,
		Value: parsedValue,
	}, nil
}

// ValidateMetricName валидирует имя метрики
func ValidateMetricName(name string) error {
	if name == "" {
		return models.ValidationError{
			Field:   "name",
			Value:   name,
			Message: "cannot be empty",
		}
	}
	return nil
}

// ValidateMetricType валидирует тип метрики
func ValidateMetricType(metricType string) error {
	if metricType != models.Gauge && metricType != models.Counter {
		return models.ValidationError{
			Field:   "type",
			Value:   metricType,
			Message: "must be 'gauge' or 'counter'",
		}
	}
	return nil
}
