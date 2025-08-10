package service

import (
	"fmt"
	"strconv"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/repository"
)

// MetricsService сервис для работы с метриками
type MetricsService struct {
	repository repository.MetricsRepository
}

// NewMetricsService создает новый экземпляр MetricsService
func NewMetricsService(repository repository.MetricsRepository) *MetricsService {
	return &MetricsService{
		repository: repository,
	}
}

// validateMetricValue валидирует значение метрики в зависимости от типа
func (s *MetricsService) validateMetricValue(metricType, name, value string) error {
	// Валидация типа метрики
	if metricType != models.Gauge && metricType != models.Counter {
		return models.ValidationError{
			Field:   "type",
			Value:   metricType,
			Message: "must be 'gauge' or 'counter'",
		}
	}

	// Валидация имени метрики
	if name == "" {
		return models.ValidationError{
			Field:   "name",
			Value:   name,
			Message: "cannot be empty",
		}
	}

	// Валидация значения в зависимости от типа
	switch metricType {
	case models.Gauge:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return models.ValidationError{
				Field:   "value",
				Value:   value,
				Message: "must be a valid float number",
			}
		}
	case models.Counter:
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return models.ValidationError{
				Field:   "value",
				Value:   value,
				Message: "must be a valid integer number",
			}
		}
	}

	return nil
}

// UpdateMetric обновляет метрику по типу, имени и значению
func (s *MetricsService) UpdateMetric(metricType, name, value string) error {
	// Сначала валидируем входные данные
	if err := s.validateMetricValue(metricType, name, value); err != nil {
		return err
	}

	// Если валидация прошла успешно, обновляем метрику
	switch metricType {
	case models.Gauge:
		val, _ := strconv.ParseFloat(value, 64) // Ошибка уже проверена в валидации
		return s.repository.UpdateGauge(name, val)

	case models.Counter:
		val, _ := strconv.ParseInt(value, 10, 64) // Ошибка уже проверена в валидации
		return s.repository.UpdateCounter(name, val)

	default:
		// Этот случай не должен произойти после валидации, но на всякий случай
		return fmt.Errorf("unknown metric type: %s", metricType)
	}
}

// GetGauge возвращает значение gauge метрики
func (s *MetricsService) GetGauge(name string) (float64, bool, error) {
	return s.repository.GetGauge(name)
}

// GetCounter возвращает значение counter метрики
func (s *MetricsService) GetCounter(name string) (int64, bool, error) {
	return s.repository.GetCounter(name)
}

// GetAllGauges возвращает все gauge метрики
func (s *MetricsService) GetAllGauges() (models.GaugeMetrics, error) {
	return s.repository.GetAllGauges()
}

// GetAllCounters возвращает все counter метрики
func (s *MetricsService) GetAllCounters() (models.CounterMetrics, error) {
	return s.repository.GetAllCounters()
}
