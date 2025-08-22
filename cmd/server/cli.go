package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// parseFlags парсит флаги командной строки
func parseFlags() (string, error) {
	var addr string

	// Получаем значение по умолчанию с учетом переменной окружения
	defaultAddr := getEnvOrDefault("ADDRESS", "localhost:8080")

	cmd := &cobra.Command{
		Use:   "server",
		Short: "HTTP сервер для сбора метрик",
		Long: `HTTP сервер для приема метрик от агентов по протоколу HTTP.

Environment variables:
  ADDRESS: адрес эндпоинта HTTP-сервера`,
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Проверяем на неизвестные аргументы
			if len(args) > 0 {
				return fmt.Errorf("неизвестные аргументы: %v", args)
			}

			// Валидируем адрес
			if err := validateAddress(addr); err != nil {
				return err
			}

			return nil
		},
	}

	// Добавляем флаг для адреса
	cmd.Flags().StringVarP(&addr, "address", "a", defaultAddr, "адрес эндпоинта HTTP-сервера")

	// Парсим аргументы
	if err := cmd.Execute(); err != nil {
		return "", err
	}

	// Проверяем, не был ли запрошен help
	if helpFlag := cmd.Flags().Lookup("help"); helpFlag != nil && helpFlag.Changed {
		return "", HelpRequestedError{}
	}

	// Проверяем, не был ли запрошен version
	if versionFlag := cmd.Flags().Lookup("version"); versionFlag != nil && versionFlag.Changed {
		return "", VersionRequestedError{}
	}

	// Получаем финальное значение с учетом приоритета
	finalAddr := getFinalValue("ADDRESS", addr, "localhost:8080")

	// Валидируем финальный адрес
	if err := validateAddress(finalAddr); err != nil {
		return "", err
	}

	return finalAddr, nil
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getFinalValue возвращает финальное значение с учетом приоритета
func getFinalValue(envKey, flagValue, defaultValue string) string {
	// 1. Переменная окружения (высший приоритет)
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	// 2. Флаг командной строки (средний приоритет)
	if flagValue != "" && flagValue != defaultValue {
		return flagValue
	}
	// 3. Значение по умолчанию (низший приоритет)
	return defaultValue
}
