package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/IgorKilipenko/metrical/internal/httpserver"
	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/service"
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
}

// New создает новое приложение с заданной конфигурацией
func New(config Config) *App {
	return &App{
		config: config,
		addr:   config.Addr + ":" + config.Port,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	log.Printf("Starting metrics server on %s", a.addr)

	// Создаем логгер
	appLogger := logger.NewSlogLogger()

	// Создаем зависимости (Dependency Injection)
	repository := repository.NewInMemoryMetricsRepository(appLogger, a.config.FileStoragePath, a.config.Restore)

	// Устанавливаем синхронное сохранение, если интервал = 0
	if a.config.StoreInterval == 0 {
		repository.SetSyncSave(true)
	}

	service := service.NewMetricsService(repository, appLogger)
	handler := handler.NewMetricsHandler(service, appLogger)

	// Создаем сервер с переданными зависимостями
	server, err := httpserver.NewServer(a.addr, handler, appLogger)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	a.server = server

	// Запускаем периодическое сохранение метрик, если интервал > 0
	if a.config.StoreInterval > 0 {
		go a.startPeriodicSaving(repository, appLogger)
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем сервер в горутине
	go func() {
		if err := a.server.Start(); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Ожидаем сигналы для graceful shutdown
	return a.waitForShutdown(ctx, repository, appLogger)
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
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		// Останавливаем периодическое сохранение перед завершением
		if a.config.StoreInterval > 0 {
			if err := repo.SaveToFile(); err != nil {
				logger.Error("failed to save metrics to file on shutdown", "error", err)
			} else {
				logger.Debug("metrics saved to file on shutdown successfully")
			}
		}
	case <-ctx.Done():
		log.Println("Server stopped, shutting down...")
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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully останавливаем сервер
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
		return err
	}

	log.Println("Server shutdown complete")
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
