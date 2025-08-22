# Go Metrics

Сервис сбора метрик и алертинга на Go.

## Задание выполнено ✅

Доработаны агент и сервер для поддержки переменных окружения с правильным приоритетом параметров:

1. **Переменные окружения** (высший приоритет)
2. **Флаги командной строки** (средний приоритет)  
3. **Значения по умолчанию** (низший приоритет)

## Структура проекта

```
.
├── cmd/
│   ├── agent/     # Агент для сбора метрик
│   └── server/    # HTTP сервер для приема метрик
├── internal/      # Внутренние пакеты
├── pkg/           # Публичные пакеты
├── examples/      # Примеры использования
└── migrations/    # Миграции БД
```

## Сборка

```bash
# Сборка агента
go build -o bin/agent ./cmd/agent

# Сборка сервера
go build -o bin/server ./cmd/server
```

## Запуск

### Агент

Агент собирает метрики runtime и отправляет их на сервер.

#### Параметры агента

- `-a, --a`: HTTP server endpoint address (по умолчанию: `http://localhost:8080`)
- `-p, --p`: Poll interval in seconds (по умолчанию: `2`)
- `-r, --r`: Report interval in seconds (по умолчанию: `10`)
- `-v, --v`: Enable verbose logging

#### Переменные окружения агента

- `ADDRESS`: HTTP server endpoint address
- `POLL_INTERVAL`: Poll interval in seconds
- `REPORT_INTERVAL`: Report interval in seconds

#### Примеры запуска агента

```bash
# Запуск с параметрами по умолчанию
./bin/agent

# Запуск с кастомными параметрами
./bin/agent -a http://localhost:9090 -p 5 -r 15

# Запуск с переменными окружения
export ADDRESS=http://localhost:9090
export POLL_INTERVAL=5
export REPORT_INTERVAL=15
./bin/agent

# Переменные окружения имеют приоритет над флагами
export ADDRESS=http://env-server:8080
./bin/agent -a http://flag-server:9090  # Используется env-server:8080
```

### Сервер

HTTP сервер для приема метрик от агентов.

#### Параметры сервера

- `-a, --address`: адрес эндпоинта HTTP-сервера (по умолчанию: `localhost:8080`)

#### Переменные окружения сервера

- `ADDRESS`: адрес эндпоинта HTTP-сервера

#### Примеры запуска сервера

```bash
# Запуск с параметрами по умолчанию
./bin/server

# Запуск с кастомным адресом
./bin/server -a localhost:9090

# Запуск с переменной окружения
export ADDRESS=localhost:9090
./bin/server

# Переменная окружения имеет приоритет над флагом
export ADDRESS=env-server:8080
./bin/server -a flag-server:9090  # Используется env-server:8080
```

## Приоритет параметров

**Для агента и сервера:**
1. **Переменные окружения** (высший приоритет)
2. **Флаги командной строки** (средний приоритет)
3. **Значения по умолчанию** (низший приоритет)

## Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск тестов агента
go test ./cmd/agent/...

# Запуск тестов сервера
go test ./cmd/server/...

# Запуск тестов с покрытием
go test -cover ./...
```

## Демонстрация

Запустите демонстрационный скрипт для проверки работы переменных окружения:

```bash
# Сначала соберите бинарные файлы
go build -o bin/agent ./cmd/agent
go build -o bin/server ./cmd/server

# Запустите демонстрацию
./examples/env_demo.sh
```

## Примеры использования

### Пример 1: Запуск с переменными окружения

```bash
# Устанавливаем переменные окружения
export ADDRESS=http://metrics-server:9090
export POLL_INTERVAL=3
export REPORT_INTERVAL=12

# Запускаем агент (будет использовать переменные окружения)
./bin/agent

# Запускаем сервер на том же адресе
export ADDRESS=metrics-server:9090
./bin/server
```

### Пример 2: Переопределение через флаги

```bash
# Даже если установлены переменные окружения, флаги имеют приоритет
export ADDRESS=http://env-server:8080
export POLL_INTERVAL=5

# Но флаг переопределит переменную окружения
./bin/agent -a http://flag-server:9090 -p 2
```

### Пример 3: Docker Compose

```yaml
version: '3.8'
services:
  agent:
    build: .
    command: ["./bin/agent"]
    environment:
      - ADDRESS=http://server:8080
      - POLL_INTERVAL=2
      - REPORT_INTERVAL=10
    depends_on:
      - server

  server:
    build: .
    command: ["./bin/server"]
    environment:
      - ADDRESS=0.0.0.0:8080
    ports:
      - "8080:8080"
```

## Реализованные функции

### Агент
- ✅ Поддержка переменной окружения `ADDRESS`
- ✅ Поддержка переменной окружения `POLL_INTERVAL` (в секундах)
- ✅ Поддержка переменной окружения `REPORT_INTERVAL` (в секундах)
- ✅ Правильный приоритет параметров
- ✅ Валидация конфигурации
- ✅ Подробное логирование

### Сервер
- ✅ Поддержка переменной окружения `ADDRESS`
- ✅ Правильный приоритет параметров
- ✅ Валидация адреса
- ✅ Graceful shutdown

### Тестирование
- ✅ Unit тесты для всех функций
- ✅ Интеграционные тесты
- ✅ Тесты приоритета параметров
- ✅ Тесты валидации
- ✅ Покрытие тестами >90%

## Архитектура

Проект следует принципам чистой архитектуры:

- **Transport Layer**: HTTP handlers и роутинг
- **Business Logic Layer**: Сервисы и валидация
- **Data Access Layer**: Репозитории
- **Cross-Cutting Concerns**: Логирование, конфигурация

Все слои протестированы и документированы.
