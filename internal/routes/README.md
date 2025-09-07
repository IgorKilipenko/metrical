# Routes Package

Пакет для настройки HTTP маршрутов приложения.

## Описание

Пакет `routes` предоставляет функции для настройки HTTP маршрутов. Основная цель - отделить логику настройки маршрутов от сервера и сделать код более модульным и тестируемым.

## Компоненты

### SetupMetricsRoutes

Функция для настройки маршрутов метрик:

```go
func SetupMetricsRoutes(handler *handler.MetricsHandler) *chi.Mux
```

Настраивает следующие маршруты:
- `GET /` - отображение всех метрик (HTML)
- `POST /update/{type}/{name}/{value}` - обновление метрики (legacy)
- `GET /value/{type}/{name}` - получение значения метрики (legacy)
- `POST /update` - обновление метрики через JSON API
- `POST /value` - получение метрики через JSON API

### Архитектура маршрутов

```mermaid
graph TB
    subgraph "Routes Package"
        MR[SetupMetricsRoutes]
        HR[SetupHealthRoutes]
        CHI[Chi Router]
    end
    
    subgraph "HTTP Endpoints"
        GET_ALL[GET /]
        POST_UPDATE[POST /update/{type}/{name}/{value}]
        GET_VALUE[GET /value/{type}/{name}]
        GET_HEALTH[GET /health]
    end
    
    subgraph "Handlers"
        MH[MetricsHandler]
        HH[HealthHandler]
    end
    
    MR --> CHI
    HR --> CHI
    CHI --> GET_ALL
    CHI --> POST_UPDATE
    CHI --> GET_VALUE
    CHI --> GET_HEALTH
    
    GET_ALL --> MH
    POST_UPDATE --> MH
    GET_VALUE --> MH
    GET_HEALTH --> HH
    
    style MR fill:#f3e5f5
    style HR fill:#f3e5f5
    style CHI fill:#e8f5e8
    style GET_ALL fill:#e3f2fd
    style POST_UPDATE fill:#e3f2fd
    style GET_VALUE fill:#e3f2fd
    style GET_HEALTH fill:#e3f2fd
    style MH fill:#fff3e0
    style HH fill:#fff3e0
```

### Структура маршрутов

```mermaid
graph LR
    subgraph "Metrics Routes"
        R1[GET /]
        R2[POST /update/{type}/{name}/{value}]
        R3[GET /value/{type}/{name}]
    end
    
    subgraph "Health Routes"
        R4[GET /health]
    end
    
    subgraph "Handler Mapping"
        H1[MetricsHandler.GetAllMetrics]
        H2[MetricsHandler.UpdateMetric]
        H3[MetricsHandler.GetMetricValue]
        H4[HealthHandler.Health]
    end
    
    R1 --> H1
    R2 --> H2
    R3 --> H3
    R4 --> H4
    
    style R1 fill:#e3f2fd
    style R2 fill:#e3f2fd
    style R3 fill:#e3f2fd
    style R4 fill:#e8f5e8
    style H1 fill:#fff3e0
    style H2 fill:#fff3e0
    style H3 fill:#fff3e0
    style H4 fill:#fff3e0
```

### SetupHealthRoutes

Функция для настройки маршрутов health check:

```go
func SetupHealthRoutes() *chi.Mux
```

Настраивает маршрут:
- `GET /health` - проверка состояния сервиса

## Использование

### В httpserver

```go
// internal/httpserver/server.go
func (s *Server) createRouter() *router.Router {
    // Используем отдельный пакет для настройки маршрутов
    chiRouter := routes.SetupMetricsRoutes(s.handler)
    return router.NewWithChiRouter(chiRouter)
}
```

### Добавление новых маршрутов

```go
// internal/routes/admin.go
func SetupAdminRoutes(adminHandler *handler.AdminHandler) *chi.Mux {
    r := chi.NewRouter()
    
    r.Get("/admin/users", adminHandler.GetUsers)
    r.Post("/admin/users", adminHandler.CreateUser)
    
    return r
}
```

## Преимущества

1. **Разделение ответственности** - настройка маршрутов отделена от сервера
2. **Модульность** - каждый тип маршрутов в отдельной функции
3. **Тестируемость** - легко тестировать маршруты изолированно
4. **Масштабируемость** - простое добавление новых групп маршрутов
5. **Переиспользование** - маршруты можно использовать в разных серверах

## Архитектурные принципы

- **Single Responsibility** - каждая функция отвечает за одну группу маршрутов
- **Dependency Injection** - хендлеры передаются как параметры
- **Composition** - маршруты можно комбинировать
- **Testability** - каждый маршрут можно тестировать отдельно

## Тестирование

```bash
go test -v ./internal/routes
```

## Расширение

Для добавления новых групп маршрутов:

1. Создайте новую функцию в пакете `routes`
2. Настройте маршруты с помощью chi
3. Добавьте тесты
4. Обновите документацию
