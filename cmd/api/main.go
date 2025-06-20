package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/marcelofabianov/redtogreen/internal/app"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("application startup failed: %v", err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, relying on environment variables.")
	}

	application, err := app.New()
	if err != nil {
		return err
	}

	return application.Run()
}
