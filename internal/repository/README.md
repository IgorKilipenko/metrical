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

Основной интерфейс для работы с метриками:

```go
type MetricsRepository interface {
    UpdateGauge(name string, value float64) error
    UpdateCounter(name string, value int64) error
    GetGauge(name string) (float64, bool, error)
    GetCounter(name string) (int64, bool, error)
    GetAllGauges() (models.GaugeMetrics, error)
    GetAllCounters() (models.CounterMetrics, error)
}
```

### InMemoryMetricsRepository (Реализация)

Реализация репозитория в памяти с потокобезопасностью:

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
// Обновление метрик
err := repo.UpdateGauge("temperature", 23.5)
err := repo.UpdateCounter("requests", 100)

// Получение метрик
value, exists, err := repo.GetGauge("temperature")
value, exists, err := repo.GetCounter("requests")

// Получение всех метрик
gauges, err := repo.GetAllGauges()
counters, err := repo.GetAllCounters()
```

## Преимущества

- **Абстракция данных** - сервис не зависит от конкретной реализации хранения
- **Легкое тестирование** - можно легко мокать репозиторий
- **Расширяемость** - легко добавить новые реализации (PostgreSQL, Redis)
- **Потокобезопасность** - встроенная защита от гонки данных
- **Чистая архитектура** - четкое разделение ответственности

## Тестирование

```bash
go test -v ./internal/repository
```

## Примеры

### Базовое использование

```go
package main

import (
    "github.com/IgorKilipenko/metrical/internal/repository"
    "github.com/IgorKilipenko/metrical/internal/service"
)

func main() {
    // Создаем репозиторий
    repo := repository.NewInMemoryMetricsRepository()
    
    // Создаем сервис
    service := service.NewMetricsService(repo)
    
    // Используем сервис
    err := service.UpdateMetric("gauge", "temperature", "23.5")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Тестирование с моками

```go
func TestServiceWithMockRepository(t *testing.T) {
    // Создаем мок репозитория
    mockRepo := &MockMetricsRepository{}
    
    // Настраиваем ожидания
    mockRepo.On("UpdateGauge", "test", 23.5).Return(nil)
    
    // Создаем сервис с моком
    service := service.NewMetricsService(mockRepo)
    
    // Тестируем
    err := service.UpdateMetric("gauge", "test", "23.5")
    assert.NoError(t, err)
    
    // Проверяем, что мок был вызван
    mockRepo.AssertExpectations(t)
}
```
