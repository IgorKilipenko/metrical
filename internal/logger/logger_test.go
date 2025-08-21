package logger

import (
	"context"
	"testing"
)

func TestNewSlogLogger(t *testing.T) {
	logger := NewSlogLogger()
	if logger == nil {
		t.Fatal("NewSlogLogger() returned nil")
	}
}

func TestNewSlogLoggerWithConfig(t *testing.T) {
	config := LoggerConfig{
		Level:  DebugLevel,
		Format: "text",
	}

	logger := NewSlogLoggerWithConfig(config)
	if logger == nil {
		t.Fatal("NewSlogLoggerWithConfig() returned nil")
	}
}

func TestDefaultLoggerConfig(t *testing.T) {
	config := DefaultLoggerConfig()
	if config.Level != InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}
	if config.Format != "text" {
		t.Errorf("Expected 'text', got %s", config.Format)
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("LogLevel(%d).String() = %s, want %s", tt.level, result, tt.expected)
		}
	}
}

func TestSlogLogger_WithContext(t *testing.T) {
	logger := NewSlogLogger()
	ctx := context.Background()

	newLogger := logger.WithContext(ctx)
	if newLogger == nil {
		t.Fatal("WithContext() returned nil")
	}

	// Проверяем, что это новый экземпляр
	if newLogger == logger {
		t.Error("WithContext() should return a new logger instance")
	}
}

func TestSlogLogger_WithFields(t *testing.T) {
	logger := NewSlogLogger()
	fields := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	newLogger := logger.WithFields(fields)
	if newLogger == nil {
		t.Fatal("WithFields() returned nil")
	}

	// Проверяем, что это новый экземпляр
	if newLogger == logger {
		t.Error("WithFields() should return a new logger instance")
	}
}

func TestSlogLogger_SetLevel(t *testing.T) {
	logger := NewSlogLogger().(*SlogLogger)

	// Устанавливаем уровень
	logger.SetLevel(DebugLevel)
	if logger.level != DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", logger.level)
	}

	// Изменяем уровень
	logger.SetLevel(ErrorLevel)
	if logger.level != ErrorLevel {
		t.Errorf("Expected ErrorLevel, got %v", logger.level)
	}
}

func TestSlogLogger_Sync(t *testing.T) {
	logger := NewSlogLogger()

	// Sync должен возвращать nil
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}
}
