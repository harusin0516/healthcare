package serial

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// DRI Physiological Data Subrecord Types
const (
	DRI_PH_ECG           = 1  // ECG data
	DRI_PH_INVP          = 2  // Invasive blood pressure data
	DRI_PH_PLETH         = 3  // Plethysmograph data
	DRI_PH_CO2           = 4  // CO2 data
	DRI_PH_O2            = 5  // O2 data
	DRI_PH_N2O           = 6  // N2O data
	DRI_PH_AA            = 7  // Anesthesia agent data
	DRI_PH_AWP           = 8  // Airway pressure data
	DRI_PH_FLOW          = 9  // Airway flow data
	DRI_PH_RESP          = 10 // Respiratory data
	DRI_PH_TEMP          = 11 // Temperature data
	DRI_PH_EEG           = 12 // EEG data
	DRI_PH_BIS           = 13 // BIS data
	DRI_PH_ENT           = 14 // Entropy data
	DRI_PH_SPI           = 15 // Spirometry data
	DRI_PH_DISPL         = 16 // Displayed Values
	DRI_PH_10S_TREND     = 17 // 10 Second Trended Values
	DRI_PH_60S_TREND     = 18 // 60 Second Trended Values
	DRI_PH_AUX_INFO      = 19 // Auxiliary Information
)

// DRI Message Types (Record main types)
// Table 2-3 Record main types
const (
	DRI_MT_PHDB    = 0 // Physiological data and related transmission requests
	DRI_MT_WAVE    = 1 // Waveform data and related transmission requests
	DRI_MT_ALARM   = 4 // Alarm data and related transmission requests (network interface only)
	DRI_MT_NETWORK = 5 // Patient identification data and demographics (network interface only)
	DRI_MT_FO      = 8 // Anesthesia record keeping event data (network interface only)
)

// DRI Interface Levels
// 2.3. Supported DRI levels
// Table 2-5 Supported DRI levels
const (
	DRI_LEVEL_95 = 2  // 1995 '95
	DRI_LEVEL_97 = 3  // 1997 '97
	DRI_LEVEL_98 = 4  // 1998 '98
	DRI_LEVEL_99 = 5  // 1999 '99
	DRI_LEVEL_00 = 6  // 2001 '01
	DRI_LEVEL_01 = 7  // 2002 '02
	DRI_LEVEL_02 = 8  // 2003 '03
	DRI_LEVEL_03 = 9  // 2005 '05
	DRI_LEVEL_04 = 10 // 2009 '09
	DRI_LEVEL_05 = 11 // 2015 '15
	DRI_LEVEL_06 = 12 // 2019 '19
)

// Legacy DRI Interface Levels (for backward compatibility)
const (
	DRI_LEVEL_1  = 1  // Basic level
	DRI_LEVEL_2  = 2  // Enhanced level
	DRI_LEVEL_3  = 3  // Standard level
	DRI_LEVEL_5  = 5  // Advanced level
	DRI_LEVEL_8  = 8  // Network level (patient demographics supported)
	DRI_LEVEL_9  = 9  // Extended level
	DRI_LEVEL_11 = 11 // High resolution level
)

// Subrecord End of List indicator
const (
	DRI_EOL_SUBR_LIST = 0xFF // Subrecord type to indicate end of subrecord list
)

// Subrecord Descriptor structure (Updated based on PDF)
// Table 2-4 Subrecord field contents
// C struct equivalent:
// struct sr_desc {
//     short sr_offset; // Relative pointer to the subrecord
//     byte sr_type;    // Contains the subrecord type
// };
type SrDesc struct {
	SrOffset int16  // Relative pointer to the subrecord (offset from start of data area)
	SrType   byte   // Contains the subrecord type (0xFF indicates end of subrecord list)
}

// Size returns the size of SrDesc in bytes
func (s *SrDesc) Size() int {
	return 3 // 2 + 1 bytes
}

// MarshalBinary converts the subrecord descriptor to binary format
func (s *SrDesc) MarshalBinary() ([]byte, error) {
	buf := make([]byte, s.Size())
	
	// sr_offset: Relative pointer to the subrecord
	binary.LittleEndian.PutUint16(buf[0:2], uint16(s.SrOffset))
	
	// sr_type: Contains the subrecord type
	buf[2] = s.SrType
	
	return buf, nil
}

// UnmarshalBinary converts binary data to subrecord descriptor
func (s *SrDesc) UnmarshalBinary(data []byte) error {
	if len(data) < s.Size() {
		return ErrInvalidDataLength
	}
	
	// sr_offset: Relative pointer to the subrecord
	s.SrOffset = int16(binary.LittleEndian.Uint16(data[0:2]))
	
	// sr_type: Contains the subrecord type
	s.SrType = data[2]
	
	return nil
}

// IsEndOfList returns true if this subrecord indicates the end of the subrecord list
func (s *SrDesc) IsEndOfList() bool {
	return s.SrType == DRI_EOL_SUBR_LIST
}

// IsValid returns true if this subrecord is valid (not end of list)
func (s *SrDesc) IsValid() bool {
	return s.SrType != DRI_EOL_SUBR_LIST
}

// Datex-Ohmeda Record Header structure
// Table 2-2 Datex-Ohmeda Record header field contents
// C struct equivalent:
// struct datex_hdr {
//     short r_len;        // Total length of the record, including the header
//     byte r_nbr;         // Record number
//     byte dri_level;     // DRI level the monitor supports
//     word plug_id;       // Plug identifier number of the sending monitor
//     dword r_time;       // Time when the record was transmitted (seconds since 1.1.1970)
//     byte n_subnet;      // Reserved field (must be zeroed)
//     byte reserved2;     // Reserved field (must be zeroed)
//     word reserved3;     // Reserved field (must be zeroed)
//     short r_maintype;   // Main type of the record
//     struct sr_desc sr_desc[8]; // Array describing data in subrecords
// };
type DatexHeader struct {
	RLen      int16      // Total length of the record, including the header
	RNbr      byte       // Record number
	DriLevel  byte       // DRI level the monitor supports (see 2.3. Supported DRI levels)
	PlugID    uint16     // Plug identifier number of the sending monitor
	RTime     uint32     // Time when the record was transmitted (seconds since 1.1.1970)
	NSubnet   byte       // Reserved field (must be zeroed)
	Reserved2 byte       // Reserved field (must be zeroed)
	Reserved3 uint16     // Reserved field (must be zeroed)
	RMainType int16      // Main type of the record (subrecord types are subtypes of this)
	SrDesc    [8]SrDesc  // Array describing the data in the subrecords
}

// Size returns the size of DatexHeader in bytes
func (h *DatexHeader) Size() int {
	return 2 + 1 + 1 + 2 + 4 + 1 + 1 + 2 + 2 + 8*3 // 32 bytes total (updated for new SrDesc size)
}

