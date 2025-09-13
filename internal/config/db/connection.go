package db

import (
	"context"
	"fmt"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseConnection интерфейс для работы с базой данных.
// Предоставляет методы для управления подключением к PostgreSQL.
type DatabaseConnection interface {
	// Pool возвращает пул соединений для выполнения SQL запросов
	Pool() *pgxpool.Pool

	// Ping проверяет доступность базы данных
	Ping(ctx context.Context) error

	// PingWithRetry выполняет ping с повторными попытками при сбоях
	PingWithRetry(ctx context.Context, maxRetries int) error

	// Close закрывает все соединения и освобождает ресурсы
	Close()

	// Stats возвращает статистику использования пула соединений
	Stats() *pgxpool.Stat

	// HealthCheck выполняет комплексную проверку здоровья БД
	HealthCheck(ctx context.Context) error

	// HealthCheckWithRetry выполняет health check с повторными попытками
	HealthCheckWithRetry(ctx context.Context, maxRetries int) error
}

// Connection представляет подключение к базе данных
type Connection struct {
	pool   *pgxpool.Pool
	config Config
	logger logger.Logger
}

// NewConnection создает новое подключение к базе данных.
//
// Функция выполняет следующие действия:
//  1. Валидирует конфигурацию подключения
//  2. Создает пул соединений с настройками из config
//  3. Проверяет доступность базы данных через ping
//  4. Возвращает готовое к использованию подключение
//
// Параметры:
//   - config: конфигурация подключения (DSN, таймауты, размер пула)
//   - logger: логгер для записи событий подключения
//
// Возвращает:
//   - *Connection: готовое подключение к БД
//   - error: ошибка создания подключения или nil при успехе
//
// Пример использования:
//
//	config := db.DefaultConfig()
//	config.DSN = "postgres://user:pass@localhost:5432/db"
//	logger := logger.New()
//
//	conn, err := db.NewConnection(config, logger)
//	if err != nil {
//	    log.Fatalf("Failed to connect to database: %v", err)
//	}
//	defer conn.Close()
//
// Возможные ошибки:
//   - "invalid database config": некорректная конфигурация
//   - "failed to parse database config": ошибка парсинга DSN
//   - "failed to create connection pool": ошибка создания пула
//   - "failed to ping database": база данных недоступна
func NewConnection(config Config, logger logger.Logger) (*Connection, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	logger.Info("connecting to database", "config", config.String())

	// Создаем конфигурацию пула соединений
	poolConfig, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Настраиваем пул соединений
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime

	// Создаем контекст с таймаутом для подключения
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// Создаем пул соединений
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("successfully connected to database",
		"max_conns", config.MaxConns,
		"min_conns", config.MinConns)

	return &Connection{
		pool:   pool,
		config: config,
		logger: logger,
	}, nil
}

// Pool возвращает пул соединений
func (c *Connection) Pool() *pgxpool.Pool {
	return c.pool
}

// Ping проверяет соединение с базой данных.
//
// Выполняет быструю проверку доступности базы данных через TCP соединение.
// Использует таймаут из конфигурации (config.PingTimeout).
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//
// Возвращает:
//   - error: ошибка подключения или nil при успехе
//
// Пример использования:
//
//	ctx := context.Background()
//	err := conn.Ping(ctx)
//	if err != nil {
//	    log.Printf("Database is not available: %v", err)
//	}
//
// Возможные ошибки:
//   - "database ping failed": база данных недоступна
//   - context.DeadlineExceeded: превышен таймаут
//   - context.Canceled: операция отменена
func (c *Connection) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.PingTimeout)
	defer cancel()

	if err := c.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	c.logger.Debug("database ping successful")
	return nil
}

