package agent

import (
	"math/rand"
	"runtime"

	models "github.com/IgorKilipenko/metrical/internal/model"
)

// Metrics структура для хранения метрик.
// Содержит отдельные map'ы для gauge и counter метрик.
type Metrics struct {
	// Gauges содержит gauge метрики (заменяют предыдущие значения)
	Gauges models.GaugeMetrics

	// Counters содержит counter метрики (накапливают значения)
	Counters models.CounterMetrics
}

// NewMetrics создает новый экземпляр Metrics.
// Инициализирует пустые map'ы для gauge и counter метрик.
//
// Возвращает:
//   - *Metrics: указатель на новую структуру Metrics
func NewMetrics() *Metrics {
	return &Metrics{
		Gauges:   make(models.GaugeMetrics),
		Counters: make(models.CounterMetrics),
	}
}

// Константы для имен метрик runtime
const (
	// Gauge метрики из runtime.MemStats

	// MetricAlloc - текущее количество байт, выделенных и еще не освобожденных
	MetricAlloc = "Alloc"

	// MetricBuckHashSys - количество байт, используемых хеш-таблицами профилирования
	MetricBuckHashSys = "BuckHashSys"

	// MetricFrees - общее количество освобождений памяти
	MetricFrees = "Frees"

	// MetricGCCPUFraction - доля времени CPU, затраченного на сборку мусора
	MetricGCCPUFraction = "GCCPUFraction"

	// MetricGCSys - количество байт, используемых системой сборки мусора
	MetricGCSys = "GCSys"

	// MetricHeapAlloc - количество байт в использовании heap
	MetricHeapAlloc = "HeapAlloc"

	// MetricHeapIdle - количество байт в неиспользуемых span'ах
	MetricHeapIdle = "HeapIdle"

	// MetricHeapInuse - количество байт в используемых span'ах
	MetricHeapInuse = "HeapInuse"

	// MetricHeapObjects - количество выделенных объектов
	MetricHeapObjects = "HeapObjects"

	// MetricHeapReleased - количество байт, возвращенных операционной системе
	MetricHeapReleased = "HeapReleased"

	// MetricHeapSys - общее количество байт, полученных от операционной системы
	MetricHeapSys = "HeapSys"

	// MetricLastGC - время последней сборки мусора в наносекундах
	MetricLastGC = "LastGC"

	// MetricLookups - количество указателей, просмотренных runtime
	MetricLookups = "Lookups"

	// MetricMCacheInuse - количество байт в используемых mcache структурах
	MetricMCacheInuse = "MCacheInuse"

	// MetricMCacheSys - количество байт, используемых mcache структурами
	MetricMCacheSys = "MCacheSys"

	// MetricMSpanInuse - количество байт в используемых mspan структурах
	MetricMSpanInuse = "MSpanInuse"

	// MetricMSpanSys - количество байт, используемых mspan структурами
	MetricMSpanSys = "MSpanSys"

	// MetricMallocs - общее количество аллокаций памяти
	MetricMallocs = "Mallocs"

	// MetricNextGC - целевое значение heap size для следующей сборки мусора
	MetricNextGC = "NextGC"

	// MetricNumForcedGC - количество принудительных сборок мусора
	MetricNumForcedGC = "NumForcedGC"

	// MetricNumGC - количество завершенных сборок мусора
	MetricNumGC = "NumGC"

	// MetricOtherSys - количество байт, используемых другими системными аллокациями
	MetricOtherSys = "OtherSys"

	// MetricPauseTotalNs - общее время пауз сборки мусора в наносекундах
	MetricPauseTotalNs = "PauseTotalNs"

	// MetricStackInuse - количество байт в использовании stack
	MetricStackInuse = "StackInuse"

	// MetricStackSys - количество байт, полученных от ОС для stack
	MetricStackSys = "StackSys"

	// MetricSys - общее количество байт, полученных от операционной системы
	MetricSys = "Sys"

	// MetricTotalAlloc - общее количество аллокаций памяти
	MetricTotalAlloc = "TotalAlloc"

	// Дополнительные метрики

	// MetricRandomValue - случайное значение от 0 до 1 для тестирования
	MetricRandomValue = "RandomValue"

	// MetricPollCount - счетчик обновлений метрик (накапливается)
	MetricPollCount = "PollCount"
)

// Типы метрик
const (
	// MetricTypeGauge - тип метрики gauge (заменяет предыдущее значение)
	MetricTypeGauge = "gauge"

	// MetricTypeCounter - тип метрики counter (накапливает значения)
	MetricTypeCounter = "counter"
)

// FillRuntimeMetrics заполняет структуру метриками из runtime.MemStats.
// Принимает указатель на Metrics и структуру runtime.MemStats.
// Заполняет все 27 gauge метрик из runtime пакета.
//
// Параметры:
//   - metrics: указатель на структуру Metrics для заполнения
//   - memStats: структура runtime.MemStats с данными о памяти
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

// FillAdditionalMetrics заполняет структуру дополнительными метриками.
// Добавляет RandomValue - случайное значение от 0 до 1.
// Используется для тестирования и демонстрации.
//
// Параметры:
//   - metrics: указатель на структуру Metrics для заполнения
func FillAdditionalMetrics(metrics *Metrics) {
	metrics.Gauges[MetricRandomValue] = rand.Float64()
}

// UpdateCounterMetrics обновляет counter метрики (накапливает значения).
// Увеличивает PollCount на 1 при каждом вызове.
// PollCount используется для отслеживания количества обновлений метрик.
//
// Параметры:
//   - metrics: указатель на структуру Metrics для обновления
func UpdateCounterMetrics(metrics *Metrics) {
	metrics.Counters[MetricPollCount]++
}

// GetAllMetrics возвращает все метрики в виде map[string]any для совместимости.
// Объединяет gauge и counter метрики в один map.
// Используется для отправки метрик на сервер.
//
// Возвращает:
//   - map[string]any: объединенный map всех метрик
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