// MarshalBinary converts the header to binary format
func (h *DatexHeader) MarshalBinary() ([]byte, error) {
	buf := make([]byte, h.Size())
	offset := 0
	
	// r_len: Total length of the record, including the header
	binary.LittleEndian.PutUint16(buf[offset:], uint16(h.RLen))
	offset += 2
	
	// r_nbr: Record number
	buf[offset] = h.RNbr
	offset += 1
	
	// dri_level: DRI level the monitor supports
	buf[offset] = h.DriLevel
	offset += 1
	
	// plug_id: Plug identifier number of the sending monitor
	binary.LittleEndian.PutUint16(buf[offset:], h.PlugID)
	offset += 2
	
	// r_time: Time when the record was transmitted (seconds since 1.1.1970)
	binary.LittleEndian.PutUint32(buf[offset:], h.RTime)
	offset += 4
	
	// n_subnet: Reserved field (must be zeroed)
	buf[offset] = h.NSubnet
	offset += 1
	
	// reserved2: Reserved field (must be zeroed)
	buf[offset] = h.Reserved2
	offset += 1
	
	// reserved3: Reserved field (must be zeroed)
	binary.LittleEndian.PutUint16(buf[offset:], h.Reserved3)
	offset += 2
	
	// r_maintype: Main type of the record
	binary.LittleEndian.PutUint16(buf[offset:], uint16(h.RMainType))
	offset += 2
	
	// sr_desc: Array describing the data in the subrecords
	for i := 0; i < 8; i++ {
		srDescBytes, err := h.SrDesc[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(buf[offset:], srDescBytes)
		offset += h.SrDesc[i].Size()
	}
	
	return buf, nil
}

// UnmarshalBinary converts binary data to header
func (h *DatexHeader) UnmarshalBinary(data []byte) error {
	if len(data) < h.Size() {
		return ErrInvalidDataLength
	}
	
	offset := 0
	
	// r_len: Total length of the record, including the header
	h.RLen = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	// r_nbr: Record number
	h.RNbr = data[offset]
	offset += 1
	
	// dri_level: DRI level the monitor supports
	h.DriLevel = data[offset]
	offset += 1
	
	// plug_id: Plug identifier number of the sending monitor
	h.PlugID = binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	
	// r_time: Time when the record was transmitted (seconds since 1.1.1970)
	h.RTime = binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	
	// n_subnet: Reserved field
	h.NSubnet = data[offset]
	offset += 1
	
	// reserved2: Reserved field
	h.Reserved2 = data[offset]
	offset += 1
	
	// reserved3: Reserved field
	h.Reserved3 = binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	
	// r_maintype: Main type of the record
	h.RMainType = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	// sr_desc: Array describing the data in the subrecords
	for i := 0; i < 8; i++ {
		if err := h.SrDesc[i].UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		offset += h.SrDesc[i].Size()
	}
	
	return nil
}

// GetActiveSubrecordCount returns the number of active subrecords
func (h *DatexHeader) GetActiveSubrecordCount() int {
	count := 0
	for i := 0; i < 8; i++ {
		if h.SrDesc[i].IsValid() {
			count++
		}
	}
	return count
}

// GetSubrecordType returns the subrecord type at the specified index
func (h *DatexHeader) GetSubrecordType(index int) byte {
	if index >= 0 && index < 8 {
		return h.SrDesc[index].SrType
	}
	return 0
}

// GetSubrecordOffset returns the subrecord offset at the specified index
func (h *DatexHeader) GetSubrecordOffset(index int) int16 {
	if index >= 0 && index < 8 {
		return h.SrDesc[index].SrOffset
	}
	return 0
}

// SetSubrecord sets the subrecord descriptor at the specified index
func (h *DatexHeader) SetSubrecord(index int, srOffset int16, srType byte) error {
	if index < 0 || index >= 8 {
		return fmt.Errorf("invalid subrecord index: %d", index)
	}
	h.SrDesc[index].SrOffset = srOffset
	h.SrDesc[index].SrType = srType
	return nil
}

// ClearSubrecords clears all subrecord descriptors
func (h *DatexHeader) ClearSubrecords() {
	for i := 0; i < 8; i++ {
		h.SrDesc[i].SrOffset = 0
		h.SrDesc[i].SrType = DRI_EOL_SUBR_LIST
	}
}

// GetMainTypeName returns the human-readable name for the main record type
func (h *DatexHeader) GetMainTypeName() string {
	switch h.RMainType {
	case DRI_MT_PHDB:
		return "Physiological Data"
	case DRI_MT_WAVE:
		return "Waveform Data"
	case DRI_MT_ALARM:
		return "Alarm Data"
	case DRI_MT_NETWORK:
		return "Network Data"
	case DRI_MT_FO:
		return "Foreground Data"
	default:
		return fmt.Sprintf("Unknown Type %d", h.RMainType)
	}
}

// GetDriLevelDescription returns the description of the DRI level
func (h *DatexHeader) GetDriLevelDescription() string {
	switch h.DriLevel {
	case DRI_LEVEL_95:
		return "1995 '95"
	case DRI_LEVEL_97:
		return "1997 '97"
	case DRI_LEVEL_98:
		return "1998 '98"
	case DRI_LEVEL_99:
		return "1999 '99"
	case DRI_LEVEL_00:
		return "2001 '01"
	case DRI_LEVEL_01:
		return "2002 '02"
	case DRI_LEVEL_02:
		return "2003 '03"
	case DRI_LEVEL_03:
		return "2005 '05"
	case DRI_LEVEL_04:
		return "2009 '09"
	case DRI_LEVEL_05:
		return "2015 '15"
	case DRI_LEVEL_06:
		return "2019 '19"
	default:
		return fmt.Sprintf("Unknown Level %d", h.DriLevel)
	}
}

// IsNetworkInterface returns true if this record type is valid for network interface only
func (h *DatexHeader) IsNetworkInterface() bool {
	switch h.RMainType {
	case DRI_MT_ALARM, DRI_MT_NETWORK, DRI_MT_FO:
		return true
	default:
		return false
	}
}

// ValidateReservedFields validates that reserved fields are zeroed
func (h *DatexHeader) ValidateReservedFields() error {
	if h.NSubnet != 0 {
		return fmt.Errorf("n_subnet must be zeroed, got %d", h.NSubnet)
	}
	if h.Reserved2 != 0 {
		return fmt.Errorf("reserved2 must be zeroed, got %d", h.Reserved2)
	}
	if h.Reserved3 != 0 {
		return fmt.Errorf("reserved3 must be zeroed, got %d", h.Reserved3)
	}
	return nil
}

// ZeroReservedFields sets all reserved fields to zero
func (h *DatexHeader) ZeroReservedFields() {
	h.NSubnet = 0
	h.Reserved2 = 0
	h.Reserved3 = 0
}

// DRI Waveform Subrecord Types
const (
	DRI_WF_CO2           = 9  // CO2 Interface level 3
	DRI_WF_O2            = 10 // O2 Interface level 3
	DRI_WF_N2O           = 11 // N2O Interface level 3
	DRI_WF_AA            = 12 // AA Interface level 3
	DRI_WF_AWP           = 13 // Airway pressure Interface level 3
	DRI_WF_FLOW          = 14 // Airway flow Interface level 3
	DRI_WF_RESP          = 15 // ECG respiratory waveform Interface level 3
	DRI_WF_INVP5         = 16 // Invasive Pressure channel 5 Interface level 3
	DRI_WF_INVP6         = 17 // Invasive Pressure channel 6 Interface level 3
	DRI_WF_EEG1          = 18 // EEG channel 1 Interface level 5
	DRI_WF_EEG2          = 19 // EEG channel 2 Interface level 5
	DRI_WF_EEG3          = 20 // EEG channel 3 Interface level 5
	DRI_WF_EEG4          = 21 // EEG channel 4 Interface level 5
	DRI_WF_ECG12         = 22 // 12 lead ECG packet Interface level 5
	DRI_WF_VOL           = 23 // Airway volume Interface level 5
	DRI_WF_TONO_PRESS    = 24 // Tonometry catheter pressure Interface level 5
	DRI_WF_SPI_LOOP_STATUS = 29 // Spirometry loop bit pattern Interface level 5
	DRI_WF_ENT_100       = 32 // Entropy Interface level 8
	DRI_WF_EEG_BIS       = 35 // BIS Interface level 8
	DRI_WF_INVP7         = 36 // Invasive Pressure channel 7 Interface level 9
	DRI_WF_INVP8         = 37 // Invasive Pressure channel 8 Interface level 9
	DRI_WF_PLETH_2       = 38 // Second Plethysmograph Interface level 9
	DRI_WF_RESP_100      = 39 // High resolution impedance respiration Interface level 11
)

// Waveform Status Bits
const (
	WF_STATUS_GAP        = 0x0001 // gap in sampling
	WF_STATUS_PACER_DET  = 0x0004 // pacer detected
	WF_STATUS_LEAD_OFF   = 0x0008 // ecg channel is off
)

// Sampling Rates per waveform subrecord (samples/s)
const (
	SAMPLE_RATE_ECG      = 300 // ECG x: μV
	SAMPLE_RATE_ECG12    = 500 // ECG x: μV (CARESCAPE monitors with software version 3.X)
	SAMPLE_RATE_INVP     = 100 // Invasive blood pressure x: 1/100 mmHg
	SAMPLE_RATE_PLETH    = 100 // Plethysmograph: modulation, 1/100%
	SAMPLE_RATE_CO2      = 25  // CO2 concentration: 1/100%
	SAMPLE_RATE_O2       = 25  // O2 concentration: 1/100%
	SAMPLE_RATE_N2O      = 25  // N2O concentration: 1/100%
	SAMPLE_RATE_AA       = 25  // Anesthesia agent: 1/100%
)

// WaveformHeader represents the waveform header structure
// C struct equivalent:
// struct wf_hdr {
//     short act_len;
//     word status;
//     word label;
// };
type WaveformHeader struct {
	ActLen int16  // The number of 16-bit waveform samples in the subrecord
	Status uint16 // Handled as bitfield. See Status bits in status field
	Label  uint16 // Reserved for future use
}

// Size returns the size of WaveformHeader in bytes
func (h *WaveformHeader) Size() int {
	return 6 // 2 + 2 + 2 bytes
}

// MarshalBinary converts the header to binary format
func (h *WaveformHeader) MarshalBinary() ([]byte, error) {
	buf := make([]byte, h.Size())
	binary.LittleEndian.PutUint16(buf[0:2], uint16(h.ActLen))
	binary.LittleEndian.PutUint16(buf[2:4], h.Status)
	binary.LittleEndian.PutUint16(buf[4:6], h.Label)
	return buf, nil
}

// UnmarshalBinary converts binary data to header
func (h *WaveformHeader) UnmarshalBinary(data []byte) error {
	if len(data) < h.Size() {
		return ErrInvalidDataLength
	}
	h.ActLen = int16(binary.LittleEndian.Uint16(data[0:2]))
	h.Status = binary.LittleEndian.Uint16(data[2:4])
	h.Label = binary.LittleEndian.Uint16(data[4:6])
	return nil
}

// HasGap returns true if there is a gap in sampling
func (h *WaveformHeader) HasGap() bool {
	return (h.Status & WF_STATUS_GAP) != 0
}

// HasPacerDetected returns true if pacer is detected
func (h *WaveformHeader) HasPacerDetected() bool {
	return (h.Status & WF_STATUS_PACER_DET) != 0
}

// HasLeadOff returns true if ECG channel is off
func (h *WaveformHeader) HasLeadOff() bool {
	return (h.Status & WF_STATUS_LEAD_OFF) != 0
}

// WaveformData represents waveform samples
type WaveformData struct {
	Header  WaveformHeader
	Samples []int16 // All samples are signed short integers (16 bits)
}

// Size returns the total size of WaveformData in bytes
func (w *WaveformData) Size() int {
	return w.Header.Size() + len(w.Samples)*2
}

// MarshalBinary converts the waveform data to binary format
func (w *WaveformData) MarshalBinary() ([]byte, error) {
	headerBytes, err := w.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}
	
	buf := make([]byte, w.Size())
	copy(buf, headerBytes)
	
	// Convert samples to bytes
	for i, sample := range w.Samples {
		binary.LittleEndian.PutUint16(buf[w.Header.Size()+i*2:], uint16(sample))
	}
	
	return buf, nil
}

// UnmarshalBinary converts binary data to waveform data
func (w *WaveformData) UnmarshalBinary(data []byte) error {
	if len(data) < w.Header.Size() {
		return ErrInvalidDataLength
	}
	
	// Parse header
	if err := w.Header.UnmarshalBinary(data[:w.Header.Size()]); err != nil {
		return err
	}
	
	// Parse samples
	sampleCount := int(w.Header.ActLen)
	if len(data) < w.Header.Size()+sampleCount*2 {
		return ErrInvalidDataLength
	}
	
	w.Samples = make([]int16, sampleCount)
	for i := 0; i < sampleCount; i++ {
		offset := w.Header.Size() + i*2
		w.Samples[i] = int16(binary.LittleEndian.Uint16(data[offset : offset+2]))
	}
	
	return nil
}

// IsControlCode returns true if the sample value is a control code
// Values less or equal to -32000 are no true measurement data but control codes
func IsControlCode(sample int16) bool {
	return sample <= -32000
}

// ConvertSampleToPhysicalValue converts sample value to physical value based on subrecord type
func ConvertSampleToPhysicalValue(sample int16, subrecordType int) float64 {
	if IsControlCode(sample) {
		return math.NaN()
	}
	
	switch subrecordType {
	case DRI_WF_ECG12:
		return float64(sample) // μV
	case DRI_WF_INVP5, DRI_WF_INVP6, DRI_WF_INVP7, DRI_WF_INVP8:
		return float64(sample) / 100.0 // mmHg
	case DRI_WF_PLETH, DRI_WF_PLETH_2:
		return float64(sample) / 100.0 // %
	case DRI_WF_CO2, DRI_WF_O2, DRI_WF_N2O, DRI_WF_AA:
		return float64(sample) / 100.0 // %
	default:
		return float64(sample)
	}
}

// GetSamplingRate returns the sampling rate for the given subrecord type
func GetSamplingRate(subrecordType int) int {
	switch subrecordType {
	case DRI_WF_ECG12:
		return SAMPLE_RATE_ECG12
	case DRI_WF_INVP5, DRI_WF_INVP6, DRI_WF_INVP7, DRI_WF_INVP8:
		return SAMPLE_RATE_INVP
	case DRI_WF_PLETH, DRI_WF_PLETH_2:
		return SAMPLE_RATE_PLETH
	case DRI_WF_CO2, DRI_WF_O2, DRI_WF_N2O, DRI_WF_AA:
		return SAMPLE_RATE_CO2
	default:
		return SAMPLE_RATE_ECG
	}
}

// Error definitions
var (
	ErrInvalidDataLength = &DRIError{Message: "invalid data length"}
)

// DRIError represents DRI-specific errors
type DRIError struct {
	Message string
}

func (e *DRIError) Error() string {
	return "DRI error: " + e.Message
}

// Physiological Database Record Structure (Updated based on PDF)
// 3.3.3 Displayed, 10S Trend and 60S Trend values
// C struct equivalent:
// struct dri_phdb {
//     dword time;
//     union {
//         struct basic_phdb basic;
//         struct ext1_phdb ext1;
//         struct ext2_phdb ext2;
//         struct ext3_phdb ext3;
//     } physdata;
//     byte marker;
//     byte reserved;
//     word cl_drilvl_subt;
// };
type PhysiologicalDatabaseRecord struct {
	Time           uint32                        // Contains the time stamp of the record in Unix time
	PhysData       PhysiologicalDataUnion       // Union of physiological data structures
	Marker         byte                          // Contains the number of the latest entered mark
	Reserved       byte                          // Reserved for future use
	ClDriLvlSubt   uint16                       // See Table 3-5 Usage of cl_drilvl_subt
}

