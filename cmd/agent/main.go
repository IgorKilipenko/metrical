package main

import (
	"log"
	"os"
)

// Version приложения (можно установить при сборке через ldflags)
var Version = "dev"

func main() {
	log.Printf("Starting metrics agent v%s", Version)

	if err := Execute(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}

	log.Println("Agent shutdown completed successfully")
}
