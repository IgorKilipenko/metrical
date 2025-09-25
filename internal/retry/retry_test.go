package retry

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

// Helper функции для тестов
func createFastTestConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}
}

func createSlowTestConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 2,
		Delays:      []time.Duration{100 * time.Millisecond},
	}
}

func createInvalidConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 0,
		Delays:      []time.Duration{1 * time.Second},
	}
}

func createEmptyDelaysConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{},
	}
}

// Константы для тестов
const (
	fastTestDelay1      = 10 * time.Millisecond
	fastTestDelay2      = 20 * time.Millisecond
	expectedMinDuration = fastTestDelay1 + fastTestDelay2
	slowTestDelay       = 100 * time.Millisecond
)

func TestRetryableError(t *testing.T) {
	t.Run("retryable error implements interface", func(t *testing.T) {
		originalErr := errors.New("test error")
		retryableErr := NewRetryableError(originalErr)

		assert.True(t, IsRetryableError(retryableErr))
		assert.Equal(t, "test error", retryableErr.Error())

		// Проверяем Unwrap через errors.Unwrap
		unwrapped := errors.Unwrap(retryableErr)
		assert.Equal(t, originalErr, unwrapped)
	})

	t.Run("non-retryable error", func(t *testing.T) {
		originalErr := errors.New("test error")
		assert.False(t, IsRetryableError(originalErr))
	})
}

func TestIsPostgreSQLConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		pgErr    *pgconn.PgError
		expected bool
	}{
		{
			name: "ConnectionException",
			pgErr: &pgconn.PgError{
				Code: pgerrcode.ConnectionException,
			},
			expected: true,
		},
		{
			name: "ConnectionDoesNotExist",
			pgErr: &pgconn.PgError{
				Code: pgerrcode.ConnectionDoesNotExist,
			},
			expected: true,
		},
		{
			name: "ConnectionFailure",
			pgErr: &pgconn.PgError{
				Code: pgerrcode.ConnectionFailure,
			},
			expected: true,
		},
		{
			name: "SQLClientUnableToEstablishSQLConnection",
			pgErr: &pgconn.PgError{
				Code: pgerrcode.SQLClientUnableToEstablishSQLConnection,
			},
			expected: true,
		},
		{
			name: "SQLServerRejectedEstablishmentOfSQLConnection",
			pgErr: &pgconn.PgError{
				Code: pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			},
			expected: true,
		},
		{
			name: "TransactionResolutionUnknown",
			pgErr: &pgconn.PgError{
				Code: pgerrcode.TransactionResolutionUnknown,
			},
			expected: true,
		},
		{
			name: "ProtocolViolation",
			pgErr: &pgconn.PgError{
				Code: pgerrcode.ProtocolViolation,
			},
			expected: true,
		},
		{
			name: "Non-connection error",
			pgErr: &pgconn.PgError{
				Code: pgerrcode.UniqueViolation,
			},
			expected: false,
		},
		{
			name:     "Non-PG error",
			pgErr:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.pgErr != nil {
				err = tt.pgErr
			} else {
				err = errors.New("regular error")
			}

			result := IsPostgreSQLConnectionError(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsHTTPRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "connection reset",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "connection timeout",
			err:      errors.New("connection timeout"),
			expected: true,
		},
		{
			name:     "no such host",
			err:      errors.New("no such host"),
			expected: true,
		},
		{
			name:     "network is unreachable",
			err:      errors.New("network is unreachable"),
			expected: true,
		},
		{
			name:     "temporary failure",
			err:      errors.New("temporary failure in name resolution"),
			expected: true,
		},
		{
			name:     "i/o timeout",
			err:      errors.New("i/o timeout"),
			expected: true,
		},
		{
			name:     "context deadline exceeded",
			err:      errors.New("context deadline exceeded"),
			expected: true,
		},
		{
			name:     "connection lost",
			err:      errors.New("connection lost"),
			expected: true,
		},
		{
			name:     "broken pipe",
			err:      errors.New("broken pipe"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      errors.New("validation error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHTTPRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetry_Success(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() error {
		attempts++
		return nil
	}

	err := Retry(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			return NewRetryableError(errors.New("temporary error"))
		}
		return nil
	}

	start := time.Now()
	err := Retry(ctx, logger, config, operation)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
	// Проверяем, что прошло достаточно времени для retry
	assert.GreaterOrEqual(t, duration, expectedMinDuration)
}

func TestRetry_AllAttemptsFailed(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	operation := func() error {
		attempts++
		return NewRetryableError(errors.New("persistent error"))
	}

	start := time.Now()
	err := Retry(ctx, logger, config, operation)
	duration := time.Since(start)

	assert.Error(t, err)

	var retryErr *RetryExhaustedError
	assert.True(t, errors.As(err, &retryErr))
	assert.Equal(t, 3, retryErr.Attempts)
	assert.Equal(t, 3, attempts)
	// Проверяем, что прошло достаточно времени для всех retry
	assert.GreaterOrEqual(t, duration, expectedMinDuration)
}

func TestRetry_NonRetryableError(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() error {
		attempts++
		return errors.New("non-retryable error")
	}

	err := Retry(ctx, logger, DefaultRetryConfig, operation)

	assert.Error(t, err)
	assert.Equal(t, "non-retryable error", err.Error())
	assert.Equal(t, 1, attempts) // Должна быть только одна попытка
}

func TestRetry_ContextCancellation(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	operation := func() error {
		attempts++
		if attempts == 1 {
			// Отменяем контекст после первой попытки
			cancel()
		}
		return NewRetryableError(errors.New("temporary error"))
	}

	err := Retry(ctx, logger, DefaultRetryConfig, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
	assert.Equal(t, 1, attempts)
}

func TestRetry_ContextTimeout(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	attempts := 0
	operation := func() error {
		attempts++
		return NewRetryableError(errors.New("temporary error"))
	}

	err := Retry(ctx, logger, DefaultRetryConfig, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
	assert.GreaterOrEqual(t, attempts, 1)
}

func TestRetryWithResult_Success(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() (string, error) {
		attempts++
		return "success", nil
	}

	result, err := RetryWithResult(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 1, attempts)
}

func TestRetryWithResult_SuccessAfterRetries(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	operation := func() (int, error) {
		attempts++
		if attempts < 3 {
			return 0, NewRetryableError(errors.New("temporary error"))
		}
		return 42, nil
	}

	result, err := RetryWithResult(ctx, logger, config, operation)

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 3, attempts)
}

func TestRetryWithResult_AllAttemptsFailed(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	operation := func() (bool, error) {
		attempts++
		return false, NewRetryableError(errors.New("persistent error"))
	}

	result, err := RetryWithResult(ctx, logger, config, operation)

	assert.Error(t, err)

	var retryErr *RetryExhaustedError
	assert.True(t, errors.As(err, &retryErr))
	assert.Equal(t, 3, retryErr.Attempts)
	assert.False(t, result)
	assert.Equal(t, 3, attempts)
}

func TestRetry_PostgreSQLConnectionError(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			pgErr := &pgconn.PgError{
				Code: pgerrcode.ConnectionException,
			}
			return pgErr
		}
		return nil
	}

	err := Retry(ctx, logger, config, operation)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestRetry_HTTPRetryableError(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 2 {
			return &net.OpError{
				Op:  "dial",
				Err: errors.New("connection refused"),
			}
		}
		return nil
	}

	err := Retry(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestRetry_CustomConfig(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createSlowTestConfig()

	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 2 {
			return NewRetryableError(errors.New("temporary error"))
		}
		return nil
	}

	start := time.Now()
	err := Retry(ctx, logger, config, operation)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
	// Проверяем, что прошло время для одного retry
	assert.GreaterOrEqual(t, duration, slowTestDelay)
	assert.Less(t, duration, 2*slowTestDelay)
}

func TestRetry_EmptyDelays(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createEmptyDelaysConfig()

	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			return NewRetryableError(errors.New("temporary error"))
		}
		return nil
	}

	// Это должно вернуть ошибку валидации
	err := Retry(ctx, logger, config, operation)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidConfig)
}

func TestRetry_ZeroMaxAttempts(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createInvalidConfig()

	attempts := 0
	operation := func() error {
		attempts++
		return NewRetryableError(errors.New("temporary error"))
	}

	err := Retry(ctx, logger, config, operation)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidConfig)
	assert.Equal(t, 0, attempts)
}

func TestRetry_OneMaxAttempt(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := RetryConfig{
		MaxAttempts: 1,
		Delays:      []time.Duration{1 * time.Second},
	}

	attempts := 0
	operation := func() error {
		attempts++
		return NewRetryableError(errors.New("temporary error"))
	}

	err := Retry(ctx, logger, config, operation)

	assert.Error(t, err)

	var retryErr *RetryExhaustedError
	assert.True(t, errors.As(err, &retryErr))
	assert.Equal(t, 1, retryErr.Attempts)
	assert.Equal(t, 1, attempts)
}

