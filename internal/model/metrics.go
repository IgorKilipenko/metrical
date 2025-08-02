package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

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
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

// MemStorage структура для хранения метрик в памяти
type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

// NewMemStorage создает новый экземпляр MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

// UpdateGauge обновляет значение gauge метрики
func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.Gauges[name] = value
}

// UpdateCounter добавляет значение к counter метрике
func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.Counters[name] += value
}

// GetGauge возвращает значение gauge метрики
func (m *MemStorage) GetGauge(name string) (float64, bool) {
	value, exists := m.Gauges[name]
	return value, exists
}

// GetCounter возвращает значение counter метрики
func (m *MemStorage) GetCounter(name string) (int64, bool) {
	value, exists := m.Counters[name]
	return value, exists
}

// GetAllGauges возвращает все gauge метрики
func (m *MemStorage) GetAllGauges() map[string]float64 {
	return m.Gauges
}

// GetAllCounters возвращает все counter метрики
func (m *MemStorage) GetAllCounters() map[string]int64 {
	return m.Counters
}
