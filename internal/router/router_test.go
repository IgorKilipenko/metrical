package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New() returned nil")
	}
	if r.router == nil {
		t.Fatal("router is nil")
	}
}

func TestHandleFunc(t *testing.T) {
	r := New()
	called := false

	r.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandle(t *testing.T) {
	r := New()
	called := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	r.Handle("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestServeHTTP(t *testing.T) {
	r := New()
	called := false

	r.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetRouter(t *testing.T) {
	r := New()
	chiRouter := r.GetRouter()
	if chiRouter == nil {
		t.Fatal("GetRouter() returned nil")
	}
}

func TestGetMux(t *testing.T) {
	r := New()
	mux := r.GetMux()
	if mux != nil {
		t.Error("GetMux() should return nil with chi router")
	}
}
