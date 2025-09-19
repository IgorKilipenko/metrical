package agent

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

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
// ВНИМАНИЕ: Caller должен закрыть response.Body после использования
func createTestResponse(statusCode int, body string) *http.Response {
	resp := &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	// Возвращаем response, caller должен закрыть Body
	return resp
}

// createTestRetryClient создает тестовый RetryHTTPClient с моком
func createTestRetryClient(mockClient *MockHTTPClient) *RetryHTTPClient {
	mockLogger := testutils.NewMockLogger()
	return NewRetryHTTPClient(mockClient, mockLogger)
}

// setupMockClient настраивает мок клиент с ответами
func setupMockClient(mockClient *MockHTTPClient, responses []*http.Response, errors []error) {
	mockClient.SetResponses(responses, errors)
}

func TestNewRetryHTTPClient(t *testing.T) {
	mockClient := &MockHTTPClient{}
	mockLogger := testutils.NewMockLogger()

	client := NewRetryHTTPClient(mockClient, mockLogger)

	assert.NotNil(t, client)
	assert.Equal(t, mockClient, client.client)
	assert.Equal(t, mockLogger, client.logger)
}

func TestRetryHTTPClient_Do_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")
	expectedResp := createTestResponse(http.StatusOK, "success")
	defer expectedResp.Body.Close()

	setupMockClient(mockClient, []*http.Response{expectedResp}, []error{nil})
	client := createTestRetryClient(mockClient)

	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, mockClient.doCalls, 1)
	resp.Body.Close()
}

func TestRetryHTTPClient_Do_RetryOn5xx(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")

	// Первая попытка - 500 ошибка, вторая - успех
	resp1 := createTestResponse(http.StatusInternalServerError, "server error")
	resp2 := createTestResponse(http.StatusOK, "success")
	defer resp1.Body.Close()
	defer resp2.Body.Close()

	setupMockClient(mockClient, []*http.Response{resp1, resp2}, []error{nil, nil})
	client := createTestRetryClient(mockClient)

	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, mockClient.doCalls, 2)
	resp.Body.Close()
}

func TestRetryHTTPClient_Do_NoRetryOn4xx(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")

	// 404 ошибка - не должна вызывать retry
	resp := createTestResponse(http.StatusNotFound, "not found")
	defer resp.Body.Close()

	setupMockClient(mockClient, []*http.Response{resp}, []error{nil})
	client := createTestRetryClient(mockClient)

	resp, err := client.Do(req)

	assert.NoError(t, err) // Теперь 4xx ошибки возвращаются как успешные ответы
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Len(t, mockClient.doCalls, 1)
	resp.Body.Close()
}

func TestRetryHTTPClient_Do_MaxRetriesExceeded(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")

	// Все попытки возвращают 500 ошибку (4 попытки в DefaultRetryConfig)
	resp := createTestResponse(http.StatusInternalServerError, "server error")
	defer resp.Body.Close()

	// Настраиваем 4 попытки
	setupMockClient(mockClient, []*http.Response{resp, resp, resp, resp}, []error{nil, nil, nil, nil})
	client := createTestRetryClient(mockClient)

	resp, err := client.Do(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP operation failed after 4 attempts")
	assert.Len(t, mockClient.doCalls, 4)
	if resp != nil {
		resp.Body.Close()
	}
}

func TestRetryHTTPClient_Do_NetworkError(t *testing.T) {
	mockClient := &MockHTTPClient{}
	req := createTestRequest("GET", "http://example.com")
	networkErr := errors.New("connection refused") // Используем retryable ошибку

	// Настраиваем 4 попытки
	setupMockClient(mockClient, []*http.Response{nil, nil, nil, nil}, []error{networkErr, networkErr, networkErr, networkErr})
	client := createTestRetryClient(mockClient)

	resp, err := client.Do(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP operation failed after 4 attempts")
	assert.Nil(t, resp)
	assert.Len(t, mockClient.doCalls, 4)
	if resp != nil {
		resp.Body.Close()
	}
}

func TestRetryHTTPClient_Post_Success(t *testing.T) {
	mockClient := &MockHTTPClient{}
	expectedResp := createTestResponse(http.StatusOK, "success")
	defer expectedResp.Body.Close()

	setupMockClient(mockClient, []*http.Response{expectedResp}, []error{nil})
	client := createTestRetryClient(mockClient)

	resp, err := client.Post("http://example.com", "application/json", strings.NewReader("data"))

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, mockClient.postCalls, 1)
	resp.Body.Close()
}
