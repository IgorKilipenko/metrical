package repository

import (
	"context"
	"sync"

	models "github.com/IgorKilipenko/metrical/internal/model"
)

// InMemoryMetricsRepository реализация репозитория в памяти
type InMemoryMetricsRepository struct {
	Gauges   models.GaugeMetrics
	Counters models.CounterMetrics
	mu       sync.RWMutex // Мьютекс для потокобезопасности
}

// NewInMemoryMetricsRepository создает новый экземпляр InMemoryMetricsRepository
func NewInMemoryMetricsRepository() *InMemoryMetricsRepository {
	return &InMemoryMetricsRepository{
		Gauges:   make(models.GaugeMetrics),
		Counters: make(models.CounterMetrics),
	}
}

// UpdateGauge обновляет значение gauge метрики
func (r *InMemoryMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.Gauges[name] = value
	return nil
}

// UpdateCounter добавляет значение к counter метрике
func (r *InMemoryMetricsRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.Counters[name] += value
	return nil
}

// GetGauge возвращает значение gauge метрики
func (r *InMemoryMetricsRepository) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return 0, false, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	value, exists := r.Gauges[name]
	return value, exists, nil
}

// GetCounter возвращает значение counter метрики
func (r *InMemoryMetricsRepository) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		return 0, false, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	value, exists := r.Counters[name]
	return value, exists, nil
}

// GetAllGauges возвращает все gauge метрики
func (r *InMemoryMetricsRepository) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
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
	return result, nil
}

// GetAllCounters возвращает все counter метрики
func (r *InMemoryMetricsRepository) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
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
	return result, nil
}
