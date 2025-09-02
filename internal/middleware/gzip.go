package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// GzipMiddleware middleware для поддержки gzip сжатия и распаковки
func GzipMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Обрабатываем входящие сжатые запросы
			if r.Header.Get("Content-Encoding") == "gzip" {
				// Создаем gzip reader для распаковки тела запроса
				gzReader, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Failed to read gzip content", http.StatusBadRequest)
					return
				}
				defer gzReader.Close()

				// Читаем распакованное содержимое
				body, err := io.ReadAll(gzReader)
				if err != nil {
					http.Error(w, "Failed to decompress gzip content", http.StatusBadRequest)
					return
				}

				// Заменяем тело запроса на распакованное содержимое
				r.Body = io.NopCloser(bytes.NewBuffer(body))
				r.ContentLength = int64(len(body))
			}

			// Проверяем, поддерживает ли клиент gzip
			acceptEncoding := r.Header.Get("Accept-Encoding")
			canGzip := strings.Contains(acceptEncoding, "gzip")

			// Если клиент поддерживает gzip, оборачиваем response writer
			if canGzip {
				gzipWriter := &gzipResponseWriter{
					ResponseWriter: w,
					gzipWriter:     gzip.NewWriter(w),
				}
				defer gzipWriter.Close()
				w = gzipWriter
			}

			next.ServeHTTP(w, r)
		})
	}
}

// gzipResponseWriter оборачивает http.ResponseWriter для сжатия ответа
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
	statusCode int
	headers    http.Header
}

// Write записывает данные через gzip writer
func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	// Если статус еще не установлен, устанавливаем по умолчанию
	if g.statusCode == 0 {
		g.WriteHeader(http.StatusOK)
	}
	return g.gzipWriter.Write(data)
}

// WriteHeader устанавливает заголовки ответа
func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	g.statusCode = statusCode

	// Устанавливаем заголовок Content-Encoding только для поддерживаемых типов контента
	contentType := g.Header().Get("Content-Type")
	if isCompressibleContentType(contentType) {
		g.Header().Set("Content-Encoding", "gzip")
	}

	g.ResponseWriter.WriteHeader(statusCode)
}

// Header возвращает заголовки ответа
func (g *gzipResponseWriter) Header() http.Header {
	return g.ResponseWriter.Header()
}

// Close закрывает gzip writer
func (g *gzipResponseWriter) Close() error {
	return g.gzipWriter.Close()
}

// isCompressibleContentType проверяет, можно ли сжимать данный тип контента
func isCompressibleContentType(contentType string) bool {
	// Поддерживаем сжатие для JSON и HTML
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/plain")
}
