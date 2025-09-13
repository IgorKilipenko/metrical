package db

import (
	"context"
	"fmt"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL константы для миграций
const (
	// Создание таблицы метрик
	createMetricsTableSQL = `
CREATE TABLE IF NOT EXISTS metrics (
	id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	type VARCHAR(50) NOT NULL,
	value DOUBLE PRECISION,
	delta BIGINT,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW(),
	UNIQUE(name, type)
);`

	// Создание индексов для таблицы метрик
	createMetricsIndexesSQL = `
CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);
CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type);
CREATE INDEX IF NOT EXISTS idx_metrics_updated_at ON metrics(updated_at);`

	// Создание таблицы миграций
	createMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	description VARCHAR(255) NOT NULL,
	applied_at TIMESTAMP DEFAULT NOW()
);`

	// Запросы для работы с миграциями
	selectMigrationsSQL = "SELECT version FROM schema_migrations ORDER BY version"
	insertMigrationSQL  = "INSERT INTO schema_migrations (version, description, applied_at) VALUES ($1, $2, $3)"

	// Удаление таблиц и индексов (для rollback)
	dropMetricsTableSQL   = "DROP TABLE IF EXISTS metrics;"
	dropMetricsIndexesSQL = `
DROP INDEX IF EXISTS idx_metrics_name;
DROP INDEX IF EXISTS idx_metrics_type;
DROP INDEX IF EXISTS idx_metrics_updated_at;`

	// Rollback запросы
	rollbackMigrationSQL = "DELETE FROM schema_migrations WHERE version = $1"

	// Метрики миграций
	selectMigrationStatsSQL = `
SELECT 
	COUNT(*) as total_migrations,
	MAX(version) as latest_version,
	MIN(applied_at) as first_migration,
	MAX(applied_at) as last_migration
FROM schema_migrations`
)

// Migration представляет миграцию базы данных
type Migration struct {
	Version     int
	Description string
	Up          func(ctx context.Context, tx pgx.Tx) error
	Down        func(ctx context.Context, tx pgx.Tx) error
}

// MigrationError представляет ошибку валидации миграции
type MigrationError struct {
	Version int
	Message string
}

func (e *MigrationError) Error() string {
	return fmt.Sprintf("migration %d: %s", e.Version, e.Message)
}

// MigrationStats представляет статистику миграций
type MigrationStats struct {
	TotalMigrations int       `json:"total_migrations"`
	LatestVersion   int       `json:"latest_version"`
	FirstMigration  time.Time `json:"first_migration"`
	LastMigration   time.Time `json:"last_migration"`
}

// MigrationResult представляет результат выполнения миграции
type MigrationResult struct {
	Version     int           `json:"version"`
	Description string        `json:"description"`
	AppliedAt   time.Time     `json:"applied_at"`
	Duration    time.Duration `json:"duration"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
}

// ValidateMigrations проверяет корректность миграций
func ValidateMigrations(migrations []Migration) error {
	if len(migrations) == 0 {
		return &MigrationError{Version: 0, Message: "no migrations provided"}
	}

	// Проверяем уникальность версий
	versions := make(map[int]bool)
	for _, migration := range migrations {
		if versions[migration.Version] {
			return &MigrationError{Version: migration.Version, Message: "duplicate version"}
		}
		versions[migration.Version] = true

		// Проверяем обязательные поля
		if migration.Version <= 0 {
			return &MigrationError{Version: migration.Version, Message: "version must be positive"}
		}
		if migration.Description == "" {
			return &MigrationError{Version: migration.Version, Message: "description cannot be empty"}
		}
		if migration.Up == nil {
			return &MigrationError{Version: migration.Version, Message: "Up function cannot be nil"}
		}
		if migration.Down == nil {
			return &MigrationError{Version: migration.Version, Message: "Down function cannot be nil"}
		}
	}

	// Проверяем последовательность версий (опционально)
	for i := 1; i <= len(migrations); i++ {
		if !versions[i] {
			return &MigrationError{Version: i, Message: "missing migration version"}
		}
	}

	return nil
}

// CreateMetricsTable создает таблицу метрик
func CreateMetricsTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, createMetricsTableSQL)
	return err
}

// CreateMetricsIndexes создает индексы для таблицы метрик
func CreateMetricsIndexes(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, createMetricsIndexesSQL)
	return err
}

// CreateMigrationsTable создает таблицу для отслеживания миграций
func CreateMigrationsTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, createMigrationsTableSQL)
	return err
}

// Migrations содержит все миграции
var Migrations = []Migration{
	{
		Version:     1,
		Description: "Create metrics table",
		Up:          CreateMetricsTable,
		Down: func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, dropMetricsTableSQL)
			return err
		},
	},
	{
		Version:     2,
		Description: "Create metrics indexes",
		Up:          CreateMetricsIndexes,
		Down: func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, dropMetricsIndexesSQL)
			return err
		},
	},
}

