package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/IgorKilipenko/metrical/internal/logger"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/repository"
	"github.com/IgorKilipenko/metrical/internal/validation"
)

// Константы для сообщений об ошибках
const (
	ErrMsgUnsupportedMetricType = "unsupported metric type"
	ErrMsgMetricNameEmpty       = "metric name cannot be empty"
	ErrMsgMetricIDEmpty         = "metric ID cannot be empty"
)

// checkContextCancellation проверяет отмену контекста
func (s *MetricsService) checkContextCancellation(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

// validateMetricName проверяет, что имя метрики не пустое
func (s *MetricsService) validateMetricName(name string) error {
	if name == "" {
		return errors.New(ErrMsgMetricNameEmpty)
	}
	return nil
}

// validateMetricID проверяет, что ID метрики не пустое
func (s *MetricsService) validateMetricID(id string) error {
	if id == "" {
		return errors.New(ErrMsgMetricIDEmpty)
	}
	return nil
}

// MetricsService сервис для работы с метриками
type MetricsService struct {
	repository repository.MetricsRepository
	logger     logger.Logger
}

// NewMetricsService создает новый экземпляр MetricsService.
//
// Функция принимает репозиторий для работы с данными и логгер для записи событий.
// Оба параметра обязательны и не могут быть nil.
//
// Параметры:
//   - repository: репозиторий для работы с метриками (не может быть nil)
//   - logger: логгер для записи событий (не может быть nil)
//
// Возвращает:
//   - *MetricsService: готовый к использованию сервис
//
// Паникует если:
//   - repository == nil
//   - logger == nil
//
// Пример использования:
//
//	repo := repository.NewInMemoryMetricsRepository(logger)
//	service := NewMetricsService(repo, logger)
func NewMetricsService(repository repository.MetricsRepository, logger logger.Logger) *MetricsService {
	if repository == nil {
		panic("repository cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	return &MetricsService{
		repository: repository,
		logger:     logger,
	}
}

// UpdateMetric обновляет метрику с готовыми валидированными данными.
//
// Функция принимает валидированный MetricRequest и делегирует обновление
// соответствующему методу в зависимости от типа метрики.
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - req: валидированный запрос на обновление метрики
//
// Возвращает:
//   - error: ошибка операции или nil при успехе
//
// Пример использования:
//
//	err := service.UpdateMetric(ctx, &validation.MetricRequest{
//	    Type:  "gauge",
//	    Name:  "temperature",
//	    Value: 23.5,
//	})
func (s *MetricsService) UpdateMetric(ctx context.Context, req *validation.MetricRequest) error {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return err
	}

	// Валидируем входные параметры
	if req == nil {
		return fmt.Errorf("metric request cannot be nil")
	}
	if err := s.validateMetricName(req.Name); err != nil {
		return err
	}

	s.logger.Info("updating metric", "name", req.Name, "type", req.Type, "value", req.Value)

	// Безопасное приведение типов с проверкой
	switch req.Type {
	case models.Gauge:
		value, ok := req.Value.(float64)
		if !ok {
			return fmt.Errorf("invalid gauge value type: expected float64, got %T", req.Value)
		}
		return s.updateGaugeMetric(ctx, req.Name, value)
	case models.Counter:
		value, ok := req.Value.(int64)
		if !ok {
			return fmt.Errorf("invalid counter value type: expected int64, got %T", req.Value)
		}
		return s.updateCounterMetric(ctx, req.Name, value)
	default:
		return fmt.Errorf(ErrMsgUnsupportedMetricType+": %s", req.Type)
	}
}

// UpdateMetricJSON обновляет метрику из JSON структуры.
//
// Функция принимает структуру Metrics и обновляет соответствующую метрику
// в зависимости от типа (gauge или counter).
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - metric: структура метрики с данными для обновления
//
// Возвращает:
//   - error: ошибка операции или nil при успехе
//
// Пример использования:
//
//	err := service.UpdateMetricJSON(ctx, &models.Metrics{
//	    ID:    "temperature",
//	    MType: "gauge",
//	    Value: &value,
//	})
func (s *MetricsService) UpdateMetricJSON(ctx context.Context, metric *models.Metrics) error {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return err
	}

	// Валидируем входные параметры
	if metric == nil {
		return fmt.Errorf("metric cannot be nil")
	}
	if err := s.validateMetricID(metric.ID); err != nil {
		return err
	}

	s.logger.Info("updating metric from JSON", "id", metric.ID, "type", metric.MType)

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("value is required for gauge metric")
		}
		return s.updateGaugeMetric(ctx, metric.ID, *metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("delta is required for counter metric")
		}
		return s.updateCounterMetric(ctx, metric.ID, *metric.Delta)
	default:
		return fmt.Errorf(ErrMsgUnsupportedMetricType+": %s", metric.MType)
	}
}

