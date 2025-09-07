package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// ServerConfig содержит конфигурацию сервера
type ServerConfig struct {
	Address         string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

// parseFlags парсит флаги командной строки
func parseFlags() (ServerConfig, error) {
	var config ServerConfig

	cmd := &cobra.Command{
		Use:   "server",
		Short: "HTTP сервер для сбора метрик",
		Long: `HTTP сервер для приема метрик от агентов по протоколу HTTP.

Environment variables:
  ADDRESS: адрес эндпоинта HTTP-сервера
  STORE_INTERVAL: интервал сохранения метрик в секундах (по умолчанию 300)
  FILE_STORAGE_PATH: путь к файлу для сохранения метрик
  RESTORE: загружать ли метрики при старте (true/false)`,
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Проверяем на неизвестные аргументы
			if len(args) > 0 {
				return fmt.Errorf("неизвестные аргументы: %v", args)
			}

			return nil
		},
	}

	// Добавляем флаги
	cmd.Flags().StringVarP(&config.Address, "address", "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	cmd.Flags().IntVarP(&config.StoreInterval, "interval", "i", 300, "интервал сохранения метрик в секундах (0 для синхронного сохранения)")
	cmd.Flags().StringVarP(&config.FileStoragePath, "file", "f", "/tmp/metrics-db.json", "путь к файлу для сохранения метрик")
	cmd.Flags().BoolVarP(&config.Restore, "restore", "r", true, "загружать ли метрики при старте")

	// Парсим аргументы
	if err := cmd.Execute(); err != nil {
		return ServerConfig{}, err
	}

	// Проверяем, не был ли запрошен help
	if helpFlag := cmd.Flags().Lookup("help"); helpFlag != nil && helpFlag.Changed {
		return ServerConfig{}, HelpRequestedError{}
	}

	// Проверяем, не был ли запрошен version
	if versionFlag := cmd.Flags().Lookup("version"); versionFlag != nil && versionFlag.Changed {
		return ServerConfig{}, VersionRequestedError{}
	}

	// Получаем финальные значения с учетом приоритета
	config.Address = getFinalValue("ADDRESS", config.Address, "localhost:8080")
	config.StoreInterval = getFinalIntValue("STORE_INTERVAL", config.StoreInterval, 300)
	config.FileStoragePath = getFinalValue("FILE_STORAGE_PATH", config.FileStoragePath, "/tmp/metrics-db.json")
	config.Restore = getFinalBoolValue("RESTORE", config.Restore, true)

	// Валидируем финальный адрес
	if err := validateAddress(config.Address); err != nil {
		return ServerConfig{}, err
	}

	return config, nil
}

// getFinalValue возвращает финальное значение с учетом приоритета
func getFinalValue(envKey, flagValue, defaultValue string) string {
	// 1. Переменная окружения (высший приоритет)
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	// 2. Флаг командной строки (средний приоритет)
	// Проверяем, был ли флаг установлен (даже если значение пустое)
	if flagValue != defaultValue {
		return flagValue
	}
	// 3. Значение по умолчанию (низший приоритет)
	return defaultValue
}

// getFinalIntValue возвращает финальное целочисленное значение с учетом приоритета
func getFinalIntValue(envKey string, flagValue, defaultValue int) int {
	// 1. Переменная окружения (высший приоритет)
	if envValue := os.Getenv(envKey); envValue != "" {
		if intValue, err := strconv.Atoi(envValue); err == nil {
			return intValue
		}
	}
	// 2. Флаг командной строки (средний приоритет)
	return flagValue
}

// getFinalBoolValue возвращает финальное булево значение с учетом приоритета
func getFinalBoolValue(envKey string, flagValue, defaultValue bool) bool {
	// 1. Переменная окружения (высший приоритет)
	if envValue := os.Getenv(envKey); envValue != "" {
		if boolValue, err := strconv.ParseBool(envValue); err == nil {
			return boolValue
		}
	}
	// 2. Флаг командной строки (средний приоритет)
	return flagValue
}