// Close закрывает соединение с базой данных и освобождает ресурсы.
//
// Безопасно закрывает все соединения в пуле и освобождает связанные ресурсы.
// Функция идемпотентна - можно вызывать многократно без побочных эффектов.
// Рекомендуется вызывать при завершении работы приложения.
//
// Пример использования:
//
//	conn, err := db.NewConnection(config, logger)
//	if err != nil {
//	    return err
//	}
//	defer conn.Close() // Гарантированное закрытие при выходе
//
// Примечания:
//   - После вызова Close() все операции с подключением будут возвращать ошибки
//   - Функция не возвращает ошибок, так как закрытие всегда возможно
func (c *Connection) Close() {
	if c.pool != nil {
		c.logger.Info("closing database connection pool")
		c.pool.Close()
	}
}

// Stats возвращает статистику пула соединений
func (c *Connection) Stats() *pgxpool.Stat {
	return c.pool.Stat()
}

// HealthCheck выполняет комплексную проверку здоровья базы данных.
//
// Выполняет двухэтапную проверку:
//  1. Ping - проверка TCP соединения
//  2. SQL запрос - проверка возможности выполнения запросов
//
// Использует таймаут из конфигурации (config.HealthCheckTimeout).
// Рекомендуется для health check endpoints в микросервисах.
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//
// Возвращает:
//   - error: ошибка проверки или nil при успехе
//
// Пример использования:
//
//	ctx := context.Background()
//	err := conn.HealthCheck(ctx)
//	if err != nil {
//	    // База данных не готова к работе
//	    http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
//	    return
//	}
//	// База данных готова к работе
//
// Возможные ошибки:
//   - "database ping failed": TCP соединение недоступно
//   - "database health check failed": SQL запрос не выполнен
//   - "unexpected health check result": неожиданный результат запроса
//   - context.DeadlineExceeded: превышен таймаут
func (c *Connection) HealthCheck(ctx context.Context) error {
	// Проверяем ping
	if err := c.Ping(ctx); err != nil {
		return err
	}

	// Выполняем простой запрос для проверки работоспособности
	ctx, cancel := context.WithTimeout(ctx, c.config.HealthCheckTimeout)
	defer cancel()

	var result int
	err := c.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}

	c.logger.Debug("database health check successful")
	return nil
}

// PingWithRetry выполняет ping с повторными попытками при сбоях.
//
// Реализует retry логику с экспоненциальной задержкой:
//   - 1-я попытка: немедленно
//   - 2-я попытка: через 1 секунду
//   - 3-я попытка: через 2 секунды
//   - N-я попытка: через (N-1) секунд
//
// Поддерживает отмену через контекст между попытками.
// Рекомендуется для критически важных операций.
//
// Параметры:
//   - ctx: контекст с возможностью отмены операции
//   - maxRetries: максимальное количество попыток (включая первую)
//
// Возвращает:
//   - error: ошибка после всех попыток или nil при успехе
//
// Пример использования:
//
//	ctx := context.Background()
//	err := conn.PingWithRetry(ctx, 3) // 3 попытки
//	if err != nil {
//	    log.Printf("Database unavailable after 3 attempts: %v", err)
//	}
//
// Возможные ошибки:
//   - "ping failed after N retries": все попытки исчерпаны
//   - context.Canceled: операция отменена между попытками
func (c *Connection) PingWithRetry(ctx context.Context, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := c.Ping(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// Если это не последняя попытка, ждем перед повтором
		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
				// Экспоненциальная задержка: 1s, 2s, 3s, ...
			}
		}
	}

	return fmt.Errorf("ping failed after %d retries: %w", maxRetries, lastErr)
}

// HealthCheckWithRetry выполняет health check с повторными попытками
func (c *Connection) HealthCheckWithRetry(ctx context.Context, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := c.HealthCheck(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// Если это не последняя попытка, ждем перед повтором
		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
				// Экспоненциальная задержка: 1s, 2s, 3s, ...
			}
		}
	}

	return fmt.Errorf("health check failed after %d retries: %w", maxRetries, lastErr)
}