func TestRetry_ContextCancellationDuringDelay(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	operation := func() error {
		attempts++
		if attempts == 1 {
			// Отменяем контекст во время задержки
			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()
		}
		return NewRetryableError(errors.New("temporary error"))
	}

	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{200 * time.Millisecond, 200 * time.Millisecond},
	}

	err := Retry(ctx, logger, config, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled during retry delay")
	assert.Equal(t, 1, attempts)
}

func TestRetry_ContextTimeoutDuringDelay(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	attempts := 0
	operation := func() error {
		attempts++
		return NewRetryableError(errors.New("temporary error"))
	}

	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{200 * time.Millisecond, 200 * time.Millisecond},
	}

	err := Retry(ctx, logger, config, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled during retry delay")
	assert.Equal(t, 1, attempts)
}

// Benchmark тесты
func BenchmarkRetry_Success(b *testing.B) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	operation := func() error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Retry(ctx, logger, DefaultRetryConfig, operation)
	}
}

func TestRetryWithResultAndExists_Success(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() (string, bool, error) {
		attempts++
		return "success", true, nil
	}

	result, exists, err := RetryWithResultAndExists(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.True(t, exists)
	assert.Equal(t, 1, attempts)
}

func TestRetryWithResultAndExists_SuccessAfterRetries(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	operation := func() (int, bool, error) {
		attempts++
		if attempts < 3 {
			return 0, false, NewRetryableError(errors.New("temporary error"))
		}
		return 42, true, nil
	}

	result, exists, err := RetryWithResultAndExists(ctx, logger, config, operation)

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.True(t, exists)
	assert.Equal(t, 3, attempts)
}

func TestRetryWithResultAndExists_AllAttemptsFailed(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	operation := func() (bool, bool, error) {
		attempts++
		return false, false, NewRetryableError(errors.New("persistent error"))
	}

	result, exists, err := RetryWithResultAndExists(ctx, logger, config, operation)

	assert.Error(t, err)

	var retryErr *RetryExhaustedError
	assert.True(t, errors.As(err, &retryErr))
	assert.Equal(t, 3, retryErr.Attempts)
	assert.False(t, result)
	assert.False(t, exists)
	assert.Equal(t, 3, attempts)
}

func TestRetryWithResultAndExists_NonRetryableError(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() (string, bool, error) {
		attempts++
		return "", false, errors.New("non-retryable error")
	}

	result, exists, err := RetryWithResultAndExists(ctx, logger, DefaultRetryConfig, operation)

	assert.Error(t, err)
	assert.Equal(t, "non-retryable error", err.Error())
	assert.Equal(t, "", result)
	assert.False(t, exists)
	assert.Equal(t, 1, attempts) // Должна быть только одна попытка
}

func TestRetryWithResultAndExists_ContextCancellation(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	operation := func() (int, bool, error) {
		attempts++
		if attempts == 1 {
			// Отменяем контекст после первой попытки
			cancel()
		}
		return 0, false, NewRetryableError(errors.New("temporary error"))
	}

	result, exists, err := RetryWithResultAndExists(ctx, logger, DefaultRetryConfig, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
	assert.Equal(t, 0, result)
	assert.False(t, exists)
	assert.Equal(t, 1, attempts)
}

func BenchmarkRetryWithResult_Success(b *testing.B) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	operation := func() (string, error) {
		return "success", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RetryWithResult(ctx, logger, DefaultRetryConfig, operation)
	}
}

func BenchmarkRetryWithResultAndExists_Success(b *testing.B) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	operation := func() (string, bool, error) {
		return "success", true, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RetryWithResultAndExists(ctx, logger, DefaultRetryConfig, operation)
	}
}

// Тесты для RetryHTTP
func TestRetryHTTP_Success(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	// Создаем тестовый HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	attempts := 0
	operation := func() (*http.Response, error) {
		attempts++
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		client := &http.Client{}
		return client.Do(req)
	}

	resp, err := RetryHTTP(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, attempts)
}

func TestRetryHTTP_SuccessAfterRetries(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	operation := func() (*http.Response, error) {
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		client := &http.Client{}
		return client.Do(req)
	}

	start := time.Now()
	resp, err := RetryHTTP(ctx, logger, config, operation)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, attempts)
	// Проверяем, что прошло достаточно времени для retry
	assert.GreaterOrEqual(t, duration, expectedMinDuration)
}

func TestRetryHTTP_5xxRetry(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	operation := func() (*http.Response, error) {
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		client := &http.Client{}
		return client.Do(req)
	}

	resp, err := RetryHTTP(ctx, logger, config, operation)

	assert.Error(t, err)
	assert.Nil(t, resp)
	if resp != nil {
		defer resp.Body.Close()
	}

	var retryErr *RetryExhaustedError
	assert.True(t, errors.As(err, &retryErr))
	assert.Equal(t, 3, retryErr.Attempts)
	assert.Equal(t, 3, attempts)
}

