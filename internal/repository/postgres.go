package repository

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/IgorKilipenko/metrical/internal/config/db"
	"github.com/IgorKilipenko/metrical/internal/logger"
	models "github.com/IgorKilipenko/metrical/internal/model"
	"github.com/IgorKilipenko/metrical/internal/retry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL запросы для работы с метриками
const (
	// Запросы для gauge метрик
	insertGaugeQuery = `
		INSERT INTO gauge_metrics (id, value, updated_at) 
		VALUES ($1, $2, NOW())
		ON CONFLICT (id) 
		DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`

	selectGaugeQuery = `
		SELECT value FROM gauge_metrics WHERE id = $1`

	selectAllGaugesQuery = `
		SELECT id, value FROM gauge_metrics`

	// Запросы для counter метрик
	insertCounterQuery = `
		INSERT INTO counter_metrics (id, value, updated_at) 
		VALUES ($1, $2, NOW())
		ON CONFLICT (id) 
		DO UPDATE SET value = counter_metrics.value + EXCLUDED.value, updated_at = NOW()`

	selectCounterQuery = `
		SELECT value FROM counter_metrics WHERE id = $1`

	selectAllCountersQuery = `
		SELECT id, value FROM counter_metrics`
)

// Ограничения для операций
const (
	// Максимальный размер батча для обновления метрик
	maxBatchSize = 1000
	// Интервал проверки контекста в циклах (каждые N итераций)
	contextCheckInterval = 100
)

// DatabasePool интерфейс для работы с пулом соединений базы данных.
//
// Предоставляет абстракцию для лучшей тестируемости и позволяет
// использовать mock-объекты в тестах вместо реального подключения к БД.
//
// Интерфейс включает все основные операции для работы с PostgreSQL:
// - Выполнение SQL команд (Exec)
// - Получение одной строки (QueryRow)
// - Получение множества строк (Query)
// - Управление транзакциями (Begin)
// - Проверка соединения (Ping)
// - Управление ресурсами (Close)
// - Получение статистики (Stat)
type DatabasePool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Close()
	Stat() *pgxpool.Stat
}

// PostgreSQLMetricsRepository реализация репозитория для PostgreSQL.
//
// Предоставляет полную реализацию интерфейса MetricsRepository для работы
// с метриками в PostgreSQL базе данных. Использует connection pooling
// для эффективного управления соединениями с БД.
//
// Основные возможности:
// - Обновление gauge и counter метрик
// - Получение отдельных метрик по имени
// - Получение всех метрик определенного типа
// - Валидация входных данных
// - Структурированное логирование
// - Обработка контекста и отмены операций
//
// Пример использования:
//
//	pool, err := pgxpool.New(ctx, dsn)
//	if err != nil {
//	    return err
//	}
//	logger := logger.New()
//	repo := NewPostgreSQLMetricsRepository(pool, logger)
//
//	err = repo.UpdateGauge(ctx, "temperature", 23.5)
//	if err != nil {
//	    log.Printf("Failed to update metric: %v", err)
//	}
type PostgreSQLMetricsRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewPostgreSQLMetricsRepository создает новый экземпляр PostgreSQLMetricsRepository.
//
// Функция принимает готовый пул соединений с PostgreSQL и логгер,
// возвращает готовый к использованию репозиторий для работы с метриками.
//
// Параметры:
//   - pool: пул соединений с PostgreSQL (должен быть уже инициализирован)
//   - logger: логгер для записи событий и ошибок
//
// Возвращает:
//   - *PostgreSQLMetricsRepository: готовый репозиторий
//
// Пример использования:
//
//	pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost/db")
//	if err != nil {
//	    return err
//	}
//	logger := logger.New()
//	repo := NewPostgreSQLMetricsRepository(pool, logger)
//
// Примечания:
//   - Пул соединений должен быть уже создан и готов к использованию
//   - Логгер не должен быть nil
//   - Репозиторий не закрывает пул соединений - это ответственность вызывающего кода
func NewPostgreSQLMetricsRepository(pool *pgxpool.Pool, logger logger.Logger) *PostgreSQLMetricsRepository {
	return &PostgreSQLMetricsRepository{
		pool:   pool,
		logger: logger,
	}
}

