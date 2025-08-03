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
	// Конфигурация агента с значениями по умолчанию
	config := agent.NewConfig()

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
