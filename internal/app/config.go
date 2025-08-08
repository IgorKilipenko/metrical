package app

import "os"

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() Config {
	return Config{
		Port: getEnv("SERVER_PORT", "8080"),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
