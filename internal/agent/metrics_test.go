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
	metrics := NewMetrics()

	// Добавляем тестовые метрики
	metrics.Gauges["test_gauge"] = 123.45
	metrics.Counters["test_counter"] = 678

	// Получаем все метрики
	allMetrics := metrics.GetAllMetrics()

	// Проверяем, что все метрики присутствуют
	assert.Equal(t, 123.45, allMetrics["test_gauge"])
	assert.Equal(t, int64(678), allMetrics["test_counter"])
	assert.Len(t, allMetrics, 2, "Should have 2 metrics total")
}

func TestFillRuntimeMetrics(t *testing.T) {
	// Создаем тестовую структуру метрик
	metrics := NewMetrics()

	// Создаем тестовые MemStats
	var memStats runtime.MemStats
	memStats.Alloc = 1024
	memStats.HeapAlloc = 2048
	memStats.NumGC = 5
	memStats.GCCPUFraction = 0.1

	// Заполняем метрики
	FillRuntimeMetrics(metrics, memStats)

	// Проверяем, что метрики заполнены
	assert.Equal(t, float64(1024), metrics.Gauges[MetricAlloc])
	assert.Equal(t, float64(2048), metrics.Gauges[MetricHeapAlloc])
	assert.Equal(t, float64(5), metrics.Gauges[MetricNumGC])
	assert.Equal(t, 0.1, metrics.Gauges[MetricGCCPUFraction])

	// Проверяем количество метрик
	expectedCount := 27 // 27 runtime метрик из FillRuntimeMetrics
	assert.Len(t, metrics.Gauges, expectedCount)

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
}

func TestFillAdditionalMetrics(t *testing.T) {
	// Создаем тестовую структуру метрик
	metrics := NewMetrics()

	// Заполняем дополнительные метрики
	FillAdditionalMetrics(metrics)

	// Проверяем, что метрика добавлена
	_, exists := metrics.Gauges[MetricRandomValue]
	assert.True(t, exists, "RandomValue metric should exist")

	// Проверяем тип значения
	value := metrics.Gauges[MetricRandomValue]
	assert.GreaterOrEqual(t, value, 0.0, "RandomValue should be >= 0")
	assert.Less(t, value, 1.0, "RandomValue should be < 1")

	// Проверяем количество метрик
	assert.Len(t, metrics.Gauges, 1, "Should have exactly 1 additional metric")
}

func TestUpdateCounterMetrics(t *testing.T) {
	// Создаем тестовую структуру метрик
	metrics := NewMetrics()

	// Первое обновление - должно установить значение 1
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(1), metrics.Counters[MetricPollCount])

	// Второе обновление - должно увеличить до 2
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(2), metrics.Counters[MetricPollCount])

	// Третье обновление - должно увеличить до 3
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(3), metrics.Counters[MetricPollCount])

	// Проверяем количество метрик
	assert.Len(t, metrics.Counters, 1, "Should have exactly 1 counter metric")
}

func TestUpdateCounterMetrics_WithExistingValue(t *testing.T) {
	// Создаем тестовую структуру метрик с существующим значением
	metrics := NewMetrics()
	metrics.Counters[MetricPollCount] = 100

	// Обновляем counter метрики
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(101), metrics.Counters[MetricPollCount])

	// Еще раз обновляем
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(102), metrics.Counters[MetricPollCount])
}
