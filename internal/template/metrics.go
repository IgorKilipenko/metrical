package template

import (
	"bytes"
	"text/template"

	models "github.com/IgorKilipenko/metrical/internal/model"
)

// MetricsData содержит данные для отображения метрик
type MetricsData struct {
	Gauges       models.GaugeMetrics
	Counters     models.CounterMetrics
	GaugeCount   int
	CounterCount int
}

// HTML шаблон для отображения метрик
const metricsHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metric-section { margin-bottom: 30px; }
        .metric-item { 
            padding: 8px; 
            margin: 4px 0; 
            background-color: #f5f5f5; 
            border-radius: 4px;
            display: flex;
            justify-content: space-between;
        }
        .metric-name { font-weight: bold; }
        .metric-value { color: #666; }
        h2 { color: #333; border-bottom: 2px solid #ddd; padding-bottom: 10px; }
        .header { text-align: center; margin-bottom: 30px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Metrics Dashboard</h1>
        <p>Current metrics values</p>
    </div>
    
    <div class="metric-section">
        <h2>Gauge Metrics ({{.GaugeCount}})</h2>
        {{range $name, $value := .Gauges}}
        <div class="metric-item">
            <span class="metric-name">{{$name}}</span>
            <span class="metric-value">{{$value}}</span>
        </div>
        {{else}}
        <p><em>No gauge metrics available</em></p>
        {{end}}
    </div>
    
    <div class="metric-section">
        <h2>Counter Metrics ({{.CounterCount}})</h2>
        {{range $name, $value := .Counters}}
        <div class="metric-item">
            <span class="metric-name">{{$name}}</span>
            <span class="metric-value">{{$value}}</span>
        </div>
        {{else}}
        <p><em>No counter metrics available</em></p>
        {{end}}
    </div>
</body>
</html>`

// MetricsTemplate предоставляет методы для работы с HTML шаблонами метрик
type MetricsTemplate struct {
	template *template.Template
}

// NewMetricsTemplate создает новый экземпляр шаблона метрик
func NewMetricsTemplate() (*MetricsTemplate, error) {
	tmpl, err := template.New("metrics").Parse(metricsHTMLTemplate)
	if err != nil {
		return nil, err
	}

	return &MetricsTemplate{
		template: tmpl,
	}, nil
}

// Execute выполняет шаблон с переданными данными
func (mt *MetricsTemplate) Execute(data MetricsData) ([]byte, error) {
	var buf bytes.Buffer
	err := mt.template.Execute(&buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
