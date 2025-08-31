package serial

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"
)

// TrendJSON represents the JSON output for trend data
type TrendJSON struct {
	Timestamp     string                 `json:"timestamp"`
	UnixTimestamp uint32                 `json:"unix_timestamp"`
	RecordType    string                 `json:"record_type"`
	RecordNumber  int                    `json:"record_number"`
	DriLevel      int                    `json:"dri_level"`
	DriLevelDesc  string                 `json:"dri_level_description"`
	PlugID        int                    `json:"plug_id"`
	MainType      int                    `json:"main_type"`
	MainTypeName  string                 `json:"main_type_name"`
	Subrecords    []SubrecordJSON        `json:"subrecords"`
	Groups        map[string]interface{} `json:"groups"`
	IsValid       bool                   `json:"is_valid"`
	ParseErrors   []string               `json:"parse_errors,omitempty"`
}

// SubrecordJSON represents a subrecord in JSON format
type SubrecordJSON struct {
	Index      int    `json:"index"`
	Offset     int16  `json:"offset"`
	Type       byte   `json:"type"`
	TypeName   string `json:"type_name"`
	IsValid    bool   `json:"is_valid"`
	IsEndOfList bool  `json:"is_end_of_list"`
	Data       interface{} `json:"data,omitempty"`
}

// GroupJSON represents a physiological data group in JSON format
type GroupJSON struct {
	Type        string                 `json:"type"`
	Status      uint16                 `json:"status"`
	Label       uint16                 `json:"label"`
	Data        map[string]interface{} `json:"data"`
	IsValid     bool                   `json:"is_valid"`
	ParseErrors []string               `json:"parse_errors,omitempty"`
}

// TrendParser handles parsing of trend data from binary format to JSON
type TrendParser struct {
	errors []string
}

// NewTrendParser creates a new trend parser
func NewTrendParser() *TrendParser {
	return &TrendParser{
		errors: make([]string, 0),
	}
}

// ParseTrendData parses binary trend data and converts it to JSON
func (p *TrendParser) ParseTrendData(data []byte) (*TrendJSON, error) {
	p.errors = make([]string, 0)
	
	if len(data) < 32 {
		return nil, fmt.Errorf("data too short for trend record: %d bytes", len(data))
	}
	
	// Parse the Datex-Ohmeda Record
	record := &DatexRecord{}
	if err := record.UnmarshalBinary(data); err != nil {
		p.addError("Failed to parse Datex-Ohmeda record: " + err.Error())
		return nil, err
	}
	
	// Create JSON structure
	trendJSON := &TrendJSON{
		Timestamp:     time.Unix(int64(record.Header.RTime), 0).Format(time.RFC3339),
		UnixTimestamp: record.Header.RTime,
		RecordType:     "Trend Data",
		RecordNumber:   int(record.Header.RNbr),
		DriLevel:       int(record.Header.DriLevel),
		DriLevelDesc:   record.Header.GetDriLevelDescription(),
		PlugID:         int(record.Header.PlugID),
		MainType:       int(record.Header.RMainType),
		MainTypeName:   record.Header.GetMainTypeName(),
		Subrecords:     make([]SubrecordJSON, 0),
		Groups:         make(map[string]interface{}),
		IsValid:        record.Header.IsValid(),
	}
	
	// Parse subrecords
	p.parseSubrecords(record, trendJSON)
	
	// Parse physiological data if this is a PHDB record
	if record.Header.RMainType == DRI_MT_PHDB {
		p.parsePhysiologicalData(record, trendJSON)
	}
	
	trendJSON.ParseErrors = p.errors
	return trendJSON, nil
}

