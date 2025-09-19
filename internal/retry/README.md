# Retry Package

Пакет `retry` предоставляет механизм повторных попыток выполнения операций с экспоненциальным backoff для обработки временных ошибок.

## Особенности

- ✅ **Экспоненциальный backoff** - интервалы между попытками увеличиваются (1s, 3s, 5s)
- ✅ **Ограниченное количество попыток** - максимум 4 попытки (1 основная + 3 повтора)
- ✅ **Поддержка контекста** - корректная обработка отмены и таймаутов
- ✅ **Автоматическое определение retryable ошибок** - PostgreSQL connection errors, HTTP network errors
- ✅ **Структурированное логирование** - детальные логи всех попыток
- ✅ **Generic поддержка** - работа с любыми типами результатов

## Использование

### Базовое использование

```go
import "github.com/IgorKilipenko/metrical/internal/retry"

// Простая операция без результата
err := retry.Retry(ctx, logger, retry.DefaultRetryConfig, func() error {
    return someOperation()
})

// Операция с результатом
result, err := retry.RetryWithResult(ctx, logger, retry.DefaultRetryConfig, func() (string, error) {
    return someOperationWithResult()
})
```

### Кастомная конфигурация

```go
config := retry.RetryConfig{
    MaxAttempts: 5, // 5 попыток
    Delays: []time.Duration{
        500 * time.Millisecond,
        1 * time.Second,
        2 * time.Second,
        4 * time.Second,
    },
}

err := retry.Retry(ctx, logger, config, operation)
```

### Создание retryable ошибок

```go
// Создание retryable ошибки
retryableErr := retry.NewRetryableError(errors.New("temporary error"))

// Проверка, является ли ошибка retryable
if retry.IsRetryableError(err) {
    // Ошибка будет повторена
}
```

## Типы ошибок

### Публичные ошибки

Пакет предоставляет типизированные ошибки для обработки пользователями:

#### ErrInvalidConfig
```go
var ErrInvalidConfig = errors.New("invalid retry configuration")
```

Возвращается при неверной конфигурации retry (например, MaxAttempts <= 0).

#### RetryExhaustedError
```go
type RetryExhaustedError struct {
    Attempts  int
    LastError error
}
```

Возвращается когда исчерпаны все попытки retry. Содержит:
- `Attempts` - количество выполненных попыток
- `LastError` - последняя ошибка, которая привела к неудаче

Пример использования:
```go
err := retry.Retry(ctx, logger, config, operation)
if err != nil {
    var retryErr *retry.RetryExhaustedError
    if errors.As(err, &retryErr) {
        log.Printf("Failed after %d attempts, last error: %v", 
            retryErr.Attempts, retryErr.LastError)
    }
}
```

### Автоматически определяемые retryable ошибки

#### PostgreSQL Connection Errors (Class 08)

Автоматически определяются как retryable:
- `ConnectionException` (08000)
- `ConnectionDoesNotExist` (08003)
- `ConnectionFailure` (08006)
- `SQLClientUnableToEstablishSQLConnection` (08001)
- `SQLServerRejectedEstablishmentOfSQLConnection` (08004)
- `TransactionResolutionUnknown` (08007)
- `ProtocolViolation` (08P01)

### HTTP Network Errors

Автоматически определяются как retryable:
- `connection refused`
- `connection reset`
- `connection timeout`
- `no such host`
- `network is unreachable`
- `temporary failure`
- `i/o timeout`
- `context deadline exceeded`
- `connection lost`
- `broken pipe`

### Custom Retryable Errors

```go
type MyRetryableError struct {
    err error
}

func (e *MyRetryableError) Error() string {
    return e.err.Error()
}

func (e *MyRetryableError) Unwrap() error {
    return e.err
}

func (e *MyRetryableError) IsRetryable() bool {
    return true
}
```

## Конфигурация

### DefaultRetryConfig

```go
var DefaultRetryConfig = RetryConfig{
    MaxAttempts: 4, // 1 основная + 3 повтора
    Delays: []time.Duration{
        1 * time.Second,  // После 1-й неудачи
        3 * time.Second,  // После 2-й неудачи
        5 * time.Second,  // После 3-й неудачи
    },
}
```

### RetryConfig

```go
type RetryConfig struct {
    MaxAttempts int             // Максимальное количество попыток
    Delays      []time.Duration // Интервалы между попытками
}
```

## Логирование

