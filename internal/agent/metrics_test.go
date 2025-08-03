package agent

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()

	assert.NotNil(t, metrics.Gauges, "Gauges map should be initialized")
	assert.NotNil(t, metrics.Counters, "Counters map should be initialized")
	assert.Len(t, metrics.Gauges, 0, "Gauges should be empty initially")
	assert.Len(t, metrics.Counters, 0, "Counters should be empty initially")
}

func TestMetrics_GetAllMetrics(t *testing.T) {
	tests := []struct {
		name            string
		gaugeMetrics    map[string]float64
		counterMetrics  map[string]int64
		expectedCount   int
		expectedGauge   float64
		expectedCounter int64
	}{
		{
			name: "Single gauge and counter",
			gaugeMetrics: map[string]float64{
				"test_gauge": 123.45,
			},
			counterMetrics: map[string]int64{
				"test_counter": 678,
			},
			expectedCount:   2,
			expectedGauge:   123.45,
			expectedCounter: 678,
		},
		{
			name: "Multiple metrics",
			gaugeMetrics: map[string]float64{
				"gauge1": 1.1,
				"gauge2": 2.2,
			},
			counterMetrics: map[string]int64{
				"counter1": 100,
				"counter2": 200,
			},
			expectedCount:   4,
			expectedGauge:   1.1,
			expectedCounter: 100,
		},
		{
			name:           "Empty metrics",
			gaugeMetrics:   map[string]float64{},
			counterMetrics: map[string]int64{},
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()

			// Добавляем тестовые метрики
			for name, value := range tt.gaugeMetrics {
				metrics.Gauges[name] = value
			}
			for name, value := range tt.counterMetrics {
				metrics.Counters[name] = value
			}

			// Получаем все метрики
			allMetrics := metrics.GetAllMetrics()

			// Проверяем количество метрик
			assert.Len(t, allMetrics, tt.expectedCount, "Should have correct number of metrics")

			// Проверяем конкретные значения для непустых тестов
			if tt.expectedCount > 0 {
				if len(tt.gaugeMetrics) > 0 {
					for name, expectedValue := range tt.gaugeMetrics {
						assert.Equal(t, expectedValue, allMetrics[name], "Gauge metric %s should match", name)
					}
				}
				if len(tt.counterMetrics) > 0 {
					for name, expectedValue := range tt.counterMetrics {
						assert.Equal(t, expectedValue, allMetrics[name], "Counter metric %s should match", name)
					}
				}
			}
		})
	}
}

func TestFillRuntimeMetrics(t *testing.T) {
	tests := []struct {
		name     string
		memStats runtime.MemStats
		expected map[string]float64
	}{
		{
			name: "Basic metrics",
			memStats: runtime.MemStats{
				Alloc:         1024,
				HeapAlloc:     2048,
				NumGC:         5,
				GCCPUFraction: 0.1,
			},
			expected: map[string]float64{
				MetricAlloc:         1024,
				MetricHeapAlloc:     2048,
				MetricNumGC:         5,
				MetricGCCPUFraction: 0.1,
			},
		},
		{
			name:     "Zero values",
			memStats: runtime.MemStats{},
			expected: map[string]float64{
				MetricAlloc:         0,
				MetricHeapAlloc:     0,
				MetricNumGC:         0,
				MetricGCCPUFraction: 0,
			},
		},
		{
			name: "Large values",
			memStats: runtime.MemStats{
				Alloc:         999999999,
				HeapAlloc:     888888888,
				NumGC:         1000,
				GCCPUFraction: 0.99,
			},
			expected: map[string]float64{
				MetricAlloc:         999999999,
				MetricHeapAlloc:     888888888,
				MetricNumGC:         1000,
				MetricGCCPUFraction: 0.99,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()

			// Заполняем метрики
			FillRuntimeMetrics(metrics, tt.memStats)

			// Проверяем количество метрик
			expectedCount := 27 // 27 runtime метрик из FillRuntimeMetrics
			assert.Len(t, metrics.Gauges, expectedCount, "Should have correct number of runtime metrics")

			// Проверяем конкретные значения
			for metricName, expectedValue := range tt.expected {
				assert.Equal(t, expectedValue, metrics.Gauges[metricName],
					"Metric %s should match expected value", metricName)
			}

			// Проверяем, что все обязательные метрики присутствуют
			requiredMetrics := []string{
				MetricAlloc, MetricBuckHashSys, MetricFrees, MetricGCCPUFraction, MetricGCSys,
				MetricHeapAlloc, MetricHeapIdle, MetricHeapInuse, MetricHeapObjects, MetricHeapReleased,
				MetricHeapSys, MetricLastGC, MetricLookups, MetricMCacheInuse, MetricMCacheSys,
				MetricMSpanInuse, MetricMSpanSys, MetricMallocs, MetricNextGC, MetricNumForcedGC,
				MetricNumGC, MetricOtherSys, MetricPauseTotalNs, MetricStackInuse, MetricStackSys,
				MetricSys, MetricTotalAlloc,
			}

			for _, metricName := range requiredMetrics {
				_, exists := metrics.Gauges[metricName]
				assert.True(t, exists, "Required metric %s should exist", metricName)
			}
		})
	}
}

func TestFillAdditionalMetrics(t *testing.T) {
	tests := []struct {
		name          string
		iterations    int
		expectedCount int
	}{
		{
			name:          "Single iteration",
			iterations:    1,
			expectedCount: 1,
		},
		{
			name:          "Multiple iterations",
			iterations:    5,
			expectedCount: 1, // RandomValue перезаписывается каждый раз
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()

			// Заполняем дополнительные метрики несколько раз
			for i := 0; i < tt.iterations; i++ {
				FillAdditionalMetrics(metrics)
			}

			// Проверяем, что метрика добавлена
			_, exists := metrics.Gauges[MetricRandomValue]
			assert.True(t, exists, "RandomValue metric should exist")

			// Проверяем тип значения
			value := metrics.Gauges[MetricRandomValue]
			assert.GreaterOrEqual(t, value, 0.0, "RandomValue should be >= 0")
			assert.Less(t, value, 1.0, "RandomValue should be < 1")

			// Проверяем количество метрик
			assert.Len(t, metrics.Gauges, tt.expectedCount, "Should have correct number of additional metrics")
		})
	}
}

func TestUpdateCounterMetrics(t *testing.T) {
	tests := []struct {
		name           string
		initialValue   int64
		iterations     int
		expectedValues []int64
	}{
		{
			name:           "Start from zero",
			initialValue:   0,
			iterations:     3,
			expectedValues: []int64{1, 2, 3},
		},
		{
			name:           "Start from existing value",
			initialValue:   100,
			iterations:     2,
			expectedValues: []int64{101, 102},
		},
		{
			name:           "Single iteration",
			initialValue:   0,
			iterations:     1,
			expectedValues: []int64{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewMetrics()

			// Устанавливаем начальное значение
			if tt.initialValue > 0 {
				metrics.Counters[MetricPollCount] = tt.initialValue
			}

			// Выполняем обновления
			for i := 0; i < tt.iterations; i++ {
				UpdateCounterMetrics(metrics)
				expectedValue := tt.expectedValues[i]
				assert.Equal(t, expectedValue, metrics.Counters[MetricPollCount],
					"PollCount should be %d after %d iterations", expectedValue, i+1)
			}

			// Проверяем количество метрик
			assert.Len(t, metrics.Counters, 1, "Should have exactly 1 counter metric")
		})
	}
}
