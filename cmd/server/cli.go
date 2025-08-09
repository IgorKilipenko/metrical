package main

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
)

// HelpRequestedError представляет ошибку запроса справки
type HelpRequestedError struct{}

func (e HelpRequestedError) Error() string {
	return "help requested"
}

// IsHelpRequested проверяет, является ли ошибка запросом справки
func IsHelpRequested(err error) bool {
	_, ok := err.(HelpRequestedError)
	return ok
}

// InvalidAddressError представляет ошибку некорректного адреса
type InvalidAddressError struct {
	Address string
	Reason  string
}

func (e InvalidAddressError) Error() string {
	return fmt.Sprintf("некорректный адрес '%s': %s", e.Address, e.Reason)
}

// IsInvalidAddress проверяет, является ли ошибка ошибкой некорректного адреса
func IsInvalidAddress(err error) bool {
	_, ok := err.(InvalidAddressError)
	return ok
}

// validateAddress проверяет корректность адреса
func validateAddress(addr string) error {
	if addr == "" {
		return InvalidAddressError{Address: addr, Reason: "адрес не может быть пустым"}
	}

	// Пытаемся разобрать адрес как host:port
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// Если не удалось разобрать как host:port, пробуем как только порт
		if _, err := net.LookupPort("tcp", addr); err != nil {
			return InvalidAddressError{Address: addr, Reason: "некорректный формат адреса"}
		}
		// Если это только порт, добавляем localhost
		return nil
	}

	// Проверяем, что хост не пустой (кроме случая :port)
	if host == "" && port == "" {
		return InvalidAddressError{Address: addr, Reason: "адрес должен содержать хост или порт"}
	}

	// Проверяем, что порт не пустой
	if port == "" {
		return InvalidAddressError{Address: addr, Reason: "порт не может быть пустым"}
	}

	// Проверяем, что порт является числом
	if _, err := net.LookupPort("tcp", port); err != nil {
		return InvalidAddressError{Address: addr, Reason: "некорректный порт"}
	}

	return nil
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
