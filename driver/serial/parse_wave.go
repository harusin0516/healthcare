package serial

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// WaveformJSON represents the JSON structure for waveform data
type WaveformJSON struct {
	Timestamp     time.Time       `json:"timestamp"`
	SubrecordType int             `json:"subrecord_type"`
	TypeName      string          `json:"type_name"`
	Header        WaveformHeaderJSON `json:"header"`
	Samples       []SampleJSON    `json:"samples"`
	SamplingRate  int             `json:"sampling_rate"`
	Duration      float64         `json:"duration_seconds"`
	TotalSamples  int             `json:"total_samples"`
}

// WaveformHeaderJSON represents the header in JSON format
type WaveformHeaderJSON struct {
	ActLen           int    `json:"act_len"`
	Status           uint16 `json:"status"`
	Label            uint16 `json:"label"`
	HasGap           bool   `json:"has_gap"`
	HasPacerDetected bool   `json:"has_pacer_detected"`
	HasLeadOff       bool   `json:"has_lead_off"`
	StatusBits       []bool `json:"status_bits"`
}

// SampleJSON represents a single sample in JSON format
type SampleJSON struct {
	Index           int     `json:"index"`
	RawValue        int16   `json:"raw_value"`
	PhysicalValue   float64 `json:"physical_value"`
	Unit            string  `json:"unit"`
	IsControlCode   bool    `json:"is_control_code"`
	Timestamp       time.Time `json:"timestamp"`
}

// WaveformParser handles parsing of waveform binary data
type WaveformParser struct {
	subrecordType int
	samplingRate  int
	startTime     time.Time
}

// NewWaveformParser creates a new waveform parser
func NewWaveformParser(subrecordType int) *WaveformParser {
	return &WaveformParser{
		subrecordType: subrecordType,
		samplingRate:  GetSamplingRate(subrecordType),
		startTime:     time.Now(),
	}
}

// ParseWaveformData parses binary waveform data and returns JSON
func (wp *WaveformParser) ParseWaveformData(data []byte) (*WaveformJSON, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("data too short: %d bytes", len(data))
	}

	// Parse header
	header := &WaveformHeader{}
	if err := header.UnmarshalBinary(data[:6]); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Validate data length
	expectedLength := 6 + int(header.ActLen)*2
	if len(data) < expectedLength {
		return nil, fmt.Errorf("data length mismatch: expected %d, got %d", expectedLength, len(data))
	}

	// Parse samples
	samples := make([]int16, header.ActLen)
	for i := 0; i < int(header.ActLen); i++ {
		offset := 6 + i*2
		samples[i] = int16(binary.LittleEndian.Uint16(data[offset : offset+2]))
	}

	// Convert to JSON format
	return wp.convertToJSON(header, samples)
}

// ParseMultipleWaveforms parses multiple waveform records from binary data
func ParseMultipleWaveforms(data []byte, subrecordType int) ([]*WaveformJSON, error) {
	var results []*WaveformJSON
	offset := 0
	parser := NewWaveformParser(subrecordType)

	for offset < len(data) {
		if offset+6 > len(data) {
			break
		}

		// Read header to get length
		header := &WaveformHeader{}
		if err := header.UnmarshalBinary(data[offset : offset+6]); err != nil {
			return nil, fmt.Errorf("failed to parse header at offset %d: %w", offset, err)
		}

		// Calculate total length for this waveform
		waveformLength := 6 + int(header.ActLen)*2
		if offset+waveformLength > len(data) {
			return nil, fmt.Errorf("incomplete waveform data at offset %d", offset)
		}

		// Parse single waveform
		waveformData := data[offset : offset+waveformLength]
		waveform, err := parser.ParseWaveformData(waveformData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse waveform at offset %d: %w", offset, err)
		}

		results = append(results, waveform)
		offset += waveformLength
	}

	return results, nil
}

// convertToJSON converts parsed data to JSON format
func (wp *WaveformParser) convertToJSON(header *WaveformHeader, samples []int16) (*WaveformJSON, error) {
	now := time.Now()
	
	// Create header JSON
	headerJSON := WaveformHeaderJSON{
		ActLen:           int(header.ActLen),
		Status:           header.Status,
		Label:            header.Label,
		HasGap:           header.HasGap(),
		HasPacerDetected: header.HasPacerDetected(),
		HasLeadOff:       header.HasLeadOff(),
		StatusBits:       wp.parseStatusBits(header.Status),
	}

	// Create samples JSON
	samplesJSON := make([]SampleJSON, len(samples))
	sampleInterval := time.Duration(float64(time.Second) / float64(wp.samplingRate))
	
	for i, sample := range samples {
		physicalValue := ConvertSampleToPhysicalValue(sample, wp.subrecordType)
		unit := wp.getUnit(wp.subrecordType)
		
		samplesJSON[i] = SampleJSON{
			Index:         i,
			RawValue:      sample,
			PhysicalValue: physicalValue,
			Unit:          unit,
			IsControlCode: IsControlCode(sample),
			Timestamp:     now.Add(time.Duration(i) * sampleInterval),
		}
	}

	// Calculate duration
	duration := float64(len(samples)) / float64(wp.samplingRate)

	return &WaveformJSON{
		Timestamp:     now,
		SubrecordType: wp.subrecordType,
		TypeName:      wp.getTypeName(wp.subrecordType),
		Header:        headerJSON,
		Samples:       samplesJSON,
		SamplingRate:  wp.samplingRate,
		Duration:      duration,
		TotalSamples:  len(samples),
	}, nil
}