// parseSubrecords parses subrecord descriptors
func (p *TrendParser) parseSubrecords(record *DatexRecord, trendJSON *TrendJSON) {
	for i := 0; i < 8; i++ {
		srDesc := record.Header.SrDesc[i]
		subrecord := SubrecordJSON{
			Index:       i,
			Offset:      srDesc.SrOffset,
			Type:        srDesc.SrType,
			TypeName:    p.getSubrecordTypeName(srDesc.SrType),
			IsValid:     srDesc.IsValid(),
			IsEndOfList: srDesc.IsEndOfList(),
		}
		
		if srDesc.IsValid() {
			// Try to parse the actual subrecord data
			if int(srDesc.SrOffset) < len(record.Data) {
				subrecordData := record.Data[srDesc.SrOffset:]
				parsedData := p.parseSubrecordData(srDesc.SrType, subrecordData)
				if parsedData != nil {
					subrecord.Data = parsedData
				}
			}
		}
		
		trendJSON.Subrecords = append(trendJSON.Subrecords, subrecord)
	}
}

// parsePhysiologicalData parses physiological database records
func (p *TrendParser) parsePhysiologicalData(record *DatexRecord, trendJSON *TrendJSON) {
	// Parse physiological subrecords
	phSubrecords := &PhysiologicalSubrecords{}
	if err := phSubrecords.UnmarshalBinary(record.Data); err != nil {
		p.addError("Failed to parse physiological subrecords: " + err.Error())
		return
	}
	
	// Add physiological data to groups
	trendJSON.Groups["physiological_data"] = phSubrecords.ToJSON()
	
	// Parse individual physiological database records
	for i, phRecord := range phSubrecords.Records {
		if phRecord != nil {
			groupKey := fmt.Sprintf("ph_record_%d", i)
			trendJSON.Groups[groupKey] = phRecord.ToJSON()
		}
	}
}

// parseSubrecordData parses individual subrecord data based on type
func (p *TrendParser) parseSubrecordData(subrecordType byte, data []byte) interface{} {
	switch subrecordType {
	case DRI_PH_DISPL, DRI_PH_10S_TREND, DRI_PH_60S_TREND:
		return p.parsePhysiologicalDatabaseRecord(data)
	case DRI_PH_AUX_INFO:
		return p.parseAuxiliaryPhysiologicalInfo(data)
	default:
		// For unknown types, return raw data
		return map[string]interface{}{
			"raw_data": data,
			"size":     len(data),
		}
	}
}

// parsePhysiologicalDatabaseRecord parses a physiological database record
func (p *TrendParser) parsePhysiologicalDatabaseRecord(data []byte) interface{} {
	if len(data) < 8 {
		p.addError("Physiological database record too short")
		return nil
	}
	
	phRecord := &PhysiologicalDatabaseRecord{}
	if err := phRecord.UnmarshalBinary(data); err != nil {
		p.addError("Failed to parse physiological database record: " + err.Error())
		return nil
	}
	
	return phRecord.ToJSON()
}

// parseAuxiliaryPhysiologicalInfo parses auxiliary physiological information
func (p *TrendParser) parseAuxiliaryPhysiologicalInfo(data []byte) interface{} {
	if len(data) < 114 {
		p.addError("Auxiliary physiological info too short")
		return nil
	}
	
	auxInfo := &AuxiliaryPhysiologicalInfo{}
	if err := auxInfo.UnmarshalBinary(data); err != nil {
		p.addError("Failed to parse auxiliary physiological info: " + err.Error())
		return nil
	}
	
	return auxInfo.ToJSON()
}

// getSubrecordTypeName returns the human-readable name for subrecord type
func (p *TrendParser) getSubrecordTypeName(subrecordType byte) string {
	switch subrecordType {
	case DRI_PH_DISPL:
		return "Displayed Values"
	case DRI_PH_10S_TREND:
		return "10 Second Trended Values"
	case DRI_PH_60S_TREND:
		return "60 Second Trended Values"
	case DRI_PH_AUX_INFO:
		return "Auxiliary Information"
	case DRI_EOL_SUBR_LIST:
		return "End of List"
	default:
		return fmt.Sprintf("Unknown Type %d", subrecordType)
	}
}

