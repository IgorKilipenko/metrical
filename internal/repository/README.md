# internal/repository

Этот пакет содержит реализацию работы с базой данных, а также со внешними сервисами.

Важно, чтобы репозиторий не содержал бизнес-логику.

Репозиторий реализует паттерн Repository и служит абстракцией над различными источниками данных, такими как:
- базы данных (PostgreSQL, MySQL и др.)
- внешние API
- файловые системы
- кэши (Redis, Memcached)
- другие источники данных.

## Архитектура

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│     Service     │───▶│   Repository     │───▶│   Data Source   │
│                 │    │   (Interface)    │    │   (Memory/DB)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Компоненты

### MetricsRepository (Интерфейс)

Основной интерфейс для работы с метриками с поддержкой контекста:

```go
type MetricsRepository interface {
    UpdateGauge(ctx context.Context, name string, value float64) error
    UpdateCounter(ctx context.Context, name string, value int64) error
    GetGauge(ctx context.Context, name string) (float64, bool, error)
    GetCounter(ctx context.Context, name string) (int64, bool, error)
    GetAllGauges(ctx context.Context) (models.GaugeMetrics, error)
    GetAllCounters(ctx context.Context) (models.CounterMetrics, error)
}
```

### InMemoryMetricsRepository (Реализация)

Реализация репозитория в памяти с потокобезопасностью и поддержкой контекста:

```go
type InMemoryMetricsRepository struct {
    Gauges   models.GaugeMetrics
    Counters models.CounterMetrics
    mu       sync.RWMutex
}
```

## Использование

### Создание репозитория

```go
// Создаем репозиторий в памяти
repo := repository.NewInMemoryMetricsRepository()

// Создаем сервис с репозиторием
service := service.NewMetricsService(repo)
```

### Основные операции

```go
ctx := context.Background()

// Обновление метрик с контекстом
err := repo.UpdateGauge(ctx, "temperature", 23.5)
err := repo.UpdateCounter(ctx, "requests", 100)

// Получение метрик с контекстом
value, exists, err := repo.GetGauge(ctx, "temperature")
value, exists, err := repo.GetCounter(ctx, "requests")

// Получение всех метрик с контекстом
gauges, err := repo.GetAllGauges(ctx)
counters, err := repo.GetAllCounters(ctx)
```

### Работа с таймаутами и отменой

```go
// Создание контекста с таймаутом
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Операция будет отменена через 5 секунд
err := repo.UpdateGauge(ctx, "temperature", 23.5)
if err != nil {
    if err == context.DeadlineExceeded {
        log.Println("Operation timed out")
    } else if err == context.Canceled {
        log.Println("Operation was canceled")
    }
}
```

## Преимущества

- **Абстракция данных** - сервис не зависит от конкретной реализации хранения
- **Легкое тестирование** - можно легко мокать репозиторий
- **Расширяемость** - легко добавить новые реализации (PostgreSQL, Redis)
- **Потокобезопасность** - встроенная защита от гонки данных
- **Поддержка контекста** - отмена операций, таймауты, graceful shutdown
- **Чистая архитектура** - четкое разделение ответственности

## Особенности реализации

### Обработка контекста

Все методы репозитория проверяют отмену контекста:

```go
func (r *InMemoryMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
    // Проверяем отмену контекста
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    r.mu.Lock()
    defer r.mu.Unlock()
    r.Gauges[name] = value
    return nil
}
```

### Потокобезопасность

Реализация использует `sync.RWMutex` для обеспечения потокобезопасности:

- Операции записи (`UpdateGauge`, `UpdateCounter`) используют `Lock()`
- Операции чтения (`GetGauge`, `GetCounter`, `GetAllGauges`, `GetAllCounters`) используют `RLock()`

## Тестирование

```bash
go test -v ./internal/repository
```

### Тестирование с контекстом

```go
func TestRepositoryWithContext(t *testing.T) {
    repo := repository.NewInMemoryMetricsRepository()
    ctx := context.Background()
    
    // Тест с обычным контекстом
    err := repo.UpdateGauge(ctx, "test", 23.5)
    assert.NoError(t, err)
    
    // Тест с отмененным контекстом
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    
    err = repo.UpdateGauge(ctx, "test", 23.5)
    assert.Equal(t, context.Canceled, err)
}
```

## Примеры

### Базовое использование

```go
package main

import (
    "context"
    "time"
    
    "github.com/IgorKilipenko/metrical/internal/repository"
    "github.com/IgorKilipenko/metrical/internal/service"
)

func main() {
    // Создаем репозиторий
    repo := repository.NewInMemoryMetricsRepository()
    
    // Создаем сервис
    service := service.NewMetricsService(repo)
    
    // Создаем контекст с таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Используем сервис с контекстом
    err := service.UpdateMetric(ctx, &validation.MetricRequest{
        Type:  "gauge",
        Name:  "temperature",
        Value: 23.5,
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### Graceful Shutdown

```go
func gracefulShutdown(repo repository.MetricsRepository) {
    // Создаем контекст для graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Выполняем финальные операции
    gauges, err := repo.GetAllGauges(ctx)
    if err != nil {
        log.Printf("Error getting final gauges: %v", err)
        return
    }
    
    // Сохраняем данные или выполняем cleanup
    log.Printf("Final gauges: %v", gauges)
}
```

### Тестирование с моками

```go
func TestServiceWithMockRepository(t *testing.T) {
    // Создаем мок репозитория
    mockRepo := &MockMetricsRepository{}
    
    // Настраиваем ожидания с контекстом
    mockRepo.On("UpdateGauge", mock.Anything, "test", 23.5).Return(nil)
    
    // Создаем сервис с моком
    service := service.NewMetricsService(mockRepo)
    
    // Тестируем с контекстом
    ctx := context.Background()
    err := service.UpdateMetric(ctx, &validation.MetricRequest{
        Type:  "gauge",
        Name:  "test",
        Value: 23.5,
    })
    assert.NoError(t, err)
    
    // Проверяем, что мок был вызван
    mockRepo.AssertExpectations(t)
}
```
