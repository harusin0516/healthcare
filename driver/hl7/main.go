package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"driver/hl7"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Configuration file path")
	flag.Parse()

	// Load configuration
	config, err := hl7.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create HL7 server
	server := hl7.NewHL7Server(config)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Print server status
	status := server.GetServerStatus()
	statusJSON, _ := json.MarshalIndent(status, "", "  ")
	fmt.Printf("HL7 Server Status:\n%s\n", statusJSON)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down HL7 server...")

	// Stop server
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	fmt.Println("HL7 server stopped")
}
