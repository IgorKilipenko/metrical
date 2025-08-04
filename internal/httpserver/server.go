package httpserver

import (
	"log"
	"net/http"

	"github.com/IgorKilipenko/metrical/internal/handler"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/router"
	"github.com/IgorKilipenko/metrical/internal/service"
)

// Server представляет HTTP сервер для метрик
type Server struct {
	addr    string
	handler *handler.MetricsHandler
}

// NewServer создает новый экземпляр сервера
func NewServer(addr string) *Server {
	// Создаем хранилище метрик
	storage := models.NewMemStorage()

	// Создаем сервис для работы с метриками
	metricsService := service.NewMetricsService(storage)

	// Создаем HTTP обработчик
	metricsHandler := handler.NewMetricsHandler(metricsService)

	return &Server{
		addr:    addr,
		handler: metricsHandler,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	// Настраиваем маршруты
	r := router.New()
	r.HandleFunc("/update/", s.handler.UpdateMetric)

	// Запускаем сервер
	log.Printf("Starting server on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

// GetMux возвращает настроенный ServeMux для использования в тестах
func (s *Server) GetMux() *http.ServeMux {
	r := router.New()
	r.HandleFunc("/update/", s.handler.UpdateMetric)
	return r.GetMux()
}

// ServeHTTP реализует интерфейс http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := s.getRouter()
	router.ServeHTTP(w, r)
}

// getRouter возвращает настроенный роутер
func (s *Server) getRouter() *router.Router {
	r := router.New()
	r.HandleFunc("/update/", s.handler.UpdateMetric)
	return r
}
