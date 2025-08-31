package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"driver/hl7"
)

func main() {
	// Parse command line flags
	serverHost := flag.String("host", "localhost", "HL7 server host")
	serverPort := flag.Int("port", 8080, "HL7 server port")
	messageType := flag.String("message", "ADT_Admission", "Message type to send")
	flag.Parse()

	// Create sample messages
	samples := hl7.NewSampleHL7Messages()
	
	// Get the specified message
	message, exists := samples.GetAllSampleMessages()[*messageType]
	if !exists {
		log.Fatalf("Unknown message type: %s", *messageType)
	}

	// Connect to HL7 server
	address := fmt.Sprintf("%s:%d", *serverHost, *serverPort)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Failed to connect to HL7 server: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to HL7 server at %s\n", address)
	fmt.Printf("Sending message: %s\n", samples.GetMessageDescription(*messageType))

	// Send the HL7 message
	_, err = conn.Write([]byte(message))
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	fmt.Println("Message sent successfully")

	// Wait for acknowledgment
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString(0x0D) // Read until CR
	if err != nil {
		log.Printf("Failed to read acknowledgment: %v", err)
		return
	}

	fmt.Printf("Received acknowledgment: %s\n", response)

	// Parse the acknowledgment
	parser := hl7.NewHL7Parser()
	ackMessage, err := parser.ParseMessage(response)
	if err != nil {
		log.Printf("Failed to parse acknowledgment: %v", err)
		return
	}

	// Convert acknowledgment to JSON
	ackJSON, err := ackMessage.ToJSON()
	if err != nil {
		log.Printf("Failed to convert acknowledgment to JSON: %v", err)
		return
	}

	fmt.Printf("Acknowledgment JSON:\n%s\n", ackJSON)
}

// TestClient represents a test client for HL7 server
type TestClient struct {
	host string
	port int
	conn net.Conn
}

// NewTestClient creates a new test client
func NewTestClient(host string, port int) *TestClient {
	return &TestClient{
		host: host,
		port: port,
	}
}

// Connect connects to the HL7 server
func (c *TestClient) Connect() error {
	address := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to HL7 server: %v", err)
	}
	c.conn = conn
	return nil
}

// Disconnect disconnects from the HL7 server
func (c *TestClient) Disconnect() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendMessage sends an HL7 message and waits for acknowledgment
func (c *TestClient) SendMessage(message string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected to server")
	}

	// Send the message
	_, err := c.conn.Write([]byte(message))
	if err != nil {
		return "", fmt.Errorf("failed to send message: %v", err)
	}

	// Wait for acknowledgment
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	reader := bufio.NewReader(c.conn)
	response, err := reader.ReadString(0x0D) // Read until CR
	if err != nil {
		return "", fmt.Errorf("failed to read acknowledgment: %v", err)
	}

	return response, nil
}

// SendAllSampleMessages sends all sample messages
func (c *TestClient) SendAllSampleMessages() error {
	samples := hl7.NewSampleHL7Messages()
	messages := samples.GetAllSampleMessages()

	fmt.Println("Sending all sample messages...")

	for msgType, message := range messages {
		fmt.Printf("\nSending: %s\n", samples.GetMessageDescription(msgType))
		
		response, err := c.SendMessage(message)
		if err != nil {
			fmt.Printf("Error sending %s: %v\n", msgType, err)
			continue
		}

		fmt.Printf("Received acknowledgment: %s\n", response)

		// Parse and display acknowledgment
		parser := hl7.NewHL7Parser()
		ackMessage, err := parser.ParseMessage(response)
		if err != nil {
			fmt.Printf("Failed to parse acknowledgment: %v\n", err)
			continue
		}

		ackJSON, err := ackMessage.ToJSON()
		if err != nil {
			fmt.Printf("Failed to convert acknowledgment to JSON: %v\n", err)
			continue
		}

		fmt.Printf("Acknowledgment JSON:\n%s\n", ackJSON)

		// Wait a bit between messages
		time.Sleep(1 * time.Second)
	}

	return nil
}

// RunPerformanceTest runs a performance test
func (c *TestClient) RunPerformanceTest(messageCount int) error {
	samples := hl7.NewSampleHL7Messages()
	message := samples.GetADTMessage()

	fmt.Printf("Running performance test with %d messages...\n", messageCount)

	startTime := time.Now()

	for i := 0; i < messageCount; i++ {
		_, err := c.SendMessage(message)
		if err != nil {
			return fmt.Errorf("failed to send message %d: %v", i+1, err)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("Sent %d messages...\n", i+1)
		}
	}

	duration := time.Since(startTime)
	rate := float64(messageCount) / duration.Seconds()

	fmt.Printf("Performance test completed:\n")
	fmt.Printf("  Total messages: %d\n", messageCount)
	fmt.Printf("  Total time: %v\n", duration)
	fmt.Printf("  Rate: %.2f messages/second\n", rate)

	return nil
}
