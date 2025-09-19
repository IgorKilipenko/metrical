# Сервис сбора метрик и алертинга

Сервер для сбора рантайм-метрик, принимает репорты от агентов по протоколу HTTP.

## 🎯 Примеры использования

### Быстрый старт с Docker
```bash
# 1. Запуск всей системы
make run-full

# 2. Проверка статуса
make status

# 3. Тестирование API
curl http://localhost:9090/ping

# 4. Остановка системы
make stop-all
```

### Разработка с тестами
```bash
# 1. Запуск тестов с Docker
make test-db

# 2. Быстрые unit тесты
make quick-test

# 3. Проверка покрытия
go test ./... -v -cover
```

### Работа с базой данных
```bash
# 1. Запуск основной БД
make db-up

# 2. Подключение к БД
make db-shell

# 3. Просмотр логов
make db-logs

# 4. Остановка БД
make db-down
```

## 🔄 Retry логика

Система использует интеллектуальную retry логику для обработки временных ошибок:

### Конфигурация
- **Количество попыток**: 4 (1 основная + 3 повтора)
- **Интервалы**: 1s, 3s, 5s (экспоненциальный backoff)
- **Поддержка контекста**: корректная обработка отмены и таймаутов

### Retryable ошибки

#### HTTP/Network ошибки
- `connection refused`
- `connection reset`
- `connection timeout`
- `no such host`
- `network is unreachable`
- `temporary failure`
- `i/o timeout`
- `context deadline exceeded`
- `connection lost`
- `broken pipe`
- HTTP 5xx ошибки сервера

#### PostgreSQL connection errors (Class 08)
- `ConnectionException` (08000)
- `ConnectionDoesNotExist` (08003)
- `ConnectionFailure` (08006)
- `SQLClientUnableToEstablishSQLConnection` (08001)
- `SQLServerRejectedEstablishmentOfSQLConnection` (08004)
- `TransactionResolutionUnknown` (08007)
- `ProtocolViolation` (08P01)

### Не-retryable ошибки
- HTTP 4xx клиентские ошибки
- Ошибки валидации данных
- Ошибки бизнес-логики

### Применение
- **Agent**: HTTP запросы к серверу
- **PostgreSQL Repository**: операции с базой данных
- **Service Layer**: делегирует retry логику в репозиторий

## 🔧 Конфигурация портов

| Сервис | Порт | Описание |
|--------|------|----------|
| **Сервер** | 9090 | HTTP API сервера |
| **Основная БД** | 5434 | PostgreSQL для продакшена |
| **Тестовая БД** | 5433 | PostgreSQL для тестов |

### Переменные окружения
```bash
# Основная база данных
export DATABASE_DSN="postgres://metricaldb:Secret@localhost:5434/metricaldb?sslmode=disable"

# Тестовая база данных
export TEST_DATABASE_URL="postgres://test:test@localhost:5433/testdb?sslmode=disable"
```

## 🚀 Быстрый старт

### 🐳 Docker инфраструктура (Рекомендуется)

Проект теперь включает полную Docker инфраструктуру для разработки и тестирования:

```bash
# Запуск всей системы (БД + сервер)
make run-full

# Проверка статуса сервисов
make status

# Остановка всех сервисов
make stop-all

# Управление базой данных
make db-up          # Запустить БД
make db-down        # Остановить БД
make db-logs        # Просмотр логов БД
make db-shell       # Подключение к БД
```

### VS Code задачи

```bash
# Сборка проекта
Ctrl+Shift+B

# Запуск всех тестов
Ctrl+Shift+P → "Tasks: Run Task" → "Full Test Suite"

# Запуск сервера
Ctrl+Shift+P → "Tasks: Run Task" → "Run Server"

# Запуск агента
Ctrl+Shift+P → "Tasks: Run Task" → "Run Agent"
```

📖 **Подробная документация:** [.vscode/README.md](.vscode/README.md)

### Ручной запуск

```bash
# Сборка
go build -o cmd/server/server cmd/server/main.go cmd/server/cli.go cmd/server/cliutils.go
go build -o cmd/agent/agent cmd/agent/main.go cmd/agent/cli.go

# Запуск сервера
./cmd/server/server -a=localhost:9090

# Запуск с персистентностью
./bin/server -a=localhost:9090 -i 300 -f /tmp/metrics.json -r true

# Запуск агента
./cmd/agent/agent -a=localhost:9090 -r=2s
```

