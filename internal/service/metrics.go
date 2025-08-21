package service

import (
	"context"
	"fmt"

	"github.com/IgorKilipenko/metrical/internal/logger"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/validation"
)

// MetricsService сервис для работы с метриками
type MetricsService struct {
	repository repository.MetricsRepository
	logger     logger.Logger
}

// NewMetricsService создает новый экземпляр MetricsService
func NewMetricsService(repository repository.MetricsRepository, logger logger.Logger) *MetricsService {
	if repository == nil {
		panic("repository cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &MetricsService{
		repository: repository,
		logger:     logger,
	}
}

// UpdateMetric обновляет метрику с готовыми валидированными данными
func (s *MetricsService) UpdateMetric(ctx context.Context, req *validation.MetricRequest) error {
	s.logger.Info("updating metric", "name", req.Name, "type", req.Type, "value", req.Value)

	// Только бизнес-логика
	switch req.Type {
	case models.Gauge:
		return s.updateGaugeMetric(ctx, req.Name, req.Value.(float64))
	case models.Counter:
		return s.updateCounterMetric(ctx, req.Name, req.Value.(int64))
	default:
		s.logger.Error("unsupported metric type", "type", req.Type, "name", req.Name)
		return fmt.Errorf("unsupported metric type: %s", req.Type)
	}
}

// updateGaugeMetric содержит бизнес-логику для обновления gauge метрик
func (s *MetricsService) updateGaugeMetric(ctx context.Context, name string, value float64) error {
	s.logger.Debug("updating gauge metric", "name", name, "value", value)

	// Здесь может быть бизнес-логика:
	// - Проверка лимитов
	// - Валидация бизнес-правил
	// - Агрегация данных
	// - Уведомления
	// - Аудит операций

	// Пока просто делегируем в репозиторий
	err := s.repository.UpdateGauge(ctx, name, value)
	if err != nil {
		s.logger.Error("failed to update gauge metric", "name", name, "value", value, "error", err)
		return err
	}

	s.logger.Debug("gauge metric updated successfully", "name", name, "value", value)
	return nil
}

// updateCounterMetric содержит бизнес-логику для обновления counter метрик
func (s *MetricsService) updateCounterMetric(ctx context.Context, name string, value int64) error {
	s.logger.Debug("updating counter metric", "name", name, "value", value)

	// Здесь может быть бизнес-логика:
	// - Проверка лимитов счетчиков
	// - Валидация бизнес-правил
	// - Агрегация данных
	// - Уведомления при превышении порогов

	// Пока просто делегируем в репозиторий
	err := s.repository.UpdateCounter(ctx, name, value)
	if err != nil {
		s.logger.Error("failed to update counter metric", "name", name, "value", value, "error", err)
		return err
	}

	s.logger.Debug("counter metric updated successfully", "name", name, "value", value)
	return nil
}

// GetGauge возвращает значение gauge метрики
func (s *MetricsService) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	s.logger.Debug("getting gauge metric", "name", name)

	value, exists, err := s.repository.GetGauge(ctx, name)
	if err != nil {
		s.logger.Error("failed to get gauge metric", "name", name, "error", err)
		return 0, false, err
	}

	if exists {
		s.logger.Debug("gauge metric retrieved", "name", name, "value", value)
	} else {
		s.logger.Debug("gauge metric not found", "name", name)
	}

	return value, exists, nil
}

// GetCounter возвращает значение counter метрики
func (s *MetricsService) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	s.logger.Debug("getting counter metric", "name", name)

	value, exists, err := s.repository.GetCounter(ctx, name)
	if err != nil {
		s.logger.Error("failed to get counter metric", "name", name, "error", err)
		return 0, false, err
	}

	if exists {
		s.logger.Debug("counter metric retrieved", "name", name, "value", value)
	} else {
		s.logger.Debug("counter metric not found", "name", name)
	}

	return value, exists, nil
}

// GetAllGauges возвращает все gauge метрики
func (s *MetricsService) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	s.logger.Debug("getting all gauge metrics")

	gauges, err := s.repository.GetAllGauges(ctx)
	if err != nil {
		s.logger.Error("failed to get all gauge metrics", "error", err)
		return nil, err
	}

	s.logger.Debug("all gauge metrics retrieved", "count", len(gauges))
	return gauges, nil
}

// GetAllCounters возвращает все counter метрики
func (s *MetricsService) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	s.logger.Debug("getting all counter metrics")

	counters, err := s.repository.GetAllCounters(ctx)
	if err != nil {
		s.logger.Error("failed to get all counter metrics", "error", err)
		return nil, err
	}

	s.logger.Debug("all counter metrics retrieved", "count", len(counters))
	return counters, nil
}
