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

// Константы для retry логики
const (
	DefaultMaxRetries = 2
	DefaultRetryDelay = 100 * time.Millisecond
)

// MetricValue структура для хранения метрики
type MetricValue struct {
	Value     float64
	Type      string // "gauge" или "counter"
	Timestamp time.Time
}

// MetricInfo структура для хранения информации о метрике для отправки
type MetricInfo struct {
	Name  string
	Type  string
	Value string
	URL   string
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
	if a.config.VerboseLogging {
		log.Printf("Collected %d metrics (gauges: %d, counters: %d)",
			totalMetrics, len(a.metrics.Gauges), len(a.metrics.Counters))
	} else {
		log.Printf("Collected %d metrics", totalMetrics)
	}
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

// prepareMetricInfo подготавливает информацию о метрике для отправки
func (a *Agent) prepareMetricInfo(name string, value interface{}) (*MetricInfo, error) {
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
		return nil, fmt.Errorf("unknown metric type for %s: %T", name, value)
	}

	// Убеждаемся, что URL содержит протокол
	serverURL := a.config.ServerURL
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	url := fmt.Sprintf("%s/update/%s/%s/%s", serverURL, metricType, name, stringValue)

	return &MetricInfo{
		Name:  name,
		Type:  metricType,
		Value: stringValue,
		URL:   url,
	}, nil
}

// sendHTTPRequest выполняет HTTP запрос с retry логикой
func (a *Agent) sendHTTPRequest(url string) error {
	for attempt := 1; attempt <= DefaultMaxRetries; attempt++ {
		resp, err := a.httpClient.Post(url, "text/plain", nil)
		if err != nil {
			// Если это последняя попытка, возвращаем ошибку
			if attempt == DefaultMaxRetries {
				return fmt.Errorf("failed after %d attempts: %w", DefaultMaxRetries, err)
			}
			// Небольшая задержка перед повторной попыткой
			time.Sleep(DefaultRetryDelay)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		} else {
			// Читаем тело ответа для лучшей диагностики
			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			if attempt == DefaultMaxRetries {
				return fmt.Errorf("server returned status %d: %s", resp.StatusCode, bodyStr)
			}
			time.Sleep(DefaultRetryDelay)
		}
	}

	return fmt.Errorf("failed to send request after %d attempts", DefaultMaxRetries)
}

// sendSingleMetric отправляет одну метрику с retry логикой
func (a *Agent) sendSingleMetric(name string, value interface{}) error {
	// Подготавливаем информацию о метрике
	metricInfo, err := a.prepareMetricInfo(name, value)
	if err != nil {
		return err
	}

	// Отправляем HTTP запрос
	if err := a.sendHTTPRequest(metricInfo.URL); err != nil {
		return fmt.Errorf("failed to send metric %s: %w", name, err)
	}

	// Логируем успешную отправку
	if a.config.VerboseLogging {
		log.Printf("Sent metric %s = %s, status: 200", metricInfo.Name, metricInfo.Value)
	}

	return nil
}
