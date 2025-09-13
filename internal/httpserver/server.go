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
	"github.com/go-chi/chi/v5"
)

// HTTPServer интерфейс для HTTP сервера
type HTTPServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// RouterFactory интерфейс для создания роутеров
type RouterFactory interface {
	CreateRouter(handler *handler.MetricsHandler, pinger handler.DatabasePinger) *router.Router
}

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

// ServerOptions опции для создания сервера
type ServerOptions struct {
	Config  *ServerConfig
	Handler *handler.MetricsHandler
	Router  *router.Router
	Logger  logger.Logger
}

// validateServerOptions валидирует опции сервера
func validateServerOptions(opts ServerOptions) error {
	if opts.Config == nil {
		return errors.New("config cannot be nil")
	}
	if opts.Config.Addr == "" {
		return errors.New("address cannot be empty")
	}
	if opts.Logger == nil {
		return errors.New("logger cannot be nil")
	}
	// Router может быть nil, если создается через NewServerWithChiRouter
	// Handler может быть nil, если создается через NewServerWithChiRouter
	return nil
}

// defaultRouterFactory реализация RouterFactory по умолчанию
type defaultRouterFactory struct{}

// CreateRouter создает роутер с переданными зависимостями
func (f *defaultRouterFactory) CreateRouter(handler *handler.MetricsHandler, pinger handler.DatabasePinger) *router.Router {
	chiRouter := routes.SetupMetricsRoutes(handler, pinger)
	return router.NewWithChiRouter(chiRouter)
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

	opts := ServerOptions{
		Config:  config,
		Handler: handler,
		Logger:  logger,
	}

	return NewServerWithOptions(opts)
}

// NewServerWithOptions создает новый HTTP сервер с опциями
func NewServerWithOptions(opts ServerOptions) (*Server, error) {
	if err := validateServerOptions(opts); err != nil {
		return nil, err
	}

	opts.Logger.Info("creating server with options", "addr", opts.Config.Addr)

	srv := &Server{
		config:  opts.Config,
		handler: opts.Handler,
		logger:  opts.Logger,
	}

	// Если роутер не передан, создаем его
	if opts.Router == nil {
		if opts.Handler == nil {
			return nil, errors.New("handler is required when router is not provided")
		}

		opts.Logger.Info("creating router")
		factory := &defaultRouterFactory{}
		srv.router = factory.CreateRouter(opts.Handler, &noDatabasePinger{})
		opts.Logger.Info("router created successfully")
	} else {
		srv.router = opts.Router
		opts.Logger.Info("using provided router")
	}

	return srv, nil
}

// NewServerWithConfig создает новый HTTP сервер с конфигурацией (deprecated, используйте NewServerWithOptions)
func NewServerWithConfig(config *ServerConfig, handler *handler.MetricsHandler, logger logger.Logger) (*Server, error) {
	opts := ServerOptions{
		Config:  config,
		Handler: handler,
		Logger:  logger,
	}
	return NewServerWithOptions(opts)
}

// NewServerWithChiRouter создает новый HTTP сервер с готовым chi роутером
func NewServerWithChiRouter(addr string, chiRouter *chi.Mux, logger logger.Logger) (*Server, error) {
	config := DefaultServerConfig()
	config.Addr = addr

	if chiRouter == nil {
		return nil, errors.New("router cannot be nil")
	}

	opts := ServerOptions{
		Config: config,
		Router: router.NewWithChiRouter(chiRouter),
		Logger: logger,
	}

	return NewServerWithOptions(opts)
}

// Start запускает HTTP сервер
func (s *Server) Start(ctx context.Context) error {
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

	// Запускаем сервер в горутине для поддержки контекста
	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Ждем либо завершения сервера, либо отмены контекста
	select {
	case err := <-errChan:
		s.logger.Error("server error", "error", err)
		return fmt.Errorf("failed to start server: %w", err)
	case <-ctx.Done():
		s.logger.Info("server start cancelled", "error", ctx.Err())
		return ctx.Err()
	}
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

// noDatabasePinger заглушка для случая когда БД не используется
type noDatabasePinger struct{}

func (p *noDatabasePinger) Ping(ctx context.Context) error {
	return fmt.Errorf("database not configured")
}
