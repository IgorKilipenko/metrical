package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Router представляет HTTP роутер
type Router struct {
	router *chi.Mux
}

// New создает новый роутер
func New() *Router {
	return &Router{
		router: chi.NewRouter(),
	}
}

// HandleFunc регистрирует обработчик для указанного пути
func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.router.HandleFunc(pattern, handler)
}

// Handle регистрирует обработчик для указанного пути
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.router.Handle(pattern, handler)
}

// ServeHTTP реализует интерфейс http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// GetRouter возвращает chi роутер для более гибкой настройки маршрутов
func (r *Router) GetRouter() *chi.Mux {
	return r.router
}

// GetMux оставлен для обратной совместимости
func (r *Router) GetMux() *http.ServeMux {
	// Возвращаем nil, так как теперь используем chi
	return nil
}
