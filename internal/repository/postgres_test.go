package repository

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLMetricsRepository_UpdateGauge тестирует обновление gauge метрики
func TestPostgreSQLMetricsRepository_UpdateGauge(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
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

// TestPostgreSQLMetricsRepository_UpdateCounter тестирует обновление counter метрики
func TestPostgreSQLMetricsRepository_UpdateCounter(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
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

// TestPostgreSQLMetricsRepository_UpdateCounter_Incremental тестирует инкрементальное обновление counter
func TestPostgreSQLMetricsRepository_UpdateCounter_Incremental(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Добавляем к counter несколько раз
	err := repo.UpdateCounter(ctx, "requests", 50)
	require.NoError(t, err, "Failed to update counter metric")

	err = repo.UpdateCounter(ctx, "requests", 30)
	require.NoError(t, err, "Failed to update counter metric")

	err = repo.UpdateCounter(ctx, "requests", 20)
	require.NoError(t, err, "Failed to update counter metric")

	// Проверяем, что значение накопилось
	value, exists, err := repo.GetCounter(ctx, "requests")
	require.NoError(t, err, "Failed to get counter metric")
	assert.True(t, exists, "Counter metric should exist")
	assert.Equal(t, int64(100), value, "Counter value should be sum of all updates")
}

// TestPostgreSQLMetricsRepository_GetGauge_NotExists тестирует получение несуществующей gauge метрики
func TestPostgreSQLMetricsRepository_GetGauge_NotExists(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Проверяем несуществующую метрику
	value, exists, err := repo.GetGauge(ctx, "nonexistent")
	require.NoError(t, err, "Failed to get non-existent gauge metric")
	assert.False(t, exists, "Non-existent gauge should not exist")
	assert.Equal(t, 0.0, value, "Non-existent gauge should return 0")
}

// TestPostgreSQLMetricsRepository_GetCounter_NotExists тестирует получение несуществующей counter метрики
func TestPostgreSQLMetricsRepository_GetCounter_NotExists(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Проверяем несуществующую метрику
	value, exists, err := repo.GetCounter(ctx, "nonexistent")
	require.NoError(t, err, "Failed to get non-existent counter metric")
	assert.False(t, exists, "Non-existent counter should not exist")
	assert.Equal(t, int64(0), value, "Non-existent counter should return 0")
}

// TestPostgreSQLMetricsRepository_GetAllGauges тестирует получение всех gauge метрик
func TestPostgreSQLMetricsRepository_GetAllGauges(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
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

// TestPostgreSQLMetricsRepository_GetAllCounters тестирует получение всех counter метрик
func TestPostgreSQLMetricsRepository_GetAllCounters(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
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

// TestPostgreSQLMetricsRepository_ContextCancellation тестирует отмену контекста
func TestPostgreSQLMetricsRepository_ContextCancellation(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()

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

// TestPostgreSQLMetricsRepository_ContextTimeout тестирует таймаут контекста
func TestPostgreSQLMetricsRepository_ContextTimeout(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()

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

// TestPostgreSQLMetricsRepository_Concurrency тестирует конкурентные операции
func TestPostgreSQLMetricsRepository_Concurrency(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Тестируем конкурентные операции
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

// TestPostgreSQLMetricsRepository_EdgeCases тестирует граничные случаи
func TestPostgreSQLMetricsRepository_EdgeCases(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("Zero values", func(t *testing.T) {
		// Тестируем нулевые значения
		err := repo.UpdateGauge(ctx, "zero_gauge", 0.0)
		require.NoError(t, err)

		err = repo.UpdateCounter(ctx, "zero_counter", 0)
		require.NoError(t, err)

		value, exists, err := repo.GetGauge(ctx, "zero_gauge")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, 0.0, value)

		valueInt, exists, err := repo.GetCounter(ctx, "zero_counter")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, int64(0), valueInt)
	})

	t.Run("Negative values", func(t *testing.T) {
		// Тестируем отрицательные значения
		err := repo.UpdateGauge(ctx, "negative_gauge", -10.5)
		require.NoError(t, err)

		err = repo.UpdateCounter(ctx, "negative_counter", -5)
		require.NoError(t, err)

		value, exists, err := repo.GetGauge(ctx, "negative_gauge")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, -10.5, value)

		valueInt, exists, err := repo.GetCounter(ctx, "negative_counter")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, int64(-5), valueInt)
	})

	t.Run("Large values", func(t *testing.T) {
		// Тестируем большие значения
		largeFloat := math.MaxFloat64 / 2
		largeInt := int64(9223372036854775807) // MaxInt64

		err := repo.UpdateGauge(ctx, "large_gauge", largeFloat)
		require.NoError(t, err)

		err = repo.UpdateCounter(ctx, "large_counter", largeInt)
		require.NoError(t, err)

		value, exists, err := repo.GetGauge(ctx, "large_gauge")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, largeFloat, value)

		valueInt, exists, err := repo.GetCounter(ctx, "large_counter")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, largeInt, valueInt)
	})

	t.Run("Special float values", func(t *testing.T) {
		// Тестируем специальные значения float
		specialValues := []float64{
			math.Inf(1),  // +Inf
			math.Inf(-1), // -Inf
			math.NaN(),   // NaN
		}

		for i, val := range specialValues {
			err := repo.UpdateGauge(ctx, fmt.Sprintf("special_%d", i), val)
			require.NoError(t, err)

			value, exists, err := repo.GetGauge(ctx, fmt.Sprintf("special_%d", i))
			require.NoError(t, err)
			assert.True(t, exists)

			if math.IsNaN(val) {
				assert.True(t, math.IsNaN(value), "NaN should be preserved")
			} else {
				assert.Equal(t, val, value, "Special float value should be preserved")
			}
		}
	})
}

// TestPostgreSQLMetricsRepository_InterfaceCompatibility тестирует совместимость с интерфейсом
func TestPostgreSQLMetricsRepository_InterfaceCompatibility(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()

	// Проверяем, что репозиторий реализует интерфейс MetricsRepository
	var _ MetricsRepository = repo

	// Тестируем методы заглушки
	err := repo.SaveToFile()
	assert.NoError(t, err, "SaveToFile should not return error")

	err = repo.LoadFromFile()
	assert.NoError(t, err, "LoadFromFile should not return error")

	repo.SetSyncSave(true)
	repo.SetSyncSave(false)
	// SetSyncSave не должен паниковать
}

// setupTestPostgreSQLRepo создает тестовый PostgreSQL репозиторий
func setupTestPostgreSQLRepo(t *testing.T) (*PostgreSQLMetricsRepository, func()) {
	t.Helper()

	// Проверяем, доступна ли тестовая БД
	testDSN := "postgres://test:test@localhost:5432/testdb?sslmode=disable"

	// Пытаемся подключиться к тестовой БД
	pool, err := pgxpool.New(context.Background(), testDSN)
	if err != nil {
		t.Skipf("Skipping PostgreSQL tests: failed to connect to test database: %v", err)
		return nil, func() {}
	}

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Skipping PostgreSQL tests: failed to ping test database: %v", err)
		return nil, func() {}
	}

	// Создаем тестовую таблицу
	if err := createTestTable(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Создаем репозиторий
	repo := NewPostgreSQLMetricsRepository(pool, testutils.NewMockLogger())

	// Возвращаем cleanup функцию
	cleanup := func() {
		// Очищаем тестовые данные
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := pool.Exec(ctx, "DELETE FROM metrics")
		if err != nil {
			t.Logf("Failed to clean up test data: %v", err)
		}

		pool.Close()
	}

	return repo, cleanup
}

// createTestTable создает тестовую таблицу
func createTestTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS metrics (
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			value DOUBLE PRECISION,
			delta BIGINT,
			updated_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (name, type)
		)`

	_, err := pool.Exec(ctx, query)
	return err
}
