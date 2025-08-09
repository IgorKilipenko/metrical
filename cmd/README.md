# cmd

В данной директории содержится код, который скомпилируется в бинарное приложение.

Рекомендуется помещать только код, необходимый для запуска приложения, но не бизнес-логику.

Название директории должно соответствовать названию приложения.

Директория `cmd/app_name` содержит:
- точку входа в приложение (функция `main`)
- инициализацию зависимостей (можно вынести в отдельный пакет `internal/app`)
- настройку и запуск HTTP-сервера (можно вынести в отдельный пакет `internal/router`)
- обработку сигналов завершения работы приложения

## Архитектура команд

```mermaid
graph TB
    subgraph "Command Line Applications"
        SERVER[cmd/server]
        AGENT[cmd/agent]
    end
    
    subgraph "Entry Points"
        SERVER_MAIN[main.go]
        AGENT_MAIN[main.go]
    end
    
    subgraph "Dependencies"
        INTERNAL_APP[internal/app]
        INTERNAL_AGENT[internal/agent]
        INTERNAL_SERVER[internal/httpserver]
    end
    
    subgraph "Binaries"
        SERVER_BIN[server binary]
        AGENT_BIN[agent binary]
    end
    
    SERVER --> SERVER_MAIN
    AGENT --> AGENT_MAIN
    
    SERVER_MAIN --> INTERNAL_APP
    SERVER_MAIN --> INTERNAL_SERVER
    AGENT_MAIN --> INTERNAL_AGENT
    
    SERVER_MAIN --> SERVER_BIN
    AGENT_MAIN --> AGENT_BIN
    
    style SERVER fill:#e3f2fd
    style AGENT fill:#e3f2fd
    style SERVER_MAIN fill:#f3e5f5
    style AGENT_MAIN fill:#f3e5f5
    style INTERNAL_APP fill:#e8f5e8
    style INTERNAL_AGENT fill:#e8f5e8
    style INTERNAL_SERVER fill:#e8f5e8
    style SERVER_BIN fill:#fff3e0
    style AGENT_BIN fill:#fff3e0
```

### Поток компиляции и запуска

```mermaid
sequenceDiagram
    participant Developer
    participant Go
    participant Main
    participant Internal
    participant Binary
    
    Developer->>Go: go build cmd/server
    Go->>Main: Compile main.go
    Main->>Internal: Import dependencies
    Internal-->>Main: Dependencies ready
    Main-->>Go: Compiled
    Go-->>Developer: server binary
    
    Developer->>Binary: ./server
    Binary->>Internal: Initialize app
    Internal-->>Binary: App ready
    Binary-->>Developer: Server running
    
    Note over Main,Internal: Минимальная логика в main
    Note over Internal: Бизнес-логика в internal
```