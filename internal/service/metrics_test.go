package service

import (
	"testing"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsService_UpdateMetric(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

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
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Обновляем gauge метрику дважды
	err := service.UpdateMetric("gauge", "temperature", "23.5")
	require.NoError(t, err, "Failed to update gauge metric")

	err = service.UpdateMetric("gauge", "temperature", "25.0")
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
	err := service.UpdateMetric("counter", "requests", "100")
	require.NoError(t, err, "Failed to update counter metric")

	err = service.UpdateMetric("counter", "requests", "50")
	require.NoError(t, err, "Failed to update counter metric")

	// Проверяем, что значения накопились
	value, exists, err := service.GetCounter("requests")
	require.NoError(t, err, "Failed to get counter metric")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(150), value, "Counter value should be accumulated")
}

func TestMetricsService_GetGauge(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Добавляем gauge метрику
	err := service.UpdateMetric("gauge", "temperature", "23.5")
	require.NoError(t, err, "Failed to update gauge metric")

	// Получаем значение
	value, exists, err := service.GetGauge("temperature")
	require.NoError(t, err, "Failed to get gauge metric")
	assert.True(t, exists, "Gauge metric should exist")
	assert.Equal(t, 23.5, value, "Gauge value should match")

	// Проверяем несуществующую метрику
	value, exists, err = service.GetGauge("nonexistent")
	require.NoError(t, err, "Failed to get non-existent gauge metric")
	assert.False(t, exists, "Non-existent gauge should not exist")
	assert.Equal(t, 0.0, value, "Non-existent gauge should return 0")
}

func TestMetricsService_GetCounter(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Добавляем counter метрику
	err := service.UpdateMetric("counter", "requests", "100")
	require.NoError(t, err, "Failed to update counter metric")

	// Получаем значение
	value, exists, err := service.GetCounter("requests")
	require.NoError(t, err, "Failed to get counter metric")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(100), value, "Counter value should match")

	// Проверяем несуществующую метрику
	value, exists, err = service.GetCounter("nonexistent")
	require.NoError(t, err, "Failed to get non-existent counter metric")
	assert.False(t, exists, "Non-existent counter should not exist")
	assert.Equal(t, int64(0), value, "Non-existent counter should return 0")
}

func TestMetricsService_GetAllGauges(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Добавляем несколько gauge метрик
	err := service.UpdateMetric("gauge", "temp1", "10.5")
	require.NoError(t, err, "Failed to update gauge metric")

	err = service.UpdateMetric("gauge", "temp2", "20.7")
	require.NoError(t, err, "Failed to update gauge metric")

	// Получаем все gauge метрики
	gauges, err := service.GetAllGauges()
	require.NoError(t, err, "Failed to get all gauges")
	assert.Len(t, gauges, 2, "Should have 2 gauge metrics")
	assert.Equal(t, 10.5, gauges["temp1"], "First gauge value should match")
	assert.Equal(t, 20.7, gauges["temp2"], "Second gauge value should match")
}

func TestMetricsService_GetAllCounters(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repository)

	// Добавляем несколько counter метрик
	err := service.UpdateMetric("counter", "req1", "100")
	require.NoError(t, err, "Failed to update counter metric")

	err = service.UpdateMetric("counter", "req2", "200")
	require.NoError(t, err, "Failed to update counter metric")

	// Получаем все counter метрики
	counters, err := service.GetAllCounters()
	require.NoError(t, err, "Failed to get all counters")
	assert.Len(t, counters, 2, "Should have 2 counter metrics")
	assert.Equal(t, int64(100), counters["req1"], "First counter value should match")
	assert.Equal(t, int64(200), counters["req2"], "Second counter value should match")
}

func TestMetricsService_ValidateMetricValue(t *testing.T) {
	// Создаем реальный репозиторий
	repo := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repo)

	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue string
		wantErr     bool
		errType     string
	}{
		{
			name:        "Valid gauge metric",
			metricType:  "gauge",
			metricName:  "memory_usage",
			metricValue: "85.7",
			wantErr:     false,
		},
		{
			name:        "Valid counter metric",
			metricType:  "counter",
			metricName:  "request_count",
			metricValue: "123",
			wantErr:     false,
		},
		{
			name:        "Invalid metric type",
			metricType:  "unknown",
			metricName:  "test",
			metricValue: "123",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Empty metric name",
			metricType:  "gauge",
			metricName:  "",
			metricValue: "123.45",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Invalid gauge value",
			metricType:  "gauge",
			metricName:  "test",
			metricValue: "abc",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Invalid counter value",
			metricType:  "counter",
			metricName:  "test",
			metricValue: "123.45",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Empty value",
			metricType:  "gauge",
			metricName:  "test",
			metricValue: "",
			wantErr:     true,
			errType:     "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateMetricValue(tt.metricType, tt.metricName, tt.metricValue)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType == "validation" {
					assert.True(t, models.IsValidationError(err))

					// Проверяем, что это именно ValidationError
					var validationErr models.ValidationError
					assert.ErrorAs(t, err, &validationErr)

					// Проверяем, что сообщение об ошибке информативное
					assert.Contains(t, err.Error(), "validation error")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsService_UpdateMetric_WithValidation(t *testing.T) {
	// Создаем реальный репозиторий
	repo := repository.NewInMemoryMetricsRepository()
	service := NewMetricsService(repo)

	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue string
		wantErr     bool
		errType     string
	}{
		{
			name:        "Valid gauge metric - success",
			metricType:  "gauge",
			metricName:  "memory_usage",
			metricValue: "85.7",
			wantErr:     false,
		},
		{
			name:        "Valid counter metric - success",
			metricType:  "counter",
			metricName:  "request_count",
			metricValue: "123",
			wantErr:     false,
		},
		{
			name:        "Invalid metric type - validation error",
			metricType:  "unknown",
			metricName:  "test",
			metricValue: "123",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Invalid gauge value - validation error",
			metricType:  "gauge",
			metricName:  "test",
			metricValue: "abc",
			wantErr:     true,
			errType:     "validation",
		},
		{
			name:        "Empty metric name - validation error",
			metricType:  "gauge",
			metricName:  "",
			metricValue: "123.45",
			wantErr:     true,
			errType:     "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateMetric(tt.metricType, tt.metricName, tt.metricValue)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType == "validation" {
					assert.True(t, models.IsValidationError(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
