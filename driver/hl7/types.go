package hl7

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// HL7 Message Types
const (
	HL7_MSG_ADT = "ADT" // Admission, Discharge, Transfer
	HL7_MSG_ORU = "ORU" // Observation Result
	HL7_MSG_ORM = "ORM" // Order Message
	HL7_MSG_ACK = "ACK" // Acknowledgment
	HL7_MSG_NACK = "NAK" // Negative Acknowledgment
)

// HL7 Segment Types
const (
	HL7_SEG_MSH = "MSH" // Message Header
	HL7_SEG_PID = "PID" // Patient Identification
	HL7_SEG_PV1 = "PV1" // Patient Visit
	HL7_SEG_OBR = "OBR" // Observation Request
	HL7_SEG_OBX = "OBX" // Observation Result
	HL7_SEG_ORC = "ORC" // Order
	HL7_SEG_AL1 = "AL1" // Allergy Information
	HL7_SEG_DG1 = "DG1" // Diagnosis
	HL7_SEG_PRX = "PRX" // Patient Result
)

// HL7 Message Structure
type HL7Message struct {
	Segments []HL7Segment `json:"segments"`
	Raw      string       `json:"raw_message"`
	Version  string       `json:"version"`
	Type     string       `json:"message_type"`
	ID       string       `json:"message_id"`
	Time     time.Time    `json:"timestamp"`
}

// HL7 Segment Structure
type HL7Segment struct {
	Type     string        `json:"segment_type"`
	Fields   []HL7Field    `json:"fields"`
	Raw      string        `json:"raw_segment"`
}

// HL7 Field Structure
type HL7Field struct {
	Value      string        `json:"value"`
	Components []HL7Component `json:"components,omitempty"`
	Repetitions []HL7Field   `json:"repetitions,omitempty"`
}

// HL7 Component Structure
type HL7Component struct {
	Value         string           `json:"value"`
	Subcomponents []HL7Subcomponent `json:"subcomponents,omitempty"`
}

// HL7 Subcomponent Structure
type HL7Subcomponent struct {
	Value string `json:"value"`
}

// HL7 Parser Configuration
type HL7Config struct {
	Version                string `json:"version"`
	Encoding               string `json:"encoding"`
	FieldSeparator         string `json:"field_separator"`
	ComponentSeparator     string `json:"component_separator"`
	SubcomponentSeparator  string `json:"subcomponent_separator"`
	RepetitionSeparator    string `json:"repetition_separator"`
	EscapeCharacter        string `json:"escape_character"`
}

// HL7 Server Configuration
type ServerConfig struct {
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	Timeout        int      `json:"timeout"`
	MaxConnections int      `json:"max_connections"`
	AllowedIPs     []string `json:"allowed_ips"`
}

// HL7 Parser
type HL7Parser struct {
	config HL7Config
}

// NewHL7Parser creates a new HL7 parser with default configuration
func NewHL7Parser() *HL7Parser {
	return &HL7Parser{
		config: HL7Config{
			Version:               "2.5",
			Encoding:              "UTF-8",
			FieldSeparator:        "|",
			ComponentSeparator:    "^",
			SubcomponentSeparator: "&",
			RepetitionSeparator:   "~",
			EscapeCharacter:       "\\",
		},
	}
}

// NewHL7ParserWithConfig creates a new HL7 parser with custom configuration
func NewHL7ParserWithConfig(config HL7Config) *HL7Parser {
	return &HL7Parser{
		config: config,
	}
}

