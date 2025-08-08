package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/IgorKilipenko/metrical/internal/handler"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/router"
	"github.com/IgorKilipenko/metrical/internal/routes"
	"github.com/IgorKilipenko/metrical/internal/service"
)

// Server представляет HTTP сервер
type Server struct {
	addr    string
	handler *handler.MetricsHandler
	router  *router.Router // Кэшированный роутер
	server  *http.Server   // Ссылка на HTTP сервер для graceful shutdown
}

// NewServer создает новый HTTP сервер
func NewServer(addr string) (*Server, error) {
	if addr == "" {
		return nil, errors.New("address cannot be empty")
	}

	storage := models.NewMemStorage()
	repo := repository.NewInMemoryMetricsRepository(storage)
	service := service.NewMetricsService(repo)
	handler := handler.NewMetricsHandler(service)

	srv := &Server{
		addr:    addr,
		handler: handler,
	}

	// Инициализируем роутер один раз
	srv.router = srv.createRouter()

	return srv, nil
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.addr)
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.router,
	}

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Server error: %v", err)
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// Shutdown gracefully останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server gracefully...")
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