## API Endpoints

### Legacy Endpoints (для обратной совместимости)

- `POST /update/{type}/{name}/{value}` - обновление метрики
- `GET /value/{type}/{name}` - получение значения метрики
- `GET /` - HTML дашборд со всеми метриками

### JSON API Endpoints

#### Обновление метрики
```http
POST /update
Content-Type: application/json

{
  "id": "LastGC",
  "type": "gauge",
  "value": 1744184459
}
```

#### Получение метрики
```http
POST /value
Content-Type: application/json

{
  "id": "LastGC",
  "type": "gauge"
}
```

Ответ:
```json
{
  "id": "LastGC",
  "type": "gauge",
  "value": 1744184459
}
```

### Структура метрики

```go
type Metrics struct {
    ID    string   `json:"id"`              // имя метрики
    MType string   `json:"type"`            // тип: "gauge" или "counter"
    Delta *int64   `json:"delta,omitempty"` // значение для counter метрик
    Value *float64 `json:"value,omitempty"` // значение для gauge метрик
}
```

## Архитектура

Проект следует принципам чистой архитектуры с разделением на слои:

### 🏗️ **Ключевые принципы:**
- **Clean Architecture** - четкое разделение слоев
- **Dependency Injection** - инверсия зависимостей
- **Validation Layer** - отдельный слой валидации данных
- **Logger Abstraction** - абстракция логирования через все слои
- **Persistence Layer** - слой персистентности метрик (файлы + PostgreSQL)
- **Error Handling** - детальная обработка ошибок
- **Test-Driven Development** - полное покрытие тестами
- **Gzip Middleware** - автоматическое сжатие/распаковка HTTP данных
- **Database Abstraction** - интерфейсы для легкого переключения между хранилищами
- **Connection Pooling** - эффективное управление соединениями с БД
- **Retry Logic** - интеллектуальные повторы при сбоях с экспоненциальным backoff

### Архитектура

```mermaid
graph TB
    subgraph "Transport Layer"
        H[HTTP Handler]
        R[Router]
    end
    
    subgraph "Validation Layer"
        V[Validation Package]
    end
    
    subgraph "Business Logic Layer"
        S[Service]
        T[Template]
    end
    
    subgraph "Data Access Layer"
        REPO[Repository Interface]
        IMR[InMemory Repository]
        PGR[PostgreSQL Repository]
    end
    
    subgraph "Database Layer"
        PG[(PostgreSQL)]
        FS[(File System)]
    end
    
    subgraph "Cross-Cutting Concerns"
        L[Logger Abstraction]
        CP[Connection Pool]
        RT[Retry Logic]
    end
    
    H --> V
    R --> H
    V --> S
    S --> REPO
    REPO --> IMR
    REPO --> PGR
    IMR --> FS
    PGR --> PG
    PGR --> CP
    PGR --> RT
    S --> T
    
    H -.-> L
    S -.-> L
    REPO -.-> L
    PGR -.-> L
    
    style H fill:#e3f2fd
    style R fill:#e3f2fd
    style V fill:#e8f5e8
    style S fill:#f3e5f5
    style T fill:#f3e5f5
    style REPO fill:#e8f5e8
    style IMR fill:#e8f5e8
    style PGR fill:#e8f5e8
    style PG fill:#ffebee
    style FS fill:#ffebee
    style L fill:#fff3e0
    style CP fill:#fff3e0
    style RT fill:#fff3e0
```

## Структура проекта

```
go-metrics/
├── cmd/
│   ├── server/             # Сервер приложения
│   └── agent/              # Агент сбора метрик
├── internal/
│   ├── app/                # Основная логика приложения
│   ├── httpserver/         # HTTP сервер
│   ├── router/             # Роутер
│   ├── handler/            # HTTP обработчики
│   ├── service/            # Бизнес-логика
│   ├── validation/         # Валидация данных
│   ├── template/           # HTML шаблоны
│   ├── routes/             # HTTP маршруты
│   ├── model/              # Структуры данных
│   ├── repository/         # Работа с данными (InMemory + PostgreSQL)
│   ├── logger/             # Абстракция логирования
│   ├── middleware/         # Middleware (gzip, logging)
│   ├── testutils/          # Утилиты для тестирования
│   ├── agent/              # Логика агента (с gzip поддержкой)
│   └── config/             # Конфигурация (включая DB настройки)
│       └── db/             # Конфигурация базы данных
├── migrations/             # Миграции БД
├── pkg/                    # Публичные пакеты
└── README.md              # Документация проекта
```

