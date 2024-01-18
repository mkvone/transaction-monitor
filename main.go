package main

import (
	"flag"
	"log"

	"github.com/mkvone/transaction-monitor/pkg"
)

func main() {
	// Define a flag for the configuration path
	var configPath string
	flag.StringVar(&configPath, "config-path", "./config.yml", "Path to configuration file")
	flag.Parse() // Parse the flags

	// Load the configuration using the path provided in the command line argument
	cfg, err := pkg.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Run the application with the loaded configuration
	pkg.Run(cfg)
}
