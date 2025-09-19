package retry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// RetryConfig конфигурация для retry механизма
type RetryConfig struct {
	MaxAttempts int             // Максимальное количество попыток (включая первую)
	Delays      []time.Duration // Интервалы между попытками
}

// Validate проверяет корректность конфигурации retry
func (c *RetryConfig) Validate() error {
	if c.MaxAttempts <= 0 {
		return fmt.Errorf("MaxAttempts must be greater than 0, got %d", c.MaxAttempts)
	}
	if len(c.Delays) == 0 {
		return fmt.Errorf("Delays cannot be empty")
	}
	if len(c.Delays) < c.MaxAttempts-1 {
		return fmt.Errorf("Delays length (%d) must be at least MaxAttempts-1 (%d)", len(c.Delays), c.MaxAttempts-1)
	}
	for i, delay := range c.Delays {
		if delay < 0 {
			return fmt.Errorf("Delay at index %d cannot be negative, got %v", i, delay)
		}
	}
	return nil
}

// DefaultRetryConfig стандартная конфигурация retry
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 4, // 1 основная + 3 повтора
	Delays:      []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
}

// RetryableError интерфейс для ошибок, которые можно повторить
type RetryableError interface {
	error
	IsRetryable() bool
}

// retryableError структура для retryable ошибок
type retryableError struct {
	err error
}

func (e *retryableError) Error() string {
	return e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}

func (e *retryableError) IsRetryable() bool {
	return true
}

// NewRetryableError создает retryable ошибку
func NewRetryableError(err error) RetryableError {
	return &retryableError{err: err}
}

// IsRetryableError проверяет, является ли ошибка retryable
func IsRetryableError(err error) bool {
	var retryableErr RetryableError
	return errors.As(err, &retryableErr) && retryableErr.IsRetryable()
}

// IsPostgreSQLConnectionError проверяет, является ли ошибка PostgreSQL connection error
func IsPostgreSQLConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	// Class 08 — Connection Exception
	return pgErr.Code == pgerrcode.ConnectionException ||
		pgErr.Code == pgerrcode.ConnectionDoesNotExist ||
		pgErr.Code == pgerrcode.ConnectionFailure ||
		pgErr.Code == pgerrcode.SQLClientUnableToEstablishSQLConnection ||
		pgErr.Code == pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection ||
		pgErr.Code == pgerrcode.TransactionResolutionUnknown ||
		pgErr.Code == pgerrcode.ProtocolViolation
}

// IsHTTPRetryableError проверяет, является ли HTTP ошибка retryable
func IsHTTPRetryableError(err error) bool {
	// Сетевые ошибки, таймауты, 5xx ошибки сервера
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Сетевые ошибки
	networkErrors := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"no such host",
		"network is unreachable",
		"temporary failure",
		"i/o timeout",
		"context deadline exceeded",
		"connection lost",
		"broken pipe",
	}

	for _, networkErr := range networkErrors {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(networkErr)) {
			return true
		}
	}

	return false
}

// IsHTTPResponseRetryable проверяет, является ли HTTP ответ retryable
func IsHTTPResponseRetryable(resp *http.Response) bool {
	if resp == nil {
		return false
	}

	// Retry только при серверных ошибках (5xx)
	return resp.StatusCode >= 500 && resp.StatusCode < 600
}

// getDelay безопасно получает задержку для попытки
func getDelay(config RetryConfig, attempt int) (time.Duration, error) {
	if attempt >= len(config.Delays) {
		return 0, fmt.Errorf("no delay configured for attempt %d", attempt+1)
	}
	return config.Delays[attempt], nil
}

// Retry выполняет операцию с retry логикой
func Retry(ctx context.Context, logger logger.Logger, config RetryConfig, operation func() error) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid retry config: %w", err)
	}

	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Проверяем контекст перед каждой попыткой
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled before attempt %d: %w", attempt+1, err)
		}

		// Выполняем операцию
		err := operation()
		if err == nil {
			// Успех
			if attempt > 0 {
				logger.Info("operation succeeded after retry", "attempt", attempt+1, "total_attempts", config.MaxAttempts)
			}
			return nil
		}

		lastErr = err

		// Проверяем, можно ли повторить операцию
		if !IsRetryableError(err) && !IsPostgreSQLConnectionError(err) && !IsHTTPRetryableError(err) {
			logger.Debug("error is not retryable", "error", err, "attempt", attempt+1)
			return err
		}

		// Если это последняя попытка, возвращаем ошибку
		if attempt == config.MaxAttempts-1 {
			logger.Error("operation failed after all retry attempts",
				"error", err,
				"total_attempts", config.MaxAttempts)
			return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, err)
		}

		// Ждем перед следующей попыткой
		delay, delayErr := getDelay(config, attempt)
		if delayErr != nil {
			return fmt.Errorf("failed to get delay: %w", delayErr)
		}
		logger.Warn("operation failed, retrying",
			"error", err,
			"attempt", attempt+1,
			"next_attempt_in", delay,
			"total_attempts", config.MaxAttempts)

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
		case <-time.After(delay):
			// Продолжаем к следующей попытке
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithResult выполняет операцию с retry логикой и возвращает результат
func RetryWithResult[T any](ctx context.Context, logger logger.Logger, config RetryConfig, operation func() (T, error)) (T, error) {
	if err := config.Validate(); err != nil {
		var zero T
		return zero, fmt.Errorf("invalid retry config: %w", err)
	}

	var zero T
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Проверяем контекст перед каждой попыткой
		if err := ctx.Err(); err != nil {
			return zero, fmt.Errorf("context cancelled before attempt %d: %w", attempt+1, err)
		}

		// Выполняем операцию
		result, err := operation()
		if err == nil {
			// Успех
			if attempt > 0 {
				logger.Info("operation succeeded after retry", "attempt", attempt+1, "total_attempts", config.MaxAttempts)
			}
			return result, nil
		}

		lastErr = err

		// Проверяем, можно ли повторить операцию
		if !IsRetryableError(err) && !IsPostgreSQLConnectionError(err) && !IsHTTPRetryableError(err) {
			logger.Debug("error is not retryable", "error", err, "attempt", attempt+1)
			return zero, err
		}

		// Если это последняя попытка, возвращаем ошибку
		if attempt == config.MaxAttempts-1 {
			logger.Error("operation failed after all retry attempts",
				"error", err,
				"total_attempts", config.MaxAttempts)
			return zero, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, err)
		}

		// Ждем перед следующей попыткой
		delay, delayErr := getDelay(config, attempt)
		if delayErr != nil {
			return zero, fmt.Errorf("failed to get delay: %w", delayErr)
		}
		logger.Warn("operation failed, retrying",
			"error", err,
			"attempt", attempt+1,
			"next_attempt_in", delay,
			"total_attempts", config.MaxAttempts)

		select {
		case <-ctx.Done():
			return zero, fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
		case <-time.After(delay):
			// Продолжаем к следующей попытке
		}
	}

	return zero, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryHTTP выполняет HTTP операцию с retry логикой
