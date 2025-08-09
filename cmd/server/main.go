package main

import (
	"log"
	"os"

	"github.com/IgorKilipenko/metrical/internal/app"
)

func main() {
	addr, err := parseFlags()
	if err != nil {
		// Если это help, просто выходим без ошибки
		if err.Error() == "help requested" {
			os.Exit(0)
		}
		log.Fatal(err)
	}

	config, err := app.NewConfig(addr)
	if err != nil {
		log.Fatal(err)
	}

	application := app.New(config)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