// updateGaugeMetric содержит бизнес-логику для обновления gauge метрик
func (s *MetricsService) updateGaugeMetric(ctx context.Context, name string, value float64) error {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return err
	}

	// Валидируем входные параметры
	if err := s.validateMetricName(name); err != nil {
		return err
	}

	s.logger.Debug("updating gauge metric", "name", name, "value", value)

	// Здесь может быть бизнес-логика:
	// - Проверка лимитов
	// - Валидация бизнес-правил
	// - Агрегация данных
	// - Уведомления
	// - Аудит операций

	// Пока просто делегируем в репозиторий
	err := s.repository.UpdateGauge(ctx, name, value)
	if err != nil {
		return err
	}

	s.logger.Debug("gauge metric updated successfully", "name", name, "value", value)
	return nil
}

// updateCounterMetric содержит бизнес-логику для обновления counter метрик
func (s *MetricsService) updateCounterMetric(ctx context.Context, name string, value int64) error {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return err
	}

	// Валидируем входные параметры
	if err := s.validateMetricName(name); err != nil {
		return err
	}

	s.logger.Debug("updating counter metric", "name", name, "value", value)

	// Здесь может быть бизнес-логика:
	// - Проверка лимитов счетчиков
	// - Валидация бизнес-правил
	// - Агрегация данных
	// - Уведомления при превышении порогов

	// Пока просто делегируем в репозиторий
	err := s.repository.UpdateCounter(ctx, name, value)
	if err != nil {
		return err
	}

	s.logger.Debug("counter metric updated successfully", "name", name, "value", value)
	return nil
}

// GetGauge возвращает значение gauge метрики.
//
// Функция выполняет поиск gauge метрики по имени и возвращает её значение.
// Если метрика не найдена, возвращается false в качестве второго значения.
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - name: имя метрики для поиска
//
// Возвращает:
//   - float64: значение метрики (0 если не найдена)
//   - bool: true если метрика найдена, false если нет
//   - error: ошибка операции или nil при успехе
func (s *MetricsService) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return 0, false, err
	}

	// Валидируем входные параметры
	if err := s.validateMetricName(name); err != nil {
		return 0, false, err
	}

	s.logger.Debug("getting gauge metric", "name", name)

	value, exists, err := s.repository.GetGauge(ctx, name)
	if err != nil {
		return 0, false, err
	}

	if exists {
		s.logger.Debug("gauge metric retrieved", "name", name, "value", value)
	} else {
		s.logger.Debug("gauge metric not found", "name", name)
	}

	return value, exists, nil
}

// GetCounter возвращает значение counter метрики.
//
// Функция выполняет поиск counter метрики по имени и возвращает её накопительное значение.
// Если метрика не найдена, возвращается false в качестве второго значения.
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - name: имя метрики для поиска
//
// Возвращает:
//   - int64: накопительное значение метрики (0 если не найдена)
//   - bool: true если метрика найдена, false если нет
//   - error: ошибка операции или nil при успехе
func (s *MetricsService) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return 0, false, err
	}

	// Валидируем входные параметры
	if err := s.validateMetricName(name); err != nil {
		return 0, false, err
	}

	s.logger.Debug("getting counter metric", "name", name)

	value, exists, err := s.repository.GetCounter(ctx, name)
	if err != nil {
		return 0, false, err
	}

	if exists {
		s.logger.Debug("counter metric retrieved", "name", name, "value", value)
	} else {
		s.logger.Debug("counter metric not found", "name", name)
	}

	return value, exists, nil
}

