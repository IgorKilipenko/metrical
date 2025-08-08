# internal

В этой директории размещается код внутренних модулей приложения. Код внутри этого пакета недоступен для импорта в других приложениях.

Структуру дирктории `internal/` можно разбивать по логическим блокам приложения, выделяя пакеты по функциональному назначению. 
Например, `internal/agent`, `internal/server` и т.д.

Директория `internal/` является специальной в Go и обеспечивает инкапсуляцию кода на уровне модуля. Компилятор Go запрещает импорт пакетов из `internal/` за пределами родительского модуля.

## Архитектура внутренних пакетов

```mermaid
graph TB
    subgraph "Application Layer"
        APP[internal/app]
        CONFIG[internal/config]
    end
    
    subgraph "Server Layer"
        HTTPSERVER[internal/httpserver]
        ROUTER[internal/router]
        ROUTES[internal/routes]
        HANDLER[internal/handler]
    end
    
    subgraph "Business Layer"
        SERVICE[internal/service]
        AGENT[internal/agent]
    end
    
    subgraph "Data Layer"
        REPO[internal/repository]
        MODEL[internal/model]
        TEMPLATE[internal/template]
    end
    
    APP --> CONFIG
    APP --> HTTPSERVER
    
    HTTPSERVER --> ROUTER
    HTTPSERVER --> HANDLER
    ROUTER --> ROUTES
    
    HANDLER --> SERVICE
    SERVICE --> REPO
    REPO --> MODEL
    SERVICE --> TEMPLATE
    
    AGENT --> MODEL
    
    style APP fill:#e3f2fd
    style CONFIG fill:#e3f2fd
    style HTTPSERVER fill:#f3e5f5
    style ROUTER fill:#f3e5f5
    style ROUTES fill:#f3e5f5
    style HANDLER fill:#f3e5f5
    style SERVICE fill:#e8f5e8
    style AGENT fill:#e8f5e8
    style REPO fill:#fff3e0
    style MODEL fill:#fff3e0
    style TEMPLATE fill:#fff3e0
```

### Иерархия зависимостей

```mermaid
graph TD
    subgraph "Public Interface"
        CMD[cmd/*]
    end
    
    subgraph "Internal Packages"
        INTERNAL[internal/*]
    end
    
    subgraph "External Dependencies"
        PKG[pkg/*]
        VENDOR[vendor/*]
    end
    
    CMD --> INTERNAL
    INTERNAL --> PKG
    INTERNAL --> VENDOR
    
    style CMD fill:#e3f2fd
    style INTERNAL fill:#f3e5f5
    style PKG fill:#e8f5e8
    style VENDOR fill:#fff3e0
```

### Принципы организации

- **Инкапсуляция** - код недоступен для внешних импортов
- **Модульность** - каждый пакет отвечает за свою область
- **Зависимости** - четкая иерархия зависимостей
- **Тестируемость** - каждый пакет можно тестировать изолированно
