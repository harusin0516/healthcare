package hl7

import (
	"fmt"
	"log"
	"os"
)

// HL7Driver represents the main HL7 communication driver
type HL7Driver struct {
	server *HL7Server
	config *ServerConfig
	logger *log.Logger
}

// NewHL7Driver creates a new HL7 driver
func NewHL7Driver(configFile string) (*HL7Driver, error) {
	// Load configuration
	config, err := LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create server
	server := NewHL7Server(config)

	// Create logger
	logger := log.New(os.Stdout, "[HL7-DRIVER] ", log.LstdFlags)

	return &HL7Driver{
		server: server,
		config: config,
		logger: logger,
	}, nil
}

// Start starts the HL7 driver
func (d *HL7Driver) Start() error {
	d.logger.Printf("Starting HL7 Driver on %s:%d", d.config.Host, d.config.Port)
	
	// Start the server
	if err := d.server.Start(); err != nil {
		return fmt.Errorf("failed to start HL7 server: %v", err)
	}

	d.logger.Println("HL7 Driver started successfully")
	return nil
}

// Stop stops the HL7 driver
func (d *HL7Driver) Stop() error {
	d.logger.Println("Stopping HL7 Driver...")
	
	// Stop the server
	if err := d.server.Stop(); err != nil {
		return fmt.Errorf("failed to stop HL7 server: %v", err)
	}

	d.logger.Println("HL7 Driver stopped successfully")
	return nil
}

// GetStatus returns the current status of the HL7 driver
func (d *HL7Driver) GetStatus() map[string]interface{} {
	return d.server.GetServerStatus()
}

// GetConnectedClients returns the list of connected clients
func (d *HL7Driver) GetConnectedClients() []*Client {
	return d.server.GetConnectedClients()
}

// DisconnectClient disconnects a specific client
func (d *HL7Driver) DisconnectClient(clientID string) error {
	return d.server.DisconnectClient(clientID)
}
