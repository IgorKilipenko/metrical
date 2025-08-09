package app

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// LoadConfig загружает конфигурацию из флагов командной строки
func LoadConfig() (Config, error) {
	var addr string

	// Создаем команду с Cobra
	cmd := &cobra.Command{
		Use:   "server",
		Short: "HTTP сервер для сбора метрик",
		Long:  `HTTP сервер для приема метрик от агентов по протоколу HTTP.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Проверяем на неизвестные аргументы
			if len(args) > 0 {
				return fmt.Errorf("неизвестные аргументы: %v", args)
			}
			return nil
		},
	}

	// Добавляем флаг для адреса
	cmd.Flags().StringVarP(&addr, "address", "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")

	// Парсим аргументы
	if err := cmd.Execute(); err != nil {
		return Config{}, err
	}

	// Проверяем, не был ли запрошен help
	if cmd.Flags().Lookup("help").Changed {
		return Config{}, fmt.Errorf("help requested")
	}

	// Парсим адрес и порт
	serverAddr, serverPort, err := parseAddr(addr)
	if err != nil {
		return Config{}, fmt.Errorf("некорректный адрес сервера: %w", err)
	}

	return Config{
		Addr: serverAddr,
		Port: serverPort,
	}, nil
}

// parseAddr парсит строку адреса в адрес и порт
func parseAddr(addr string) (string, string, error) {
	// Если адрес содержит двоеточие, разделяем на адрес и порт
	if strings.Contains(addr, ":") {
		parts := strings.SplitN(addr, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("некорректный формат адреса: %s", addr)
		}
		return parts[0], parts[1], nil
	}

	// Если адрес не содержит двоеточие, считаем что это только порт
	return "localhost", addr, nil
}
