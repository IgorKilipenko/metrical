package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// getEnvOrDefault получает значение из переменной окружения или возвращает значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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

			// Определяем финальный адрес с учетом приоритета:
			// 1. Переменная окружения ADDRESS
			// 2. Флаг командной строки
			// 3. Значение по умолчанию
			finalAddr := getEnvOrDefault("ADDRESS", addr)

			// Валидируем адрес
			if err := validateAddress(finalAddr); err != nil {
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

	// Возвращаем финальный адрес с учетом приоритета
	return getEnvOrDefault("ADDRESS", addr), nil
}
