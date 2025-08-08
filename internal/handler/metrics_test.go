package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/go-chi/chi/v5"
)

// createTestHandler создает тестовый обработчик
func createTestHandler() *MetricsHandler {
	storage := models.NewMemStorage()
	repo := repository.NewInMemoryMetricsRepository(storage)
	service := service.NewMetricsService(repo)
	return NewMetricsHandler(service)
}

// createChiContext создает контекст chi для тестирования
func createChiContext(pattern string, params map[string]string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Создаем chi контекст
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req, w
}

func TestMetricsHandler_UpdateMetric(t *testing.T) {
	handler := createTestHandler()

	tests := []struct {
		name           string
		method         string
		params         map[string]string
		expectedStatus int
	}{
		{
			name:   "Valid gauge metric",
			method: "POST",
			params: map[string]string{
				"type":  "gauge",
				"name":  "temperature",
				"value": "23.5",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Valid counter metric",
			method: "POST",
			params: map[string]string{
				"type":  "counter",
				"name":  "requests",
				"value": "100",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Invalid HTTP method",
			method: "GET",
			params: map[string]string{
				"type":  "gauge",
				"name":  "temperature",
				"value": "23.5",
			},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Empty metric name",
			method: "POST",
			params: map[string]string{
				"type":  "gauge",
				"name":  "",
				"value": "23.5",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Invalid metric type",
			method: "POST",
			params: map[string]string{
				"type":  "invalid",
				"name":  "test",
				"value": "123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid gauge value",
			method: "POST",
			params: map[string]string{
				"type":  "gauge",
				"name":  "test",
				"value": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid counter value",
			method: "POST",
			params: map[string]string{
				"type":  "counter",
				"name":  "test",
				"value": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, w := createChiContext("/update/{type}/{name}/{value}", tt.params)
			req.Method = tt.method

			handler.UpdateMetric(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMetricsHandler_GetMetricValue(t *testing.T) {
	handler := createTestHandler()

	// Добавляем тестовые метрики
	handler.service.UpdateMetric("gauge", "temperature", "23.5")
	handler.service.UpdateMetric("counter", "requests", "100")

	tests := []struct {
		name           string
		method         string
		params         map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Valid gauge metric",
			method: "GET",
			params: map[string]string{
				"type": "gauge",
				"name": "temperature",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "23.5",
		},
		{
			name:   "Valid counter metric",
			method: "GET",
			params: map[string]string{
				"type": "counter",
				"name": "requests",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "100",
		},
		{
			name:   "Invalid HTTP method",
			method: "POST",
			params: map[string]string{
				"type": "gauge",
				"name": "temperature",
			},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Empty metric name",
			method: "GET",
			params: map[string]string{
				"type": "gauge",
				"name": "",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Invalid metric type",
			method: "GET",
			params: map[string]string{
				"type": "invalid",
				"name": "test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Metric not found",
			method: "GET",
			params: map[string]string{
				"type": "gauge",
				"name": "nonexistent",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, w := createChiContext("/value/{type}/{name}", tt.params)
			req.Method = tt.method

			handler.GetMetricValue(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, body)
				}
			}
		})
	}
}

func TestMetricsHandler_GetAllMetrics(t *testing.T) {
	handler := createTestHandler()

	// Добавляем тестовые метрики
	handler.service.UpdateMetric("gauge", "temperature", "23.5")
	handler.service.UpdateMetric("counter", "requests", "100")

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:           "Valid GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"temperature", "23.5", "requests", "100"},
		},
		{
			name:           "Invalid HTTP method",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			handler.GetAllMetrics(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				body := w.Body.String()
				for _, expected := range tt.expectedBody {
					if !strings.Contains(body, expected) {
						t.Errorf("Expected body to contain '%s', got '%s'", expected, body)
					}
				}

				// Проверяем Content-Type
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, "text/html") {
					t.Errorf("Expected Content-Type to contain 'text/html', got '%s'", contentType)
				}
			}
		})
	}
}

func TestMetricsHandler_UpdateMetric_CounterAccumulation(t *testing.T) {
	handler := createTestHandler()

	// Добавляем counter метрику несколько раз
	handler.service.UpdateMetric("counter", "requests", "10")
	handler.service.UpdateMetric("counter", "requests", "20")
	handler.service.UpdateMetric("counter", "requests", "30")

	// Проверяем, что значения накапливаются
	req, w := createChiContext("/value/{type}/{name}", map[string]string{
		"type": "counter",
		"name": "requests",
	})
	req.Method = "GET"

	handler.GetMetricValue(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expectedValue := "60" // 10 + 20 + 30
	if strings.TrimSpace(w.Body.String()) != expectedValue {
		t.Errorf("Expected value %s, got %s", expectedValue, w.Body.String())
	}
}

func TestMetricsHandler_UpdateMetric_GaugeReplacement(t *testing.T) {
	handler := createTestHandler()

	// Добавляем gauge метрику несколько раз
	handler.service.UpdateMetric("gauge", "temperature", "20.0")
	handler.service.UpdateMetric("gauge", "temperature", "25.5")
	handler.service.UpdateMetric("gauge", "temperature", "30.0")

	// Проверяем, что последнее значение заменяет предыдущие
	req, w := createChiContext("/value/{type}/{name}", map[string]string{
		"type": "gauge",
		"name": "temperature",
	})
	req.Method = "GET"

	handler.GetMetricValue(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expectedValue := "30" // Последнее значение
	if strings.TrimSpace(w.Body.String()) != expectedValue {
		t.Errorf("Expected value %s, got %s", expectedValue, w.Body.String())
	}
}
