package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IgorKilipenko/metrical/internal/handler"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/service"
)

func TestSetupMetricsRoutes(t *testing.T) {
	// Создаем мок хендлер для тестирования
	storage := models.NewMemStorage()
	service := service.NewMetricsService(storage)
	handler := handler.NewMetricsHandler(service)

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
		if contentType != "text/html; charset=utf-8" {
			t.Errorf("Expected Content-Type text/html, got %s", contentType)
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
		// Сначала добавляем метрику
		storage.UpdateGauge("test", 123.45)

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
