# Сервис сбора метрик и алертинга

Сервер для сбора рантайм-метрик, принимает репорты от агентов по протоколу HTTP.

## Архитектура

Проект следует принципам чистой архитектуры с разделением на слои:

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
        REPO[Repository]
        M[Model/Storage]
    end
    
    H --> S
    R --> H
    S --> REPO
    REPO --> M
    S --> T
    
    style H fill:#e3f2fd
    style R fill:#e3f2fd
    style S fill:#f3e5f5
    style T fill:#f3e5f5
    style REPO fill:#e8f5e8
    style M fill:#e8f5e8
```

### Поток данных

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repository
    participant Storage
    
    Client->>Handler: POST /update/{type}/{name}/{value}
    Handler->>Service: UpdateMetric(type, name, value)
    Service->>Repository: UpdateGauge/UpdateCounter
    Repository->>Storage: UpdateGauge/UpdateCounter
    Storage-->>Repository: Success
    Repository-->>Service: Success
    Service-->>Handler: Success
    Handler-->>Client: 200 OK
    
    Note over Client,Storage: Получение метрики
    Client->>Handler: GET /value/{type}/{name}
    Handler->>Service: GetGauge/GetCounter
    Service->>Repository: GetGauge/GetCounter
    Repository->>Storage: GetGauge/GetCounter
    Storage-->>Repository: Value
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
        REPO[Repository]
        STORAGE[MemStorage]
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
    REPO --> STORAGE
    SERVICE --> TEMPLATE
    
    style AGENT fill:#e1f5fe
    style SERVER fill:#f3e5f5
    style APP fill:#e8f5e8
    style SERVICE fill:#fff3e0
    style STORAGE fill:#e3f2fd
```

- **`cmd/`** - точки входа в приложение (server, agent)
- **`internal/`** - внутренняя логика приложения
  - **`app/`** - основная логика инициализации приложения
  - **`httpserver/`** - HTTP сервер и его логика
  - **`router/`** - роутер (обертка над chi роутером)
  - **`routes/`** - настройка HTTP маршрутов
  - **`model/`** - модели данных и интерфейсы (потокобезопасные)
  - **`service/`** - бизнес-логика
  - **`handler/`** - HTTP обработчики
  - **`template/`** - HTML шаблоны
  - **`agent/`** - агент для сбора метрик
  - **`config/`** - конфигурация
  - **`repository/`** - работа с данными

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
│   │   ├── metrics.go       # Модели данных
│   │   └── metrics_test.go  # Тесты модели
│   ├── agent/
│   │   ├── agent.go         # Логика агента
│   │   ├── config.go        # Конфигурация агента
│   │   ├── metrics.go       # Сбор метрик
│   │   └── *_test.go        # Тесты агента
│   ├── config/              # Конфигурация
│   └── repository/          # Работа с данными
├── migrations/              # Миграции БД
├── pkg/                     # Публичные пакеты
├── go.mod                   # Зависимости
├── go.sum                   # Хеши зависимостей
└── README.md               # Документация проекта
```

## Функциональность

### Поддерживаемые типы метрик

1. **Gauge** (float64) - новое значение замещает предыдущее
2. **Counter** (int64) - новое значение добавляется к предыдущему

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

## Запуск

### Сервер

```bash
go run cmd/server/main.go
```

Сервер запустится на порту 8080.

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

### Типы-алиасы

Для улучшения читаемости кода определены типы-алиасы:

```go
type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64
```

### MemStorage

Потокобезопасное хранилище метрик в памяти с интерфейсом `Storage`:

```go
type Storage interface {
    UpdateGauge(name string, value float64)
    UpdateCounter(name string, value int64)
    GetGauge(name string) (float64, bool)
    GetCounter(name string) (int64, bool)
    GetAllGauges() GaugeMetrics
    GetAllCounters() CounterMetrics
}

type MemStorage struct {
    Gauges   GaugeMetrics
    Counters CounterMetrics
    mu       sync.RWMutex // Потокобезопасность
}
```

**Особенности:**
- **Потокобезопасность** - все операции защищены RWMutex
- **Типизированные метрики** - использование GaugeMetrics и CounterMetrics
- **Безопасное копирование** - GetAllGauges/GetAllCounters возвращают копии

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

## Архитектурные решения

### Разделение ответственности

- **`cmd/server/main.go`** - минимальная точка входа, только создание и запуск сервера
- **`internal/httpserver/`** - инкапсуляция всей логики HTTP сервера
- **`internal/router/`** - абстракция над `chi` роутером для будущей расширяемости
- **`internal/handler/`** - HTTP обработчики, только парсинг запросов и валидация
- **`internal/service/`** - бизнес-логика, работа с метриками
- **`internal/template/`** - HTML шаблоны для отображения метрик
- **`internal/routes/`** - настройка HTTP маршрутов и их группировка
- **`internal/model/`** - модели данных и интерфейсы хранилища

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

- **Абстракция** - скрывает детали работы с источниками данных
- **Тестируемость** - легко создавать моки для тестирования
- **Гибкость** - можно легко заменить реализацию
- **Обработка ошибок** - все методы возвращают ошибки
- **Разделение ответственности** - репозиторий не содержит бизнес-логику

### Преимущества пакета routes

- **Разделение ответственности** - настройка маршрутов отделена от сервера
- **Модульность** - каждый тип маршрутов в отдельной функции
- **Тестируемость** - легко тестировать маршруты изолированно
- **Масштабируемость** - простое добавление новых групп маршрутов
- **Переиспользование** - маршруты можно использовать в разных серверах

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

### Работа с хранилищем метрик

```go
// Создание потокобезопасного хранилища
storage := models.NewMemStorage()

// Обновление метрик
storage.UpdateGauge("temperature", 23.5)
storage.UpdateCounter("requests", 100)

// Получение метрик
value, exists := storage.GetGauge("temperature")
if exists {
    fmt.Printf("Temperature: %.2f\n", value)
}

// Получение всех метрик (типизированные)
allGauges := storage.GetAllGauges()     // GaugeMetrics
allCounters := storage.GetAllCounters() // CounterMetrics

// Работа с типизированными метриками
for name, value := range allGauges {
    fmt.Printf("Gauge %s: %.2f\n", name, value)
}

for name, value := range allCounters {
    fmt.Printf("Counter %s: %d\n", name, value)
}
```

### Работа с репозиторием

```go
// Создание репозитория
storage := models.NewMemStorage()
repo := repository.NewInMemoryMetricsRepository(storage)

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
