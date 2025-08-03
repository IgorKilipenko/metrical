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
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL format\n",
		},
		{
			name:           "Invalid URL format - too many parts",
			method:         "POST",
			path:           "/update/gauge/temperature/23.5/extra",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL format\n",
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
