package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()

	assert.NotNil(t, storage, "Storage should not be nil")
	assert.NotNil(t, storage.Gauges, "Gauges map should be initialized")
	assert.NotNil(t, storage.Counters, "Counters map should be initialized")
	assert.Empty(t, storage.Gauges, "Gauges map should be empty initially")
	assert.Empty(t, storage.Counters, "Counters map should be empty initially")
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	storage := NewMemStorage()

	tests := []struct {
		name  string
		value float64
	}{
		{"temperature", 23.5},
		{"memory_usage", 85.7},
		{"cpu_usage", 0.0},
		{"disk_usage", -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.UpdateGauge(tt.name, tt.value)

			// Проверяем, что значение сохранилось
			value, exists := storage.GetGauge(tt.name)
			assert.True(t, exists, "Gauge should exist after update")
			assert.Equal(t, tt.value, value, "Gauge value should match")
		})
	}
}

func TestMemStorage_UpdateGauge_Replacement(t *testing.T) {
	storage := NewMemStorage()

	// Обновляем gauge метрику дважды
	storage.UpdateGauge("temperature", 23.5)
	storage.UpdateGauge("temperature", 25.0)

	// Проверяем, что значение заменилось
	value, exists := storage.GetGauge("temperature")
	assert.True(t, exists, "Gauge should exist")
	assert.Equal(t, 25.0, value, "Gauge value should be replaced")
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	storage := NewMemStorage()

	tests := []struct {
		name  string
		value int64
	}{
		{"requests", 100},
		{"errors", 0},
		{"warnings", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.UpdateCounter(tt.name, tt.value)

			// Проверяем, что значение сохранилось
			value, exists := storage.GetCounter(tt.name)
			assert.True(t, exists, "Counter should exist after update")
			assert.Equal(t, tt.value, value, "Counter value should match")
		})
	}
}

func TestMemStorage_UpdateCounter_Accumulation(t *testing.T) {
	storage := NewMemStorage()

	// Обновляем counter метрику несколько раз
	storage.UpdateCounter("requests", 100)
	storage.UpdateCounter("requests", 50)
	storage.UpdateCounter("requests", 25)

	// Проверяем, что значения накопились
	value, exists := storage.GetCounter("requests")
	assert.True(t, exists, "Counter should exist")
	assert.Equal(t, int64(175), value, "Counter value should be accumulated")
}

func TestMemStorage_GetGauge(t *testing.T) {
	storage := NewMemStorage()

	// Проверяем получение несуществующей метрики
	value, exists := storage.GetGauge("nonexistent")
	assert.False(t, exists, "Non-existent gauge should not exist")
	assert.Equal(t, 0.0, value, "Non-existent gauge should return zero value")

	// Добавляем метрику и проверяем получение
	storage.UpdateGauge("temperature", 23.5)
	value, exists = storage.GetGauge("temperature")
	assert.True(t, exists, "Existing gauge should exist")
	assert.Equal(t, 23.5, value, "Gauge value should match")
}

func TestMemStorage_GetCounter(t *testing.T) {
	storage := NewMemStorage()

	// Проверяем получение несуществующей метрики
	value, exists := storage.GetCounter("nonexistent")
	assert.False(t, exists, "Non-existent counter should not exist")
	assert.Equal(t, int64(0), value, "Non-existent counter should return zero value")

	// Добавляем метрику и проверяем получение
	storage.UpdateCounter("requests", 100)
	value, exists = storage.GetCounter("requests")
	assert.True(t, exists, "Existing counter should exist")
	assert.Equal(t, int64(100), value, "Counter value should match")
}

