# Сервис сбора метрик и алертинга

Сервер для сбора рантайм-метрик, принимает репорты от агентов по протоколу HTTP.

## Архитектура

Проект следует принципам чистой архитектуры с разделением на слои:
```
┌─────────────┐    HTTP POST    ┌─────────────┐
│    Agent    │ ──────────────► │   Server    │
│             │                 │             │
│ • Сбор      │                 │ • Прием     │
│   runtime   │                 │   метрик    │
│   метрик    │                 │ • Хранение  │
│ • Отправка  │                 │ • API       │
└─────────────┘                 └─────────────┘
```

- **`cmd/`** - точки входа в приложение (server, agent)
- **`internal/`** - внутренняя логика приложения
  - **`httpserver/`** - HTTP сервер и его логика
  - **`router/`** - роутер (обертка над chi роутером)
  - **`model/`** - модели данных и интерфейсы
  - **`service/`** - бизнес-логика
  - **`handler/`** - HTTP обработчики
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
│   ├── httpserver/
│   │   ├── server.go        # Логика HTTP сервера
│   │   ├── server_test.go   # Тесты сервера
│   │   └── README.md        # Документация пакета httpserver
│   ├── router/
│   │   ├── router.go        # Роутер (обертка над http.ServeMux)
│   │   ├── router_test.go   # Тесты роутера
│   │   └── README.md        # Документация пакета router
│   ├── handler/
│   │   ├── metrics.go       # HTTP обработчики
│   │   └── metrics_test.go  # Тесты обработчиков
│   ├── service/
│   │   ├── metrics.go       # Бизнес-логика
│   │   └── metrics_test.go  # Тесты сервиса
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
# Тесты HTTP сервера
go test ./internal/httpserver/... -v

# Тесты роутера
go test ./internal/router/... -v

# Тесты хендлеров
go test ./internal/handler/... -v

# Тесты сервиса
go test ./internal/service/... -v

# Тесты модели
go test ./internal/model/... -v

# Тесты агента
go test ./internal/agent/... -v
```

### Покрытие тестами

Проект покрыт юнит-тестами для всех основных компонентов:

- ✅ **HTTP сервер** - интеграционные тесты сервера
- ✅ **Роутер** - тестирование маршрутизации
- ✅ **HTTP хендлеры** - тестирование API endpoints
- ✅ **Сервисный слой** - тестирование бизнес-логики
- ✅ **Модели данных** - тестирование структур и интерфейсов
- ✅ **Агент** - тестирование сбора метрик
- ✅ **Валидация** - тестирование обработки ошибок
- ✅ **Потокобезопасность** - тестирование конкурентного доступа

## Структура данных

### MemStorage

Хранилище метрик в памяти с интерфейсом `Storage`:

```go
type Storage interface {
    UpdateGauge(name string, value float64)
    UpdateCounter(name string, value int64)
    GetGauge(name string) (float64, bool)
    GetCounter(name string) (int64, bool)
    GetAllGauges() map[string]float64
    GetAllCounters() map[string]int64
}
```

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

## Архитектурные решения

### Разделение ответственности

- **`cmd/server/main.go`** - минимальная точка входа, только создание и запуск сервера
- **`internal/httpserver/`** - инкапсуляция всей логики HTTP сервера
- **`internal/router/`** - абстракция над `chi` роутером для будущей расширяемости
- **`internal/handler/`** - HTTP обработчики, только парсинг запросов и валидация
- **`internal/service/`** - бизнес-логика, работа с метриками
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