// NewPostgreSQLMetricsRepositoryWithPool создает новый экземпляр с интерфейсом DatabasePool
// Используется для лучшей тестируемости
func NewPostgreSQLMetricsRepositoryWithPool(pool DatabasePool, logger logger.Logger) (*PostgreSQLMetricsRepository, error) {
	// Приводим к конкретному типу для внутреннего использования
	pgxPool, ok := pool.(*pgxpool.Pool)
	if !ok {
		return nil, fmt.Errorf("pool must be *pgxpool.Pool, got %T", pool)
	}

	return &PostgreSQLMetricsRepository{
		pool:   pgxPool,
		logger: logger,
	}, nil
}

// NewPostgreSQLMetricsRepositoryWithMigrations создает новый экземпляр PostgreSQL репозитория с автоматическими миграциями.
//
// Принимает:
// - pool: пул соединений с PostgreSQL базой данных
// - logger: логгер для записи событий и ошибок
//
// Возвращает:
// - *PostgreSQLMetricsRepository: новый экземпляр репозитория
// - error: ошибка при выполнении миграций
//
// Автоматически выполняет миграции для создания необходимых таблиц.
func NewPostgreSQLMetricsRepositoryWithMigrations(pool DatabasePool, logger logger.Logger) (*PostgreSQLMetricsRepository, error) {
	ctx := context.Background()

	// Приводим к конкретному типу для миграций
	pgxPool, ok := pool.(*pgxpool.Pool)
	if !ok {
		return nil, fmt.Errorf("pool must be *pgxpool.Pool for migrations")
	}

	// Создаем менеджер миграций
	migrationManager := db.NewMigrationManager(pgxPool, logger)

	// Загружаем миграции из файловой системы
	logger.Info("Loading migrations from filesystem")
	migrations, err := migrationManager.LoadMigrationsFromFS(os.DirFS("."), "migrations")
	if err != nil {
		logger.Error("Failed to load migrations", "error", err)
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	logger.Info("Loaded migrations", "count", len(migrations))

	// Выполняем миграции
	logger.Info("Running migrations")
	err = migrationManager.RunMigrations(ctx, migrations)
	if err != nil {
		logger.Error("Failed to run migrations", "error", err)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("PostgreSQL repository initialized with migrations")

	return &PostgreSQLMetricsRepository{
		pool:   pgxPool,
		logger: logger,
	}, nil
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

// UpdateGauge обновляет значение gauge метрики в базе данных.
//
// Gauge метрики представляют мгновенные значения (например, температура, память).
// При обновлении gauge метрики новое значение полностью заменяет старое.
//
// Функция выполняет следующие действия:
//  1. Валидирует входные параметры (имя метрики и значение)
//  2. Проверяет отмену контекста
//  3. Выполняет UPSERT операцию в БД (INSERT или UPDATE)
//  4. Логирует результат операции
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - name: имя метрики (не может быть пустым)
//   - value: новое значение метрики (не может быть NaN или бесконечностью)
//
// Возвращает:
//   - error: ошибка операции или nil при успехе
//
// Пример использования:
//
//	err := repo.UpdateGauge(ctx, "temperature", 23.5)
//	if err != nil {
//	    log.Printf("Failed to update gauge: %v", err)
//	}
//
// Возможные ошибки:
//   - "metric name cannot be empty": пустое имя метрики
//   - "gauge value cannot be NaN": значение NaN
//   - "gauge value cannot be infinite": бесконечное значение
//   - context.DeadlineExceeded: превышен таймаут
//   - context.Canceled: операция отменена
//   - Ошибки базы данных: проблемы с подключением или SQL
func (r *PostgreSQLMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.logger.Debug("operation completed", "operation", "UpdateGauge", "duration", duration)
	}()

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

	// Используем retry логику для операций с базой данных
	return retry.Retry(ctx, r.logger, retry.DefaultRetryConfig, func() error {
		_, err := r.pool.Exec(ctx, insertGaugeQuery, name, value)
		if err != nil {
			r.logger.Error("failed to update gauge metric", "name", name, "value", value, "error", err)
			return err
		}

		r.logger.Debug("updated gauge metric", "name", name, "value", value)
		return nil
	})
}

// UpdateCounter добавляет значение к counter метрике в базе данных.
//
// Counter метрики представляют накопительные значения (например, количество запросов).
// При обновлении counter метрики новое значение добавляется к существующему.
//
// Функция выполняет следующие действия:
//  1. Валидирует входные параметры (имя метрики)
//  2. Проверяет отмену контекста
//  3. Выполняет атомарную UPSERT операцию с инкрементом
//  4. Логирует результат операции
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - name: имя метрики (не может быть пустым)
//   - value: значение для добавления к существующему счетчику
//
// Возвращает:
//   - error: ошибка операции или nil при успехе
//
// Пример использования:
//
//	err := repo.UpdateCounter(ctx, "requests_total", 1)
//	if err != nil {
//	    log.Printf("Failed to update counter: %v", err)
//	}
//
// Примечания:
//   - Операция атомарна - нет race conditions при concurrent доступе
//   - Если метрика не существует, создается новая с переданным значением
//   - Если метрика существует, значение добавляется к текущему
//
// Возможные ошибки:
//   - "metric name cannot be empty": пустое имя метрики
//   - context.DeadlineExceeded: превышен таймаут
//   - context.Canceled: операция отменена
//   - Ошибки базы данных: проблемы с подключением или SQL
func (r *PostgreSQLMetricsRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.logger.Debug("operation completed", "operation", "UpdateCounter", "duration", duration)
	}()

	// Проверяем валидность входных данных
	if err := r.validateMetricName(name); err != nil {
		return err
	}

	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "counter update"); err != nil {
		return err
	}

	// Используем retry логику для операций с базой данных
	return retry.Retry(ctx, r.logger, retry.DefaultRetryConfig, func() error {
		// Используем оптимизированный SQL запрос для атомарного обновления
		_, err := r.pool.Exec(ctx, insertCounterQuery, name, value)

		if err != nil {
			r.logger.Error("failed to update counter metric", "name", name, "value", value, "error", err)
			return err
		}

		r.logger.Debug("updated counter metric", "name", name, "value", value)
		return nil
	})
}

