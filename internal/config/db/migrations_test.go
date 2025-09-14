package db

import (
	"context"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/jackc/pgx/v5"
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

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	require.NoError(t, err, "Failed to create test connection pool")

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

func TestValidateMigrations(t *testing.T) {
	tests := []struct {
		name        string
		migrations  []Migration
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty migrations",
			migrations:  []Migration{},
			expectError: true,
			errorMsg:    "migration 0: no migrations provided",
		},
		{
			name: "valid migrations",
			migrations: []Migration{
				{
					Version:     1,
					Description: "Test migration 1",
					Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
					Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
				},
				{
					Version:     2,
					Description: "Test migration 2",
					Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
					Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
				},
			},
			expectError: false,
		},
		{
			name: "duplicate versions",
			migrations: []Migration{
				{
					Version:     1,
					Description: "Test migration 1",
					Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
					Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
				},
				{
					Version:     1,
					Description: "Test migration 1 duplicate",
					Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
					Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
				},
			},
			expectError: true,
			errorMsg:    "migration 1: duplicate version",
		},
		{
			name: "zero version",
			migrations: []Migration{
				{
					Version:     0,
					Description: "Test migration with zero version",
					Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
					Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
				},
			},
			expectError: true,
			errorMsg:    "migration 0: version must be positive",
		},
		{
			name: "empty description",
			migrations: []Migration{
				{
					Version:     1,
					Description: "",
					Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
					Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
				},
			},
			expectError: true,
			errorMsg:    "migration 1: description cannot be empty",
		},
		{
			name: "nil Up function",
			migrations: []Migration{
				{
					Version:     1,
					Description: "Test migration with nil Up",
					Up:          nil,
					Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
				},
			},
			expectError: true,
			errorMsg:    "migration 1: Up function cannot be nil",
		},
		{
			name: "nil Down function",
			migrations: []Migration{
				{
					Version:     1,
					Description: "Test migration with nil Down",
					Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
					Down:        nil,
				},
			},
			expectError: true,
			errorMsg:    "migration 1: Down function cannot be nil",
		},
		{
			name: "missing version",
			migrations: []Migration{
				{
					Version:     1,
					Description: "Test migration 1",
					Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
					Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
				},
				{
					Version:     3,
					Description: "Test migration 3",
					Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
					Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
				},
			},
			expectError: true,
			errorMsg:    "migration 2: missing migration version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMigrations(tt.migrations)

			if tt.expectError {
				require.Error(t, err, "Expected error but got none")
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error(), "Error message should match expected")
				}
			} else {
				assert.NoError(t, err, "Expected no error")
			}
		})
	}
}

func TestMigrationError(t *testing.T) {
	err := &MigrationError{
		Version: 1,
		Message: "test error",
	}

	expected := "migration 1: test error"
	assert.Equal(t, expected, err.Error(), "Error message should match expected format")
}

func TestGetMigrationStats(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	conn := &Connection{pool: pool}
	ctx := context.Background()

	// Сначала создаем таблицу миграций
	tx, err := pool.Begin(ctx)
	require.NoError(t, err, "Failed to begin transaction")
	defer tx.Rollback(ctx)

	err = CreateMigrationsTable(ctx, tx)
	require.NoError(t, err, "Failed to create migrations table")

	// Очищаем существующие миграции перед тестом
	_, err = tx.Exec(ctx, "DELETE FROM schema_migrations")
	require.NoError(t, err, "Failed to clear existing migrations")

	// Добавляем тестовые миграции
	testMigrations := []struct {
		version     int
		description string
		appliedAt   time.Time
	}{
		{1, "Test migration 1", time.Now().Add(-2 * time.Hour)},
		{2, "Test migration 2", time.Now().Add(-1 * time.Hour)},
	}

	for _, m := range testMigrations {
		_, err := tx.Exec(ctx, insertMigrationSQL, m.version, m.description, m.appliedAt)
		require.NoError(t, err, "Failed to insert test migration")
	}

	err = tx.Commit(ctx)
	require.NoError(t, err, "Failed to commit transaction")

	// Тестируем GetMigrationStats
	stats, err := GetMigrationStats(ctx, conn)
	require.NoError(t, err, "Failed to get migration stats")

	assert.Equal(t, 2, stats.TotalMigrations, "Should have 2 total migrations")
	assert.Equal(t, 2, stats.LatestVersion, "Latest version should be 2")
	assert.False(t, stats.FirstMigration.IsZero(), "First migration time should be set")
	assert.False(t, stats.LastMigration.IsZero(), "Last migration time should be set")
}