// addError adds an error to the parser's error list
func (p *TrendParser) addError(err string) {
	p.errors = append(p.errors, err)
}

// ParseMultipleTrends parses multiple trend records from binary data
func (p *TrendParser) ParseMultipleTrends(data []byte) ([]*TrendJSON, error) {
	var trends []*TrendJSON
	offset := 0
	
	for offset < len(data) {
		if offset+2 > len(data) {
			break
		}
		
		// Read record length
		recordLen := int(binary.LittleEndian.Uint16(data[offset:offset+2]))
		if recordLen <= 0 || offset+recordLen > len(data) {
			break
		}
		
		// Parse single trend record
		trendData := data[offset:offset+recordLen]
		trend, err := p.ParseTrendData(trendData)
		if err != nil {
			p.addError(fmt.Sprintf("Failed to parse trend at offset %d: %v", offset, err))
		} else {
			trends = append(trends, trend)
		}
		
		offset += recordLen
	}
	
	return trends, nil
}

// ToJSON converts trend data to JSON string
func (p *TrendParser) ToJSON(trend *TrendJSON) (string, error) {
	jsonBytes, err := json.MarshalIndent(trend, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// Convenience functions for easy usage

// ParseAndConvertToJSON parses binary trend data and returns JSON string
func ParseAndConvertToJSON(data []byte) (string, error) {
	parser := NewTrendParser()
	trend, err := parser.ParseTrendData(data)
	if err != nil {
		return "", err
	}
	return parser.ToJSON(trend)
}

// ParseAndConvertToStruct parses binary trend data and returns TrendJSON struct
func ParseAndConvertToStruct(data []byte) (*TrendJSON, error) {
	parser := NewTrendParser()
	return parser.ParseTrendData(data)
}

// ParseMultipleTrendsToJSON parses multiple trend records and returns JSON string
func ParseMultipleTrendsToJSON(data []byte) (string, error) {
	parser := NewTrendParser()
	trends, err := parser.ParseMultipleTrends(data)
	if err != nil {
		return "", err
	}
	
	jsonBytes, err := json.MarshalIndent(trends, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ValidateTrendData validates trend data structure
func ValidateTrendData(data []byte) error {
	if len(data) < 32 {
		return fmt.Errorf("data too short for trend record: %d bytes", len(data))
	}
	
	// Check record length
	recordLen := int(binary.LittleEndian.Uint16(data[0:2]))
	if recordLen <= 0 || recordLen > len(data) {
		return fmt.Errorf("invalid record length: %d", recordLen)
	}
	
	// Check DRI level
	driLevel := data[3]
	if driLevel < DRI_LEVEL_95 || driLevel > DRI_LEVEL_06 {
		return fmt.Errorf("invalid DRI level: %d", driLevel)
	}
	
	// Check main type
	mainType := int16(binary.LittleEndian.Uint16(data[28:30]))
	if mainType < 0 {
		return fmt.Errorf("invalid main type: %d", mainType)
	}
	
	return nil
}

// GetTrendSummary returns a summary of trend data
func GetTrendSummary(trend *TrendJSON) map[string]interface{} {
	summary := map[string]interface{}{
		"timestamp":        trend.Timestamp,
		"record_type":      trend.RecordType,
		"main_type":        trend.MainTypeName,
		"dri_level":        trend.DriLevelDesc,
		"subrecord_count":  len(trend.Subrecords),
		"group_count":      len(trend.Groups),
		"is_valid":         trend.IsValid,
	}
	
	// Count valid subrecords
	validSubrecords := 0
	for _, sr := range trend.Subrecords {
		if sr.IsValid {
			validSubrecords++
		}
	}
	summary["valid_subrecord_count"] = validSubrecords
	
	// Add parse errors if any
	if len(trend.ParseErrors) > 0 {
		summary["parse_error_count"] = len(trend.ParseErrors)
		summary["parse_errors"] = trend.ParseErrors
	}
	
	return summary
}
