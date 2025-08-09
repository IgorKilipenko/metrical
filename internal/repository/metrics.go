package repository

import (
	models "github.com/IgorKilipenko/metrical/internal/model"
)

// MetricsRepository интерфейс для работы с метриками
type MetricsRepository interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGauge(name string) (float64, bool, error)
	GetCounter(name string) (int64, bool, error)
	GetAllGauges() (models.GaugeMetrics, error)
	GetAllCounters() (models.CounterMetrics, error)
}
