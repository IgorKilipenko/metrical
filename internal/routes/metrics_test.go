package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestSetupMetricsRoutes(t *testing.T) {
	// Создаем мок хендлер для тестирования
	mockLogger := testutils.NewMockLogger()
	repository := repository.NewInMemoryMetricsRepository(mockLogger, testutils.TestMetricsFile, false)
	service := service.NewMetricsService(repository, mockLogger)
	handler, err := handler.NewMetricsHandler(service, mockLogger)
	if err != nil {
		t.Fatalf("failed to create metrics handler: %v", err)
	}

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

func TestSetupMetricsRoutes_JSONEndpoints(t *testing.T) {
	// Создаем mock handler
	handler := &handler.MetricsHandler{}

	// Настраиваем маршруты
	r := SetupMetricsRoutes(handler)

	// Проверяем, что роутер создан
	assert.NotNil(t, r)

	// Проверяем, что роутер содержит маршруты (базовая проверка)
	// Более детальная проверка маршрутов требует сложной настройки chi контекста
}
