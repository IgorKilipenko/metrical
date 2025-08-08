# HTTPServer Package

Пакет `httpserver` содержит логику HTTP сервера для работы с метриками.

## Описание

Пакет предоставляет структуру `Server`, которая инкапсулирует всю логику HTTP сервера, включая:
- Создание и настройку хранилища метрик
- Инициализацию сервисов и обработчиков
- Настройку маршрутов HTTP
- Запуск сервера

### Архитектура HTTP сервера

```mermaid
graph TB
    subgraph "HTTPServer Package"
        SERVER[Server]
        HANDLER[MetricsHandler]
        ROUTER[Router]
    end
    
    subgraph "Dependencies"
        SERVICE[Service]
        REPO[Repository]
        STORAGE[MemStorage]
        TEMPLATE[Template]
    end
    
    subgraph "HTTP Layer"
        HTTP_SERVER[HTTP Server]
        MUX[ServeMux]
    end
    
    SERVER --> HANDLER
    SERVER --> ROUTER
    HANDLER --> SERVICE
    SERVICE --> REPO
    REPO --> STORAGE
    SERVICE --> TEMPLATE
    
    HTTP_SERVER --> SERVER
    MUX --> SERVER
    
    style SERVER fill:#f3e5f5
    style HANDLER fill:#e3f2fd
    style ROUTER fill:#e3f2fd
    style SERVICE fill:#e8f5e8
    style REPO fill:#fff3e0
    style STORAGE fill:#e1f5fe
    style TEMPLATE fill:#fff3e0
```

### Жизненный цикл сервера

```mermaid
stateDiagram-v2
    [*] --> NewServer
    NewServer --> Configure
    Configure --> Start
    Start --> Running
    
    Running --> HandleRequest : HTTP Request
    HandleRequest --> Running
    
    Running --> Shutdown : SIGINT/SIGTERM
    Shutdown --> GracefulShutdown
    GracefulShutdown --> Stopped
    Stopped --> [*]
    
    note right of Configure
        • Create Handler
        • Setup Router
        • Initialize Dependencies
    end note
    
    note right of HandleRequest
        • Parse Request
        • Route to Handler
        • Process Business Logic
        • Return Response
    end note
```

## Использование

```go
// Создание сервера с обработкой ошибок
server, err := httpserver.NewServer(":8080")
if err != nil {
    log.Fatalf("Failed to create server: %v", err)
}

// Запуск сервера
if err := server.Start(); err != nil {
    log.Printf("Server error: %v", err)
}

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := server.Shutdown(ctx); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

## Структуры

### Server

```go
type Server struct {
    addr    string
    handler *handler.MetricsHandler
    router  *router.Router
    server  *http.Server
}
```

- `addr` - адрес для запуска сервера
- `handler` - HTTP обработчик для метрик
- `router` - кэшированный роутер
- `server` - ссылка на HTTP сервер для graceful shutdown

## Методы

### NewServer(addr string) (*Server, error)

Создает новый экземпляр сервера с указанным адресом. Возвращает ошибку при пустом адресе.

### Start() error

Запускает HTTP сервер и блокирует выполнение до завершения работы сервера. Корректно обрабатывает ошибки и логирует их.

### Shutdown(ctx context.Context) error

Gracefully останавливает сервер с использованием переданного контекста. Корректно завершает все текущие запросы.



### ServeHTTP(w http.ResponseWriter, r *http.Request)

Реализует интерфейс `http.Handler`, что позволяет использовать сервер напрямую в тестах.

## Маршруты

- `POST /update/<тип>/<имя>/<значение>` - обновление метрики 