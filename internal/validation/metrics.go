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

// ValidateMetricsBatch валидирует батч метрик
func ValidateMetricsBatch(metrics []models.Metrics) error {
	if metrics == nil {
		return models.ValidationError{
			Field:   "metrics",
			Value:   "nil",
			Message: "metrics slice cannot be nil",
		}
	}

	if len(metrics) == 0 {
		return models.ValidationError{
			Field:   "metrics",
			Value:   "empty slice",
			Message: "metrics slice cannot be empty",
		}
	}

	for i, metric := range metrics {
		// Валидируем ID метрики
		if err := ValidateMetricName(metric.ID); err != nil {
			return models.ValidationError{
				Field:   "id",
				Value:   metric.ID,
				Message: "validation error for metric at index " + strconv.Itoa(i) + ": " + err.Error(),
			}
		}

		// Валидируем тип метрики
		if err := ValidateMetricType(metric.MType); err != nil {
			return models.ValidationError{
				Field:   "type",
				Value:   metric.MType,
				Message: "validation error for metric at index " + strconv.Itoa(i) + ": " + err.Error(),
			}
		}

		// Валидируем значения в зависимости от типа
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return models.ValidationError{
					Field:   "value",
					Value:   "nil",
					Message: "validation error for metric at index " + strconv.Itoa(i) + ": value is required for gauge metric",
				}
			}
		case models.Counter:
			if metric.Delta == nil {
				return models.ValidationError{
					Field:   "delta",
					Value:   "nil",
					Message: "validation error for metric at index " + strconv.Itoa(i) + ": delta is required for counter metric",
				}
			}
		}
	}

	return nil
}