// Physiological Data Union Structure
// C struct equivalent:
// union {
//     struct basic_phdb basic;
//     struct ext1_phdb ext1;
//     struct ext2_phdb ext2;
//     struct ext3_phdb ext3;
// } physdata;
type PhysiologicalDataUnion struct {
	Basic *BasicPhysiologicalData
	Ext1  *Extended1PhysiologicalData
	Ext2  *Extended2PhysiologicalData
	Ext3  *Extended3PhysiologicalData
}

// Size returns the total size of PhysiologicalDatabaseRecord in bytes
func (p *PhysiologicalDatabaseRecord) Size() int {
	baseSize := 4 + 1 + 1 + 2 // time + marker + reserved + cl_drilvl_subt
	if p.PhysData.Basic != nil {
		return baseSize + p.PhysData.Basic.Size()
	} else if p.PhysData.Ext1 != nil {
		return baseSize + p.PhysData.Ext1.Size()
	} else if p.PhysData.Ext2 != nil {
		return baseSize + p.PhysData.Ext2.Size()
	} else if p.PhysData.Ext3 != nil {
		return baseSize + p.PhysData.Ext3.Size()
	}
	return baseSize
}

// UnmarshalBinary converts binary data to physiological database record
func (p *PhysiologicalDatabaseRecord) UnmarshalBinary(data []byte) error {
	if len(data) < 8 {
		return ErrInvalidDataLength
	}

	offset := 0

	// time: Contains the time stamp of the record in Unix time
	p.Time = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// physdata: Union of physiological data structures
	// Note: This would need to be determined based on the cl_drilvl_subt field or other context
	// For now, we'll assume Basic class as default
	p.PhysData.Basic = &BasicPhysiologicalData{}
	if err := p.PhysData.Basic.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += p.PhysData.Basic.Size()

	// marker: Contains the number of the latest entered mark
	p.Marker = data[offset]
	offset += 1

	// reserved: Reserved for future use
	p.Reserved = data[offset]
	offset += 1

	// cl_drilvl_subt: See Table 3-5 Usage of cl_drilvl_subt
	p.ClDriLvlSubt = binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	return nil
}

// IsValid returns true if this physiological database record is valid
func (p *PhysiologicalDatabaseRecord) IsValid() bool {
	return p.Time > 0 && (p.PhysData.Basic != nil || p.PhysData.Ext1 != nil || p.PhysData.Ext2 != nil || p.PhysData.Ext3 != nil)
}

// GetDataClass returns the data class based on the union content
func (p *PhysiologicalDatabaseRecord) GetDataClass() int {
	if p.PhysData.Basic != nil {
		return PH_DATA_CLASS_BASIC
	} else if p.PhysData.Ext1 != nil {
		return PH_DATA_CLASS_EXT1
	} else if p.PhysData.Ext2 != nil {
		return PH_DATA_CLASS_EXT2
	} else if p.PhysData.Ext3 != nil {
		return PH_DATA_CLASS_EXT3
	}
	return PH_DATA_CLASS_BASIC
}

// GetSubrecordTypeName returns the human-readable name for the subrecord type
func (p *PhysiologicalDatabaseRecord) GetSubrecordTypeName() string {
	switch p.SubrecordType {
	case DRI_PH_DISPL:
		return "Displayed Values"
	case DRI_PH_10S_TREND:
		return "10 Second Trended Values"
	case DRI_PH_60S_TREND:
		return "60 Second Trended Values"
	case DRI_PH_AUX_INFO:
		return "Auxiliary Information"
	default:
		return fmt.Sprintf("Unknown Type %d", p.SubrecordType)
	}
}

// GetSubrecordClass returns the subrecord class based on subrecord type
func (p *PhysiologicalDatabaseRecord) GetSubrecordClass() int {
	switch p.SubrecordType {
	case DRI_PH_DISPL:
		return PH_CLASS_DISPLAYED
	case DRI_PH_10S_TREND:
		return PH_CLASS_TREND_10S
	case DRI_PH_60S_TREND:
		return PH_CLASS_TREND_60S
	case DRI_PH_AUX_INFO:
		return PH_CLASS_AUXILIARY
	default:
		return PH_CLASS_DISPLAYED
	}
}

// IsAuxiliaryData returns true if this is auxiliary physiological information
func (p *PhysiologicalDatabaseRecord) IsAuxiliaryData() bool {
	return p.SubrecordType == DRI_PH_AUX_INFO
}

// IsTrendedData returns true if this is trended data (10s or 60s)
func (p *PhysiologicalDatabaseRecord) IsTrendedData() bool {
	return p.SubrecordType == DRI_PH_10S_TREND || p.SubrecordType == DRI_PH_60S_TREND
}

// IsDisplayedData returns true if this is displayed data
func (p *PhysiologicalDatabaseRecord) IsDisplayedData() bool {
	return p.SubrecordType == DRI_PH_DISPL
}

// GetTrendInterval returns the trend interval in seconds
func (p *PhysiologicalDatabaseRecord) GetTrendInterval() int {
	switch p.SubrecordType {
	case DRI_PH_10S_TREND:
		return 10
	case DRI_PH_60S_TREND:
		return 60
	default:
		return 0
	}
}

// IsComputerInterfaceOnly returns true if this subrecord type is valid for computer interface only
func (p *PhysiologicalDatabaseRecord) IsComputerInterfaceOnly() bool {
	return p.SubrecordType == DRI_PH_10S_TREND
}

// GetTimestamp returns the timestamp as time.Time
func (p *PhysiologicalDatabaseRecord) GetTimestamp() time.Time {
	return time.Unix(int64(p.Time), 0)
}

// SetTimestamp sets the timestamp from time.Time
func (p *PhysiologicalDatabaseRecord) SetTimestamp(t time.Time) {
	p.Time = uint32(t.Unix())
}

// ToJSON converts the physiological database record to JSON format
func (p *PhysiologicalDatabaseRecord) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"timestamp":        p.GetTimestamp().Format(time.RFC3339),
		"unix_timestamp":   p.Time,
		"marker":           p.Marker,
		"reserved":         p.Reserved,
		"cl_drilvl_subt":   p.ClDriLvlSubt,
		"data_class":       p.GetDataClass(),
		"data_class_name":  GetDataClassName(p.GetDataClass()),
		"is_valid":         p.IsValid(),
	}

	// Add physiological data based on the union content
	if p.PhysData.Basic != nil {
		result["physiological_data"] = map[string]interface{}{
			"type": "basic",
			"data": p.PhysData.Basic.Data,
		}
	} else if p.PhysData.Ext1 != nil {
		result["physiological_data"] = map[string]interface{}{
			"type": "extended1",
			"data": p.PhysData.Ext1.Data,
		}
	} else if p.PhysData.Ext2 != nil {
		result["physiological_data"] = map[string]interface{}{
			"type": "extended2",
			"data": p.PhysData.Ext2.Data,
		}
	} else if p.PhysData.Ext3 != nil {
		result["physiological_data"] = map[string]interface{}{
			"type": "extended3",
			"data": p.PhysData.Ext3.Data,
		}
	}

	return result
}

// Basic Physiological Data Structure
// C struct equivalent:
// struct basic_phdb {
//     // Basic physiological data fields
//     // This structure would contain ECG, blood pressures, temperatures, SpO2, gases, etc.
// };
type BasicPhysiologicalData struct {
	// Basic physiological data fields would be defined here
	// ECG, blood pressures, temperatures, SpO2, gases, spirometry flow and volume, C.O., PCWP, NMT, SvO2, etc.
	Data []byte // Placeholder for actual data structure
}

// Size returns the size of BasicPhysiologicalData in bytes
func (b *BasicPhysiologicalData) Size() int {
	return len(b.Data)
}

// UnmarshalBinary converts binary data to basic physiological data
func (b *BasicPhysiologicalData) UnmarshalBinary(data []byte) error {
	b.Data = make([]byte, len(data))
	copy(b.Data, data)
	return nil
}

// ToJSON converts the basic physiological data to JSON format
func (b *BasicPhysiologicalData) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type": "basic",
		"data": b.Data,
		"size": len(b.Data),
	}
}

// Extended 1 Physiological Data Structure
// C struct equivalent:
// struct ext1_phdb {
//     // Extended 1 physiological data fields
//     // Arrhythmia analysis and ST analysis data, 12-lead ECG data, etc.
// };
type Extended1PhysiologicalData struct {
	// Extended 1 physiological data fields would be defined here
	// Arrhythmia analysis and ST analysis data, 12-lead ECG data, invasive blood pressure channels 7 and 8, 2nd SpO2 channel, temperature channels 5 and 6
	Data []byte // Placeholder for actual data structure
}

// Size returns the size of Extended1PhysiologicalData in bytes
func (e *Extended1PhysiologicalData) Size() int {
	return len(e.Data)
}

// UnmarshalBinary converts binary data to extended 1 physiological data
func (e *Extended1PhysiologicalData) UnmarshalBinary(data []byte) error {
	e.Data = make([]byte, len(data))
	copy(e.Data, data)
	return nil
}

// ToJSON converts the extended 1 physiological data to JSON format
func (e *Extended1PhysiologicalData) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type": "extended1",
		"data": e.Data,
		"size": len(e.Data),
	}
}

// Extended 2 Physiological Data Structure
// C struct equivalent:
// struct ext2_phdb {
//     // Extended 2 physiological data fields
//     // More NMT data, EEG, entropy, surgical pleth index data
// };
type Extended2PhysiologicalData struct {
	// Extended 2 physiological data fields would be defined here
	// More NMT data, EEG, entropy, surgical pleth index data
	Data []byte // Placeholder for actual data structure
}

// Size returns the size of Extended2PhysiologicalData in bytes
func (e *Extended2PhysiologicalData) Size() int {
	return len(e.Data)
}

// UnmarshalBinary converts binary data to extended 2 physiological data
func (e *Extended2PhysiologicalData) UnmarshalBinary(data []byte) error {
	e.Data = make([]byte, len(data))
	copy(e.Data, data)
	return nil
}

// ToJSON converts the extended 2 physiological data to JSON format
func (e *Extended2PhysiologicalData) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type": "extended2",
		"data": e.Data,
		"size": len(e.Data),
	}
}

// Extended 3 Physiological Data Structure
// C struct equivalent:
// struct ext3_phdb {
//     // Extended 3 physiological data fields
//     // More gas measurement data, gas exchange data, more spirometry parameters, etc.
// };
type Extended3PhysiologicalData struct {
	// Extended 3 physiological data fields would be defined here
	// More gas measurement data, gas exchange data, more spirometry parameters, tonometry, invasive pressure data, delta pressure, CPP and PiCCO data
	Data []byte // Placeholder for actual data structure
}

// Size returns the size of Extended3PhysiologicalData in bytes
func (e *Extended3PhysiologicalData) Size() int {
	return len(e.Data)
}

// UnmarshalBinary converts binary data to extended 3 physiological data
func (e *Extended3PhysiologicalData) UnmarshalBinary(data []byte) error {
	e.Data = make([]byte, len(data))
	copy(e.Data, data)
	return nil
}

// ToJSON converts the extended 3 physiological data to JSON format
func (e *Extended3PhysiologicalData) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type": "extended3",
		"data": e.Data,
		"size": len(e.Data),
	}
}

