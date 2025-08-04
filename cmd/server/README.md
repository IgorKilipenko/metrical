# cmd/server

HTTP сервер для приема метрик от агента.

## Описание

Сервер принимает POST запросы для обновления метрик в формате:
```
POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
```

### Поддерживаемые типы метрик:
- `gauge` - метрики с плавающей точкой (заменяют предыдущее значение)
- `counter` - счетчики (накапливают значения)

### Примеры запросов:
```bash
# Обновление gauge метрики
curl -X POST http://localhost:8080/update/gauge/memory_usage/85.7

# Обновление counter метрики
curl -X POST http://localhost:8080/update/counter/request_count/1
```

## Структура проекта

```
cmd/server/
├── main.go          # Точка входа сервера
└── main_test.go     # Интеграционные тесты сервера

internal/
├── handler/
│   ├── metrics.go      # HTTP обработчики
│   └── metrics_test.go # Тесты обработчиков
├── service/
│   ├── metrics.go      # Бизнес-логика
│   └── metrics_test.go # Тесты сервиса
└── model/
    ├── metrics.go      # Модели данных
    └── metrics_test.go # Тесты модели
```

## Тесты

### Интеграционные тесты сервера (`cmd/server/main_test.go`)
- `TestServerFullIntegration` - полная интеграция всех компонентов сервера
- `TestServerEndToEndFlow` - end-to-end тестирование полного цикла работы с метриками
- `TestServerEdgeCases` - тестирование граничных случаев в полной интеграции
- `TestServerConcurrentRequests` - тестирование конкурентных запросов в полной интеграции

**Назначение:** Проверяют работу всех компонентов вместе, имитируя реальное использование сервера.

### Unit тесты обработчиков (`internal/handler/metrics_test.go`)
- `TestMetricsHandler_UpdateMetric` - unit тестирование HTTP обработчика
- `TestMetricsHandler_UpdateMetric_CounterAccumulation` - unit тестирование накопления счетчиков
- `TestMetricsHandler_UpdateMetric_GaugeReplacement` - unit тестирование замены gauge метрик
- `TestSplitPath` - unit тестирование функции разбора URL путей

**Назначение:** Быстрые unit тесты для изолированного тестирования логики HTTP обработчика и вспомогательных функций.

### Тесты сервиса (`internal/service/metrics_test.go`)
- `TestMetricsService_UpdateMetric` - тестирование обновления метрик
- `TestMetricsService_UpdateMetric_GaugeReplacement` - тестирование замены gauge
- `TestMetricsService_UpdateMetric_CounterAccumulation` - тестирование накопления counter
- `TestMetricsService_GetGauge` - тестирование получения gauge метрик
- `TestMetricsService_GetCounter` - тестирование получения counter метрик
- `TestMetricsService_GetAllGauges` - тестирование получения всех gauge метрик
- `TestMetricsService_GetAllCounters` - тестирование получения всех counter метрик

### Тесты модели (`internal/model/metrics_test.go`)
- `TestNewMemStorage` - тестирование создания хранилища
- `TestMemStorage_UpdateGauge` - тестирование обновления gauge метрик
- `TestMemStorage_UpdateCounter` - тестирование обновления counter метрик
- `TestMemStorage_GetGauge` - тестирование получения gauge метрик
- `TestMemStorage_GetCounter` - тестирование получения counter метрик
- `TestMemStorage_GetAllGauges` - тестирование получения всех gauge метрик
- `TestMemStorage_GetAllCounters` - тестирование получения всех counter метрик
- `TestMemStorage_Isolation` - тестирование изоляции хранилищ
- `TestMemStorage_EdgeCases` - тестирование граничных случаев

## Запуск тестов

```bash
# Все тесты сервера
go test ./cmd/server/... ./internal/handler/... ./internal/service/... ./internal/model/... -v

# Только интеграционные тесты сервера
go test ./cmd/server/... -v

# Только unit тесты обработчиков
go test ./internal/handler/... -v

# Только тесты сервиса
go test ./internal/service/... -v

# Только тесты модели
go test ./internal/model/... -v
```

## Разница между типами тестов

### Unit тесты (`internal/handler/metrics_test.go`)
- **Цель:** Тестирование изолированной логики HTTP обработчика
- **Скорость:** Быстрые (миллисекунды)
- **Зависимости:** Минимальные (только handler + service + storage)
- **Использование:** При разработке и рефакторинге handler

### Интеграционные тесты (`cmd/server/main_test.go`)
- **Цель:** Тестирование полной интеграции всех компонентов
- **Скорость:** Средние (несколько секунд)
- **Зависимости:** Все компоненты сервера
- **Использование:** При проверке end-to-end сценариев и регрессий

## Запуск сервера

```bash
go run cmd/server/main.go
```

Сервер запустится на порту 8080.
