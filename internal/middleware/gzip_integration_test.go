package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGzipMiddlewareIntegration тестирует интеграцию gzip middleware с реальными HTTP запросами
func TestGzipMiddlewareIntegration(t *testing.T) {
	// Создаем тестовый handler, который возвращает JSON
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем Content-Type для JSON
		w.Header().Set("Content-Type", "application/json")

		// Возвращаем JSON данные
		response := map[string]interface{}{
			"status": "success",
			"data":   "test response",
			"count":  42,
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}

		w.Write(jsonData)
	})

	// Создаем middleware
	middleware := GzipMiddleware()
	wrappedHandler := middleware(handler)

	// Тест 1: Клиент поддерживает gzip, получаем сжатый ответ
	t.Run("client supports gzip - compressed response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Проверяем статус
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Проверяем заголовок Content-Encoding
		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding: gzip, got %s", w.Header().Get("Content-Encoding"))
		}

		// Проверяем Content-Type
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", w.Header().Get("Content-Type"))
		}

		// Распаковываем ответ
		body := w.Body.Bytes()
		gzReader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()

		uncompressed, err := io.ReadAll(gzReader)
		if err != nil {
			t.Fatalf("Failed to decompress response: %v", err)
		}

		// Проверяем, что распакованные данные - это валидный JSON
		var response map[string]interface{}
		if err := json.Unmarshal(uncompressed, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Проверяем содержимое ответа
		if response["status"] != "success" {
			t.Errorf("Expected status 'success', got %v", response["status"])
		}
	})

	// Тест 2: Клиент не поддерживает gzip, получаем несжатый ответ
	t.Run("client does not support gzip - uncompressed response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		// Не устанавливаем Accept-Encoding

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Проверяем статус
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Проверяем, что ответ не сжат
		if w.Header().Get("Content-Encoding") != "" {
			t.Errorf("Expected no Content-Encoding, got %s", w.Header().Get("Content-Encoding"))
		}

		// Проверяем Content-Type
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", w.Header().Get("Content-Type"))
		}

		// Проверяем, что тело ответа - это валидный JSON
		body := w.Body.Bytes()
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Проверяем содержимое ответа
		if response["status"] != "success" {
			t.Errorf("Expected status 'success', got %v", response["status"])
		}
	})

	// Тест 3: Отправляем сжатый запрос, получаем ответ
	t.Run("send compressed request - receive response", func(t *testing.T) {
		// Создаем сжатые данные для запроса
		requestData := map[string]interface{}{
			"action": "test",
			"value":  123,
		}

		jsonData, err := json.Marshal(requestData)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		// Сжимаем данные запроса
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		gzWriter.Write(jsonData)
		gzWriter.Close()

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(buf.Bytes()))
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Проверяем статус
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Проверяем, что ответ сжат (клиент поддерживает gzip)
		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding: gzip, got %s", w.Header().Get("Content-Encoding"))
		}

		// Распаковываем ответ
		body := w.Body.Bytes()
		gzReader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()

		uncompressed, err := io.ReadAll(gzReader)
		if err != nil {
			t.Fatalf("Failed to decompress response: %v", err)
		}

		// Проверяем содержимое ответа
		var response map[string]interface{}
		if err := json.Unmarshal(uncompressed, &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["status"] != "success" {
			t.Errorf("Expected status 'success', got %v", response["status"])
		}
	})
}
