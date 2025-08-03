package main

import (
	"log"
	"time"

	"github.com/IgorKilipenko/metrical/internal/agent"
)

func main() {
	// Конфигурация агента
	config := &agent.Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	// Создаем агент
	metricsAgent := agent.NewAgent(config)

	// Запускаем агент
	log.Println("Starting metrics agent...")
	metricsAgent.Run()
}
