# Router Package

Простой HTTP роутер-обертка над `http.ServeMux`.

## Описание

Пакет `router` предоставляет простую обертку над стандартным `http.ServeMux` Go, которая:
- Упрощает регистрацию маршрутов
- Обеспечивает совместимость с `http.Handler`
- Позволяет легко расширять функциональность в будущем
- Сохраняет производительность стандартного `http.ServeMux`

## Использование

```go
// Создание роутера
r := router.New()

// Регистрация обработчиков
r.HandleFunc("/update/", metricsHandler.UpdateMetric)
r.Handle("/health", healthHandler)

// Использование в HTTP сервере
http.ListenAndServe(":8080", r)
```

## API

### New() *Router

Создает новый экземпляр роутера.

### HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))

Регистрирует функцию-обработчик для указанного пути.

### Handle(pattern string, handler http.Handler)

Регистрирует обработчик для указанного пути.

### ServeHTTP(w http.ResponseWriter, req *http.Request)

Реализует интерфейс `http.Handler`.

### GetMux() *http.ServeMux

Возвращает внутренний `http.ServeMux` для совместимости.

## Преимущества

1. **Простота**: Легкий и понятный API
2. **Совместимость**: Полная совместимость с `http.Handler`
3. **Производительность**: Минимальные накладные расходы
4. **Расширяемость**: Легко добавить middleware и другие функции
5. **Тестируемость**: Удобно для unit-тестирования

## Тестирование

```bash
go test -v ./internal/router
``` 