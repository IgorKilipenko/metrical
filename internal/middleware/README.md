# Middleware Package

Пакет содержит middleware компоненты для HTTP сервера.

## Logging Middleware

`LoggingMiddleware` предоставляет функциональность для логирования HTTP запросов и ответов с использованием библиотеки `github.com/rs/zerolog`.

### Функциональность

Middleware логирует следующие сведения:

**О запросах:**
- URI запроса
- HTTP метод
- Время начала обработки
- IP адрес клиента
- User-Agent

**Об ответах:**
- Код статуса HTTP
- Размер содержимого ответа
- Время, затраченное на выполнение запроса

### Использование

```go
import "github.com/IgorKilipenko/metrical/internal/middleware"

// Создание middleware
loggingMiddleware := middleware.LoggingMiddleware()

// Применение к роутеру
router.Use(loggingMiddleware)
```

### Пример логов

```
{"level":"info","method":"GET","uri":"/value/counter/test","remote_addr":"127.0.0.1:12345","user_agent":"curl/7.68.0","time":"2024-01-15T10:30:00Z","message":"HTTP request started"}
{"level":"info","method":"GET","uri":"/value/counter/test","status_code":200,"response_size":15,"duration":0.001234,"time":"2024-01-15T10:30:00Z","message":"HTTP request completed"}
```
