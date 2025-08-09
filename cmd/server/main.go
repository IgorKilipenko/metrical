package main

import (
	"fmt"
	"log"
	"os"

	"github.com/IgorKilipenko/metrical/internal/app"
	"github.com/spf13/cobra"
)

func main() {
	addr, err := parseFlags()
	if err != nil {
		// Если это help, просто выходим без ошибки
		if err.Error() == "help requested" {
			os.Exit(0)
		}
		log.Fatal(err)
	}

	config, err := app.NewConfig(addr)
	if err != nil {
		log.Fatal(err)
	}
	application := app.New(config)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}

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
		return "", fmt.Errorf("help requested")
	}

	return addr, nil
}
