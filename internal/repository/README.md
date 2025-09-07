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
    SaveToFile() error
    LoadFromFile() error
    SetSyncSave(sync bool)
}
```

### InMemoryMetricsRepository (Реализация)

Реализация репозитория в памяти с потокобезопасностью, поддержкой контекста и логированием:

```go
type InMemoryMetricsRepository struct {
    Gauges          models.GaugeMetrics
    Counters        models.CounterMetrics
    mu              sync.RWMutex
    logger          logger.Logger
    fileStoragePath string
    restore         bool
    syncSave        bool
}
```

## Использование

### Создание репозитория

```go
// Создаем логгер
appLogger := logger.NewSlogLogger()

// Создаем репозиторий в памяти с логгером и настройками персистентности
repo := repository.NewInMemoryMetricsRepository(
    appLogger,
    "/tmp/metrics.json",  // путь к файлу для сохранения
    true,                 // загружать метрики при старте
)

// Создаем сервис с репозиторием и логгером
service := service.NewMetricsService(repo, appLogger)
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

## Логирование

Репозиторий интегрирован с системой логирования для отслеживания операций:

```go
// Логирование операций обновления
repo.UpdateGauge(ctx, "temperature", 23.5)
// Логи: "Updating gauge metric" name=temperature value=23.5

// Логирование операций получения
value, exists, err := repo.GetGauge(ctx, "temperature")
// Логи: "Retrieved gauge metric" name=temperature value=23.5 exists=true

// Логирование ошибок контекста
if err == context.Canceled {
    // Логи: "Context canceled during operation" operation=UpdateGauge
}
```

### Уровни логирования

- **Debug**: Детальная информация об операциях
- **Info**: Основные операции (создание, обновление, получение)
- **Error**: Ошибки операций и отмены контекста

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

## 💾 Персистентность метрик

Репозиторий поддерживает сохранение и загрузку метрик в/из JSON файла:

### Сохранение метрик

```go
// Сохранение всех метрик в файл
err := repo.SaveToFile()
if err != nil {
    log.Printf("Failed to save metrics: %v", err)
}
```

### Загрузка метрик

```go
// Загрузка метрик из файла при старте
err := repo.LoadFromFile()
if err != nil {
    log.Printf("Failed to load metrics: %v", err)
}
```

### Синхронное сохранение

```go
// Включение синхронного сохранения (каждое обновление сразу на диск)
repo.SetSyncSave(true)

// Теперь каждое обновление метрики автоматически сохраняется
err := repo.UpdateGauge(ctx, "temperature", 23.5)
// Метрики автоматически сохраняются в файл
```

### Формат файла

Метрики сохраняются в JSON формате:

```json
[
  {"id":"LastGC","type":"gauge","value":1257894000000000000},
  {"id":"NumGC","type":"counter","delta":42}
]
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
    // Создаем логгер
    appLogger := logger.NewSlogLogger()
    
    // Создаем репозиторий с логгером
    repo := repository.NewInMemoryMetricsRepository(appLogger)
    
    // Создаем сервис с логгером
    service := service.NewMetricsService(repo, appLogger)
    
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
    
    // Создаем мок логгера
    mockLogger := &MockLogger{}
    
    // Создаем сервис с моком и логгером
    service := service.NewMetricsService(mockRepo, mockLogger)
    
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
