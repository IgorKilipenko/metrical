package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IgorKilipenko/metrical/internal/handler"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/stretchr/testify/assert"
)

// createTestServer создает тестовый HTTP сервер с настроенными обработчиками
func createTestServer() *httptest.Server {
	// Создаем хранилище метрик
	storage := models.NewMemStorage()

	// Создаем сервис для работы с метриками
	metricsService := service.NewMetricsService(storage)

	// Создаем HTTP обработчик
	metricsHandler := handler.NewMetricsHandler(metricsService)

	// Настраиваем маршруты
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", metricsHandler.UpdateMetric)

	// Создаем тестовый сервер
	return httptest.NewServer(mux)
}

// TestServerIntegration тестирует интеграцию HTTP сервера
func TestServerIntegration(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Update gauge metric via HTTP",
			method:         "POST",
			path:           "/update/gauge/temperature/23.5",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Update counter metric via HTTP",
			method:         "POST",
			path:           "/update/counter/requests/100",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Invalid HTTP method",
			method:         "GET",
			path:           "/update/gauge/temperature/23.5",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "Invalid URL format",
			method:         "POST",
			path:           "/update/gauge/temperature",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL format\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем HTTP запрос к тестовому серверу
			req, err := http.NewRequest(tt.method, server.URL+tt.path, nil)
			assert.NoError(t, err)

			// Выполняем запрос
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Status code mismatch")

			// Проверяем тело ответа
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, string(body), "Response body mismatch")
		})
	}
}

// TestServerEndToEnd тестирует полный end-to-end цикл через HTTP
func TestServerEndToEnd(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	// Тестируем обновление gauge метрики через HTTP
	t.Run("Gauge metric end-to-end", func(t *testing.T) {
		req, err := http.NewRequest("POST", server.URL+"/update/gauge/memory_usage/85.7", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Тестируем обновление counter метрики через HTTP
	t.Run("Counter metric end-to-end", func(t *testing.T) {
		req, err := http.NewRequest("POST", server.URL+"/update/counter/request_count/1", nil)
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestServerBasicFunctionality тестирует базовую функциональность HTTP сервера
func TestServerBasicFunctionality(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	// Тестируем простой запрос
	req, err := http.NewRequest("POST", server.URL+"/update/gauge/test/123.45", nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestServerRedirects тестирует автоматические редиректы Go HTTP сервера
// Go HTTP сервер автоматически выполняет редиректы для путей с двойными слешами
func TestServerRedirects(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	// Создаем клиент, который НЕ следует редиректам
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	tests := []struct {
		name             string
		path             string
		expectedStatus   int
		expectedRedirect bool
	}{
		{
			name:             "Path with double slash",
			path:             "/update/gauge//123.45",
			expectedStatus:   http.StatusMovedPermanently, // Go HTTP server automatically redirects double slashes
			expectedRedirect: true,
		},
		{
			name:             "Path with trailing slash",
			path:             "/update/gauge/test/123.45/",
			expectedStatus:   http.StatusOK, // Обрабатывается напрямую
			expectedRedirect: false,
		},
		{
			name:             "Normal path",
			path:             "/update/gauge/test/123.45",
			expectedStatus:   http.StatusOK,
			expectedRedirect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", server.URL+tt.path, nil)
			assert.NoError(t, err)

			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Status code mismatch for %s", tt.path)

			if tt.expectedRedirect {
				location := resp.Header.Get("Location")
				assert.NotEmpty(t, location, "Redirect should have Location header")
				t.Logf("Redirect from %s to %s", tt.path, location)
			}
		})
	}
}
