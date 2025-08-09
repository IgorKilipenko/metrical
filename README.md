# Сервис сбора метрик и алертинга

Сервер для сбора рантайм-метрик, принимает репорты от агентов по протоколу HTTP.

## Архитектура

Проект следует принципам чистой архитектуры с разделением на слои и включает современные практики разработки:

### 🏗️ **Ключевые принципы:**
- **Clean Architecture** - четкое разделение слоев
- **Dependency Injection** - инверсия зависимостей
- **Graceful Shutdown** - корректная остановка сервера
- **Error Handling** - детальная обработка ошибок
- **Test-Driven Development** - полное покрытие тестами

### Общая архитектура системы

```mermaid
graph TB
    subgraph "Agent"
        A[Agent Process]
        A --> A1[Сбор runtime метрик]
        A --> A2[Отправка HTTP POST]
    end
    
    subgraph "Server"
        B[HTTP Server]
        B --> B1[Прием метрик]
        B --> B2[Хранение в памяти]
        B --> B3[API endpoints]
    end
    
    A2 -->|HTTP POST| B1
    
    style A fill:#e1f5fe
    style B fill:#f3e5f5
    style A1 fill:#fff3e0
    style A2 fill:#fff3e0
    style B1 fill:#e8f5e8
    style B2 fill:#e8f5e8
    style B3 fill:#e8f5e8
```

### Архитектура сервера (Clean Architecture)

```mermaid
graph TB
    subgraph "Transport Layer"
        H[HTTP Handler]
        R[Router]
    end
    
    subgraph "Business Logic Layer"
        S[Service]
        T[Template]
    end
    
    subgraph "Data Access Layer"
        REPO[Repository Interface]
        IMR[InMemory Repository]
    end
    
    subgraph "Data Models"
        M[Models]
    end
    
    H --> S
    R --> H
    S --> REPO
    REPO --> IMR
    IMR --> M
    S --> T
    
    style H fill:#e3f2fd
    style R fill:#e3f2fd
    style S fill:#f3e5f5
    style T fill:#f3e5f5
    style REPO fill:#e8f5e8
    style IMR fill:#e8f5e8
    style M fill:#fff3e0
```

### Поток данных

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repository
    participant Models
    
    Client->>Handler: POST /update/{type}/{name}/{value}
    Handler->>Service: UpdateMetric(type, name, value)
    Service->>Repository: UpdateGauge/UpdateCounter
    Repository->>Models: UpdateGauge/UpdateCounter
    Models-->>Repository: Success
    Repository-->>Service: Success
    Service-->>Handler: Success
    Handler-->>Client: 200 OK
    
    Note over Client,Models: Получение метрики
    Client->>Handler: GET /value/{type}/{name}
    Handler->>Service: GetGauge/GetCounter
    Service->>Repository: GetGauge/GetCounter
    Repository->>Models: GetGauge/GetCounter
    Models-->>Repository: Value
    Repository-->>Service: Value
    Service-->>Handler: Value
    Handler-->>Client: 200 OK + Value
```

### Полная архитектура системы

```mermaid
graph TB
    subgraph "Agent Process"
        AGENT[Agent]
        COLLECTOR[Metrics Collector]
        SENDER[HTTP Sender]
        AGENT_CONFIG[Agent Config]
    end
    
    subgraph "Server Process"
        APP[App]
        SERVER[HTTPServer]
        ROUTER[Router/Chi]
        HANDLER[Handler]
        SERVICE[Service]
        REPO[Repository Interface]
        IMR[InMemory Repository]
        MODELS[Models]
        TEMPLATE[Template]
    end
    
    subgraph "External"
        RUNTIME[runtime.MemStats]
        ENV[Environment]
        SIGNALS[OS Signals]
    end
    
    AGENT --> COLLECTOR
    AGENT --> SENDER
    AGENT --> AGENT_CONFIG
    COLLECTOR --> RUNTIME
    SENDER --> SERVER
    
    APP --> SERVER
    APP --> ENV
    APP --> SIGNALS
    SERVER --> ROUTER
    ROUTER --> HANDLER
    HANDLER --> SERVICE
    SERVICE --> REPO
    REPO --> IMR
    IMR --> MODELS
    SERVICE --> TEMPLATE
    
    style AGENT fill:#e1f5fe
    style SERVER fill:#f3e5f5
    style APP fill:#e8f5e8
    style SERVICE fill:#fff3e0
    style REPO fill:#e8f5e8
    style IMR fill:#e8f5e8
    style MODELS fill:#fff3e0