// Physiological Data Subrecord Classes
// 3.3.1 Physiological subrecord classes
const (
	PH_CLASS_DISPLAYED   = 0 // Displayed values of the physiological database
	PH_CLASS_TREND_10S   = 1 // 10 second trended values (computer interface only)
	PH_CLASS_TREND_60S   = 2 // 60 second trended values
	PH_CLASS_AUXILIARY   = 3 // Auxiliary Physiological Information
)

// Physiological Subrecord Data Classes
// Table 3-2 Physiological subrecord data classes
const (
	PH_DATA_CLASS_BASIC = 0 // Basic physiological data: ECG, blood pressures, temperatures, SpO2, gases, spirometry flow and volume, C.O., PCWP, NMT, SvO2, etc.
	PH_DATA_CLASS_EXT1  = 1 // Arrhythmia analysis and ST analysis data, 12-lead ECG data, invasive blood pressure channels 7 and 8, 2nd SpO2 channel, temperature channels 5 and 6
	PH_DATA_CLASS_EXT2  = 2 // More NMT data, EEG, entropy, surgical pleth index data
	PH_DATA_CLASS_EXT3  = 3 // More gas measurement data, gas exchange data, more spirometry parameters, tonometry, invasive pressure data, delta pressure, CPP and PiCCO data
)

// Physiological Data Class Bit Masks
// Table 3-3 Bit description for the phdb_class_bf
const (
	DRI_PHDBCL_REQ_BASIC_MASK  = 0x0000 // Enable sending of Basic physiological data class
	DRI_PHDBCL_DENY_BASIC_MASK = 0x0001 // Disable sending of Basic physiological data class
)

// DRI Physiological Database Class Enumeration
// Table 3-5 Usage of cl_drilvl_subt - enum dri_phdb_class
const (
	DRI_PHDBCL_BASIC = 0 // Basic physiological data class
	DRI_PHDBCL_EXT1  = 1 // Extended 1 physiological data class
	DRI_PHDBCL_EXT2  = 2 // Extended 2 physiological data class
	DRI_PHDBCL_EXT3  = 3 // Extended 3 physiological data class
)

// Bit masks for cl_drilvl_subt field
// Table 3-5 Usage of cl_drilvl_subt
const (
	CL_DRILVL_SUBT_CLASS_MASK = 0x0F00 // Bits 8-11: Class of the subrecord
	CL_DRILVL_SUBT_RESERVED1  = 0x00FF // Bits 0-7: Reserved for future extensions
	CL_DRILVL_SUBT_RESERVED2  = 0xF000 // Bits 12-15: Reserved for future extensions
)

// Physiological Data Class Bit Field Structure
// C struct equivalent:
// struct phdb_class_bf {
//     word basic_class;
//     word ext1_class;
//     word ext2_class;
//     word ext3_class;
// };
type PhysiologicalDataClassBitField struct {
	BasicClass uint16 // Basic physiological data class bit mask
	Ext1Class  uint16 // Extended 1 physiological data class bit mask
	Ext2Class  uint16 // Extended 2 physiological data class bit mask
	Ext3Class  uint16 // Extended 3 physiological data class bit mask
}

// Size returns the size of PhysiologicalDataClassBitField in bytes
func (p *PhysiologicalDataClassBitField) Size() int {
	return 8 // 2 + 2 + 2 + 2 bytes
}

// UnmarshalBinary converts binary data to physiological data class bit field
func (p *PhysiologicalDataClassBitField) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return ErrInvalidDataLength
	}

	offset := 0
	p.BasicClass = binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	p.Ext1Class = binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	p.Ext2Class = binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	p.Ext3Class = binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	return nil
}

// IsBasicClassEnabled returns true if Basic physiological data class is enabled
func (p *PhysiologicalDataClassBitField) IsBasicClassEnabled() bool {
	return p.BasicClass == DRI_PHDBCL_REQ_BASIC_MASK
}

// IsBasicClassDisabled returns true if Basic physiological data class is disabled
func (p *PhysiologicalDataClassBitField) IsBasicClassDisabled() bool {
	return p.BasicClass == DRI_PHDBCL_DENY_BASIC_MASK
}

// EnableBasicClass enables Basic physiological data class
func (p *PhysiologicalDataClassBitField) EnableBasicClass() {
	p.BasicClass = DRI_PHDBCL_REQ_BASIC_MASK
}

// DisableBasicClass disables Basic physiological data class
func (p *PhysiologicalDataClassBitField) DisableBasicClass() {
	p.BasicClass = DRI_PHDBCL_DENY_BASIC_MASK
}

// Auxiliary Physiological Information Structure
// 3.3.4 Auxiliary Physiological Information
// C struct equivalent:
// struct aux_phdb_info {
//     dword nibp_time;
//     short reserved1;
//     dword co_time;
//     dword pcwp_time;
//     short pat_bsa;
//     byte reserved[98];
// };
type AuxiliaryPhysiologicalInfo struct {
	NibpTime  uint32   // Time of the latest NIBP measurement (seconds since 1.1.1970)
	Reserved1 int16    // Reserved
	CoTime    uint32   // Time of the latest Cardiac Output measurement (seconds since 1.1.1970)
	PcwpTime  uint32   // Time of the latest PCWP measurement (seconds since 1.1.1970)
	PatBsa    int16    // Patient's body surface area (1/100 m2)
	Reserved  [98]byte // Reserved
}

// Size returns the size of AuxiliaryPhysiologicalInfo in bytes
func (a *AuxiliaryPhysiologicalInfo) Size() int {
	return 4 + 2 + 4 + 4 + 2 + 98 // 114 bytes total
}

// MarshalBinary converts the auxiliary physiological info to binary format
func (a *AuxiliaryPhysiologicalInfo) MarshalBinary() ([]byte, error) {
	buf := make([]byte, a.Size())
	offset := 0

	// nibp_time: Time of the latest NIBP measurement
	binary.LittleEndian.PutUint32(buf[offset:], a.NibpTime)
	offset += 4

	// reserved1: Reserved
	binary.LittleEndian.PutUint16(buf[offset:], uint16(a.Reserved1))
	offset += 2

	// co_time: Time of the latest Cardiac Output measurement
	binary.LittleEndian.PutUint32(buf[offset:], a.CoTime)
	offset += 4

	// pcwp_time: Time of the latest PCWP measurement
	binary.LittleEndian.PutUint32(buf[offset:], a.PcwpTime)
	offset += 4

	// pat_bsa: Patient's body surface area
	binary.LittleEndian.PutUint16(buf[offset:], uint16(a.PatBsa))
	offset += 2

	// reserved: Reserved
	copy(buf[offset:], a.Reserved[:])
	offset += 98

	return buf, nil
}

// UnmarshalBinary converts binary data to auxiliary physiological info
func (a *AuxiliaryPhysiologicalInfo) UnmarshalBinary(data []byte) error {
	if len(data) < a.Size() {
		return ErrInvalidDataLength
	}

	offset := 0

	// nibp_time: Time of the latest NIBP measurement
	a.NibpTime = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// reserved1: Reserved
	a.Reserved1 = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2

	// co_time: Time of the latest Cardiac Output measurement
	a.CoTime = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// pcwp_time: Time of the latest PCWP measurement
	a.PcwpTime = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// pat_bsa: Patient's body surface area
	a.PatBsa = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2

	// reserved: Reserved
	copy(a.Reserved[:], data[offset:])
	offset += 98

	return nil
}

// GetNibpTime returns the NIBP measurement time as time.Time
func (a *AuxiliaryPhysiologicalInfo) GetNibpTime() time.Time {
	if a.NibpTime == 0 {
		return time.Time{} // Return zero time if not known
	}
	return time.Unix(int64(a.NibpTime), 0)
}

// SetNibpTime sets the NIBP measurement time from time.Time
func (a *AuxiliaryPhysiologicalInfo) SetNibpTime(t time.Time) {
	if t.IsZero() {
		a.NibpTime = 0
	} else {
		a.NibpTime = uint32(t.Unix())
	}
}

// GetCoTime returns the Cardiac Output measurement time as time.Time
func (a *AuxiliaryPhysiologicalInfo) GetCoTime() time.Time {
	if a.CoTime == 0 {
		return time.Time{} // Return zero time if not known
	}
	return time.Unix(int64(a.CoTime), 0)
}

// SetCoTime sets the Cardiac Output measurement time from time.Time
func (a *AuxiliaryPhysiologicalInfo) SetCoTime(t time.Time) {
	if t.IsZero() {
		a.CoTime = 0
	} else {
		a.CoTime = uint32(t.Unix())
	}
}

// GetPcwpTime returns the PCWP measurement time as time.Time
func (a *AuxiliaryPhysiologicalInfo) GetPcwpTime() time.Time {
	if a.PcwpTime == 0 {
		return time.Time{} // Return zero time if not known
	}
	return time.Unix(int64(a.PcwpTime), 0)
}

// SetPcwpTime sets the PCWP measurement time from time.Time
func (a *AuxiliaryPhysiologicalInfo) SetPcwpTime(t time.Time) {
	if t.IsZero() {
		a.PcwpTime = 0
	} else {
		a.PcwpTime = uint32(t.Unix())
	}
}

// GetBodySurfaceArea returns the body surface area in m2
func (a *AuxiliaryPhysiologicalInfo) GetBodySurfaceArea() float64 {
	return float64(a.PatBsa) / 100.0 // Convert from 1/100 m2 to m2
}

// SetBodySurfaceArea sets the body surface area in m2
func (a *AuxiliaryPhysiologicalInfo) SetBodySurfaceArea(bsa float64) {
	a.PatBsa = int16(bsa * 100.0) // Convert from m2 to 1/100 m2
}

// IsValid returns true if this auxiliary physiological info is valid
func (a *AuxiliaryPhysiologicalInfo) IsValid() bool {
	// Check if at least one measurement time is valid (non-zero)
	return a.NibpTime > 0 || a.CoTime > 0 || a.PcwpTime > 0
}

// GetDataClassName returns the human-readable name for the data class
func GetDataClassName(dataClass int) string {
	switch dataClass {
	case PH_DATA_CLASS_BASIC:
		return "Basic"
	case PH_DATA_CLASS_EXT1:
		return "Extended 1"
	case PH_DATA_CLASS_EXT2:
		return "Extended 2"
	case PH_DATA_CLASS_EXT3:
		return "Extended 3"
	default:
		return fmt.Sprintf("Unknown Class %d", dataClass)
	}
}

// GetDataClassDescription returns the detailed description for the data class
func GetDataClassDescription(dataClass int) string {
	switch dataClass {
	case PH_DATA_CLASS_BASIC:
		return "Basic physiological data: ECG, blood pressures, temperatures, SpO2, gases, spirometry flow and volume, C.O., PCWP, NMT, SvO2, etc."
	case PH_DATA_CLASS_EXT1:
		return "Arrhythmia analysis and ST analysis data, 12-lead ECG data, invasive blood pressure channels 7 and 8, 2nd SpO2 channel, temperature channels 5 and 6"
	case PH_DATA_CLASS_EXT2:
		return "More NMT data, EEG, entropy, surgical pleth index data"
	case PH_DATA_CLASS_EXT3:
		return "More gas measurement data, gas exchange data, more spirometry parameters, tonometry, invasive pressure data, delta pressure, CPP and PiCCO data"
	default:
		return fmt.Sprintf("Unknown data class %d", dataClass)
	}
}

