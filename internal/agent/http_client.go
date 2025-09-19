package agent

import (
	"context"
	"io"
	"net/http"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/IgorKilipenko/metrical/internal/retry"
)

// HTTPClient интерфейс для HTTP клиента
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// RetryHTTPClient HTTP клиент с retry логикой
type RetryHTTPClient struct {
	client HTTPClient
	logger logger.Logger
}

// NewRetryHTTPClient создает новый HTTP клиент с retry логикой
func NewRetryHTTPClient(client HTTPClient, logger logger.Logger) *RetryHTTPClient {
	return &RetryHTTPClient{
		client: client,
		logger: logger,
	}
}

// Do выполняет HTTP запрос с retry логикой
func (c *RetryHTTPClient) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	return retry.RetryHTTP(ctx, c.logger, retry.DefaultRetryConfig, func() (*http.Response, error) {
		return c.client.Do(req)
	})
}

// Post выполняет POST запрос с retry логикой
func (c *RetryHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	ctx := context.Background()

	return retry.RetryHTTP(ctx, c.logger, retry.DefaultRetryConfig, func() (*http.Response, error) {
		return c.client.Post(url, contentType, body)
	})
}
