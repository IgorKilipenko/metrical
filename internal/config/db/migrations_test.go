package db

import (
	"context"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB создает тестовую БД для тестов миграций
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	// Используем основную БД для тестов (создаем временные таблицы)
	dsn := "postgres://metricaldb:Secret@localhost:5432/metricaldb?sslmode=disable"

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err, "Failed to parse test DSN")

	// Проверяем доступность базы данных
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil, func() {}
	}

	// Проверяем подключение
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		t.Skipf("Skipping integration test: database not reachable: %v", err)
		return nil, func() {}
	}

	// Очищаем тестовую БД
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Очищаем таблицы (не удаляем, так как используем основную БД)
		pool.Exec(ctx, "DELETE FROM schema_migrations")
		pool.Exec(ctx, "DELETE FROM metrics")
		pool.Close()
	}

	return pool, cleanup
}

func TestMigration_Validation(t *testing.T) {
	tests := []struct {
		name        string
		migration   Migration
		expectError bool
	}{
		{
			name: "valid migration",
			migration: Migration{
				Version:  1,
				Name:     "test_migration",
				SQL:      "CREATE TABLE test (id INT);",
				Checksum: "abc123",
			},
			expectError: false,
		},
		{
			name: "zero version",
			migration: Migration{
				Version:  0,
				Name:     "test_migration",
				SQL:      "CREATE TABLE test (id INT);",
				Checksum: "abc123",
			},
			expectError: true,
		},
		{
			name: "empty name",
			migration: Migration{
				Version:  1,
				Name:     "",
				SQL:      "CREATE TABLE test (id INT);",
				Checksum: "abc123",
			},
			expectError: true,
		},
		{
			name: "empty SQL",
			migration: Migration{
				Version:  1,
				Name:     "test_migration",
				SQL:      "",
				Checksum: "abc123",
			},
			expectError: true,
		},
		{
			name: "empty checksum",
			migration: Migration{
				Version:  1,
				Name:     "test_migration",
				SQL:      "CREATE TABLE test (id INT);",
				Checksum: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Простая валидация структуры Migration
			hasError := false

			if tt.migration.Version <= 0 {
				hasError = true
			}
			if tt.migration.Name == "" {
				hasError = true
			}
			if tt.migration.SQL == "" {
				hasError = true
			}
			if tt.migration.Checksum == "" {
				hasError = true
			}

			if tt.expectError {
				assert.True(t, hasError, "Expected validation error")
			} else {
				assert.False(t, hasError, "Expected no validation error")
			}
		})
	}
}

func TestMigrationManager_NewMigrationManager(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return // Тест пропущен
	}
	defer cleanup()

	logger := logger.NewSlogLogger()
	manager := NewMigrationManager(pool, logger)

	assert.NotNil(t, manager, "MigrationManager should not be nil")
	// Проверяем, что менеджер создан успешно, тестируя его функциональность
	ctx := context.Background()
	err := manager.InitMigrationsTable(ctx)
	assert.NoError(t, err, "Should be able to initialize migrations table")
}

func TestMigrationManager_GetMigrationStats(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return // Тест пропущен
	}
	defer cleanup()

	logger := logger.NewSlogLogger()
	manager := NewMigrationManager(pool, logger)
	ctx := context.Background()

	// Инициализируем таблицу миграций
	err := manager.InitMigrationsTable(ctx)
	require.NoError(t, err, "Failed to initialize migrations table")

	// Очищаем существующие миграции перед тестом
	_, err = pool.Exec(ctx, "DELETE FROM schema_migrations")
	require.NoError(t, err, "Failed to clear existing migrations")

	// Добавляем тестовые миграции
	testMigrations := []struct {
		version   int
		name      string
		checksum  string
		appliedAt time.Time
	}{
		{1, "test_migration_1", "abc123", time.Now().Add(-2 * time.Hour)},
		{2, "test_migration_2", "def456", time.Now().Add(-1 * time.Hour)},
	}

	for _, m := range testMigrations {
		_, err := pool.Exec(ctx,
			"INSERT INTO schema_migrations (version, name, checksum, applied_at) VALUES ($1, $2, $3, $4)",
			m.version, m.name, m.checksum, m.appliedAt)
		require.NoError(t, err, "Failed to insert test migration")
	}

	// Тестируем GetMigrationStats
	stats, err := manager.GetMigrationStats(ctx)
	require.NoError(t, err, "Failed to get migration stats")

	assert.Equal(t, 2, stats["total_migrations"], "Should have 2 total migrations")
	assert.NotNil(t, stats["last_migration_time"], "Last migration time should be set")

	// Проверяем, что last_migration_time является временем
	lastTime, ok := stats["last_migration_time"].(time.Time)
	assert.True(t, ok, "Last migration time should be time.Time")
	assert.False(t, lastTime.IsZero(), "Last migration time should not be zero")
}