## 💾 Персистентность метрик

Сервер поддерживает сохранение метрик на диск с настраиваемыми параметрами:

### Конфигурация

```bash
# Запуск с настройками персистентности
./bin/server \
  -i 300 \                    # Интервал сохранения в секундах (0 = синхронно)
  -f /tmp/metrics.json \      # Путь к файлу для сохранения
  -r true                     # Загружать метрики при старте
```

### Переменные окружения

```bash
export STORE_INTERVAL=300        # Интервал сохранения (секунды)
export FILE_STORAGE_PATH="/tmp/metrics.json"  # Путь к файлу
export RESTORE=true              # Восстановление при старте
```

### Приоритет конфигурации

1. **Переменные окружения** (высший приоритет)
2. **CLI флаги**
3. **Значения по умолчанию**

### Формат файла

```json
[
  {"id":"LastGC","type":"gauge","value":1257894000000000000},
  {"id":"NumGC","type":"counter","delta":42}
]
```

## 🗄️ PostgreSQL поддержка

Сервер поддерживает хранение метрик в PostgreSQL базе данных с **автоматическими миграциями** и полной функциональностью:

### 🆕 Автоматические миграции

Сервис теперь **автоматически создает все необходимые таблицы** при запуске:

```bash
# Запуск сервера с PostgreSQL - таблицы создаются автоматически!
./cmd/server/server \
  -a=localhost:9090 \
  -d="postgres://user:password@localhost:5432/metrics_db?sslmode=disable"
```

**Логи автоматических миграций:**
```
{"level":"info","message":"Creating PostgreSQL repository with migrations"}
{"level":"info","message":"Loading migrations from filesystem"}
{"level":"info","count":1,"message":"Loaded migrations"}
{"level":"info","message":"Running migrations"}
{"level":"info","message":"Initializing migrations table"}
{"level":"info","message":"Migration applied successfully"}
{"level":"info","message":"All migrations completed successfully"}
```

### Конфигурация PostgreSQL

```bash
# Переменные окружения для PostgreSQL
export DATABASE_DSN="postgres://user:password@localhost:5432/metrics_db?sslmode=disable"
export DB_MAX_CONNS=10
export DB_MIN_CONNS=2
export DB_MAX_CONN_LIFETIME=1h
export DB_MAX_CONN_IDLE_TIME=30m
export DB_CONNECT_TIMEOUT=10s
export DB_PING_TIMEOUT=5s
export DB_HEALTH_CHECK_TIMEOUT=5s
```

### 🏗️ Новая схема базы данных

**Автоматически создаваемые таблицы:**

```sql
-- Таблица для gauge метрик
CREATE TABLE gauge_metrics (
    id VARCHAR(255) PRIMARY KEY,
    value DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица для counter метрик  
CREATE TABLE counter_metrics (
    id VARCHAR(255) PRIMARY KEY,
    value BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица для отслеживания миграций
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    checksum VARCHAR(64) NOT NULL
);
```

### 🚀 Приоритет хранения

Сервис автоматически выбирает тип хранилища по приоритету:

1. **PostgreSQL** (если указан `DATABASE_DSN` или `-d`)
2. **Файл** (если указан `FILE_STORAGE_PATH` или `-f`)  
3. **Память** (по умолчанию)

```bash
# PostgreSQL (приоритет 1)
./cmd/server/server -d="postgres://user:pass@localhost:5432/db"

# Файл (приоритет 2) 
./cmd/server/server -f="/tmp/metrics.json"

# Память (приоритет 3)
./cmd/server/server
```

### 🐳 Docker Compose инфраструктура

Проект включает две Docker Compose конфигурации:

#### Основная база данных (`docker-compose.yml`)
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    container_name: metrics-main-db
    restart: always
    environment:
      POSTGRES_DB: metricaldb
      POSTGRES_USER: metricaldb
      POSTGRES_PASSWORD: Secret
    ports:
      - "5434:5432"  # Основной порт PostgreSQL
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U metricaldb -d metricaldb"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