func TestMemStorage_GetAllGauges(t *testing.T) {
	storage := NewMemStorage()

	// Проверяем пустое хранилище
	gauges := storage.GetAllGauges()
	assert.Empty(t, gauges, "Empty storage should return empty gauges map")

	// Добавляем несколько gauge метрик
	storage.UpdateGauge("temperature", 23.5)
	storage.UpdateGauge("memory_usage", 85.7)
	storage.UpdateGauge("cpu_usage", 45.2)

	// Проверяем получение всех метрик
	gauges = storage.GetAllGauges()
	assert.Len(t, gauges, 3, "Should return all 3 gauge metrics")
	assert.Equal(t, 23.5, gauges["temperature"])
	assert.Equal(t, 85.7, gauges["memory_usage"])
	assert.Equal(t, 45.2, gauges["cpu_usage"])
}

func TestMemStorage_GetAllCounters(t *testing.T) {
	storage := NewMemStorage()

	// Проверяем пустое хранилище
	counters := storage.GetAllCounters()
	assert.Empty(t, counters, "Empty storage should return empty counters map")

	// Добавляем несколько counter метрик
	storage.UpdateCounter("requests", 100)
	storage.UpdateCounter("errors", 5)
	storage.UpdateCounter("warnings", 10)

	// Проверяем получение всех метрик
	counters = storage.GetAllCounters()
	assert.Len(t, counters, 3, "Should return all 3 counter metrics")
	assert.Equal(t, int64(100), counters["requests"])
	assert.Equal(t, int64(5), counters["errors"])
	assert.Equal(t, int64(10), counters["warnings"])
}

func TestMemStorage_Isolation(t *testing.T) {
	// Создаем два отдельных хранилища
	storage1 := NewMemStorage()
	storage2 := NewMemStorage()

	// Добавляем метрики в первое хранилище
	storage1.UpdateGauge("temperature", 23.5)
	storage1.UpdateCounter("requests", 100)

	// Проверяем, что второе хранилище не содержит метрики первого
	gaugeValue, exists := storage2.GetGauge("temperature")
	assert.False(t, exists, "Storage2 should not contain storage1's gauge")
	assert.Equal(t, 0.0, gaugeValue)

	counterValue, exists := storage2.GetCounter("requests")
	assert.False(t, exists, "Storage2 should not contain storage1's counter")
	assert.Equal(t, int64(0), counterValue)

	// Проверяем, что первое хранилище содержит свои метрики
	gaugeValue, exists = storage1.GetGauge("temperature")
	assert.True(t, exists, "Storage1 should contain its own gauge")
	assert.Equal(t, 23.5, gaugeValue)

	counterValue, exists = storage1.GetCounter("requests")
	assert.True(t, exists, "Storage1 should contain its own counter")
	assert.Equal(t, int64(100), counterValue)
}

func TestMemStorage_EdgeCases(t *testing.T) {
	storage := NewMemStorage()

	// Тестируем граничные значения для gauge
	storage.UpdateGauge("max_float", 1.7976931348623157e+308)
	storage.UpdateGauge("min_float", -1.7976931348623157e+308)
	storage.UpdateGauge("zero", 0.0)

	value, exists := storage.GetGauge("max_float")
	assert.True(t, exists)
	assert.Equal(t, 1.7976931348623157e+308, value)

	value, exists = storage.GetGauge("min_float")
	assert.True(t, exists)
	assert.Equal(t, -1.7976931348623157e+308, value)

	value, exists = storage.GetGauge("zero")
	assert.True(t, exists)
	assert.Equal(t, 0.0, value)

	// Тестируем граничные значения для counter
	storage.UpdateCounter("max_int", 9223372036854775807)
	storage.UpdateCounter("min_int", -9223372036854775808)
	storage.UpdateCounter("zero", 0)

	counterValue, exists := storage.GetCounter("max_int")
	assert.True(t, exists)
	assert.Equal(t, int64(9223372036854775807), counterValue)

	counterValue, exists = storage.GetCounter("min_int")
	assert.True(t, exists)
	assert.Equal(t, int64(-9223372036854775808), counterValue)

	counterValue, exists = storage.GetCounter("zero")
	assert.True(t, exists)
	assert.Equal(t, int64(0), counterValue)
}
