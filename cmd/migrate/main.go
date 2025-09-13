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

	// Создаем конфигурацию БД
	config := db.NewConfig()
	config.DSN = *dsn

	// Подключаемся к БД
	conn, err := db.NewConnection(config, appLogger)
	if err != nil {
		appLogger.Error("failed to connect to database", "error", err)
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Выполняем команду
	switch *command {
	case "migrate":
		err = runMigrate(ctx, conn, appLogger)
	case "rollback":
		if *version == 0 {
			log.Fatal("Version is required for rollback command. Use -version flag")
		}
		err = runRollback(ctx, conn, appLogger, *version)
	case "stats":
		err = runStats(ctx, conn, appLogger, *format)
	case "history":
		err = runHistory(ctx, conn, appLogger, *format)
	default:
		log.Fatal("Unknown command:", *command)
	}

	if err != nil {
		appLogger.Error("command failed", "command", *command, "error", err)
		log.Fatal("Command failed:", err)
	}

	appLogger.Info("command completed successfully", "command", *command)
}

func runMigrate(ctx context.Context, conn *db.Connection, logger logger.Logger) error {
	logger.Info("running migrations")
	return db.Migrate(ctx, conn, logger)
}

func runRollback(ctx context.Context, conn *db.Connection, logger logger.Logger, version int) error {
	logger.Info("rolling back migrations", "target_version", version)
	return db.RollbackMigration(ctx, conn, logger, version)
}

func runStats(ctx context.Context, conn *db.Connection, logger logger.Logger, format string) error {
	stats, err := db.GetMigrationStats(ctx, conn)
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
	case "text":
		fmt.Printf("Migration Statistics:\n")
		fmt.Printf("  Total migrations: %d\n", stats.TotalMigrations)
		fmt.Printf("  Latest version: %d\n", stats.LatestVersion)
		if !stats.FirstMigration.IsZero() {
			fmt.Printf("  First migration: %s\n", stats.FirstMigration.Format(time.RFC3339))
		}
		if !stats.LastMigration.IsZero() {
			fmt.Printf("  Last migration: %s\n", stats.LastMigration.Format(time.RFC3339))
		}
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	return nil
}

func runHistory(ctx context.Context, conn *db.Connection, logger logger.Logger, format string) error {
	history, err := db.GetMigrationHistory(ctx, conn)
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
	case "text":
		fmt.Printf("Migration History:\n")
		if len(history) == 0 {
			fmt.Printf("  No migrations found\n")
			return nil
		}
		for _, result := range history {
			status := "✅"
			if !result.Success {
				status = "❌"
			}
			fmt.Printf("  %s Version %d: %s (applied at %s)\n",
				status, result.Version, result.Description, result.AppliedAt.Format(time.RFC3339))
		}
	default:
		return fmt.Errorf("unknown format: %s", format)
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
