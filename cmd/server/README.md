# cmd/server

HTTP сервер для приема метрик от агента.

## Описание

Сервер принимает POST запросы для обновления метрик в формате:
```
POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
```

Сервер использует пакет `internal/httpserver` для управления HTTP сервером и маршрутизацией.

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
└── README.md        # Документация сервера

internal/
├── httpserver/
│   ├── server.go      # Логика HTTP сервера
│   ├── server_test.go # Тесты сервера
│   └── README.md      # Документация пакета httpserver
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

### Тесты сервера (`internal/httpserver/server_test.go`)
- `TestNewServer` - тестирование создания сервера
- `TestServerGetMux` - тестирование получения ServeMux
- `TestServerIntegration` - интеграционные тесты HTTP сервера
- `TestServerEndToEnd` - end-to-end тестирование полного цикла работы с метриками
- `TestServerBasicFunctionality` - тестирование базовой функциональности
- `TestServerRedirects` - тестирование автоматических редиректов Go HTTP сервера

**Назначение:** Проверяют работу сервера, включая интеграцию всех компонентов и имитацию реального использования.

### Unit тесты обработчиков (`internal/handler/metrics_test.go`)
- `TestMetricsHandler_UpdateMetric` - unit тестирование HTTP обработчика
- `TestMetricsHandler_UpdateMetric_CounterAccumulation` - unit тестирование накопления счетчиков
- `TestMetricsHandler_UpdateMetric_GaugeReplacement` - unit тестирование замены gauge метрик

**Назначение:** Быстрые unit тесты для изолированного тестирования логики HTTP обработчика.

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

### Тесты приложения (`internal/app/app_test.go`)
- `TestLoadConfig` - тестирование загрузки конфигурации
- `TestNew` - тестирование создания приложения
- `TestApp_GetPort` - тестирование получения порта
- `TestGetEnv` - тестирование работы с переменными окружения

**Назначение:** Тестирование логики инициализации приложения и конфигурации.

## Запуск тестов

```bash
# Все тесты сервера
go test ./internal/httpserver/... ./internal/handler/... ./internal/service/... ./internal/model/... ./internal/app/... -v

# Только тесты сервера
go test ./internal/httpserver/... -v

# Только unit тесты обработчиков
go test ./internal/handler/... -v

# Только тесты сервиса
go test ./internal/service/... -v

# Только тесты модели
go test ./internal/model/... -v

# Только тесты приложения
go test ./internal/app/... -v
```

## Разница между типами тестов

### Unit тесты (`internal/handler/metrics_test.go`)
- **Цель:** Тестирование изолированной логики HTTP обработчика
- **Скорость:** Быстрые (миллисекунды)
- **Зависимости:** Минимальные (только handler + service + storage)
- **Использование:** При разработке и рефакторинге handler

### Тесты сервера (`internal/httpserver/server_test.go`)
- **Цель:** Тестирование полной интеграции всех компонентов сервера
- **Скорость:** Средние (несколько секунд)
- **Зависимости:** Все компоненты сервера
- **Использование:** При проверке end-to-end сценариев и регрессий

## Запуск сервера

```bash
go run cmd/server/main.go
```

Сервер запустится на порту 8080.

### Архитектура

Файл `main.go` содержит минимальную логику инициализации:
- Создание экземпляра сервера через `httpserver.NewServer(":8080")`
- Запуск сервера через `srv.Start()`

Вся остальная логика HTTP сервера инкапсулирована в пакете `internal/httpserver`.
