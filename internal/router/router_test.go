package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	r := New()
	assert.NotNil(t, r)
	assert.NotNil(t, r.mux)
}

func TestHandleFunc(t *testing.T) {
	r := New()

	// Регистрируем обработчик
	r.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Создаем тестовый сервер
	server := httptest.NewServer(r)
	defer server.Close()

	// Тестируем запрос
	resp, err := http.Get(server.URL + "/test")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandle(t *testing.T) {
	r := New()

	// Создаем обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Регистрируем обработчик
	r.Handle("/test", handler)

	// Создаем тестовый сервер
	server := httptest.NewServer(r)
	defer server.Close()

	// Тестируем запрос
	resp, err := http.Get(server.URL + "/test")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServeHTTP(t *testing.T) {
	r := New()

	// Регистрируем обработчик
	r.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Создаем запрос
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)

	// Создаем ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем ServeHTTP
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}

func TestGetMux(t *testing.T) {
	r := New()
	mux := r.GetMux()

	assert.NotNil(t, mux)
	assert.Equal(t, r.mux, mux)
}