```

- **`cmd/`** - точки входа в приложение (server, agent)
- **`internal/`** - внутренняя логика приложения
  - **`app/`** - основная логика инициализации приложения
  - **`httpserver/`** - HTTP сервер и его логика
  - **`router/`** - роутер (обертка над chi роутером)
  - **`routes/`** - настройка HTTP маршрутов
  - **`model/`** - структуры данных (только модели)
  - **`repository/`** - интерфейсы и реализации для работы с данными
  - **`service/`** - бизнес-логика
  - **`handler/`** - HTTP обработчики
  - **`template/`** - HTML шаблоны
  - **`agent/`** - агент для сбора метрик
  - **`config/`** - конфигурация

## Структура проекта

```
go-metrics/
├── cmd/
│   ├── server/
│   │   ├── main.go          # Точка входа сервера
│   │   └── README.md        # Документация сервера
│   └── agent/
│       ├── main.go          # Точка входа агента
│       └── README.md        # Документация агента
├── internal/
│   ├── app/
│   │   ├── app.go           # Основная логика приложения
│   │   ├── config.go        # Конфигурация приложения
│   │   ├── app_test.go      # Тесты приложения
│   │   └── README.md        # Документация пакета app
│   ├── httpserver/
│   │   ├── server.go        # Логика HTTP сервера
│   │   ├── server_test.go   # Тесты сервера
│   │   └── README.md        # Документация пакета httpserver
│   ├── router/
│   │   ├── router.go        # Роутер (обертка над chi роутером)
│   │   ├── router_test.go   # Тесты роутера
│   │   └── README.md        # Документация пакета router
│   ├── handler/
│   │   ├── metrics.go       # HTTP обработчики
│   │   └── metrics_test.go  # Тесты обработчиков
│   ├── service/
│   │   ├── metrics.go       # Бизнес-логика
│   │   └── metrics_test.go  # Тесты сервиса
│   ├── template/
│   │   ├── metrics.go       # HTML шаблоны
│   │   ├── metrics_test.go  # Тесты шаблонов
│   │   └── README.md        # Документация пакета template
│   ├── routes/
│   │   ├── metrics.go       # Настройка HTTP маршрутов
│   │   ├── metrics_test.go  # Тесты маршрутов
│   │   └── README.md        # Документация пакета routes
│   ├── model/
│   │   ├── metrics.go       # Структуры данных
│   │   └── README.md        # Документация пакета model
│   ├── repository/
│   │   ├── metrics.go       # Интерфейс Repository
│   │   ├── memory.go        # InMemory реализация
│   │   ├── memory_test.go   # Тесты Repository
│   │   └── README.md        # Документация пакета repository
│   ├── agent/
│   │   ├── agent.go         # Логика агента
│   │   ├── config.go        # Конфигурация агента
│   │   ├── metrics.go       # Сбор метрик
│   │   └── *_test.go        # Тесты агента
│   └── config/              # Конфигурация
├── migrations/              # Миграции БД
├── pkg/                     # Публичные пакеты
├── go.mod                   # Зависимости
├── go.sum                   # Хеши зависимостей
└── README.md               # Документация проекта
```

## 🚀 Функциональность

### Поддерживаемые типы метрик

1. **Gauge** (float64) - новое значение замещает предыдущее
2. **Counter** (int64) - новое значение добавляется к предыдущему

### 🔧 **Последние улучшения**

#### Graceful Shutdown
- ✅ Корректная остановка HTTP сервера
- ✅ Обработка сигналов SIGINT/SIGTERM
- ✅ Таймаут 30 секунд для завершения запросов
- ✅ Логирование процесса остановки

### HTTP API

Сервер доступен по адресу `http://localhost:8080`

