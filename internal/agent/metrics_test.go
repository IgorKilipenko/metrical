package agent

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillRuntimeMetrics(t *testing.T) {
	// Создаем тестовую мапу
	metrics := make(map[string]any)

	// Создаем тестовые MemStats
	var memStats runtime.MemStats
	memStats.Alloc = 1024
	memStats.HeapAlloc = 2048
	memStats.NumGC = 5
	memStats.GCCPUFraction = 0.1

	// Заполняем метрики
	FillRuntimeMetrics(metrics, memStats)

	// Проверяем, что метрики заполнены
	assert.Equal(t, float64(1024), metrics[MetricAlloc])
	assert.Equal(t, float64(2048), metrics[MetricHeapAlloc])
	assert.Equal(t, float64(5), metrics[MetricNumGC])
	assert.Equal(t, 0.1, metrics[MetricGCCPUFraction])

	// Проверяем количество метрик
	expectedCount := 27 // 27 runtime метрик из FillRuntimeMetrics
	assert.Len(t, metrics, expectedCount)

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
		_, exists := metrics[metricName]
		assert.True(t, exists, "Required metric %s should exist", metricName)
	}
}

func TestFillAdditionalMetrics(t *testing.T) {
	// Создаем тестовую мапу
	metrics := make(map[string]any)

	// Заполняем дополнительные метрики
	FillAdditionalMetrics(metrics)

	// Проверяем, что метрика добавлена
	_, exists := metrics[MetricRandomValue]
	assert.True(t, exists, "RandomValue metric should exist")

	// Проверяем тип значения
	value, ok := metrics[MetricRandomValue].(float64)
	assert.True(t, ok, "RandomValue should be float64")
	assert.GreaterOrEqual(t, value, 0.0, "RandomValue should be >= 0")
	assert.Less(t, value, 1.0, "RandomValue should be < 1")

	// Проверяем количество метрик
	assert.Len(t, metrics, 1, "Should have exactly 1 additional metric")
}

func TestUpdateCounterMetrics(t *testing.T) {
	// Создаем тестовую мапу
	metrics := make(map[string]any)

	// Первое обновление - должно установить значение 1
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(1), metrics[MetricPollCount])

	// Второе обновление - должно увеличить до 2
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(2), metrics[MetricPollCount])

	// Третье обновление - должно увеличить до 3
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(3), metrics[MetricPollCount])

	// Проверяем количество метрик
	assert.Len(t, metrics, 1, "Should have exactly 1 counter metric")
}

func TestUpdateCounterMetrics_WithExistingValue(t *testing.T) {
	// Создаем тестовую мапу с существующим значением
	metrics := make(map[string]any)
	metrics[MetricPollCount] = int64(100)

	// Обновляем counter метрики
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(101), metrics[MetricPollCount])

	// Еще раз обновляем
	UpdateCounterMetrics(metrics)
	assert.Equal(t, int64(102), metrics[MetricPollCount])
}