// validateBatch проверяет валидность входных параметров для batch операции
func (r *PostgreSQLMetricsRepository) validateBatch(metrics []models.Metrics) error {
	if metrics == nil {
		return fmt.Errorf("metrics slice cannot be nil")
	}
	if len(metrics) == 0 {
		return fmt.Errorf("metrics slice cannot be empty")
	}
	if len(metrics) > maxBatchSize {
		return fmt.Errorf("batch size %d exceeds maximum allowed size %d", len(metrics), maxBatchSize)
	}
	return nil
}

// beginTransactionWithRetry начинает транзакцию с retry логикой
func (r *PostgreSQLMetricsRepository) beginTransactionWithRetry(ctx context.Context) (pgx.Tx, error) {
	var tx pgx.Tx
	err := retry.Retry(ctx, r.logger, retry.DefaultRetryConfig, func() error {
		var beginErr error
		tx, beginErr = r.pool.Begin(ctx)
		if beginErr != nil {
			r.logger.Error("failed to begin transaction for batch update", "error", beginErr)
			return fmt.Errorf("failed to begin transaction: %w", beginErr)
		}
		return nil
	})
	return tx, err
}

// finalizeTransaction завершает транзакцию (коммит или rollback)
func (r *PostgreSQLMetricsRepository) finalizeTransaction(tx pgx.Tx, ctx context.Context, txErr *error) {
	if *txErr != nil {
		// Откатываем транзакцию при ошибке
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			r.logger.Error("failed to rollback transaction", "error", rollbackErr)
		}
	} else {
		// Коммитим транзакцию при успехе
		if commitErr := tx.Commit(ctx); commitErr != nil {
			r.logger.Error("failed to commit transaction", "error", commitErr)
			// Если коммит не удался, пытаемся откатить
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				r.logger.Error("failed to rollback after commit failure", "error", rollbackErr)
			}
		}
	}
}

