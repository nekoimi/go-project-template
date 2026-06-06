package main

// @title           Go Template API
// @version         1.0
// @description     A Go backend template project.
// @host            localhost:8080
// @BasePath        /v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization

import (
	"flag"
	"log"

	"github.com/nekoimi/go-project-template/internal/app"
)

func main() {
	configPath := flag.String("config", "config/config.dev.yaml", "path to config file")
	flag.Parse()

	if err := app.Run(*configPath); err != nil {
		log.Fatalf("app stopped with error: %v", err)
	}
}
