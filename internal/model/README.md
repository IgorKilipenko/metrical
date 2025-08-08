# internal/model

В этом пакете содержатся структуры данных, которые описывают основные сущности предметной области приложения.

Эти структуры используются в сервисах и хэндлерах. Данный пакет не должен содержать бизнес-логику приложения.

## Структуры данных

### Типы-алиасы

Для улучшения читаемости кода определены типы-алиасы:

```go
type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64
```

### Storage Interface

Интерфейс для работы с хранилищем метрик:

```go
type Storage interface {
    UpdateGauge(name string, value float64)
    UpdateCounter(name string, value int64)
    GetGauge(name string) (float64, bool)
    GetCounter(name string) (int64, bool)
    GetAllGauges() GaugeMetrics
    GetAllCounters() CounterMetrics
}
```

### MemStorage

Реализация хранилища метрик в памяти:

```go
type MemStorage struct {
    Gauges   GaugeMetrics
    Counters CounterMetrics
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
