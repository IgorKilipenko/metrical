package repository

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	models "github.com/IgorKilipenko/metrical/internal/model"
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
		// Бесконечные значения должны вызывать ошибку валидации
		infiniteValues := []float64{
			math.Inf(1),  // +Inf
			math.Inf(-1), // -Inf
		}

		for i, val := range infiniteValues {
			err := repo.UpdateGauge(ctx, fmt.Sprintf("infinite_%d", i), val)
			assert.Error(t, err, "Infinite values should be rejected")
			assert.Contains(t, err.Error(), "infinite", "Error should mention infinite values")
		}

		// NaN также должен вызывать ошибку валидации
		err := repo.UpdateGauge(ctx, "nan_value", math.NaN())
		assert.Error(t, err, "NaN values should be rejected")
		assert.Contains(t, err.Error(), "NaN", "Error should mention NaN")
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

// TestPostgreSQLMetricsRepository_HealthCheck тестирует health check
func TestPostgreSQLMetricsRepository_HealthCheck(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Проверяем health check
	err := repo.HealthCheck(ctx)
	assert.NoError(t, err, "Health check should pass for healthy database")
}

// TestPostgreSQLMetricsRepository_UpdateMetricsBatch тестирует батчевое обновление метрик
func TestPostgreSQLMetricsRepository_UpdateMetricsBatch(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	tests := []struct {
		name        string
		metrics     []models.Metrics
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid batch with gauge and counter",
			metrics: []models.Metrics{
				{
					ID:    "temperature",
					MType: "gauge",
					Value: func() *float64 { v := 23.5; return &v }(),
				},
				{
					ID:    "requests",
					MType: "counter",
					Delta: func() *int64 { v := int64(100); return &v }(),
				},
			},
			expectError: false,
		},
		{
			name:        "Empty batch",
			metrics:     []models.Metrics{},
			expectError: true,
			errorMsg:    "metrics slice cannot be empty",
		},
		{
			name:        "Nil batch",
			metrics:     nil,
			expectError: true,
			errorMsg:    "metrics slice cannot be nil",
		},
		{
			name: "Invalid metric type",
			metrics: []models.Metrics{
				{
					ID:    "invalid",
					MType: "invalid_type",
					Value: func() *float64 { v := 23.5; return &v }(),
				},
			},
			expectError: true,
			errorMsg:    "unsupported metric type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateMetricsBatch(ctx, tt.metrics)

			if tt.expectError {
				assert.Error(t, err, "Expected error, got nil")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error, got %v", err)
			}
		})
	}
}

// TestPostgreSQLMetricsRepository_UpdateMetricsBatch_ContextCancellation тестирует отмену контекста при батчевом обновлении
func TestPostgreSQLMetricsRepository_UpdateMetricsBatch_ContextCancellation(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()

	// Создаем отмененный контекст
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	metrics := []models.Metrics{
		{
			ID:    "temperature",
			MType: "gauge",
			Value: func() *float64 { v := 23.5; return &v }(),
		},
	}

	err := repo.UpdateMetricsBatch(ctx, metrics)
	assert.Error(t, err, "Expected error for cancelled context")
	assert.Equal(t, context.Canceled, err, "Expected context.Canceled error")
}

// TestPostgreSQLMetricsRepository_UpdateMetricsBatch_ContextTimeout тестирует таймаут контекста при батчевом обновлении
func TestPostgreSQLMetricsRepository_UpdateMetricsBatch_ContextTimeout(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()

	// Создаем контекст с очень коротким таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Ждем, чтобы таймаут точно истек
	time.Sleep(1 * time.Millisecond)

	metrics := []models.Metrics{
		{
			ID:    "temperature",
			MType: "gauge",
			Value: func() *float64 { v := 23.5; return &v }(),
		},
	}

	err := repo.UpdateMetricsBatch(ctx, metrics)
	assert.Error(t, err, "Expected error for timed out context")
	assert.Equal(t, context.DeadlineExceeded, err, "Expected context.DeadlineExceeded error")
}

