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

func TestAgent_PrepareMetricInfo(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	agent := NewAgent(NewConfig(), mockLogger)

	tests := []struct {
		name        string
		metricName  string
		value       interface{}
		expectError bool
		expected    *MetricInfo
	}{
		{
			name:       "Gauge metric",
			metricName: "test_gauge",
			value:      123.45,
			expected: &MetricInfo{
				Name:  "test_gauge",
				Type:  MetricTypeGauge,
				Value: "123.45",
				URL:   "http://localhost:8080/update/gauge/test_gauge/123.45",
			},
		},
		{
			name:       "Counter metric",
			metricName: "test_counter",
			value:      int64(678),
			expected: &MetricInfo{
				Name:  "test_counter",
				Type:  MetricTypeCounter,
				Value: "678",
				URL:   "http://localhost:8080/update/counter/test_counter/678",
			},
		},
		{
			name:        "Unknown type",
			metricName:  "test_unknown",
			value:       "string_value",
			expectError: true,
		},
		{
			name:       "URL without protocol",
			metricName: "test_gauge",
			value:      123.45,
			expected: &MetricInfo{
				Name:  "test_gauge",
				Type:  MetricTypeGauge,
				Value: "123.45",
				URL:   "http://localhost:8080/update/gauge/test_gauge/123.45",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Для теста с URL без протокола
			if tt.name == "URL without protocol" {
				agent.config.ServerURL = "localhost:8080"
			} else {
				agent.config.ServerURL = "http://localhost:8080"
			}

			result, err := agent.prepareMetricInfo(tt.metricName, tt.value)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Type, result.Type)
				assert.Equal(t, tt.expected.Value, result.Value)
				assert.Equal(t, tt.expected.URL, result.URL)
			}
		})
	}
}

func TestAgent_PrepareMetricInfo_URLHandling(t *testing.T) {
	mockLogger := testutils.NewMockLogger()
	agent := NewAgent(NewConfig(), mockLogger)

	tests := []struct {
		name      string
		serverURL string
		expected  string
	}{
		{
			name:      "URL with http protocol",
			serverURL: "http://example.com:8080",
			expected:  "http://example.com:8080/update/gauge/test/123.45",
		},
		{
			name:      "URL with https protocol",
			serverURL: "https://example.com:8080",
			expected:  "https://example.com:8080/update/gauge/test/123.45",
		},
		{
			name:      "URL without protocol",
			serverURL: "example.com:8080",
			expected:  "http://example.com:8080/update/gauge/test/123.45",
		},
		{
			name:      "URL with localhost",
			serverURL: "localhost:9091",
			expected:  "http://localhost:9091/update/gauge/test/123.45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent.config.ServerURL = tt.serverURL

			result, err := agent.prepareMetricInfo("test", 123.45)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, result.URL)
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
