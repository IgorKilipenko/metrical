# internal/agent

Агент для сбора и отправки метрик.

## Архитектура агента

```mermaid
graph TB
    subgraph "Agent Package"
        AGENT[Agent]
        CONFIG[Config]
        METRICS[Metrics Collector]
        METRIC_INFO[MetricInfo]
        HTTP_HANDLER[HTTP Handler]
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
    AGENT --> HTTP_HANDLER
    
    METRICS --> RUNTIME
    METRICS --> RANDOM
    HTTP_HANDLER --> METRIC_INFO
    HTTP_HANDLER --> HTTP_CLIENT
    HTTP_CLIENT --> SERVER
    
    style AGENT fill:#f3e5f5
    style CONFIG fill:#e8f5e8
    style METRICS fill:#e3f2fd
    style METRIC_INFO fill:#fff8e1
    style HTTP_HANDLER fill:#f3e5f5
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
    Note over HTTPClient,Server: Retry логика при ошибках
```

## Возможности

### ✅ Основные функции
- **Сбор метрик**: 27 runtime метрик + 1 дополнительная (RandomValue) + 1 counter (PollCount)
- **Отправка метрик**: HTTP POST запросы с retry логикой (legacy и JSON API)
- **Graceful shutdown**: Корректное завершение работы
- **Потокобезопасность**: Использование `sync.RWMutex`
- **Конфигурация**: Гибкие настройки через структуру Config
- **Логирование**: Структурированное логирование через logger абстракцию
- **JSON API поддержка**: Отправка метрик через новые JSON эндпоинты
- **Gzip поддержка**: Автоматическое сжатие всех отправляемых данных

### ✅ Обработка ошибок
- **Retry логика**: 2 попытки с задержкой 100ms
- **Детальная диагностика**: Чтение тела ответа при ошибках
- **Структурированное логирование**: Детальное логирование операций и ошибок

### ✅ Конфигурация
- **Валидация**: Проверка корректности настроек
- **Значения по умолчанию**: Готовые к использованию настройки
- **Гибкость**: Поддержка кастомных URL и интервалов

### ✅ Gzip поддержка
- **Автоматическое сжатие**: Все JSON метрики автоматически сжимаются перед отправкой
- **HTTP заголовки**: Автоматически устанавливаются `Content-Encoding: gzip` и `Accept-Encoding: gzip`
- **Прозрачная работа**: Сжатие происходит автоматически без изменения API
- **Эффективность**: Значительное уменьшение размера передаваемых данных

## Структура файлов

### Основные файлы
- `agent.go` - основная логика агента (сбор, отправка, retry логика)
- `config.go` - конфигурация агента с валидацией
- `metrics.go` - работа с метриками (runtime + дополнительные)

### Тестовые файлы
- `agent_test.go` - тесты агента (создание, сбор метрик, потокобезопасность, graceful shutdown, подготовка метрик)
- `config_test.go` - тесты конфигурации (создание, валидация)
- `metrics_test.go` - тесты метрик (создание, заполнение, обновление)
- `gzip_test.go` - тесты gzip функциональности (сжатие, распаковка, интеграция)

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

# Проверка линтером
go vet ./internal/agent/...

# Тесты gzip функциональности
go test ./internal/agent/... -run TestAgent_CompressData -v
go test ./internal/agent/... -run TestAgent_CompressDataIntegration -v
```

## Конфигурация по умолчанию

```go
DefaultServerURL      = "http://localhost:8080"
DefaultPollInterval   = 2 * time.Second
DefaultReportInterval = 10 * time.Second
DefaultHTTPTimeout    = 10 * time.Second
DefaultMaxRetries     = 2
DefaultRetryDelay     = 100 * time.Millisecond
```

## Логирование

Агент использует структурированное логирование для отслеживания всех операций:

```go
// Создание агента с логгером
agentLogger := logger.NewSlogLogger()
agent := agent.NewAgent(config, agentLogger)

// Логирование сбора метрик
agent.collectMetrics()
// Логи: "collected metrics" total=29 gauges=28 counters=1

// Логирование отправки метрик
agent.sendMetrics()
// Логи: "successfully sent metrics" count=29

// Логирование ошибок (при verbose режиме)
// Логи: "error sending metric" name=Alloc error="connection refused"

// Логирование graceful shutdown
agent.Stop()
// Логи: "stopping agent"
// Логи: "polling stopped"  
// Логи: "reporting stopped"
// Логи: "agent stopped gracefully"
```

### Уровни логирования

- **Debug**: Детальная информация о метриках (при verbose режиме)
- **Info**: Основные события (сбор метрик, отправка, остановка)
- **Warn**: Предупреждения (частичные ошибки отправки)
- **Error**: Ошибки отправки метрик (при verbose режиме)