// updateMetricInTransaction обновляет одну метрику в рамках транзакции
func (r *PostgreSQLMetricsRepository) updateMetricInTransaction(ctx context.Context, tx pgx.Tx, metric models.Metrics) error {
	// Валидируем имя метрики
	if err := r.validateMetricName(metric.ID); err != nil {
		return fmt.Errorf("validation error for metric %s: %w", metric.ID, err)
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("validation error for gauge metric %s: value is required", metric.ID)
		}

		// Валидируем значение gauge
		if err := r.validateGaugeValue(*metric.Value); err != nil {
			return fmt.Errorf("validation error for gauge metric %s: %w", metric.ID, err)
		}

		_, err := tx.Exec(ctx, insertGaugeQuery, metric.ID, *metric.Value)
		if err != nil {
			r.logger.Error("failed to update gauge metric in batch", "id", metric.ID, "value", *metric.Value, "error", err)
			return fmt.Errorf("failed to update gauge metric %s: %w", metric.ID, err)
		}
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("validation error for counter metric %s: delta is required", metric.ID)
		}

		_, err := tx.Exec(ctx, insertCounterQuery, metric.ID, *metric.Delta)
		if err != nil {
			r.logger.Error("failed to update counter metric in batch", "id", metric.ID, "delta", *metric.Delta, "error", err)
			return fmt.Errorf("failed to update counter metric %s: %w", metric.ID, err)
		}
	default:
		return fmt.Errorf("unsupported metric type: %s", metric.MType)
	}

	return nil
}

// UpdateMetricsBatch обновляет множество метрик в рамках одной транзакции.
//
// Функция принимает слайс метрик и обновляет их все в рамках одной транзакции.
// Это позволяет избежать race conditions и обеспечивает атомарность операции.
//
// Функция выполняет следующие действия:
//  1. Валидирует входные параметры
//  2. Проверяет отмену контекста
//  3. Начинает транзакцию
//  4. Обновляет все метрики в рамках транзакции
//  5. Коммитит транзакцию или откатывает при ошибке
//  6. Логирует результат операции
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
//	err := repo.UpdateMetricsBatch(ctx, metrics)
//
// Примечания:
//   - Все операции выполняются в рамках одной транзакции
//   - При ошибке в любой метрике вся транзакция откатывается
//   - Операция атомарна - нет race conditions
//
// Возможные ошибки:
//   - "metrics slice cannot be nil": nil слайс метрик
//   - "metrics slice cannot be empty": пустой слайс метрик
//   - "batch size X exceeds maximum allowed size 1000": превышен лимит размера батча
//   - context.DeadlineExceeded: превышен таймаут
//   - context.Canceled: операция отменена
//   - Ошибки базы данных: проблемы с подключением или SQL
func (r *PostgreSQLMetricsRepository) UpdateMetricsBatch(ctx context.Context, metrics []models.Metrics) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.logger.Debug("operation completed", "operation", "UpdateMetricsBatch", "duration", duration, "count", len(metrics))
	}()

	// Валидируем входные параметры
	if err := r.validateBatch(metrics); err != nil {
		return err
	}

	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "batch update"); err != nil {
		return err
	}

	// Начинаем транзакцию с retry логикой
	tx, err := r.beginTransactionWithRetry(ctx)
	if err != nil {
		return err
	}

	// Используем отдельную переменную для отслеживания ошибок
	var txErr error
	defer r.finalizeTransaction(tx, ctx, &txErr)

	// Обновляем все метрики в рамках транзакции
	for _, metric := range metrics {
		if err := r.updateMetricInTransaction(ctx, tx, metric); err != nil {
			txErr = err
			return txErr
		}
	}

	// Транзакция будет закоммичена в defer функции
	r.logger.Debug("batch update completed successfully", "count", len(metrics))
	return nil
}

