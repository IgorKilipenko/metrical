package service

import (
	"context"
	"testing"
	"time"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/IgorKilipenko/metrical/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsService_UpdateMetric(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())
	ctx := context.Background()

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
			err := service.UpdateMetric(ctx, tt.metricReq)

			if tt.expectError {
				assert.Error(t, err, "Expected error, got nil")
			} else {
				assert.NoError(t, err, "Expected no error, got %v", err)
			}
		})
	}
}

func TestMetricsService_UpdateMetric_GaugeReplacement(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())
	ctx := context.Background()

	// Обновляем gauge метрику дважды
	err := service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "gauge",
		Name:  "temperature",
		Value: 23.5,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	err = service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "gauge",
		Name:  "temperature",
		Value: 25.0,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	// Проверяем, что значение заменилось
	value, exists, err := service.GetGauge(ctx, "temperature")
	require.NoError(t, err, "Failed to get gauge metric")
	assert.True(t, exists, "Gauge metric should exist")
	assert.Equal(t, 25.0, value, "Gauge value should be replaced")
}

func TestMetricsService_UpdateMetric_CounterAccumulation(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())
	ctx := context.Background()

	// Обновляем counter метрику дважды
	err := service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "counter",
		Name:  "requests",
		Value: int64(100),
	})
	require.NoError(t, err, "Failed to update counter metric")

	err = service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "counter",
		Name:  "requests",
		Value: int64(50),
	})
	require.NoError(t, err, "Failed to update counter metric")

	// Проверяем, что значения накапливаются
	value, exists, err := service.GetCounter(ctx, "requests")
	require.NoError(t, err, "Failed to get counter metric")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(150), value, "Counter value should accumulate")
}

func TestMetricsService_GetGauge(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())
	ctx := context.Background()

	// Добавляем тестовую метрику
	err := service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "gauge",
		Name:  "temperature",
		Value: 23.5,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	// Получаем значение
	value, exists, err := service.GetGauge(ctx, "temperature")
	require.NoError(t, err, "Failed to get gauge metric")
	assert.True(t, exists, "Gauge metric should exist")
	assert.Equal(t, 23.5, value, "Gauge value should match")

	// Проверяем несуществующую метрику
	_, exists, err = service.GetGauge(ctx, "nonexistent")
	require.NoError(t, err, "Should not error for nonexistent metric")
	assert.False(t, exists, "Nonexistent metric should not exist")
}

func TestMetricsService_GetCounter(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())
	ctx := context.Background()

	// Добавляем тестовую метрику
	err := service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "counter",
		Name:  "requests",
		Value: int64(100),
	})
	require.NoError(t, err, "Failed to update counter metric")

	// Получаем значение
	value, exists, err := service.GetCounter(ctx, "requests")
	require.NoError(t, err, "Failed to get counter metric")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(100), value, "Counter value should match")

	// Проверяем несуществующую метрику
	_, exists, err = service.GetCounter(ctx, "nonexistent")
	require.NoError(t, err, "Should not error for nonexistent metric")
	assert.False(t, exists, "Nonexistent metric should not exist")
}

func TestMetricsService_GetAllGauges(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())
	ctx := context.Background()

	// Добавляем несколько gauge метрик
	err := service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "gauge",
		Name:  "temp1",
		Value: 20.0,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	err = service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "gauge",
		Name:  "temp2",
		Value: 25.5,
	})
	require.NoError(t, err, "Failed to update gauge metric")

	// Получаем все gauge метрики
	gauges, err := service.GetAllGauges(ctx)
	require.NoError(t, err, "Failed to get all gauges")
	assert.Len(t, gauges, 2, "Should have 2 gauge metrics")
	assert.Equal(t, 20.0, gauges["temp1"], "First gauge value should match")
	assert.Equal(t, 25.5, gauges["temp2"], "Second gauge value should match")
}

func TestMetricsService_GetAllCounters(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())
	ctx := context.Background()

	// Добавляем несколько counter метрик
	err := service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "counter",
		Name:  "req1",
		Value: int64(100),
	})
	require.NoError(t, err, "Failed to update counter metric")

	err = service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "counter",
		Name:  "req2",
		Value: int64(200),
	})
	require.NoError(t, err, "Failed to update counter metric")

	// Получаем все counter метрики
	counters, err := service.GetAllCounters(ctx)
	require.NoError(t, err, "Failed to get all counters")
	assert.Len(t, counters, 2, "Should have 2 counter metrics")
	assert.Equal(t, int64(100), counters["req1"], "First counter value should match")
	assert.Equal(t, int64(200), counters["req2"], "Second counter value should match")
}

