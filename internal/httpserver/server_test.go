package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/handler"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/service"
	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers

// createTestHandler creates a test handler
func createTestHandler() *handler.MetricsHandler {
	mockLogger := testutils.NewMockLogger()
	repository := repository.NewInMemoryMetricsRepository(mockLogger, "/tmp/test-metrics.json", false)
	service := service.NewMetricsService(repository, mockLogger)
	return handler.NewMetricsHandler(service, mockLogger)
}

// createTestServer creates a test server with default configuration
func createTestServer(t *testing.T) *Server {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()
	srv, err := NewServer(":8080", handler, mockLogger)
	require.NoError(t, err)
	return srv
}

// createTestServerWithConfig creates a test server with custom configuration
func createTestServerWithConfig(t *testing.T, config *ServerConfig) *Server {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()
	srv, err := NewServerWithConfig(config, handler, mockLogger)
	require.NoError(t, err)
	return srv
}

// makeHTTPRequest helper for making HTTP requests in tests
func makeHTTPRequest(t *testing.T, srv *Server, method, path string) (*httptest.ResponseRecorder, int) {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w, w.Code
}

// assertHTTPResponse helper for asserting HTTP responses
func assertHTTPResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedBody string) {
	assert.Equal(t, expectedStatus, w.Code, "HTTP status code mismatch")
	if expectedBody != "" {
		body := strings.TrimSpace(w.Body.String())
		assert.Equal(t, expectedBody, body, "Response body mismatch")
	}
}

func TestNewServer(t *testing.T) {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()

	srv, err := NewServer(":8080", handler, mockLogger)

	require.NoError(t, err)
	require.NotNil(t, srv)
	assert.Equal(t, ":8080", srv.config.Addr)
	assert.NotNil(t, srv.handler)
}

func TestNewServerWithEmptyAddr(t *testing.T) {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()

	srv, err := NewServer("", handler, mockLogger)

	assert.Error(t, err)
	assert.Nil(t, srv)
	assert.Contains(t, err.Error(), "address cannot be empty")
}

func TestNewServerWithNilHandler(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	srv, err := NewServer(":8080", nil, mockLogger)

	assert.Error(t, err)
	assert.Nil(t, srv)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestServerIntegration(t *testing.T) {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()
	srv, err := NewServer(":8080", handler, mockLogger)
	require.NoError(t, err)

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

			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP status code mismatch")

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				assert.Equal(t, tt.expectedBody, body, "Response body mismatch")
			}
		})
	}
}

func TestServerEndToEnd(t *testing.T) {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()
	srv, err := NewServer(":8080", handler, mockLogger)
	require.NoError(t, err)

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

			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP status code mismatch")
		})
	}
}

func TestServerBasicFunctionality(t *testing.T) {
	srv := createTestServer(t)

	// Testing metric update
	w1, _ := makeHTTPRequest(t, srv, "POST", "/update/gauge/test/42.0")
	assertHTTPResponse(t, w1, http.StatusOK, "")

	// Testing metric retrieval
	w2, _ := makeHTTPRequest(t, srv, "GET", "/value/gauge/test")
	assertHTTPResponse(t, w2, http.StatusOK, "42")
}

func TestServerRedirects(t *testing.T) {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()
	srv, err := NewServer(":8080", handler, mockLogger)
	require.NoError(t, err)

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
			expectedStatus:   http.StatusMethodNotAllowed, // GET request to POST endpoint with trailing slashes
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
			require.NoError(t, err, "Failed to make request")
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "HTTP status code mismatch")

			if tt.expectedRedirect {
				location := resp.Header.Get("Location")
				assert.NotEmpty(t, location, "Expected redirect location header")
				t.Logf("Redirect from %s to %s", tt.path, location)
			}
		})
	}
}

func TestServerShutdownWithNilServer(t *testing.T) {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()
	srv, err := NewServer(":8080", handler, mockLogger)
	require.NoError(t, err)

	// Тестируем shutdown без запущенного сервера
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	assert.NoError(t, err, "Shutdown() should not return error when server is nil")
}

func TestServerServeHTTP(t *testing.T) {
	srv := createTestServer(t)

	// Testing ServeHTTP directly
	w, _ := makeHTTPRequest(t, srv, "GET", "/")
	assertHTTPResponse(t, w, http.StatusOK, "")
}

func TestServerCreateRouter(t *testing.T) {
	srv := createTestServer(t)

	// Checking that router is created
	assert.NotNil(t, srv.router, "Router should not be nil")
}

func TestNewServerWithConfig(t *testing.T) {
	config := &ServerConfig{
		Addr:         ":9090",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  20 * time.Second,
	}

	srv := createTestServerWithConfig(t, config)

	assert.Equal(t, ":9090", srv.config.Addr)
	assert.Equal(t, 10*time.Second, srv.config.ReadTimeout)
	assert.Equal(t, 10*time.Second, srv.config.WriteTimeout)
	assert.Equal(t, 20*time.Second, srv.config.IdleTimeout)
}

