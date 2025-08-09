package main

import (
	"log"

	"github.com/IgorKilipenko/metrical/internal/httpserver"
)

func main() {
	// Создаем и запускаем сервер
	srv := httpserver.NewServer(":8080")
	log.Fatal(srv.Start())
}
