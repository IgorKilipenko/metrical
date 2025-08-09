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
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/service"
)

// App представляет основное приложение
type App struct {
	server *httpserver.Server
	port   string
}

// Config содержит конфигурацию приложения
type Config struct {
	Port string
}

// New создает новое приложение с заданной конфигурацией
func New(config Config) *App {
	return &App{
		port: config.Port,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	log.Printf("Starting metrics server on port %s", a.port)

	// Создаем зависимости (Dependency Injection)
	repository := repository.NewInMemoryMetricsRepository()
	service := service.NewMetricsService(repository)
	handler := handler.NewMetricsHandler(service)

	// Создаем сервер с переданными зависимостями
	server, err := httpserver.NewServer(":"+a.port, handler)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	a.server = server

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
	return a.waitForShutdown(ctx)
}

// waitForShutdown ожидает сигналы для graceful shutdown
func (a *App) waitForShutdown(ctx context.Context) error {
	// Создаем канал для сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигнал или ошибку сервера
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
	case <-ctx.Done():
		log.Println("Server stopped, shutting down...")
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

// GetPort возвращает порт приложения
func (a *App) GetPort() string {
	return a.port
}
