package agent

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// MetricValue структура для хранения метрики
type MetricValue struct {
	Value     float64
	Type      string // "gauge" или "counter"
	Timestamp time.Time
}

// Agent агент для сбора и отправки метрик
type Agent struct {
	config     *Config
	metrics    *Metrics
	mu         sync.RWMutex
	httpClient *http.Client
	done       chan struct{} // Канал для graceful shutdown
}

// NewAgent создает новый экземпляр агента
func NewAgent(config *Config) *Agent {
	return &Agent{
		config:  config,
		metrics: NewMetrics(),
		httpClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
		done: make(chan struct{}),
	}
}

// Stop останавливает агента gracefully
func (a *Agent) Stop() {
	log.Println("Stopping agent...")
	close(a.done)
}

// Run запускает агента
func (a *Agent) Run() {
	// Запускаем сбор метрик в отдельной горутине
	go a.pollMetrics()

	// Запускаем отправку метрик в отдельной горутине
	go a.reportMetrics()

	// Ждем сигнала завершения
	<-a.done
	log.Println("Agent stopped gracefully")
}

// pollMetrics собирает метрики из runtime
func (a *Agent) pollMetrics() {
	ticker := time.NewTicker(a.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.collectMetrics()
		case <-a.done:
			log.Println("Polling stopped")
			return
		}
	}
}

// collectMetrics собирает все метрики
func (a *Agent) collectMetrics() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Собираем метрики из runtime
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Заполняем runtime метрики
	FillRuntimeMetrics(a.metrics, memStats)

	// Заполняем дополнительные метрики
	FillAdditionalMetrics(a.metrics)

	// Обновляем counter метрики
	UpdateCounterMetrics(a.metrics)

	totalMetrics := len(a.metrics.Gauges) + len(a.metrics.Counters)
	log.Printf("Collected %d metrics", totalMetrics)
}

// reportMetrics отправляет метрики на сервер
func (a *Agent) reportMetrics() {
	ticker := time.NewTicker(a.config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.sendMetrics()
		case <-a.done:
			log.Println("Reporting stopped")
			return
		}
	}
}

// sendMetrics отправляет все метрики на сервер
func (a *Agent) sendMetrics() {
	a.mu.RLock()
	metrics := a.metrics.GetAllMetrics()
	a.mu.RUnlock()

	for name, value := range metrics {
		var metricType string
		var stringValue string

		switch v := value.(type) {
		case float64:
			metricType = MetricTypeGauge
			stringValue = strconv.FormatFloat(v, 'f', -1, 64)
		case int64:
			metricType = MetricTypeCounter
			stringValue = strconv.FormatInt(v, 10)
		default:
			log.Printf("Unknown metric type for %s: %T", name, value)
			continue
		}

		url := fmt.Sprintf("%s/update/%s/%s/%s", a.config.ServerURL, metricType, name, stringValue)
		resp, err := a.httpClient.Post(url, "text/plain", nil)
		if err != nil {
			log.Printf("Error sending metric %s: %v", name, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("Sent metric %s = %s, status: %d", name, stringValue, resp.StatusCode)
		} else {
			log.Printf("Failed to send metric %s, status: %d", name, resp.StatusCode)
		}
	}
}
