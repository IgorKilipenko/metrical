package repository

import (
	"context"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryMetricsRepository_UpdateGauge(t *testing.T) {
	repo := NewInMemoryMetricsRepository(testutils.NewMockLogger())
	ctx := context.Background()

	// Обновляем gauge метрику
	err := repo.UpdateGauge(ctx, "temperature", 23.5)
	require.NoError(t, err, "Failed to update gauge metric")

	// Проверяем, что значение сохранилось
	value, exists, err := repo.GetGauge(ctx, "temperature")
	require.NoError(t, err, "Failed to get gauge metric")
	assert.True(t, exists, "Gauge metric should exist")
	assert.Equal(t, 23.5, value, "Gauge value should match")
}

func TestInMemoryMetricsRepository_UpdateCounter(t *testing.T) {
	repo := NewInMemoryMetricsRepository(testutils.NewMockLogger())
	ctx := context.Background()

	// Обновляем counter метрику
	err := repo.UpdateCounter(ctx, "requests", 100)
	require.NoError(t, err, "Failed to update counter metric")

	// Проверяем, что значение сохранилось
	value, exists, err := repo.GetCounter(ctx, "requests")
	require.NoError(t, err, "Failed to get counter metric")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(100), value, "Counter value should match")
}

func TestInMemoryMetricsRepository_GetGauge_NotExists(t *testing.T) {
	repo := NewInMemoryMetricsRepository(testutils.NewMockLogger())
	ctx := context.Background()

	// Проверяем несуществующую метрику
	value, exists, err := repo.GetGauge(ctx, "nonexistent")
	require.NoError(t, err, "Failed to get non-existent gauge metric")
	assert.False(t, exists, "Non-existent gauge should not exist")
	assert.Equal(t, 0.0, value, "Non-existent gauge should return 0")
}

func TestInMemoryMetricsRepository_GetCounter_NotExists(t *testing.T) {
	repo := NewInMemoryMetricsRepository(testutils.NewMockLogger())
	ctx := context.Background()

	// Проверяем несуществующую метрику
	value, exists, err := repo.GetCounter(ctx, "nonexistent")
	require.NoError(t, err, "Failed to get non-existent counter metric")
	assert.False(t, exists, "Non-existent counter should not exist")
	assert.Equal(t, int64(0), value, "Non-existent counter should return 0")
}

func TestInMemoryMetricsRepository_GetAllGauges(t *testing.T) {
	repo := NewInMemoryMetricsRepository(testutils.NewMockLogger())
	ctx := context.Background()

	// Добавляем несколько gauge метрик
	err := repo.UpdateGauge(ctx, "temp1", 10.5)
	require.NoError(t, err, "Failed to update gauge metric")

	err = repo.UpdateGauge(ctx, "temp2", 20.7)
	require.NoError(t, err, "Failed to update gauge metric")

	// Получаем все gauge метрики
	gauges, err := repo.GetAllGauges(ctx)
	require.NoError(t, err, "Failed to get all gauges")
	assert.Len(t, gauges, 2, "Should have 2 gauge metrics")
	assert.Equal(t, 10.5, gauges["temp1"], "First gauge value should match")
	assert.Equal(t, 20.7, gauges["temp2"], "Second gauge value should match")
}

func TestInMemoryMetricsRepository_GetAllCounters(t *testing.T) {
	repo := NewInMemoryMetricsRepository(testutils.NewMockLogger())
	ctx := context.Background()

	// Добавляем несколько counter метрик
	err := repo.UpdateCounter(ctx, "req1", 100)
	require.NoError(t, err, "Failed to update counter metric")

	err = repo.UpdateCounter(ctx, "req2", 200)
	require.NoError(t, err, "Failed to update counter metric")

	// Получаем все counter метрики
	counters, err := repo.GetAllCounters(ctx)
	require.NoError(t, err, "Failed to get all counters")
	assert.Len(t, counters, 2, "Should have 2 counter metrics")
	assert.Equal(t, int64(100), counters["req1"], "First counter value should match")
	assert.Equal(t, int64(200), counters["req2"], "Second counter value should match")
}

// Тесты на отмену контекста
func TestInMemoryMetricsRepository_ContextCancellation(t *testing.T) {
	repo := NewInMemoryMetricsRepository(testutils.NewMockLogger())

	tests := []struct {
		name string
		test func(context.Context) error
	}{
		{
			name: "UpdateGauge with cancelled context",
			test: func(ctx context.Context) error {
				return repo.UpdateGauge(ctx, "test", 42.0)
			},
		},
		{
			name: "UpdateCounter with cancelled context",
			test: func(ctx context.Context) error {
				return repo.UpdateCounter(ctx, "test", 42)
			},
		},
		{
			name: "GetGauge with cancelled context",
			test: func(ctx context.Context) error {
				_, _, err := repo.GetGauge(ctx, "test")
				return err
			},
		},
		{
			name: "GetCounter with cancelled context",
			test: func(ctx context.Context) error {
				_, _, err := repo.GetCounter(ctx, "test")
				return err
			},
		},
		{
			name: "GetAllGauges with cancelled context",
			test: func(ctx context.Context) error {
				_, err := repo.GetAllGauges(ctx)
				return err
			},
		},
		{
			name: "GetAllCounters with cancelled context",
			test: func(ctx context.Context) error {
				_, err := repo.GetAllCounters(ctx)
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

func TestInMemoryMetricsRepository_ContextTimeout(t *testing.T) {
	repo := NewInMemoryMetricsRepository(testutils.NewMockLogger())

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Ждем, пока контекст истечет
	time.Sleep(1 * time.Millisecond)

	// Пытаемся выполнить операцию с истекшим контекстом
	err := repo.UpdateGauge(ctx, "test", 42.0)

	// Проверяем, что получили ошибку таймаута
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestInMemoryMetricsRepository_ConcurrencyWithContext(t *testing.T) {
	repo := NewInMemoryMetricsRepository(testutils.NewMockLogger())
	ctx := context.Background()

	// Тестируем конкурентные операции с контекстом
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Обновляем gauge
			err := repo.UpdateGauge(ctx, "concurrent_gauge", float64(id))
			assert.NoError(t, err)

			// Обновляем counter
			err = repo.UpdateCounter(ctx, "concurrent_counter", int64(id))
			assert.NoError(t, err)

			// Читаем значения
			_, _, err = repo.GetGauge(ctx, "concurrent_gauge")
			assert.NoError(t, err)

			_, _, err = repo.GetCounter(ctx, "concurrent_counter")
			assert.NoError(t, err)
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Проверяем финальные значения
	value, exists, err := repo.GetCounter(ctx, "concurrent_counter")
	require.NoError(t, err)
	assert.True(t, exists)
	// Counter должен накопиться: 0+1+2+...+9 = 45
	assert.Equal(t, int64(45), value)
}
