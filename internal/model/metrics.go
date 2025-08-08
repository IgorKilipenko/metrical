package models

import (
	"sync"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Типы-алиасы для улучшения читаемости
type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

// Storage интерфейс для работы с хранилищем метрик
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() GaugeMetrics
	GetAllCounters() CounterMetrics
}

// MemStorage структура для хранения метрик в памяти
type MemStorage struct {
	Gauges   GaugeMetrics
	Counters CounterMetrics
	mu       sync.RWMutex // Мьютекс для потокобезопасности
}

// NewMemStorage создает новый экземпляр MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(GaugeMetrics),
		Counters: make(CounterMetrics),
	}
}

// UpdateGauge обновляет значение gauge метрики
func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Gauges[name] = value
}

// UpdateCounter добавляет значение к counter метрике
func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Counters[name] += value
}

// GetGauge возвращает значение gauge метрики
func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.Gauges[name]
	return value, exists
}

// GetCounter возвращает значение counter метрики
func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.Counters[name]
	return value, exists
}

// GetAllGauges возвращает все gauge метрики
func (m *MemStorage) GetAllGauges() GaugeMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Создаем копию для безопасного возврата
	result := make(GaugeMetrics, len(m.Gauges))
	for k, v := range m.Gauges {
		result[k] = v
	}
	return result
}

// GetAllCounters возвращает все counter метрики
func (m *MemStorage) GetAllCounters() CounterMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Создаем копию для безопасного возврата
	result := make(CounterMetrics, len(m.Counters))
	for k, v := range m.Counters {
		result[k] = v
	}
	return result
}
