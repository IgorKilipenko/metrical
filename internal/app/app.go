package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IgorKilipenko/metrical/internal/config/db"
	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/IgorKilipenko/metrical/internal/httpserver"
	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/routes"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/go-chi/chi/v5"
)

// Константы для таймаутов
const (
	DefaultMigrationTimeout = 30 * time.Second
	DefaultShutdownTimeout  = 30 * time.Second
)

// App представляет основное приложение
type App struct {
	server *httpserver.Server
	addr   string
	config Config
}

// Config содержит конфигурацию приложения
type Config struct {
	Addr            string // Адрес сервера (например, "localhost")
	Port            string // Порт сервера (например, "8080")
	FileStoragePath string // Путь к файлу для сохранения метрик
	Restore         bool   // Флаг для восстановления метрик из файла
	StoreInterval   int    // Интервал сохранения метрик в секундах
	DatabaseDSN     string // Строка подключения к базе данных PostgreSQL
}

// New создает новое приложение с заданной конфигурацией
func New(config Config) *App {
	return &App{
		config: config,
		addr:   config.Addr + ":" + config.Port,
	}
}

// validateConfig проверяет корректность конфигурации
func (a *App) validateConfig() error {
	if a.config.Addr == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if a.config.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	return nil
}

// Run запускает приложение
func (a *App) Run() error {
	// Валидируем конфигурацию
	if err := a.validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Создаем логгер
	appLogger := logger.NewSlogLogger()
	appLogger.Info("starting metrics server",
		"addr", a.addr,
		"database_configured", a.config.DatabaseDSN != "",
		"store_interval", a.config.StoreInterval,
		"file_storage", a.config.FileStoragePath)

	// Создаем репозиторий в зависимости от конфигурации
	var repo repository.MetricsRepository
	var dbConnection *db.Connection
	var err error

	if a.config.DatabaseDSN != "" {
		// Используем PostgreSQL
		appLogger.Info("initializing PostgreSQL database")

		// Создаем конфигурацию БД
		dbConfig := db.NewConfig()
		dbConfig.DSN = a.config.DatabaseDSN

		// Создаем подключение к БД
		dbConnection, err = db.NewConnection(dbConfig, appLogger)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() {
			if dbConnection != nil {
				dbConnection.Close()
			}
		}()

		// Выполняем миграции
		ctx, cancel := context.WithTimeout(context.Background(), DefaultMigrationTimeout)
		defer cancel()
		if err := db.Migrate(ctx, dbConnection, appLogger); err != nil {
			return fmt.Errorf("failed to run database migrations: %w", err)
		}

		// Создаем PostgreSQL репозиторий
		repo = repository.NewPostgreSQLMetricsRepository(dbConnection.Pool(), appLogger)
		appLogger.Info("using PostgreSQL repository")
	} else {
		// Используем in-memory репозиторий как fallback
		appLogger.Info("using in-memory repository")
		inMemoryRepo := repository.NewInMemoryMetricsRepository(appLogger, a.config.FileStoragePath, a.config.Restore)

		// Устанавливаем синхронное сохранение, если интервал = 0
		if a.config.StoreInterval == 0 {
			inMemoryRepo.SetSyncSave(true)
		}
		repo = inMemoryRepo
	}

	// Создаем сервис
	service := service.NewMetricsService(repo, appLogger)

	// Создаем обработчик
	metricsHandler, err := handler.NewMetricsHandler(service, appLogger)
	if err != nil {
		return fmt.Errorf("failed to create metrics handler: %w", err)
	}

	// Создаем роутер с правильными зависимостями
	var chiRouter *chi.Mux
	if dbConnection != nil {
		// Используем БД connection для ping
		chiRouter = routes.SetupMetricsRoutes(metricsHandler, dbConnection)
	} else {
		// Создаем заглушку для ping (всегда возвращает ошибку)
		chiRouter = routes.SetupMetricsRoutes(metricsHandler, &noDatabasePinger{})
	}

	// Создаем сервер
	server, err := httpserver.NewServerWithChiRouter(a.addr, chiRouter, appLogger)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	a.server = server

	// Запускаем периодическое сохранение метрик только для in-memory репозитория
	if a.config.DatabaseDSN == "" && a.config.StoreInterval > 0 {
		go a.startPeriodicSaving(repo, appLogger)
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем сервер в горутине
	go func() {
		if err := a.server.Start(ctx); err != nil {
			appLogger.Error("server error", "error", err)
			cancel()
		}
	}()

	// Ожидаем сигналы для graceful shutdown
	return a.waitForShutdown(ctx, repo, appLogger)
}

// startPeriodicSaving запускает периодическое сохранение метрик
func (a *App) startPeriodicSaving(repo repository.MetricsRepository, logger logger.Logger) {
	ticker := time.NewTicker(time.Duration(a.config.StoreInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := repo.SaveToFile(); err != nil {
			logger.Error("failed to save metrics to file", "error", err)
		} else {
			logger.Debug("metrics saved to file successfully")
		}
	}
}

// waitForShutdown ожидает сигналы для graceful shutdown
func (a *App) waitForShutdown(ctx context.Context, repo repository.MetricsRepository, logger logger.Logger) error {
	// Создаем канал для сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигнал или ошибку сервера
	select {
	case sig := <-sigChan:
		logger.Info("received signal, shutting down gracefully", "signal", sig)
		// Останавливаем периодическое сохранение перед завершением
		if a.config.StoreInterval > 0 {
			if err := repo.SaveToFile(); err != nil {
				logger.Error("failed to save metrics to file on shutdown", "error", err)
			} else {
				logger.Debug("metrics saved to file on shutdown successfully")
			}
		}
	case <-ctx.Done():
		logger.Info("server stopped, shutting down")
		// Останавливаем периодическое сохранение перед завершением
		if a.config.StoreInterval > 0 {
			if err := repo.SaveToFile(); err != nil {
				logger.Error("failed to save metrics to file on graceful shutdown", "error", err)
			} else {
				logger.Debug("metrics saved to file on graceful shutdown successfully")
			}
		}
	}

	// Даем время на завершение текущих запросов
	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	// Gracefully останавливаем сервер
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		logger.Error("error during shutdown", "error", err)
		return err
	}

	logger.Info("server shutdown complete")
	return nil
}

// GetServer возвращает экземпляр сервера (для тестирования)
func (a *App) GetServer() *httpserver.Server {
	return a.server
}

// GetPort возвращает адрес приложения
func (a *App) GetPort() string {
	return a.addr
}

// noDatabasePinger заглушка для случая когда БД не используется
type noDatabasePinger struct{}

func (p *noDatabasePinger) Ping(ctx context.Context) error {
	return fmt.Errorf("database not configured")
}
