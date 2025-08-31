package hl7

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// HL7Server represents the HL7 server
type HL7Server struct {
	config     *ServerConfig
	parser     *HL7Parser
	listener   net.Listener
	clients    map[string]*Client
	mutex      sync.RWMutex
	messageChan chan *HL7Message
	stopChan   chan bool
	logger     *log.Logger
}

// Client represents a connected client
type Client struct {
	ID       string
	Conn     net.Conn
	Address  string
	LastSeen time.Time
}

// NewHL7Server creates a new HL7 server
func NewHL7Server(config *ServerConfig) *HL7Server {
	return &HL7Server{
		config:     config,
		parser:     NewHL7Parser(),
		clients:    make(map[string]*Client),
		messageChan: make(chan *HL7Message, 100),
		stopChan:   make(chan bool),
		logger:     log.New(os.Stdout, "[HL7-SERVER] ", log.LstdFlags),
	}
}

// LoadConfig loads server configuration from file
func LoadConfig(filename string) (*ServerConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config struct {
		Server ServerConfig `json:"server"`
	}

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	return &config.Server, nil
}

// Start starts the HL7 server
func (s *HL7Server) Start() error {
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start server on %s: %v", address, err)
	}
	
	s.listener = listener
	s.logger.Printf("HL7 server started on %s", address)
	
	// Start message processor
	go s.processMessages()
	
	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return nil
			default:
				s.logger.Printf("Failed to accept connection: %v", err)
				continue
			}
		}
		
		// Check if client is allowed
		if !s.isClientAllowed(conn.RemoteAddr().String()) {
			s.logger.Printf("Connection rejected from %s", conn.RemoteAddr().String())
			conn.Close()
			continue
		}
		
		// Handle client connection
		go s.handleClient(conn)
	}
}

// Stop stops the HL7 server
func (s *HL7Server) Stop() error {
	s.logger.Println("Stopping HL7 server...")
	
	// Signal stop
	close(s.stopChan)
	
	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}
	
	// Close all client connections
	s.mutex.Lock()
	for _, client := range s.clients {
		client.Conn.Close()
	}
	s.clients = make(map[string]*Client)
	s.mutex.Unlock()
	
	s.logger.Println("HL7 server stopped")
	return nil
}

// handleClient handles a single client connection
func (s *HL7Server) handleClient(conn net.Conn) {
	clientID := conn.RemoteAddr().String()
	
	client := &Client{
		ID:       clientID,
		Conn:     conn,
		Address:  conn.RemoteAddr().String(),
		LastSeen: time.Now(),
	}
	
	// Add client to list
	s.mutex.Lock()
	s.clients[clientID] = client
	s.mutex.Unlock()
	
	s.logger.Printf("Client connected: %s", clientID)
	
	// Set connection timeout
	conn.SetDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))
	
	// Handle client messages
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		if message == "" {
			continue
		}
		
		// Update client last seen time
		client.LastSeen = time.Now()
		conn.SetDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))
		
		// Parse HL7 message
		hl7Message, err := s.parser.ParseMessage(message)
		if err != nil {
			s.logger.Printf("Failed to parse HL7 message from %s: %v", clientID, err)
			continue
		}
		
		// Send acknowledgment
		ack := s.createAcknowledgment(hl7Message)
		if err := s.sendAcknowledgment(conn, ack); err != nil {
			s.logger.Printf("Failed to send acknowledgment to %s: %v", clientID, err)
		}
		
		// Process message
		s.messageChan <- hl7Message
		
		s.logger.Printf("Received HL7 message from %s: %s", clientID, hl7Message.Type)
	}
	
	// Remove client from list
	s.mutex.Lock()
	delete(s.clients, clientID)
	s.mutex.Unlock()
	
	conn.Close()
	s.logger.Printf("Client disconnected: %s", clientID)
}

// processMessages processes received HL7 messages
func (s *HL7Server) processMessages() {
	for {
		select {
		case message := <-s.messageChan:
			s.handleMessage(message)
		case <-s.stopChan:
			return
		}
	}
}

// handleMessage handles a single HL7 message
func (s *HL7Server) handleMessage(message *HL7Message) {
	// Log message details
	s.logger.Printf("Processing HL7 message: Type=%s, ID=%s", message.Type, message.ID)
	
	// Convert to JSON
	jsonStr, err := message.ToJSON()
	if err != nil {
		s.logger.Printf("Failed to convert message to JSON: %v", err)
		return
	}
	
	// Log JSON output
	s.logger.Printf("HL7 Message JSON:\n%s", jsonStr)
	
	// Handle different message types
	switch message.Type {
	case HL7_MSG_ADT:
		s.handleADTMessage(message)
	case HL7_MSG_ORU:
		s.handleORUMessage(message)
	case HL7_MSG_ORM:
		s.handleORMMessage(message)
	default:
		s.logger.Printf("Unknown message type: %s", message.Type)
	}
}

