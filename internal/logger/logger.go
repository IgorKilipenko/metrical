package logger

import (
	"context"
	"log/slog"
	"os"
)

// LogLevel представляет уровень логирования
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String возвращает строковое представление уровня логирования
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger интерфейс для логирования
type Logger interface {
	// Уровни логирования
	SetLevel(level LogLevel)
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	// Контекст и поля
	WithContext(ctx context.Context) Logger
	WithFields(fields map[string]any) Logger

	// Утилиты
	Sync() error
}

// SlogLogger реализация логгера на основе slog
type SlogLogger struct {
	logger *slog.Logger
	level  LogLevel
}

// NewSlogLogger создает новый логгер на основе slog
func NewSlogLogger() Logger {
	return &SlogLogger{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		level:  InfoLevel,
	}
}

// NewSlogLoggerWithConfig создает логгер с конфигурацией
func NewSlogLoggerWithConfig(config LoggerConfig) Logger {
	var handler slog.Handler

	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, nil)
	default:
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	return &SlogLogger{
		logger: slog.New(handler),
		level:  config.Level,
	}
}

// LoggerConfig конфигурация логгера
type LoggerConfig struct {
	Level  LogLevel
	Format string // "text" или "json"
}

// DefaultLoggerConfig возвращает конфигурацию по умолчанию
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:  InfoLevel,
		Format: "text",
	}
}

// SetLevel устанавливает уровень логирования
func (l *SlogLogger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug логирует сообщение на уровне DEBUG
func (l *SlogLogger) Debug(msg string, args ...any) {
	if l.level <= DebugLevel {
		l.logger.Debug(msg, args...)
	}
}

// Info логирует сообщение на уровне INFO
func (l *SlogLogger) Info(msg string, args ...any) {
	if l.level <= InfoLevel {
		l.logger.Info(msg, args...)
	}
}

// Warn логирует сообщение на уровне WARN
func (l *SlogLogger) Warn(msg string, args ...any) {
	if l.level <= WarnLevel {
		l.logger.Warn(msg, args...)
	}
}

// Error логирует сообщение на уровне ERROR
func (l *SlogLogger) Error(msg string, args ...any) {
	if l.level <= ErrorLevel {
		l.logger.Error(msg, args...)
	}
}

// WithContext создает новый логгер с контекстом
func (l *SlogLogger) WithContext(ctx context.Context) Logger {
	return &SlogLogger{
		logger: l.logger.With("context", ctx),
		level:  l.level,
	}
}

// WithFields создает новый логгер с дополнительными полями
func (l *SlogLogger) WithFields(fields map[string]any) Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &SlogLogger{
		logger: l.logger.With(args...),
		level:  l.level,
	}
}

// Sync синхронизирует буферы логгера
func (l *SlogLogger) Sync() error {
	// slog автоматически синхронизирует буферы
	return nil
}
