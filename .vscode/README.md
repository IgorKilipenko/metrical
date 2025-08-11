# VS Code Tasks для Go Metrics

Этот файл содержит описание задач VS Code для работы с проектом Go Metrics.

## ⚡ Быстрый старт

### Основные команды:
```bash
# Сборка проекта
Ctrl+Shift+B

# Запуск всех тестов (unit + автотесты)
Ctrl+Shift+P → "Tasks: Run Task" → "Full Test Suite"

# Запуск только автотестов
Ctrl+Shift+P → "Tasks: Run Task" → "Run Auto Tests"
```

### Проверка работоспособности:
```bash
# 1. Сборка
Ctrl+Shift+B

# 2. Запуск автотестов
Ctrl+Shift+P → "Tasks: Run Task" → "Run Auto Tests"

# 3. Если все прошло успешно - проект готов к работе!
```

## 🚀 Доступные задачи

### Сборка (Build Tasks)

| Задача | Описание | Горячие клавиши |
|--------|----------|-----------------|
| **Build Agent** | Сборка агента метрик | `Ctrl+Shift+P` → "Tasks: Run Task" → "Build Agent" |
| **Build Server** | Сборка сервера метрик | `Ctrl+Shift+P` → "Tasks: Run Task" → "Build Server" |
| **Build All** | Сборка агента и сервера (по умолчанию) | `Ctrl+Shift+B` |
| **Clean Build** | Очистка сборки | `Ctrl+Shift+P` → "Tasks: Run Task" → "Clean Build" |

### Тестирование (Test Tasks)

| Задача | Описание | Горячие клавиши |
|--------|----------|-----------------|
| **Run All Tests** | Запуск всех unit тестов | `Ctrl+Shift+P` → "Tasks: Run Task" → "Run All Tests" |
| **Run Tests with Coverage** | Запуск тестов с покрытием | `Ctrl+Shift+P` → "Tasks: Run Task" → "Run Tests with Coverage" |
| **Run Auto Tests** | Запуск автотестов для Iteration4 | `Ctrl+Shift+P` → "Tasks: Run Task" → "Run Auto Tests" |
| **Run All Auto Tests** | Запуск всех автотестов | `Ctrl+Shift+P` → "Tasks: Run Task" → "Run All Auto Tests" |
| **Full Test Suite** | Полный набор тестов (по умолчанию) | `Ctrl+Shift+P` → "Tasks: Run Task" → "Full Test Suite" |

### Запуск приложений (Run Tasks)

| Задача | Описание | Горячие клавиши |
|--------|----------|-----------------|
| **Run Server** | Запуск сервера на localhost:9090 | `Ctrl+Shift+P` → "Tasks: Run Task" → "Run Server" |
| **Run Agent** | Запуск агента на localhost:8080 | `Ctrl+Shift+P` → "Tasks: Run Task" → "Run Agent" |

## 🎯 Рекомендуемый рабочий процесс

### 1. Разработка
```bash
# Сборка проекта
Ctrl+Shift+B

# Запуск unit тестов
Ctrl+Shift+P → "Run All Tests"
```

### 2. Тестирование
```bash
# Полный набор тестов
Ctrl+Shift+P → "Full Test Suite"

# Или по отдельности:
Ctrl+Shift+P → "Run Tests with Coverage"  # Unit тесты с покрытием
Ctrl+Shift+P → "Run Auto Tests"           # Автотесты
```

### 3. Запуск приложений
```bash
# Запуск сервера
Ctrl+Shift+P → "Run Server"

# В другом терминале - запуск агента
Ctrl+Shift+P → "Run Agent"
```

## 📋 Зависимости задач

- **Run Auto Tests** зависит от **Build All**
- **Run All Auto Tests** зависит от **Build All**
- **Full Test Suite** зависит от **Run All Tests** и **Run Auto Tests**
- **Run Server** зависит от **Build Server**
- **Run Agent** зависит от **Build Agent**

## 🔧 Конфигурация автотестов

Автотесты используют следующие параметры:
- **Путь к автотестам**: `/mnt/SSD_2TB_2023/git/yandex-practicum/go/auto-tests/metricstest`
- **Порт сервера**: `9091`
- **Путь к серверу**: `cmd/server/server`
- **Путь к агенту**: `cmd/agent/agent`
- **Исходный код**: `.` (текущая директория)

## 🚨 Устранение неполадок

### Проблема: Автотесты не запускаются
```bash
# Проверьте, что файл автотестов существует
ls -la /mnt/SSD_2TB_2023/git/yandex-practicum/go/auto-tests/metricstest

# Проверьте права доступа
chmod +x /mnt/SSD_2TB_2023/git/yandex-practicum/go/auto-tests/metricstest
```

### Проблема: Сборка не проходит
```bash
# Очистите сборку
Ctrl+Shift+P → "Clean Build"

# Проверьте зависимости
go mod tidy
go mod download
```

### Проблема: Порт занят
```bash
# Найдите процесс, использующий порт
lsof -i :9091
lsof -i :9090
lsof -i :8080

# Завершите процесс
kill -9 <PID>
```

## 📊 Результаты тестирования

После выполнения задач вы увидите:
- **Unit тесты**: Покрытие кода и результаты тестов
- **Автотесты**: Результаты интеграционных тестов
- **Сборка**: Статус компиляции и ошибки

## 🎉 Готово!

Теперь вы можете эффективно работать с проектом, используя встроенные задачи VS Code!
