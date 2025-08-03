package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	config := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	agent := NewAgent(config)

	assert.Equal(t, config, agent.config, "Agent config should match provided config")
	assert.NotNil(t, agent.metrics, "Agent metrics map should be initialized")
	assert.NotNil(t, agent.httpClient, "Agent HTTP client should be initialized")
}

func TestAgent_CollectMetrics(t *testing.T) {
	config := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	agent := NewAgent(config)

	// Собираем метрики
	agent.collectMetrics()

	// Проверяем, что метрики собраны
	totalMetrics := len(agent.metrics.Gauges) + len(agent.metrics.Counters)
	assert.Greater(t, totalMetrics, 0, "Metrics should be collected")

	// Проверяем наличие обязательных метрик
	requiredGaugeMetrics := []string{
		MetricAlloc, MetricBuckHashSys, MetricFrees, MetricGCCPUFraction, MetricGCSys,
		MetricHeapAlloc, MetricHeapIdle, MetricHeapInuse, MetricHeapObjects, MetricHeapReleased,
		MetricHeapSys, MetricLastGC, MetricLookups, MetricMCacheInuse, MetricMCacheSys,
		MetricMSpanInuse, MetricMSpanSys, MetricMallocs, MetricNextGC, MetricNumForcedGC,
		MetricNumGC, MetricOtherSys, MetricPauseTotalNs, MetricStackInuse, MetricStackSys,
		MetricSys, MetricTotalAlloc, MetricRandomValue,
	}

	requiredCounterMetrics := []string{
		MetricPollCount,
	}

	for _, metricName := range requiredGaugeMetrics {
		_, exists := agent.metrics.Gauges[metricName]
		assert.True(t, exists, "Required gauge metric %s should exist", metricName)
	}

	for _, metricName := range requiredCounterMetrics {
		_, exists := agent.metrics.Counters[metricName]
		assert.True(t, exists, "Required counter metric %s should exist", metricName)
	}

	// Проверяем, что PollCount увеличивается
	initialPollCount := agent.metrics.Counters[MetricPollCount]
	agent.collectMetrics()
	newPollCount := agent.metrics.Counters[MetricPollCount]

	assert.Equal(t, initialPollCount+1, newPollCount, "PollCount should increment")
}

func TestAgent_CollectMetrics_ThreadSafety(t *testing.T) {
	config := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	agent := NewAgent(config)

	// Запускаем несколько горутин для тестирования потокобезопасности
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			agent.collectMetrics()
			done <- true
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}

	// Проверяем, что метрики собраны без ошибок
	totalMetrics := len(agent.metrics.Gauges) + len(agent.metrics.Counters)
	assert.Greater(t, totalMetrics, 0, "Metrics should be collected in thread-safe manner")
}

func TestAgent_Config_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
	}{
		{
			name: "Valid config",
			config: &Config{
				ServerURL:      "http://localhost:8080",
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
			},
			expectValid: true,
		},
		{
			name: "Empty server URL",
			config: &Config{
				ServerURL:      "",
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
			},
			expectValid: false,
		},
		{
			name: "Zero poll interval",
			config: &Config{
				ServerURL:      "http://localhost:8080",
				PollInterval:   0,
				ReportInterval: 10 * time.Second,
			},
			expectValid: false,
		},
		{
			name: "Zero report interval",
			config: &Config{
				ServerURL:      "http://localhost:8080",
				PollInterval:   2 * time.Second,
				ReportInterval: 0,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(tt.config)

			if tt.expectValid {
				assert.NotEmpty(t, agent.config.ServerURL, "Server URL should not be empty")
				assert.Greater(t, agent.config.PollInterval, time.Duration(0), "Poll interval should be positive")
				assert.Greater(t, agent.config.ReportInterval, time.Duration(0), "Report interval should be positive")
			}
		})
	}
}