// GetGauge возвращает значение gauge метрики из базы данных.
//
// Функция выполняет поиск gauge метрики по имени и возвращает её значение.
// Если метрика не найдена, возвращается false в качестве второго значения.
//
// Функция выполняет следующие действия:
//  1. Валидирует входные параметры (имя метрики)
//  2. Проверяет отмену контекста
//  3. Выполняет SELECT запрос в БД
//  4. Обрабатывает случай отсутствия метрики
//  5. Логирует результат операции
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - name: имя метрики для поиска
//
// Возвращает:
//   - float64: значение метрики (0 если не найдена)
//   - bool: true если метрика найдена, false если нет
//   - error: ошибка операции или nil при успехе
//
// Пример использования:
//
//	value, exists, err := repo.GetGauge(ctx, "temperature")
//	if err != nil {
//	    log.Printf("Failed to get gauge: %v", err)
//	    return
//	}
//	if exists {
//	    log.Printf("Temperature: %.2f", value)
//	} else {
//	    log.Printf("Temperature metric not found")
//	}
//
// Возможные ошибки:
//   - "metric name cannot be empty": пустое имя метрики
//   - context.DeadlineExceeded: превышен таймаут
//   - context.Canceled: операция отменена
//   - Ошибки базы данных: проблемы с подключением или SQL
func (r *PostgreSQLMetricsRepository) GetGauge(ctx context.Context, name string) (float64, bool, error) {
	// Проверяем валидность входных данных
	if err := r.validateMetricName(name); err != nil {
		return 0, false, err
	}

	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "gauge retrieval"); err != nil {
		return 0, false, err
	}

	// Используем retry логику для операций с базой данных
	return retry.RetryWithResultAndExists(ctx, r.logger, retry.DefaultRetryConfig, func() (float64, bool, error) {
		var value float64
		err := r.pool.QueryRow(ctx, selectGaugeQuery, name).Scan(&value)

		if err == pgx.ErrNoRows {
			r.logger.Debug("gauge metric not found", "name", name)
			return 0, false, nil
		}

		if err != nil {
			r.logger.Error("failed to get gauge metric", "name", name, "error", err)
			return 0, false, err
		}

		r.logger.Debug("retrieved gauge metric", "name", name, "value", value)
		return value, true, nil
	})
}

// GetCounter возвращает значение counter метрики из базы данных.
//
// Функция выполняет поиск counter метрики по имени и возвращает её накопительное значение.
// Если метрика не найдена, возвращается false в качестве второго значения.
//
// Функция выполняет следующие действия:
//  1. Валидирует входные параметры (имя метрики)
//  2. Проверяет отмену контекста
//  3. Выполняет SELECT запрос в БД
//  4. Обрабатывает случай отсутствия метрики
//  5. Логирует результат операции
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - name: имя метрики для поиска
//
// Возвращает:
//   - int64: накопительное значение метрики (0 если не найдена)
//   - bool: true если метрика найдена, false если нет
//   - error: ошибка операции или nil при успехе
//
// Пример использования:
//
//	value, exists, err := repo.GetCounter(ctx, "requests_total")
//	if err != nil {
//	    log.Printf("Failed to get counter: %v", err)
//	    return
//	}
//	if exists {
//	    log.Printf("Total requests: %d", value)
//	} else {
//	    log.Printf("Requests counter not found")
//	}
//
// Примечания:
//   - Возвращаемое значение представляет общую сумму всех инкрементов
//   - Если метрика не существует, возвращается 0 и false
//
// Возможные ошибки:
//   - "metric name cannot be empty": пустое имя метрики
//   - context.DeadlineExceeded: превышен таймаут
//   - context.Canceled: операция отменена
//   - Ошибки базы данных: проблемы с подключением или SQL
func (r *PostgreSQLMetricsRepository) GetCounter(ctx context.Context, name string) (int64, bool, error) {
	// Проверяем валидность входных данных
	if err := r.validateMetricName(name); err != nil {
		return 0, false, err
	}

	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "counter retrieval"); err != nil {
		return 0, false, err
	}

	// Используем retry логику для операций с базой данных
	return retry.RetryWithResultAndExists(ctx, r.logger, retry.DefaultRetryConfig, func() (int64, bool, error) {
		var value int64
		err := r.pool.QueryRow(ctx, selectCounterQuery, name).Scan(&value)

		if err == pgx.ErrNoRows {
			r.logger.Debug("counter metric not found", "name", name)
			return 0, false, nil
		}

		if err != nil {
			r.logger.Error("failed to get counter metric", "name", name, "error", err)
			return 0, false, err
		}

		r.logger.Debug("retrieved counter metric", "name", name, "value", value)
		return value, true, nil
	})
}

