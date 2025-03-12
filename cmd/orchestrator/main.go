package main

import (
	"log"

	"github.com/AzizovHikmatullo/calc-go_V2/internal/orchestrator"
	"github.com/joho/godotenv"
)

// Starts orchestrator
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	orch := orchestrator.NewOrchestrator()

	orch.Run()
}
