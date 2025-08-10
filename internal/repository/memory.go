package repository

import (
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
func (r *InMemoryMetricsRepository) UpdateGauge(name string, value float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Gauges[name] = value
	return nil
}

// UpdateCounter добавляет значение к counter метрике
func (r *InMemoryMetricsRepository) UpdateCounter(name string, value int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Counters[name] += value
	return nil
}

// GetGauge возвращает значение gauge метрики
func (r *InMemoryMetricsRepository) GetGauge(name string) (float64, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, exists := r.Gauges[name]
	return value, exists, nil
}

// GetCounter возвращает значение counter метрики
func (r *InMemoryMetricsRepository) GetCounter(name string) (int64, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, exists := r.Counters[name]
	return value, exists, nil
}

// GetAllGauges возвращает все gauge метрики
func (r *InMemoryMetricsRepository) GetAllGauges() (models.GaugeMetrics, error) {
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
func (r *InMemoryMetricsRepository) GetAllCounters() (models.CounterMetrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Создаем копию для безопасного возврата
	result := make(models.CounterMetrics, len(r.Counters))
	for k, v := range r.Counters {
		result[k] = v
	}
	return result, nil
}
