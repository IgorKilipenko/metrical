package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryMetricsRepository_UpdateGauge(t *testing.T) {
	repo := NewInMemoryMetricsRepository()

	// Обновляем gauge метрику
	err := repo.UpdateGauge("temperature", 23.5)
	require.NoError(t, err, "Failed to update gauge metric")

	// Проверяем, что значение сохранилось
	value, exists, err := repo.GetGauge("temperature")
	require.NoError(t, err, "Failed to get gauge metric")
	assert.True(t, exists, "Gauge metric should exist")
	assert.Equal(t, 23.5, value, "Gauge value should match")
}

func TestInMemoryMetricsRepository_UpdateCounter(t *testing.T) {
	repo := NewInMemoryMetricsRepository()

	// Обновляем counter метрику
	err := repo.UpdateCounter("requests", 100)
	require.NoError(t, err, "Failed to update counter metric")

	// Проверяем, что значение сохранилось
	value, exists, err := repo.GetCounter("requests")
	require.NoError(t, err, "Failed to get counter metric")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(100), value, "Counter value should match")
}

func TestInMemoryMetricsRepository_GetGauge_NotExists(t *testing.T) {
	repo := NewInMemoryMetricsRepository()

	// Проверяем несуществующую метрику
	value, exists, err := repo.GetGauge("nonexistent")
	require.NoError(t, err, "Failed to get non-existent gauge metric")
	assert.False(t, exists, "Non-existent gauge should not exist")
	assert.Equal(t, 0.0, value, "Non-existent gauge should return 0")
}

func TestInMemoryMetricsRepository_GetCounter_NotExists(t *testing.T) {
	repo := NewInMemoryMetricsRepository()

	// Проверяем несуществующую метрику
	value, exists, err := repo.GetCounter("nonexistent")
	require.NoError(t, err, "Failed to get non-existent counter metric")
	assert.False(t, exists, "Non-existent counter should not exist")
	assert.Equal(t, int64(0), value, "Non-existent counter should return 0")
}

func TestInMemoryMetricsRepository_GetAllGauges(t *testing.T) {
	repo := NewInMemoryMetricsRepository()

	// Добавляем несколько gauge метрик
	err := repo.UpdateGauge("temp1", 10.5)
	require.NoError(t, err, "Failed to update gauge metric")

	err = repo.UpdateGauge("temp2", 20.7)
	require.NoError(t, err, "Failed to update gauge metric")

	// Получаем все gauge метрики
	gauges, err := repo.GetAllGauges()
	require.NoError(t, err, "Failed to get all gauges")
	assert.Len(t, gauges, 2, "Should have 2 gauge metrics")
	assert.Equal(t, 10.5, gauges["temp1"], "First gauge value should match")
	assert.Equal(t, 20.7, gauges["temp2"], "Second gauge value should match")
}

func TestInMemoryMetricsRepository_GetAllCounters(t *testing.T) {
	repo := NewInMemoryMetricsRepository()

	// Добавляем несколько counter метрик
	err := repo.UpdateCounter("req1", 100)
	require.NoError(t, err, "Failed to update counter metric")

	err = repo.UpdateCounter("req2", 200)
	require.NoError(t, err, "Failed to update counter metric")

	// Получаем все counter метрики
	counters, err := repo.GetAllCounters()
	require.NoError(t, err, "Failed to get all counters")
	assert.Len(t, counters, 2, "Should have 2 counter metrics")
	assert.Equal(t, int64(100), counters["req1"], "First counter value should match")
	assert.Equal(t, int64(200), counters["req2"], "Second counter value should match")
}