#### Тестовая база данных (`docker-compose.test.yml`)
```yaml
version: '3.8'
services:
  postgres-test:
    image: postgres:15-alpine
    container_name: metrics-test-db
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5433:5432"  # Тестовый порт PostgreSQL
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test -d testdb"]
      interval: 5s
      timeout: 5s
      retries: 5
    tmpfs:
      - /var/lib/postgresql/data:rw  # Быстрая работа и автоматическая очистка
```

#### Команды управления Docker

```bash
# Основная БД
make db-up          # Запустить основную БД
make db-down        # Остановить основную БД
make db-logs        # Просмотр логов основной БД
make db-shell       # Подключение к основной БД
make db-reset       # Сбросить основную БД

# Тестовая БД
make test-db        # Полный цикл тестов (up → test → down)
make test-db-up     # Запустить тестовую БД
make test-db-down   # Остановить тестовую БД
make test-db-run    # Запустить тесты с БД

# Управление сервисами
make run-full       # Запустить БД + сервер
make stop-all       # Остановить все сервисы
make status         # Показать статус всех сервисов
```

### 🔒 Безопасность и надежность

- ✅ **Автоматические миграции** - создание таблиц при запуске
- ✅ **Контрольные суммы** - проверка целостности миграций
- ✅ **Настройки PostgreSQL** - автоматическая настройка совместимости
- ✅ **Connection Pooling** - эффективное управление соединениями
- ✅ **Атомарные операции** - UPSERT с ON CONFLICT для thread-safety
- ✅ **Валидация данных** - проверка входных параметров
- ✅ **Структурированное логирование** - детальное отслеживание операций
- ✅ **Health Check** - проверка состояния базы данных
- ✅ **Retry логика** - автоматические повторы при сбоях
- ✅ **Конфигурируемые таймауты** - настройка всех временных параметров

📖 **Подробная документация:** [internal/repository/README.md](internal/repository/README.md)

## 🚀 Функциональность

### Поддерживаемые типы метрик

1. **Gauge** (float64) - новое значение замещает предыдущее
2. **Counter** (int64) - новое значение добавляется к предыдущему

### Gzip поддержка

Проект поддерживает автоматическое сжатие и распаковку данных с помощью gzip:

- **Сжатие ответов**: Сервер автоматически сжимает ответы, если клиент поддерживает gzip
- **Распаковка запросов**: Сервер автоматически распаковывает входящие сжатые запросы
- **Автоматическое сжатие в агенте**: Все отправляемые JSON метрики автоматически сжимаются
- **Умная фильтрация**: Сжатие применяется только к поддерживаемым типам контента (JSON, HTML, plain text)

#### Примеры использования gzip

**Сжатые ответы сервера:**
```bash
# Запрос с поддержкой gzip
curl -H "Accept-Encoding: gzip" http://localhost:8080/

# Ответ будет сжат и содержать заголовок Content-Encoding: gzip
```

**Отправка сжатых запросов:**
```bash
# Отправка сжатого JSON
echo '{"id":"test","type":"gauge","value":42.5}' | gzip | \
curl -X POST http://localhost:8080/update \
  -H "Content-Type: application/json" \
  -H "Content-Encoding: gzip" \
  --data-binary @-
```

**Автоматическое сжатие в агенте:**
```go
// Агент автоматически сжимает все JSON метрики
err := agent.sendSingleMetricJSON("test_metric", value)
// Данные сжимаются и отправляются с gzip заголовками
```

### 🛡️ Улучшенная обработка ошибок и логирование

Проект включает продвинутую систему обработки ошибок и структурированного логирования:

#### Структурированное логирование
```go
// Все компоненты используют единую систему логирования
logger.Info("Server started", "port", 8080, "env", "production")
logger.Error("Database connection failed", "error", err, "retry_count", 3)
logger.Debug("Processing metric", "name", "temperature", "value", 23.5)
```

