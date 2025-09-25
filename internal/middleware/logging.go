package middleware

import (
	"net/http"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
)

// ResponseWriter обертка для http.ResponseWriter для отслеживания статуса и размера ответа
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader перехватывает статус код ответа
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write перехватывает размер ответа
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// LoggingMiddleware создает middleware для логирования HTTP запросов и ответов
func LoggingMiddleware() func(http.Handler) http.Handler {
	logger := logger.NewSlogLogger()
	return LoggingMiddlewareWithLogger(logger)
}

// LoggingMiddlewareWithLogger создает middleware с кастомным логгером
func LoggingMiddlewareWithLogger(logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем обертку для ResponseWriter
			wrappedWriter := &ResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // По умолчанию 200
			}

			// Логируем информацию о запросе
			logger.Info("HTTP request started",
				"method", r.Method,
				"uri", r.RequestURI,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Выполняем следующий обработчик
			next.ServeHTTP(wrappedWriter, r)

			// Вычисляем время выполнения
			duration := time.Since(start)

			// Логируем информацию об ответе
			logger.Info("HTTP request completed",
				"method", r.Method,
				"uri", r.RequestURI,
				"status_code", wrappedWriter.statusCode,
				"response_size", wrappedWriter.size,
				"duration", duration,
			)
		})
	}
}
