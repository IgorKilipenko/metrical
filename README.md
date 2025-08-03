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
  - **`model/`** - модели данных и интерфейсы
  - **`service/`** - бизнес-логика
  - **`handler/`** - HTTP обработчики
  - **`agent/`** - агент для сбора метрик
  - **`config/`** - конфигурация
  - **`repository/`** - работа с данными

## Функциональность

### Поддерживаемые типы метрик

1. **Gauge** (float64) - новое значение замещает предыдущее
2. **Counter** (int64) - новое значение добавляется к предыдущему

### HTTP API

Сервер доступен по адресу `http://localhost:8080`

#### Обновление метрики

```
POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
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

#### Коды ответов

- `200 OK` - метрика успешно обновлена
- `400 Bad Request` - некорректный тип метрики или значение
- `404 Not Found` - отсутствует имя метрики
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
# Тесты хендлеров
go test ./internal/handler/... -v

# Тесты сервиса
go test ./internal/service/... -v

# Тесты агента
go test ./internal/agent/... -v
```

### Покрытие тестами

Проект покрыт юнит-тестами для всех основных компонентов:

- ✅ **HTTP хендлеры** - тестирование API endpoints
- ✅ **Сервисный слой** - тестирование бизнес-логики
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

## Отладка

Настроена конфигурация VS Code для отладки:

- **Debug Server** - отладка сервера
- **Debug Agent** - отладка агента

## Пример работы

1. Запустите сервер: `go run cmd/server/main.go`
2. Запустите агент: `go run cmd/agent/main.go`
3. Агент будет автоматически:
   - Собирать 29 метрик из runtime каждые 2 секунды
   - Отправлять их на сервер каждые 10 секунд
   - Логировать все операции

Все запросы возвращают статус 200 OK при успешном выполнении.
