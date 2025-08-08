# internal/repository

Этот пакет содержит реализацию работы с базой данных, а также со внешними сервисами.

Важно, чтобы репозиторий не содержал бизнес-логику.

Репозиторий реализует паттерн Repository и служит абстракцией над различными источниками данных, такими как:
- базы данных (PostgreSQL, MySQL и др.)
- внешние API
- файловые системы
- кэши (Redis, Memcached)
- другие источники данных.

## Компоненты

### MetricsRepository

Интерфейс для работы с метриками:

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

### Архитектура репозитория

```mermaid
graph TB
    subgraph "Repository Layer"
        REPO[MetricsRepository Interface]
        IMPL[InMemoryRepository]
        ERR[Error Handling]
    end
    
    subgraph "Data Sources"
        MEM[In-Memory Storage]
        DB[(Database)]
        FILE[File System]
        CACHE[Cache]
    end
    
    subgraph "Service Layer"
        SERVICE[Service]
    end
    
    SERVICE --> REPO
    REPO --> IMPL
    IMPL --> MEM
    IMPL --> ERR
    
    REPO -.-> DB
    REPO -.-> FILE
    REPO -.-> CACHE
    
    style REPO fill:#e8f5e8
    style IMPL fill:#f3e5f5
    style ERR fill:#fff3e0
    style MEM fill:#e3f2fd
    style SERVICE fill:#e1f5fe
    style DB fill:#fff3e0
    style FILE fill:#fff3e0
    style CACHE fill:#fff3e0
```

### Паттерн Repository

```mermaid
classDiagram
    class MetricsRepository {
        <<interface>>
        +UpdateGauge(name, value) error
        +UpdateCounter(name, value) error
        +GetGauge(name) (float64, bool, error)
        +GetCounter(name) (int64, bool, error)
        +GetAllGauges() (GaugeMetrics, error)
        +GetAllCounters() (CounterMetrics, error)
    }
    
    class InMemoryRepository {
        -storage Storage
        +UpdateGauge(name, value) error
        +UpdateCounter(name, value) error
        +GetGauge(name) (float64, bool, error)
        +GetCounter(name) (int64, bool, error)
        +GetAllGauges() (GaugeMetrics, error)
        +GetAllCounters() (CounterMetrics, error)
    }
    
    class Storage {
        <<interface>>
        +UpdateGauge(name, value)
        +UpdateCounter(name, value)
        +GetGauge(name) (float64, bool)
        +GetCounter(name) (int64, bool)
        +GetAllGauges() GaugeMetrics
        +GetAllCounters() CounterMetrics
    }
    
    MetricsRepository <|.. InMemoryRepository
    InMemoryRepository --> Storage
```

### InMemoryMetricsRepository

Реализация репозитория в памяти:

```go
type InMemoryMetricsRepository struct {
    storage models.Storage
}
```

## Использование

```go
// Создание репозитория
storage := models.NewMemStorage()
repo := repository.NewInMemoryMetricsRepository(storage)

// Обновление метрик
err := repo.UpdateGauge("temperature", 23.5)
err = repo.UpdateCounter("requests", 100)

// Получение метрик
value, exists, err := repo.GetGauge("temperature")
value, exists, err := repo.GetCounter("requests")

// Получение всех метрик
gauges, err := repo.GetAllGauges()
counters, err := repo.GetAllCounters()
```

## Преимущества

1. **Абстракция** - скрывает детали работы с источниками данных
2. **Тестируемость** - легко создавать моки для тестирования
3. **Гибкость** - можно легко заменить реализацию
4. **Обработка ошибок** - все методы возвращают ошибки
5. **Разделение ответственности** - репозиторий не содержит бизнес-логику

## Тестирование

```bash
go test -v ./internal/repository
```
