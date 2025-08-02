package service

import (
	"fmt"
	"strconv"

	"github.com/IgorKilipenko/metrical/internal/model"
)

// MetricsService сервис для работы с метриками
type MetricsService struct {
	storage models.Storage
}

// NewMetricsService создает новый экземпляр MetricsService
func NewMetricsService(storage models.Storage) *MetricsService {
	return &MetricsService{
		storage: storage,
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
		s.storage.UpdateGauge(name, val)
		return nil

	case models.Counter:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter value: %s", value)
		}
		s.storage.UpdateCounter(name, val)
		return nil

	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}
}

// GetGauge возвращает значение gauge метрики
func (s *MetricsService) GetGauge(name string) (float64, bool) {
	return s.storage.GetGauge(name)
}

// GetCounter возвращает значение counter метрики
func (s *MetricsService) GetCounter(name string) (int64, bool) {
	return s.storage.GetCounter(name)
}

// GetAllGauges возвращает все gauge метрики
func (s *MetricsService) GetAllGauges() map[string]float64 {
	return s.storage.GetAllGauges()
}

// GetAllCounters возвращает все counter метрики
func (s *MetricsService) GetAllCounters() map[string]int64 {
	return s.storage.GetAllCounters()
}
