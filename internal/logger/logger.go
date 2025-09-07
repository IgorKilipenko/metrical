package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
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

// ZerologLogger реализация логгера на основе zerolog
type ZerologLogger struct {
	logger zerolog.Logger
	level  LogLevel
}

// NewSlogLogger создает новый логгер на основе zerolog
func NewSlogLogger() Logger {
	return &ZerologLogger{
		logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
		level:  InfoLevel,
	}
}

// NewSlogLoggerWithConfig создает логгер с конфигурацией на основе zerolog
func NewSlogLoggerWithConfig(config LoggerConfig) Logger {
	var logger zerolog.Logger

	switch config.Format {
	case "json":
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	default:
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	// Устанавливаем уровень логирования
	switch config.Level {
	case DebugLevel:
		logger = logger.Level(zerolog.DebugLevel)
	case InfoLevel:
		logger = logger.Level(zerolog.InfoLevel)
	case WarnLevel:
		logger = logger.Level(zerolog.WarnLevel)
	case ErrorLevel:
		logger = logger.Level(zerolog.ErrorLevel)
	}

	return &ZerologLogger{
		logger: logger,
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
func (l *ZerologLogger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug логирует сообщение на уровне DEBUG
func (l *ZerologLogger) Debug(msg string, args ...any) {
	if l.level <= DebugLevel {
		event := l.logger.Debug()
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				event = event.Interface(args[i].(string), args[i+1])
			}
		}
		event.Msg(msg)
	}
}

// Info логирует сообщение на уровне INFO
func (l *ZerologLogger) Info(msg string, args ...any) {
	if l.level <= InfoLevel {
		event := l.logger.Info()
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				event = event.Interface(args[i].(string), args[i+1])
			}
		}
		event.Msg(msg)
	}
}

// Warn логирует сообщение на уровне WARN
func (l *ZerologLogger) Warn(msg string, args ...any) {
	if l.level <= WarnLevel {
		event := l.logger.Warn()
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				event = event.Interface(args[i].(string), args[i+1])
			}
		}
		event.Msg(msg)
	}
}

// Error логирует сообщение на уровне ERROR
func (l *ZerologLogger) Error(msg string, args ...any) {
	if l.level <= ErrorLevel {
		event := l.logger.Error()
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				event = event.Interface(args[i].(string), args[i+1])
			}
		}
		event.Msg(msg)
	}
}

// WithContext создает новый логгер с контекстом
func (l *ZerologLogger) WithContext(ctx context.Context) Logger {
	return &ZerologLogger{
		logger: l.logger.With().Interface("context", ctx).Logger(),
		level:  l.level,
	}
}

// WithFields создает новый логгер с дополнительными полями
func (l *ZerologLogger) WithFields(fields map[string]any) Logger {
	logger := l.logger.With()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}

	return &ZerologLogger{
		logger: logger.Logger(),
		level:  l.level,
	}
}

// Sync синхронизирует буферы логгера
func (l *ZerologLogger) Sync() error {
	// zerolog автоматически синхронизирует буферы
	return nil
}
