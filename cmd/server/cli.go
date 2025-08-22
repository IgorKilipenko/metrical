package main

import (
	"fmt"
	"strings"

	"github.com/IgorKilipenko/metrical/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// parseFlags парсит флаги командной строки
func parseFlags() (string, error) {
	var addr string

	// Настраиваем Viper для работы с переменными окружения
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Привязываем переменную окружения к ключу конфигурации
	viper.BindEnv("address", "ADDRESS")

	// Устанавливаем значение по умолчанию
	viper.SetDefault("address", "localhost:8080")

	// Получаем значение по умолчанию с учетом переменной окружения
	defaultAddr := viper.GetString("address")

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

			// Загружаем конфигурацию с помощью Viper
			serverConfig, err := config.LoadServerConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Валидируем адрес
			if err := validateAddress(serverConfig.Address); err != nil {
				return err
			}

			return nil
		},
	}

	// Добавляем флаг для адреса
	cmd.Flags().StringVarP(&addr, "address", "a", defaultAddr, "адрес эндпоинта HTTP-сервера")

	// Привязываем флаг к Viper
	viper.BindPFlag("address", cmd.Flags().Lookup("address"))

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

	// Возвращаем финальный адрес из конфигурации
	return viper.GetString("address"), nil
}
