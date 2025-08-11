# internal/handler

Пакет для обработки HTTP-запросов. Содержит адаптеры между HTTP-транспортом и бизнес-логикой приложения.

## Назначение

- Обработка HTTP-запросов
- Валидация входных данных через пакет `validation`
- Вызовы сервисов с бизнес-логикой
- Формирование HTTP-ответов

## Архитектура обработчиков

```mermaid
graph TB
    subgraph "HTTP Layer"
        REQ[HTTP Request]
        RESP[HTTP Response]
    end
    
    subgraph "Handler Layer"
        MH[MetricsHandler]
        VAL[Validation Package]
        PARSER[Request Parser]
    end
    
    subgraph "Business Layer"
        SERVICE[Service]
        TEMPLATE[Template]
    end
    
    REQ --> MH
    MH --> VAL
    MH --> PARSER
    MH --> SERVICE
    MH --> TEMPLATE
    VAL --> SERVICE
    SERVICE --> RESP
    TEMPLATE --> RESP
    
    style REQ fill:#e3f2fd
    style RESP fill:#e3f2fd
    style MH fill:#f3e5f5
    style VAL fill:#e8f5e8
    style PARSER fill:#e8f5e8
    style SERVICE fill:#fff3e0
    style TEMPLATE fill:#fff3e0
```

### Поток обработки запроса

```mermaid
sequenceDiagram
    participant Router
    participant Handler
    participant Validation
    participant Service
    participant Response
    
    Router->>Handler: HTTP Request
    Handler->>Validation: ValidateMetricRequest(type, name, value)
    Validation-->>Handler: MetricRequest/ValidationError
    
    alt Valid Request
        Handler->>Service: UpdateMetric(MetricRequest)
        Service-->>Handler: Success
        Handler->>Response: Format Response
        Response-->>Router: HTTP Response
    else Invalid Request
        Handler->>Response: Error Response (400 Bad Request)
        Response-->>Router: HTTP Error
    end
    
    Note over Handler,Validation: Валидация через пакет validation
    Note over Handler,Service: Передача валидированных данных
```

## Компоненты

### MetricsHandler

Основной обработчик для работы с метриками:

```go
type MetricsHandler struct {
    service  *service.MetricsService
    template *template.MetricsTemplate
}
```

### Основные методы

- `UpdateMetric(w, r)` - обновление метрики с валидацией
- `GetMetricValue(w, r)` - получение значения метрики  
- `GetAllMetrics(w, r)` - получение всех метрик (HTML)
- `getAllMetricsData()` - приватный метод для получения данных метрик

## Принципы

- **Адаптер** - преобразует HTTP в вызовы сервисов
- **Валидация** - использует пакет `validation` для проверки входных данных
- **Обработка ошибок** - возвращает соответствующие HTTP коды
- **Разделение ответственности** - только HTTP логика, без бизнес-логики
- **Типобезопасность** - передача валидированных структур в сервисы