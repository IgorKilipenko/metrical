package app

import (
	"flag"
	"fmt"
	"strings"
)

// LoadConfig загружает конфигурацию из флагов и переменных окружения
func LoadConfig() (Config, error) {
	// Определяем флаги
	var addr string
	flag.StringVar(&addr, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")

	// Парсим флаги
	flag.Parse()

	// Проверяем на неизвестные флаги
	if flag.NArg() > 0 {
		return Config{}, fmt.Errorf("неизвестные аргументы: %v", flag.Args())
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
