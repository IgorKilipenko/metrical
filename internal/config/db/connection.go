package db

import (
	"context"
	"fmt"
	"time"

	"github.com/IgorKilipenko/metrical/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseConnection интерфейс для работы с базой данных
type DatabaseConnection interface {
	Pool() *pgxpool.Pool
	Ping(ctx context.Context) error
	PingWithRetry(ctx context.Context, maxRetries int) error
	Close()
	Stats() *pgxpool.Stat
	HealthCheck(ctx context.Context) error
	HealthCheckWithRetry(ctx context.Context, maxRetries int) error
}

// Connection представляет подключение к базе данных
type Connection struct {
	pool   *pgxpool.Pool
	config Config
	logger logger.Logger
}

// NewConnection создает новое подключение к базе данных
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

// Ping проверяет соединение с базой данных
func (c *Connection) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.PingTimeout)
	defer cancel()

	if err := c.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	c.logger.Debug("database ping successful")
	return nil
}

// Close закрывает соединение с базой данных
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

// HealthCheck выполняет проверку здоровья базы данных
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

// PingWithRetry выполняет ping с повторными попытками
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
