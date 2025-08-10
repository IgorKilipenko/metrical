package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/IgorKilipenko/metrical/internal/router"
	"github.com/IgorKilipenko/metrical/internal/routes"
)

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DefaultServerConfig возвращает конфигурацию по умолчанию
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// Server представляет HTTP сервер
type Server struct {
	config  *ServerConfig
	handler *handler.MetricsHandler
	router  *router.Router // Кэшированный роутер
	server  *http.Server   // Ссылка на HTTP сервер для graceful shutdown
}

// NewServer создает новый HTTP сервер с переданными зависимостями
func NewServer(addr string, handler *handler.MetricsHandler) (*Server, error) {
	return NewServerWithConfig(&ServerConfig{Addr: addr}, handler)
}

// NewServerWithConfig создает новый HTTP сервер с конфигурацией
func NewServerWithConfig(config *ServerConfig, handler *handler.MetricsHandler) (*Server, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if config.Addr == "" {
		return nil, errors.New("address cannot be empty")
	}
	if handler == nil {
		return nil, errors.New("handler cannot be nil")
	}

	srv := &Server{
		config:  config,
		handler: handler,
	}

	// Инициализируем роутер один раз
	srv.router = srv.createRouter()

	return srv, nil
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	slog.Info("starting HTTP server", "addr", s.config.Addr)
	s.server = &http.Server{
		Addr:         s.config.Addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Shutdown gracefully останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("shutting down server gracefully")
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// ServeHTTP реализует интерфейс http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// createRouter создает и настраивает роутер с маршрутами
func (s *Server) createRouter() *router.Router {
	// Используем отдельный пакет для настройки маршрутов
	chiRouter := routes.SetupMetricsRoutes(s.handler)
	return router.NewWithChiRouter(chiRouter)
}
