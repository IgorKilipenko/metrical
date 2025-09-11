package agent

import (
	"net/http"

	models "github.com/IgorKilipenko/metrical/internal/model"
)

// MetricsCollector интерфейс для сбора метрик
type MetricsCollector interface {
	Collect() map[string]any
	GetAllMetrics() map[string]any
}

// MetricsSender интерфейс для отправки метрик
type MetricsSender interface {
	Send(metrics map[string]any) error
	SendSingle(name string, value interface{}) error
}

// MetricPreparer интерфейс для подготовки метрик
type MetricPreparer interface {
	PrepareJSON(name string, value any) (*models.Metrics, error)
	PrepareInfo(name string, value any) (*MetricInfo, error)
}

// HTTPRequestBuilder интерфейс для создания HTTP запросов
type HTTPRequestBuilder interface {
	BuildJSONRequest(metric *models.Metrics) (*http.Request, error)
	BuildURLRequest(url string) (*http.Request, error)
}
