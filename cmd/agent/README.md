# cmd/agent

В данной директории содержится код **Агента для сбора метрик**, который компилируется в бинарное приложение.

## 📊 Функциональность

Агент автоматически собирает runtime метрики из Go приложения и отправляет их на сервер по HTTP.

### 🔄 Основные возможности

- **Сбор метрик** - автоматический сбор 29+ метрик из `runtime.MemStats`
- **Периодическая отправка** - отправка метрик на сервер каждые 10 секунд
- **Graceful shutdown** - корректная остановка при получении сигналов
- **Потокобезопасность** - использование `sync.RWMutex` для конкурентного доступа
- **Структурированное логирование** - детальное логирование через logger абстракцию

### 📈 Собираемые метрики

#### Runtime метрики (Gauge):
- `Alloc` - текущее использование памяти
- `HeapAlloc` - использование heap памяти
- `HeapIdle` - свободная heap память
- `HeapInuse` - используемая heap память
- `NumGC` - количество сборок мусора
- `GCCPUFraction` - доля CPU времени на GC
- `PauseTotalNs` - общее время пауз GC
- `Mallocs` - количество аллокаций
- `Frees` - количество освобождений
- `TotalAlloc` - общий объем аллокаций
- И еще 17 метрик из `runtime.MemStats`

#### Дополнительные метрики:
- `RandomValue` (gauge) - случайное значение от 0 до 1
- `PollCount` (counter) - счетчик обновлений метрик

## ⚙️ Конфигурация

```go
type Config struct {
    ServerURL      string        // URL сервера (по умолчанию: http://localhost:8080)
    PollInterval   time.Duration // Интервал сбора метрик (по умолчанию: 2s)
    ReportInterval time.Duration // Интервал отправки метрик (по умолчанию: 10s)
}
```

## 🚀 Запуск

```bash
# Запуск агента
go run cmd/agent/main.go

# Компиляция в бинарный файл
go build -o agent ./cmd/agent/

# Компиляция с версией
go build -ldflags "-X main.Version=1.0.0" -o agent ./cmd/agent/

# Запуск
./agent
```

### 📦 Версионирование

Приложение поддерживает версионирование через ldflags:

```bash
# Установка версии при сборке
go build -ldflags "-X main.Version=1.0.0" -o agent ./cmd/agent/

# Проверка версии
./agent --help
# Выведет: Starting metrics agent v1.0.0
```

### 📋 Флаги командной строки

```bash
# Показать справку
./agent --help

# Запуск с кастомными настройками
./agent -a http://example.com:9090 -p 5 -r 15 -v

# Только verbose логирование
./agent -v
```

| Флаг | Описание | По умолчанию |
|------|----------|--------------|
| `-a, --a` | HTTP server endpoint address | `http://localhost:8080` |
| `-p, --p` | Poll interval in seconds | `2` |
| `-r, --r` | Report interval in seconds | `10` |
| `-v, --v` | Enable verbose logging | `false` |
| `-h, --help` | Show help | - |

## 🛑 Graceful Shutdown

Агент корректно обрабатывает сигналы завершения:

```bash
# Остановка Ctrl+C
^C
2025/08/03 09:12:40 Received signal: terminated
2025/08/03 09:12:40 Stopping agent...
2025/08/03 09:12:40 Agent stopped gracefully
2025/08/03 09:12:40 Polling stopped
2025/08/03 09:12:40 Reporting stopped
2025/08/03 09:12:41 Agent shutdown completed
```

## 🧪 Тестирование

```bash
# Тесты агента
go test ./internal/agent/... -v

# Запуск с таймаутом для проверки работы
timeout 10s go run cmd/agent/main.go
```

## 📁 Структура кода

- `main.go` - точка входа, настройка graceful shutdown
- `internal/agent/agent.go` - основная логика агента
- `internal/agent/metrics.go` - структуры и функции для работы с метриками
- `internal/agent/agent_test.go` - unit тесты

## 🔄 Жизненный цикл

```mermaid
stateDiagram-v2
    [*] --> Инициализация
    Инициализация --> ЗапускГорутин
    ЗапускГорутин --> СборМетрик
    ЗапускГорутин --> ОтправкаМетрик
    
    СборМетрик --> СборМетрик : каждые 2s
    ОтправкаМетрик --> ОтправкаМетрик : каждые 10s
    
    СборМетрик --> GracefulShutdown : SIGINT/SIGTERM
    ОтправкаМетрик --> GracefulShutdown : SIGINT/SIGTERM
    
    GracefulShutdown --> ОстановкаСбора
    GracefulShutdown --> ОстановкаОтправки
    
    ОстановкаСбора --> Завершение
    ОстановкаОтправки --> Завершение
    Завершение --> [*]
    
    note right of СборМетрик
        • runtime.MemStats
        • 29+ метрик
        • Потокобезопасность
    end note
    
    note right of ОтправкаМетрик
        • HTTP POST
        • JSON формат
        • Retry логика
    end note
```

1. **Инициализация** - создание агента с конфигурацией
2. **Запуск горутин** - параллельный сбор и отправка метрик
3. **Сбор метрик** - каждые 2 секунды из `runtime.MemStats`
4. **Отправка метрик** - каждые 10 секунд на сервер
5. **Graceful shutdown** - корректная остановка при сигналах

## 📊 Пример логов

```
2025/08/03 09:12:30 Starting metrics agent v1.0.0
2025/08/03 09:12:30 Agent configuration: server=http://localhost:8080, poll=2s, report=10s, verbose=false
2025/08/03 09:12:30 INFO collected metrics total=29
2025/08/03 09:12:32 INFO collected metrics total=29
2025/08/03 09:12:40 INFO successfully sent metrics count=29
2025/08/03 09:12:40 Received signal: terminated
2025/08/03 09:12:40 INFO stopping agent
2025/08/03 09:12:40 INFO polling stopped
2025/08/03 09:12:40 INFO reporting stopped
2025/08/03 09:12:40 INFO agent stopped gracefully
```

## 📝 Логирование

Агент использует структурированное логирование для отслеживания операций:

### Основные события
- **Запуск**: Конфигурация и инициализация
- **Сбор метрик**: Количество собранных метрик
- **Отправка метрик**: Статистика отправки
- **Ошибки**: Детальная информация при verbose режиме
- **Graceful shutdown**: Процесс остановки

### Verbose режим
При включении флага `-v` логируются дополнительные детали:
```
INFO collected metrics total=29 gauges=28 counters=1
DEBUG sent metric successfully name=Alloc value=1048576 status=200
ERROR error sending metric name=HeapIdle error="connection refused"
WARN sent metrics with errors successful=25 failed=4
```
