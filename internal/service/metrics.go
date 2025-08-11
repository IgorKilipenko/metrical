package service

import (
	"context"
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
	if repository == nil {
		panic("repository cannot be nil")
	}

	return &MetricsService{
		repository: repository,
	}
}

// UpdateMetric обновляет метрику с готовыми валидированными данными
func (s *MetricsService) UpdateMetric(ctx context.Context, req *validation.MetricRequest) error {
	// Только бизнес-логика
	switch req.Type {
	case models.Gauge:
		return s.updateGaugeMetric(ctx, req.Name, req.Value.(float64))
	case models.Counter:
		return s.updateCounterMetric(ctx, req.Name, req.Value.(int64))
	default:
		return fmt.Errorf("unsupported metric type: %s", req.Type)
	}
}

// updateGaugeMetric содержит бизнес-логику для обновления gauge метрик
func (s *MetricsService) updateGaugeMetric(ctx context.Context, name string, value float64) error {
	// Здесь может быть бизнес-логика:
	// - Проверка лимитов
	// - Валидация бизнес-правил
	// - Агрегация данных
	// - Уведомления
	// - Аудит операций

	// Пока просто делегируем в репозиторий
	return s.repository.UpdateGauge(ctx, name, value)
}

// updateCounterMetric содержит бизнес-логику для обновления counter метрик
func (s *MetricsService) updateCounterMetric(ctx context.Context, name string, value int64) error {
	// Здесь может быть бизнес-логика:
	// - Проверка лимитов счетчиков
	// - Валидация бизнес-правил
	// - Агрегация данных
	// - Уведомления при превышении порогов

	// Пока просто делегируем в репозиторий
	return s.repository.UpdateCounter(ctx, name, value)
}

// GetGauge возвращает значение gauge метрики
func (s *MetricsService) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	return s.repository.GetGauge(ctx, name)
}

// GetCounter возвращает значение counter метрики
func (s *MetricsService) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	return s.repository.GetCounter(ctx, name)
}

// GetAllGauges возвращает все gauge метрики
func (s *MetricsService) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	return s.repository.GetAllGauges(ctx)
}

// GetAllCounters возвращает все counter метрики
func (s *MetricsService) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	return s.repository.GetAllCounters(ctx)
}