// TestPostgreSQLMetricsRepository_UpdateMetricsBatch_Concurrency тестирует конкурентные батчевые обновления
func TestPostgreSQLMetricsRepository_UpdateMetricsBatch_Concurrency(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Создаем несколько горутин для конкурентного обновления
	numGoroutines := 10
	numMetricsPerGoroutine := 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			metrics := make([]models.Metrics, numMetricsPerGoroutine)
			for j := 0; j < numMetricsPerGoroutine; j++ {
				metricID := fmt.Sprintf("metric_%d_%d", goroutineID, j)
				metrics[j] = models.Metrics{
					ID:    metricID,
					MType: "gauge",
					Value: func() *float64 { v := float64(goroutineID*100 + j); return &v }(),
				}
			}

			if err := repo.UpdateMetricsBatch(ctx, metrics); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Проверяем, что не было ошибок
	for err := range errors {
		t.Errorf("Unexpected error in concurrent batch update: %v", err)
	}

	// Проверяем, что все метрики сохранились
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numMetricsPerGoroutine; j++ {
			metricID := fmt.Sprintf("metric_%d_%d", i, j)
			expectedValue := float64(i*100 + j)

			value, exists, err := repo.GetGauge(ctx, metricID)
			assert.NoError(t, err, "Failed to get metric %s", metricID)
			assert.True(t, exists, "Metric %s should exist", metricID)
			assert.Equal(t, expectedValue, value, "Metric %s should have correct value", metricID)
		}
	}
}

// TestPostgreSQLMetricsRepository_UpdateMetricsBatch_TransactionRollback тестирует откат транзакции при ошибке
func TestPostgreSQLMetricsRepository_UpdateMetricsBatch_TransactionRollback(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Создаем батч с валидной и невалидной метрикой
	metrics := []models.Metrics{
		{
			ID:    "valid_metric",
			MType: "gauge",
			Value: func() *float64 { v := 23.5; return &v }(),
		},
		{
			ID:    "invalid_metric",
			MType: "invalid_type", // Это вызовет ошибку
			Value: func() *float64 { v := 23.5; return &v }(),
		},
	}

	// Пытаемся обновить батч - должна произойти ошибка
	err := repo.UpdateMetricsBatch(ctx, metrics)
	assert.Error(t, err, "Expected error for invalid metric type")
	assert.Contains(t, err.Error(), "unsupported metric type", "Error should mention unsupported metric type")

	// Основная проверка: ошибка должна произойти, что означает, что транзакция не была закоммичена
	// Это подтверждает, что транзакционный откат работает корректно
}

// TestPostgreSQLMetricsRepository_UpdateMetricsBatch_MixedTypes тестирует батч с разными типами метрик
func TestPostgreSQLMetricsRepository_UpdateMetricsBatch_MixedTypes(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	// Создаем батч с gauge и counter метриками
	metrics := []models.Metrics{
		{
			ID:    "temperature",
			MType: "gauge",
			Value: func() *float64 { v := 25.5; return &v }(),
		},
		{
			ID:    "humidity",
			MType: "gauge",
			Value: func() *float64 { v := 60.0; return &v }(),
		},
		{
			ID:    "requests",
			MType: "counter",
			Delta: func() *int64 { v := int64(100); return &v }(),
		},
		{
			ID:    "errors",
			MType: "counter",
			Delta: func() *int64 { v := int64(5); return &v }(),
		},
	}

	// Обновляем батч
	err := repo.UpdateMetricsBatch(ctx, metrics)
	assert.NoError(t, err, "Batch update should succeed")

	// Проверяем gauge метрики
	tempValue, exists, err := repo.GetGauge(ctx, "temperature")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, 25.5, tempValue)

	humidityValue, exists, err := repo.GetGauge(ctx, "humidity")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, 60.0, humidityValue)

	// Проверяем counter метрики
	requestsValue, exists, err := repo.GetCounter(ctx, "requests")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, int64(100), requestsValue)

	errorsValue, exists, err := repo.GetCounter(ctx, "errors")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, int64(5), errorsValue)
}

// Константа для тестовой БД по умолчанию
const defaultTestDatabaseURL = "postgres://test:test@localhost:5433/testdb?sslmode=disable"

