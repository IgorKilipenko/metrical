package db

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration представляет одну миграцию
type Migration struct {
	Version   int
	Name      string
	SQL       string
	AppliedAt *time.Time
	Checksum  string
}

// Validate проверяет корректность миграции
func (m *Migration) Validate() error {
	if m.Version <= 0 {
		return fmt.Errorf("invalid version: %d (must be positive)", m.Version)
	}
	if m.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if m.SQL == "" {
		return fmt.Errorf("SQL cannot be empty")
	}
	if m.Checksum == "" {
		return fmt.Errorf("checksum cannot be empty")
	}
	return nil
}

// MigrationManager управляет миграциями базы данных
type MigrationManager struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewMigrationManager создает новый менеджер миграций
func NewMigrationManager(pool *pgxpool.Pool, logger logger.Logger) *MigrationManager {
	return &MigrationManager{
		pool:   pool,
		logger: logger,
	}
}

// InitMigrationsTable создает таблицу для отслеживания миграций
func (m *MigrationManager) InitMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			checksum VARCHAR(64) NOT NULL
		);
	`

	_, err := m.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	m.logger.Info("Migrations table initialized")
	return nil
}

// LoadMigrationsFromFS загружает миграции из файловой системы
func (m *MigrationManager) LoadMigrationsFromFS(fsys fs.FS, dir string) ([]Migration, error) {
	var migrations []Migration

	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		// Парсим имя файла: 001_create_table.sql -> version=1, name=create_table
		filename := filepath.Base(path)
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid migration filename format: %s", filename)
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid migration version: %s", parts[0])
		}

		name := strings.TrimSuffix(parts[1], ".sql")

		// Читаем содержимое файла
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		migrations = append(migrations, Migration{
			Version:  version,
			Name:     name,
			SQL:      string(content),
			Checksum: calculateChecksum(string(content)),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	// Сортируем по версии
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// GetAppliedMigrations возвращает список примененных миграций
func (m *MigrationManager) GetAppliedMigrations(ctx context.Context) (map[int]Migration, error) {
	query := `
		SELECT version, name, applied_at, checksum 
		FROM schema_migrations 
		ORDER BY version
	`

	m.logger.Info("Executing query for applied migrations", "query", query)
	rows, err := m.pool.Query(ctx, query)
	if err != nil {
		m.logger.Error("Failed to query applied migrations", "error", err)
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]Migration)
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.Version, &migration.Name, &migration.AppliedAt, &migration.Checksum)
		if err != nil {
			m.logger.Error("Failed to scan migration", "error", err)
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}
		applied[migration.Version] = migration
	}

	m.logger.Info("Found applied migrations", "count", len(applied))
	return applied, nil
}

// ApplyMigration применяет одну миграцию
func (m *MigrationManager) ApplyMigration(ctx context.Context, migration Migration) error {
	// Валидируем миграцию перед применением
	if err := migration.Validate(); err != nil {
		return fmt.Errorf("invalid migration: %w", err)
	}

	m.logger.Info("Applying migration", "version", migration.Version, "name", migration.Name)

	// Начинаем транзакцию
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Выполняем SQL миграции
	_, err = tx.Exec(ctx, migration.SQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
	}

	// Записываем информацию о миграции
	insertQuery := `
		INSERT INTO schema_migrations (version, name, checksum) 
		VALUES ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, insertQuery, migration.Version, migration.Name, migration.Checksum)
	if err != nil {
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}

	// Подтверждаем транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
	}

	m.logger.Info("Migration applied successfully", "version", migration.Version, "name", migration.Name)
	return nil
}

// RunMigrations выполняет все непримененные миграции
func (m *MigrationManager) RunMigrations(ctx context.Context, migrations []Migration) error {
	// Инициализируем таблицу миграций
	m.logger.Info("Initializing migrations table")
	err := m.InitMigrationsTable(ctx)
	if err != nil {
		m.logger.Error("Failed to init migrations table", "error", err)
		return fmt.Errorf("failed to init migrations table: %w", err)
	}

	// Получаем список примененных миграций
	m.logger.Info("Getting applied migrations")
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		m.logger.Error("Failed to get applied migrations", "error", err)
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	m.logger.Info("Found applied migrations", "count", len(applied))

	// Применяем непримененные миграции
	for _, migration := range migrations {
		if appliedMigration, exists := applied[migration.Version]; exists {
			// Проверяем контрольную сумму
			if appliedMigration.Checksum != migration.Checksum {
				return fmt.Errorf("migration %d checksum mismatch: applied=%s, current=%s",
					migration.Version, appliedMigration.Checksum, migration.Checksum)
			}
			m.logger.Info("Migration already applied", "version", migration.Version, "name", migration.Name)
			continue
		}

		m.logger.Info("Applying new migration", "version", migration.Version, "name", migration.Name)
		err := m.ApplyMigration(ctx, migration)
		if err != nil {
			m.logger.Error("Failed to apply migration", "version", migration.Version, "error", err)
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	m.logger.Info("All migrations completed successfully")
	return nil
}

// GetMigrationStats возвращает статистику миграций
func (m *MigrationManager) GetMigrationStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_migrations,
			MAX(applied_at) as last_migration_time
		FROM schema_migrations
	`

	var totalMigrations int
	var lastMigrationTime *time.Time

	err := m.pool.QueryRow(ctx, query).Scan(&totalMigrations, &lastMigrationTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_migrations": totalMigrations,
	}

	if lastMigrationTime != nil {
		stats["last_migration_time"] = *lastMigrationTime
	}

	return stats, nil
}

// calculateChecksum вычисляет контрольную сумму для миграции
func calculateChecksum(content string) string {
	// Используем SHA-256 для надежной контрольной суммы
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}
