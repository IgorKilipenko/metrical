package service

import (
	"fmt"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/validation"
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

// UpdateMetric обновляет метрику с готовыми валидированными данными
func (s *MetricsService) UpdateMetric(req *validation.MetricRequest) error {
	// Только бизнес-логика
	switch req.Type {
	case models.Gauge:
		return s.updateGaugeMetric(req.Name, req.Value.(float64))
	case models.Counter:
		return s.updateCounterMetric(req.Name, req.Value.(int64))
	default:
		return fmt.Errorf("unsupported metric type: %s", req.Type)
	}
}

// updateGaugeMetric содержит бизнес-логику для обновления gauge метрик
func (s *MetricsService) updateGaugeMetric(name string, value float64) error {
	// Здесь может быть бизнес-логика:
	// - Проверка лимитов
	// - Валидация бизнес-правил
	// - Агрегация данных
	// - Уведомления
	// - Аудит операций

	// Пока просто делегируем в репозиторий
	return s.repository.UpdateGauge(name, value)
}

// updateCounterMetric содержит бизнес-логику для обновления counter метрик
func (s *MetricsService) updateCounterMetric(name string, value int64) error {
	// Здесь может быть бизнес-логика:
	// - Проверка лимитов счетчиков
	// - Валидация бизнес-правил
	// - Агрегация данных
	// - Уведомления при превышении порогов

	// Пока просто делегируем в репозиторий
	return s.repository.UpdateCounter(name, value)
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
