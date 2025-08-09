package router

import (
	"net/http"
)

// Router представляет HTTP роутер
type Router struct {
	mux *http.ServeMux
}

// New создает новый роутер
func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// HandleFunc регистрирует обработчик для указанного пути
func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.mux.HandleFunc(pattern, handler)
}

// Handle регистрирует обработчик для указанного пути
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

// ServeHTTP реализует интерфейс http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// GetMux возвращает внутренний ServeMux (для совместимости)
func (r *Router) GetMux() *http.ServeMux {
	return r.mux
}
