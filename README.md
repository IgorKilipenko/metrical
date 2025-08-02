# Сервис сбора метрик и алертинга

Сервер для сбора рантайм-метрик, принимает репорты от агентов по протоколу HTTP.

## Архитектура

Проект следует принципам чистой архитектуры с разделением на слои:

- **`cmd/`** - точки входа в приложение (server, agent)
- **`internal/`** - внутренняя логика приложения
  - **`model/`** - модели данных и интерфейсы
  - **`service/`** - бизнес-логика
  - **`handler/`** - HTTP обработчики
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

## Запуск

### Сервер

```bash
go run cmd/server/main.go
```

Сервер запустится на порту 8080.

### Агент (для тестирования)

```bash
go run cmd/agent/main.go
```

Агент отправит тестовые метрики на сервер.

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
- **Debug Tests** - отладка тестов

## Тестирование

Проект включает тестовый агент, который демонстрирует работу с API:

1. Отправляет gauge метрики (temperature, humidity, pressure)
2. Отправляет counter метрики (requests, errors, connections)
3. Демонстрирует накопление counter метрик

Все запросы возвращают статус 200 OK при успешном выполнении.
