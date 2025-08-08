package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	srv, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() returned error: %v", err)
	}
	if srv == nil {
		t.Fatal("NewServer() returned nil")
	}
	if srv.addr != ":8080" {
		t.Errorf("Expected addr :8080, got %s", srv.addr)
	}
	if srv.handler == nil {
		t.Fatal("handler is nil")
	}
}

func TestNewServerWithEmptyAddr(t *testing.T) {
	srv, err := NewServer("")
	if err == nil {
		t.Fatal("Expected error for empty address")
	}
	if srv != nil {
		t.Fatal("Expected nil server for empty address")
	}
}

func TestServerIntegration(t *testing.T) {
	srv, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() returned error: %v", err)
	}
	server := httptest.NewServer(srv)
	defer server.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Update gauge metric via HTTP",
			method:         "POST",
			path:           "/update/gauge/temperature/23.5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Update counter metric via HTTP",
			method:         "POST",
			path:           "/update/counter/requests/100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get gauge metric value",
			method:         "GET",
			path:           "/value/gauge/temperature",
			expectedStatus: http.StatusOK,
			expectedBody:   "23.5",
		},
		{
			name:           "Get counter metric value",
			method:         "GET",
			path:           "/value/counter/requests",
			expectedStatus: http.StatusOK,
			expectedBody:   "100",
		},
		{
			name:           "Get all metrics",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid HTTP method",
			method:         "GET",
			path:           "/update/gauge/temperature/23.5",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid URL format",
			method:         "POST",
			path:           "/update/gauge/temperature",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, body)
				}
			}
		})
	}
}

func TestServerEndToEnd(t *testing.T) {
	srv, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() returned error: %v", err)
	}
	server := httptest.NewServer(srv)
	defer server.Close()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Gauge metric end-to-end",
			method:         "POST",
			path:           "/update/gauge/temperature/23.5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Counter metric end-to-end",
			method:         "POST",
			path:           "/update/counter/requests/100",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestServerBasicFunctionality(t *testing.T) {
	srv, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() returned error: %v", err)
	}
	server := httptest.NewServer(srv)
	defer server.Close()

	// Тестируем обновление метрики
	req1 := httptest.NewRequest("POST", "/update/gauge/test/42.0", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w1.Code)
	}

	// Тестируем получение значения метрики
	req2 := httptest.NewRequest("GET", "/value/gauge/test", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w2.Code)
	}

	expectedValue := "42"
	if strings.TrimSpace(w2.Body.String()) != expectedValue {
		t.Errorf("Expected value %s, got %s", expectedValue, w2.Body.String())
	}
}

func TestServerRedirects(t *testing.T) {
	srv, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() returned error: %v", err)
	}
	server := httptest.NewServer(srv)
	defer server.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	tests := []struct {
		name             string
		path             string
		expectedStatus   int
		expectedRedirect bool
	}{
		{
			name:             "Path with double slash",
			path:             "/update/gauge//123.45",
			expectedStatus:   http.StatusMethodNotAllowed, // Chi router returns 405 for malformed paths
			expectedRedirect: false,
		},
		{
			name:             "Path with trailing slash",
			path:             "/update/gauge/test/123.45/",
			expectedStatus:   http.StatusNotFound, // Chi router doesn't redirect trailing slashes by default
			expectedRedirect: false,
		},
		{
			name:             "Normal path",
			path:             "/update/gauge/test/123.45",
			expectedStatus:   http.StatusMethodNotAllowed, // GET request to POST endpoint
			expectedRedirect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(server.URL + tt.path)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedRedirect {
				location := resp.Header.Get("Location")
				if location == "" {
					t.Error("Expected redirect location header")
				} else {
					t.Logf("Redirect from %s to %s", tt.path, location)
				}
			}
		})
	}
}
