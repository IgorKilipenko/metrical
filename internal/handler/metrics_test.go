package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler_UpdateMetric(t *testing.T) {
	// Создаем мок хранилище
	storage := models.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewMetricsHandler(metricsService)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid gauge metric",
			method:         "POST",
			path:           "/update/gauge/temperature/23.5",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Valid counter metric",
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
			name:           "Invalid URL format - too few parts",
			method:         "POST",
			path:           "/update/gauge/temperature",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not found\n",
		},
		{
			name:           "Invalid URL format - too many parts",
			method:         "POST",
			path:           "/update/gauge/temperature/23.5/extra",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not found\n",
		},
		{
			name:           "Empty metric name",
			method:         "POST",
			path:           "/update/gauge//123.45",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not found\n",
		},
		{
			name:           "Invalid metric type",
			method:         "POST",
			path:           "/update/invalid/name/100",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unknown metric type: invalid\n",
		},
		{
			name:           "Invalid gauge value",
			method:         "POST",
			path:           "/update/gauge/name/invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid gauge value: invalid\n",
		},
		{
			name:           "Invalid counter value",
			method:         "POST",
			path:           "/update/counter/name/invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid counter value: invalid\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.UpdateMetric(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
			assert.Equal(t, tt.expectedBody, w.Body.String(), "Response body mismatch")
		})
	}
}

func TestMetricsHandler_UpdateMetric_CounterAccumulation(t *testing.T) {
	storage := models.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewMetricsHandler(metricsService)

	// Отправляем первое значение
	req1 := httptest.NewRequest("POST", "/update/counter/requests/100", nil)
	w1 := httptest.NewRecorder()
	handler.UpdateMetric(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code, "First request should succeed")

	// Отправляем второе значение
	req2 := httptest.NewRequest("POST", "/update/counter/requests/50", nil)
	w2 := httptest.NewRecorder()
	handler.UpdateMetric(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code, "Second request should succeed")

	// Проверяем, что значения накопились
	value, exists := storage.GetCounter("requests")
	assert.True(t, exists, "Counter 'requests' should exist")
	assert.Equal(t, int64(150), value, "Counter value should be accumulated")
}

func TestMetricsHandler_UpdateMetric_GaugeReplacement(t *testing.T) {
	storage := models.NewMemStorage()
	metricsService := service.NewMetricsService(storage)
	handler := NewMetricsHandler(metricsService)

	// Отправляем первое значение
	req1 := httptest.NewRequest("POST", "/update/gauge/temperature/23.5", nil)
	w1 := httptest.NewRecorder()
	handler.UpdateMetric(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code, "First request should succeed")

	// Отправляем второе значение
	req2 := httptest.NewRequest("POST", "/update/gauge/temperature/25.0", nil)
	w2 := httptest.NewRecorder()
	handler.UpdateMetric(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code, "Second request should succeed")

	// Проверяем, что значение заменилось
	value, exists := storage.GetGauge("temperature")
	assert.True(t, exists, "Gauge 'temperature' should exist")
	assert.Equal(t, 25.0, value, "Gauge value should be replaced")
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expected    []string
		expectError bool
	}{
		{
			name:        "Normal path with 4 parts",
			path:        "/update/gauge/temperature/23.5",
			expected:    []string{"update", "gauge", "temperature", "23.5"},
			expectError: false,
		},
		{
			name:        "Path with empty segment",
			path:        "/update/gauge//123.45",
			expected:    []string{"update", "gauge", "", "123.45"},
			expectError: false,
		},
		{
			name:        "Path without leading slash",
			path:        "update/gauge/test/123",
			expected:    []string{"update", "gauge", "test", "123"},
			expectError: false,
		},
		{
			name:        "Path with trailing slash",
			path:        "/update/gauge/test/123/",
			expected:    []string{"update", "gauge", "test", "123"},
			expectError: false,
		},
		{
			name:        "Path with special characters",
			path:        "/update/gauge/test-metric_123/42.0",
			expected:    []string{"update", "gauge", "test-metric_123", "42.0"},
			expectError: false,
		},
		{
			name:        "Path with numbers and dots",
			path:        "/update/counter/requests/100",
			expected:    []string{"update", "counter", "requests", "100"},
			expectError: false,
		},
		{
			name:        "Empty path",
			path:        "",
			expected:    []string{""},
			expectError: false,
		},
		{
			name:        "Single slash",
			path:        "/",
			expected:    []string{""},
			expectError: false,
		},
		{
			name:        "Double slash",
			path:        "//",
			expected:    []string{""},
			expectError: false,
		},
		{
			name:        "Path with consecutive slashes",
			path:        "/update///test/123",
			expected:    []string{"update", "", "", "test", "123"},
			expectError: false,
		},
		{
			name:        "Path with unicode characters",
			path:        "/update/gauge/тест/123",
			expected:    []string{"update", "gauge", "тест", "123"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := splitPath(tt.path)

			if tt.expectError {
				assert.Error(t, err, "Expected error for unsupported characters")
				assert.Nil(t, result, "Result should be nil on error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tt.expected, result, "Path splitting should match expected result")
			}
		})
	}
}
