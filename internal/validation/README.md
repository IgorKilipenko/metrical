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

### ValidateMetricsBatch
**НОВОЕ**: Валидация батча метрик:

```go
func ValidateMetricsBatch(metrics []models.Metrics) error
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

### Батчевая валидация (НОВОЕ)
```go
func (h *MetricsHandler) UpdateMetricsBatch(w http.ResponseWriter, r *http.Request) {
    var metrics []models.Metrics
    if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
        http.Error(w, "Invalid JSON format", http.StatusBadRequest)
        return
    }

    // Валидируем весь батч
    if err := validation.ValidateMetricsBatch(metrics); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Передаем валидированные данные в сервис
    err := h.service.UpdateMetricsBatch(r.Context(), metrics)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
```

### Примеры батчевой валидации
```go
// Валидный батч
metrics := []models.Metrics{
    {
        ID:    "temperature",
        MType: "gauge",
        Value: func() *float64 { v := 23.5; return &v }(),
    },
    {
        ID:    "requests",
        MType: "counter",
        Delta: func() *int64 { v := int64(100); return &v }(),
    },
}
err := validation.ValidateMetricsBatch(metrics)
// err == nil

// Невалидный батч (пустой)
err = validation.ValidateMetricsBatch([]models.Metrics{})
// err = ValidationError{Field: "metrics", Value: "empty slice", Message: "metrics slice cannot be empty"}

// Невалидный батч (nil значение)
metrics = []models.Metrics{
    {
        ID:    "temperature",
        MType: "gauge",
        Value: nil, // Ошибка!
    },
}
err = validation.ValidateMetricsBatch(metrics)
// err = ValidationError{Field: "value", Value: "nil", Message: "validation error for metric at index 0: value is required for gauge metric"}
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
