package app

import (
	"fmt"
	"strings"
)

// NewConfig создает конфигурацию из строки адреса и дополнительных параметров
func NewConfig(addr string, storeInterval int, fileStoragePath string, restore bool, databaseDSN string) (Config, error) {
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
		DatabaseDSN:     databaseDSN,
	}, nil
}

// GetStorageType определяет тип хранилища на основе конфигурации
// Приоритет: PostgreSQL -> Файл -> Память
func (c Config) GetStorageType() string {
	// 1. PostgreSQL - если указан DATABASE_DSN
	if c.DatabaseDSN != "" {
		return "postgresql"
	}

	// 2. Файл - если указан путь к файлу
	if c.FileStoragePath != "" {
		return "file"
	}

	// 3. Память - по умолчанию
	return "memory"
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
