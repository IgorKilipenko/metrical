package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IgorKilipenko/metrical/internal/config/db"
	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var (
		dsn     = flag.String("dsn", "", "Database connection string")
		command = flag.String("command", "", "Command to execute: migrate, rollback, stats, history")
		version = flag.Int("version", 0, "Target version for rollback")
		format  = flag.String("format", "text", "Output format: text, json")
	)
	flag.Parse()

	if *dsn == "" {
		*dsn = os.Getenv("DATABASE_DSN")
		if *dsn == "" {
			log.Fatal("Database DSN is required. Use -dsn flag or DATABASE_DSN environment variable")
		}
	}

	if *command == "" {
		log.Fatal("Command is required. Use -command flag with one of: migrate, rollback, stats, history")
	}

	// Создаем логгер
	appLogger := logger.NewSlogLogger()
	appLogger.Info("starting migration tool", "command", *command, "dsn", maskDSN(*dsn))

	// Создаем пул соединений
	pool, err := pgxpool.New(context.Background(), *dsn)
	if err != nil {
		appLogger.Error("failed to create connection pool", "error", err)
		log.Fatal("Failed to create connection pool:", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Выполняем команду
	switch *command {
	case "migrate":
		err = runMigrate(ctx, pool, appLogger)
	case "rollback":
		if *version == 0 {
			log.Fatal("Version is required for rollback command. Use -version flag")
		}
		err = runRollback(ctx, pool, appLogger, *version)
	case "stats":
		err = runStats(ctx, pool, appLogger, *format)
	case "history":
		err = runHistory(ctx, pool, appLogger, *format)
	default:
		log.Fatal("Unknown command:", *command)
	}

	if err != nil {
		appLogger.Error("command failed", "command", *command, "error", err)
		log.Fatal("Command failed:", err)
	}

	appLogger.Info("command completed successfully", "command", *command)
}

func runMigrate(ctx context.Context, pool *pgxpool.Pool, logger logger.Logger) error {
	logger.Info("running migrations")

	// Создаем менеджер миграций
	migrationManager := db.NewMigrationManager(pool, logger)

	// Загружаем миграции из файловой системы
	migrations, err := migrationManager.LoadMigrationsFromFS(os.DirFS("."), "migrations")
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Выполняем миграции
	return migrationManager.RunMigrations(ctx, migrations)
}

func runRollback(ctx context.Context, pool *pgxpool.Pool, logger logger.Logger, version int) error {
	logger.Info("rolling back migrations", "target_version", version)
	// TODO: Реализовать rollback
	return fmt.Errorf("rollback not implemented yet")
}

func runStats(ctx context.Context, pool *pgxpool.Pool, logger logger.Logger, format string) error {
	// Создаем менеджер миграций
	migrationManager := db.NewMigrationManager(pool, logger)

	stats, err := migrationManager.GetMigrationStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration stats: %w", err)
	}

	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal stats to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	default:
		fmt.Printf("Migration Stats:\n")
		fmt.Printf("  Total migrations: %d\n", stats["total_migrations"])
		if lastTime, ok := stats["last_migration_time"]; ok && lastTime != nil {
			fmt.Printf("  Last migration: %v\n", lastTime)
		}
	}
	return nil
}

func runHistory(ctx context.Context, pool *pgxpool.Pool, logger logger.Logger, format string) error {
	// Создаем менеджер миграций
	migrationManager := db.NewMigrationManager(pool, logger)

	history, err := migrationManager.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get migration history: %w", err)
	}

	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(history, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal history to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	default:
		fmt.Printf("Migration History:\n")
		if len(history) == 0 {
			fmt.Printf("  No migrations found\n")
			return nil
		}
		for _, migration := range history {
			fmt.Printf("  %d: %s (applied at: %v)\n",
				migration.Version, migration.Name, migration.AppliedAt)
		}
	}
	return nil
}

func maskDSN(dsn string) string {
	// Простая маскировка пароля в DSN
	if len(dsn) > 20 {
		return dsn[:20] + "***"
	}
	return "***"
}
