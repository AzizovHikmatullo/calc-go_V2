package main

import (
	"log"

	"github.com/AzizovHikmatullo/calc-go_V2/internal/agent"
	"github.com/AzizovHikmatullo/calc-go_V2/pkg"
	"github.com/joho/godotenv"
)

// Gets constants from env and start agent
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	cntGoroutines := pkg.GetEnvIntWithDefault("COMPUTING_POWER", 5)
	pingTime := pkg.GetEnvIntWithDefault("PING_MS", 1000)

	newAgent := agent.NewAgent(cntGoroutines, pingTime)

	newAgent.Run()
}
