package main

import (
	"fmt"
	"log"
	"os"

	"openvpnadvanced/cmd/cli"
	"openvpnadvanced/cmd/config"
	"openvpnadvanced/cmd/logger"
)

func main() {
	// Ensure logs directory exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		_ = os.Mkdir("logs", 0755)
	}

	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Load configuration
	if err := config.LoadINIConfig("config.ini"); err != nil {
		log.Fatalf("Failed to load config.ini: %v", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "--start" {
		cli.StartConsole()
	} else {
		fmt.Println(`Usage:
  sudo ./openvpnadvanced --start     Launch interactive console
  sudo ./openvpnadvanced             Show this help message`)
		os.Exit(0)
	}
}