// GetRequiredInterfaceLevel returns the required interface level for the data class
func GetRequiredInterfaceLevel(dataClass int) int {
	switch dataClass {
	case PH_DATA_CLASS_BASIC:
		return 1
	case PH_DATA_CLASS_EXT1, PH_DATA_CLASS_EXT2, PH_DATA_CLASS_EXT3:
		return 3
	default:
		return 1
	}
}

// IsDataClassSupported returns true if the data class is supported at the given interface level
func IsDataClassSupported(dataClass int, interfaceLevel int) bool {
	requiredLevel := GetRequiredInterfaceLevel(dataClass)
	return interfaceLevel >= requiredLevel
}

// GetDataClassFromClDriLvlSubt extracts the data class from cl_drilvl_subt field
func GetDataClassFromClDriLvlSubt(clDriLvlSubt uint16) int {
	classBits := (clDriLvlSubt & CL_DRILVL_SUBT_CLASS_MASK) >> 8
	return int(classBits)
}

// SetDataClassInClDriLvlSubt sets the data class in cl_drilvl_subt field
func SetDataClassInClDriLvlSubt(clDriLvlSubt uint16, dataClass int) uint16 {
	// Clear the class bits (bits 8-11)
	clDriLvlSubt &= ^CL_DRILVL_SUBT_CLASS_MASK
	// Set the new class bits
	clDriLvlSubt |= uint16(dataClass) << 8
	return clDriLvlSubt
}

// GetDriLvlSubtClassName returns the human-readable name for the data class from cl_drilvl_subt
func GetDriLvlSubtClassName(clDriLvlSubt uint16) string {
	dataClass := GetDataClassFromClDriLvlSubt(clDriLvlSubt)
	switch dataClass {
	case DRI_PHDBCL_BASIC:
		return "Basic"
	case DRI_PHDBCL_EXT1:
		return "Extended 1"
	case DRI_PHDBCL_EXT2:
		return "Extended 2"
	case DRI_PHDBCL_EXT3:
		return "Extended 3"
	default:
		return fmt.Sprintf("Unknown Class %d", dataClass)
	}
}

// Group Header Structure (Common for all groups)
// C struct equivalent:
// struct group_hdr {
//     word status;
//     word label;
// };
type GroupHeader struct {
	Status uint16 // Status field with group-specific bits
	Label  uint16 // Label field with group-specific values
}

// Size returns the size of GroupHeader in bytes
func (h *GroupHeader) Size() int {
	return 4 // 2 + 2 bytes
}

// UnmarshalBinary converts binary data to group header
func (h *GroupHeader) UnmarshalBinary(data []byte) error {
	if len(data) < h.Size() {
		return ErrInvalidDataLength
	}
	h.Status = binary.LittleEndian.Uint16(data[0:2])
	h.Label = binary.LittleEndian.Uint16(data[2:4])
	return nil
}

// ToJSON converts the GroupHeader to JSON format
func (h *GroupHeader) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"status": h.Status,
		"label":  h.Label,
	}
}

// O2 Group Structure
// Table 3-31 O2 data fields
// C struct equivalent:
// struct o2_group {
//     struct group_hdr hdr;
//     short et;
//     short fi;
// };
type O2Group struct {
	Header GroupHeader // Group header with status and label
	Et     int16      // Expiratory concentration (1/100%)
	Fi     int16      // Inspiratory concentration (1/100%)
}

// Size returns the size of O2Group in bytes
func (o *O2Group) Size() int {
	return o.Header.Size() + 4 // header + 2 + 2 bytes
}

// UnmarshalBinary converts binary data to O2 group
func (o *O2Group) UnmarshalBinary(data []byte) error {
	if len(data) < o.Size() {
		return ErrInvalidDataLength
	}
	
	offset := 0
	if err := o.Header.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += o.Header.Size()
	
	o.Et = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	o.Fi = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	return nil
}

// GetExpiratoryConcentration returns the expiratory concentration in %
func (o *O2Group) GetExpiratoryConcentration() float64 {
	return float64(o.Et) / 100.0
}

// GetInspiratoryConcentration returns the inspiratory concentration in %
func (o *O2Group) GetInspiratoryConcentration() float64 {
	return float64(o.Fi) / 100.0
}

// ToJSON converts the O2Group to JSON format
func (o *O2Group) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"header": o.Header.ToJSON(),
		"et": map[string]interface{}{
			"raw_value": o.Et,
			"percent":   o.GetExpiratoryConcentration(),
			"unit":      "%",
		},
		"fi": map[string]interface{}{
			"raw_value": o.Fi,
			"percent":   o.GetInspiratoryConcentration(),
			"unit":      "%",
		},
	}
}

// N2O Group Structure
// Table 3-33 N2O data fields
// C struct equivalent:
// struct n2o_group {
//     struct group_hdr hdr;
//     short et;
//     short fi;
// };
type N2OGroup struct {
	Header GroupHeader // Group header with status and label
	Et     int16      // Expiratory concentration (1/100%)
	Fi     int16      // Inspiratory concentration (1/100%)
}

// Size returns the size of N2OGroup in bytes
func (n *N2OGroup) Size() int {
	return n.Header.Size() + 4 // header + 2 + 2 bytes
}

// UnmarshalBinary converts binary data to N2O group
func (n *N2OGroup) UnmarshalBinary(data []byte) error {
	if len(data) < n.Size() {
		return ErrInvalidDataLength
	}
	
	offset := 0
	if err := n.Header.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += n.Header.Size()
	
	n.Et = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	n.Fi = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	return nil
}

// GetExpiratoryConcentration returns the expiratory concentration in %
func (n *N2OGroup) GetExpiratoryConcentration() float64 {
	return float64(n.Et) / 100.0
}

// GetInspiratoryConcentration returns the inspiratory concentration in %
func (n *N2OGroup) GetInspiratoryConcentration() float64 {
	return float64(n.Fi) / 100.0
}

// IsCalibrating returns true if N2O is calibrating
func (n *N2OGroup) IsCalibrating() bool {
	return (n.Header.Status & 0x0004) != 0 // Bit 2
}

// IsMeasurementOff returns true if N2O measurement is off
func (n *N2OGroup) IsMeasurementOff() bool {
	return (n.Header.Status & 0x0008) != 0 // Bit 3
}

// ToJSON converts the N2OGroup to JSON format
func (n *N2OGroup) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"header": map[string]interface{}{
			"status": n.Header.Status,
			"label":  n.Header.Label,
			"is_calibrating": n.IsCalibrating(),
			"is_measurement_off": n.IsMeasurementOff(),
		},
		"et": map[string]interface{}{
			"raw_value": n.Et,
			"percent":   n.GetExpiratoryConcentration(),
			"unit":      "%",
		},
		"fi": map[string]interface{}{
			"raw_value": n.Fi,
			"percent":   n.GetInspiratoryConcentration(),
			"unit":      "%",
		},
	}
}

// Anesthesia Agent Label Constants
// Table 3-35 Anesthesia Agent label field (word) values
const (
	DRI_UNKNOWN = 0 // Unknown
	DRI_NONE    = 1 // NONE
	DRI_HAL     = 2 // HAL
	DRI_ENF     = 3 // ENF
	DRI_ISO     = 4 // ISO
	DRI_DES     = 5 // DES
	DRI_SEV     = 6 // SEV
)

// Anesthesia Agent Group Structure
// Table 3-36 Anesthesia Agent data fields
// C struct equivalent:
// struct aa_group {
//     struct group_hdr hdr;
//     short et;
//     short fi;
//     short mac_sum;
// };
type AnesthesiaAgentGroup struct {
	Header GroupHeader // Group header with status and label
	Et     int16      // Expiratory concentration (1/100%)
	Fi     int16      // Inspiratory concentration (1/100%)
	MacSum int16      // Total Minimum Alveolar Concentration (1/100)
}

// Size returns the size of AnesthesiaAgentGroup in bytes
func (a *AnesthesiaAgentGroup) Size() int {
	return a.Header.Size() + 6 // header + 2 + 2 + 2 bytes
}

// UnmarshalBinary converts binary data to Anesthesia Agent group
func (a *AnesthesiaAgentGroup) UnmarshalBinary(data []byte) error {
	if len(data) < a.Size() {
		return ErrInvalidDataLength
	}
	
	offset := 0
	if err := a.Header.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += a.Header.Size()
	
	a.Et = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	a.Fi = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	a.MacSum = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	return nil
}

// GetExpiratoryConcentration returns the expiratory concentration in %
func (a *AnesthesiaAgentGroup) GetExpiratoryConcentration() float64 {
	return float64(a.Et) / 100.0
}

// GetInspiratoryConcentration returns the inspiratory concentration in %
func (a *AnesthesiaAgentGroup) GetInspiratoryConcentration() float64 {
	return float64(a.Fi) / 100.0
}

// GetMacSum returns the MAC sum value
func (a *AnesthesiaAgentGroup) GetMacSum() float64 {
	return float64(a.MacSum) / 100.0
}

// GetAgentLabel returns the human-readable agent label
func (a *AnesthesiaAgentGroup) GetAgentLabel() string {
	switch a.Header.Label {
	case DRI_UNKNOWN:
		return "Unknown"
	case DRI_NONE:
		return "NONE"
	case DRI_HAL:
		return "HAL"
	case DRI_ENF:
		return "ENF"
	case DRI_ISO:
		return "ISO"
	case DRI_DES:
		return "DES"
	case DRI_SEV:
		return "SEV"
	default:
		return fmt.Sprintf("Unknown Agent %d", a.Header.Label)
	}
}

// IsCalibrating returns true if Anesthesia Agent is calibrating
func (a *AnesthesiaAgentGroup) IsCalibrating() bool {
	return (a.Header.Status & 0x0004) != 0 // Bit 2
}

// IsMeasurementOff returns true if Anesthesia Agent measurement is off
func (a *AnesthesiaAgentGroup) IsMeasurementOff() bool {
	return (a.Header.Status & 0x0008) != 0 // Bit 3
}

// ToJSON converts the AnesthesiaAgentGroup to JSON format
func (a *AnesthesiaAgentGroup) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"header": map[string]interface{}{
			"status": a.Header.Status,
			"label":  a.Header.Label,
			"agent_label": a.GetAgentLabel(),
			"is_calibrating": a.IsCalibrating(),
			"is_measurement_off": a.IsMeasurementOff(),
		},
		"et": map[string]interface{}{
			"raw_value": a.Et,
			"percent":   a.GetExpiratoryConcentration(),
			"unit":      "%",
		},
		"fi": map[string]interface{}{
			"raw_value": a.Fi,
			"percent":   a.GetInspiratoryConcentration(),
			"unit":      "%",
		},
		"mac_sum": map[string]interface{}{
			"raw_value": a.MacSum,
			"value":     a.GetMacSum(),
			"unit":      "MAC",
		},
	}
}

// TV Base Constants
// Table 3-37 Flow & Volume status field bits - enum dri_tv_base
const (
	DRI_ATPD = 0 // Atmospheric/ambient temperature and pressure, dry gas
	DRI_NTPD = 1 // Normal temperature and pressure, dry gas
	DRI_BTPS = 2 // Body temperature and pressure, saturated gas
	DRI_STPD = 3 // Standard temperature and pressure, dry gas
	DRI_NR_TV_BASE = 4
)

