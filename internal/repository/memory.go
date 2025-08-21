package repository

import (
	"context"
	"sync"

	"github.com/IgorKilipenko/metrical/internal/logger"
	models "github.com/IgorKilipenko/metrical/internal/model"
)

// InMemoryMetricsRepository реализация репозитория в памяти
type InMemoryMetricsRepository struct {
	Gauges   models.GaugeMetrics
	Counters models.CounterMetrics
	mu       sync.RWMutex // Мьютекс для потокобезопасности
	logger   logger.Logger
}

// NewInMemoryMetricsRepository создает новый экземпляр InMemoryMetricsRepository
func NewInMemoryMetricsRepository(logger logger.Logger) *InMemoryMetricsRepository {
	return &InMemoryMetricsRepository{
		Gauges:   make(models.GaugeMetrics),
		Counters: make(models.CounterMetrics),
		logger:   logger,
	}
}

// UpdateGauge обновляет значение gauge метрики
func (r *InMemoryMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during gauge update", "name", name, "value", value)
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldValue, exists := r.Gauges[name]
	r.Gauges[name] = value

	if exists {
		r.logger.Debug("updated existing gauge metric", "name", name, "old_value", oldValue, "new_value", value)
	} else {
		r.logger.Debug("created new gauge metric", "name", name, "value", value)
	}

	return nil
}

// UpdateCounter добавляет значение к counter метрике
func (r *InMemoryMetricsRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during counter update", "name", name, "value", value)
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldValue := r.Counters[name]
	r.Counters[name] += value

	r.logger.Debug("updated counter metric", "name", name, "added_value", value, "old_total", oldValue, "new_total", r.Counters[name])

	return nil
}

// GetGauge возвращает значение gauge метрики
func (r *InMemoryMetricsRepository) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during gauge retrieval", "name", name)
		return 0, false, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	value, exists := r.Gauges[name]

	if exists {
		r.logger.Debug("retrieved gauge metric", "name", name, "value", value)
	} else {
		r.logger.Debug("gauge metric not found", "name", name)
	}

	return value, exists, nil
}

// GetCounter возвращает значение counter метрики
func (r *InMemoryMetricsRepository) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during counter retrieval", "name", name)
		return 0, false, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	value, exists := r.Counters[name]

	if exists {
		r.logger.Debug("retrieved counter metric", "name", name, "value", value)
	} else {
		r.logger.Debug("counter metric not found", "name", name)
	}

	return value, exists, nil
}

// GetAllGauges возвращает все gauge метрики
func (r *InMemoryMetricsRepository) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during getAllGauges")
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Создаем копию для безопасного возврата
	result := make(models.GaugeMetrics, len(r.Gauges))
	for k, v := range r.Gauges {
		result[k] = v
	}

	r.logger.Debug("retrieved all gauge metrics", "count", len(result))
	return result, nil
}

// GetAllCounters возвращает все counter метрики
func (r *InMemoryMetricsRepository) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during getAllCounters")
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Создаем копию для безопасного возврата
	result := make(models.CounterMetrics, len(r.Counters))
	for k, v := range r.Counters {
		result[k] = v
	}

	r.logger.Debug("retrieved all counter metrics", "count", len(result))
	return result, nil
}
