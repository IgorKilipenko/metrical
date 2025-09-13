package db

import (
	"context"
	"fmt"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration представляет миграцию базы данных
type Migration struct {
	Version     int
	Description string
	Up          func(ctx context.Context, tx pgx.Tx) error
	Down        func(ctx context.Context, tx pgx.Tx) error
}

// CreateMetricsTable создает таблицу метрик
func CreateMetricsTable(ctx context.Context, tx pgx.Tx) error {
	query := `
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

	_, err := tx.Exec(ctx, query)
	return err
}

// CreateMetricsIndexes создает индексы для таблицы метрик
func CreateMetricsIndexes(ctx context.Context, tx pgx.Tx) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);",
		"CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type);",
		"CREATE INDEX IF NOT EXISTS idx_metrics_updated_at ON metrics(updated_at);",
	}

	for _, index := range indexes {
		if _, err := tx.Exec(ctx, index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// CreateMigrationsTable создает таблицу для отслеживания миграций
func CreateMigrationsTable(ctx context.Context, tx pgx.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		description VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP DEFAULT NOW()
	);`

	_, err := tx.Exec(ctx, query)
	return err
}

// Migrations содержит все миграции
var Migrations = []Migration{
	{
		Version:     1,
		Description: "Create metrics table",
		Up:          CreateMetricsTable,
		Down: func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "DROP TABLE IF EXISTS metrics;")
			return err
		},
	},
	{
		Version:     2,
		Description: "Create metrics indexes",
		Up:          CreateMetricsIndexes,
		Down: func(ctx context.Context, tx pgx.Tx) error {
			indexes := []string{
				"DROP INDEX IF EXISTS idx_metrics_name;",
				"DROP INDEX IF EXISTS idx_metrics_type;",
				"DROP INDEX IF EXISTS idx_metrics_updated_at;",
			}
			for _, index := range indexes {
				if _, err := tx.Exec(ctx, index); err != nil {
					return err
				}
			}
			return nil
		},
	},
}

// Migrate выполняет миграции базы данных
func Migrate(ctx context.Context, conn *Connection, logger logger.Logger) error {
	logger.Info("starting database migrations")

	// Создаем таблицу миграций
	logger.Debug("creating migrations table")
	tx, err := conn.Pool().Begin(ctx)
	if err != nil {
		logger.Error("failed to begin transaction for migrations table", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := CreateMigrationsTable(ctx, tx); err != nil {
		logger.Error("failed to create migrations table", "error", err)
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("failed to commit migrations table transaction", "error", err)
		return fmt.Errorf("failed to commit migrations table transaction: %w", err)
	}
	logger.Debug("migrations table created successfully")

	// Получаем список примененных миграций
	appliedMigrations, err := getAppliedMigrations(ctx, conn.Pool())
	if err != nil {
		logger.Error("failed to get applied migrations", "error", err)
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Применяем новые миграции
	for _, migration := range Migrations {
		if _, exists := appliedMigrations[migration.Version]; exists {
			logger.Debug("migration already applied", "version", migration.Version, "description", migration.Description)
			continue
		}

		logger.Info("applying migration", "version", migration.Version, "description", migration.Description)

		tx, err := conn.Pool().Begin(ctx)
		if err != nil {
			logger.Error("failed to begin transaction for migration", "version", migration.Version, "error", err)
			return fmt.Errorf("failed to begin transaction for migration %d: %w", migration.Version, err)
		}
		defer tx.Rollback(ctx)

		// Выполняем миграцию
		if err := migration.Up(ctx, tx); err != nil {
			logger.Error("migration failed", "version", migration.Version, "error", err)
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		// Записываем информацию о примененной миграции
		if _, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations (version, description, applied_at) VALUES ($1, $2, $3)",
			migration.Version, migration.Description, time.Now()); err != nil {
			logger.Error("failed to record migration", "version", migration.Version, "error", err)
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// Коммитим транзакцию
		if err := tx.Commit(ctx); err != nil {
			logger.Error("failed to commit migration", "version", migration.Version, "error", err)
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		logger.Info("migration applied successfully", "version", migration.Version, "description", migration.Description)
	}

	logger.Info("database migrations completed successfully")
	return nil
}

// getAppliedMigrations возвращает список примененных миграций
func getAppliedMigrations(ctx context.Context, pool *pgxpool.Pool) (map[int]bool, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
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