#### Обновление метрики

```
POST /update/{ТИП_МЕТРИКИ}/{ИМЯ_МЕТРИКИ}/{ЗНАЧЕНИЕ_МЕТРИКИ}
Content-Type: text/plain
```

**Примеры:**
```bash
# Gauge метрика
curl -X POST "http://localhost:8080/update/gauge/temperature/23.5" \
     -H "Content-Type: text/plain"

# Counter метрика
curl -X POST "http://localhost:8080/update/counter/requests/100" \
     -H "Content-Type: text/plain"
```

#### Получение значения метрики

```
GET /value/{ТИП_МЕТРИКИ}/{ИМЯ_МЕТРИКИ}
Content-Type: text/plain
```

**Примеры:**
```bash
# Получить значение gauge метрики
curl "http://localhost:8080/value/gauge/temperature"
# Ответ: 23.5

# Получить значение counter метрики
curl "http://localhost:8080/value/counter/requests"
# Ответ: 100
```

#### Просмотр всех метрик

```
GET /
Content-Type: text/html
```

**Пример:**
```bash
# Открыть в браузере или получить HTML
curl "http://localhost:8080/"
```

Возвращает HTML-страницу со списком всех метрик, сгруппированных по типам.

#### Коды ответов

- `200 OK` - запрос выполнен успешно
- `400 Bad Request` - некорректный тип метрики или значение
- `404 Not Found` - метрика не найдена или отсутствует имя метрики
- `405 Method Not Allowed` - неподдерживаемый HTTP метод

### Агент для сбора метрик

Агент автоматически собирает метрики из пакета `runtime` и отправляет их на сервер:

#### Собираемые метрики

**Gauge метрики из runtime:**
- Alloc, BuckHashSys, Frees, GCCPUFraction, GCSys
- HeapAlloc, HeapIdle, HeapInuse, HeapObjects, HeapReleased
- HeapSys, LastGC, Lookups, MCacheInuse, MCacheSys
- MSpanInuse, MSpanSys, Mallocs, NextGC, NumForcedGC
- NumGC, OtherSys, PauseTotalNs, StackInuse, StackSys
- Sys, TotalAlloc

**Дополнительные метрики:**
- RandomValue (gauge) - случайное значение
- PollCount (counter) - счетчик обновлений

#### Конфигурация агента

- **PollInterval**: 2 секунды - частота сбора метрик
- **ReportInterval**: 10 секунд - частота отправки метрик
- **ServerURL**: http://localhost:8080

## Зависимости

### Внешние пакеты

- **`github.com/go-chi/chi/v5`** - HTTP роутер для маршрутизации запросов
- **`github.com/stretchr/testify/assert`** - библиотека для тестирования (в некоторых тестах)

### Стандартные пакеты

- **`sync`** - потокобезопасность (RWMutex)
- **`net/http`** - HTTP сервер и клиент
- **`text/template`** - HTML шаблоны

## 🚀 Запуск

### Сервер

```bash
# Запуск с портом по умолчанию
go run cmd/server/main.go

# Запуск с кастомным портом
SERVER_PORT=9090 go run cmd/server/main.go
```

Сервер запустится на порту 8080 (или указанном в переменной `SERVER_PORT`).

#### Graceful Shutdown

Сервер корректно обрабатывает сигналы завершения:

```bash
# Остановка Ctrl+C
^C
2025/08/08 09:11:02 Received signal: terminated
2025/08/08 09:11:02 Shutting down server gracefully...
2025/08/08 09:11:02 Server shutdown complete
```

### Агент

```bash
go run cmd/agent/main.go
```

Агент начнет собирать и отправлять метрики автоматически.

## Тестирование

### Запуск всех тестов

```bash
go test ./...
```

### Запуск тестов по пакетам

```bash
# Тесты приложения
go test ./internal/app/... -v

# Тесты HTTP сервера
go test ./internal/httpserver/... -v

# Тесты роутера
go test ./internal/router/... -v

# Тесты хендлеров
go test ./internal/handler/... -v

# Тесты сервиса
go test ./internal/service/... -v

# Тесты репозитория
go test ./internal/repository/... -v

# Тесты модели
go test ./internal/model/... -v

# Тесты агента
go test ./internal/agent/... -v

# Тесты шаблонов
go test ./internal/template/... -v

# Тесты маршрутов
go test ./internal/routes/... -v
```

