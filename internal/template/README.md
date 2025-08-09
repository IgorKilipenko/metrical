# Template Package

Пакет для работы с HTML шаблонами метрик.

## Описание

Пакет `template` предоставляет функциональность для генерации HTML страниц с отображением метрик. Основная цель - отделить логику формирования HTML от HTTP обработчиков.

## Компоненты

### MetricsData

Структура данных для передачи метрик в шаблон:

```go
type MetricsData struct {
    Gauges       models.GaugeMetrics  // Gauge метрики
    Counters     models.CounterMetrics // Counter метрики
    GaugeCount   int                  // Количество gauge метрик
    CounterCount int                  // Количество counter метрик
}
```

### Архитектура шаблонов

```mermaid
graph TB
    subgraph "Template Engine"
        MT[MetricsTemplate]
        HTML[HTML Template]
        CSS[CSS Styles]
    end
    
    subgraph "Data Flow"
        MD[MetricsData]
        GM[GaugeMetrics]
        CM[CounterMetrics]
    end
    
    subgraph "Output"
        HTML_OUT[HTML Output]
        HTTP_RESP[HTTP Response]
    end
    
    MD --> GM
    MD --> CM
    MT --> HTML
    MT --> CSS
    MT --> MD
    MT --> HTML_OUT
    HTML_OUT --> HTTP_RESP
    
    style MT fill:#f3e5f5
    style HTML fill:#e3f2fd
    style CSS fill:#e3f2fd
    style MD fill:#e8f5e8
    style GM fill:#e1f5fe
    style CM fill:#e1f5fe
    style HTML_OUT fill:#fff3e0
    style HTTP_RESP fill:#fff3e0
```

### Поток генерации HTML

```mermaid
sequenceDiagram
    participant Handler
    participant Template
    participant Data
    participant HTML
    
    Handler->>Template: Execute(data)
    Template->>Data: Prepare MetricsData
    Data-->>Template: Structured Data
    Template->>HTML: Generate HTML
    HTML-->>Template: HTML Bytes
    Template-->>Handler: HTML Response
    
    Note over Template,HTML: Включает CSS стили
    Note over Data: Gauge + Counter метрики
```

### MetricsTemplate

Основной класс для работы с HTML шаблонами:

```go
type MetricsTemplate struct {
    template *template.Template
}
```

## Использование

```go
// Создание шаблона
mt, err := template.NewMetricsTemplate()
if err != nil {
    log.Fatal(err)
}

// Подготовка данных
data := template.MetricsData{
    Gauges: models.GaugeMetrics{
        "temperature": 23.5,
        "memory":      1024.0,
    },
    Counters: models.CounterMetrics{
        "requests": 100,
        "errors":   5,
    },
    GaugeCount:   2,
    CounterCount: 2,
}

// Генерация HTML
htmlBytes, err := mt.Execute(data)
if err != nil {
    log.Fatal(err)
}

// Отправка в HTTP ответ
w.Header().Set("Content-Type", "text/html; charset=utf-8")
w.Write(htmlBytes)
```

## HTML Шаблон

Шаблон включает:
- Современный CSS дизайн
- Отдельные секции для gauge и counter метрик
- Счетчики метрик
- Сообщения при отсутствии метрик
- Адаптивную верстку

## Преимущества

1. **Разделение ответственности** - HTML логика отделена от HTTP обработчиков
2. **Переиспользование** - шаблон можно использовать в разных частях приложения
3. **Тестируемость** - легко тестировать генерацию HTML
4. **Поддержка** - простое изменение дизайна и структуры
5. **Производительность** - шаблон парсится один раз при создании

## Тестирование

```bash
go test -v ./internal/template
```
