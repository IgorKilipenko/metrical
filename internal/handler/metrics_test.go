package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/IgorKilipenko/metrical/internal/validation"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// createTestHandler создает тестовый handler
func createTestHandler() *MetricsHandler {
	mockLogger := testutils.NewMockLogger()
	repository := repository.NewInMemoryMetricsRepository(mockLogger, testutils.TestMetricsFile, false)
	service := service.NewMetricsService(repository, mockLogger)
	handler, err := NewMetricsHandler(service, mockLogger)
	if err != nil {
		panic(fmt.Sprintf("failed to create test handler: %v", err))
	}
	return handler
}

// createChiContext создает chi контекст для тестирования
func createChiContext(path string, params map[string]string) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()

	// Создаем chi контекст
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	return r, w
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
			name:   "Empty metric name",
			method: "POST",
			params: map[string]string{
				"type":  "gauge",
				"name":  "",
				"value": "23.5",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid gauge value",
			method: "POST",
			params: map[string]string{
				"type":  "gauge",
				"name":  "test",
				"value": "abc",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid counter value",
			method: "POST",
			params: map[string]string{
				"type":  "counter",
				"name":  "test",
				"value": "123.45",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, w := createChiContext("/update/{type}/{name}/{value}", tt.params)
			req.Method = tt.method

			handler.UpdateMetric(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestMetricsHandler_GetMetricValue(t *testing.T) {
	handler := createTestHandler()
	ctx := context.Background()

	// Добавляем тестовые метрики через валидацию
	metricReq1, _ := validation.ValidateMetricRequest("gauge", "temperature", "23.5")
	handler.service.UpdateMetric(ctx, metricReq1)

	metricReq2, _ := validation.ValidateMetricRequest("counter", "requests", "100")
	handler.service.UpdateMetric(ctx, metricReq2)

	tests := []struct {
		name           string
		method         string
		params         map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Get gauge metric",
			method: "GET",
			params: map[string]string{
				"type": "gauge",
				"name": "temperature",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "23.5",
		},
		{
			name:   "Get counter metric",
			method: "GET",
			params: map[string]string{
				"type": "counter",
				"name": "requests",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "100",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, w := createChiContext("/value/{type}/{name}", tt.params)
			req.Method = tt.method

			handler.GetMetricValue(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestMetricsHandler_GetAllMetrics(t *testing.T) {
	handler := createTestHandler()
	ctx := context.Background()

	// Добавляем тестовые метрики через валидацию
	metricReq1, _ := validation.ValidateMetricRequest("gauge", "temperature", "23.5")
	handler.service.UpdateMetric(ctx, metricReq1)

	metricReq2, _ := validation.ValidateMetricRequest("counter", "requests", "100")
	handler.service.UpdateMetric(ctx, metricReq2)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "Get all metrics",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, w := createChiContext("/", nil)
			req.Method = tt.method

			handler.GetAllMetrics(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
			assert.Contains(t, w.Body.String(), "Metrics Dashboard")
		})
	}
}

func TestMetricsHandler_UpdateMetric_CounterAccumulation(t *testing.T) {
	handler := createTestHandler()
	ctx := context.Background()

	// Добавляем counter метрику несколько раз через валидацию
	metricReq1, _ := validation.ValidateMetricRequest("counter", "requests", "10")
	handler.service.UpdateMetric(ctx, metricReq1)

	metricReq2, _ := validation.ValidateMetricRequest("counter", "requests", "20")
	handler.service.UpdateMetric(ctx, metricReq2)

	metricReq3, _ := validation.ValidateMetricRequest("counter", "requests", "30")
	handler.service.UpdateMetric(ctx, metricReq3)

	// Проверяем, что значения накапливаются
	req, w := createChiContext("/value/{type}/{name}", map[string]string{
		"type": "counter",
		"name": "requests",
	})
	req.Method = "GET"

	handler.GetMetricValue(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "60", w.Body.String()) // 10 + 20 + 30 = 60
}

func TestMetricsHandler_UpdateMetric_GaugeReplacement(t *testing.T) {
	handler := createTestHandler()
	ctx := context.Background()

	// Добавляем gauge метрику несколько раз через валидацию
	metricReq1, _ := validation.ValidateMetricRequest("gauge", "temperature", "20.0")
	handler.service.UpdateMetric(ctx, metricReq1)

	metricReq2, _ := validation.ValidateMetricRequest("gauge", "temperature", "25.5")
	handler.service.UpdateMetric(ctx, metricReq2)

	metricReq3, _ := validation.ValidateMetricRequest("gauge", "temperature", "30.0")
	handler.service.UpdateMetric(ctx, metricReq3)

	// Проверяем, что последнее значение заменяет предыдущие
	req, w := createChiContext("/value/{type}/{name}", map[string]string{
		"type": "gauge",
		"name": "temperature",
	})
	req.Method = "GET"

	handler.GetMetricValue(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "30", w.Body.String()) // Последнее значение
}

func TestMetricsHandler_UpdateMetricJSON(t *testing.T) {
	handler := createTestHandler()

	tests := []struct {
		name           string
		requestBody    string
		contentType    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful gauge metric update",
			requestBody:    `{"id": "TestMetric", "type": "gauge", "value": 42.5}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "successful counter metric update",
			requestBody:    `{"id": "TestCounter", "type": "counter", "delta": 100}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "invalid content type",
			requestBody:    `{"id": "TestMetric", "type": "gauge", "value": 42.5}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Content-Type must be application/json\n",
		},
		{
			name:           "invalid JSON format",
			requestBody:    `{invalid json}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON format\n",
		},
		{
			name:           "missing metric ID",
			requestBody:    `{"type": "gauge", "value": 42.5}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "metric ID is required\n",
		},
		{
			name:           "missing metric type",
			requestBody:    `{"id": "TestMetric", "value": 42.5}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "metric type is required\n",
		},
		{
			name:           "gauge metric without value",
			requestBody:    `{"id": "TestMetric", "type": "gauge"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "value is required for gauge metric\n",
		},
		{
			name:           "counter metric without delta",
			requestBody:    `{"id": "TestCounter", "type": "counter"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "delta is required for counter metric\n",
		},
		{
			name:           "unsupported metric type",
			requestBody:    `{"id": "TestMetric", "type": "invalid", "value": 42.5}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unsupported metric type: invalid\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/update", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			handler.UpdateMetricJSON(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestMetricsHandler_GetMetricJSON(t *testing.T) {
	handler := createTestHandler()

	// Добавляем тестовые метрики через JSON API
	req1 := httptest.NewRequest("POST", "/update", strings.NewReader(`{"id": "TestGauge", "type": "gauge", "value": 42.5}`))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler.UpdateMetricJSON(w1, req1)

	req2 := httptest.NewRequest("POST", "/update", strings.NewReader(`{"id": "TestCounter", "type": "counter", "delta": 100}`))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.UpdateMetricJSON(w2, req2)

	tests := []struct {
		name           string
		requestBody    string
		contentType    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful gauge metric retrieval",
			requestBody:    `{"id": "TestGauge", "type": "gauge"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"TestGauge","type":"gauge","value":42.5}` + "\n",
		},
		{
			name:           "successful counter metric retrieval",
			requestBody:    `{"id": "TestCounter", "type": "counter"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"TestCounter","type":"counter","delta":100}` + "\n",
		},
		{
			name:           "invalid content type",
			requestBody:    `{"id": "TestGauge", "type": "gauge"}`,
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Content-Type must be application/json\n",
		},
		{
			name:           "invalid JSON format",
			requestBody:    `{invalid json}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON format\n",
		},
		{
			name:           "missing metric ID",
			requestBody:    `{"type": "gauge"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "metric ID is required\n",
		},
		{
			name:           "missing metric type",
			requestBody:    `{"id": "TestMetric"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "metric type is required\n",
		},
		{
			name:           "unsupported metric type",
			requestBody:    `{"id": "TestMetric", "type": "invalid"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unsupported metric type: invalid\n",
		},
		{
			name:           "metric not found",
			requestBody:    `{"id": "NonExistent", "type": "gauge"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "gauge metric not found: NonExistent\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/value", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			handler.GetMetricJSON(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}
