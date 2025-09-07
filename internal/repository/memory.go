package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/IgorKilipenko/metrical/internal/logger"
	models "github.com/IgorKilipenko/metrical/internal/model"
)

// InMemoryMetricsRepository реализация репозитория в памяти
type InMemoryMetricsRepository struct {
	Gauges          models.GaugeMetrics
	Counters        models.CounterMetrics
	mu              sync.RWMutex // Мьютекс для потокобезопасности
	logger          logger.Logger
	fileStoragePath string // Путь к файлу для сохранения/загрузки метрик
	restore         bool   // Флаг для восстановления метрик из файла
	syncSave        bool   // Флаг для синхронного сохранения при каждом обновлении
}

// NewInMemoryMetricsRepository создает новый экземпляр InMemoryMetricsRepository
func NewInMemoryMetricsRepository(logger logger.Logger, fileStoragePath string, restore bool) *InMemoryMetricsRepository {
	repo := &InMemoryMetricsRepository{
		Gauges:          make(models.GaugeMetrics),
		Counters:        make(models.CounterMetrics),
		logger:          logger,
		fileStoragePath: fileStoragePath,
		restore:         restore,
		syncSave:        false, // По умолчанию синхронное сохранение отключено
	}
	if restore {
		if err := repo.LoadFromFile(); err != nil {
			logger.Warn("failed to load metrics from file", "error", err)
		} else {
			logger.Info("metrics loaded from file successfully")
		}
	}
	return repo
}

// SetSyncSave устанавливает флаг синхронного сохранения
func (r *InMemoryMetricsRepository) SetSyncSave(sync bool) {
	r.syncSave = sync
}

// UpdateGauge обновляет значение gauge метрики
func (r *InMemoryMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during gauge update", "name", name, "value", value)
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldValue, exists := r.Gauges[name]
	r.Gauges[name] = value

	if exists {
		r.logger.Debug("updated existing gauge metric", "name", name, "old_value", oldValue, "new_value", value)
	} else {
		r.logger.Debug("created new gauge metric", "name", name, "value", value)
	}

	// Синхронное сохранение, если включено
	if r.syncSave {
		if err := r.saveToFileUnsafe(); err != nil {
			r.logger.Error("failed to save metrics synchronously", "error", err)
			return fmt.Errorf("failed to save metrics synchronously: %w", err)
		}
		r.logger.Debug("metrics saved synchronously after gauge update")
	}

	return nil
}

// UpdateCounter добавляет значение к counter метрике
func (r *InMemoryMetricsRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during counter update", "name", name, "value", value)
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	oldValue := r.Counters[name]
	r.Counters[name] += value

	r.logger.Debug("updated counter metric", "name", name, "added_value", value, "old_total", oldValue, "new_total", r.Counters[name])

	// Синхронное сохранение, если включено
	if r.syncSave {
		if err := r.saveToFileUnsafe(); err != nil {
			r.logger.Error("failed to save metrics synchronously", "error", err)
			return fmt.Errorf("failed to save metrics synchronously: %w", err)
		}
		r.logger.Debug("metrics saved synchronously after counter update")
	}

	return nil
}

// GetGauge возвращает значение gauge метрики
func (r *InMemoryMetricsRepository) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during gauge retrieval", "name", name)
		return 0, false, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	value, exists := r.Gauges[name]

	if exists {
		r.logger.Debug("retrieved gauge metric", "name", name, "value", value)
	} else {
		r.logger.Debug("gauge metric not found", "name", name)
	}

	return value, exists, nil
}

// GetCounter возвращает значение counter метрики
func (r *InMemoryMetricsRepository) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during counter retrieval", "name", name)
		return 0, false, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	value, exists := r.Counters[name]

	if exists {
		r.logger.Debug("retrieved counter metric", "name", name, "value", value)
	} else {
		r.logger.Debug("counter metric not found", "name", name)
	}

	return value, exists, nil
}

// GetAllGauges возвращает все gauge метрики
func (r *InMemoryMetricsRepository) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during getAllGauges")
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Создаем копию для безопасного возврата
	result := make(models.GaugeMetrics, len(r.Gauges))
	for k, v := range r.Gauges {
		result[k] = v
	}

	r.logger.Debug("retrieved all gauge metrics", "count", len(result))
	return result, nil
}

// GetAllCounters возвращает все counter метрики
func (r *InMemoryMetricsRepository) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	// Проверяем отмену контекста
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during getAllCounters")
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Создаем копию для безопасного возврата
	result := make(models.CounterMetrics, len(r.Counters))
	for k, v := range r.Counters {
		result[k] = v
	}

	r.logger.Debug("retrieved all counter metrics", "count", len(result))
	return result, nil
}

// SaveToFile сохраняет все метрики в файл
func (r *InMemoryMetricsRepository) SaveToFile() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.saveToFileUnsafe()
}

// saveToFileUnsafe сохраняет метрики в файл без блокировки (для внутреннего использования)
func (r *InMemoryMetricsRepository) saveToFileUnsafe() error {
	// Создаем слайс метрик для сохранения
	var metrics []models.Metrics

	// Добавляем gauge метрики
	for name, value := range r.Gauges {
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		})
	}

	// Добавляем counter метрики
	for name, delta := range r.Counters {
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &delta,
		})
	}

	// Кодируем в JSON
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics to JSON: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(r.fileStoragePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics to file: %w", err)
	}

	r.logger.Debug("metrics saved to file", "path", r.fileStoragePath, "count", len(metrics))
	return nil
}

// LoadFromFile загружает метрики из файла
func (r *InMemoryMetricsRepository) LoadFromFile() error {
	// Проверяем существование файла
	if _, err := os.Stat(r.fileStoragePath); os.IsNotExist(err) {
		r.logger.Debug("metrics file does not exist, skipping load", "path", r.fileStoragePath)
		return nil
	}

	// Читаем файл
	data, err := os.ReadFile(r.fileStoragePath)
	if err != nil {
		return fmt.Errorf("failed to read metrics file: %w", err)
	}

	// Декодируем JSON
	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("failed to unmarshal metrics from JSON: %w", err)
	}

	// Очищаем текущие метрики
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Gauges = make(models.GaugeMetrics)
	r.Counters = make(models.CounterMetrics)

	// Загружаем метрики
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				r.Gauges[metric.ID] = *metric.Value
			}
		case models.Counter:
			if metric.Delta != nil {
				r.Counters[metric.ID] = *metric.Delta
			}
		}
	}

	r.logger.Debug("metrics loaded from file", "path", r.fileStoragePath, "count", len(metrics))
	return nil
}
