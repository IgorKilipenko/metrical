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
- `POST /update/{type}/{name}/{value}` - обновление метрики
- `GET /value/{type}/{name}` - получение значения метрики

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
