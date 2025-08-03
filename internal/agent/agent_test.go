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
			name:                    "Default configuration",
			config:                  NewConfig(),
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
			name:               "10 goroutines with default config",
			config:             NewConfig(),
			goroutines:         10,
			expectedMinMetrics: 28, // 27 runtime + 1 additional
		},
		{
			name:               "5 goroutines with custom config",
			config:             NewConfigWithURL("http://example.com:9090"),
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

func TestAgent_GracefulShutdown(t *testing.T) {
	config := NewConfig()
	// Переопределяем интервалы для быстрого тестирования
	config.PollInterval = 100 * time.Millisecond
	config.ReportInterval = 200 * time.Millisecond

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
