# HTTPServer Package

Пакет `httpserver` содержит логику HTTP сервера для работы с метриками.

## Описание

Пакет предоставляет структуру `Server`, которая инкапсулирует всю логику HTTP сервера, включая:
- Создание и настройку хранилища метрик
- Инициализацию сервисов и обработчиков
- Настройку маршрутов HTTP
- Запуск сервера

## Использование

```go
// Создание сервера
srv := httpserver.NewServer(":8080")

// Запуск сервера
err := srv.Start()
if err != nil {
    log.Fatal(err)
}
```

## Структуры

### Server

```go
type Server struct {
    addr    string
    handler *handler.MetricsHandler
}
```

- `addr` - адрес для запуска сервера
- `handler` - HTTP обработчик для метрик

## Методы

### NewServer(addr string) *Server

Создает новый экземпляр сервера с указанным адресом.

### Start() error

Запускает HTTP сервер и блокирует выполнение до завершения работы сервера.

### GetMux() *http.ServeMux

Возвращает настроенный ServeMux для использования в тестах.

### ServeHTTP(w http.ResponseWriter, r *http.Request)

Реализует интерфейс `http.Handler`, что позволяет использовать сервер напрямую в тестах.

## Маршруты

- `POST /update/<тип>/<имя>/<значение>` - обновление метрики 