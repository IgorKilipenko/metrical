package repository

import (
	models "github.com/IgorKilipenko/metrical/internal/model"
)

// MetricsRepository интерфейс для работы с метриками
type MetricsRepository interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGauge(name string) (float64, bool, error)
	GetCounter(name string) (int64, bool, error)
	GetAllGauges() (models.GaugeMetrics, error)
	GetAllCounters() (models.CounterMetrics, error)
}

// InMemoryMetricsRepository реализация репозитория в памяти
type InMemoryMetricsRepository struct {
	storage models.Storage
}

// NewInMemoryMetricsRepository создает новый репозиторий в памяти
func NewInMemoryMetricsRepository(storage models.Storage) *InMemoryMetricsRepository {
	return &InMemoryMetricsRepository{
		storage: storage,
	}
}

// UpdateGauge обновляет значение gauge метрики
func (r *InMemoryMetricsRepository) UpdateGauge(name string, value float64) error {
	r.storage.UpdateGauge(name, value)
	return nil
}

// UpdateCounter добавляет значение к counter метрике
func (r *InMemoryMetricsRepository) UpdateCounter(name string, value int64) error {
	r.storage.UpdateCounter(name, value)
	return nil
}

// GetGauge возвращает значение gauge метрики
func (r *InMemoryMetricsRepository) GetGauge(name string) (float64, bool, error) {
	value, exists := r.storage.GetGauge(name)
	return value, exists, nil
}

// GetCounter возвращает значение counter метрики
func (r *InMemoryMetricsRepository) GetCounter(name string) (int64, bool, error) {
	value, exists := r.storage.GetCounter(name)
	return value, exists, nil
}

// GetAllGauges возвращает все gauge метрики
func (r *InMemoryMetricsRepository) GetAllGauges() (models.GaugeMetrics, error) {
	return r.storage.GetAllGauges(), nil
}

// GetAllCounters возвращает все counter метрики
func (r *InMemoryMetricsRepository) GetAllCounters() (models.CounterMetrics, error) {
	return r.storage.GetAllCounters(), nil
}
