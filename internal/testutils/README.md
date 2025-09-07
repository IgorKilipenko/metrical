# Test Utils

Пакет `testutils` содержит общие утилиты для тестирования, используемые во всем проекте.

## 📦 Содержимое

### MockLogger

`MockLogger` - это реализация интерфейса `logger.Logger` для использования в тестах.

#### Особенности:
- **No-op операции**: Все методы логирования ничего не делают
- **Thread-safe**: Безопасен для использования в конкурентных тестах
- **Легковесный**: Минимальные накладные расходы
- **Консистентный**: Одинаковое поведение во всех тестах

#### Использование:

```go
import (
    "github.com/IgorKilipenko/metrical/internal/testutils"
    "github.com/IgorKilipenko/metrical/internal/logger"
)

func TestSomething(t *testing.T) {
    // Создаем мок логгер
    mockLogger := testutils.NewMockLogger()
    
    // Используем в тестируемом коде
    service := service.NewMetricsService(repository, mockLogger)
    
    // Тестируем функциональность...
}
```

### TestFilePaths

`TestFilePaths` содержит константы с путями к тестовым файлам для избежания дублирования в тестах.

#### Доступные константы:
- `TestMetricsFile` - основной тестовый файл метрик (`/tmp/test-metrics.json`)
- `TestSyncFile` - для тестов синхронного сохранения (`/tmp/test-sync-metrics.json`)
- `TestEnvFile` - для тестов переменных окружения (`/tmp/test-env-metrics.json`)
- `TestFlagFile` - для тестов CLI флагов (`/tmp/test-flag-metrics.json`)

#### Функции:
- `GetTestFilePath(name string)` - генерирует путь к тестовому файлу с произвольным именем

#### Использование:

```go
import "github.com/IgorKilipenko/metrical/internal/testutils"

func TestRepository(t *testing.T) {
    // Используем готовую константу
    repo := repository.NewInMemoryMetricsRepository(
        testutils.NewMockLogger(), 
        testutils.TestMetricsFile, 
        false,
    )
    
    // Или генерируем путь динамически
    customFile := testutils.GetTestFilePath("custom")
    // Результат: "/tmp/test-custom-metrics.json"
}
```

## 🎯 Преимущества

1. **DRY принцип**: Избегаем дублирования кода в тестах
2. **Консистентность**: Одинаковое поведение мока во всех тестах
3. **Простота**: Легко использовать и понимать
4. **Производительность**: Быстрые тесты без реального логирования

## 🔄 Миграция

Для миграции существующих тестов:

1. Заменить локальные определения `MockLogger` на импорт `testutils`
2. Заменить `newMockLogger()` на `testutils.NewMockLogger()`
3. Удалить дублированный код

### Пример миграции:

**До:**
```go
// MockLogger для тестирования
type MockLogger struct{}

func (m *MockLogger) SetLevel(level logger.LogLevel) {}
// ... остальные методы

func newMockLogger() logger.Logger {
    return &MockLogger{}
}

func TestSomething(t *testing.T) {
    mockLogger := newMockLogger()
    // ...
}
```

**После:**
```go
import "github.com/IgorKilipenko/metrical/internal/testutils"

func TestSomething(t *testing.T) {
    mockLogger := testutils.NewMockLogger()
    // ...
}
```

## 🏗️ Архитектура

```
internal/
├── testutils/
│   ├── logger.go      # MockLogger и утилиты
│   ├── files.go       # Константы и утилиты для тестовых файлов
│   └── README.md      # Документация
└── ...
```

## 📝 Планы развития

В будущем пакет может быть расширен:
- Mock для других интерфейсов
- Утилиты для создания тестовых данных
- Хелперы для интеграционных тестов
- Утилиты для бенчмарков