// GetAllGauges возвращает все gauge метрики.
//
// Функция выполняет запрос всех gauge метрик и возвращает их в виде map,
// где ключ - имя метрики, значение - её значение.
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//
// Возвращает:
//   - models.GaugeMetrics: map всех gauge метрик (может быть пустым)
//   - error: ошибка операции или nil при успехе
func (s *MetricsService) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return nil, err
	}

	s.logger.Debug("getting all gauge metrics")

	gauges, err := s.repository.GetAllGauges(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("all gauge metrics retrieved", "count", len(gauges))
	return gauges, nil
}

// GetAllCounters возвращает все counter метрики.
//
// Функция выполняет запрос всех counter метрик и возвращает их в виде map,
// где ключ - имя метрики, значение - её накопительное значение.
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//
// Возвращает:
//   - models.CounterMetrics: map всех counter метрик (может быть пустым)
//   - error: ошибка операции или nil при успехе
func (s *MetricsService) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return nil, err
	}

	s.logger.Debug("getting all counter metrics")

	counters, err := s.repository.GetAllCounters(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("all counter metrics retrieved", "count", len(counters))
	return counters, nil
}

// GetMetricJSON возвращает метрику в JSON формате.
//
// Функция принимает структуру Metrics с ID и типом, выполняет поиск метрики
// и возвращает полную структуру с данными.
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - metric: структура метрики с ID и типом для поиска
//
// Возвращает:
//   - *models.Metrics: полная структура метрики с данными
//   - error: ошибка операции или nil при успехе
func (s *MetricsService) GetMetricJSON(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return nil, err
	}

	// Валидируем входные параметры
	if metric == nil {
		return nil, fmt.Errorf("metric cannot be nil")
	}
	if err := s.validateMetricID(metric.ID); err != nil {
		return nil, err
	}

	s.logger.Info("getting metric as JSON", "id", metric.ID, "type", metric.MType)

	result := &models.Metrics{
		ID:    metric.ID,
		MType: metric.MType,
	}

	switch metric.MType {
	case models.Gauge:
		value, exists, err := s.GetGauge(ctx, metric.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("gauge metric not found: %s", metric.ID)
		}
		result.Value = &value

	case models.Counter:
		value, exists, err := s.GetCounter(ctx, metric.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("counter metric not found: %s", metric.ID)
		}
		result.Delta = &value

	default:
		return nil, fmt.Errorf(ErrMsgUnsupportedMetricType+": %s", metric.MType)
	}

	return result, nil
}

// UpdateMetricsBatch обновляет множество метрик в рамках одной транзакции.
//
// Функция принимает слайс метрик и обновляет их все в рамках одной транзакции.
// Это позволяет избежать race conditions и обеспечивает атомарность операции.
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - metrics: слайс метрик для обновления
//
// Возвращает:
//   - error: ошибка операции или nil при успехе
//
// Пример использования:
//
//	metrics := []models.Metrics{
//	    {ID: "temperature", MType: "gauge", Value: &temp},
//	    {ID: "requests", MType: "counter", Delta: &count},
//	}
//	err := service.UpdateMetricsBatch(ctx, metrics)
func (s *MetricsService) UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error {
	// Проверяем отмену контекста
	if err := s.checkContextCancellation(ctx); err != nil {
		return err
	}

	// Валидируем входные параметры
	if metrics == nil {
		return fmt.Errorf("metrics slice cannot be nil")
	}

	// Проверяем, что слайс не пустой
	if len(metrics) == 0 {
		return fmt.Errorf("metrics slice cannot be empty")
	}

	s.logger.Info("updating metrics batch", "count", len(metrics))

	// Валидируем каждую метрику в батче
	for i, metric := range metrics {
		if err := s.validateMetricID(metric.ID); err != nil {
			return fmt.Errorf("validation error for metric at index %d: %w", i, err)
		}

		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("validation error for metric at index %d: value is required for gauge metric", i)
			}
		case models.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("validation error for metric at index %d: delta is required for counter metric", i)
			}
		default:
			return fmt.Errorf("validation error for metric at index %d: unsupported metric type: %s", i, metric.MType)
		}
	}

	// Обновляем все метрики в рамках одной транзакции
	err := s.repository.UpdateMetricsBatch(ctx, metrics)
	if err != nil {
		s.logger.Error("failed to update metrics batch", "count", len(metrics), "error", err)
		return err
	}

	s.logger.Info("metrics batch updated successfully", "count", len(metrics))
	return nil
}