// Flow & Volume Group Structure
// Table 3-38 Flow & Volume data fields
// C struct equivalent:
// struct flow_vol_group {
//     struct group_hdr hdr;
//     short rr;
//     short ppeak;
//     short peep;
//     short pplat;
//     short tv_insp;
//     short tv_exp;
//     short compliance;
//     short mv_exp;
// };
type FlowVolumeGroup struct {
	Header    GroupHeader // Group header with status and label
	Rr        int16       // Respiration rate (1/min)
	Ppeak     int16       // Peak pressure (1/100 cmH2O)
	Peep      int16       // Positive end expiratory pressure (1/100 cmH2O)
	Pplat     int16       // Plateau pressure (1/100 cmH2O)
	TvInsp    int16       // Inspiratory tidal volume (1/10 ml)
	TvExp     int16       // Expiratory tidal volume (1/10 ml)
	Compliance int16      // Compliance (1/100 ml/cmH2O)
	MvExp     int16       // Expiratory minute volume (1/100 l/min)
}

// Size returns the size of FlowVolumeGroup in bytes
func (f *FlowVolumeGroup) Size() int {
	return f.Header.Size() + 16 // header + 8 * 2 bytes
}

// UnmarshalBinary converts binary data to Flow & Volume group
func (f *FlowVolumeGroup) UnmarshalBinary(data []byte) error {
	if len(data) < f.Size() {
		return ErrInvalidDataLength
	}
	
	offset := 0
	if err := f.Header.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += f.Header.Size()
	
	f.Rr = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	f.Ppeak = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	f.Peep = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	f.Pplat = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	f.TvInsp = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	f.TvExp = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	f.Compliance = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	f.MvExp = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	return nil
}

// GetRespirationRate returns the respiration rate in breaths/min
func (f *FlowVolumeGroup) GetRespirationRate() float64 {
	return float64(f.Rr)
}

// GetPeakPressure returns the peak pressure in cmH2O
func (f *FlowVolumeGroup) GetPeakPressure() float64 {
	return float64(f.Ppeak) / 100.0
}

// GetPeep returns the PEEP in cmH2O
func (f *FlowVolumeGroup) GetPeep() float64 {
	return float64(f.Peep) / 100.0
}

// GetPlateauPressure returns the plateau pressure in cmH2O
func (f *FlowVolumeGroup) GetPlateauPressure() float64 {
	return float64(f.Pplat) / 100.0
}

// GetInspiratoryTidalVolume returns the inspiratory tidal volume in ml
func (f *FlowVolumeGroup) GetInspiratoryTidalVolume() float64 {
	return float64(f.TvInsp) / 10.0
}

// GetExpiratoryTidalVolume returns the expiratory tidal volume in ml
func (f *FlowVolumeGroup) GetExpiratoryTidalVolume() float64 {
	return float64(f.TvExp) / 10.0
}

// GetCompliance returns the compliance in ml/cmH2O
func (f *FlowVolumeGroup) GetCompliance() float64 {
	return float64(f.Compliance) / 100.0
}

// GetExpiratoryMinuteVolume returns the expiratory minute volume in l/min
func (f *FlowVolumeGroup) GetExpiratoryMinuteVolume() float64 {
	return float64(f.MvExp) / 100.0
}

// GetTvBase returns the TV measuring conditions
func (f *FlowVolumeGroup) GetTvBase() int {
	return int((f.Header.Status >> 8) & 0x03) // Bits 8-9
}

// GetTvBaseDescription returns the human-readable TV base description
func (f *FlowVolumeGroup) GetTvBaseDescription() string {
	switch f.GetTvBase() {
	case DRI_ATPD:
		return "Atmospheric/ambient temperature and pressure, dry gas"
	case DRI_NTPD:
		return "Normal temperature and pressure, dry gas"
	case DRI_BTPS:
		return "Body temperature and pressure, saturated gas"
	case DRI_STPD:
		return "Standard temperature and pressure, dry gas"
	default:
		return "Unknown TV base"
	}
}

// IsDisconnection returns true if there is a disconnection
func (f *FlowVolumeGroup) IsDisconnection() bool {
	return (f.Header.Status & 0x0004) != 0 // Bit 2
}

// IsCalibrating returns true if Flow & Volume is calibrating
func (f *FlowVolumeGroup) IsCalibrating() bool {
	return (f.Header.Status & 0x0008) != 0 // Bit 3
}

// IsZeroing returns true if Flow & Volume is zeroing
func (f *FlowVolumeGroup) IsZeroing() bool {
	return (f.Header.Status & 0x0010) != 0 // Bit 4
}

// IsObstruction returns true if there is an obstruction
func (f *FlowVolumeGroup) IsObstruction() bool {
	return (f.Header.Status & 0x0020) != 0 // Bit 5
}

// IsLeak returns true if there is a leak
func (f *FlowVolumeGroup) IsLeak() bool {
	return (f.Header.Status & 0x0040) != 0 // Bit 6
}

// IsMeasurementOff returns true if Flow & Volume measurement is off
func (f *FlowVolumeGroup) IsMeasurementOff() bool {
	return (f.Header.Status & 0x0080) != 0 // Bit 7
}

// ToJSON converts the FlowVolumeGroup to JSON format
func (f *FlowVolumeGroup) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"header": map[string]interface{}{
			"status": f.Header.Status,
			"label":  f.Header.Label,
			"tv_base": map[string]interface{}{
				"value": f.GetTvBase(),
				"description": f.GetTvBaseDescription(),
			},
			"status_bits": map[string]interface{}{
				"disconnection": f.IsDisconnection(),
				"calibrating": f.IsCalibrating(),
				"zeroing": f.IsZeroing(),
				"obstruction": f.IsObstruction(),
				"leak": f.IsLeak(),
				"measurement_off": f.IsMeasurementOff(),
			},
		},
		"rr": map[string]interface{}{
			"raw_value": f.Rr,
			"value":     f.GetRespirationRate(),
			"unit":      "breaths/min",
		},
		"ppeak": map[string]interface{}{
			"raw_value": f.Ppeak,
			"value":     f.GetPeakPressure(),
			"unit":      "cmH2O",
		},
		"peep": map[string]interface{}{
			"raw_value": f.Peep,
			"value":     f.GetPeep(),
			"unit":      "cmH2O",
		},
		"pplat": map[string]interface{}{
			"raw_value": f.Pplat,
			"value":     f.GetPlateauPressure(),
			"unit":      "cmH2O",
		},
		"tv_insp": map[string]interface{}{
			"raw_value": f.TvInsp,
			"value":     f.GetInspiratoryTidalVolume(),
			"unit":      "ml",
		},
		"tv_exp": map[string]interface{}{
			"raw_value": f.TvExp,
			"value":     f.GetExpiratoryTidalVolume(),
			"unit":      "ml",
		},
		"compliance": map[string]interface{}{
			"raw_value": f.Compliance,
			"value":     f.GetCompliance(),
			"unit":      "ml/cmH2O",
		},
		"mv_exp": map[string]interface{}{
			"raw_value": f.MvExp,
			"value":     f.GetExpiratoryMinuteVolume(),
			"unit":      "l/min",
		},
	}
}

// CO & PCWP Label Bit Constants
// Table 3-39 CO & PCWP label field bits usage
const (
	LBIT_CO_OVER_60S_OLD   = 0 // Age of CO reading is > 60 s
	LBIT_PCWP_OVER_60S_OLD = 1 // Age of PCWP reading is > 60 s
)

// Cardiac Output & Wedge Pressure Group Structure
// Table 3-40 CO & PCWP data fields
// C struct equivalent:
// struct co_wedge_group {
//     struct group_hdr hdr;
//     short co;
//     short blood_temp;
//     short ref;
//     short pcwp;
// };
type COWedgeGroup struct {
	Header    GroupHeader // Group header with status and label
	Co        int16       // Cardiac output (ml/min)
	BloodTemp int16       // Blood temperature (1/100 °C)
	Ref       int16       // Right heart ejection fraction (1/100 %)
	Pcwp      int16       // Wedge pressure (1/100 mmHg)
}

// Size returns the size of COWedgeGroup in bytes
func (c *COWedgeGroup) Size() int {
	return c.Header.Size() + 8 // header + 4 * 2 bytes
}

// UnmarshalBinary converts binary data to CO & Wedge Pressure group
func (c *COWedgeGroup) UnmarshalBinary(data []byte) error {
	if len(data) < c.Size() {
		return ErrInvalidDataLength
	}
	
	offset := 0
	if err := c.Header.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += c.Header.Size()
	
	c.Co = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	c.BloodTemp = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	c.Ref = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	c.Pcwp = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	return nil
}

// GetCardiacOutput returns the cardiac output in ml/min
func (c *COWedgeGroup) GetCardiacOutput() float64 {
	return float64(c.Co)
}

// GetBloodTemperature returns the blood temperature in °C
func (c *COWedgeGroup) GetBloodTemperature() float64 {
	return float64(c.BloodTemp) / 100.0
}

// GetRightHeartEjectionFraction returns the right heart ejection fraction in %
func (c *COWedgeGroup) GetRightHeartEjectionFraction() float64 {
	return float64(c.Ref) / 100.0
}

// GetWedgePressure returns the wedge pressure in mmHg
func (c *COWedgeGroup) GetWedgePressure() float64 {
	return float64(c.Pcwp) / 100.0
}

// IsCOOver60sOld returns true if CO reading is > 60s old
func (c *COWedgeGroup) IsCOOver60sOld() bool {
	return (c.Header.Label & (1 << LBIT_CO_OVER_60S_OLD)) != 0
}

// IsPCWPOver60sOld returns true if PCWP reading is > 60s old
func (c *COWedgeGroup) IsPCWPOver60sOld() bool {
	return (c.Header.Label & (1 << LBIT_PCWP_OVER_60S_OLD)) != 0
}

// GetCOMode returns the CO mode (0-3)
func (c *COWedgeGroup) GetCOMode() int {
	return int((c.Header.Label >> 2) & 0x07) // Bits 2-4
}

// GetCOModeDescription returns the human-readable CO mode description
func (c *COWedgeGroup) GetCOModeDescription() string {
	switch c.GetCOMode() {
	case 0:
		return "No mode"
	case 1:
		return "Bolus mode"
	case 2:
		return "Continuous mode"
	case 3:
		return "Reserved"
	default:
		return "Unknown mode"
	}
}

// ToJSON converts the COWedgeGroup to JSON format
func (c *COWedgeGroup) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"header": map[string]interface{}{
			"status": c.Header.Status,
			"label":  c.Header.Label,
			"co_over_60s_old": c.IsCOOver60sOld(),
			"pcwp_over_60s_old": c.IsPCWPOver60sOld(),
			"co_mode": map[string]interface{}{
				"value": c.GetCOMode(),
				"description": c.GetCOModeDescription(),
			},
		},
		"co": map[string]interface{}{
			"raw_value": c.Co,
			"value":     c.GetCardiacOutput(),
			"unit":      "ml/min",
		},
		"blood_temp": map[string]interface{}{
			"raw_value": c.BloodTemp,
			"value":     c.GetBloodTemperature(),
			"unit":      "°C",
		},
		"ref": map[string]interface{}{
			"raw_value": c.Ref,
			"value":     c.GetRightHeartEjectionFraction(),
			"unit":      "%",
		},
		"pcwp": map[string]interface{}{
			"raw_value": c.Pcwp,
			"value":     c.GetWedgePressure(),
			"unit":      "mmHg",
		},
	}
}