// parseStatusBits converts status field to individual bits
func (wp *WaveformParser) parseStatusBits(status uint16) []bool {
	bits := make([]bool, 16)
	for i := 0; i < 16; i++ {
		bits[i] = (status & (1 << i)) != 0
	}
	return bits
}

// getTypeName returns the human-readable name for the subrecord type
func (wp *WaveformParser) getTypeName(subrecordType int) string {
	switch subrecordType {
	case DRI_WF_CO2:
		return "CO2"
	case DRI_WF_O2:
		return "O2"
	case DRI_WF_N2O:
		return "N2O"
	case DRI_WF_AA:
		return "Anesthesia Agent"
	case DRI_WF_AWP:
		return "Airway Pressure"
	case DRI_WF_FLOW:
		return "Airway Flow"
	case DRI_WF_RESP:
		return "ECG Respiratory"
	case DRI_WF_INVP5:
		return "Invasive Pressure 5"
	case DRI_WF_INVP6:
		return "Invasive Pressure 6"
	case DRI_WF_INVP7:
		return "Invasive Pressure 7"
	case DRI_WF_INVP8:
		return "Invasive Pressure 8"
	case DRI_WF_EEG1:
		return "EEG 1"
	case DRI_WF_EEG2:
		return "EEG 2"
	case DRI_WF_EEG3:
		return "EEG 3"
	case DRI_WF_EEG4:
		return "EEG 4"
	case DRI_WF_ECG12:
		return "12-Lead ECG"
	case DRI_WF_VOL:
		return "Airway Volume"
	case DRI_WF_TONO_PRESS:
		return "Tonometry Pressure"
	case DRI_WF_SPI_LOOP_STATUS:
		return "Spirometry Loop"
	case DRI_WF_ENT_100:
		return "Entropy"
	case DRI_WF_EEG_BIS:
		return "BIS"
	case DRI_WF_PLETH_2:
		return "Plethysmograph 2"
	case DRI_WF_RESP_100:
		return "High Resolution Respiration"
	default:
		return fmt.Sprintf("Unknown Type %d", subrecordType)
	}
}

// getUnit returns the unit for the given subrecord type
func (wp *WaveformParser) getUnit(subrecordType int) string {
	switch subrecordType {
	case DRI_WF_ECG12:
		return "μV"
	case DRI_WF_INVP5, DRI_WF_INVP6, DRI_WF_INVP7, DRI_WF_INVP8:
		return "mmHg"
	case DRI_WF_PLETH, DRI_WF_PLETH_2:
		return "%"
	case DRI_WF_CO2, DRI_WF_O2, DRI_WF_N2O, DRI_WF_AA:
		return "%"
	case DRI_WF_AWP:
		return "cmH2O"
	case DRI_WF_FLOW:
		return "L/min"
	case DRI_WF_VOL:
		return "mL"
	case DRI_WF_TONO_PRESS:
		return "mmHg"
	default:
		return "raw"
	}
}

// ToJSON converts WaveformData to JSON string
func (wd *WaveformData) ToJSON(subrecordType int) (string, error) {
	parser := NewWaveformParser(subrecordType)
	jsonData, err := parser.convertToJSON(&wd.Header, wd.Samples)
	if err != nil {
		return "", err
	}
	
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return string(jsonBytes), nil
}

// ParseAndConvertToJSON is a convenience function that parses binary data and returns JSON string
func ParseAndConvertToJSON(data []byte, subrecordType int) (string, error) {
	parser := NewWaveformParser(subrecordType)
	waveform, err := parser.ParseWaveformData(data)
	if err != nil {
		return "", err
	}
	
	jsonBytes, err := json.MarshalIndent(waveform, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return string(jsonBytes), nil
}

// ParseAndConvertToStruct parses binary data and returns WaveformJSON struct
func ParseAndConvertToStruct(data []byte, subrecordType int) (*WaveformJSON, error) {
	parser := NewWaveformParser(subrecordType)
	return parser.ParseWaveformData(data)
}

// ValidateWaveformData validates the binary waveform data
func ValidateWaveformData(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("data too short: minimum 6 bytes required")
	}
	
	header := &WaveformHeader{}
	if err := header.UnmarshalBinary(data[:6]); err != nil {
		return fmt.Errorf("invalid header: %w", err)
	}
	
	expectedLength := 6 + int(header.ActLen)*2
	if len(data) < expectedLength {
		return fmt.Errorf("data length mismatch: expected %d, got %d", expectedLength, len(data))
	}
	
	if header.ActLen < 0 {
		return fmt.Errorf("invalid act_len: %d", header.ActLen)
	}
	
	return nil
}
