# internal/repository

Этот пакет содержит реализацию работы с базой данных, а также со внешними сервисами.

Важно, чтобы репозиторий не содержал бизнес-логику.

Репозиторий реализует паттерн Repository и служит абстракцией над различными источниками данных, такими как:
- базы данных (PostgreSQL, MySQL и др.)
- внешние API
- файловые системы
- кэши (Redis, Memcached)
- другие источники данных.

## Архитектура

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│     Service     │───▶│   Repository     │───▶│   Data Source   │
│                 │    │   (Interface)    │    │   (Memory/DB)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Компоненты

### MetricsRepository (Интерфейс)

Основной интерфейс для работы с метриками с поддержкой контекста:

```go
type MetricsRepository interface {
    UpdateGauge(ctx context.Context, name string, value float64) error
    UpdateCounter(ctx context.Context, name string, value int64) error
    UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error // НОВОЕ: батчевое обновление
    GetGauge(ctx context.Context, name string) (float64, bool, error)
    GetCounter(ctx context.Context, name string) (int64, bool, error)
    GetAllGauges(ctx context.Context) (models.GaugeMetrics, error)
    GetAllCounters(ctx context.Context) (models.CounterMetrics, error)
    SaveToFile() error
    LoadFromFile() error
    SetSyncSave(sync bool)
}
```

### InMemoryMetricsRepository (Реализация)

Реализация репозитория в памяти с потокобезопасностью, поддержкой контекста и логированием:

```go
type InMemoryMetricsRepository struct {
    Gauges          models.GaugeMetrics
    Counters        models.CounterMetrics
    mu              sync.RWMutex
    logger          logger.Logger
    fileStoragePath string
    restore         bool
    syncSave        bool
}
```

### PostgreSQLMetricsRepository (Реализация)

Реализация репозитория для PostgreSQL с connection pooling, поддержкой контекста, структурированным логированием и **автоматическими миграциями**:

```go
type PostgreSQLMetricsRepository struct {
    pool   *pgxpool.Pool
    logger logger.Logger
}
```

**Особенности PostgreSQL реализации:**
- 🗄️ **Персистентное хранение** - данные сохраняются в PostgreSQL
- 🔄 **Connection pooling** - эффективное управление соединениями
- ⚡ **Оптимизированные SQL запросы** - UPSERT операции с ON CONFLICT
- 🛡️ **Валидация данных** - проверка входных параметров
- 📊 **Атомарные операции** - нет race conditions при concurrent доступе
- 🔍 **Структурированное логирование** - детальное отслеживание операций
- 🚀 **Автоматические миграции** - создание таблиц при запуске
- 🔒 **Контрольные суммы** - проверка целостности миграций
- ⚙️ **Настройки PostgreSQL** - автоматическая настройка совместимости

## Использование

### Создание репозитория

#### In-Memory репозиторий

```go
// Создаем логгер
appLogger := logger.NewSlogLogger()

// Создаем репозиторий в памяти с логгером и настройками персистентности
repo := repository.NewInMemoryMetricsRepository(
    appLogger,
    "/tmp/metrics.json",  // путь к файлу для сохранения
    true,                 // загружать метрики при старте
)

// Создаем сервис с репозиторием и логгером
service := service.NewMetricsService(repo, appLogger)
```

#### PostgreSQL репозиторий с автоматическими миграциями

```go
import (
    "context"
    "time"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/IgorKilipenko/metrical/internal/logger"
    "github.com/IgorKilipenko/metrical/internal/repository"
)

// Создаем логгер
appLogger := logger.NewSlogLogger()

// Создаем пул соединений с PostgreSQL
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

dsn := "postgres://user:password@localhost:5432/metrics_db?sslmode=disable"
pool, err := pgxpool.New(ctx, dsn)
if err != nil {
    log.Fatal("Failed to create connection pool:", err)
}
defer pool.Close()

// Создаем PostgreSQL репозиторий с автоматическими миграциями
// Миграции будут применены автоматически при создании репозитория
repo := repository.NewPostgreSQLMetricsRepositoryWithMigrations(pool, appLogger)

// Создаем сервис с репозиторием и логгером
service := service.NewMetricsService(repo, appLogger)
```

#### PostgreSQL репозиторий без миграций (ручное управление)

