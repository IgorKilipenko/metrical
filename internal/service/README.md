# internal/service

Пакет `service` содержит бизнес-логику и играет ключевую роль в реализации функциональности приложения.

В нём описаны правила, процессы и операции, которые определяют поведение приложения.

Принципы организации:
- Сервисы должны быть независимы от деталей транспорта (HTTP, gRPC и т.д.).
- Взаимодействие с данными происходит через интерфейсы репозиториев.
- Каждый сервис должен иметь четко определенную область ответственности.

## Компоненты

### MetricsService

Сервис для работы с метриками:

```go
type MetricsService struct {
    repo repository.MetricsRepository
}
```

### Архитектура сервисного слоя

```mermaid
graph TB
    subgraph "Service Layer"
        MS[MetricsService]
        VAL[Validation Logic]
        BL[Business Rules]
    end
    
    subgraph "Dependencies"
        REPO[Repository Interface]
        MODEL[Data Models]
    end
    
    subgraph "External"
        HANDLER[HTTP Handler]
        TEMPLATE[Template Engine]
    end
    
    HANDLER --> MS
    MS --> VAL
    MS --> BL
    MS --> REPO
    REPO --> MODEL
    MS --> TEMPLATE
    
    style MS fill:#f3e5f5
    style VAL fill:#e8f5e8
    style BL fill:#e8f5e8
    style REPO fill:#fff3e0
    style MODEL fill:#fff3e0
    style HANDLER fill:#e3f2fd
    style TEMPLATE fill:#e3f2fd
```

### Поток бизнес-логики

```mermaid
sequenceDiagram
    participant Handler
    participant Service
    participant Validation
    participant BusinessLogic
    participant Repository
    
    Handler->>Service: UpdateMetric(type, name, value)
    Service->>Validation: Validate Input
    Validation-->>Service: Valid/Invalid
    
    alt Valid Input
        Service->>BusinessLogic: Process Metric
        BusinessLogic->>Repository: Update Data
        Repository-->>BusinessLogic: Success
        BusinessLogic-->>Service: Success
        Service-->>Handler: Success
    else Invalid Input
        Service-->>Handler: Error
    end
    
    Note over Service,BusinessLogic: Бизнес-правила
    Note over Validation: Валидация типов и значений
```

## Основные методы

### UpdateMetric
Обновляет метрику по типу, имени и значению:
```go
func (s *MetricsService) UpdateMetric(metricType, name, value string) error
```

### GetGauge/GetCounter
Получает значение метрики:
```go
func (s *MetricsService) GetGauge(name string) (float64, bool, error)
func (s *MetricsService) GetCounter(name string) (int64, bool, error)
```

### GetAllGauges/GetAllCounters
Получает все метрики:
```go
func (s *MetricsService) GetAllGauges() (models.GaugeMetrics, error)
func (s *MetricsService) GetAllCounters() (models.CounterMetrics, error)
```

## Использование

```go
// Создание сервиса
storage := models.NewMemStorage()
repo := repository.NewInMemoryMetricsRepository(storage)
service := service.NewMetricsService(repo)

// Обновление метрик
err := service.UpdateMetric("gauge", "temperature", "23.5")
err = service.UpdateMetric("counter", "requests", "100")

// Получение метрик
value, exists, err := service.GetGauge("temperature")
value, exists, err := service.GetCounter("requests")

// Получение всех метрик
gauges, err := service.GetAllGauges()
counters, err := service.GetAllCounters()
```

## Преимущества

1. **Разделение ответственности** - бизнес-логика отделена от транспорта и данных
2. **Тестируемость** - легко тестировать с моками репозитория
3. **Независимость** - не зависит от конкретных реализаций
4. **Обработка ошибок** - все методы возвращают ошибки
5. **Валидация** - проверяет корректность входных данных

## Тестирование

```bash
go test -v ./internal/service
```
