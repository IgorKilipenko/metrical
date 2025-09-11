package agent

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// MockHTTPClient мок для HTTPClient интерфейса
type MockHTTPClient struct {
	doCalls   []*http.Request
	postCalls []struct {
		url         string
		contentType string
		body        io.Reader
	}
	responses []*http.Response
	errors    []error
	callIndex int
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.doCalls = append(m.doCalls, req)
	if m.callIndex < len(m.responses) {
		resp := m.responses[m.callIndex]
		err := m.errors[m.callIndex]
		m.callIndex++
		return resp, err
	}
	return nil, errors.New("no more responses configured")
}

func (m *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	m.postCalls = append(m.postCalls, struct {
		url         string
		contentType string
		body        io.Reader
	}{url, contentType, body})
	if m.callIndex < len(m.responses) {
		resp := m.responses[m.callIndex]
		err := m.errors[m.callIndex]
		m.callIndex++
		return resp, err
	}
	return nil, errors.New("no more responses configured")
}

func (m *MockHTTPClient) SetResponses(responses []*http.Response, errors []error) {
	m.responses = responses
	m.errors = errors
	m.callIndex = 0
}

// Test helpers для устранения дублирования кода

// createTestRequest создает тестовый HTTP запрос
func createTestRequest(method, url string) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	return req
}

// createTestResponse создает тестовый HTTP ответ
func createTestResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// createTestRetryClient создает тестовый RetryHTTPClient с моком
func createTestRetryClient(mockClient *MockHTTPClient) *RetryHTTPClient {
	mockLogger := testutils.NewMockLogger()
	return NewRetryHTTPClient(mockClient, 2, 1*time.Millisecond, mockLogger)
}

// setupMockClient настраивает мок клиент с ответами
func setupMockClient(mockClient *MockHTTPClient, responses []*http.Response, errors []error) {
	mockClient.SetResponses(responses, errors)
}

func TestNewRetryHTTPClient(t *testing.T) {
	mockClient := &MockHTTPClient{}
	mockLogger := testutils.NewMockLogger()

	client := NewRetryHTTPClient(mockClient, 3, 100*time.Millisecond, mockLogger)

	assert.NotNil(t, client)
	assert.Equal(t, mockClient, client.client)
	assert.Equal(t, 3, client.maxRetries)
	assert.Equal(t, 100*time.Millisecond, client.retryDelay)
	assert.Equal(t, mockLogger, client.logger)
}

func TestRetryHTTPClient_Do_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")
	expectedResp := createTestResponse(http.StatusOK, "success")

	setupMockClient(mockClient, []*http.Response{expectedResp}, []error{nil})
	client := createTestRetryClient(mockClient)

	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, mockClient.doCalls, 1)
}

func TestRetryHTTPClient_Do_RetryOn5xx(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")

	// Первая попытка - 500 ошибка, вторая - успех
	resp1 := createTestResponse(http.StatusInternalServerError, "server error")
	resp2 := createTestResponse(http.StatusOK, "success")

	setupMockClient(mockClient, []*http.Response{resp1, resp2}, []error{nil, nil})
	client := createTestRetryClient(mockClient)

	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, mockClient.doCalls, 2)
}

func TestRetryHTTPClient_Do_NoRetryOn4xx(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")

	// 404 ошибка - не должна вызывать retry
	resp := createTestResponse(http.StatusNotFound, "not found")

	setupMockClient(mockClient, []*http.Response{resp}, []error{nil})
	client := createTestRetryClient(mockClient)

	_, err := client.Do(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client error: status 404")
	assert.Len(t, mockClient.doCalls, 1)
}

func TestRetryHTTPClient_Do_MaxRetriesExceeded(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")

	// Все попытки возвращают 500 ошибку
	resp := createTestResponse(http.StatusInternalServerError, "server error")

	setupMockClient(mockClient, []*http.Response{resp, resp}, []error{nil, nil})
	client := createTestRetryClient(mockClient)

	_, err := client.Do(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server error after 2 attempts")
	assert.Len(t, mockClient.doCalls, 2)
}

func TestRetryHTTPClient_Do_NetworkError(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")
	networkErr := errors.New("network error")

	setupMockClient(mockClient, []*http.Response{nil, nil}, []error{networkErr, networkErr})
	client := createTestRetryClient(mockClient)

	resp, err := client.Do(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed after 2 attempts")
	assert.Nil(t, resp)
	assert.Len(t, mockClient.doCalls, 2)
}

func TestRetryHTTPClient_Post_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	expectedResp := createTestResponse(http.StatusOK, "success")

	setupMockClient(mockClient, []*http.Response{expectedResp}, []error{nil})
	client := createTestRetryClient(mockClient)

	resp, err := client.Post("http://example.com", "application/json", strings.NewReader("data"))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, mockClient.postCalls, 1)
}

func TestRetryHTTPClient_readResponseBody(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	client := &RetryHTTPClient{logger: mockLogger}

	tests := []struct {
		name     string
		body     string
		readErr  error
		expected string
	}{
		{
			name:     "successful read",
			body:     "response body",
			readErr:  nil,
			expected: "response body",
		},
		{
			name:     "EOF error",
			body:     "response body",
			readErr:  io.EOF,
			expected: "response body",
		},
		{
			name:     "other read error",
			body:     "response body",
			readErr:  errors.New("read failed"),
			expected: "response body (read error: read failed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Body: io.NopCloser(strings.NewReader(tt.body)),
			}

			// Мокаем Read для возврата ошибки
			if tt.readErr != nil {
				resp.Body = &mockReadCloser{
					Reader:  strings.NewReader(tt.body),
					readErr: tt.readErr,
				}
			}

			result, err := client.readResponseBody(resp)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// mockReadCloser для тестирования ошибок чтения
type mockReadCloser struct {
	io.Reader
	readErr error
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.readErr != nil {
		n, _ = m.Reader.Read(p)
		return n, m.readErr
	}
	return m.Reader.Read(p)
}

func (m *mockReadCloser) Close() error {
	return nil
}
