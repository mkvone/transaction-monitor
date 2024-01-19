package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mkvone/transaction-monitor/pkg"
)

func main() {
	fmt.Println("File hierarchy of the current directory:")
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}

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
