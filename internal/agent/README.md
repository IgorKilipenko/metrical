# internal/agent

Агент для сбора и отправки метрик.

## Архитектура агента

```mermaid
graph TB
    subgraph "Agent Package"
        AGENT[Agent]
        CONFIG[Config]
        METRICS[Metrics Collector]
    end
    
    subgraph "Runtime"
        RUNTIME[runtime.MemStats]
        RANDOM[Random Generator]
    end
    
    subgraph "Network"
        HTTP_CLIENT[HTTP Client]
        SERVER[Server]
    end
    
    AGENT --> CONFIG
    AGENT --> METRICS
    AGENT --> HTTP_CLIENT
    
    METRICS --> RUNTIME
    METRICS --> RANDOM
    HTTP_CLIENT --> SERVER
    
    style AGENT fill:#f3e5f5
    style CONFIG fill:#e8f5e8
    style METRICS fill:#e3f2fd
    style RUNTIME fill:#fff3e0
    style RANDOM fill:#fff3e0
    style HTTP_CLIENT fill:#e1f5fe
    style SERVER fill:#e1f5fe
```

### Поток сбора метрик

```mermaid
sequenceDiagram
    participant Agent
    participant Collector
    participant Runtime
    participant HTTPClient
    participant Server
    
    loop Every 2 seconds
        Agent->>Collector: Collect Metrics
        Collector->>Runtime: Get MemStats
        Runtime-->>Collector: 29+ Metrics
        Collector-->>Agent: Metrics Data
    end
    
    loop Every 10 seconds
        Agent->>HTTPClient: Send Metrics
        HTTPClient->>Server: HTTP POST
        Server-->>HTTPClient: 200 OK
        HTTPClient-->>Agent: Success
    end
    
    Note over Agent,Collector: Потокобезопасный сбор
    Note over HTTPClient,Server: JSON формат данных
```

## Структура файлов

### Основные файлы
- `agent.go` - основная логика агента
- `config.go` - конфигурация агента
- `metrics.go` - работа с метриками

### Тестовые файлы
- `agent_test.go` - тесты агента (создание, сбор метрик, потокобезопасность, graceful shutdown)
- `config_test.go` - тесты конфигурации (создание, валидация)
- `metrics_test.go` - тесты метрик (создание, заполнение, обновление)

## Запуск тестов

```bash
# Все тесты агента
go test ./internal/agent/... -v

# Только тесты конфигурации
go test ./internal/agent/config_test.go ./internal/agent/config.go -v

# Только тесты агента
go test ./internal/agent/agent_test.go ./internal/agent/agent.go ./internal/agent/config.go ./internal/agent/metrics.go -v

# Только тесты метрик
go test ./internal/agent/metrics_test.go ./internal/agent/metrics.go -v
```