// Stimulus Type Constants
// Table 3-41 NMT status field bits - enum stim_typ
const (
	TOF      = 0 // Train Of Four (TOF mode)
	DBS      = 1 // Double Burst (DB mode)
	ST_STIM  = 2 // Single Twitch (ST mode)
	PTC_STIM = 3 // Post-tetanic count
	NR_STIM_TYPES = 4
)

// Pulse Width Constants
// Table 3-41 NMT status field bits - enum pulse_width_types
const (
	PULSE_NOT_USED = 0 // not used
	PULSE_100      = 1 // 100 us
	PULSE_200      = 2 // 200 us
	PULSE_300      = 3 // 300 us
	PULSE_NR       = 4
)

// NMT Group Structure
// Table 3-42 NMT data fields
// C struct equivalent:
// struct nmt_group {
//     struct group_hdr hdr;
//     short t1;
//     short tratio;
//     short ptc;
// };
type NMTGroup struct {
	Header GroupHeader // Group header with status and label
	T1     int16      // TOF Twitch 1 (1/10 %)
	Tratio int16      // t4/t1 in TOF mode, t2/t1 in DB mode (1/10 %)
	Ptc    int16      // Split into a bit field (see Table 3-43)
}

// Size returns the size of NMTGroup in bytes
func (n *NMTGroup) Size() int {
	return n.Header.Size() + 6 // header + 3 * 2 bytes
}

// UnmarshalBinary converts binary data to NMT group
func (n *NMTGroup) UnmarshalBinary(data []byte) error {
	if len(data) < n.Size() {
		return ErrInvalidDataLength
	}
	
	offset := 0
	if err := n.Header.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += n.Header.Size()
	
	n.T1 = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	n.Tratio = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	n.Ptc = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	return nil
}

// GetT1 returns the T1 value in %
func (n *NMTGroup) GetT1() float64 {
	return float64(n.T1) / 10.0
}

// GetTratio returns the T ratio value in %
func (n *NMTGroup) GetTratio() float64 {
	return float64(n.Tratio) / 10.0
}

// GetStimulusMode returns the stimulus mode
func (n *NMTGroup) GetStimulusMode() int {
	return int((n.Header.Status >> 2) & 0x03) // Bits 2-3
}

// GetStimulusModeDescription returns the human-readable stimulus mode description
func (n *NMTGroup) GetStimulusModeDescription() string {
	switch n.GetStimulusMode() {
	case TOF:
		return "Train Of Four (TOF mode)"
	case DBS:
		return "Double Burst (DB mode)"
	case ST_STIM:
		return "Single Twitch (ST mode)"
	case PTC_STIM:
		return "Post-tetanic count"
	default:
		return "Unknown stimulus mode"
	}
}

// GetPulseWidth returns the pulse width
func (n *NMTGroup) GetPulseWidth() int {
	return int((n.Header.Status >> 4) & 0x03) // Bits 4-5
}

// GetPulseWidthDescription returns the human-readable pulse width description
func (n *NMTGroup) GetPulseWidthDescription() string {
	switch n.GetPulseWidth() {
	case PULSE_NOT_USED:
		return "not used"
	case PULSE_100:
		return "100 us"
	case PULSE_200:
		return "200 us"
	case PULSE_300:
		return "300 us"
	default:
		return "Unknown pulse width"
	}
}

// IsSupramaxCurrentFound returns true if supramax current is found
func (n *NMTGroup) IsSupramaxCurrentFound() bool {
	return (n.Header.Status & 0x0040) != 0 // Bit 6
}

// IsCalibrated returns true if NMT is calibrated
func (n *NMTGroup) IsCalibrated() bool {
	return (n.Header.Status & 0x0080) != 0 // Bit 7
}

// GetPostTetanicCount returns the post tetanic count (bits 0-4)
func (n *NMTGroup) GetPostTetanicCount() int {
	return int(n.Ptc & 0x1F) // Bits 0-4
}

// GetTOFCount returns the TOF count (bits 5-8)
func (n *NMTGroup) GetTOFCount() int {
	return int((n.Ptc >> 5) & 0x0F) // Bits 5-8
}

// GetStimulusCurrent returns the stimulus current in mA (bits 9-15)
func (n *NMTGroup) GetStimulusCurrent() int {
	return int((n.Ptc >> 9) & 0x7F) // Bits 9-15
}

// ToJSON converts the NMTGroup to JSON format
func (n *NMTGroup) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"header": map[string]interface{}{
			"status": n.Header.Status,
			"label":  n.Header.Label,
			"stimulus_mode": map[string]interface{}{
				"value": n.GetStimulusMode(),
				"description": n.GetStimulusModeDescription(),
			},
			"pulse_width": map[string]interface{}{
				"value": n.GetPulseWidth(),
				"description": n.GetPulseWidthDescription(),
			},
			"is_supramax_current_found": n.IsSupramaxCurrentFound(),
			"is_calibrated": n.IsCalibrated(),
		},
		"t1": map[string]interface{}{
			"raw_value": n.T1,
			"value":     n.GetT1(),
			"unit":      "%",
		},
		"tratio": map[string]interface{}{
			"raw_value": n.Tratio,
			"value":     n.GetTratio(),
			"unit":      "%",
		},
		"ptc": map[string]interface{}{
			"raw_value": n.Ptc,
			"post_tetanic_count": n.GetPostTetanicCount(),
			"tof_count": n.GetTOFCount(),
			"stimulus_current": map[string]interface{}{
				"value": n.GetStimulusCurrent(),
				"unit":  "mA",
			},
		},
	}
}

// ECG Extra Group Structure
// Table 3-44 ECG Extra data fields
// C struct equivalent:
// struct ecg_extra_group {
//     short hr_ecg;
//     short hr_max;
//     short hr_min;
// };
type ECGExtraGroup struct {
	HrEcg int16 // Heart rate as derived from the ECG signal
	HrMax int16 // Maximum heart rate
	HrMin int16 // Minimum heart rate
}

// Size returns the size of ECGExtraGroup in bytes
func (e *ECGExtraGroup) Size() int {
	return 6 // 3 * 2 bytes (no header)
}

// UnmarshalBinary converts binary data to ECG Extra group
func (e *ECGExtraGroup) UnmarshalBinary(data []byte) error {
	if len(data) < e.Size() {
		return ErrInvalidDataLength
	}
	
	offset := 0
	e.HrEcg = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	e.HrMax = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	e.HrMin = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	return nil
}

// GetHeartRate returns the heart rate in bpm
func (e *ECGExtraGroup) GetHeartRate() float64 {
	return float64(e.HrEcg)
}

// GetMaxHeartRate returns the maximum heart rate in bpm
func (e *ECGExtraGroup) GetMaxHeartRate() float64 {
	return float64(e.HrMax)
}

// GetMinHeartRate returns the minimum heart rate in bpm
func (e *ECGExtraGroup) GetMinHeartRate() float64 {
	return float64(e.HrMin)
}

// ToJSON converts the ECGExtraGroup to JSON format
func (e *ECGExtraGroup) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"hr_ecg": map[string]interface{}{
			"raw_value": e.HrEcg,
			"value":     e.GetHeartRate(),
			"unit":      "bpm",
		},
		"hr_max": map[string]interface{}{
			"raw_value": e.HrMax,
			"value":     e.GetMaxHeartRate(),
			"unit":      "bpm",
		},
		"hr_min": map[string]interface{}{
			"raw_value": e.HrMin,
			"value":     e.GetMinHeartRate(),
			"unit":      "bpm",
		},
	}
}

// SvO2 Label Constants
// Table 3-46 SpO2 label values
const (
	DRI_SO2  = 0 // SO2
	DRI_SAO2 = 1 // SaO2
	DRI_SVO2 = 2 // SvO2
)

// SvO2 Group Structure
// Table 3-44 SvO2 data fields
// C struct equivalent:
// struct svo2_group {
//     struct group_hdr hdr;
//     short svo2;
// };
type SvO2Group struct {
	Header GroupHeader // Group header with status and label
	SvO2   int16      // SvO2 value
}

// Size returns the size of SvO2Group in bytes
func (s *SvO2Group) Size() int {
	return s.Header.Size() + 2 // header + 2 bytes
}

// UnmarshalBinary converts binary data to SvO2 group
func (s *SvO2Group) UnmarshalBinary(data []byte) error {
	if len(data) < s.Size() {
		return ErrInvalidDataLength
	}
	
	offset := 0
	if err := s.Header.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += s.Header.Size()
	
	s.SvO2 = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2
	
	return nil
}

// GetSvO2Value returns the SvO2 value
func (s *SvO2Group) GetSvO2Value() float64 {
	return float64(s.SvO2)
}

// GetSaturationType returns the saturation measurement type
func (s *SvO2Group) GetSaturationType() string {
	switch s.Header.Label {
	case DRI_SO2:
		return "SO2"
	case DRI_SAO2:
		return "SaO2"
	case DRI_SVO2:
		return "SvO2"
	default:
		return fmt.Sprintf("Unknown type %d", s.Header.Label)
	}
}

// IsCalibratedOver24hAgo returns true if calibrated over 24h ago
func (s *SvO2Group) IsCalibratedOver24hAgo() bool {
	return (s.Header.Status & 0x0004) != 0 // Bit 2
}

// IsFaultyCable returns true if cable is faulty
func (s *SvO2Group) IsFaultyCable() bool {
	return (s.Header.Status & 0x0008) != 0 // Bit 3
}

// IsNoCable returns true if no cable is connected
func (s *SvO2Group) IsNoCable() bool {
	return (s.Header.Status & 0x0010) != 0 // Bit 4
}

// IsNotCalibrated returns true if not calibrated
func (s *SvO2Group) IsNotCalibrated() bool {
	return (s.Header.Status & 0x0020) != 0 // Bit 5
}

// IsRecalibrated returns true if re-calibrated
func (s *SvO2Group) IsRecalibrated() bool {
	return (s.Header.Status & 0x0040) != 0 // Bit 6
}

// IsSvO2OutOfRange returns true if SvO2 is out of range
func (s *SvO2Group) IsSvO2OutOfRange() bool {
	return (s.Header.Status & 0x0080) != 0 // Bit 7
}

// IsCheckCatheterPosition returns true if catheter position should be checked
func (s *SvO2Group) IsCheckCatheterPosition() bool {
	return (s.Header.Status & 0x0100) != 0 // Bit 8
}

// IsIntensityShift returns true if there is an intensity shift
func (s *SvO2Group) IsIntensityShift() bool {
	return (s.Header.Status & 0x0200) != 0 // Bit 9
}

// ToJSON converts the SvO2Group to JSON format
func (s *SvO2Group) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"header": map[string]interface{}{
			"status": s.Header.Status,
			"label":  s.Header.Label,
			"saturation_type": s.GetSaturationType(),
			"status_bits": map[string]interface{}{
				"calibrated_over_24h_ago": s.IsCalibratedOver24hAgo(),
				"faulty_cable": s.IsFaultyCable(),
				"no_cable": s.IsNoCable(),
				"not_calibrated": s.IsNotCalibrated(),
				"recalibrated": s.IsRecalibrated(),
				"svo2_out_of_range": s.IsSvO2OutOfRange(),
				"check_catheter_position": s.IsCheckCatheterPosition(),
				"intensity_shift": s.IsIntensityShift(),
			},
		},
		"svo2": map[string]interface{}{
			"raw_value": s.SvO2,
			"value":     s.GetSvO2Value(),
			"unit":      "%",
		},
	}
}

