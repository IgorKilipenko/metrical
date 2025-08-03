package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// Создаем контекст с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем агент в горутине
	go func() {
		log.Println("Starting metrics agent...")
		metricsAgent.Run()
	}()

	// Ждем сигнала завершения
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case <-ctx.Done():
		log.Println("Context cancelled")
	}

	// Graceful shutdown
	metricsAgent.Stop()

	// Даем время на завершение горутин
	time.Sleep(1 * time.Second)
	log.Println("Agent shutdown completed")
}
