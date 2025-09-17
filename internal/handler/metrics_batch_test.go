package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IgorKilipenko/metrical/internal/logger"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandler_UpdateMetricsBatch(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := &mockMetricsRepository{}
	mockLogger := logger.NewSlogLogger()

	// Создаем сервис и хендлер
	metricsService := service.NewMetricsService(mockRepo, mockLogger)
	handler, err := NewMetricsHandler(metricsService, mockLogger)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    []models.Metrics
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful batch update",
			requestBody: []models.Metrics{
				{
					ID:    "temperature",
					MType: "gauge",
					Value: func() *float64 { v := 23.5; return &v }(),
				},
				{
					ID:    "requests",
					MType: "counter",
					Delta: func() *int64 { v := int64(100); return &v }(),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty batch",
			requestBody:    []models.Metrics{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Metrics batch cannot be empty",
		},
		{
			name: "invalid metric type",
			requestBody: []models.Metrics{
				{
					ID:    "invalid",
					MType: "invalid_type",
					Value: func() *float64 { v := 23.5; return &v }(),
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation error for metric at index 0",
		},
		{
			name: "missing gauge value",
			requestBody: []models.Metrics{
				{
					ID:    "temperature",
					MType: "gauge",
					Value: nil,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation error for metric at index 0",
		},
		{
			name: "missing counter delta",
			requestBody: []models.Metrics{
				{
					ID:    "requests",
					MType: "counter",
					Delta: nil,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation error for metric at index 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготавливаем JSON
			jsonData, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Создаем запрос
			req := httptest.NewRequest("POST", "/updates", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Создаем ResponseRecorder
			w := httptest.NewRecorder()

			// Выполняем запрос
			handler.UpdateMetricsBatch(w, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Проверяем ошибку, если ожидается
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			// Проверяем, что метод UpdateMetricsBatch был вызван для успешных случаев
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, mockRepo.updateMetricsBatchCalled)
				assert.Equal(t, tt.requestBody, mockRepo.lastBatchMetrics)
			}
		})
	}
}

func TestMetricsHandler_UpdateMetricsBatch_InvalidContentType(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := &mockMetricsRepository{}
	mockLogger := logger.NewSlogLogger()

	// Создаем сервис и хендлер
	metricsService := service.NewMetricsService(mockRepo, mockLogger)
	handler, err := NewMetricsHandler(metricsService, mockLogger)
	require.NoError(t, err)

	// Подготавливаем JSON
	jsonData, err := json.Marshal([]models.Metrics{
		{
			ID:    "temperature",
			MType: "gauge",
			Value: func() *float64 { v := 23.5; return &v }(),
		},
	})
	require.NoError(t, err)

	// Создаем запрос с неправильным Content-Type
	req := httptest.NewRequest("POST", "/updates", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "text/plain")

	// Создаем ResponseRecorder
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.UpdateMetricsBatch(w, req)

	// Проверяем статус
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Content-Type must be application/json")
}

func TestMetricsHandler_UpdateMetricsBatch_InvalidJSON(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := &mockMetricsRepository{}
	mockLogger := logger.NewSlogLogger()

	// Создаем сервис и хендлер
	metricsService := service.NewMetricsService(mockRepo, mockLogger)
	handler, err := NewMetricsHandler(metricsService, mockLogger)
	require.NoError(t, err)

	// Создаем запрос с невалидным JSON
	req := httptest.NewRequest("POST", "/updates", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.UpdateMetricsBatch(w, req)

	// Проверяем статус
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON format")
}

// Мок репозитория для тестирования
type mockMetricsRepository struct {
	updateMetricsBatchCalled bool
	lastBatchMetrics         []models.Metrics
	updateMetricsBatchError  error
}

func (m *mockMetricsRepository) UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error {
	m.updateMetricsBatchCalled = true
	m.lastBatchMetrics = metrics
	return m.updateMetricsBatchError
}

// Заглушки для остальных методов интерфейса MetricsRepository
func (m *mockMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	return nil
}
func (m *mockMetricsRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	return nil
}
func (m *mockMetricsRepository) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	return 0, false, nil
}
func (m *mockMetricsRepository) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	return 0, false, nil
}
func (m *mockMetricsRepository) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	return nil, nil
}
func (m *mockMetricsRepository) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	return nil, nil
}
func (m *mockMetricsRepository) SaveToFile() error {
	return nil
}
func (m *mockMetricsRepository) LoadFromFile() error {
	return nil
}
func (m *mockMetricsRepository) SetSyncSave(sync bool) {
}

func TestMetricsHandler_UpdateMetricsBatch_GzipCompression(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := &mockMetricsRepository{}
	mockLogger := logger.NewSlogLogger()

	// Создаем сервис и хендлер
	metricsService := service.NewMetricsService(mockRepo, mockLogger)
	handler, err := NewMetricsHandler(metricsService, mockLogger)
	require.NoError(t, err)

	// Подготавливаем тестовые метрики
	metrics := []models.Metrics{
		{
			ID:    "temperature",
			MType: "gauge",
			Value: func() *float64 { v := 23.5; return &v }(),
		},
		{
			ID:    "requests",
			MType: "counter",
			Delta: func() *int64 { v := int64(100); return &v }(),
		},
	}

	// Кодируем в JSON
	jsonData, err := json.Marshal(metrics)
	require.NoError(t, err)

	// Сжимаем данные
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err = gzWriter.Write(jsonData)
	require.NoError(t, err)
	err = gzWriter.Close()
	require.NoError(t, err)

	// Создаем запрос с gzip сжатием
	req := httptest.NewRequest("POST", "/updates", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	// Создаем ResponseRecorder
	w := httptest.NewRecorder()

	// Создаем middleware для gzip декомпрессии
	gzipMiddleware := func(next http.Handler) http.Handler {
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
			next.ServeHTTP(w, r)
		})
	}

	// Создаем хендлер с middleware
	handlerWithMiddleware := gzipMiddleware(http.HandlerFunc(handler.UpdateMetricsBatch))

	// Выполняем запрос
	handlerWithMiddleware.ServeHTTP(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, mockRepo.updateMetricsBatchCalled)
	assert.Equal(t, metrics, mockRepo.lastBatchMetrics)

	// Проверяем, что сжатие работает
	originalSize := len(jsonData)
	compressedSize := buf.Len()
	compressionRatio := float64(originalSize-compressedSize) / float64(originalSize) * 100

	t.Logf("Original size: %d bytes", originalSize)
	t.Logf("Compressed size: %d bytes", compressedSize)
	t.Logf("Compression ratio: %.1f%%", compressionRatio)

	// Проверяем, что сжатие действительно произошло
	assert.True(t, compressedSize < originalSize, "Data should be compressed")
}
