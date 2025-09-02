# Middleware Package

Пакет содержит middleware компоненты для HTTP сервера.

## Доступные Middleware

- **LoggingMiddleware** - логирование HTTP запросов и ответов
- **GzipMiddleware** - поддержка gzip сжатия и распаковки

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

## Gzip Middleware

`GzipMiddleware` обеспечивает поддержку gzip сжатия и распаковки для HTTP запросов и ответов.

### Функциональность

- **Сжатие ответов**: Автоматически сжимает ответы сервера, если клиент поддерживает gzip (заголовок `Accept-Encoding: gzip`)
- **Распаковка запросов**: Автоматически распаковывает входящие запросы с заголовком `Content-Encoding: gzip`
- **Умная фильтрация**: Сжимает только поддерживаемые типы контента (JSON, HTML, plain text)

### Поддерживаемые типы контента для сжатия

- `application/json`
- `text/html`
- `text/plain`

### HTTP заголовки

**Входящие запросы:**
- `Content-Encoding: gzip` - указывает, что тело запроса сжато

**Исходящие ответы:**
- `Content-Encoding: gzip` - указывает, что тело ответа сжато (устанавливается автоматически)
- `Accept-Encoding: gzip` - клиент указывает поддержку gzip

### Использование

```go
import "github.com/IgorKilipenko/metrical/internal/middleware"

// Создание middleware
gzipMiddleware := middleware.GzipMiddleware()

// Применение к роутеру
router.Use(gzipMiddleware)
```

### Примеры использования

```go
// В роутере
r := chi.NewRouter()
r.Use(middleware.GzipMiddleware())
r.Use(middleware.LoggingMiddleware())

// Обработчик автоматически получит распакованные данные
r.Post("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
    // Тело запроса уже распаковано, если было сжато
    body, _ := io.ReadAll(r.Body)
    
    // Ответ автоматически сжимается, если клиент поддерживает gzip
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"status": "ok"}`))
})
```

### Особенности реализации

- Middleware автоматически определяет, нужно ли сжимать ответ на основе заголовка `Accept-Encoding`
- Для входящих запросов с gzip автоматически распаковывает тело и обновляет `Content-Length`
- Использует `gzipResponseWriter` для прозрачного сжатия ответов
- Корректно обрабатывает ошибки сжатия/распаковки

### Производительность

- Сжатие происходит "на лету" без буферизации всего ответа в память
- Поддерживает потоковую обработку больших ответов
- Минимальные накладные расходы для несжатых запросов/ответов