```go
// Создаем PostgreSQL репозиторий без автоматических миграций
repo := repository.NewPostgreSQLMetricsRepository(pool, appLogger)

// Создаем сервис с репозиторием и логгером
service := service.NewMetricsService(repo, appLogger)
```

#### Создание с интерфейсом для тестирования

```go
// Для лучшей тестируемости можно использовать интерфейс DatabasePool
var pool repository.DatabasePool = pgxPool // или mock в тестах
repo := repository.NewPostgreSQLMetricsRepositoryWithPool(pool, appLogger)
```

### Основные операции

```go
ctx := context.Background()

// Обновление метрик с контекстом
err := repo.UpdateGauge(ctx, "temperature", 23.5)
err := repo.UpdateCounter(ctx, "requests", 100)

// Получение метрик с контекстом
value, exists, err := repo.GetGauge(ctx, "temperature")
value, exists, err := repo.GetCounter(ctx, "requests")

// Получение всех метрик с контекстом
gauges, err := repo.GetAllGauges(ctx)
counters, err := repo.GetAllCounters(ctx)

// Батчевое обновление метрик (НОВОЕ)
metrics := []models.Metrics{
    {
        ID:    "temperature",
        MType: "gauge",
        Value: func() *float64 { v := 23.5; return &v }(),
    },
    {
        ID:    "requests_total",
        MType: "counter",
        Delta: func() *int64 { v := int64(100); return &v }(),
    },
}
err = repo.UpdateMetricsBatch(ctx, metrics)
```

### Работа с таймаутами и отменой

```go
// Создание контекста с таймаутом
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Операция будет отменена через 5 секунд
err := repo.UpdateGauge(ctx, "temperature", 23.5)
if err != nil {
    if err == context.DeadlineExceeded {
        log.Println("Operation timed out")
    } else if err == context.Canceled {
        log.Println("Operation was canceled")
    }
}
```

## Логирование

Репозиторий интегрирован с системой логирования для отслеживания операций:

```go
// Логирование операций обновления
repo.UpdateGauge(ctx, "temperature", 23.5)
// Логи: "Updating gauge metric" name=temperature value=23.5

// Логирование операций получения
value, exists, err := repo.GetGauge(ctx, "temperature")
// Логи: "Retrieved gauge metric" name=temperature value=23.5 exists=true

// Логирование ошибок контекста
if err == context.Canceled {
    // Логи: "Context canceled during operation" operation=UpdateGauge
}
```

### Уровни логирования

- **Debug**: Детальная информация об операциях
- **Info**: Основные операции (создание, обновление, получение)
- **Error**: Ошибки операций и отмены контекста

## Преимущества

- **Абстракция данных** - сервис не зависит от конкретной реализации хранения
- **Легкое тестирование** - можно легко мокать репозиторий
- **Расширяемость** - легко добавить новые реализации (PostgreSQL, Redis)
- **Потокобезопасность** - встроенная защита от гонки данных
- **Поддержка контекста** - отмена операций, таймауты, graceful shutdown
- **Чистая архитектура** - четкое разделение ответственности

## Особенности реализации

### Обработка контекста

Все методы репозитория проверяют отмену контекста:

```go
func (r *InMemoryMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
    // Проверяем отмену контекста
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    r.mu.Lock()
    defer r.mu.Unlock()
    r.Gauges[name] = value
    return nil
}
```

### Потокобезопасность

#### In-Memory репозиторий
Реализация использует `sync.RWMutex` для обеспечения потокобезопасности:

- Операции записи (`UpdateGauge`, `UpdateCounter`) используют `Lock()`
- Операции чтения (`GetGauge`, `GetCounter`, `GetAllGauges`, `GetAllCounters`) используют `RLock()`

#### PostgreSQL репозиторий
Потокобезопасность обеспечивается на уровне базы данных:

- **Connection pooling** - пул соединений управляет concurrent доступом
- **Атомарные SQL операции** - UPSERT с ON CONFLICT предотвращает race conditions
- **ACID транзакции** - PostgreSQL гарантирует консистентность данных
- **Row-level locking** - автоматическая блокировка на уровне строк

## Тестирование

```bash
go test -v ./internal/repository
```

### Тестирование с контекстом

```go
func TestRepositoryWithContext(t *testing.T) {
    repo := repository.NewInMemoryMetricsRepository()
    ctx := context.Background()
    
    // Тест с обычным контекстом
    err := repo.UpdateGauge(ctx, "test", 23.5)
    assert.NoError(t, err)
    
    // Тест с отмененным контекстом
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    
    err = repo.UpdateGauge(ctx, "test", 23.5)
    assert.Equal(t, context.Canceled, err)
}
```

