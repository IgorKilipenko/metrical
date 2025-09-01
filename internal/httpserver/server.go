package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/IgorKilipenko/metrical/internal/logger"
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
	logger  logger.Logger
}

// NewServer создает новый HTTP сервер с переданными зависимостями
func NewServer(addr string, handler *handler.MetricsHandler, logger logger.Logger) (*Server, error) {
	config := DefaultServerConfig()
	config.Addr = addr
	return NewServerWithConfig(config, handler, logger)
}

// NewServerWithConfig создает новый HTTP сервер с конфигурацией
func NewServerWithConfig(config *ServerConfig, handler *handler.MetricsHandler, logger logger.Logger) (*Server, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if config.Addr == "" {
		return nil, errors.New("address cannot be empty")
	}
	if handler == nil {
		return nil, errors.New("handler cannot be nil")
	}
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	logger.Info("creating server with config", "addr", config.Addr)

	srv := &Server{
		config:  config,
		handler: handler,
		logger:  logger,
	}

	// Инициализируем роутер один раз
	logger.Info("creating router")
	srv.router = srv.createRouter()
	logger.Info("router created successfully")

	return srv, nil
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	s.logger.Info("starting HTTP server",
		"addr", s.config.Addr,
		"read_timeout", s.config.ReadTimeout,
		"write_timeout", s.config.WriteTimeout,
		"idle_timeout", s.config.IdleTimeout)

	s.server = &http.Server{
		Addr:         s.config.Addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("server error", "error", err)
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

// Shutdown gracefully останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server gracefully")
	if s.server != nil {
		err := s.server.Shutdown(ctx)
		if err != nil {
			s.logger.Error("error during server shutdown", "error", err)
			return err
		}
		s.logger.Info("server shutdown completed successfully")
		return nil
	}
	s.logger.Warn("shutdown called on nil server")
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
