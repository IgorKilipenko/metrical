package service

import (
	"testing"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsService_UpdateMetric(t *testing.T) {
	storage := models.NewMemStorage()
	service := NewMetricsService(storage)

	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue string
		expectError bool
	}{
		{
			name:        "Valid gauge metric",
			metricType:  "gauge",
			metricName:  "temperature",
			metricValue: "23.5",
			expectError: false,
		},
		{
			name:        "Valid counter metric",
			metricType:  "counter",
			metricName:  "requests",
			metricValue: "100",
			expectError: false,
		},
		{
			name:        "Invalid metric type",
			metricType:  "invalid",
			metricName:  "name",
			metricValue: "100",
			expectError: true,
		},
		{
			name:        "Invalid gauge value",
			metricType:  "gauge",
			metricName:  "name",
			metricValue: "invalid",
			expectError: true,
		},
		{
			name:        "Invalid counter value",
			metricType:  "counter",
			metricName:  "name",
			metricValue: "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateMetric(tt.metricType, tt.metricName, tt.metricValue)

			if tt.expectError {
				assert.Error(t, err, "Expected error, got nil")
			} else {
				assert.NoError(t, err, "Expected no error, got %v", err)
			}
		})
	}
}

func TestMetricsService_UpdateMetric_GaugeReplacement(t *testing.T) {
	storage := models.NewMemStorage()
	service := NewMetricsService(storage)

	// Обновляем gauge метрику дважды
	err := service.UpdateMetric("gauge", "temperature", "23.5")
	require.NoError(t, err, "Failed to update gauge metric")

	err = service.UpdateMetric("gauge", "temperature", "25.0")
	require.NoError(t, err, "Failed to update gauge metric")

	// Проверяем, что значение заменилось
	value, exists := service.GetGauge("temperature")
	assert.True(t, exists, "Gauge metric should exist")
	assert.Equal(t, 25.0, value, "Gauge value should be replaced")
}

func TestMetricsService_UpdateMetric_CounterAccumulation(t *testing.T) {
	storage := models.NewMemStorage()
	service := NewMetricsService(storage)

	// Обновляем counter метрику дважды
	err := service.UpdateMetric("counter", "requests", "100")
	require.NoError(t, err, "Failed to update counter metric")

	err = service.UpdateMetric("counter", "requests", "50")
	require.NoError(t, err, "Failed to update counter metric")

	// Проверяем, что значения накопились
	value, exists := service.GetCounter("requests")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(150), value, "Counter value should be accumulated")
}

func TestMetricsService_GetGauge(t *testing.T) {
	storage := models.NewMemStorage()
	service := NewMetricsService(storage)

	// Проверяем несуществующую метрику
	_, exists := service.GetGauge("nonexistent")
	assert.False(t, exists, "Non-existent gauge should not exist")

	// Добавляем метрику и проверяем
	err := service.UpdateMetric("gauge", "temperature", "23.5")
	require.NoError(t, err, "Failed to update gauge metric")

	value, exists := service.GetGauge("temperature")
	assert.True(t, exists, "Gauge metric should exist")
	assert.Equal(t, 23.5, value, "Gauge value should match")
}

func TestMetricsService_GetCounter(t *testing.T) {
	storage := models.NewMemStorage()
	service := NewMetricsService(storage)

	// Проверяем несуществующую метрику
	_, exists := service.GetCounter("nonexistent")
	assert.False(t, exists, "Non-existent counter should not exist")

	// Добавляем метрику и проверяем
	err := service.UpdateMetric("counter", "requests", "100")
	require.NoError(t, err, "Failed to update counter metric")

	value, exists := service.GetCounter("requests")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(100), value, "Counter value should match")
}

func TestMetricsService_GetAllGauges(t *testing.T) {
	storage := models.NewMemStorage()
	service := NewMetricsService(storage)

	// Добавляем несколько gauge метрик
	err1 := service.UpdateMetric("gauge", "temperature", "23.5")
	require.NoError(t, err1, "Failed to update temperature metric")

	err2 := service.UpdateMetric("gauge", "humidity", "65.2")
	require.NoError(t, err2, "Failed to update humidity metric")

	gauges := service.GetAllGauges()
	assert.Len(t, gauges, 2, "Should have 2 gauges")
	assert.Equal(t, 23.5, gauges["temperature"], "Temperature should match")
	assert.Equal(t, 65.2, gauges["humidity"], "Humidity should match")
}

func TestMetricsService_GetAllCounters(t *testing.T) {
	storage := models.NewMemStorage()
	service := NewMetricsService(storage)

	// Добавляем несколько counter метрик
	err1 := service.UpdateMetric("counter", "requests", "100")
	require.NoError(t, err1, "Failed to update requests metric")

	err2 := service.UpdateMetric("counter", "errors", "5")
	require.NoError(t, err2, "Failed to update errors metric")

	counters := service.GetAllCounters()
	assert.Len(t, counters, 2, "Should have 2 counters")
	assert.Equal(t, int64(100), counters["requests"], "Requests should match")
	assert.Equal(t, int64(5), counters["errors"], "Errors should match")
}
