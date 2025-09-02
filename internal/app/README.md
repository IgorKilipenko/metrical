# Пакет `app`

Пакет `app` предоставляет основную логику инициализации и запуска приложения метрик.

## Назначение

Пакет инкапсулирует всю логику запуска приложения, включая:
- Создание конфигурации из строки адреса
- Создание и запуск HTTP сервера
- Graceful shutdown при получении сигналов
- Обработку ошибок и логирование

## Компоненты

### `App`
Основная структура приложения, которая управляет жизненным циклом сервера.

```go
type App struct {
    server *httpserver.Server
    addr   string
}
```

### `Config`
Конфигурация приложения.

```go
type Config struct {
    Addr            string // Адрес сервера (например, "localhost")
    Port            string // Порт сервера (например, "8080")
    StoreInterval   int    // Интервал сохранения метрик в секундах
    FileStoragePath string // Путь к файлу для сохранения метрик
    Restore         bool   // Загружать ли метрики при старте
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
        CLI[CLI Arguments]
        SIGNALS[OS Signals]
    end
    
    CLI --> CONFIG
    APP --> CONFIG
    APP --> HTTPSERVER
    APP --> GRACEFUL
    GRACEFUL --> SIGNALS
    HTTPSERVER --> HANDLER
    HTTPSERVER --> ROUTER
    
    style APP fill:#f3e5f5
    style CONFIG fill:#e8f5e8
    style GRACEFUL fill:#fff3e0
    style HTTPSERVER fill:#e3f2fd
    style HANDLER fill:#e1f5fe
    style ROUTER fill:#e1f5fe
    style CLI fill:#fff3e0
    style SIGNALS fill:#fff3e0
```

### Жизненный цикл приложения

```mermaid
stateDiagram-v2
    [*] --> ParseAddress
    ParseAddress --> CreateConfig
    CreateConfig --> CreateApp
    CreateApp --> StartServer
    StartServer --> Running
    
    Running --> GracefulShutdown : SIGINT/SIGTERM
    GracefulShutdown --> StopServer
    StopServer --> WaitForRequests
    WaitForRequests --> ShutdownComplete
    ShutdownComplete --> [*]
    
    Running --> Running : Handle Requests
    
    note right of ParseAddress
        • Парсинг адреса из строки
        • Поддержка форматов:
        • localhost:8080
        • 9090 (localhost:9090)
        • 127.0.0.1:9090
    end note
    
    note right of GracefulShutdown
        • Stop accepting new requests
        • Wait for current requests
        • Timeout: 30 seconds
    end note
```

## Основные методы

### `NewConfig(addr string, storeInterval int, fileStoragePath string, restore bool) (Config, error)`
Создает конфигурацию из строки адреса и параметров персистентности. Поддерживает различные форматы адреса:
- `localhost:8080` - полный адрес
- `9090` - только порт (хост по умолчанию: localhost)
- `127.0.0.1:9090` - IP адрес с портом

**Параметры персистентности:**
- `storeInterval` - интервал сохранения метрик в секундах (0 для синхронного сохранения)
- `fileStoragePath` - путь к файлу для сохранения метрик
- `restore` - загружать ли метрики при старте

### `New(config Config) *App`
Создает новое приложение с заданной конфигурацией.

### `Run() error`
Запускает приложение и ожидает сигналы для graceful shutdown.

## Пример использования

```go
package main

import (
    "log"
    "github.com/IgorKilipenko/metrical/internal/app"
    "github.com/IgorKilipenko/metrical/internal/logger"
)

func main() {
    // Создаем логгер
    appLogger := logger.NewSlogLogger()
    
    // Создаем конфигурацию из строки адреса и параметров персистентности
    config, err := app.NewConfig("localhost:8080", 300, "/tmp/metrics.json", true)
    if err != nil {
        log.Fatal(err)
    }
    
    // Создаем приложение
    application := app.New(config)
    
    // Запускаем приложение
    if err := application.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Логирование

Приложение интегрировано с системой логирования через Dependency Injection:

```go
// Создание логгера и внедрение во все слои
appLogger := logger.NewSlogLogger()

// Создание зависимостей с логгером
repository := repository.NewInMemoryMetricsRepository(appLogger)
service := service.NewMetricsService(repository, appLogger)
handler := handler.NewMetricsHandler(service, appLogger)
server := httpserver.NewServer(addr, handler, appLogger)

// Логирование событий приложения
// Логи: "Starting metrics server on localhost:8080"
// Логи: "Server shutdown completed successfully"
```

### Уровни логирования

- **Info**: События жизненного цикла приложения (запуск, остановка)
- **Error**: Ошибки инициализации и выполнения

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
- Создание конфигурации из различных форматов адреса
- Парсинг адреса и порта
- Создание приложения
- Получение адреса приложения

Запуск тестов:
```bash
go test -v ./internal/app
```

## Преимущества

1. **Разделение ответственности** - логика инициализации отделена от CLI
2. **Тестируемость** - легко тестировать компоненты изолированно
3. **Конфигурируемость** - гибкая настройка через строку адреса
4. **Надежность** - graceful shutdown для корректного завершения
5. **Переиспользование** - можно использовать в разных точках входа
6. **Чистая архитектура** - CLI логика находится в транспортном слое