## 💾 Персистентность метрик

Репозиторий поддерживает сохранение и загрузку метрик в/из JSON файла:

### Сохранение метрик

```go
// Сохранение всех метрик в файл
err := repo.SaveToFile()
if err != nil {
    log.Printf("Failed to save metrics: %v", err)
}
```

### Загрузка метрик

```go
// Загрузка метрик из файла при старте
err := repo.LoadFromFile()
if err != nil {
    log.Printf("Failed to load metrics: %v", err)
}
```

### Синхронное сохранение

```go
// Включение синхронного сохранения (каждое обновление сразу на диск)
repo.SetSyncSave(true)

// Теперь каждое обновление метрики автоматически сохраняется
err := repo.UpdateGauge(ctx, "temperature", 23.5)
// Метрики автоматически сохраняются в файл
```

### Формат файла

Метрики сохраняются в JSON формате:

```json
[
  {"id":"LastGC","type":"gauge","value":1257894000000000000},
  {"id":"NumGC","type":"counter","delta":42}
]
```

## 🗄️ PostgreSQL особенности

### Схема базы данных

PostgreSQL репозиторий использует следующую схему таблицы:

```sql
CREATE TABLE metrics (
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION,
    delta BIGINT,
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (name, type)
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_metrics_type ON metrics(type);
CREATE INDEX idx_metrics_updated_at ON metrics(updated_at);
```

### Оптимизированные SQL запросы

#### Gauge метрики (UPSERT)
```sql
INSERT INTO metrics (name, type, value, updated_at) 
VALUES ($1, $2, $3, NOW())
ON CONFLICT (name, type) 
DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
```

#### Counter метрики (атомарный инкремент)
```sql
INSERT INTO metrics (name, type, delta, updated_at) 
VALUES ($1, $2, $3, NOW())
ON CONFLICT (name, type) 
DO UPDATE SET delta = metrics.delta + EXCLUDED.delta, updated_at = NOW()
```

#### Батчевое обновление (НОВОЕ)
```sql
-- Начинаем транзакцию
BEGIN;

-- Обновляем gauge метрики
INSERT INTO gauge_metrics (id, value, updated_at) 
VALUES ($1, $2, NOW())
ON CONFLICT (id) 
DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();

-- Обновляем counter метрики
INSERT INTO counter_metrics (id, value, updated_at) 
VALUES ($1, $2, NOW())
ON CONFLICT (id) 
DO UPDATE SET value = counter_metrics.value + EXCLUDED.value, updated_at = NOW();

-- Коммитим транзакцию
COMMIT;
```

### Валидация данных

PostgreSQL репозиторий включает встроенную валидацию:

```go
// Валидация имени метрики
func (r *PostgreSQLMetricsRepository) validateMetricName(name string) error {
    if name == "" {
        return fmt.Errorf("metric name cannot be empty")
    }
    return nil
}

// Валидация значения gauge
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

### Connection Pooling

Репозиторий использует pgxpool для эффективного управления соединениями:

```go
// Настройка пула соединений
config, err := pgxpool.ParseConfig(dsn)
if err != nil {
    return err
}

// Настройки пула
config.MaxConns = 10                    // Максимум соединений
config.MinConns = 2                     // Минимум соединений
config.MaxConnLifetime = time.Hour      // Время жизни соединения
config.MaxConnIdleTime = time.Minute * 30 // Время простоя соединения

pool, err := pgxpool.NewWithConfig(ctx, config)
```

### Обработка ошибок

PostgreSQL репозиторий обрабатывает специфичные ошибки базы данных:

```go
// Обработка отсутствующих записей
if err == pgx.ErrNoRows {
    r.logger.Debug("metric not found", "name", name)
    return 0, false, nil
}

// Обработка ошибок подключения
if err != nil {
    r.logger.Error("database operation failed", "error", err)
    return err
}
```

### Логирование операций

Структурированное логирование всех операций:

```go
// Успешные операции
r.logger.Debug("updated gauge metric", "name", name, "value", value)
r.logger.Debug("retrieved gauge metric", "name", name, "value", value)