Пакет автоматически логирует:
- **Info** - успешное выполнение после retry
- **Warn** - неудачная попытка с информацией о следующей попытке
- **Error** - окончательная неудача после всех попыток
- **Debug** - не-retryable ошибки

Пример логов:
```
WARN operation failed, retrying error="connection refused" attempt=1 next_attempt_in=1s total_attempts=4
WARN operation failed, retrying error="connection refused" attempt=2 next_attempt_in=3s total_attempts=4
INFO operation succeeded after retry attempt=3 total_attempts=4
```

## Работа с контекстом

Пакет полностью поддерживает `context.Context`:
- Проверка отмены перед каждой попыткой
- Прерывание задержек при отмене контекста
- Поддержка таймаутов

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := retry.Retry(ctx, logger, config, operation)
```

## Примеры использования

### HTTP Client

```go
func (c *HTTPClient) SendMetrics(ctx context.Context, metrics []models.Metrics) error {
    return retry.Retry(ctx, c.logger, retry.DefaultRetryConfig, func() error {
        return c.sendMetricsOnce(ctx, metrics)
    })
}
```

### PostgreSQL Repository

```go
func (r *PostgreSQLMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
    return retry.Retry(ctx, r.logger, retry.DefaultRetryConfig, func() error {
        return r.updateGaugeOnce(ctx, name, value)
    })
}
```

### Service Layer

```go
func (s *MetricsService) UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error {
    return retry.Retry(ctx, s.logger, retry.DefaultRetryConfig, func() error {
        return s.repository.UpdateMetricsBatch(ctx, metrics)
    })
}
```

## Тестирование

Пакет включает comprehensive тесты:
- Успешное выполнение с первой попытки
- Успешное выполнение после retry
- Окончательная неудача после всех попыток
- Не-retryable ошибки
- Отмена контекста
- Таймауты контекста
- PostgreSQL connection errors
- HTTP network errors
- Кастомные конфигурации
- Benchmark тесты

Запуск тестов:
```bash
go test -v ./internal/retry/
```

## Производительность

- Минимальный overhead для успешных операций
- Эффективная проверка типов ошибок
- Оптимизированные задержки
- Поддержка benchmark тестов

## Безопасность

- Защита от бесконечных циклов (ограниченное количество попыток)
- Корректная обработка контекста
- Безопасная работа с горутинами
- Валидация конфигурации

## Интеграция

Пакет легко интегрируется в существующий код:
1. Импортируйте пакет
2. Оберните операцию в retry.Retry или retry.RetryWithResult
3. Настройте логирование
4. Обработайте финальные ошибки

## Лучшие практики

1. **Используйте контекст** - всегда передавайте context.Context
2. **Настройте логирование** - используйте структурированное логирование
3. **Обрабатывайте финальные ошибки** - не игнорируйте ошибки после всех попыток
4. **Используйте типизированные ошибки** - проверяйте `ErrInvalidConfig` и `RetryExhaustedError`
5. **Тестируйте retry логику** - убедитесь, что retry работает корректно
6. **Мониторьте метрики** - отслеживайте количество retry в production
7. **Настройте конфигурацию** - адаптируйте под ваши нужды

### Обработка ошибок

```go
err := retry.Retry(ctx, logger, config, operation)
if err != nil {
    // Проверяем тип ошибки
    if errors.Is(err, retry.ErrInvalidConfig) {
        // Обработка неверной конфигурации
        log.Error("Invalid retry configuration", "error", err)
        return err
    }
    
    var retryErr *retry.RetryExhaustedError
    if errors.As(err, &retryErr) {
        // Обработка исчерпания попыток
        log.Error("All retry attempts exhausted", 
            "attempts", retryErr.Attempts,
            "last_error", retryErr.LastError)
        
        // Можно попробовать альтернативную стратегию
        return handleRetryExhaustion(retryErr)
    }
    
    // Другие ошибки
    return err
}
```

## Troubleshooting

### Частые проблемы

1. **Слишком много retry** - проверьте, что ошибки действительно retryable
2. **Медленные операции** - настройте таймауты контекста
3. **Неожиданные retry** - проверьте логи для понимания причин
4. **Контекст отменяется** - убедитесь, что контекст не отменяется слишком рано

### Отладка

Включите debug логирование для детальной информации:
```go
logger.SetLevel("debug")
```

Проверьте логи на наличие:
- Причин retry
- Времени между попытками
- Финальных ошибок
- Отмены контекста
