package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// parseFlags парсит флаги командной строки
func parseFlags() (string, error) {
	var addr string

	cmd := &cobra.Command{
		Use:   "server",
		Short: "HTTP сервер для сбора метрик",
		Long:  `HTTP сервер для приема метрик от агентов по протоколу HTTP.`,
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
	cmd.Flags().StringVarP(&addr, "address", "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")

	// Парсим аргументы
	if err := cmd.Execute(); err != nil {
		return "", err
	}

	// Проверяем, не был ли запрошен help
	if cmd.Flags().Lookup("help").Changed {
		return "", HelpRequestedError{}
	}

	return addr, nil
}
