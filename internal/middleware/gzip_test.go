package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding: gzip, got %s", w.Header().Get("Content-Encoding"))
		}

		// Проверяем, что тело ответа действительно сжато
		body := w.Body.Bytes()
		if len(body) == 0 {
			t.Error("Response body is empty")
		}

		// Пытаемся распаковать ответ
		gzReader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			t.Errorf("Failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()

		uncompressed, err := io.ReadAll(gzReader)
		if err != nil {
			t.Errorf("Failed to decompress response: %v", err)
		}

		expected := `{"message": "test response"}`
		if string(uncompressed) != expected {
			t.Errorf("Expected %s, got %s", expected, string(uncompressed))
		}
	})

	// Тест 2: Клиент не поддерживает gzip
	t.Run("client does not support gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		// Не устанавливаем Accept-Encoding

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Проверяем, что ответ не сжат
		if w.Header().Get("Content-Encoding") != "" {
			t.Errorf("Expected no Content-Encoding, got %s", w.Header().Get("Content-Encoding"))
		}

		// Проверяем, что тело ответа не сжато
		body := w.Body.String()
		expected := `{"message": "test response"}`
		if body != expected {
			t.Errorf("Expected %s, got %s", expected, body)
		}
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
		gzWriter.Write([]byte("compressed test data"))
		gzWriter.Close()

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(buf.Bytes()))
		req.Header.Set("Content-Encoding", "gzip")

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Проверяем, что сервер корректно обработал сжатый запрос
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		expected := "compressed test data"
		if body != expected {
			t.Errorf("Expected %s, got %s", expected, body)
		}
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
		if w.Header().Get("Content-Encoding") != "" {
			t.Errorf("Expected no Content-Encoding for binary content, got %s", w.Header().Get("Content-Encoding"))
		}
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
		result := isCompressibleContentType(test.contentType)
		if result != test.expected {
			t.Errorf("For content type '%s': expected %v, got %v", test.contentType, test.expected, result)
		}
	}
}