func TestRetryHTTP_4xxNoRetry(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	operation := func() (*http.Response, error) {
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		client := &http.Client{}
		return client.Do(req)
	}

	resp, err := RetryHTTP(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, 1, attempts) // Должна быть только одна попытка
}

func TestRetryHTTP_NetworkError(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	attempts := 0
	operation := func() (*http.Response, error) {
		attempts++
		if attempts < 3 {
			// Симулируем сетевую ошибку
			return nil, &net.OpError{
				Op:  "dial",
				Err: errors.New("connection refused"),
			}
		}
		// На третьей попытке возвращаем успешный ответ
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		client := &http.Client{}
		return client.Do(req)
	}

	resp, err := RetryHTTP(ctx, logger, config, operation)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, attempts)
}

func TestRetryHTTP_ContextCancellation(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	operation := func() (*http.Response, error) {
		attempts++
		if attempts == 1 {
			cancel() // Отменяем контекст после первой попытки
		}
		// Симулируем сетевую ошибку
		return nil, &net.OpError{
			Op:  "dial",
			Err: errors.New("connection refused"),
		}
	}

	resp, err := RetryHTTP(ctx, logger, DefaultRetryConfig, operation)

	assert.Error(t, err)
	assert.Nil(t, resp)
	if resp != nil {
		defer resp.Body.Close()
	}
	assert.Contains(t, err.Error(), "context cancelled")
	assert.Equal(t, 1, attempts)
}

// Тесты для валидации конфигурации
func TestRetryConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  RetryConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: RetryConfig{
				MaxAttempts: 3,
				Delays:      []time.Duration{1 * time.Second, 2 * time.Second},
			},
			wantErr: false,
		},
		{
			name: "valid config with single attempt",
			config: RetryConfig{
				MaxAttempts: 1,
				Delays:      []time.Duration{},
			},
			wantErr: false,
		},
		{
			name: "valid config with exact delays",
			config: RetryConfig{
				MaxAttempts: 3,
				Delays:      []time.Duration{1 * time.Second, 2 * time.Second},
			},
			wantErr: false,
		},
		{
			name: "valid config with more delays than needed",
			config: RetryConfig{
				MaxAttempts: 2,
				Delays:      []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
			},
			wantErr: false,
		},
		{
			name: "zero max attempts",
			config: RetryConfig{
				MaxAttempts: 0,
				Delays:      []time.Duration{1 * time.Second},
			},
			wantErr: true,
			errMsg:  "maxAttempts must be greater than 0, got 0",
		},
		{
			name: "negative max attempts",
			config: RetryConfig{
				MaxAttempts: -1,
				Delays:      []time.Duration{1 * time.Second},
			},
			wantErr: true,
			errMsg:  "maxAttempts must be greater than 0, got -1",
		},
		{
			name: "empty delays",
			config: RetryConfig{
				MaxAttempts: 3,
				Delays:      []time.Duration{},
			},
			wantErr: true,
			errMsg:  "delays cannot be empty",
		},
		{
			name: "nil delays",
			config: RetryConfig{
				MaxAttempts: 3,
				Delays:      nil,
			},
			wantErr: true,
			errMsg:  "delays cannot be empty",
		},
		{
			name: "insufficient delays",
			config: RetryConfig{
				MaxAttempts: 4,
				Delays:      []time.Duration{1 * time.Second},
			},
			wantErr: true,
			errMsg:  "delays length (1) must be at least MaxAttempts-1 (3)",
		},
		{
			name: "negative delay",
			config: RetryConfig{
				MaxAttempts: 3,
				Delays:      []time.Duration{1 * time.Second, -2 * time.Second},
			},
			wantErr: true,
			errMsg:  "delay at index 1 cannot be negative, got -2s",
		},
		{
			name: "zero delay",
			config: RetryConfig{
				MaxAttempts: 3,
				Delays:      []time.Duration{1 * time.Second, 0},
			},
			wantErr: false, // Zero delay is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Тесты для новых типов ошибок
func TestErrInvalidConfig(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	// Тест с неверной конфигурацией
	config := createInvalidConfig()

	operation := func() error {
		return nil
	}

	err := Retry(ctx, logger, config, operation)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidConfig)
}

