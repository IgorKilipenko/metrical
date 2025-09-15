# internal/config/db

Пакет для конфигурации подключения к базе данных и управления миграциями PostgreSQL.

## 📋 Содержание

### 🔧 Конфигурация подключения
- Структуры конфигурации для подключения к PostgreSQL
- Параметры соединения и пула подключений
- Настройки таймаутов и лимитов соединений
- Валидация конфигурации

### 🚀 Система миграций
- **MigrationManager** - менеджер миграций базы данных
- Автоматическая загрузка SQL миграций из файловой системы
- Контрольные суммы для проверки целостности
- Детальное логирование и мониторинг

## 🏗️ Архитектура

```
internal/config/db/
├── config.go          # Конфигурация подключения
├── connection.go      # Управление соединениями
├── migrations.go      # Система миграций
└── migrations_test.go # Тесты миграций
```

## 🔄 Система миграций

### Структура миграции
```go
type Migration struct {
    Version   int        // Номер версии
    Name      string     // Имя миграции
    SQL       string     // SQL код миграции
    AppliedAt *time.Time // Время применения
    Checksum  string     // Контрольная сумма
}
```

### Миграционный менеджер
```go
type MigrationManager struct {
    pool   *pgxpool.Pool
    logger logger.Logger
}
```

### Основные методы
- `InitMigrationsTable()` - инициализация таблицы миграций
- `LoadMigrationsFromFS()` - загрузка миграций из файлов
- `ApplyMigration()` - применение миграции
- `RunMigrations()` - выполнение всех миграций
- `GetMigrationStats()` - статистика миграций

## 📁 Формат миграций

Миграции хранятся в директории `migrations/` в формате:
```
migrations/
├── 001_create_metrics_tables.sql
├── 002_add_indexes.sql
└── 003_update_schema.sql
```

### Пример миграции
```sql
-- Миграция 001: Создание таблиц для метрик
-- Автор: Igor Kilipenko
-- Описание: Создание базовых таблиц для хранения gauge и counter метрик

-- Настройка PostgreSQL
SET standard_conforming_strings = on;

-- Создание таблицы для gauge метрик
CREATE TABLE IF NOT EXISTS gauge_metrics (
    id VARCHAR(255) PRIMARY KEY,
    value DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 🔒 Безопасность

- **Контрольные суммы** - проверка целостности SQL файлов
- **Версионирование** - строгое соблюдение порядка миграций
- **Транзакции** - атомарное применение миграций
- **Настройки PostgreSQL** - автоматическая настройка совместимости

## 📊 Мониторинг

Система предоставляет детальную информацию:
- Количество применённых миграций
- Время последней миграции
- Статистика по версиям
- Логирование всех операций

## 🧪 Тестирование

Пакет включает комплексные тесты:
- Валидация конфигурации
- Тестирование миграций
- Проверка статистики
- Интеграционные тесты с PostgreSQL