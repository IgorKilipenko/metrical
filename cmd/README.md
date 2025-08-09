# cmd

В данной директории содержится код, который скомпилируется в бинарное приложение.

Рекомендуется помещать только код, необходимый для запуска приложения, но не бизнес-логику.

Название директории должно соответствовать названию приложения.

Директория `cmd/app_name` содержит:
- точку входа в приложение (функция `main`)
- инициализацию зависимостей (можно вынести в отдельный пакет `internal/app`)
- настройку и запуск HTTP-сервера (можно вынести в отдельный пакет `internal/router`)
- обработку сигналов завершения работы приложения
- CLI логику и парсинг флагов
- обработку ошибок и валидацию входных данных

## Архитектура команд

```mermaid
graph TB
    subgraph "Command Line Applications"
        SERVER[cmd/server]
        AGENT[cmd/agent]
    end
    
    subgraph "Server Components"
        SERVER_MAIN[main.go]
        SERVER_CLI[cli.go]
        SERVER_UTILS[cliutils.go]
        SERVER_TESTS[main_test.go]
        SERVER_CLI_TESTS[cli_test.go]
        SERVER_UTILS_TESTS[cliutils_test.go]
    end
    
    subgraph "Entry Points"
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
    SERVER --> SERVER_CLI
    SERVER --> SERVER_UTILS
    SERVER --> SERVER_TESTS
    SERVER --> SERVER_CLI_TESTS
    SERVER --> SERVER_UTILS_TESTS
    AGENT --> AGENT_MAIN
    
    SERVER_MAIN --> INTERNAL_APP
    SERVER_MAIN --> INTERNAL_SERVER
    SERVER_CLI --> SERVER_UTILS
    AGENT_MAIN --> INTERNAL_AGENT
    
    SERVER_MAIN --> SERVER_BIN
    AGENT_MAIN --> AGENT_BIN
    
    style SERVER fill:#e3f2fd
    style AGENT fill:#e3f2fd
    style SERVER_MAIN fill:#f3e5f5
    style SERVER_CLI fill:#f3e5f5
    style SERVER_UTILS fill:#f3e5f5
    style SERVER_TESTS fill:#fff3e0
    style SERVER_CLI_TESTS fill:#fff3e0
    style SERVER_UTILS_TESTS fill:#fff3e0
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
    participant CLI
    participant Utils
    participant Internal
    participant Binary
    
    Developer->>Go: go build cmd/server
    Go->>Main: Compile main.go
    Main->>CLI: Import CLI logic
    CLI->>Utils: Import utilities
    Main->>Internal: Import dependencies
    Internal-->>Main: Dependencies ready
    Utils-->>CLI: Utilities ready
    CLI-->>Main: CLI ready
    Main-->>Go: Compiled
    Go-->>Developer: server binary
    
    Developer->>Binary: ./server
    Binary->>CLI: Parse flags
    CLI->>Utils: Validate input
    Utils-->>CLI: Validation result
    CLI-->>Binary: Parsed config
    Binary->>Internal: Initialize app
    Internal-->>Binary: App ready
    Binary-->>Developer: Server running
    
    Note over Main,Internal: Минимальная логика в main
    Note over Internal: Бизнес-логика в internal
    Note over CLI,Utils: CLI логика и валидация
```

## Структура cmd/server

```
cmd/server/
├── main.go          # Точка входа сервера
├── main_test.go     # Тесты main функции
├── cli.go           # CLI логика и парсинг флагов
├── cli_test.go      # Тесты CLI логики
├── cliutils.go      # Утилиты CLI и кастомные ошибки
├── cliutils_test.go # Тесты утилит CLI
└── README.md        # Документация сервера
```

### Компоненты cmd/server

#### main.go
- Точка входа сервера
- Парсинг флагов командной строки
- Создание конфигурации
- Инициализация и запуск приложения
- Централизованная обработка ошибок через `handleError`

#### cli.go
- Настройка Cobra команд
- Парсинг аргументов
- Валидация входных данных
- Обработка help флага с безопасной проверкой на nil
- Кастомный тип ошибки `HelpRequestedError`
- Валидация адреса с кастомным типом ошибки `InvalidAddressError`

#### cliutils.go
- Кастомные типы ошибок
- Функции-предикаты для проверки типов ошибок
- Валидация адреса с детальными сообщениями об ошибках

#### Тесты
- **main_test.go** - тестирование обработки ошибок, кодов выхода, интеграционные тесты
- **cli_test.go** - тестирование парсинга флагов, валидации адреса, обработки ошибок
- **cliutils_test.go** - тестирование кастомных типов ошибок и функций валидации

### Статистика тестирования cmd/server

- **Всего тестов:** 20+ функций тестирования
- **Подтестов:** 60+ сценариев
- **Покрытие:** 81.6% общего покрытия
- **Время выполнения:** ~34ms

### Запуск тестов

```bash
# Все тесты cmd/server
go test ./cmd/server/... -v

# Только тесты main функции
go test ./cmd/server/... -run "TestHandleError|TestMainFunction"

# Покрытие тестами
go test ./cmd/server/... -cover
```