#### Валидация данных
```go
// Встроенная валидация всех входных данных
func (r *PostgreSQLMetricsRepository) validateMetricName(name string) error {
    if name == "" {
        return fmt.Errorf("metric name cannot be empty")
    }
    return nil
}

func (r *PostgreSQLMetricsRepository) validateGaugeValue(value float64) error {
    if math.IsNaN(value) {
        return fmt.Errorf("gauge value cannot be NaN")
    }
    if math.IsInf(value, 0) {
        return fmt.Errorf("gauge value cannot be infinite")
    }
    return nil
}
```

#### Retry логика
```go
// Автоматические повторы при сбоях с экспоненциальной задержкой
func (c *Connection) PingWithRetry(ctx context.Context, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        if err := c.Ping(ctx); err == nil {
            return nil
        }
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return fmt.Errorf("ping failed after %d retries", maxRetries)
}
```

### 🔧 Конфигурация и переменные окружения

Проект поддерживает гибкую конфигурацию через переменные окружения с константами:

#### Константы для переменных окружения
```go
const (
    EnvDatabaseDSN            = "DATABASE_DSN"
    EnvDBMaxConns             = "DB_MAX_CONNS"
    EnvDBMinConns             = "DB_MIN_CONNS"
    EnvDBMaxConnLifetime      = "DB_MAX_CONN_LIFETIME"
    EnvDBMaxConnIdleTime      = "DB_MAX_CONN_IDLE_TIME"
    EnvDBConnectTimeout       = "DB_CONNECT_TIMEOUT"
    EnvDBPingTimeout          = "DB_PING_TIMEOUT"
    EnvDBHealthCheckTimeout   = "DB_HEALTH_CHECK_TIMEOUT"
)
```

#### Безопасная маскировка паролей
```go
// DSN с паролем автоматически маскируется в логах
dsn := "postgres://user:secret@localhost:5432/db"
masked := maskDSN(dsn) // "postgres://user:***@localhost:5432/db"
```

### 🧪 Продвинутое тестирование

Проект включает комплексную систему тестирования:

#### Unit тесты с моками
```go
func TestPostgreSQLRepositoryWithMockPool(t *testing.T) {
    mockPool := &MockDatabasePool{}
    mockPool.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(pgconn.CommandTag("INSERT 0 1"), nil)
    
    repo := repository.NewPostgreSQLMetricsRepositoryWithPool(mockPool, logger)
    // Тестирование без реальной БД
}
```

#### Интеграционные тесты
```go
func TestPostgreSQLRepository(t *testing.T) {
    testDB := setupTestDB(t)
    defer testDB.Close()
    
    repo := repository.NewPostgreSQLMetricsRepository(testDB, logger)
    // Тестирование с реальной БД
}
```

#### Бенчмарки производительности
```go
func BenchmarkPostgreSQLRepository(b *testing.B) {
    b.Run("UpdateGauge", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            err := repo.UpdateGauge(ctx, "benchmark_gauge", float64(i))
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

### HTTP API

#### Обновление метрики
```bash
POST /update/{ТИП_МЕТРИКИ}/{ИМЯ_МЕТРИКИ}/{ЗНАЧЕНИЕ_МЕТРИКИ}

# Примеры:
curl -X POST "http://localhost:8080/update/gauge/temperature/23.5"
curl -X POST "http://localhost:8080/update/counter/requests/100"
```

#### Получение значения метрики
```bash
GET /value/{ТИП_МЕТРИКИ}/{ИМЯ_МЕТРИКИ}

# Примеры:
curl "http://localhost:8080/value/gauge/temperature"
curl "http://localhost:8080/value/counter/requests"
```

#### Просмотр всех метрик
```bash
GET /

# Открыть в браузере: http://localhost:8080/
```

## 🧪 Тестирование

### 🐳 Docker тестирование (Рекомендуется)

Проект включает полную Docker инфраструктуру для тестирования:

```bash
# Полный цикл тестов с Docker
make test-db        # Запустить тестовую БД → тесты → остановить БД

# Управление тестовой БД
make test-db-up     # Запустить тестовую БД
make test-db-run    # Запустить тесты с БД
make test-db-down   # Остановить тестовую БД

# Быстрые тесты
make quick-test     # Только unit тесты (без БД)
```

### VS Code задачи
```bash
# Полный набор тестов
Ctrl+Shift+P → "Tasks: Run Task" → "Full Test Suite"

# Только unit тесты
Ctrl+Shift+P → "Tasks: Run Task" → "Run All Tests"