// ParseMessage parses a raw HL7 message string into HL7Message structure
func (p *HL7Parser) ParseMessage(rawMessage string) (*HL7Message, error) {
	// Remove MLLP wrapper if present
	message := p.removeMLLPWrapper(rawMessage)
	
	// Split message into segments
	segments := strings.Split(message, "\r")
	
	hl7Message := &HL7Message{
		Segments: make([]HL7Segment, 0, len(segments)),
		Raw:      rawMessage,
		Time:     time.Now(),
	}
	
	for _, segmentRaw := range segments {
		segmentRaw = strings.TrimSpace(segmentRaw)
		if segmentRaw == "" {
			continue
		}
		
		segment, err := p.parseSegment(segmentRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse segment: %v", err)
		}
		
		hl7Message.Segments = append(hl7Message.Segments, *segment)
		
		// Extract message header information from MSH segment
		if segment.Type == HL7_SEG_MSH && len(segment.Fields) >= 9 {
			hl7Message.Version = segment.Fields[8].Value
			if len(segment.Fields) >= 10 {
				hl7Message.Type = segment.Fields[8].Value
			}
			if len(segment.Fields) >= 10 {
				hl7Message.ID = segment.Fields[9].Value
			}
		}
	}
	
	return hl7Message, nil
}

// parseSegment parses a single HL7 segment
func (p *HL7Parser) parseSegment(segmentRaw string) (*HL7Segment, error) {
	fields := strings.Split(segmentRaw, p.config.FieldSeparator)
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty segment")
	}
	
	segment := &HL7Segment{
		Type:   fields[0],
		Fields: make([]HL7Field, 0, len(fields)),
		Raw:    segmentRaw,
	}
	
	for i, fieldRaw := range fields {
		if i == 0 {
			// Skip segment type field
			continue
		}
		
		field, err := p.parseField(fieldRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse field %d: %v", i, err)
		}
		
		segment.Fields = append(segment.Fields, *field)
	}
	
	return segment, nil
}

// parseField parses a single HL7 field
func (p *HL7Parser) parseField(fieldRaw string) (*HL7Field, error) {
	// Check for repetitions
	if strings.Contains(fieldRaw, p.config.RepetitionSeparator) {
		repetitions := strings.Split(fieldRaw, p.config.RepetitionSeparator)
		field := &HL7Field{
			Value:      repetitions[0],
			Repetitions: make([]HL7Field, 0, len(repetitions)),
		}
		
		for _, repetition := range repetitions {
			repField, err := p.parseField(repetition)
			if err != nil {
				return nil, err
			}
			field.Repetitions = append(field.Repetitions, *repField)
		}
		
		return field, nil
	}
	
	// Check for components
	if strings.Contains(fieldRaw, p.config.ComponentSeparator) {
		components := strings.Split(fieldRaw, p.config.ComponentSeparator)
		field := &HL7Field{
			Value:      components[0],
			Components: make([]HL7Component, 0, len(components)),
		}
		
		for _, componentRaw := range components {
			component, err := p.parseComponent(componentRaw)
			if err != nil {
				return nil, err
			}
			field.Components = append(field.Components, *component)
		}
		
		return field, nil
	}
	
	// Simple field
	return &HL7Field{
		Value: fieldRaw,
	}, nil
}

// parseComponent parses a single HL7 component
func (p *HL7Parser) parseComponent(componentRaw string) (*HL7Component, error) {
	// Check for subcomponents
	if strings.Contains(componentRaw, p.config.SubcomponentSeparator) {
		subcomponents := strings.Split(componentRaw, p.config.SubcomponentSeparator)
		component := &HL7Component{
			Value:         subcomponents[0],
			Subcomponents: make([]HL7Subcomponent, 0, len(subcomponents)),
		}
		
		for _, subcomponentRaw := range subcomponents {
			component.Subcomponents = append(component.Subcomponents, HL7Subcomponent{
				Value: subcomponentRaw,
			})
		}
		
		return component, nil
	}
	
	// Simple component
	return &HL7Component{
		Value: componentRaw,
	}, nil
}

