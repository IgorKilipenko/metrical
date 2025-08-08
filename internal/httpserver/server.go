package httpserver

import (
	"log"
	"net/http"

	"github.com/IgorKilipenko/metrical/internal/handler"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/router"
	"github.com/IgorKilipenko/metrical/internal/service"
)

// Server представляет HTTP сервер
type Server struct {
	addr    string
	handler *handler.MetricsHandler
}

// NewServer создает новый HTTP сервер
func NewServer(addr string) *Server {
	storage := models.NewMemStorage()
	service := service.NewMetricsService(storage)
	handler := handler.NewMetricsHandler(service)

	return &Server{
		addr:    addr,
		handler: handler,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	r := router.New()
	chiRouter := r.GetRouter()

	// Настраиваем маршруты с помощью chi
	chiRouter.Get("/", s.handler.GetAllMetrics)
	chiRouter.Post("/update/{type}/{name}/{value}", s.handler.UpdateMetric)
	chiRouter.Get("/value/{type}/{name}", s.handler.GetMetricValue)

	log.Printf("Starting server on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

// GetMux оставлен для обратной совместимости
func (s *Server) GetMux() *http.ServeMux {
	// Возвращаем nil, так как теперь используем chi
	return nil
}

// ServeHTTP реализует интерфейс http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := s.getRouter()
	router.ServeHTTP(w, r)
}

// getRouter создает роутер с настроенными маршрутами
func (s *Server) getRouter() *router.Router {
	r := router.New()
	chiRouter := r.GetRouter()

	// Настраиваем маршруты с помощью chi
	chiRouter.Get("/", s.handler.GetAllMetrics)
	chiRouter.Post("/update/{type}/{name}/{value}", s.handler.UpdateMetric)
	chiRouter.Get("/value/{type}/{name}", s.handler.GetMetricValue)

	return r
}