# Автотесты
Ctrl+Shift+P → "Tasks: Run Task" → "Run Auto Tests Iteration4"
```

### Ручной запуск
```bash
# Все тесты
go test ./... -v

# Тесты с покрытием
go test ./... -v -cover

# Тесты gzip функциональности
go test ./internal/middleware/... -v
go test ./internal/agent/... -v

# Тесты PostgreSQL репозитория (автоматически использует значение по умолчанию)
go test ./internal/repository/... -v

# Или с кастомной БД
TEST_DATABASE_URL="postgres://user:pass@localhost:5432/mydb" go test ./internal/repository/... -v

# Тесты конфигурации базы данных
go test ./internal/config/db/... -v

# Бенчмарки производительности
go test -bench=. -benchmem ./internal/repository

# Тесты с профилированием
go test -cpuprofile=cpu.prof -bench=.
go test -memprofile=mem.prof -bench=.
```

### 🔧 Настройка тестовой среды

#### Переменные окружения для тестов
```bash
# Тестовая база данных (опционально - есть значение по умолчанию)
export TEST_DATABASE_URL="postgres://test:test@localhost:5433/testdb?sslmode=disable"

# Основная база данных
export DATABASE_DSN="postgres://metricaldb:Secret@localhost:5434/metricaldb?sslmode=disable"
```

**Примечание**: `TEST_DATABASE_URL` не обязательна - тесты автоматически используют значение по умолчанию `postgres://test:test@localhost:5433/testdb?sslmode=disable`

#### Скрипты для тестирования
```bash
# Запуск тестов с автоматической настройкой БД
./scripts/test-db.sh

# Остановка всех сервисов
./scripts/stop-all.sh
```

## Документация пакетов

- 📖 **Сервер:** [cmd/server/README.md](cmd/server/README.md)
- 📖 **Агент:** [cmd/agent/README.md](cmd/agent/README.md)
- 📖 **Приложение:** [internal/app/README.md](internal/app/README.md)
- 📖 **HTTP сервер:** [internal/httpserver/README.md](internal/httpserver/README.md)
- 📖 **Роутер:** [internal/router/README.md](internal/router/README.md)
- 📖 **Обработчики:** [internal/handler/README.md](internal/handler/README.md)
- 📖 **Сервис:** [internal/service/README.md](internal/service/README.md)
- 📖 **Валидация:** [internal/validation/README.md](internal/validation/README.md)
- 📖 **Шаблоны:** [internal/template/README.md](internal/template/README.md)
- 📖 **Маршруты:** [internal/routes/README.md](internal/routes/README.md)
- 📖 **Модели:** [internal/model/README.md](internal/model/README.md)
- 📖 **Репозиторий:** [internal/repository/README.md](internal/repository/README.md)
- 📖 **Логгер:** [internal/logger/README.md](internal/logger/README.md)
- 📖 **Middleware:** [internal/middleware/README.md](internal/middleware/README.md)
- 📖 **Test Utils:** [internal/testutils/README.md](internal/testutils/README.md)
- 📖 **Конфигурация БД:** [internal/config/db/README.md](internal/config/db/README.md)

## 🆕 Новые возможности

### 🗄️ PostgreSQL интеграция
- ✅ **Полная поддержка PostgreSQL** - хранение метрик в БД
- ✅ **Connection pooling** - эффективное управление соединениями
- ✅ **Атомарные операции** - UPSERT с ON CONFLICT
- ✅ **Retry логика** - автоматические повторы при сбоях
- ✅ **Health checks** - проверка состояния БД

### 🐳 Docker инфраструктура
- ✅ **Docker Compose** - полная инфраструктура для разработки
- ✅ **Изолированные тесты** - отдельная тестовая БД в Docker
- ✅ **Автоматические миграции** - создание схемы при запуске
- ✅ **Health checks** - проверка готовности контейнеров
- ✅ **Управление сервисами** - простые команды make для всех операций

### 🛠️ Улучшенное управление сервисами
- ✅ **make run-full** - запуск всей системы одной командой
- ✅ **make stop-all** - надежная остановка всех сервисов
- ✅ **make status** - проверка статуса всех компонентов
- ✅ **Автоматические скрипты** - упрощение рутинных операций
- ✅ **Цветной вывод** - информативные сообщения с эмодзи
