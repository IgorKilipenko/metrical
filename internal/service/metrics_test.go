package service

import (
	"testing"

	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsService_UpdateMetric(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	tests := []struct {
		name        string
		metricReq   *validation.MetricRequest
		expectError bool
	}{
		{
			name: "Valid gauge metric",
			metricReq: &validation.MetricRequest{
				Type:  "gauge",
				Name:  "temperature",
				Value: 23.5,
			},
			expectError: false,
		},
		{
			name: "Valid counter metric",
			metricReq: &validation.MetricRequest{
				Type:  "counter",
				Name:  "requests",
				Value: int64(100),
			},
			expectError: false,
		},
		{
			name: "Invalid metric type",
			metricReq: &validation.MetricRequest{
				Type:  "invalid",
				Name:  "name",
				Value: 100,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateMetric(tt.metricReq)

			if tt.expectError {
				assert.Error(t, err, "Expected error, got nil")
			} else {
				assert.NoError(t, err, "Expected no error, got %v", err)
			}
		})
	}
}

func TestMetricsService_UpdateMetric_GaugeReplacement(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Обновляем gauge метрику дважды
	err := service.UpdateMetric(&validation.MetricRequest{
		Type:  "gauge",
		Name:  "temperature",
		Value: 23.5,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	err = service.UpdateMetric(&validation.MetricRequest{
		Type:  "gauge",
		Name:  "temperature",
		Value: 25.0,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	// Проверяем, что значение заменилось
	value, exists, err := service.GetGauge("temperature")
	require.NoError(t, err, "Failed to get gauge metric")
	assert.True(t, exists, "Gauge metric should exist")
	assert.Equal(t, 25.0, value, "Gauge value should be replaced")
}

func TestMetricsService_UpdateMetric_CounterAccumulation(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Обновляем counter метрику дважды
	err := service.UpdateMetric(&validation.MetricRequest{
		Type:  "counter",
		Name:  "requests",
		Value: int64(100),
	})
	require.NoError(t, err, "Failed to update counter metric")

	err = service.UpdateMetric(&validation.MetricRequest{
		Type:  "counter",
		Name:  "requests",
		Value: int64(50),
	})
	require.NoError(t, err, "Failed to update counter metric")

	// Проверяем, что значения накапливаются
	value, exists, err := service.GetCounter("requests")
	require.NoError(t, err, "Failed to get counter metric")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(150), value, "Counter value should accumulate")
}

func TestMetricsService_GetGauge(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Добавляем тестовую метрику
	err := service.UpdateMetric(&validation.MetricRequest{
		Type:  "gauge",
		Name:  "temperature",
		Value: 23.5,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	// Получаем значение
	value, exists, err := service.GetGauge("temperature")
	require.NoError(t, err, "Failed to get gauge metric")
	assert.True(t, exists, "Gauge metric should exist")
	assert.Equal(t, 23.5, value, "Gauge value should match")

	// Проверяем несуществующую метрику
	_, exists, err = service.GetGauge("nonexistent")
	require.NoError(t, err, "Should not error for nonexistent metric")
	assert.False(t, exists, "Nonexistent metric should not exist")
}

func TestMetricsService_GetCounter(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Добавляем тестовую метрику
	err := service.UpdateMetric(&validation.MetricRequest{
		Type:  "counter",
		Name:  "requests",
		Value: int64(100),
	})
	require.NoError(t, err, "Failed to update counter metric")

	// Получаем значение
	value, exists, err := service.GetCounter("requests")
	require.NoError(t, err, "Failed to get counter metric")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(100), value, "Counter value should match")

	// Проверяем несуществующую метрику
	_, exists, err = service.GetCounter("nonexistent")
	require.NoError(t, err, "Should not error for nonexistent metric")
	assert.False(t, exists, "Nonexistent metric should not exist")
}

func TestMetricsService_GetAllGauges(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Добавляем несколько gauge метрик
	err := service.UpdateMetric(&validation.MetricRequest{
		Type:  "gauge",
		Name:  "temp1",
		Value: 20.0,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	err = service.UpdateMetric(&validation.MetricRequest{
		Type:  "gauge",
		Name:  "temp2",
		Value: 25.5,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	// Получаем все gauge метрики
	gauges, err := service.GetAllGauges()
	require.NoError(t, err, "Failed to get all gauges")
	assert.Len(t, gauges, 2, "Should have 2 gauge metrics")
	assert.Equal(t, 20.0, gauges["temp1"], "First gauge value should match")
	assert.Equal(t, 25.5, gauges["temp2"], "Second gauge value should match")
}

func TestMetricsService_GetAllCounters(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Добавляем несколько counter метрик
	err := service.UpdateMetric(&validation.MetricRequest{
		Type:  "counter",
		Name:  "req1",
		Value: int64(100),
	})
	require.NoError(t, err, "Failed to update counter metric")

	err = service.UpdateMetric(&validation.MetricRequest{
		Type:  "counter",
		Name:  "req2",
		Value: int64(200),
	})
	require.NoError(t, err, "Failed to update counter metric")

	// Получаем все counter метрики
	counters, err := service.GetAllCounters()
	require.NoError(t, err, "Failed to get all counters")
	assert.Len(t, counters, 2, "Should have 2 counter metrics")
	assert.Equal(t, int64(100), counters["req1"], "First counter value should match")
	assert.Equal(t, int64(200), counters["req2"], "Second counter value should match")
}

func TestMetricsService_UpdateMetric_WithValidation(t *testing.T) {
	// Создаем реальный репозиторий
	repo := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repo)

	tests := []struct {
		name      string
		metricReq *validation.MetricRequest
		wantErr   bool
		errType   string
	}{
		{
			name: "Valid gauge metric - success",
			metricReq: &validation.MetricRequest{
				Type:  "gauge",
				Name:  "memory_usage",
				Value: 85.7,
			},
			wantErr: false,
		},
		{
			name: "Valid counter metric - success",
			metricReq: &validation.MetricRequest{
				Type:  "counter",
				Name:  "request_count",
				Value: int64(123),
			},
			wantErr: false,
		},
		{
			name: "Invalid metric type - error",
			metricReq: &validation.MetricRequest{
				Type:  "unknown",
				Name:  "test",
				Value: 123,
			},
			wantErr: true,
			errType: "unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateMetric(tt.metricReq)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType == "unsupported" {
					assert.Contains(t, err.Error(), "unsupported metric type")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
