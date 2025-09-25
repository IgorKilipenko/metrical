# Migration Tool

CLI инструмент для управления миграциями базы данных PostgreSQL.

## Установка

```bash
go build -o migrate ./cmd/migrate
```

## Использование

### Основные команды

```bash
# Применить все миграции
./migrate -command migrate -dsn "postgres://user:pass@localhost/db"

# Откатить до версии 1
./migrate -command rollback -version 1 -dsn "postgres://user:pass@localhost/db"

# Показать статистику миграций
./migrate -command stats -dsn "postgres://user:pass@localhost/db"

# Показать историю миграций
./migrate -command history -dsn "postgres://user:pass@localhost/db"
```

### Переменные окружения

```bash
export DATABASE_DSN="postgres://user:pass@localhost/db"
./migrate -command migrate
```

### Форматы вывода

```bash
# Текстовый формат (по умолчанию)
./migrate -command stats -format text

# JSON формат
./migrate -command stats -format json
```

## Примеры

### Применение миграций

```bash
./migrate -command migrate -dsn "postgres://metricaldb:Secret@localhost:5432/metricaldb?sslmode=disable"
```

### Откат миграций

```bash
# Откатить до версии 1 (удалить миграцию 2)
./migrate -command rollback -version 1 -dsn "postgres://metricaldb:Secret@localhost:5432/metricaldb?sslmode=disable"
```

### Статистика

```bash
./migrate -command stats -dsn "postgres://metricaldb:Secret@localhost:5432/metricaldb?sslmode=disable"
```

Вывод:
```
Migration Statistics:
  Total migrations: 2
  Latest version: 2
  First migration: 2025-09-14T05:08:16+07:00
  Last migration: 2025-09-14T05:08:16+07:00
```

### История миграций

```bash
./migrate -command history -dsn "postgres://metricaldb:Secret@localhost:5432/metricaldb?sslmode=disable"
```

Вывод:
```
Migration History:
  ✅ Version 2: Create metrics indexes (applied at 2025-09-14T05:08:16+07:00)
  ✅ Version 1: Create metrics table (applied at 2025-09-14T05:08:16+07:00)
```

## Безопасность

- Все операции выполняются в транзакциях
- При ошибке автоматически выполняется rollback
- Валидация миграций перед выполнением
- Детальное логирование всех операций
