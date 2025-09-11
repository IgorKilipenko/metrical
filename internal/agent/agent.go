package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	models "github.com/IgorKilipenko/metrical/internal/model"
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
	httpClient HTTPClient
	done       chan struct{} // Канал для graceful shutdown
	logger     logger.Logger
}

// NewAgent создает новый экземпляр агента
func NewAgent(config *Config, agentLogger logger.Logger) *Agent {
	if agentLogger == nil {
		agentLogger = logger.NewSlogLogger()
	}

	// Создаем базовый HTTP клиент
	baseClient := &http.Client{
		Timeout: DefaultHTTPTimeout,
	}

	// Обертываем в retry клиент
	retryClient := NewRetryHTTPClient(baseClient, DefaultMaxRetries, DefaultRetryDelay, agentLogger)

	return &Agent{
		config:     config,
		metrics:    NewMetrics(),
		httpClient: retryClient,
		done:       make(chan struct{}),
		logger:     agentLogger,
	}
}

// Stop останавливает агента gracefully
func (a *Agent) Stop() {
	a.logger.Info("stopping agent")
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
	a.logger.Info("agent stopped gracefully")
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
			a.logger.Info("polling stopped")
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
		a.logger.Info("collected metrics",
			"total", totalMetrics,
			"gauges", len(a.metrics.Gauges),
			"counters", len(a.metrics.Counters))
	} else {
		a.logger.Info("collected metrics", "total", totalMetrics)
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
			a.logger.Info("reporting stopped")
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
		if err := a.sendSingleMetricJSON(name, value); err != nil {
			errorCount++
			// Логируем ошибки только если включено подробное логирование
			if a.config.VerboseLogging {
				a.logger.Error("error sending metric", "name", name, "error", err)
			}
		} else {
			successCount++
		}
	}

	// Логируем итоговую статистику
	if errorCount > 0 {
		a.logger.Warn("sent metrics with errors",
			"successful", successCount,
			"failed", errorCount)
	} else {
		a.logger.Info("successfully sent metrics", "count", successCount)
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
	resp, err := a.httpClient.Post(url, "text/plain", nil)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// sendHTTPRequestWithGzip выполняет HTTP запрос с gzip сжатием и retry логикой
func (a *Agent) sendHTTPRequestWithGzip(url string) error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки для gzip
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request with gzip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// sendSingleMetric отправляет одну метрику с retry логикой
func (a *Agent) sendSingleMetric(name string, value interface{}) error {
	// Подготавливаем информацию о метрике
	metricInfo, err := a.prepareMetricInfo(name, value)
	if err != nil {
		return err
	}

	// Отправляем HTTP запрос
	if err := a.sendHTTPRequestWithGzip(metricInfo.URL); err != nil {
		return fmt.Errorf("failed to send metric %s: %w", name, err)
	}

	// Логируем успешную отправку
	if a.config.VerboseLogging {
		a.logger.Debug("sent metric successfully",
			"name", metricInfo.Name,
			"value", metricInfo.Value,
			"status", 200)
	}

	return nil
}

// sendSingleMetricJSON отправляет одну метрику в JSON формате
func (a *Agent) sendSingleMetricJSON(name string, value interface{}) error {
	// Подготавливаем метрику в JSON формате
	metric, err := a.prepareMetricJSON(name, value)
	if err != nil {
		return err
	}

	// Отправляем HTTP запрос
	if err := a.sendJSONRequest(metric); err != nil {
		return fmt.Errorf("failed to send metric %s: %w", name, err)
	}

	// Логируем успешную отправку
	if a.config.VerboseLogging {
		a.logger.Debug("sent metric successfully",
			"name", metric.ID,
			"type", metric.MType,
			"status", 200)
	}

	return nil
}

// prepareMetricJSON подготавливает метрику в JSON формате
func (a *Agent) prepareMetricJSON(name string, value interface{}) (*models.Metrics, error) {
	var metric models.Metrics
	metric.ID = name

	switch v := value.(type) {
	case float64:
		metric.MType = "gauge"
		metric.Value = &v
	case int64:
		metric.MType = "counter"
		metric.Delta = &v
	default:
		return nil, fmt.Errorf("unknown metric type for %s: %T", name, value)
	}

	return &metric, nil
}

// compressData сжимает данные с помощью gzip
func (a *Agent) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	if _, err := gzWriter.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// sendJSONRequest отправляет JSON запрос на сервер
func (a *Agent) sendJSONRequest(metric *models.Metrics) error {
	// Убеждаемся, что URL содержит протокол
	serverURL := a.config.ServerURL
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	url := fmt.Sprintf("%s/update", serverURL)

	// Кодируем метрику в JSON
	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	// Сжимаем данные
	compressedData, err := a.compressData(jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	// Создаем запрос
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(compressedData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	// Выполняем запрос с retry логикой
	return a.sendHTTPRequestWithRetry(req)
}

// sendHTTPRequestWithRetry выполняет HTTP запрос с retry логикой
func (a *Agent) sendHTTPRequestWithRetry(req *http.Request) error {
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request with retry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}
