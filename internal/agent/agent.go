package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Config конфигурация агента
type Config struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

// Agent агент для сбора и отправки метрик
type Agent struct {
	config     *Config
	metrics    map[string]interface{}
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewAgent создает новый экземпляр агента
func NewAgent(config *Config) *Agent {
	return &Agent{
		config:  config,
		metrics: make(map[string]interface{}),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Run запускает агент
func (a *Agent) Run() {
	// Запускаем сбор метрик в отдельной горутине
	go a.pollMetrics()

	// Запускаем отправку метрик в отдельной горутине
	go a.reportMetrics()

	// Ждем бесконечно
	select {}
}

// pollMetrics собирает метрики из runtime
func (a *Agent) pollMetrics() {
	ticker := time.NewTicker(a.config.PollInterval)
	defer ticker.Stop()

	for range ticker.C {
		a.collectMetrics()
	}
}

// collectMetrics собирает все метрики
func (a *Agent) collectMetrics() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Собираем метрики из runtime
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Gauge метрики из runtime
	a.metrics["Alloc"] = float64(memStats.Alloc)
	a.metrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	a.metrics["Frees"] = float64(memStats.Frees)
	a.metrics["GCCPUFraction"] = memStats.GCCPUFraction
	a.metrics["GCSys"] = float64(memStats.GCSys)
	a.metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	a.metrics["HeapIdle"] = float64(memStats.HeapIdle)
	a.metrics["HeapInuse"] = float64(memStats.HeapInuse)
	a.metrics["HeapObjects"] = float64(memStats.HeapObjects)
	a.metrics["HeapReleased"] = float64(memStats.HeapReleased)
	a.metrics["HeapSys"] = float64(memStats.HeapSys)
	a.metrics["LastGC"] = float64(memStats.LastGC)
	a.metrics["Lookups"] = float64(memStats.Lookups)
	a.metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	a.metrics["MCacheSys"] = float64(memStats.MCacheSys)
	a.metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	a.metrics["MSpanSys"] = float64(memStats.MSpanSys)
	a.metrics["Mallocs"] = float64(memStats.Mallocs)
	a.metrics["NextGC"] = float64(memStats.NextGC)
	a.metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	a.metrics["NumGC"] = float64(memStats.NumGC)
	a.metrics["OtherSys"] = float64(memStats.OtherSys)
	a.metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	a.metrics["StackInuse"] = float64(memStats.StackInuse)
	a.metrics["StackSys"] = float64(memStats.StackSys)
	a.metrics["Sys"] = float64(memStats.Sys)
	a.metrics["TotalAlloc"] = float64(memStats.TotalAlloc)

	// Дополнительные метрики
	a.metrics["RandomValue"] = rand.Float64()

	// Counter метрики
	if pollCount, exists := a.metrics["PollCount"]; exists {
		a.metrics["PollCount"] = pollCount.(int64) + 1
	} else {
		a.metrics["PollCount"] = int64(1)
	}

	log.Printf("Collected %d metrics", len(a.metrics))
}

// reportMetrics отправляет метрики на сервер
func (a *Agent) reportMetrics() {
	ticker := time.NewTicker(a.config.ReportInterval)
	defer ticker.Stop()

	for range ticker.C {
		a.sendMetrics()
	}
}

// sendMetrics отправляет все метрики на сервер
func (a *Agent) sendMetrics() {
	a.mu.RLock()
	metrics := make(map[string]interface{}, len(a.metrics))
	for k, v := range a.metrics {
		metrics[k] = v
	}
	a.mu.RUnlock()

	for name, value := range metrics {
		var metricType string
		var stringValue string

		switch v := value.(type) {
		case float64:
			metricType = "gauge"
			stringValue = strconv.FormatFloat(v, 'f', -1, 64)
		case int64:
			metricType = "counter"
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
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("Sent metric %s = %s, status: %d", name, stringValue, resp.StatusCode)
		} else {
			log.Printf("Failed to send metric %s, status: %d", name, resp.StatusCode)
		}
	}
}
