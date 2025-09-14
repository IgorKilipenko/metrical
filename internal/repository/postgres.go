package repository

import (
	"context"
	"fmt"
	"math"

	"github.com/IgorKilipenko/metrical/internal/logger"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabasePool интерфейс для работы с пулом соединений базы данных
// Предоставляет абстракцию для лучшей тестируемости
type DatabasePool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Close()
	Stat() *pgxpool.Stat
}

// PostgreSQLMetricsRepository реализация репозитория для PostgreSQL
type PostgreSQLMetricsRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewPostgreSQLMetricsRepository создает новый экземпляр PostgreSQLMetricsRepository
func NewPostgreSQLMetricsRepository(pool *pgxpool.Pool, logger logger.Logger) *PostgreSQLMetricsRepository {
	return &PostgreSQLMetricsRepository{
		pool:   pool,
		logger: logger,
	}
}

// NewPostgreSQLMetricsRepositoryWithPool создает новый экземпляр с интерфейсом DatabasePool
// Используется для лучшей тестируемости
func NewPostgreSQLMetricsRepositoryWithPool(pool DatabasePool, logger logger.Logger) *PostgreSQLMetricsRepository {
	// Приводим к конкретному типу для внутреннего использования
	pgxPool, ok := pool.(*pgxpool.Pool)
	if !ok {
		panic("pool must be *pgxpool.Pool")
	}

	return &PostgreSQLMetricsRepository{
		pool:   pgxPool,
		logger: logger,
	}
}

// checkContext проверяет отмену контекста и возвращает ошибку если контекст отменен
func (r *PostgreSQLMetricsRepository) checkContext(ctx context.Context, operation string) error {
	select {
	case <-ctx.Done():
		r.logger.Debug("context cancelled during " + operation)
		return ctx.Err()
	default:
		return nil
	}
}

// validateMetricName проверяет валидность имени метрики
func (r *PostgreSQLMetricsRepository) validateMetricName(name string) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	return nil
}

// validateGaugeValue проверяет валидность значения gauge метрики
func (r *PostgreSQLMetricsRepository) validateGaugeValue(value float64) error {
	if math.IsNaN(value) {
		return fmt.Errorf("gauge value cannot be NaN")
	}
	if math.IsInf(value, 0) {
		return fmt.Errorf("gauge value cannot be infinite")
	}
	return nil
}

// UpdateGauge обновляет значение gauge метрики
func (r *PostgreSQLMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	// Проверяем валидность входных данных
	if err := r.validateMetricName(name); err != nil {
		return err
	}
	if err := r.validateGaugeValue(value); err != nil {
		return err
	}

	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "gauge update"); err != nil {
		return err
	}

	query := `
		INSERT INTO metrics (name, type, value, updated_at) 
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (name, type) 
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`

	_, err := r.pool.Exec(ctx, query, name, models.Gauge, value)
	if err != nil {
		r.logger.Error("failed to update gauge metric", "name", name, "value", value, "error", err)
		return fmt.Errorf("failed to update gauge metric: %w", err)
	}

	r.logger.Debug("updated gauge metric", "name", name, "value", value)
	return nil
}

// UpdateCounter добавляет значение к counter метрике
func (r *PostgreSQLMetricsRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	// Проверяем валидность входных данных
	if err := r.validateMetricName(name); err != nil {
		return err
	}

	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "counter update"); err != nil {
		return err
	}

	// Используем оптимизированный SQL запрос для атомарного обновления
	_, err := r.pool.Exec(ctx, `
		INSERT INTO metrics (name, type, delta, updated_at) 
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (name, type) 
		DO UPDATE SET delta = metrics.delta + EXCLUDED.delta, updated_at = NOW()`,
		name, models.Counter, value)

	if err != nil {
		r.logger.Error("failed to update counter metric", "name", name, "value", value, "error", err)
		return fmt.Errorf("failed to update counter metric: %w", err)
	}

	r.logger.Debug("updated counter metric", "name", name, "value", value)
	return nil
}