// Migrate выполняет миграции базы данных
func Migrate(ctx context.Context, conn *Connection, logger logger.Logger) error {
	logger.Info("starting database migrations")

	// Валидируем миграции перед выполнением
	if err := ValidateMigrations(Migrations); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}
	logger.Debug("migrations validation passed", "count", len(Migrations))

	// Создаем таблицу миграций
	logger.Debug("creating migrations table")
	tx, err := conn.Pool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := CreateMigrationsTable(ctx, tx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit migrations table transaction: %w", err)
	}
	logger.Debug("migrations table created successfully")

	// Получаем список примененных миграций
	appliedMigrations, err := getAppliedMigrations(ctx, conn.Pool())
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Применяем новые миграции
	for _, migration := range Migrations {
		if _, exists := appliedMigrations[migration.Version]; exists {
			logger.Debug("migration already applied", "version", migration.Version, "description", migration.Description)
			continue
		}

		logger.Info("applying migration", "version", migration.Version, "description", migration.Description)

		startTime := time.Now()
		tx, err := conn.Pool().Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", migration.Version, err)
		}

		// Выполняем миграцию
		if err := migration.Up(ctx, tx); err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error("failed to rollback migration", "version", migration.Version, "error", rollbackErr)
			}
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		// Записываем информацию о примененной миграции
		appliedAt := time.Now()
		if _, err := tx.Exec(ctx, insertMigrationSQL,
			migration.Version, migration.Description, appliedAt); err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error("failed to rollback migration", "version", migration.Version, "error", rollbackErr)
			}
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// Коммитим транзакцию
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		duration := time.Since(startTime)
		logger.Info("migration applied successfully",
			"version", migration.Version,
			"description", migration.Description,
			"duration", duration,
			"applied_at", appliedAt)
	}

	logger.Info("database migrations completed successfully")
	return nil
}

// getAppliedMigrations возвращает список примененных миграций
func getAppliedMigrations(ctx context.Context, pool *pgxpool.Pool) (map[int]bool, error) {
	rows, err := pool.Query(ctx, selectMigrationsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return applied, nil
}

// RollbackMigration откатывает миграцию до указанной версии
func RollbackMigration(ctx context.Context, conn *Connection, logger logger.Logger, targetVersion int) error {
	logger.Info("starting migration rollback", "target_version", targetVersion)

	// Валидируем миграции
	if err := ValidateMigrations(Migrations); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	// Получаем список примененных миграций
	appliedMigrations, err := getAppliedMigrations(ctx, conn.Pool())
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Находим миграции для отката (версии больше targetVersion)
	var migrationsToRollback []Migration
	for _, migration := range Migrations {
		if migration.Version > targetVersion && appliedMigrations[migration.Version] {
			migrationsToRollback = append(migrationsToRollback, migration)
		}
	}

	if len(migrationsToRollback) == 0 {
		logger.Info("no migrations to rollback", "target_version", targetVersion)
		return nil
	}

	// Сортируем по убыванию версии (откатываем в обратном порядке)
	for i := len(migrationsToRollback) - 1; i >= 0; i-- {
		migration := migrationsToRollback[i]
		logger.Info("rolling back migration", "version", migration.Version, "description", migration.Description)

		startTime := time.Now()
		tx, err := conn.Pool().Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for rollback %d: %w", migration.Version, err)
		}

		// Выполняем rollback
		if err := migration.Down(ctx, tx); err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error("failed to rollback transaction", "version", migration.Version, "error", rollbackErr)
			}
			return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
		}

		// Удаляем запись о миграции
		if _, err := tx.Exec(ctx, rollbackMigrationSQL, migration.Version); err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error("failed to rollback transaction", "version", migration.Version, "error", rollbackErr)
			}
			return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
		}

		// Коммитим транзакцию
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit rollback %d: %w", migration.Version, err)
		}

		duration := time.Since(startTime)
		logger.Info("migration rollback completed",
			"version", migration.Version,
			"description", migration.Description,
			"duration", duration)
	}

	logger.Info("migration rollback completed successfully", "target_version", targetVersion)
	return nil
}

// GetMigrationStats возвращает статистику миграций
func GetMigrationStats(ctx context.Context, conn *Connection) (*MigrationStats, error) {
	var stats MigrationStats
	var firstMigration, lastMigration *time.Time

	err := conn.Pool().QueryRow(ctx, selectMigrationStatsSQL).Scan(
		&stats.TotalMigrations,
		&stats.LatestVersion,
		&firstMigration,
		&lastMigration,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get migration stats: %w", err)
	}

	if firstMigration != nil {
		stats.FirstMigration = *firstMigration
	}
	if lastMigration != nil {
		stats.LastMigration = *lastMigration
	}

	return &stats, nil
}

// GetMigrationHistory возвращает историю миграций
func GetMigrationHistory(ctx context.Context, conn *Connection) ([]MigrationResult, error) {
	query := `
SELECT version, description, applied_at 
FROM schema_migrations 
ORDER BY version DESC`

	rows, err := conn.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migration history: %w", err)
	}
	defer rows.Close()

	var results []MigrationResult
	for rows.Next() {
		var result MigrationResult
		if err := rows.Scan(&result.Version, &result.Description, &result.AppliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration history: %w", err)
		}
		result.Success = true
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return results, nil
}
