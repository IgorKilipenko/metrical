package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/service"
)

// MockLogger для тестирования
type MockLogger struct {
	logs []string
}

func (m *MockLogger) SetLevel(level logger.LogLevel)                 {}
func (m *MockLogger) Debug(msg string, args ...any)                  {}
func (m *MockLogger) Info(msg string, args ...any)                   {}
func (m *MockLogger) Warn(msg string, args ...any)                   {}
func (m *MockLogger) Error(msg string, args ...any)                  {}
func (m *MockLogger) WithContext(ctx context.Context) logger.Logger  { return m }
func (m *MockLogger) WithFields(fields map[string]any) logger.Logger { return m }
func (m *MockLogger) Sync() error                                    { return nil }

func newMockLogger() logger.Logger {
	return &MockLogger{}
}

func TestSetupMetricsRoutes(t *testing.T) {
	// Создаем мок хендлер для тестирования
	mockLogger := newMockLogger()
	repository := repository.NewInMemoryMetricsRepository(mockLogger)
	service := service.NewMetricsService(repository, mockLogger)
	handler := handler.NewMetricsHandler(service, mockLogger)

	// Настраиваем маршруты
	router := SetupMetricsRoutes(handler)

	// Тестируем GET /
	t.Run("GET /", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Проверяем, что возвращается HTML
		contentType := w.Header().Get("Content-Type")
		if contentType != "text/html" && contentType != "text/html; charset=utf-8" {
			t.Errorf("Expected Content-Type text/html or text/html; charset=utf-8, got %s", contentType)
		}
	})

	// Тестируем POST /update/{type}/{name}/{value}
	t.Run("POST /update/gauge/test/123.45", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/update/gauge/test/123.45", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	// Тестируем GET /value/{type}/{name}
	t.Run("GET /value/gauge/test", func(t *testing.T) {
		// Сначала добавляем метрику через repository
		ctx := context.Background()
		err := repository.UpdateGauge(ctx, "test", 123.45)
		if err != nil {
			t.Fatalf("Failed to update gauge: %v", err)
		}

		req := httptest.NewRequest("GET", "/value/gauge/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		expectedBody := "123.45"
		if w.Body.String() != expectedBody {
			t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
		}
	})

	// Тестируем несуществующий маршрут
	t.Run("GET /nonexistent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestSetupHealthRoutes(t *testing.T) {
	router := SetupHealthRoutes()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expectedBody := "OK"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
	}
}