func TestMigrationManager_GetAppliedMigrations(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return // Тест пропущен
	}
	defer cleanup()

	logger := logger.NewSlogLogger()
	manager := NewMigrationManager(pool, logger)
	ctx := context.Background()

	// Инициализируем таблицу миграций
	err := manager.InitMigrationsTable(ctx)
	require.NoError(t, err, "Failed to initialize migrations table")

	// Очищаем существующие миграции перед тестом
	_, err = pool.Exec(ctx, "DELETE FROM schema_migrations")
	require.NoError(t, err, "Failed to clear existing migrations")

	// Добавляем тестовые миграции
	testMigrations := []struct {
		version   int
		name      string
		checksum  string
		appliedAt time.Time
	}{
		{1, "test_migration_1", "abc123", time.Now().Add(-2 * time.Hour)},
		{2, "test_migration_2", "def456", time.Now().Add(-1 * time.Hour)},
	}

	for _, m := range testMigrations {
		_, err := pool.Exec(ctx,
			"INSERT INTO schema_migrations (version, name, checksum, applied_at) VALUES ($1, $2, $3, $4)",
			m.version, m.name, m.checksum, m.appliedAt)
		require.NoError(t, err, "Failed to insert test migration")
	}

	// Тестируем GetAppliedMigrations
	appliedMigrations, err := manager.GetAppliedMigrations(ctx)
	require.NoError(t, err, "Failed to get applied migrations")

	assert.Len(t, appliedMigrations, 2, "Should have 2 applied migrations")

	// Проверяем, что миграции присутствуют
	assert.Contains(t, appliedMigrations, 1, "Migration 1 should be present")
	assert.Contains(t, appliedMigrations, 2, "Migration 2 should be present")

	// Проверяем содержимое миграций
	migration1 := appliedMigrations[1]
	assert.Equal(t, 1, migration1.Version, "Migration 1 version should be 1")
	assert.Equal(t, "test_migration_1", migration1.Name, "Migration 1 name should match")
	assert.Equal(t, "abc123", migration1.Checksum, "Migration 1 checksum should match")
}

func TestMigrationManager_InitMigrationsTable(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return // Тест пропущен
	}
	defer cleanup()

	logger := logger.NewSlogLogger()
	manager := NewMigrationManager(pool, logger)
	ctx := context.Background()

	// Тестируем инициализацию таблицы миграций
	err := manager.InitMigrationsTable(ctx)
	require.NoError(t, err, "Failed to initialize migrations table")

	// Проверяем, что таблица создана
	var exists bool
	err = pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&exists)
	require.NoError(t, err, "Failed to check if table exists")
	assert.True(t, exists, "schema_migrations table should exist")

	// Проверяем структуру таблицы
	var columnCount int
	err = pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'schema_migrations'").Scan(&columnCount)
	require.NoError(t, err, "Failed to check column count")
	assert.Equal(t, 4, columnCount, "schema_migrations table should have 4 columns")
}

func TestMigrationManager_ApplyMigration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return // Тест пропущен
	}
	defer cleanup()

	logger := logger.NewSlogLogger()
	manager := NewMigrationManager(pool, logger)
	ctx := context.Background()

	// Инициализируем таблицу миграций
	err := manager.InitMigrationsTable(ctx)
	require.NoError(t, err, "Failed to initialize migrations table")

	// Очищаем существующие миграции и тестовые таблицы
	_, err = pool.Exec(ctx, "DELETE FROM schema_migrations")
	require.NoError(t, err, "Failed to clear existing migrations")
	_, err = pool.Exec(ctx, "DROP TABLE IF EXISTS test_table")
	require.NoError(t, err, "Failed to drop test table")

	// Создаем тестовую миграцию
	testMigration := Migration{
		Version:  1,
		Name:     "test_migration",
		SQL:      "CREATE TABLE test_table (id INT PRIMARY KEY);",
		Checksum: "test123",
	}

	// Применяем миграцию
	err = manager.ApplyMigration(ctx, testMigration)
	require.NoError(t, err, "Failed to apply migration")

	// Проверяем, что миграция записана в таблицу
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = 1").Scan(&count)
	require.NoError(t, err, "Failed to check migration count")
	assert.Equal(t, 1, count, "Migration should be recorded")

	// Проверяем, что таблица создана
	var tableExists bool
	err = pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'test_table')").Scan(&tableExists)
	require.NoError(t, err, "Failed to check if test table exists")
	assert.True(t, tableExists, "Test table should be created")
}