// handleADTMessage handles ADT (Admission, Discharge, Transfer) messages
func (s *HL7Server) handleADTMessage(message *HL7Message) {
	patientID := message.GetPatientID()
	patientName := message.GetPatientName()
	patientDOB := message.GetPatientDOB()
	patientSex := message.GetPatientSex()
	
	s.logger.Printf("ADT Message - Patient: ID=%s, Name=%s, DOB=%s, Sex=%s", 
		patientID, patientName, patientDOB, patientSex)
	
	// Extract additional information
	admissionDate := message.GetAdmissionDate()
	dischargeDate := message.GetDischargeDate()
	
	if admissionDate != "" {
		s.logger.Printf("Admission Date: %s", admissionDate)
	}
	if dischargeDate != "" {
		s.logger.Printf("Discharge Date: %s", dischargeDate)
	}
	
	// Get diagnoses
	diagnoses := message.GetDiagnoses()
	for i, diagnosis := range diagnoses {
		if len(diagnosis.Fields) > 2 {
			s.logger.Printf("Diagnosis %d: %s", i+1, diagnosis.Fields[2].Value)
		}
	}
	
	// Get allergies
	allergies := message.GetAllergies()
	for i, allergy := range allergies {
		if len(allergy.Fields) > 2 {
			s.logger.Printf("Allergy %d: %s", i+1, allergy.Fields[2].Value)
		}
	}
}

// handleORUMessage handles ORU (Observation Result) messages
func (s *HL7Server) handleORUMessage(message *HL7Message) {
	patientID := message.GetPatientID()
	patientName := message.GetPatientName()
	
	s.logger.Printf("ORU Message - Patient: ID=%s, Name=%s", patientID, patientName)
	
	// Get observation results
	observations := message.GetObservationResults()
	for i, observation := range observations {
		if len(observation.Fields) >= 5 {
			value := observation.Fields[4].Value
			units := ""
			if len(observation.Fields) >= 6 {
				units = observation.Fields[5].Value
			}
			s.logger.Printf("Observation %d: %s %s", i+1, value, units)
		}
	}
}

// handleORMMessage handles ORM (Order Message) messages
func (s *HL7Server) handleORMMessage(message *HL7Message) {
	patientID := message.GetPatientID()
	patientName := message.GetPatientName()
	
	s.logger.Printf("ORM Message - Patient: ID=%s, Name=%s", patientID, patientName)
	
	// Get order information from ORC segments
	orders := message.GetSegmentsByType(HL7_SEG_ORC)
	for i, order := range orders {
		if len(order.Fields) >= 2 {
			orderID := order.Fields[1].Value
			s.logger.Printf("Order %d: %s", i+1, orderID)
		}
	}
}

// createAcknowledgment creates an HL7 acknowledgment message
func (s *HL7Server) createAcknowledgment(message *HL7Message) string {
	// Create MSH segment for acknowledgment
	msh := fmt.Sprintf("MSH|^~\\&|HL7SERVER|HOSPITAL|%s|%s|%s||ACK^A01|%s|P|2.5",
		message.GetFieldValue(HL7_SEG_MSH, 2), // Sending application
		message.GetFieldValue(HL7_SEG_MSH, 3), // Sending facility
		time.Now().Format("20060102150405"),    // Message date/time
		message.ID)                             // Message control ID
	
	// Create MSA segment
	msa := fmt.Sprintf("MSA|AA|%s", message.ID) // AA = Application Accept
	
	// Create ERR segment (empty for successful acknowledgment)
	err := "ERR|"
	
	// Combine segments
	ack := fmt.Sprintf("%s\r%s\r%s\r", msh, msa, err)
	
	return ack
}

// sendAcknowledgment sends an acknowledgment to the client
func (s *HL7Server) sendAcknowledgment(conn net.Conn, ack string) error {
	// Add MLLP wrapper
	mllpAck := fmt.Sprintf("%c%s%c%c", 0x0B, ack, 0x1C, 0x0D)
	
	_, err := conn.Write([]byte(mllpAck))
	return err
}

// isClientAllowed checks if the client IP is allowed
func (s *HL7Server) isClientAllowed(clientIP string) bool {
	if len(s.config.AllowedIPs) == 0 {
		return true // Allow all if no restrictions
	}
	
	// Extract IP address from client address
	ip := strings.Split(clientIP, ":")[0]
	
	for _, allowedIP := range s.config.AllowedIPs {
		if ip == allowedIP {
			return true
		}
	}
	
	return false
}

// GetConnectedClients returns the list of connected clients
func (s *HL7Server) GetConnectedClients() []*Client {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	clients := make([]*Client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	
	return clients
}

// GetClientCount returns the number of connected clients
func (s *HL7Server) GetClientCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return len(s.clients)
}

// DisconnectClient disconnects a specific client
func (s *HL7Server) DisconnectClient(clientID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	client, exists := s.clients[clientID]
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}
	
	client.Conn.Close()
	delete(s.clients, clientID)
	
	s.logger.Printf("Client %s disconnected by server", clientID)
	return nil
}

// GetServerStatus returns the server status information
func (s *HL7Server) GetServerStatus() map[string]interface{} {
	return map[string]interface{}{
		"host":           s.config.Host,
		"port":           s.config.Port,
		"timeout":        s.config.Timeout,
		"max_connections": s.config.MaxConnections,
		"connected_clients": s.GetClientCount(),
		"is_running":     s.listener != nil,
	}
}
