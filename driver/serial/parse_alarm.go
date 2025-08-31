package serial

import (
	"encoding/json"
	"fmt"
	"time"
)

// AlarmJSON represents the overall JSON output for alarm data
type AlarmJSON struct {
	Timestamp     string                 `json:"timestamp"`
	UnixTimestamp uint32                 `json:"unix_timestamp"`
	RecordType    string                 `json:"record_type"`
	RecordNumber  int                    `json:"record_number"`
	DriLevel      int                    `json:"dri_level"`
	DriLevelDesc  string                 `json:"dri_level_description"`
	PlugID        int                    `json:"plug_id"`
	MainType      int                    `json:"main_type"`
	MainTypeName  string                 `json:"main_type_name"`
	Subrecords    []AlarmSubrecordJSON   `json:"subrecords"`
	AlarmData     map[string]interface{} `json:"alarm_data"`
	IsValid       bool                   `json:"is_valid"`
	ParseErrors   []string               `json:"parse_errors,omitempty"`
}

// AlarmSubrecordJSON represents a single alarm subrecord in JSON format
type AlarmSubrecordJSON struct {
	Index        int                    `json:"index"`
	Offset       int16                  `json:"offset"`
	Type         byte                   `json:"type"`
	TypeName     string                 `json:"type_name"`
	IsValid      bool                   `json:"is_valid"`
	IsEndOfList  bool                   `json:"is_end_of_list"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// AlarmParser manages the parsing process for alarm data
type AlarmParser struct {
	errors []string
}

// NewAlarmParser creates a new alarm parser
func NewAlarmParser() *AlarmParser {
	return &AlarmParser{
		errors: make([]string, 0),
	}
}

// ParseAlarmData parses a single binary alarm record into AlarmJSON
func (p *AlarmParser) ParseAlarmData(data []byte) (*AlarmJSON, error) {
	if len(data) < 32 { // Minimum size for DatexHeader
		p.addError("data too short for alarm record")
		return nil, ErrInvalidDataLength
	}

	// Parse the Datex-Ohmeda Record header
	header := &DatexHeader{}
	if err := header.UnmarshalBinary(data[:32]); err != nil {
		p.addError(fmt.Sprintf("failed to parse header: %v", err))
		return nil, err
	}

	// Validate that this is an alarm record
	if header.RMainType != DRI_MT_ALARM {
		p.addError(fmt.Sprintf("expected alarm record type %d, got %d", DRI_MT_ALARM, header.RMainType))
		return nil, fmt.Errorf("invalid record type for alarm data")
	}

	// Create the JSON structure
	alarmJSON := &AlarmJSON{
		Timestamp:     time.Unix(int64(header.RTime), 0).Format(time.RFC3339),
		UnixTimestamp: header.RTime,
		RecordType:    "Alarm Data",
		RecordNumber:  int(header.RNbr),
		DriLevel:      int(header.DriLevel),
		DriLevelDesc:  header.GetDriLevelDescription(),
		PlugID:        int(header.PlugID),
		MainType:      int(header.RMainType),
		MainTypeName:  header.GetMainTypeName(),
		Subrecords:    make([]AlarmSubrecordJSON, 0),
		AlarmData:     make(map[string]interface{}),
		IsValid:       true,
		ParseErrors:   make([]string, 0),
	}

	// Parse subrecords
	if err := p.parseAlarmSubrecords(header, data, alarmJSON); err != nil {
		p.addError(fmt.Sprintf("failed to parse subrecords: %v", err))
		alarmJSON.IsValid = false
	}

	// Parse alarm data
	if err := p.parseAlarmData(header, data, alarmJSON); err != nil {
		p.addError(fmt.Sprintf("failed to parse alarm data: %v", err))
		alarmJSON.IsValid = false
	}

	// Add any parsing errors
	alarmJSON.ParseErrors = p.errors

	return alarmJSON, nil
}

// parseAlarmSubrecords parses the subrecord descriptors
func (p *AlarmParser) parseAlarmSubrecords(header *DatexHeader, data []byte, alarmJSON *AlarmJSON) error {
	for i := 0; i < 8; i++ {
		srDesc := header.SrDesc[i]
		
		subrecordJSON := AlarmSubrecordJSON{
			Index:       i,
			Offset:      srDesc.SrOffset,
			Type:        srDesc.SrType,
			TypeName:    p.getAlarmSubrecordTypeName(srDesc.SrType),
			IsValid:     srDesc.IsValid(),
			IsEndOfList: srDesc.IsEndOfList(),
		}

		// Parse subrecord data if it's valid
		if srDesc.IsValid() && srDesc.SrOffset >= 0 && int(srDesc.SrOffset) < len(data) {
			subrecordData := p.parseSubrecordData(srDesc.SrType, data[srDesc.SrOffset:])
			if subrecordData != nil {
				subrecordJSON.Data = subrecordData
			}
		}

		alarmJSON.Subrecords = append(alarmJSON.Subrecords, subrecordJSON)
	}

	return nil
}

// parseAlarmData parses the actual alarm data
func (p *AlarmParser) parseAlarmData(header *DatexHeader, data []byte, alarmJSON *AlarmJSON) error {
	// Find the first valid alarm subrecord
	var alarmSubrecord *AlarmSubrecords
	for i := 0; i < 8; i++ {
		srDesc := header.SrDesc[i]
		if srDesc.IsValid() && srDesc.SrType == DRI_AL_STATUS {
			// Parse alarm subrecords
			alarmSubrecord = &AlarmSubrecords{}
			if srDesc.SrOffset >= 0 && int(srDesc.SrOffset) < len(data) {
				if err := alarmSubrecord.UnmarshalBinary(data[srDesc.SrOffset:]); err != nil {
					p.addError(fmt.Sprintf("failed to parse alarm subrecords: %v", err))
					return err
				}
			}
			break
		}
	}

	if alarmSubrecord != nil {
		alarmJSON.AlarmData = alarmSubrecord.ToJSON()
	}

	return nil
}

// parseSubrecordData dispatches parsing based on subrecord type
func (p *AlarmParser) parseSubrecordData(subrecordType byte, data []byte) map[string]interface{} {
	switch subrecordType {
	case DRI_AL_STATUS:
		return p.parseAlarmStatusData(data)
	default:
		p.addError(fmt.Sprintf("unknown alarm subrecord type: %d", subrecordType))
		return nil
	}
}

// parseAlarmStatusData parses alarm status data
func (p *AlarmParser) parseAlarmStatusData(data []byte) map[string]interface{} {
	alarmMsg := &AlarmStatusMessage{}
	if err := alarmMsg.UnmarshalBinary(data); err != nil {
		p.addError(fmt.Sprintf("failed to parse alarm status message: %v", err))
		return nil
	}

	return alarmMsg.ToJSON()
}

// getAlarmSubrecordTypeName returns the human-readable name for alarm subrecord type
func (p *AlarmParser) getAlarmSubrecordTypeName(subrecordType byte) string {
	switch subrecordType {
	case DRI_AL_STATUS:
		return "Alarm Status"
	case DRI_EOL_SUBR_LIST:
		return "End of List"
	default:
		return fmt.Sprintf("Unknown Type %d", subrecordType)
	}
}

// addError adds an error to the parser's list
func (p *AlarmParser) addError(err string) {
	p.errors = append(p.errors, err)
}

// ParseMultipleAlarms parses multiple alarm records
func (p *AlarmParser) ParseMultipleAlarms(data []byte) ([]*AlarmJSON, error) {
	var alarms []*AlarmJSON
	offset := 0

	for offset < len(data) {
		if len(data[offset:]) < 32 {
			break
		}

		// Try to parse the header to get the record length
		header := &DatexHeader{}
		if err := header.UnmarshalBinary(data[offset:offset+32]); err != nil {
			p.addError(fmt.Sprintf("failed to parse header at offset %d: %v", offset, err))
			break
		}

		recordLength := int(header.RLen)
		if recordLength <= 0 || offset+recordLength > len(data) {
			p.addError(fmt.Sprintf("invalid record length %d at offset %d", recordLength, offset))
			break
		}

		// Parse the alarm record
		alarmJSON, err := p.ParseAlarmData(data[offset : offset+recordLength])
		if err != nil {
			p.addError(fmt.Sprintf("failed to parse alarm record at offset %d: %v", offset, err))
			offset += recordLength
			continue
		}

		alarms = append(alarms, alarmJSON)
		offset += recordLength
	}

	return alarms, nil
}

// ToJSON converts AlarmJSON to a pretty-printed string
func (p *AlarmParser) ToJSON(alarm *AlarmJSON) (string, error) {
	jsonBytes, err := json.MarshalIndent(alarm, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// Convenience functions for easy use

// ParseAndConvertToJSON parses binary alarm data and converts to JSON string
func ParseAndConvertToJSON(data []byte) (string, error) {
	parser := NewAlarmParser()
	alarm, err := parser.ParseAlarmData(data)
	if err != nil {
		return "", err
	}
	return parser.ToJSON(alarm)
}

// ParseAndConvertToStruct parses binary alarm data and returns the struct
func ParseAndConvertToStruct(data []byte) (*AlarmJSON, error) {
	parser := NewAlarmParser()
	return parser.ParseAlarmData(data)
}

// ParseMultipleAlarmsToJSON parses multiple alarm records and converts to JSON string
func ParseMultipleAlarmsToJSON(data []byte) (string, error) {
	parser := NewAlarmParser()
	alarms, err := parser.ParseMultipleAlarms(data)
	if err != nil {
		return "", err
	}

	// Create a wrapper structure for multiple alarms
	result := map[string]interface{}{
		"alarm_count": len(alarms),
		"alarms":      alarms,
		"parse_errors": parser.errors,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ValidateAlarmData validates basic alarm data structure
func ValidateAlarmData(data []byte) error {
	if len(data) < 32 {
		return fmt.Errorf("data too short for alarm record")
	}

	header := &DatexHeader{}
	if err := header.UnmarshalBinary(data[:32]); err != nil {
		return fmt.Errorf("invalid header: %v", err)
	}

	if header.RMainType != DRI_MT_ALARM {
		return fmt.Errorf("expected alarm record type %d, got %d", DRI_MT_ALARM, header.RMainType)
	}

	if header.RLen <= 0 || int(header.RLen) > len(data) {
		return fmt.Errorf("invalid record length: %d", header.RLen)
	}

	return nil
}

// GetAlarmSummary provides a summary of parsed alarm data
func GetAlarmSummary(alarm *AlarmJSON) map[string]interface{} {
	summary := map[string]interface{}{
		"timestamp":        alarm.Timestamp,
		"record_number":    alarm.RecordNumber,
		"dri_level":        alarm.DriLevel,
		"plug_id":          alarm.PlugID,
		"is_valid":         alarm.IsValid,
		"subrecord_count":  len(alarm.Subrecords),
		"parse_error_count": len(alarm.ParseErrors),
	}

	// Extract alarm information if available
	if alarmData, ok := alarm.AlarmData["data"].(map[string]interface{}); ok {
		if soundOnOff, ok := alarmData["sound_on_off"].(map[string]interface{}); ok {
			if soundStatus, ok := soundOnOff["status"].(bool); ok {
				summary["sound_on"] = soundStatus
			}
		}

		if silenceInfo, ok := alarmData["silence_info"].(map[string]interface{}); ok {
			if isSilenced, ok := silenceInfo["is_silenced"].(bool); ok {
				summary["is_silenced"] = isSilenced
			}
		}

		if activeCount, ok := alarmData["active_alarm_count"].(int); ok {
			summary["active_alarm_count"] = activeCount
		}

		if alarms, ok := alarmData["alarms"].([]map[string]interface{}); ok {
			summary["alarm_count"] = len(alarms)
		}
	}

	return summary
}

// CreateSampleAlarmData creates sample alarm data for testing
func CreateSampleAlarmData() *AlarmJSON {
	now := time.Now()
	
	// Create sample alarm display
	alarmDisplay := &AlarmDisplay{}
	alarmDisplay.SetAlarmText("HR LOW")
	alarmDisplay.TextChanged = true
	alarmDisplay.Color = DRI_PR2 // Yellow
	alarmDisplay.ColorChanged = false

	// Create sample alarm status message
	alarmMsg := &AlarmStatusMessage{
		Reserved:    0,
		SoundOnOff:  true,
		Reserved2:   0,
		Reserved3:   0,
		SilenceInfo: DRI_SI_NONE,
		Reserved4:   [5]int16{0, 0, 0, 0, 0},
	}
	alarmMsg.AlDisp[0] = *alarmDisplay

	// Create sample alarm subrecords
	alarmSubrecords := &AlarmSubrecords{
		AlarmMsg: alarmMsg,
	}

	// Create sample alarm JSON
	alarmJSON := &AlarmJSON{
		Timestamp:     now.Format(time.RFC3339),
		UnixTimestamp: uint32(now.Unix()),
		RecordType:    "Alarm Data",
		RecordNumber:  1,
		DriLevel:      6,
		DriLevelDesc:  "2019 '19",
		PlugID:        12345,
		MainType:      DRI_MT_ALARM,
		MainTypeName:  "Alarm Data",
		Subrecords: []AlarmSubrecordJSON{
			{
				Index:     0,
				Offset:    0,
				Type:      DRI_AL_STATUS,
				TypeName:  "Alarm Status",
				IsValid:   true,
				IsEndOfList: false,
				Data:      alarmSubrecords.ToJSON(),
			},
			{
				Index:     1,
				Offset:    0,
				Type:      DRI_EOL_SUBR_LIST,
				TypeName:  "End of List",
				IsValid:   false,
				IsEndOfList: true,
			},
		},
		AlarmData:   alarmSubrecords.ToJSON(),
		IsValid:     true,
		ParseErrors: []string{},
	}

	return alarmJSON
}