func RetryHTTP(ctx context.Context, logger logger.Logger, config RetryConfig, operation func() (*http.Response, error)) (*http.Response, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid retry config: %w", err)
	}

	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Проверяем контекст перед каждой попыткой
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled before attempt %d: %w", attempt+1, err)
		}

		// Выполняем операцию
		resp, err := operation()
		if err == nil {
			// Проверяем статус ответа
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				// Успех
				if attempt > 0 {
					logger.Info("HTTP operation succeeded after retry", "attempt", attempt+1, "total_attempts", config.MaxAttempts, "status", resp.StatusCode)
				}
				return resp, nil
			}

			// Проверяем, нужно ли повторить запрос
			if !IsHTTPResponseRetryable(resp) {
				// Клиентские ошибки (4xx) не требуют retry
				logger.Debug("HTTP error is not retryable", "status", resp.StatusCode, "attempt", attempt+1)
				return resp, nil
			}

			// Серверные ошибки (5xx) требуют retry
			err = NewRetryableError(fmt.Errorf("server error: status %d", resp.StatusCode))
		}

		lastErr = err

		// Проверяем, можно ли повторить операцию
		if !IsRetryableError(err) && !IsPostgreSQLConnectionError(err) && !IsHTTPRetryableError(err) {
			logger.Debug("error is not retryable", "error", err, "attempt", attempt+1)
			return nil, err
		}

		// Если это последняя попытка, возвращаем ошибку
		if attempt == config.MaxAttempts-1 {
			logger.Error("HTTP operation failed after all retry attempts",
				"error", err,
				"total_attempts", config.MaxAttempts)
			return nil, fmt.Errorf("HTTP operation failed after %d attempts: %w", config.MaxAttempts, err)
		}

		// Ждем перед следующей попыткой
		delay, delayErr := getDelay(config, attempt)
		if delayErr != nil {
			return nil, fmt.Errorf("failed to get delay: %w", delayErr)
		}
		logger.Warn("HTTP operation failed, retrying",
			"error", err,
			"attempt", attempt+1,
			"next_attempt_in", delay,
			"total_attempts", config.MaxAttempts)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
		case <-time.After(delay):
			// Продолжаем к следующей попытке
		}
	}

	return nil, fmt.Errorf("HTTP operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithResult3 выполняет операцию с retry логикой и возвращает результат с тремя значениями
func RetryWithResult3[T any](ctx context.Context, logger logger.Logger, config RetryConfig, operation func() (T, bool, error)) (T, bool, error) {
	if err := config.Validate(); err != nil {
		var zero T
		return zero, false, fmt.Errorf("invalid retry config: %w", err)
	}

	var zero T
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Проверяем контекст перед каждой попыткой
		if err := ctx.Err(); err != nil {
			return zero, false, fmt.Errorf("context cancelled before attempt %d: %w", attempt+1, err)
		}

		// Выполняем операцию
		result, exists, err := operation()
		if err == nil {
			// Успех
			if attempt > 0 {
				logger.Info("operation succeeded after retry", "attempt", attempt+1, "total_attempts", config.MaxAttempts)
			}
			return result, exists, nil
		}

		lastErr = err

		// Проверяем, можно ли повторить операцию
		if !IsRetryableError(err) && !IsPostgreSQLConnectionError(err) && !IsHTTPRetryableError(err) {
			logger.Debug("error is not retryable", "error", err, "attempt", attempt+1)
			return zero, false, err
		}

		// Если это последняя попытка, возвращаем ошибку
		if attempt == config.MaxAttempts-1 {
			logger.Error("operation failed after all retry attempts",
				"error", err,
				"total_attempts", config.MaxAttempts)
			return zero, false, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, err)
		}

		// Ждем перед следующей попыткой
		delay, delayErr := getDelay(config, attempt)
		if delayErr != nil {
			return zero, false, fmt.Errorf("failed to get delay: %w", delayErr)
		}
		logger.Warn("operation failed, retrying",
			"error", err,
			"attempt", attempt+1,
			"next_attempt_in", delay,
			"total_attempts", config.MaxAttempts)

		select {
		case <-ctx.Done():
			return zero, false, fmt.Errorf("context cancelled during retry delay: %w", ctx.Err())
		case <-time.After(delay):
			// Продолжаем к следующей попытке
		}
	}

	return zero, false, fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}