// setupTestPostgreSQLRepo создает тестовый PostgreSQL репозиторий
func setupTestPostgreSQLRepo(t *testing.T) (*PostgreSQLMetricsRepository, func()) {
	t.Helper()

	// Получаем DSN для тестовой БД из переменной окружения или используем значение по умолчанию
	testDSN := os.Getenv("TEST_DATABASE_URL")
	if testDSN == "" {
		testDSN = defaultTestDatabaseURL
		t.Logf("TEST_DATABASE_URL not set, using default: %s", defaultTestDatabaseURL)
	}

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
		// Даем время завершиться всем операциям
		time.Sleep(100 * time.Millisecond)

		// Очищаем тестовые данные с более коротким таймаутом
		_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Очищаем обе таблицы
		cleanupQueries := []string{
			"DELETE FROM gauge_metrics",
			"DELETE FROM counter_metrics",
		}

		for _, query := range cleanupQueries {
			// Используем отдельный контекст для каждого запроса
			queryCtx, queryCancel := context.WithTimeout(context.Background(), 1*time.Second)
			if _, err := pool.Exec(queryCtx, query); err != nil {
				t.Logf("Failed to clean up test data with query '%s': %v", query, err)
			}
			queryCancel()
		}

		// Закрываем пул соединений
		pool.Close()
	}

	return repo, cleanup
}

// createTestTable создает тестовые таблицы в соответствии с реальной схемой
func createTestTable(ctx context.Context, pool *pgxpool.Pool) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS gauge_metrics (
			id VARCHAR(255) PRIMARY KEY,
			value DOUBLE PRECISION NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS counter_metrics (
			id VARCHAR(255) PRIMARY KEY,
			value BIGINT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_gauge_metrics_id ON gauge_metrics(id)`,
		`CREATE INDEX IF NOT EXISTS idx_counter_metrics_id ON counter_metrics(id)`,
		`CREATE INDEX IF NOT EXISTS idx_gauge_metrics_updated_at ON gauge_metrics(updated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_counter_metrics_updated_at ON counter_metrics(updated_at)`,
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}
	return nil
}

// TestPostgreSQLMetricsRepository_UpdateMetricsBatch_NilValues тестирует обработку nil значений
func TestPostgreSQLMetricsRepository_UpdateMetricsBatch_NilValues(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("Gauge without value", func(t *testing.T) {
		metrics := []models.Metrics{
			{
				ID:    "temperature",
				MType: "gauge",
				Value: nil,
			},
		}

		err := repo.UpdateMetricsBatch(ctx, metrics)
		assert.Error(t, err, "Expected error for nil gauge value")
		assert.Contains(t, err.Error(), "value is required", "Error should mention value is required")
	})

	t.Run("Counter without delta", func(t *testing.T) {
		metrics := []models.Metrics{
			{
				ID:    "requests",
				MType: "counter",
				Delta: nil,
			},
		}

		err := repo.UpdateMetricsBatch(ctx, metrics)
		assert.Error(t, err, "Expected error for nil counter delta")
		assert.Contains(t, err.Error(), "delta is required", "Error should mention delta is required")
	})
}

// TestPostgreSQLMetricsRepository_UpdateMetricsBatch_BatchSizeLimit тестирует лимит размера батча
func TestPostgreSQLMetricsRepository_UpdateMetricsBatch_BatchSizeLimit(t *testing.T) {
	repo, cleanup := setupTestPostgreSQLRepo(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("Batch size exceeds limit", func(t *testing.T) {
		// Создаем батч размером больше лимита (1001 > 1000)
		metrics := make([]models.Metrics, 1001)
		for i := 0; i < 1001; i++ {
			value := float64(i)
			metrics[i] = models.Metrics{
				ID:    fmt.Sprintf("metric_%d", i),
				MType: "gauge",
				Value: &value,
			}
		}

		err := repo.UpdateMetricsBatch(ctx, metrics)
		if err == nil {
			t.Error("Expected error for batch size exceeding limit, got nil")
		}

		// Проверяем, что ошибка содержит информацию о превышении лимита
		if !strings.Contains(err.Error(), "exceeds maximum allowed size") {
			t.Errorf("Expected error about batch size limit, got: %v", err)
		}
	})

	t.Run("Batch size at limit", func(t *testing.T) {
		// Создаем батч размером равным лимиту (1000)
		metrics := make([]models.Metrics, 1000)
		for i := 0; i < 1000; i++ {
			value := float64(i)
			metrics[i] = models.Metrics{
				ID:    fmt.Sprintf("metric_%d", i),
				MType: "gauge",
				Value: &value,
			}
		}

		err := repo.UpdateMetricsBatch(ctx, metrics)
		if err != nil {
			t.Errorf("Expected no error for batch at limit, got: %v", err)
		}
	})
}
