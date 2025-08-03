package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "Valid config",
			config: &Config{
				ServerURL:      "http://localhost:8080",
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
			},
		},
		{
			name: "Different intervals",
			config: &Config{
				ServerURL:      "http://example.com:9090",
				PollInterval:   1 * time.Second,
				ReportInterval: 5 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(tt.config)

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
			name: "Standard configuration",
			config: &Config{
				ServerURL:      "http://localhost:8080",
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
			},
			expectedMinTotalMetrics: 28, // 27 runtime + 1 additional
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
			agent := NewAgent(tt.config)

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
			name: "10 goroutines",
			config: &Config{
				ServerURL:      "http://localhost:8080",
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
			},
			goroutines:         10,
			expectedMinMetrics: 28, // 27 runtime + 1 additional
		},
		{
			name: "5 goroutines",
			config: &Config{
				ServerURL:      "http://example.com:9090",
				PollInterval:   1 * time.Second,
				ReportInterval: 5 * time.Second,
			},
			goroutines:         5,
			expectedMinMetrics: 28,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(tt.config)

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

func TestAgent_GracefulShutdown(t *testing.T) {
	config := &Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   100 * time.Millisecond,
		ReportInterval: 200 * time.Millisecond,
	}

	agent := NewAgent(config)

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