// Ошибки
r.logger.Error("failed to update gauge metric", "name", name, "value", value, "error", err)
r.logger.Error("failed to get gauge metric", "name", name, "error", err)
```

## Примеры

### Базовое использование

#### In-Memory репозиторий

```go
package main

import (
    "context"
    "time"
    
    "github.com/IgorKilipenko/metrical/internal/repository"
    "github.com/IgorKilipenko/metrical/internal/service"
)

func main() {
    // Создаем логгер
    appLogger := logger.NewSlogLogger()
    
    // Создаем репозиторий с логгером
    repo := repository.NewInMemoryMetricsRepository(appLogger)
    
    // Создаем сервис с логгером
    service := service.NewMetricsService(repo, appLogger)
    
    // Создаем контекст с таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Используем сервис с контекстом
    err := service.UpdateMetric(ctx, &validation.MetricRequest{
        Type:  "gauge",
        Name:  "temperature",
        Value: 23.5,
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

#### PostgreSQL репозиторий

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/IgorKilipenko/metrical/internal/logger"
    "github.com/IgorKilipenko/metrical/internal/repository"
    "github.com/IgorKilipenko/metrical/internal/service"
)

func main() {
    // Создаем логгер
    appLogger := logger.NewSlogLogger()
    
    // Создаем пул соединений с PostgreSQL
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    dsn := "postgres://user:password@localhost:5432/metrics_db?sslmode=disable"
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        log.Fatal("Failed to create connection pool:", err)
    }
    defer pool.Close()
    
    // Создаем PostgreSQL репозиторий
    repo := repository.NewPostgreSQLMetricsRepository(pool, appLogger)
    
    // Создаем сервис с репозиторием и логгером
    service := service.NewMetricsService(repo, appLogger)
    
    // Создаем контекст с таймаутом
    ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Используем сервис с контекстом
    err = service.UpdateMetric(ctx, &validation.MetricRequest{
        Type:  "gauge",
        Name:  "temperature",
        Value: 23.5,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Получаем все метрики
    gauges, err := repo.GetAllGauges(ctx)
    if err != nil {
        log.Printf("Failed to get gauges: %v", err)
    } else {
        log.Printf("Retrieved %d gauge metrics", len(gauges))
    }
}
```

### Graceful Shutdown

```go
func gracefulShutdown(repo repository.MetricsRepository) {
    // Создаем контекст для graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Выполняем финальные операции
    gauges, err := repo.GetAllGauges(ctx)
    if err != nil {
        log.Printf("Error getting final gauges: %v", err)
        return
    }
    
    // Сохраняем данные или выполняем cleanup
    log.Printf("Final gauges: %v", gauges)
}
```

### Тестирование с моками

#### In-Memory репозиторий

```go
func TestServiceWithMockRepository(t *testing.T) {
    // Создаем мок репозитория
    mockRepo := &MockMetricsRepository{}
    
    // Настраиваем ожидания с контекстом
    mockRepo.On("UpdateGauge", mock.Anything, "test", 23.5).Return(nil)
    
    // Создаем мок логгера
    mockLogger := &MockLogger{}
    
    // Создаем сервис с моком и логгером
    service := service.NewMetricsService(mockRepo, mockLogger)
    
    // Тестируем с контекстом
    ctx := context.Background()
    err := service.UpdateMetric(ctx, &validation.MetricRequest{
        Type:  "gauge",
        Name:  "test",
        Value: 23.5,
    })
    assert.NoError(t, err)
    
    // Проверяем, что мок был вызван
    mockRepo.AssertExpectations(t)
}
```

#### PostgreSQL репозиторий

```go
func TestPostgreSQLRepository(t *testing.T) {
    // Создаем тестовую базу данных
    testDB := setupTestDB(t)
    defer testDB.Close()
    
    // Создаем логгер
    logger := logger.NewSlogLogger()
    
    // Создаем репозиторий
    repo := repository.NewPostgreSQLMetricsRepository(testDB, logger)
    
    ctx := context.Background()
    
    // Тест обновления gauge метрики
    err := repo.UpdateGauge(ctx, "temperature", 23.5)
    assert.NoError(t, err)
    
    // Тест получения gauge метрики
    value, exists, err := repo.GetGauge(ctx, "temperature")
    assert.NoError(t, err)
    assert.True(t, exists)
    assert.Equal(t, 23.5, value)
    
    // Тест обновления counter метрики
    err = repo.UpdateCounter(ctx, "requests", 10)
    assert.NoError(t, err)
    
    // Тест повторного обновления counter (должно добавиться)
    err = repo.UpdateCounter(ctx, "requests", 5)
    assert.NoError(t, err)
    
    // Проверяем, что значение увеличилось
    value, exists, err = repo.GetCounter(ctx, "requests")
    assert.NoError(t, err)
    assert.True(t, exists)
    assert.Equal(t, int64(15), value) // 10 + 5
}

func TestPostgreSQLRepositoryWithMockPool(t *testing.T) {
    // Создаем мок пула соединений
    mockPool := &MockDatabasePool{}
    
    // Настраиваем ожидания
    mockPool.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(pgconn.CommandTag("INSERT 0 1"), nil)
    
    mockPool.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(&MockRow{value: 23.5})
    
    // Создаем логгер
    logger := logger.NewSlogLogger()
    
    // Создаем репозиторий с моком
    repo := repository.NewPostgreSQLMetricsRepositoryWithPool(mockPool, logger)
    
    ctx := context.Background()
    
    // Тестируем операции
    err := repo.UpdateGauge(ctx, "temperature", 23.5)
    assert.NoError(t, err)
    
    value, exists, err := repo.GetGauge(ctx, "temperature")
    assert.NoError(t, err)
    assert.True(t, exists)
    assert.Equal(t, 23.5, value)
    
    // Проверяем, что мок был вызван
    mockPool.AssertExpectations(t)
}
```

### Тестирование производительности

```go
func BenchmarkPostgreSQLRepository(b *testing.B) {
    testDB := setupTestDB(b)
    defer testDB.Close()
    
    logger := logger.NewSlogLogger()
    repo := repository.NewPostgreSQLMetricsRepository(testDB, logger)
    
    ctx := context.Background()
    
    b.Run("UpdateGauge", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            err := repo.UpdateGauge(ctx, "benchmark_gauge", float64(i))
            if err != nil {
                b.Fatal(err)
            }
        }
    })
    
    b.Run("UpdateCounter", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            err := repo.UpdateCounter(ctx, "benchmark_counter", int64(i))
            if err != nil {
                b.Fatal(err)
            }
        }
    })
    
    b.Run("GetGauge", func(b *testing.B) {
        // Предварительно создаем метрику
        repo.UpdateGauge(ctx, "benchmark_get", 42.0)
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            _, _, err := repo.GetGauge(ctx, "benchmark_get")
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

## 🚀 Настройка и миграции

### 🆕 Новая система миграций (рекомендуется)

Сервис теперь поддерживает **автоматические миграции** с использованием SQL файлов:

#### Структура миграций
```
migrations/
├── 001_create_metrics_tables.sql
├── 002_add_indexes.sql
└── 003_update_schema.sql
```

#### Автоматическое применение
```go
// Миграции применяются автоматически при создании репозитория
repo := repository.NewPostgreSQLMetricsRepositoryWithMigrations(pool, logger)
```

#### Пример миграции
```sql
-- migrations/001_create_metrics_tables.sql
-- Миграция 001: Создание таблиц для метрик
-- Автор: Igor Kilipenko

-- Настройка PostgreSQL для совместимости
SET standard_conforming_strings = on;

-- Создание таблицы для gauge метрик
CREATE TABLE IF NOT EXISTS gauge_metrics (
    id VARCHAR(255) PRIMARY KEY,
    value DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для counter метрик
CREATE TABLE IF NOT EXISTS counter_metrics (
    id VARCHAR(255) PRIMARY KEY,
    value BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_gauge_metrics_id ON gauge_metrics(id);
CREATE INDEX IF NOT EXISTS idx_counter_metrics_id ON counter_metrics(id);
```

#### Преимущества новой системы
- ✅ **Автоматическое применение** - миграции выполняются при запуске
- ✅ **Контрольные суммы** - проверка целостности SQL файлов
- ✅ **Версионирование** - строгий порядок применения
- ✅ **Настройки PostgreSQL** - автоматическая настройка совместимости
- ✅ **Детальное логирование** - отслеживание всех операций

### 🔧 Ручное создание базы данных (устаревший способ)

```sql
-- Создание базы данных
CREATE DATABASE metrics_db;

-- Подключение к базе данных
\c metrics_db;

-- Создание таблицы метрик (старая схема)
CREATE TABLE metrics (
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION,
    delta BIGINT,
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (name, type)
);

-- Создание индексов для оптимизации
CREATE INDEX idx_metrics_type ON metrics(type);
CREATE INDEX idx_metrics_updated_at ON metrics(updated_at);
CREATE INDEX idx_metrics_name_type ON metrics(name, type);
```

### Миграции с помощью Go

```go
package main

import (
    "context"
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
)

func runMigrations(dsn string) error {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return err
    }
    defer db.Close()
    
    ctx := context.Background()
    
    // Проверяем, существует ли таблица
    var exists bool
    err = db.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'metrics'
        )
    `).Scan(&exists)
    if err != nil {
        return err
    }
    
    if !exists {
        log.Println("Creating metrics table...")
        _, err = db.ExecContext(ctx, `
            CREATE TABLE metrics (
                name VARCHAR(255) NOT NULL,
                type VARCHAR(50) NOT NULL,
                value DOUBLE PRECISION,
                delta BIGINT,
                updated_at TIMESTAMP DEFAULT NOW(),
                PRIMARY KEY (name, type)
            )
        `)
        if err != nil {
            return err
        }
        
        // Создаем индексы
        _, err = db.ExecContext(ctx, `CREATE INDEX idx_metrics_type ON metrics(type)`)
        if err != nil {
            return err
        }
        
        _, err = db.ExecContext(ctx, `CREATE INDEX idx_metrics_updated_at ON metrics(updated_at)`)
        if err != nil {
            return err
        }
        
        log.Println("Database migration completed successfully")
    } else {
        log.Println("Metrics table already exists")
    }
    
    return nil
}
```

### Docker Compose для разработки

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: metrics_db
      POSTGRES_USER: metrics_user
      POSTGRES_PASSWORD: metrics_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U metrics_user -d metrics_db"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

### Переменные окружения

```bash
# Настройки базы данных
DATABASE_DSN=postgres://metrics_user:metrics_password@localhost:5432/metrics_db?sslmode=disable
DB_MAX_CONNS=10
DB_MIN_CONNS=2
DB_MAX_CONN_LIFETIME=1h
DB_MAX_CONN_IDLE_TIME=30m
DB_CONNECT_TIMEOUT=10s
DB_PING_TIMEOUT=5s
DB_HEALTH_CHECK_TIMEOUT=5s
```

## 📊 Мониторинг и отладка

### Логирование запросов

```go
// Включение логирования SQL запросов
config, err := pgxpool.ParseConfig(dsn)
if err != nil {
    return err
}

