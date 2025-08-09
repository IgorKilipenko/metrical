package agent

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
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

	successCount := 0
	errorCount := 0

	for name, value := range metrics {
		if err := a.sendSingleMetric(name, value); err != nil {
			errorCount++
			// Логируем ошибки только если включено подробное логирование
			if a.config.VerboseLogging {
				log.Printf("Error sending metric %s: %v", name, err)
			}
		} else {
			successCount++
		}
	}

	// Логируем итоговую статистику
	if errorCount > 0 {
		log.Printf("Sent %d metrics successfully, %d failed", successCount, errorCount)
	} else {
		log.Printf("Successfully sent %d metrics", successCount)
	}
}

// sendSingleMetric отправляет одну метрику с retry логикой
func (a *Agent) sendSingleMetric(name string, value interface{}) error {
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
		return fmt.Errorf("unknown metric type for %s: %T", name, value)
	}

	// Убеждаемся, что URL содержит протокол
	serverURL := a.config.ServerURL
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	url := fmt.Sprintf("%s/update/%s/%s/%s", serverURL, metricType, name, stringValue)

	// Retry логика для обработки временных проблем с сервером
	maxRetries := 2
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := a.httpClient.Post(url, "text/plain", nil)
		if err != nil {
			// Если это последняя попытка, возвращаем ошибку
			if attempt == maxRetries {
				return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
			}
			// Небольшая задержка перед повторной попыткой
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			if a.config.VerboseLogging {
				log.Printf("Sent metric %s = %s, status: %d", name, stringValue, resp.StatusCode)
			}
			return nil
		} else {
			if attempt == maxRetries {
				return fmt.Errorf("server returned status %d", resp.StatusCode)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	return fmt.Errorf("failed to send metric %s after %d attempts", name, maxRetries)
}
