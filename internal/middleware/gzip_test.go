package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware_Compression(t *testing.T) {
	// Создаем тестовый handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "test response"}`))
	})

	// Создаем middleware
	middleware := GzipMiddleware()
	wrappedHandler := middleware(handler)

	// Тест 1: Клиент поддерживает gzip
	t.Run("client supports gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Проверяем, что ответ сжат
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"), "Content-Encoding should be gzip")

		// Проверяем, что тело ответа действительно сжато
		body := w.Body.Bytes()
		assert.NotEmpty(t, body, "Response body should not be empty")

		// Пытаемся распаковать ответ
		gzReader, err := gzip.NewReader(bytes.NewReader(body))
		require.NoError(t, err, "Failed to create gzip reader")
		defer gzReader.Close()

		uncompressed, err := io.ReadAll(gzReader)
		require.NoError(t, err, "Failed to decompress response")

		expected := `{"message": "test response"}`
		assert.Equal(t, expected, string(uncompressed), "Decompressed content should match expected")
	})

	// Тест 2: Клиент не поддерживает gzip
	t.Run("client does not support gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		// Не устанавливаем Accept-Encoding

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Проверяем, что ответ не сжат
		assert.Empty(t, w.Header().Get("Content-Encoding"), "Content-Encoding should be empty when client doesn't support gzip")

		// Проверяем, что тело ответа не сжато
		body := w.Body.String()
		expected := `{"message": "test response"}`
		assert.Equal(t, expected, body, "Response body should match expected when not compressed")
	})
}

func TestGzipMiddleware_Decompression(t *testing.T) {
	// Создаем тестовый handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Читаем тело запроса и возвращаем его обратно
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write(body)
	})

	// Создаем middleware
	middleware := GzipMiddleware()
	wrappedHandler := middleware(handler)

	// Тест: Отправляем сжатый запрос
	t.Run("decompress gzipped request", func(t *testing.T) {
		// Создаем сжатые данные
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		_, err := gzWriter.Write([]byte("compressed test data"))
		require.NoError(t, err, "Failed to write compressed data")
		require.NoError(t, gzWriter.Close(), "Failed to close gzip writer")

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(buf.Bytes()))
		req.Header.Set("Content-Encoding", "gzip")

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Проверяем, что сервер корректно обработал сжатый запрос
		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200")

		body := w.Body.String()
		expected := "compressed test data"
		assert.Equal(t, expected, body, "Response body should match expected decompressed data")
	})
}

func TestGzipMiddleware_ContentTypeFiltering(t *testing.T) {
	// Создаем тестовый handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("binary data"))
	})

	// Создаем middleware
	middleware := GzipMiddleware()
	wrappedHandler := middleware(handler)

	// Тест: Не сжимаем бинарные типы контента
	t.Run("do not compress binary content", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Проверяем, что ответ не сжат для бинарного типа
		assert.Empty(t, w.Header().Get("Content-Encoding"), "Content-Encoding should be empty for binary content")
	})
}

func TestIsCompressibleContentType(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"application/json", true},
		{"text/html", true},
		{"text/plain", true},
		{"image/png", false},
		{"application/octet-stream", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(test.contentType, func(t *testing.T) {
			result := isCompressibleContentType(test.contentType)
			assert.Equal(t, test.expected, result, "Content type compressibility should match expected")
		})
	}
}