// DRI Alarm Subrecord Types
// Table 4-1 Alarm subrecord values and usage
const (
	DRI_AL_STATUS = 1 // Alarm status subrecord
)

// Alarm Silence Information Values
// Table 4-2 Alarm silence information values
const (
	DRI_SI_NONE     = 0 // Alarms are not silenced at bedside
	DRI_SI_APNEA    = 1 // Apnea alarms have been silenced at bedside
	DRI_SI_ASY      = 2 // Asystole alarms have been silenced at bedside
	DRI_SI_APNEA_ASY = 3 // Both apnea and asystole alarms have been silenced at bedside
	DRI_SI_ALL      = 4 // All alarms have been silenced at bedside
	DRI_SI_2MIN     = 5 // All alarms have been silenced at bedside for two minutes
	DRI_SI_5MIN     = 6 // All alarms have been silenced at bedside for five minutes
	DRI_SI_20S      = 7 // All alarms have been silenced at bedside for 20 seconds
)

// Alarm Color/Priority Constants
// enum dri_alarm_color
const (
	DRI_PR0 = 0 // No alarm
	DRI_PR1 = 1 // White
	DRI_PR2 = 2 // Yellow
	DRI_PR3 = 3 // Red
)

// Alarm Display Structure
// struct al_disp_al
type AlarmDisplay struct {
	Text         [80]byte // The actual alarm text displayed by the S/5 monitor
	TextChanged  bool     // Is true if the alarm text has changed after the previous alarm data transmission
	Color        byte     // The priority of the alarm (enum_dri_alarm_color)
	ColorChanged bool     // Is true if the alarm color has changed after the previous alarm data transmission
	Reserved     [6]int16 // Reserved for future extensions
}

// Size returns the size of AlarmDisplay in bytes
func (a *AlarmDisplay) Size() int {
	return 80 + 1 + 1 + 1 + 6*2 // text[80] + text_changed + color + color_changed + reserved[6]
}

// UnmarshalBinary converts binary data to alarm display
func (a *AlarmDisplay) UnmarshalBinary(data []byte) error {
	if len(data) < a.Size() {
		return ErrInvalidDataLength
	}

	offset := 0

	// text[]: The actual alarm text displayed by the S/5 monitor
	copy(a.Text[:], data[offset:offset+80])
	offset += 80

	// text_changed: Is true if the alarm text has changed
	a.TextChanged = data[offset] != 0
	offset += 1

	// color: The priority of the alarm
	a.Color = data[offset]
	offset += 1

	// color_changed: Is true if the alarm color has changed
	a.ColorChanged = data[offset] != 0
	offset += 1

	// reserved: Reserved for future extensions
	for i := 0; i < 6; i++ {
		a.Reserved[i] = int16(binary.LittleEndian.Uint16(data[offset:]))
		offset += 2
	}

	return nil
}

// GetAlarmText returns the alarm text as a string
func (a *AlarmDisplay) GetAlarmText() string {
	// Find the null terminator
	end := 0
	for i := 0; i < 80; i++ {
		if a.Text[i] == 0 {
			end = i
			break
		}
		end = i + 1
	}
	return string(a.Text[:end])
}

// SetAlarmText sets the alarm text
func (a *AlarmDisplay) SetAlarmText(text string) {
	// Clear the text array
	for i := 0; i < 80; i++ {
		a.Text[i] = 0
	}
	
	// Copy the text (truncate if longer than 80 characters)
	textBytes := []byte(text)
	copyLength := len(textBytes)
	if copyLength > 80 {
		copyLength = 80
	}
	copy(a.Text[:], textBytes[:copyLength])
}

// GetAlarmColor returns the alarm color as a string
func (a *AlarmDisplay) GetAlarmColor() string {
	switch a.Color {
	case DRI_PR0:
		return "No alarm"
	case DRI_PR1:
		return "White"
	case DRI_PR2:
		return "Yellow"
	case DRI_PR3:
		return "Red"
	default:
		return fmt.Sprintf("Unknown color %d", a.Color)
	}
}

// GetAlarmPriority returns the alarm priority level
func (a *AlarmDisplay) GetAlarmPriority() int {
	return int(a.Color)
}

// IsActiveAlarm returns true if this is an active alarm (priority 1 or higher)
func (a *AlarmDisplay) IsActiveAlarm() bool {
	return a.Color >= DRI_PR1
}

// ToJSON converts the AlarmDisplay to JSON format
func (a *AlarmDisplay) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"text": map[string]interface{}{
			"value":   a.GetAlarmText(),
			"changed": a.TextChanged,
		},
		"color": map[string]interface{}{
			"value":   a.Color,
			"name":    a.GetAlarmColor(),
			"changed": a.ColorChanged,
		},
		"priority": map[string]interface{}{
			"level": a.GetAlarmPriority(),
			"is_active": a.IsActiveAlarm(),
		},
		"reserved": a.Reserved,
	}
}

// Alarm Status Message Structure
// struct dri_al_msg
type AlarmStatusMessage struct {
	Reserved    int16           // Reserved for future extensions
	SoundOnOff  bool            // Indicates the on/off status of the alarm sound (0=off, 1=on)
	Reserved2   int16           // Reserved for future extensions
	Reserved3   int16           // Reserved for future extensions
	SilenceInfo byte            // Indicates the alarm silence status at the monitor
	AlDisp      [5]AlarmDisplay // Array of alarm messages (sorted in descending order by alarm color)
	Reserved4   [5]int16        // Reserved for future extensions
}

// Size returns the size of AlarmStatusMessage in bytes
func (a *AlarmStatusMessage) Size() int {
	return 2 + 1 + 2 + 2 + 1 + 5*93 + 5*2 // reserved + sound_on_off + reserved2 + reserved3 + silence_info + 5*al_disp + reserved4
}

// UnmarshalBinary converts binary data to alarm status message
func (a *AlarmStatusMessage) UnmarshalBinary(data []byte) error {
	if len(data) < a.Size() {
		return ErrInvalidDataLength
	}

	offset := 0

	// reserved: Reserved for future extensions
	a.Reserved = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2

	// sound_on_off: Indicates the on/off status of the alarm sound
	a.SoundOnOff = data[offset] != 0
	offset += 1

	// reserved2: Reserved for future extensions
	a.Reserved2 = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2

	// reserved3: Reserved for future extensions
	a.Reserved3 = int16(binary.LittleEndian.Uint16(data[offset:]))
	offset += 2

	// silence_info: Indicates the alarm silence status at the monitor
	a.SilenceInfo = data[offset]
	offset += 1

	// al_disp: Array of alarm messages
	for i := 0; i < 5; i++ {
		if err := a.AlDisp[i].UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		offset += a.AlDisp[i].Size()
	}

	// reserved4: Reserved for future extensions
	for i := 0; i < 5; i++ {
		a.Reserved4[i] = int16(binary.LittleEndian.Uint16(data[offset:]))
		offset += 2
	}

	return nil
}

// GetSilenceInfoDescription returns the human-readable silence info description
func (a *AlarmStatusMessage) GetSilenceInfoDescription() string {
	switch a.SilenceInfo {
	case DRI_SI_NONE:
		return "Alarms are not silenced at bedside"
	case DRI_SI_APNEA:
		return "Apnea alarms have been silenced at bedside"
	case DRI_SI_ASY:
		return "Asystole alarms have been silenced at bedside"
	case DRI_SI_APNEA_ASY:
		return "Both apnea and asystole alarms have been silenced at bedside"
	case DRI_SI_ALL:
		return "All alarms have been silenced at bedside"
	case DRI_SI_2MIN:
		return "All alarms have been silenced at bedside for two minutes"
	case DRI_SI_5MIN:
		return "All alarms have been silenced at bedside for five minutes"
	case DRI_SI_20S:
		return "All alarms have been silenced at bedside for 20 seconds"
	default:
		return fmt.Sprintf("Unknown silence info %d", a.SilenceInfo)
	}
}

// GetActiveAlarmCount returns the number of active alarms
func (a *AlarmStatusMessage) GetActiveAlarmCount() int {
	count := 0
	for i := 0; i < 5; i++ {
		if a.AlDisp[i].IsActiveAlarm() {
			count++
		}
	}
	return count
}

// GetHighestPriorityAlarm returns the highest priority alarm (lowest number = highest priority)
func (a *AlarmStatusMessage) GetHighestPriorityAlarm() *AlarmDisplay {
	highestPriority := 255
	var highestAlarm *AlarmDisplay

	for i := 0; i < 5; i++ {
		if a.AlDisp[i].IsActiveAlarm() && int(a.AlDisp[i].Color) < highestPriority {
			highestPriority = int(a.AlDisp[i].Color)
			highestAlarm = &a.AlDisp[i]
		}
	}

	return highestAlarm
}

// IsSoundOn returns true if the alarm sound is on
func (a *AlarmStatusMessage) IsSoundOn() bool {
	return a.SoundOnOff
}

// IsSilenced returns true if any alarms are silenced
func (a *AlarmStatusMessage) IsSilenced() bool {
	return a.SilenceInfo != DRI_SI_NONE
}

// ToJSON converts the AlarmStatusMessage to JSON format
func (a *AlarmStatusMessage) ToJSON() map[string]interface{} {
	alarms := make([]map[string]interface{}, 5)
	for i := 0; i < 5; i++ {
		alarms[i] = a.AlDisp[i].ToJSON()
	}

	return map[string]interface{}{
		"reserved": a.Reserved,
		"sound_on_off": map[string]interface{}{
			"value": a.SoundOnOff,
			"status": a.IsSoundOn(),
		},
		"reserved2": a.Reserved2,
		"reserved3": a.Reserved3,
		"silence_info": map[string]interface{}{
			"value": a.SilenceInfo,
			"description": a.GetSilenceInfoDescription(),
			"is_silenced": a.IsSilenced(),
		},
		"alarms": alarms,
		"active_alarm_count": a.GetActiveAlarmCount(),
		"highest_priority_alarm": func() interface{} {
			if highest := a.GetHighestPriorityAlarm(); highest != nil {
				return highest.ToJSON()
			}
			return nil
		}(),
		"reserved4": a.Reserved4,
	}
}

// Alarm Subrecords Union Structure
// union al_srcrds
type AlarmSubrecords struct {
	AlarmMsg *AlarmStatusMessage // struct dri_al_msg
}

// Size returns the size of AlarmSubrecords in bytes
func (a *AlarmSubrecords) Size() int {
	if a.AlarmMsg != nil {
		return a.AlarmMsg.Size()
	}
	return 0
}

// UnmarshalBinary converts binary data to alarm subrecords
func (a *AlarmSubrecords) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// For now, assume it's an alarm status message
	a.AlarmMsg = &AlarmStatusMessage{}
	return a.AlarmMsg.UnmarshalBinary(data)
}

// ToJSON converts the AlarmSubrecords to JSON format
func (a *AlarmSubrecords) ToJSON() map[string]interface{} {
	if a.AlarmMsg != nil {
		return map[string]interface{}{
			"type": "alarm_status_message",
			"data": a.AlarmMsg.ToJSON(),
		}
	}
	return map[string]interface{}{
		"type": "empty",
		"data": nil,
	}
}