// GetAllGauges возвращает все gauge метрики из базы данных.
//
// Функция выполняет запрос всех gauge метрик и возвращает их в виде map,
// где ключ - имя метрики, значение - её значение.
//
// Функция выполняет следующие действия:
//  1. Проверяет отмену контекста
//  2. Выполняет SELECT запрос всех gauge метрик
//  3. Итерирует по результатам и заполняет map
//  4. Обрабатывает ошибки сканирования и итерации
//  5. Логирует результат операции
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//
// Возвращает:
//   - models.GaugeMetrics: map всех gauge метрик (может быть пустым)
//   - error: ошибка операции или nil при успехе
//
// Пример использования:
//
//	gauges, err := repo.GetAllGauges(ctx)
//	if err != nil {
//	    log.Printf("Failed to get all gauges: %v", err)
//	    return
//	}
//	for name, value := range gauges {
//	    log.Printf("Gauge %s: %.2f", name, value)
//	}
//
// Примечания:
//   - Возвращает пустую map если метрики не найдены
//   - Автоматически закрывает rows после использования
//   - Операция выполняется в одном запросе
//
// Возможные ошибки:
//   - context.DeadlineExceeded: превышен таймаут
//   - context.Canceled: операция отменена
//   - Ошибки базы данных: проблемы с подключением или SQL
//   - Ошибки сканирования: проблемы с типами данных
func (r *PostgreSQLMetricsRepository) GetAllGauges(ctx context.Context) (models.GaugeMetrics, error) {
	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "getAllGauges"); err != nil {
		return nil, err
	}

	// Используем retry логику для операций с базой данных
	var rows pgx.Rows
	err := retry.Retry(ctx, r.logger, retry.DefaultRetryConfig, func() error {
		var queryErr error
		rows, queryErr = r.pool.Query(ctx, selectAllGaugesQuery)
		if queryErr != nil {
			r.logger.Error("failed to query all gauge metrics", "error", queryErr)
			return queryErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(models.GaugeMetrics)
	iterationCount := 0
	for rows.Next() {
		// Проверяем отмену контекста периодически
		if iterationCount%contextCheckInterval == 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				// Продолжаем итерацию
			}
		}
		iterationCount++

		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			r.logger.Error("failed to scan gauge metric", "error", err)
			return nil, err
		}
		result[name] = value
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating gauge metrics", "error", err)
		return nil, err
	}

	r.logger.Debug("retrieved all gauge metrics", "count", len(result))
	return result, nil
}

