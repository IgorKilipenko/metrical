package agent

import (
	"math/rand"
	"runtime"
)

// Константы для имен метрик runtime
const (
	// Gauge метрики из runtime.MemStats
	MetricAlloc         = "Alloc"
	MetricBuckHashSys   = "BuckHashSys"
	MetricFrees         = "Frees"
	MetricGCCPUFraction = "GCCPUFraction"
	MetricGCSys         = "GCSys"
	MetricHeapAlloc     = "HeapAlloc"
	MetricHeapIdle      = "HeapIdle"
	MetricHeapInuse     = "HeapInuse"
	MetricHeapObjects   = "HeapObjects"
	MetricHeapReleased  = "HeapReleased"
	MetricHeapSys       = "HeapSys"
	MetricLastGC        = "LastGC"
	MetricLookups       = "Lookups"
	MetricMCacheInuse   = "MCacheInuse"
	MetricMCacheSys     = "MCacheSys"
	MetricMSpanInuse    = "MSpanInuse"
	MetricMSpanSys      = "MSpanSys"
	MetricMallocs       = "Mallocs"
	MetricNextGC        = "NextGC"
	MetricNumForcedGC   = "NumForcedGC"
	MetricNumGC         = "NumGC"
	MetricOtherSys      = "OtherSys"
	MetricPauseTotalNs  = "PauseTotalNs"
	MetricStackInuse    = "StackInuse"
	MetricStackSys      = "StackSys"
	MetricSys           = "Sys"
	MetricTotalAlloc    = "TotalAlloc"

	// Дополнительные метрики
	MetricRandomValue = "RandomValue"
	MetricPollCount   = "PollCount"
)

// Типы метрик
const (
	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"
)

// FillRuntimeMetrics заполняет мапу метриками из runtime.MemStats
func FillRuntimeMetrics(metrics map[string]any, memStats runtime.MemStats) {
	// Gauge метрики из runtime
	metrics[MetricAlloc] = float64(memStats.Alloc)
	metrics[MetricBuckHashSys] = float64(memStats.BuckHashSys)
	metrics[MetricFrees] = float64(memStats.Frees)
	metrics[MetricGCCPUFraction] = memStats.GCCPUFraction
	metrics[MetricGCSys] = float64(memStats.GCSys)
	metrics[MetricHeapAlloc] = float64(memStats.HeapAlloc)
	metrics[MetricHeapIdle] = float64(memStats.HeapIdle)
	metrics[MetricHeapInuse] = float64(memStats.HeapInuse)
	metrics[MetricHeapObjects] = float64(memStats.HeapObjects)
	metrics[MetricHeapReleased] = float64(memStats.HeapReleased)
	metrics[MetricHeapSys] = float64(memStats.HeapSys)
	metrics[MetricLastGC] = float64(memStats.LastGC)
	metrics[MetricLookups] = float64(memStats.Lookups)
	metrics[MetricMCacheInuse] = float64(memStats.MCacheInuse)
	metrics[MetricMCacheSys] = float64(memStats.MCacheSys)
	metrics[MetricMSpanInuse] = float64(memStats.MSpanInuse)
	metrics[MetricMSpanSys] = float64(memStats.MSpanSys)
	metrics[MetricMallocs] = float64(memStats.Mallocs)
	metrics[MetricNextGC] = float64(memStats.NextGC)
	metrics[MetricNumForcedGC] = float64(memStats.NumForcedGC)
	metrics[MetricNumGC] = float64(memStats.NumGC)
	metrics[MetricOtherSys] = float64(memStats.OtherSys)
	metrics[MetricPauseTotalNs] = float64(memStats.PauseTotalNs)
	metrics[MetricStackInuse] = float64(memStats.StackInuse)
	metrics[MetricStackSys] = float64(memStats.StackSys)
	metrics[MetricSys] = float64(memStats.Sys)
	metrics[MetricTotalAlloc] = float64(memStats.TotalAlloc)
}

// FillAdditionalMetrics заполняет мапу дополнительными метриками
func FillAdditionalMetrics(metrics map[string]any) {
	metrics[MetricRandomValue] = rand.Float64()
}

// UpdateCounterMetrics обновляет counter метрики (накапливает значения)
func UpdateCounterMetrics(metrics map[string]any) {
	if pollCount, exists := metrics[MetricPollCount]; exists {
		metrics[MetricPollCount] = pollCount.(int64) + 1
	} else {
		metrics[MetricPollCount] = int64(1)
	}
}