// removeMLLPWrapper removes MLLP (Minimal Lower Layer Protocol) wrapper
func (p *HL7Parser) removeMLLPWrapper(message string) string {
	// MLLP wrapper: 0x0B (VT) + message + 0x1C (FS) + 0x0D (CR)
	if len(message) >= 3 {
		if message[0] == 0x0B && message[len(message)-2] == 0x1C && message[len(message)-1] == 0x0D {
			return message[1 : len(message)-2]
		}
	}
	return message
}

// ToJSON converts HL7Message to JSON format
func (m *HL7Message) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// GetSegmentByType returns the first segment of the specified type
func (m *HL7Message) GetSegmentByType(segmentType string) *HL7Segment {
	for _, segment := range m.Segments {
		if segment.Type == segmentType {
			return &segment
		}
	}
	return nil
}

// GetSegmentsByType returns all segments of the specified type
func (m *HL7Message) GetSegmentsByType(segmentType string) []*HL7Segment {
	var segments []*HL7Segment
	for i := range m.Segments {
		if m.Segments[i].Type == segmentType {
			segments = append(segments, &m.Segments[i])
		}
	}
	return segments
}

// GetFieldValue returns the value of a specific field in a segment
func (m *HL7Message) GetFieldValue(segmentType string, fieldIndex int) string {
	segment := m.GetSegmentByType(segmentType)
	if segment == nil || fieldIndex >= len(segment.Fields) {
		return ""
	}
	return segment.Fields[fieldIndex].Value
}

// GetComponentValue returns the value of a specific component in a field
func (m *HL7Message) GetComponentValue(segmentType string, fieldIndex, componentIndex int) string {
	segment := m.GetSegmentByType(segmentType)
	if segment == nil || fieldIndex >= len(segment.Fields) {
		return ""
	}
	
	field := segment.Fields[fieldIndex]
	if componentIndex >= len(field.Components) {
		return ""
	}
	
	return field.Components[componentIndex].Value
}

// IsADTMessage returns true if this is an ADT message
func (m *HL7Message) IsADTMessage() bool {
	return m.Type == HL7_MSG_ADT
}

// IsORUMessage returns true if this is an ORU message
func (m *HL7Message) IsORUMessage() bool {
	return m.Type == HL7_MSG_ORU
}

// IsORMMessage returns true if this is an ORM message
func (m *HL7Message) IsORMMessage() bool {
	return m.Type == HL7_MSG_ORM
}

// GetPatientID returns the patient ID from PID segment
func (m *HL7Message) GetPatientID() string {
	return m.GetFieldValue(HL7_SEG_PID, 2)
}

// GetPatientName returns the patient name from PID segment
func (m *HL7Message) GetPatientName() string {
	return m.GetFieldValue(HL7_SEG_PID, 4)
}

// GetPatientDOB returns the patient date of birth from PID segment
func (m *HL7Message) GetPatientDOB() string {
	return m.GetFieldValue(HL7_SEG_PID, 6)
}

// GetPatientSex returns the patient sex from PID segment
func (m *HL7Message) GetPatientSex() string {
	return m.GetFieldValue(HL7_SEG_PID, 7)
}

// GetAdmissionDate returns the admission date from PV1 segment
func (m *HL7Message) GetAdmissionDate() string {
	return m.GetFieldValue(HL7_SEG_PV1, 44)
}

// GetDischargeDate returns the discharge date from PV1 segment
func (m *HL7Message) GetDischargeDate() string {
	return m.GetFieldValue(HL7_SEG_PV1, 45)
}

// GetObservationResults returns all OBX segments
func (m *HL7Message) GetObservationResults() []*HL7Segment {
	return m.GetSegmentsByType(HL7_SEG_OBX)
}

// GetDiagnoses returns all DG1 segments
func (m *HL7Message) GetDiagnoses() []*HL7Segment {
	return m.GetSegmentsByType(HL7_SEG_DG1)
}

// GetAllergies returns all AL1 segments
func (m *HL7Message) GetAllergies() []*HL7Segment {
	return m.GetSegmentsByType(HL7_SEG_AL1)
}