func TestGetMigrationHistory(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	conn := &Connection{pool: pool}
	ctx := context.Background()

	// Сначала создаем таблицу миграций
	tx, err := pool.Begin(ctx)
	require.NoError(t, err, "Failed to begin transaction")
	defer tx.Rollback(ctx)

	err = CreateMigrationsTable(ctx, tx)
	require.NoError(t, err, "Failed to create migrations table")

	// Очищаем существующие миграции перед тестом
	_, err = tx.Exec(ctx, "DELETE FROM schema_migrations")
	require.NoError(t, err, "Failed to clear existing migrations")

	// Добавляем тестовые миграции
	testMigrations := []struct {
		version     int
		description string
		appliedAt   time.Time
	}{
		{1, "Test migration 1", time.Now().Add(-2 * time.Hour)},
		{2, "Test migration 2", time.Now().Add(-1 * time.Hour)},
	}

	for _, m := range testMigrations {
		_, err := tx.Exec(ctx, insertMigrationSQL, m.version, m.description, m.appliedAt)
		require.NoError(t, err, "Failed to insert test migration")
	}

	err = tx.Commit(ctx)
	require.NoError(t, err, "Failed to commit transaction")

	// Тестируем GetMigrationHistory
	history, err := GetMigrationHistory(ctx, conn)
	require.NoError(t, err, "Failed to get migration history")

	assert.Len(t, history, 2, "Should have 2 migrations in history")

	// Проверяем, что миграции отсортированы по убыванию версии
	assert.Equal(t, 2, history[0].Version, "First migration should be version 2")
	assert.Equal(t, 1, history[1].Version, "Second migration should be version 1")

	// Проверяем, что все миграции помечены как успешные
	for _, result := range history {
		assert.True(t, result.Success, "Migration %d should be marked as successful", result.Version)
	}
}

func TestRollbackMigration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	conn := &Connection{pool: pool}
	logger := logger.NewSlogLogger()
	ctx := context.Background()

	// Сначала создаем таблицу миграций
	tx, err := pool.Begin(ctx)
	require.NoError(t, err, "Failed to begin transaction")
	defer tx.Rollback(ctx)

	err = CreateMigrationsTable(ctx, tx)
	require.NoError(t, err, "Failed to create migrations table")

	// Очищаем существующие миграции перед тестом
	_, err = tx.Exec(ctx, "DELETE FROM schema_migrations")
	require.NoError(t, err, "Failed to clear existing migrations")

	// Добавляем тестовые миграции
	testMigrations := []struct {
		version     int
		description string
		appliedAt   time.Time
	}{
		{1, "Test migration 1", time.Now().Add(-2 * time.Hour)},
		{2, "Test migration 2", time.Now().Add(-1 * time.Hour)},
	}

	for _, m := range testMigrations {
		_, err := tx.Exec(ctx, insertMigrationSQL, m.version, m.description, m.appliedAt)
		require.NoError(t, err, "Failed to insert test migration")
	}

	err = tx.Commit(ctx)
	require.NoError(t, err, "Failed to commit transaction")

	// Создаем тестовые миграции для rollback
	testMigrationsList := []Migration{
		{
			Version:     1,
			Description: "Test migration 1",
			Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
			Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
		},
		{
			Version:     2,
			Description: "Test migration 2",
			Up:          func(ctx context.Context, tx pgx.Tx) error { return nil },
			Down:        func(ctx context.Context, tx pgx.Tx) error { return nil },
		},
	}

	// Временно заменяем глобальные миграции для теста
	originalMigrations := Migrations
	Migrations = testMigrationsList
	defer func() { Migrations = originalMigrations }()

	// Тестируем rollback до версии 1
	err = RollbackMigration(ctx, conn, logger, 1)
	require.NoError(t, err, "Failed to rollback migration")

	// Проверяем, что миграция 2 была удалена
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = 2").Scan(&count)
	require.NoError(t, err, "Failed to check migration count")
	assert.Equal(t, 0, count, "Migration 2 should be removed")

	// Проверяем, что миграция 1 осталась
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = 1").Scan(&count)
	require.NoError(t, err, "Failed to check migration count")
	assert.Equal(t, 1, count, "Migration 1 should remain")
}

func TestRollbackMigrationNoMigrationsToRollback(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	conn := &Connection{pool: pool}
	logger := logger.NewSlogLogger()
	ctx := context.Background()

	// Сначала создаем таблицу миграций
	tx, err := pool.Begin(ctx)
	require.NoError(t, err, "Failed to begin transaction")
	defer tx.Rollback(ctx)

	err = CreateMigrationsTable(ctx, tx)
	require.NoError(t, err, "Failed to create migrations table")

	// Очищаем существующие миграции перед тестом
	_, err = tx.Exec(ctx, "DELETE FROM schema_migrations")
	require.NoError(t, err, "Failed to clear existing migrations")

	err = tx.Commit(ctx)
	require.NoError(t, err, "Failed to commit transaction")

	// Тестируем rollback когда нет миграций для отката
	err = RollbackMigration(ctx, conn, logger, 1)
	assert.NoError(t, err, "Should not error when no migrations to rollback")
}
