package agent

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
)

// HTTPClient интерфейс для HTTP клиента
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// RetryHTTPClient HTTP клиент с retry логикой
type RetryHTTPClient struct {
	client     HTTPClient
	maxRetries int
	retryDelay time.Duration
	logger     logger.Logger
}

// NewRetryHTTPClient создает новый HTTP клиент с retry логикой
func NewRetryHTTPClient(client HTTPClient, maxRetries int, retryDelay time.Duration, logger logger.Logger) *RetryHTTPClient {
	return &RetryHTTPClient{
		client:     client,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
		logger:     logger,
	}
}

// Do выполняет HTTP запрос с retry логикой
func (c *RetryHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.doWithRetry(req, func(r *http.Request) (*http.Response, error) {
		return c.client.Do(r)
	})
}

// Post выполняет POST запрос с retry логикой
func (c *RetryHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return c.doWithRetry(nil, func(r *http.Request) (*http.Response, error) {
		return c.client.Post(url, contentType, body)
	})
}

// doWithRetry выполняет запрос с retry логикой
func (c *RetryHTTPClient) doWithRetry(originalReq *http.Request, doFunc func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	var lastErr error

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		var req *http.Request
		var err error

		if originalReq != nil {
			// Создаем новый запрос для каждой попытки
			req, err = http.NewRequest(originalReq.Method, originalReq.URL.String(), originalReq.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			// Копируем заголовки
			for key, values := range originalReq.Header {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
		}

		resp, err := doFunc(req)
		if err != nil {
			lastErr = err
			if attempt == c.maxRetries {
				return nil, fmt.Errorf("failed after %d attempts: %w", c.maxRetries, err)
			}
			time.Sleep(c.retryDelay)
			continue
		}

		// Проверяем статус ответа
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		// Читаем тело ответа для диагностики
		bodyStr, _ := c.readResponseBody(resp)
		resp.Body.Close()

		// Retry только при серверных ошибках (5xx)
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			if attempt == c.maxRetries {
				return nil, fmt.Errorf("server error after %d attempts: status %d: %s", c.maxRetries, resp.StatusCode, bodyStr)
			}
			time.Sleep(c.retryDelay)
			continue
		}

		// Клиентские ошибки (4xx) и другие статусы не требуют retry
		return nil, fmt.Errorf("client error: status %d: %s", resp.StatusCode, bodyStr)
	}

	return nil, fmt.Errorf("failed to send request after %d attempts: %w", c.maxRetries, lastErr)
}

// readResponseBody читает тело ответа для диагностики
func (c *RetryHTTPClient) readResponseBody(resp *http.Response) (string, error) {
	const bufferSize = 1024
	body := make([]byte, bufferSize)
	n, readErr := resp.Body.Read(body)
	bodyStr := string(body[:n])

	if readErr != nil && !errors.Is(readErr, io.EOF) {
		// Логируем ошибку чтения для диагностики
		if c.logger != nil {
			c.logger.Warn("failed to read response body", "error", readErr)
		}
		bodyStr += fmt.Sprintf(" (read error: %v)", readErr)
	}

	return bodyStr, nil
}
