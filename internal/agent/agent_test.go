package agent

import (
	"testing"
	"time"

	"github.com/IgorKilipenko/metrical/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "Default config",
			config: NewConfig(),
		},
		{
			name:   "Custom URL config",
			config: NewConfigWithURL("http://example.com:9090"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := testutils.NewMockLogger()
			agent := NewAgent(tt.config, mockLogger)

			assert.Equal(t, tt.config, agent.config, "Agent config should match provided config")
			assert.NotNil(t, agent.metrics, "Agent metrics struct should be initialized")
			assert.NotNil(t, agent.metrics.Gauges, "Agent gauges map should be initialized")
			assert.NotNil(t, agent.metrics.Counters, "Agent counters map should be initialized")
			assert.NotNil(t, agent.httpClient, "Agent HTTP client should be initialized")
			assert.NotNil(t, agent.done, "Agent done channel should be initialized")
		})
	}
}

func TestAgent_CollectMetrics(t *testing.T) {
	tests := []struct {
		name                    string
		config                  *Config
		expectedMinTotalMetrics int
		expectedGaugeMetrics    []string
		expectedCounterMetrics  []string
	}{
		{
			name:                    "Default configuration",
			config:                  NewConfig(),
			expectedMinTotalMetrics: 29, // 27 runtime + 1 additional + 1 counter
			expectedGaugeMetrics: []string{
				MetricAlloc, MetricBuckHashSys, MetricFrees, MetricGCCPUFraction, MetricGCSys,
				MetricHeapAlloc, MetricHeapIdle, MetricHeapInuse, MetricHeapObjects, MetricHeapReleased,
				MetricHeapSys, MetricLastGC, MetricLookups, MetricMCacheInuse, MetricMCacheSys,
				MetricMSpanInuse, MetricMSpanSys, MetricMallocs, MetricNextGC, MetricNumForcedGC,
				MetricNumGC, MetricOtherSys, MetricPauseTotalNs, MetricStackInuse, MetricStackSys,
				MetricSys, MetricTotalAlloc, MetricRandomValue,
			},
			expectedCounterMetrics: []string{
				MetricPollCount,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := testutils.NewMockLogger()
			agent := NewAgent(tt.config, mockLogger)

			// Собираем метрики
			agent.collectMetrics()

			// Проверяем, что метрики собраны
			totalMetrics := len(agent.metrics.Gauges) + len(agent.metrics.Counters)
			assert.GreaterOrEqual(t, totalMetrics, tt.expectedMinTotalMetrics, "Metrics should be collected")

			// Проверяем наличие обязательных gauge метрик
			for _, metricName := range tt.expectedGaugeMetrics {
				_, exists := agent.metrics.Gauges[metricName]
				assert.True(t, exists, "Required gauge metric %s should exist", metricName)
			}

			// Проверяем наличие обязательных counter метрик
			for _, metricName := range tt.expectedCounterMetrics {
				_, exists := agent.metrics.Counters[metricName]
				assert.True(t, exists, "Required counter metric %s should exist", metricName)
			}

			// Проверяем, что PollCount увеличивается
			initialPollCount := agent.metrics.Counters[MetricPollCount]
			agent.collectMetrics()
			newPollCount := agent.metrics.Counters[MetricPollCount]

			assert.Equal(t, initialPollCount+1, newPollCount, "PollCount should increment")
		})
	}
}

func TestAgent_CollectMetrics_ThreadSafety(t *testing.T) {
	tests := []struct {
		name               string
		config             *Config
		goroutines         int
		expectedMinMetrics int
	}{
		{
			name:               "10 goroutines with default config",
			config:             NewConfig(),
			goroutines:         10,
			expectedMinMetrics: 29, // 27 runtime + 1 additional + 1 counter
		},
		{
			name:               "5 goroutines with custom config",
			config:             NewConfigWithURL("http://example.com:9090"),
			goroutines:         5,
			expectedMinMetrics: 29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := testutils.NewMockLogger()
			agent := NewAgent(tt.config, mockLogger)

			// Запускаем несколько горутин для тестирования потокобезопасности
			done := make(chan bool, tt.goroutines)
			for i := 0; i < tt.goroutines; i++ {
				go func() {
					agent.collectMetrics()
					done <- true
				}()
			}

			// Ждем завершения всех горутин
			for i := 0; i < tt.goroutines; i++ {
				<-done
			}

			// Проверяем, что метрики собраны без ошибок
			totalMetrics := len(agent.metrics.Gauges) + len(agent.metrics.Counters)
			assert.GreaterOrEqual(t, totalMetrics, tt.expectedMinMetrics,
				"Metrics should be collected in thread-safe manner")
		})
	}
}

func TestAgent_GracefulShutdown(t *testing.T) {
	config := NewConfig()
	// Переопределяем интервалы для быстрого тестирования
	config.PollInterval = 100 * time.Millisecond
	config.ReportInterval = 200 * time.Millisecond

	mockLogger := testutils.NewMockLogger()
	agent := NewAgent(config, mockLogger)

	// Запускаем агент в горутине
	go agent.Run()

	// Даем время на запуск и сбор метрик
	time.Sleep(150 * time.Millisecond)

	// Проверяем, что метрики собираются
	initialMetrics := len(agent.metrics.Gauges) + len(agent.metrics.Counters)
	assert.Greater(t, initialMetrics, 0, "Metrics should be collected")

	// Останавливаем агента
	agent.Stop()

	// Даем время на завершение
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что агент остановился корректно
	// (канал done должен быть закрыт)
	select {
	case <-agent.done:
		// Ожидаемо - канал закрыт
	default:
		t.Error("Agent should be stopped")
	}
}

func TestAgent_sendSingleMetricJSON(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	config := &Config{
		ServerURL:      "localhost:8080",
		PollInterval:   1 * time.Second,
		ReportInterval: 1 * time.Second,
		VerboseLogging: true,
	}

	agent := NewAgent(config, mockLogger)

	tests := []struct {
		name        string
		metricName  string
		metricValue interface{}
		expectError bool
	}{
		{
			name:        "unsupported metric type",
			metricName:  "TestMetric",
			metricValue: "string",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := agent.sendSingleMetricJSON(tt.metricName, tt.metricValue)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAgent_prepareMetricJSON(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	config := &Config{
		ServerURL:      "localhost:8080",
		PollInterval:   1 * time.Second,
		ReportInterval: 1 * time.Second,
		VerboseLogging: true,
	}

	agent := NewAgent(config, mockLogger)

	tests := []struct {
		name         string
		metricName   string
		metricValue  interface{}
		expectError  bool
		expectedType string
	}{
		{
			name:         "gauge metric",
			metricName:   "TestGauge",
			metricValue:  42.5,
			expectError:  false,
			expectedType: "gauge",
		},
		{
			name:         "counter metric",
			metricName:   "TestCounter",
			metricValue:  int64(100),
			expectError:  false,
			expectedType: "counter",
		},
		{
			name:        "unsupported metric type",
			metricName:  "TestMetric",
			metricValue: "string",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := agent.prepareMetricJSON(tt.metricName, tt.metricValue)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, metric)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metric)
				assert.Equal(t, tt.metricName, metric.ID)
				assert.Equal(t, tt.expectedType, metric.MType)
			}
		})
	}
}
