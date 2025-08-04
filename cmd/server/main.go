package main

import (
	"log"

	"github.com/IgorKilipenko/metrical/internal/server"
)

func main() {
	// Создаем и запускаем сервер
	srv := server.NewServer(":8080")
	log.Fatal(srv.Start())
}
