# internal/validation

Пакет для валидации и парсинга данных метрик. Проверяет корректность входных данных перед передачей в бизнес-логику.

## Назначение

- Валидация типа метрики (gauge/counter)
- Валидация имени метрики (непустое)
- Парсинг значений в соответствующие типы данных
- Возврат типизированных структур или ошибок валидации

## Основные функции

### ValidateMetricRequest
Основная функция валидации и парсинга запроса:

```go
func ValidateMetricRequest(metricType, name, value string) (*MetricRequest, error)
```

**Возвращает:**
- `*MetricRequest` - типизированная структура с валидированными данными
- `error` - ошибка валидации при некорректных данных

### MetricRequest
Структура для валидированного запроса:

```go
type MetricRequest struct {
    Type  string // "gauge" или "counter"
    Name  string // имя метрики
    Value any    // float64 для gauge, int64 для counter
}
```

## Примеры использования

### Валидные запросы
```go
// Gauge метрика
req, err := ValidateMetricRequest("gauge", "temperature", "23.5")
// req.Value = 23.5 (float64)

// Counter метрика  
req, err := ValidateMetricRequest("counter", "requests", "100")
// req.Value = 100 (int64)
```

### Обработка ошибок
```go
req, err := ValidateMetricRequest("gauge", "temp", "abc")
if err != nil {
    // err = ValidationError{Field: "value", Value: "abc", Message: "must be a valid float number"}
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

### В HTTP обработчике
```go
func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
    metricType := chi.URLParam(r, "type")
    metricName := chi.URLParam(r, "name")
    metricValue := chi.URLParam(r, "value")

    metricReq, err := validation.ValidateMetricRequest(metricType, metricName, metricValue)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Передаем валидированные данные в сервис
    err = h.service.UpdateMetric(metricReq)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
```

## Тестирование

```bash
# Запуск тестов
go test ./internal/validation/... -v

# Покрытие тестами (100%)
go test ./internal/validation/... -cover
```

## Ошибки валидации

Пакет возвращает структурированные ошибки `ValidationError`:

```go
type ValidationError struct {
    Field   string // поле с ошибкой
    Value   string // некорректное значение  
    Message string // описание ошибки
}
```

**Примеры ошибок:**
- `"validation error for field 'type' with value 'unknown': must be 'gauge' or 'counter'"`
- `"validation error for field 'name' with value '': cannot be empty"`
- `"validation error for field 'value' with value 'abc': must be a valid float number"`
