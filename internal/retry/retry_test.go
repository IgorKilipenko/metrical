package retry

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
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

	// Используем быструю конфигурацию для теста
	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}

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
	// Проверяем, что прошло достаточно времени для retry (10ms + 20ms = 30ms минимум)
	assert.GreaterOrEqual(t, duration, 30*time.Millisecond)
}

func TestRetry_AllAttemptsFailed(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	// Используем быструю конфигурацию для теста
	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}

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
	// Проверяем, что прошло достаточно времени для всех retry (10ms + 20ms = 30ms минимум)
	assert.GreaterOrEqual(t, duration, 30*time.Millisecond)
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

	// Используем быструю конфигурацию для теста
	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}

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

	// Используем быструю конфигурацию для теста
	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}

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

	// Используем быструю конфигурацию для теста
	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}

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

	config := RetryConfig{
		MaxAttempts: 2,
		Delays:      []time.Duration{100 * time.Millisecond},
	}

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
	// Проверяем, что прошло время для одного retry (100ms)
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
	assert.Less(t, duration, 200*time.Millisecond)
}

func TestRetry_EmptyDelays(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{}, // Пустой слайс
	}

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

	config := RetryConfig{
		MaxAttempts: 0,
		Delays:      []time.Duration{1 * time.Second},
	}

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

func TestRetryWithResult3_Success(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() (string, bool, error) {
		attempts++
		return "success", true, nil
	}

	result, exists, err := RetryWithResult3(ctx, logger, DefaultRetryConfig, operation)

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.True(t, exists)
	assert.Equal(t, 1, attempts)
}

func TestRetryWithResult3_SuccessAfterRetries(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	// Используем быструю конфигурацию для теста
	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}

	attempts := 0
	operation := func() (int, bool, error) {
		attempts++
		if attempts < 3 {
			return 0, false, NewRetryableError(errors.New("temporary error"))
		}
		return 42, true, nil
	}

	result, exists, err := RetryWithResult3(ctx, logger, config, operation)

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.True(t, exists)
	assert.Equal(t, 3, attempts)
}

func TestRetryWithResult3_AllAttemptsFailed(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	// Используем быструю конфигурацию для теста
	config := RetryConfig{
		MaxAttempts: 3,
		Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
	}

	attempts := 0
	operation := func() (bool, bool, error) {
		attempts++
		return false, false, NewRetryableError(errors.New("persistent error"))
	}

	result, exists, err := RetryWithResult3(ctx, logger, config, operation)

	assert.Error(t, err)

	var retryErr *RetryExhaustedError
	assert.True(t, errors.As(err, &retryErr))
	assert.Equal(t, 3, retryErr.Attempts)
	assert.False(t, result)
	assert.False(t, exists)
	assert.Equal(t, 3, attempts)
}

func TestRetryWithResult3_NonRetryableError(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	attempts := 0
	operation := func() (string, bool, error) {
		attempts++
		return "", false, errors.New("non-retryable error")
	}

	result, exists, err := RetryWithResult3(ctx, logger, DefaultRetryConfig, operation)

	assert.Error(t, err)
	assert.Equal(t, "non-retryable error", err.Error())
	assert.Equal(t, "", result)
	assert.False(t, exists)
	assert.Equal(t, 1, attempts) // Должна быть только одна попытка
}

func TestRetryWithResult3_ContextCancellation(t *testing.T) {
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

	result, exists, err := RetryWithResult3(ctx, logger, DefaultRetryConfig, operation)

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

func BenchmarkRetryWithResult3_Success(b *testing.B) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	operation := func() (string, bool, error) {
		return "success", true, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RetryWithResult3(ctx, logger, DefaultRetryConfig, operation)
	}
}

// Тесты для новых типов ошибок
func TestErrInvalidConfig(t *testing.T) {
	logger := testutils.NewMockLogger()
	ctx := context.Background()

	// Тест с неверной конфигурацией
	config := RetryConfig{
		MaxAttempts: 0,
		Delays:      []time.Duration{1 * time.Second},
	}

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