// Настройка логирования
config.ConnConfig.Logger = logger.NewSlogLogger()
config.ConnConfig.LogLevel = pgx.LogLevelDebug

pool, err := pgxpool.NewWithConfig(ctx, config)
```

### Мониторинг производительности

```go
// Получение статистики пула соединений
stats := pool.Stat()
log.Printf("Pool stats: MaxConns=%d, AcquiredConns=%d, ConstructingConns=%d, 
           IdleConns=%d, TotalConns=%d", 
           stats.MaxConns(), stats.AcquiredConns(), stats.ConstructingConns(),
           stats.IdleConns(), stats.TotalConns())
```

### Health Check

```go
func (r *PostgreSQLMetricsRepository) HealthCheck(ctx context.Context) error {
    return r.pool.Ping(ctx)
}
```

## 🔧 Troubleshooting

### Частые проблемы

1. **Connection timeout**
   ```
   Error: context deadline exceeded
   ```
   **Решение**: Увеличить `DB_CONNECT_TIMEOUT` или проверить доступность БД

2. **Too many connections**
   ```
   Error: too many connections
   ```
   **Решение**: Уменьшить `DB_MAX_CONNS` или увеличить лимиты PostgreSQL

3. **Table doesn't exist**
   ```
   Error: relation "metrics" does not exist
   ```
   **Решение**: Запустить миграции для создания таблицы

4. **Permission denied**
   ```
   Error: permission denied for table metrics
   ```
   **Решение**: Проверить права пользователя БД

### Отладка SQL запросов

```go
// Включение детального логирования
config.ConnConfig.Logger = logger.NewSlogLogger()
config.ConnConfig.LogLevel = pgx.LogLevelTrace
```

### Профилирование

```go
// Бенчмарк операций
go test -bench=. -benchmem ./internal/repository

// Профилирование CPU
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

// Профилирование памяти
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```