// GetAllCounters возвращает все counter метрики из базы данных.
//
// Функция выполняет запрос всех counter метрик и возвращает их в виде map,
// где ключ - имя метрики, значение - её накопительное значение.
//
// Функция выполняет следующие действия:
//  1. Проверяет отмену контекста
//  2. Выполняет SELECT запрос всех counter метрик
//  3. Итерирует по результатам и заполняет map
//  4. Обрабатывает ошибки сканирования и итерации
//  5. Логирует результат операции
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//
// Возвращает:
//   - models.CounterMetrics: map всех counter метрик (может быть пустым)
//   - error: ошибка операции или nil при успехе
//
// Пример использования:
//
//	counters, err := repo.GetAllCounters(ctx)
//	if err != nil {
//	    log.Printf("Failed to get all counters: %v", err)
//	    return
//	}
//	for name, value := range counters {
//	    log.Printf("Counter %s: %d", name, value)
//	}
//
// Примечания:
//   - Возвращает пустую map если метрики не найдены
//   - Автоматически закрывает rows после использования
//   - Операция выполняется в одном запросе
//
// Возможные ошибки:
//   - context.DeadlineExceeded: превышен таймаут
//   - context.Canceled: операция отменена
//   - Ошибки базы данных: проблемы с подключением или SQL
//   - Ошибки сканирования: проблемы с типами данных
func (r *PostgreSQLMetricsRepository) GetAllCounters(ctx context.Context) (models.CounterMetrics, error) {
	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "getAllCounters"); err != nil {
		return nil, err
	}

	// Используем retry логику для операций с базой данных
	var rows pgx.Rows
	err := retry.Retry(ctx, r.logger, retry.DefaultRetryConfig, func() error {
		var queryErr error
		rows, queryErr = r.pool.Query(ctx, selectAllCountersQuery)
		if queryErr != nil {
			r.logger.Error("failed to query all counter metrics", "error", queryErr)
			return queryErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(models.CounterMetrics)
	iterationCount := 0
	for rows.Next() {
		// Проверяем отмену контекста периодически
		if iterationCount%contextCheckInterval == 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				// Продолжаем итерацию
			}
		}
		iterationCount++

		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			r.logger.Error("failed to scan counter metric", "error", err)
			return nil, err
		}
		result[name] = value
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating counter metrics", "error", err)
		return nil, err
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

// HealthCheck проверяет состояние connection pool и соединения с базой данных.
//
// Функция выполняет следующие проверки:
//  1. Проверяет доступность базы данных через Ping
//  2. Анализирует статистику connection pool
//  3. Проверяет, не приближается ли пул к лимиту соединений
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//
// Возвращает:
//   - error: ошибка при проблемах с БД или пулом соединений, nil при успехе
//
// Пример использования:
//
//	err := repo.HealthCheck(ctx)
//	if err != nil {
//	    log.Printf("Health check failed: %v", err)
//	}
//
// Возможные ошибки:
//   - context.DeadlineExceeded: превышен таймаут
//   - context.Canceled: операция отменена
//   - "connection pool near capacity": пул соединений близок к лимиту
//   - Ошибки базы данных: проблемы с подключением
func (r *PostgreSQLMetricsRepository) HealthCheck(ctx context.Context) error {
	// Проверяем отмену контекста
	if err := r.checkContext(ctx, "health check"); err != nil {
		return err
	}

	// Проверяем доступность базы данных
	if err := r.pool.Ping(ctx); err != nil {
		r.logger.Error("health check failed: database ping failed", "error", err)
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Анализируем статистику connection pool
	stats := r.pool.Stat()

	// Проверяем, не приближается ли пул к лимиту соединений
	if stats.AcquireCount() > int64(float64(stats.MaxConns())*0.9) {
		r.logger.Warn("connection pool near capacity",
			"acquire_count", stats.AcquireCount(),
			"max_conns", stats.MaxConns())
		return fmt.Errorf("connection pool near capacity: %d/%d connections",
			stats.AcquireCount(), stats.MaxConns())
	}

	r.logger.Debug("health check passed",
		"acquire_count", stats.AcquireCount(),
		"max_conns", stats.MaxConns(),
		"idle_conns", stats.IdleConns())

	return nil
}
