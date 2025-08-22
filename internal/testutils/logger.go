package testutils

import (
	"context"

	"github.com/IgorKilipenko/metrical/internal/logger"
)

// MockLogger - мок логгера для тестов
type MockLogger struct {
	logs []string
}

// SetLevel устанавливает уровень логирования (no-op для мока)
func (m *MockLogger) SetLevel(level logger.LogLevel) {}

// Debug логирует сообщение на уровне DEBUG (no-op для мока)
func (m *MockLogger) Debug(msg string, args ...any) {}

// Info логирует сообщение на уровне INFO (no-op для мока)
func (m *MockLogger) Info(msg string, args ...any) {}

// Warn логирует сообщение на уровне WARN (no-op для мока)
func (m *MockLogger) Warn(msg string, args ...any) {}

// Error логирует сообщение на уровне ERROR (no-op для мока)
func (m *MockLogger) Error(msg string, args ...any) {}

// WithContext создает новый логгер с контекстом (возвращает тот же мок)
func (m *MockLogger) WithContext(ctx context.Context) logger.Logger {
	return m
}

// WithFields создает новый логгер с дополнительными полями (возвращает тот же мок)
func (m *MockLogger) WithFields(fields map[string]any) logger.Logger {
	return m
}

// Sync синхронизирует буферы логгера (no-op для мока)
func (m *MockLogger) Sync() error {
	return nil
}

// NewMockLogger создает новый экземпляр MockLogger
func NewMockLogger() logger.Logger {
	return &MockLogger{}
}
