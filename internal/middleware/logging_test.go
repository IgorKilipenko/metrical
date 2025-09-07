package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	// Настраиваем zerolog для тестов
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.New(zerolog.NewTestWriter(t)))

	// Создаем middleware
	middleware := LoggingMiddleware()

	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("User-Agent", "test-agent")

	// Создаем ResponseRecorder
	rr := httptest.NewRecorder()

	// Применяем middleware
	middleware(handler).ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test response", rr.Body.String())
}

func TestResponseWriter(t *testing.T) {
	// Создаем тестовый ResponseRecorder
	rr := httptest.NewRecorder()

	// Создаем обертку ResponseWriter
	wrappedWriter := &ResponseWriter{
		ResponseWriter: rr,
		statusCode:     http.StatusOK,
	}

	// Тестируем WriteHeader
	wrappedWriter.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, wrappedWriter.statusCode)
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Тестируем Write
	response := "test response"
	size, err := wrappedWriter.Write([]byte(response))
	assert.NoError(t, err)
	assert.Equal(t, len(response), size)
	assert.Equal(t, len(response), wrappedWriter.size)
	assert.Equal(t, response, rr.Body.String())
}

func TestLoggingMiddlewareWithLogger(t *testing.T) {
	// Создаем тестовый логгер
	testWriter := zerolog.NewTestWriter(t)
	logger := zerolog.New(testWriter).Level(zerolog.InfoLevel)

	// Создаем middleware с кастомным логгером
	middleware := LoggingMiddlewareWithLogger(logger)

	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond) // Имитируем работу
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	// Создаем тестовый запрос
	req := httptest.NewRequest("POST", "/create", nil)
	req.RemoteAddr = "127.0.0.1:54321"
	req.Header.Set("User-Agent", "test-client")

	// Создаем ResponseRecorder
	rr := httptest.NewRecorder()

	// Применяем middleware
	middleware(handler).ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "created", rr.Body.String())
}

func TestLoggingMiddlewareWithError(t *testing.T) {
	// Настраиваем zerolog для тестов
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.New(zerolog.NewTestWriter(t)))

	// Создаем middleware
	middleware := LoggingMiddleware()

	// Создаем тестовый обработчик, который возвращает ошибку
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error occurred"))
	})

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/error", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	// Создаем ResponseRecorder
	rr := httptest.NewRecorder()

	// Применяем middleware
	middleware(handler).ServeHTTP(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "error occurred", rr.Body.String())
}
