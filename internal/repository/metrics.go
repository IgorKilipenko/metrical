package repository

import (
	"context"

	models "github.com/IgorKilipenko/metrical/internal/model"
)

// MetricsRepository интерфейс для работы с метриками
type MetricsRepository interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	GetGauge(ctx context.Context, name string) (float64, bool, error)
	GetCounter(ctx context.Context, name string) (int64, bool, error)
	GetAllGauges(ctx context.Context) (models.GaugeMetrics, error)
	GetAllCounters(ctx context.Context) (models.CounterMetrics, error)
	SaveToFile() error
	LoadFromFile() error
	SetSyncSave(sync bool)
}