func TestRetryExhaustedError(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := RetryConfig{
		MaxAttempts: 2,
		Delays:      []time.Duration{10 * time.Millisecond},
	}

	attempts := 0
	operation := func() error {
		attempts++
		return NewRetryableError(errors.New("persistent error"))
	}

	err := Retry(ctx, logger, config, operation)
	assert.Error(t, err)

	var retryErr *RetryExhaustedError
	assert.True(t, errors.As(err, &retryErr))
	assert.Equal(t, 2, retryErr.Attempts)
	assert.Equal(t, "persistent error", retryErr.LastError.Error())
	assert.Equal(t, 2, attempts)

	// Проверяем Unwrap
	assert.Equal(t, retryErr.LastError, errors.Unwrap(retryErr))
}

// Тесты для edge cases
func TestRetry_EmptyOperation(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	// Тест с nil операцией
	var operation func() error = nil

	// Это должно вызвать панику, но мы не тестируем панику в этом тесте
	// В реальном коде это должно быть обработано на уровне вызова
	assert.Panics(t, func() {
		Retry(ctx, logger, DefaultRetryConfig, operation)
	})
}

func TestRetry_ContextAlreadyCancelled(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем контекст сразу

	attempts := 0
	operation := func() error {
		attempts++
		return NewRetryableError(errors.New("temporary error"))
	}

	err := Retry(ctx, logger, DefaultRetryConfig, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
	assert.Equal(t, 0, attempts) // Операция не должна выполняться
}

func TestRetry_ContextWithDeadline(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond))
	defer cancel()

	attempts := 0
	operation := func() error {
		attempts++
		return NewRetryableError(errors.New("temporary error"))
	}

	config := RetryConfig{
		MaxAttempts: 5,
		Delays:      []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
	}

	err := Retry(ctx, logger, config, operation)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
	assert.GreaterOrEqual(t, attempts, 1)
}

func TestRetry_OperationReturnsNil(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() error {
		attempts++
		return nil // Всегда возвращаем nil
	}

	err := Retry(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestRetry_OperationPanics(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() error {
		attempts++
		panic("test panic")
	}

	// Это должно вызвать панику
	assert.Panics(t, func() {
		Retry(ctx, logger, DefaultRetryConfig, operation)
	})
	assert.Equal(t, 1, attempts)
}

func TestRetryWithResult_EmptyResult(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() (string, error) {
		attempts++
		return "", nil // Возвращаем пустую строку
	}

	result, err := RetryWithResult(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, 1, attempts)
}

func TestRetryWithResultAndExists_EmptyResult(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() (string, bool, error) {
		attempts++
		return "", false, nil // Возвращаем пустую строку и false
	}

	result, exists, err := RetryWithResultAndExists(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.Equal(t, "", result)
	assert.False(t, exists)
	assert.Equal(t, 1, attempts)
}

func TestRetry_DefaultConfigIsValid(t *testing.T) {
	// Проверяем, что DefaultRetryConfig валидна
	err := DefaultRetryConfig.Validate()
	assert.NoError(t, err)
}

func TestRetry_AllPostgreSQLConnectionErrors(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	// Тестируем все типы PostgreSQL connection ошибок
	pgErrorCodes := []string{
		pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure,
		pgerrcode.SQLClientUnableToEstablishSQLConnection,
		pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
		pgerrcode.TransactionResolutionUnknown,
		pgerrcode.ProtocolViolation,
	}

	for _, code := range pgErrorCodes {
		t.Run("PostgreSQL_"+code, func(t *testing.T) {
			attempts := 0
			operation := func() error {
				attempts++
				if attempts < 3 {
					pgErr := &pgconn.PgError{
						Code: code,
					}
					return pgErr
				}
				return nil
			}

			err := Retry(ctx, logger, config, operation)

			assert.NoError(t, err)
			assert.Equal(t, 3, attempts)
		})
	}
}

func TestRetry_AllHTTPRetryableErrors(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := createFastTestConfig()

	// Тестируем все типы HTTP retryable ошибок
	httpErrors := []string{
		"connection refused",
		"connection reset by peer",
		"connection timeout",
		"no such host",
		"network is unreachable",
		"temporary failure in name resolution",
		"i/o timeout",
		"context deadline exceeded",
		"connection lost",
		"broken pipe",
	}

	for _, errMsg := range httpErrors {
		t.Run("HTTP_"+errMsg, func(t *testing.T) {
			attempts := 0
			operation := func() error {
				attempts++
				if attempts < 3 {
					return &net.OpError{
						Op:  "dial",
						Err: errors.New(errMsg),
					}
				}
				return nil
			}

			err := Retry(ctx, logger, config, operation)

			assert.NoError(t, err)
			assert.Equal(t, 3, attempts)
		})
	}
}
