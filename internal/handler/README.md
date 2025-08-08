# internal/handler

В этом пакете размещаются обработчики HTTP-запросов. Здесь инкапсулируется так называемая логика представления.

Обычно хэндлеры реализуют:
- логику обработки запросов
- валидацию данных
- вызовы сервисов, в которых содержится бизнес-логика приложения
- формирование HTTP-ответов

Рекомендуется разбивать хэндлеры по функциональным группам и следовать принципу, где хэндлеры являются адаптерами между HTTP-транспортом и бизнес-логикой приложения.

## Архитектура обработчиков

```mermaid
graph TB
    subgraph "HTTP Layer"
        REQ[HTTP Request]
        RESP[HTTP Response]
    end
    
    subgraph "Handler Layer"
        MH[MetricsHandler]
        VAL[Validation]
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
    Handler->>Validation: Validate Input
    Validation-->>Handler: Valid/Invalid
    
    alt Valid Request
        Handler->>Service: Call Business Logic
        Service-->>Handler: Result
        Handler->>Response: Format Response
        Response-->>Router: HTTP Response
    else Invalid Request
        Handler->>Response: Error Response
        Response-->>Router: HTTP Error
    end
    
    Note over Handler,Validation: Валидация параметров
    Note over Handler,Service: Адаптер между HTTP и бизнес-логикой
```

## Компоненты

### MetricsHandler

Основной обработчик для работы с метриками:

```go
type MetricsHandler struct {
    service *service.MetricsService
    template *template.MetricsTemplate
}
```

### Основные методы

- `UpdateMetric(w, r)` - обновление метрики
- `GetMetricValue(w, r)` - получение значения метрики  
- `GetAllMetrics(w, r)` - получение всех метрик (HTML)

## Принципы

- **Адаптер** - преобразует HTTP в вызовы сервисов
- **Валидация** - проверяет корректность входных данных
- **Обработка ошибок** - возвращает соответствующие HTTP коды
- **Разделение ответственности** - только HTTP логика, без бизнес-логики