### Покрытие тестами

Проект покрыт юнит-тестами для всех основных компонентов:

- ✅ **Приложение** - тестирование инициализации и конфигурации
- ✅ **HTTP сервер** - интеграционные тесты сервера
- ✅ **Роутер** - тестирование маршрутизации
- ✅ **HTTP хендлеры** - тестирование API endpoints
- ✅ **Сервисный слой** - тестирование бизнес-логики
- ✅ **Репозиторий** - тестирование работы с данными
- ✅ **Модели данных** - тестирование структур и интерфейсов (включая потокобезопасность)
- ✅ **Агент** - тестирование сбора метрик
- ✅ **Шаблоны** - тестирование генерации HTML
- ✅ **Маршруты** - тестирование настройки HTTP endpoints
- ✅ **Валидация** - тестирование обработки ошибок
- ✅ **Потокобезопасность** - тестирование конкурентного доступа

## Структура данных

### Модели (internal/model)

Структуры данных для метрик:

```go
// Константы типов метрик
const (
    Counter = "counter"
    Gauge   = "gauge"
)

// Типы-алиасы
type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

// Структура метрики
type Metrics struct {
    ID    string   `json:"id"`
    MType string   `json:"type"`
    Delta *int64   `json:"delta,omitempty"`
    Value *float64 `json:"value,omitempty"`
    Hash  string   `json:"hash,omitempty"`
}
```

### Репозиторий (internal/repository)

Интерфейс для работы с данными:

```go
type MetricsRepository interface {
    UpdateGauge(name string, value float64) error
    UpdateCounter(name string, value int64) error
    GetGauge(name string) (float64, bool, error)
    GetCounter(name string) (int64, bool, error)
    GetAllGauges() (models.GaugeMetrics, error)
    GetAllCounters() (models.CounterMetrics, error)
}
```

**InMemoryRepository** - потокобезопасная реализация в памяти:

```go
type InMemoryMetricsRepository struct {
    Gauges   models.GaugeMetrics
    Counters models.CounterMetrics
    mu       sync.RWMutex
}
```

**Особенности:**
- **Потокобезопасность** - все операции защищены RWMutex
- **Абстракция данных** - сервис не зависит от конкретной реализации
- **Легкое тестирование** - можно мокать интерфейс
- **Расширяемость** - легко добавить PostgreSQL, Redis и т.д.

### Metrics

Структура для представления метрики:

```go
type Metrics struct {
    ID    string   `json:"id"`
    MType string   `json:"type"`
    Delta *int64   `json:"delta,omitempty"`
    Value *float64 `json:"value,omitempty"`
    Hash  string   `json:"hash,omitempty"`
}
```

### MetricsData

Структура для передачи данных в HTML шаблон:

```go
type MetricsData struct {
    Gauges       models.GaugeMetrics  // Gauge метрики
    Counters     models.CounterMetrics // Counter метрики
    GaugeCount   int                  // Количество gauge метрик
    CounterCount int                  // Количество counter метрик
}
```

### Routes

Функции для настройки HTTP маршрутов:

```go
// Настройка маршрутов метрик
func SetupMetricsRoutes(handler *handler.MetricsHandler) *chi.Mux

// Настройка маршрутов health check
func SetupHealthRoutes() *chi.Mux
```

**Настраиваемые маршруты:**
- `GET /` - отображение всех метрик (HTML)
- `POST /update/{type}/{name}/{value}` - обновление метрики
- `GET /value/{type}/{name}` - получение значения метрики
- `GET /health` - проверка состояния сервиса

## 🏗️ Архитектурные решения

### Разделение ответственности

