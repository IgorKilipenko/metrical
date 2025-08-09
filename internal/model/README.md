# internal/model

В этом пакете содержатся структуры данных, которые описывают основные сущности предметной области приложения.

Эти структуры используются в сервисах и хэндлерах. Данный пакет не должен содержать бизнес-логику приложения.

## Структуры

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

## Использование

```go
// Создание метрики
metric := models.Metrics{
    ID:    "temperature",
    MType: models.Gauge,
    Value: &value,
}

// Работа с типами
gauges := models.GaugeMetrics{"temp": 23.5}
counters := models.CounterMetrics{"requests": 100}
```