// GetGauge возвращает значение gauge метрики
func (r *PostgreSQLMetricsRepository) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	// Проверяем валидность входных данных
	if err := r.validateMetricName(name); err != nil {
		return 0, false, err
	}

	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "gauge retrieval"); err != nil {
		return 0, false, err
	}

	var value float64
	err := r.pool.QueryRow(ctx,
		"SELECT value FROM metrics WHERE name = $1 AND type = $2",
		name, models.Gauge).Scan(&value)

	if err == pgx.ErrNoRows {
		r.logger.Debug("gauge metric not found", "name", name)
		return 0, false, nil
	}

	if err != nil {
		r.logger.Error("failed to get gauge metric", "name", name, "error", err)
		return 0, false, fmt.Errorf("failed to get gauge metric: %w", err)
	}

	r.logger.Debug("retrieved gauge metric", "name", name, "value", value)
	return value, true, nil
}

// GetCounter возвращает значение counter метрики
func (r *PostgreSQLMetricsRepository) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	// Проверяем валидность входных данных
	if err := r.validateMetricName(name); err != nil {
		return 0, false, err
	}

	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "counter retrieval"); err != nil {
		return 0, false, err
	}

	var value int64
	err := r.pool.QueryRow(ctx,
		"SELECT COALESCE(delta, 0) FROM metrics WHERE name = $1 AND type = $2",
		name, models.Counter).Scan(&value)

	if err == pgx.ErrNoRows {
		r.logger.Debug("counter metric not found", "name", name)
		return 0, false, nil
	}

	if err != nil {
		r.logger.Error("failed to get counter metric", "name", name, "error", err)
		return 0, false, fmt.Errorf("failed to get counter metric: %w", err)
	}

	r.logger.Debug("retrieved counter metric", "name", name, "value", value)
	return value, true, nil
}

// GetAllGauges возвращает все gauge метрики
func (r *PostgreSQLMetricsRepository) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "getAllGauges"); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx,
		"SELECT name, value FROM metrics WHERE type = $1", models.Gauge)
	if err != nil {
		r.logger.Error("failed to get all gauge metrics", "error", err)
		return nil, fmt.Errorf("failed to get all gauge metrics: %w", err)
	}
	defer rows.Close()

	result := make(models.GaugeMetrics)
	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			r.logger.Error("failed to scan gauge metric", "error", err)
			return nil, fmt.Errorf("failed to scan gauge metric: %w", err)
		}
		result[name] = value
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating gauge metrics", "error", err)
		return nil, fmt.Errorf("error iterating gauge metrics: %w", err)
	}

	r.logger.Debug("retrieved all gauge metrics", "count", len(result))
	return result, nil
}

// GetAllCounters возвращает все counter метрики
func (r *PostgreSQLMetricsRepository) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "getAllCounters"); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx,
		"SELECT name, COALESCE(delta, 0) FROM metrics WHERE type = $1", models.Counter)
	if err != nil {
		r.logger.Error("failed to get all counter metrics", "error", err)
		return nil, fmt.Errorf("failed to get all counter metrics: %w", err)
	}
	defer rows.Close()

	result := make(models.CounterMetrics)
	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			r.logger.Error("failed to scan counter metric", "error", err)
			return nil, fmt.Errorf("failed to scan counter metric: %w", err)
		}
		result[name] = value
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating counter metrics", "error", err)
		return nil, fmt.Errorf("error iterating counter metrics: %w", err)
	}

	r.logger.Debug("retrieved all counter metrics", "count", len(result))
	return result, nil
}

// SaveToFile - заглушка для совместимости с интерфейсом
// В PostgreSQL версии файловое сохранение не используется
func (r *PostgreSQLMetricsRepository) SaveToFile() error {
	r.logger.Debug("SaveToFile called on PostgreSQL repository - no action needed")
	return nil
}

// LoadFromFile - заглушка для совместимости с интерфейсом
// В PostgreSQL версии файловая загрузка не используется
func (r *PostgreSQLMetricsRepository) LoadFromFile() error {
	r.logger.Debug("LoadFromFile called on PostgreSQL repository - no action needed")
	return nil
}

// SetSyncSave - заглушка для совместимости с интерфейсом
// В PostgreSQL версии синхронное сохранение не применимо
func (r *PostgreSQLMetricsRepository) SetSyncSave(sync bool) {
	r.logger.Debug("SetSyncSave called on PostgreSQL repository", "sync", sync)
}
