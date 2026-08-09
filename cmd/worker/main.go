package main

import (
	"flag"
	"log"

	"github.com/nekoimi/go-project-template/internal/app"
)

func main() {
	configPath := flag.String("config", "config/config.dev.yaml", "path to config file")
	flag.Parse()

	if err := app.RunWorker(*configPath); err != nil {
		log.Fatalf("worker stopped with error: %v", err)
	}
}