func TestNewServerWithNilConfig(t *testing.T) {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()

	srv, err := NewServerWithConfig(nil, handler, mockLogger)

	assert.Error(t, err)
	assert.Nil(t, srv)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

func TestDefaultServerConfig(t *testing.T) {
	config := DefaultServerConfig()

	require.NotNil(t, config, "DefaultServerConfig() should not return nil")
	assert.Equal(t, ":8080", config.Addr, "Default address mismatch")
	assert.Equal(t, 30*time.Second, config.ReadTimeout, "Default ReadTimeout mismatch")
	assert.Equal(t, 30*time.Second, config.WriteTimeout, "Default WriteTimeout mismatch")
	assert.Equal(t, 60*time.Second, config.IdleTimeout, "Default IdleTimeout mismatch")
}

// Дополнительные тесты с использованием testify

func TestServerConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *ServerConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid config",
			config:      &ServerConfig{Addr: ":8080"},
			expectError: false,
		},
		{
			name:        "Empty address",
			config:      &ServerConfig{Addr: ""},
			expectError: true,
			errorMsg:    "address cannot be empty",
		},
		{
			name:        "Nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := createTestHandler()
			mockLogger := testutils.NewMockLogger()

			srv, err := NewServerWithConfig(tt.config, handler, mockLogger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, srv)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, srv)
			}
		})
	}
}

func TestServerHTTPMethods(t *testing.T) {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()
	srv, err := NewServer(":8080", handler, mockLogger)
	require.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"GET root", "GET", "/", http.StatusOK},
		{"POST update", "POST", "/update/gauge/test/42.0", http.StatusOK},
		{"GET value", "GET", "/value/gauge/test", http.StatusOK},
		{"PUT not allowed", "PUT", "/update/gauge/test/42.0", http.StatusMethodNotAllowed},
		{"DELETE not allowed", "DELETE", "/value/gauge/test", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code,
				"Expected status %d for %s %s, got %d",
				tt.expectedStatus, tt.method, tt.path, w.Code)
		})
	}
}

func TestServerConcurrentRequests(t *testing.T) {
	handler := createTestHandler()
	mockLogger := testutils.NewMockLogger()
	srv, err := NewServer(":8080", handler, mockLogger)
	require.NoError(t, err)

	// Тестируем конкурентные запросы
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// Собираем результаты
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		assert.Equal(t, http.StatusOK, statusCode,
			"Concurrent request %d returned unexpected status", i+1)
	}
}

func TestServerEdgeCases(t *testing.T) {
	srv := createTestServer(t)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Very large gauge value",
			method:         "POST",
			path:           "/update/gauge/large_value/1.7976931348623157e+308",
			expectedStatus: http.StatusOK,
			description:    "Testing with maximum float64 value",
		},
		{
			name:           "Very large counter value",
			method:         "POST",
			path:           "/update/counter/large_counter/9223372036854775807",
			expectedStatus: http.StatusOK,
			description:    "Testing with maximum int64 value",
		},
		{
			name:           "Metric name with special characters",
			method:         "POST",
			path:           "/update/gauge/test-metric_123/42.5",
			expectedStatus: http.StatusOK,
			description:    "Testing metric name with hyphens and underscores",
		},
		{
			name:           "Invalid metric type",
			method:         "POST",
			path:           "/update/invalid/test/42.0",
			expectedStatus: http.StatusBadRequest,
			description:    "Testing with invalid metric type",
		},
		{
			name:           "Invalid gauge value",
			method:         "POST",
			path:           "/update/gauge/test/invalid_value",
			expectedStatus: http.StatusBadRequest,
			description:    "Testing with non-numeric gauge value",
		},
		{
			name:           "Invalid counter value",
			method:         "POST",
			path:           "/update/counter/test/invalid_value",
			expectedStatus: http.StatusBadRequest,
			description:    "Testing with non-numeric counter value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, _ := makeHTTPRequest(t, srv, tt.method, tt.path)
			assert.Equal(t, tt.expectedStatus, w.Code,
				"Test: %s - %s", tt.name, tt.description)
		})
	}
}

func TestServerPerformance(t *testing.T) {
	srv := createTestServer(t)

	// Performance test with multiple rapid requests
	const numRequests = 100
	start := time.Now()

	for i := 0; i < numRequests; i++ {
		w, _ := makeHTTPRequest(t, srv, "GET", "/")
		assert.Equal(t, http.StatusOK, w.Code,
			"Performance test request %d failed", i+1)
	}

	duration := time.Since(start)
	avgTime := duration / numRequests

	t.Logf("Performance test: %d requests in %v (avg: %v per request)",
		numRequests, duration, avgTime)

	// Assert that average response time is reasonable (less than 10ms)
	assert.Less(t, avgTime, 10*time.Millisecond,
		"Average response time should be less than 10ms")
}
