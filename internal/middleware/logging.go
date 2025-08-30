package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем обертку для ResponseWriter
			wrappedWriter := &ResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // По умолчанию 200
			}

			// Логируем информацию о запросе
			log.Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Msg("HTTP request started")

			// Выполняем следующий обработчик
			next.ServeHTTP(wrappedWriter, r)

			// Вычисляем время выполнения
			duration := time.Since(start)

			// Логируем информацию об ответе
			log.Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Int("status_code", wrappedWriter.statusCode).
				Int("response_size", wrappedWriter.size).
				Dur("duration", duration).
				Msg("HTTP request completed")
		})
	}
}

// LoggingMiddlewareWithLogger создает middleware с кастомным логгером
func LoggingMiddlewareWithLogger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем обертку для ResponseWriter
			wrappedWriter := &ResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // По умолчанию 200
			}

			// Логируем информацию о запросе
			logger.Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Msg("HTTP request started")

			// Выполняем следующий обработчик
			next.ServeHTTP(wrappedWriter, r)

			// Вычисляем время выполнения
			duration := time.Since(start)

			// Логируем информацию об ответе
			logger.Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Int("status_code", wrappedWriter.statusCode).
				Int("response_size", wrappedWriter.size).
				Dur("duration", duration).
				Msg("HTTP request completed")
		})
	}
}
