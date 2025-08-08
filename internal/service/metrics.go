package service

import (
	"fmt"
	"strconv"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/repository"
)

// MetricsService сервис для работы с метриками
type MetricsService struct {
	repo repository.MetricsRepository
}

// NewMetricsService создает новый экземпляр MetricsService
func NewMetricsService(repo repository.MetricsRepository) *MetricsService {
	return &MetricsService{
		repo: repo,
	}
}

// UpdateMetric обновляет метрику по типу, имени и значению
func (s *MetricsService) UpdateMetric(metricType, name, value string) error {
	switch metricType {
	case models.Gauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid gauge value: %s", value)
		}
		return s.repo.UpdateGauge(name, val)

	case models.Counter:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter value: %s", value)
		}
		return s.repo.UpdateCounter(name, val)

	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}
}

// GetGauge возвращает значение gauge метрики
func (s *MetricsService) GetGauge(name string) (float64, bool, error) {
	return s.repo.GetGauge(name)
}

// GetCounter возвращает значение counter метрики
func (s *MetricsService) GetCounter(name string) (int64, bool, error) {
	return s.repo.GetCounter(name)
}

// GetAllGauges возвращает все gauge метрики
func (s *MetricsService) GetAllGauges() (models.GaugeMetrics, error) {
	return s.repo.GetAllGauges()
}

// GetAllCounters возвращает все counter метрики
func (s *MetricsService) GetAllCounters() (models.CounterMetrics, error) {
	return s.repo.GetAllCounters()
}