// Тесты на отмену контекста
func TestMetricsService_ContextCancellation(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())

	tests := []struct {
		name string
		test func(context.Context) error
	}{
		{
			name: "UpdateMetric with cancelled context",
			test: func(ctx context.Context) error {
				return service.UpdateMetric(ctx, &validation.MetricRequest{
					Type:  "gauge",
					Name:  "test",
					Value: 42.0,
				})
			},
		},
		{
			name: "GetGauge with cancelled context",
			test: func(ctx context.Context) error {
				_, _, err := service.GetGauge(ctx, "test")
				return err
			},
		},
		{
			name: "GetCounter with cancelled context",
			test: func(ctx context.Context) error {
				_, _, err := service.GetCounter(ctx, "test")
				return err
			},
		},
		{
			name: "GetAllGauges with cancelled context",
			test: func(ctx context.Context) error {
				_, err := service.GetAllGauges(ctx)
				return err
			},
		},
		{
			name: "GetAllCounters with cancelled context",
			test: func(ctx context.Context) error {
				_, err := service.GetAllCounters(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем контекст с отменой
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Немедленно отменяем

			// Выполняем операцию с отмененным контекстом
			err := tt.test(ctx)

			// Проверяем, что получили ошибку отмены контекста
			assert.Error(t, err)
			assert.Equal(t, context.Canceled, err)
		})
	}
}

func TestMetricsService_ContextTimeout(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Ждем, пока контекст истечет
	time.Sleep(1 * time.Millisecond)

	// Пытаемся выполнить операцию с истекшим контекстом
	err := service.UpdateMetric(ctx, &validation.MetricRequest{
		Type:  "gauge",
		Name:  "test",
		Value: 42.0,
	})

	// Проверяем, что получили ошибку таймаута
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestMetricsService_UpdateMetric_WithValidation(t *testing.T) {
	// Создаем реальный репозиторий
	repo := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repo, testutils.NewMockLogger())
	ctx := context.Background()

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
			err := service.UpdateMetric(ctx, tt.metricReq)

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

func TestMetricsService_UpdateMetricJSON(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())
	ctx := context.Background()

	tests := []struct {
		name        string
		metric      *models.Metrics
		expectError bool
	}{
		{
			name: "successful gauge metric update",
			metric: &models.Metrics{
				ID:    "TestGauge",
				MType: "gauge",
				Value: func() *float64 { v := 42.5; return &v }(),
			},
			expectError: false,
		},
		{
			name: "successful counter metric update",
			metric: &models.Metrics{
				ID:    "TestCounter",
				MType: "counter",
				Delta: func() *int64 { v := int64(100); return &v }(),
			},
			expectError: false,
		},
		{
			name: "gauge metric without value",
			metric: &models.Metrics{
				ID:    "TestGauge",
				MType: "gauge",
				Value: nil,
			},
			expectError: true,
		},
		{
			name: "counter metric without delta",
			metric: &models.Metrics{
				ID:    "TestCounter",
				MType: "counter",
				Delta: nil,
			},
			expectError: true,
		},
		{
			name: "unsupported metric type",
			metric: &models.Metrics{
				ID:    "TestMetric",
				MType: "invalid",
				Value: func() *float64 { v := 42.5; return &v }(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateMetricJSON(ctx, tt.metric)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsService_GetMetricJSON(t *testing.T) {
	repository := repository.NewInMemoryMetricsRepository(testutils.NewMockLogger(), testutils.TestMetricsFile, false)
	service := NewMetricsService(repository, testutils.NewMockLogger())
	ctx := context.Background()

	// Добавляем тестовые метрики
	gaugeMetric := &models.Metrics{
		ID:    "TestGauge",
		MType: "gauge",
		Value: func() *float64 { v := 42.5; return &v }(),
	}
	err := service.UpdateMetricJSON(ctx, gaugeMetric)
	require.NoError(t, err, "Failed to add test gauge metric")

	counterMetric := &models.Metrics{
		ID:    "TestCounter",
		MType: "counter",
		Delta: func() *int64 { v := int64(100); return &v }(),
	}
	err = service.UpdateMetricJSON(ctx, counterMetric)
	require.NoError(t, err, "Failed to add test counter metric")

	tests := []struct {
		name        string
		metric      *models.Metrics
		expectError bool
		expected    *models.Metrics
	}{
		{
			name: "successful gauge metric retrieval",
			metric: &models.Metrics{
				ID:    "TestGauge",
				MType: "gauge",
			},
			expectError: false,
			expected: &models.Metrics{
				ID:    "TestGauge",
				MType: "gauge",
				Value: func() *float64 { v := 42.5; return &v }(),
			},
		},
		{
			name: "successful counter metric retrieval",
			metric: &models.Metrics{
				ID:    "TestCounter",
				MType: "counter",
			},
			expectError: false,
			expected: &models.Metrics{
				ID:    "TestCounter",
				MType: "counter",
				Delta: func() *int64 { v := int64(100); return &v }(),
			},
		},
		{
			name: "gauge metric not found",
			metric: &models.Metrics{
				ID:    "NonExistent",
				MType: "gauge",
			},
			expectError: true,
		},
		{
			name: "counter metric not found",
			metric: &models.Metrics{
				ID:    "NonExistent",
				MType: "counter",
			},
			expectError: true,
		},
		{
			name: "unsupported metric type",
			metric: &models.Metrics{
				ID:    "TestMetric",
				MType: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetMetricJSON(ctx, tt.metric)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.MType, result.MType)
				if tt.expected.Value != nil {
					assert.Equal(t, *tt.expected.Value, *result.Value)
				}
				if tt.expected.Delta != nil {
					assert.Equal(t, *tt.expected.Delta, *result.Delta)
				}
			}
		})
	}
}
