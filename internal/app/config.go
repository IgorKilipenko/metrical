package app

import (
	"fmt"
	"strings"
)

// NewConfig создает конфигурацию из строки адреса и дополнительных параметров
func NewConfig(addr string, storeInterval int, fileStoragePath string, restore bool) (Config, error) {
	// Парсим адрес и порт
	serverAddr, serverPort, err := parseAddr(addr)
	if err != nil {
		return Config{}, fmt.Errorf("некорректный адрес сервера: %w", err)
	}

	return Config{
		Addr:            serverAddr,
		Port:            serverPort,
		StoreInterval:   storeInterval,
		FileStoragePath: fileStoragePath,
		Restore:         restore,
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
