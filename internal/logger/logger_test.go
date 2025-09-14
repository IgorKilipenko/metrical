package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSlogLogger(t *testing.T) {
	logger := NewSlogLogger()
	require.NotNil(t, logger, "NewSlogLogger() should not return nil")
}

func TestNewSlogLoggerWithConfig(t *testing.T) {
	config := LoggerConfig{
		Level:  DebugLevel,
		Format: "text",
	}

	logger := NewSlogLoggerWithConfig(config)
	require.NotNil(t, logger, "NewSlogLoggerWithConfig() should not return nil")
}

func TestDefaultLoggerConfig(t *testing.T) {
	config := DefaultLoggerConfig()
	assert.Equal(t, InfoLevel, config.Level, "Default level should be InfoLevel")
	assert.Equal(t, "text", config.Format, "Default format should be 'text'")
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
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.expected, result, "LogLevel.String() should return correct string representation")
		})
	}
}

func TestSlogLogger_WithContext(t *testing.T) {
	logger := NewSlogLogger()
	ctx := context.Background()

	newLogger := logger.WithContext(ctx)
	require.NotNil(t, newLogger, "WithContext() should not return nil")

	// Проверяем, что это новый экземпляр
	assert.NotEqual(t, logger, newLogger, "WithContext() should return a new logger instance")
}

func TestSlogLogger_WithFields(t *testing.T) {
	logger := NewSlogLogger()
	fields := map[string]any{
		"key1": "value1",
		"key2": 42,
	}

	newLogger := logger.WithFields(fields)
	require.NotNil(t, newLogger, "WithFields() should not return nil")

	// Проверяем, что это новый экземпляр
	assert.NotEqual(t, logger, newLogger, "WithFields() should return a new logger instance")
}

func TestSlogLogger_SetLevel(t *testing.T) {
	logger := NewSlogLogger().(*ZerologLogger)

	// Устанавливаем уровень
	logger.SetLevel(DebugLevel)
	assert.Equal(t, DebugLevel, logger.level, "Level should be set to DebugLevel")

	// Изменяем уровень
	logger.SetLevel(ErrorLevel)
	assert.Equal(t, ErrorLevel, logger.level, "Level should be set to ErrorLevel")
}

func TestSlogLogger_Sync(t *testing.T) {
	logger := NewSlogLogger()

	// Sync должен возвращать nil
	err := logger.Sync()
	assert.NoError(t, err, "Sync() should not return an error")
}