- **`cmd/server/main.go`** - минимальная точка входа, только создание и запуск сервера
- **`internal/httpserver/`** - инкапсуляция всей логики HTTP сервера с graceful shutdown
- **`internal/router/`** - абстракция над `chi` роутером для будущей расширяемости
- **`internal/handler/`** - HTTP обработчики, только парсинг запросов и валидация
- **`internal/service/`** - бизнес-логика, работа с метриками
- **`internal/template/`** - HTML шаблоны для отображения метрик
- **`internal/routes/`** - настройка HTTP маршрутов и их группировка
- **`internal/model/`** - структуры данных (только модели)
- **`internal/repository/`** - абстракция над источниками данных

### 🔧 **Последние архитектурные улучшения**

#### Graceful Shutdown
- ✅ Корректная остановка HTTP сервера через `context.Context`
- ✅ Обработка сигналов SIGINT/SIGTERM в `internal/app`
- ✅ Таймаут 30 секунд для завершения текущих запросов
- ✅ Логирование процесса остановки

### Принципы

- **Чистая архитектура** - разделение на слои с четкими границами
- **Dependency Injection** - зависимости передаются через конструкторы
- **Interface Segregation** - интерфейсы разделены по назначению
- **Single Responsibility** - каждый пакет отвечает за одну область

### Преимущества chi роутера

- **Параметризованные маршруты** - поддержка URL параметров `{type}`, `{name}`, `{value}`
- **Высокая производительность** - быстрая маршрутизация запросов
- **Гибкость** - легко добавлять middleware и расширять функциональность
- **Совместимость** - полная совместимость с `net/http`
- **Читаемость** - понятные и выразительные маршруты

### Преимущества типов-алиасов

- **Улучшенная читаемость** - явно понятно назначение типов (`GaugeMetrics`, `CounterMetrics`)
- **Типобезопасность** - компилятор различает типы метрик
- **Самодокументируемость** - код становится более понятным
- **Расширяемость** - легко добавить методы к типам в будущем
- **Консистентность** - единообразное использование типов во всем проекте

### Преимущества пакета app

- **Разделение ответственности** - логика инициализации отделена от main
- **Тестируемость** - легко тестировать компоненты изолированно
- **Конфигурируемость** - гибкая настройка через переменные окружения
- **Надежность** - graceful shutdown для корректного завершения
- **Переиспользование** - можно использовать в разных точках входа

### Преимущества пакета app

- **Разделение ответственности** - логика инициализации отделена от main
- **Тестируемость** - легко тестировать компоненты изолированно
- **Конфигурируемость** - гибкая настройка через переменные окружения
- **Надежность** - graceful shutdown для корректного завершения
- **Переиспользование** - можно использовать в разных точках входа

### Преимущества пакета repository

- **Абстракция данных** - скрывает детали работы с источниками данных
- **Тестируемость** - легко создавать моки для тестирования
- **Гибкость** - можно легко заменить реализацию (PostgreSQL, Redis)
- **Обработка ошибок** - все методы возвращают ошибки
- **Разделение ответственности** - репозиторий не содержит бизнес-логику
- **Потокобезопасность** - встроенная защита от гонки данных
- **Классический паттерн** - стандартный подход в Go

### Преимущества пакета model

- **Чистота архитектуры** - только структуры данных, никакой бизнес-логики
- **Простота** - минимальный набор полей и методов
- **Независимость** - нет зависимостей от других пакетов
- **Переиспользование** - модели используются во всех слоях приложения
- **Типобезопасность** - строгая типизация для метрик

### Преимущества пакета routes

- **Разделение ответственности** - настройка маршрутов отделена от сервера
- **Модульность** - каждый тип маршрутов в отдельной функции
- **Тестируемость** - легко тестировать маршруты изолированно
- **Масштабируемость** - простое добавление новых групп маршрутов
- **Переиспользование** - маршруты можно использовать в разных серверах

## 📋 Changelog

### v2.0.0 (Текущая версия) - Рефакторинг архитектуры
- 🏗️ **Repository Pattern** - внедрен классический паттерн Repository
- 🧹 **Очистка моделей** - убраны Storage и MemStorage из models
- 📦 **Новый пакет repository** - интерфейсы и реализации для работы с данными
- 🔧 **Упрощение архитектуры** - четкое разделение ответственности
- 📚 **Обновленная документация** - отражена новая архитектура
- 🧪 **Обновленные тесты** - все тесты адаптированы под новую архитектуру

