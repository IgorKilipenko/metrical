# Пакет `app`

Пакет `app` предоставляет основную логику инициализации и запуска приложения метрик.

## Назначение

Пакет инкапсулирует всю логику запуска приложения, включая:
- Загрузку конфигурации из переменных окружения
- Создание и запуск HTTP сервера
- Graceful shutdown при получении сигналов
- Обработку ошибок и логирование

## Компоненты

### `App`
Основная структура приложения, которая управляет жизненным циклом сервера.

```go
type App struct {
    server *httpserver.Server
    port   string
}
```

### Архитектура приложения

```mermaid
graph TB
    subgraph "Application Layer"
        APP[App]
        CONFIG[Config]
        GRACEFUL[Graceful Shutdown]
    end
    
    subgraph "Server Layer"
        HTTPSERVER[HTTPServer]
        HANDLER[Handler]
        ROUTER[Router]
    end
    
    subgraph "External"
        ENV[Environment Variables]
        SIGNALS[OS Signals]
    end
    
    APP --> CONFIG
    APP --> HTTPSERVER
    APP --> GRACEFUL
    CONFIG --> ENV
    GRACEFUL --> SIGNALS
    HTTPSERVER --> HANDLER
    HTTPSERVER --> ROUTER
    
    style APP fill:#f3e5f5
    style CONFIG fill:#e8f5e8
    style GRACEFUL fill:#fff3e0
    style HTTPSERVER fill:#e3f2fd
    style HANDLER fill:#e1f5fe
    style ROUTER fill:#e1f5fe
    style ENV fill:#fff3e0
    style SIGNALS fill:#fff3e0
```

### Жизненный цикл приложения

```mermaid
stateDiagram-v2
    [*] --> LoadConfig
    LoadConfig --> CreateApp
    CreateApp --> StartServer
    StartServer --> Running
    
    Running --> GracefulShutdown : SIGINT/SIGTERM
    GracefulShutdown --> StopServer
    StopServer --> WaitForRequests
    WaitForRequests --> ShutdownComplete
    ShutdownComplete --> [*]
    
    Running --> Running : Handle Requests
    
    note right of LoadConfig
        • SERVER_PORT env var
        • Default: 8080
    end note
    
    note right of GracefulShutdown
        • Stop accepting new requests
        • Wait for current requests
        • Timeout: 30 seconds
    end note
```

### `Config`
Конфигурация приложения.

```go
type Config struct {
    Port string
}
```

## Основные методы

### `New(config Config) *App`
Создает новое приложение с заданной конфигурацией.

### `Run() error`
Запускает приложение и ожидает сигналы для graceful shutdown.

### `LoadConfig() Config`
Загружает конфигурацию из переменных окружения.

## Переменные окружения

- `SERVER_PORT` - порт для запуска сервера (по умолчанию: "8080")

## Пример использования

```go
package main

import (
    "log"
    "github.com/IgorKilipenko/metrical/internal/app"
)

func main() {
    // Загружаем конфигурацию
    config := app.LoadConfig()
    
    // Создаем приложение
    application := app.New(config)
    
    // Запускаем приложение
    if err := application.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Graceful Shutdown

Приложение корректно обрабатывает сигналы:
- `SIGINT` (Ctrl+C)
- `SIGTERM`

При получении сигнала приложение:
1. Логирует получение сигнала
2. Останавливает прием новых запросов
3. Ждет завершения текущих запросов (до 30 секунд)
4. Корректно завершает работу

## Тестирование

Пакет включает полное покрытие тестами:
- Загрузка конфигурации
- Создание приложения
- Обработка переменных окружения

Запуск тестов:
```bash
go test -v ./internal/app
```

## Преимущества

1. **Разделение ответственности** - логика инициализации отделена от main
2. **Тестируемость** - легко тестировать компоненты изолированно
3. **Конфигурируемость** - гибкая настройка через переменные окружения
4. **Надежность** - graceful shutdown для корректного завершения
5. **Переиспользование** - можно использовать в разных точках входа
