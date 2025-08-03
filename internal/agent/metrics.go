package agent

import (
	"math/rand"
	"runtime"
)

// Metrics структура для хранения метрик
type Metrics struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

// NewMetrics создает новый экземпляр Metrics
func NewMetrics() *Metrics {
	return &Metrics{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

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

// FillRuntimeMetrics заполняет структуру метриками из runtime.MemStats
func FillRuntimeMetrics(metrics *Metrics, memStats runtime.MemStats) {
	// Gauge метрики из runtime
	metrics.Gauges[MetricAlloc] = float64(memStats.Alloc)
	metrics.Gauges[MetricBuckHashSys] = float64(memStats.BuckHashSys)
	metrics.Gauges[MetricFrees] = float64(memStats.Frees)
	metrics.Gauges[MetricGCCPUFraction] = memStats.GCCPUFraction
	metrics.Gauges[MetricGCSys] = float64(memStats.GCSys)
	metrics.Gauges[MetricHeapAlloc] = float64(memStats.HeapAlloc)
	metrics.Gauges[MetricHeapIdle] = float64(memStats.HeapIdle)
	metrics.Gauges[MetricHeapInuse] = float64(memStats.HeapInuse)
	metrics.Gauges[MetricHeapObjects] = float64(memStats.HeapObjects)
	metrics.Gauges[MetricHeapReleased] = float64(memStats.HeapReleased)
	metrics.Gauges[MetricHeapSys] = float64(memStats.HeapSys)
	metrics.Gauges[MetricLastGC] = float64(memStats.LastGC)
	metrics.Gauges[MetricLookups] = float64(memStats.Lookups)
	metrics.Gauges[MetricMCacheInuse] = float64(memStats.MCacheInuse)
	metrics.Gauges[MetricMCacheSys] = float64(memStats.MCacheSys)
	metrics.Gauges[MetricMSpanInuse] = float64(memStats.MSpanInuse)
	metrics.Gauges[MetricMSpanSys] = float64(memStats.MSpanSys)
	metrics.Gauges[MetricMallocs] = float64(memStats.Mallocs)
	metrics.Gauges[MetricNextGC] = float64(memStats.NextGC)
	metrics.Gauges[MetricNumForcedGC] = float64(memStats.NumForcedGC)
	metrics.Gauges[MetricNumGC] = float64(memStats.NumGC)
	metrics.Gauges[MetricOtherSys] = float64(memStats.OtherSys)
	metrics.Gauges[MetricPauseTotalNs] = float64(memStats.PauseTotalNs)
	metrics.Gauges[MetricStackInuse] = float64(memStats.StackInuse)
	metrics.Gauges[MetricStackSys] = float64(memStats.StackSys)
	metrics.Gauges[MetricSys] = float64(memStats.Sys)
	metrics.Gauges[MetricTotalAlloc] = float64(memStats.TotalAlloc)
}

// FillAdditionalMetrics заполняет структуру дополнительными метриками
func FillAdditionalMetrics(metrics *Metrics) {
	metrics.Gauges[MetricRandomValue] = rand.Float64()
}

// UpdateCounterMetrics обновляет counter метрики (накапливает значения)
func UpdateCounterMetrics(metrics *Metrics) {
	metrics.Counters[MetricPollCount]++
}

// GetAllMetrics возвращает все метрики в виде map[string]any для совместимости
func (m *Metrics) GetAllMetrics() map[string]any {
	result := make(map[string]any)

	// Добавляем gauge метрики
	for name, value := range m.Gauges {
		result[name] = value
	}

	// Добавляем counter метрики
	for name, value := range m.Counters {
		result[name] = value
	}

	return result
}