### v1.0.0-rc2
- ✨ **Graceful Shutdown** - корректная остановка сервера
- ✨ **Улучшенная обработка ошибок** - валидация и детальное логирование
- 🧹 **Очистка API** - удален устаревший метод `GetMux()`
- 📚 **Обновленная документация** - Mermaid диаграммы и актуальные примеры
- 🧪 **Улучшенные тесты** - полное покрытие новых функций

### v1.0.0-rc.1
- 🎉 **Первоначальный релиз** - базовая функциональность
- 📊 **Сбор метрик** - gauge и counter типы
- 🌐 **HTTP API** - RESTful endpoints
- 🤖 **Агент** - автоматический сбор runtime метрик
- 🏗️ **Clean Architecture** - разделение на слои

## Отладка

Настроена конфигурация VS Code для отладки:

- **Debug Server** - отладка сервера
- **Debug Agent** - отладка агента

## Пример работы

### Запуск сервера

```bash
go run cmd/server/main.go
```

Сервер запустится на порту 8080 и будет доступен по адресу `http://localhost:8080`.

### Тестирование API

#### 1. Обновление метрик

```bash
# Добавить gauge метрику
curl -X POST "http://localhost:8080/update/gauge/temperature/23.5"

# Добавить counter метрику
curl -X POST "http://localhost:8080/update/counter/requests/100"
```

#### 2. Получение значений метрик

```bash
# Получить значение gauge метрики
curl "http://localhost:8080/value/gauge/temperature"
# Ответ: 23.5

# Получить значение counter метрики
curl "http://localhost:8080/value/counter/requests"
# Ответ: 100
```

#### 3. Просмотр всех метрик

```bash
# Открыть в браузере
open http://localhost:8080/

# Или получить HTML через curl
curl "http://localhost:8080/"
```

### Запуск агента

```bash
go run cmd/agent/main.go
```

Агент будет автоматически:
- Собирать 29 метрик из runtime каждые 2 секунды
- Отправлять их на сервер каждые 10 секунд
- Логировать все операции

### Полный цикл работы

1. Запустите сервер: `go run cmd/server/main.go`
2. Запустите агент: `go run cmd/agent/main.go`
3. Откройте браузер: `http://localhost:8080/`
4. Наблюдайте, как метрики обновляются в реальном времени

Все запросы возвращают статус 200 OK при успешном выполнении.

## Примеры использования

### Инициализация приложения

```go
package main

import (
    "log"
    "github.com/IgorKilipenko/metrical/internal/app"
)

func main() {
    // Загружаем конфигурацию из переменных окружения
    config := app.LoadConfig()
    
    // Создаем приложение
    application := app.New(config)
    
    // Запускаем приложение с graceful shutdown
    if err := application.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### Создание HTTP сервера

```go
// Создание сервера с валидацией
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

### Работа с моделями

```go
// Создание метрики
metric := models.Metrics{
    ID:    "temperature",
    MType: models.Gauge,
    Value: &value,
}

// Работа с типами-алиасами
gauges := models.GaugeMetrics{"temp": 23.5}
counters := models.CounterMetrics{"requests": 100}

// Проверка типов метрик
if metric.MType == models.Gauge {
    // обработка gauge метрики
}
```

### Работа с репозиторием

```go
// Создание репозитория
repo := repository.NewInMemoryMetricsRepository()

// Обновление метрик
err := repo.UpdateGauge("temperature", 23.5)
err = repo.UpdateCounter("requests", 100)

// Получение метрик
value, exists, err := repo.GetGauge("temperature")
value, exists, err := repo.GetCounter("requests")

// Получение всех метрик
gauges, err := repo.GetAllGauges()
counters, err := repo.GetAllCounters()
```

### Настройка маршрутов

```go
// Создание хендлера
handler := handler.NewMetricsHandler(service)

// Настройка маршрутов через пакет routes
router := routes.SetupMetricsRoutes(handler)

// Добавление health check маршрутов
healthRouter := routes.SetupHealthRoutes()
```
