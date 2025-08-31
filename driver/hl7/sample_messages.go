package hl7

import (
	"fmt"
	"time"
)

// SampleHL7Messages contains various sample HL7 messages for testing
type SampleHL7Messages struct{}

// NewSampleHL7Messages creates a new sample messages instance
func NewSampleHL7Messages() *SampleHL7Messages {
	return &SampleHL7Messages{}
}

// GetVitalSignsMessage returns a sample vital signs ORU message based on GE Healthcare format
func (s *SampleHL7Messages) GetVitalSignsMessage() string {
	// ORU^R01 - Vital Signs (GE Healthcare format)
	now := time.Now()
	timestamp := now.Format("20060102150405-0700")
	deviceID := "080019FFFE134535"
	
	message := fmt.Sprintf("MSH|^~\\&|VSP^%s^EUI-64|GE Healthcare|||%s||ORU^R01^ORU_R01|%s|P|2.6|||NE|AL||UNICODE UTF-8|||PCD_DEC_001^IHE PCD^1.3.6.1.4.1.19376.1.6.1.1.1^ISO\r"+
		"PID|||HED12^^^PID^MR||LAZY^KITTY^^^^^L|||\r"+
		"PV1||E|ICU^^79874\r"+
		"OBR|1|%s%s^VSP^%s^EUI-64|%s%s^VSP^%s^EUI-64|182777000^monitoring ofpatient^SCT|||%s\r"+
		"OBX|1||69965^MDC_DEV_MON_PHYSIO_MULTI_PARAM_MDS^MDC|1.0.0.0|||||||X\r"+
		"OBX|2||69854^MDC_DEV_METER_PRESS_BLD_VMD^MDC|1.13.0.0|||||||X\r"+
		"OBX|3||69855^MDC_DEV_METER_PRESS_BLD_CHAN^MDC|1.13.1.0|||||||X\r"+
		"OBX|4|NM|150033^MDC_PRESS_BLD_ART_SYS^MDC|1.13.1.1|120|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|5|NM|150034^MDC_PRESS_BLD_ART_DIA^MDC|1.13.1.2|80|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|6|NM|150035^MDC_PRESS_BLD_ART_MEAN^MDC|1.13.1.3|93|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|7|NM|149522^MDC_BLD_PULS_RATE_INV^MDC|1.13.1.4|72|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|8||69855^MDC_DEV_METER_PRESS_BLD_CHAN^MDC|1.13.2.0|||||||X\r"+
		"OBX|9|NM|150087^MDC_PRESS_BLD_VEN_CENT_MEAN^MDC|1.13.2.1|8|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|10||69798^MDC_DEV_ECG_VMD^MDC|1.5.0.0|||||||X\r"+
		"OBX|11|NM|147842^MDC_ECG_HEART_RATE^MDC|1.5.1.1|75|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|12|NM|148066^MDC_ECG_V_P_C_RATE^MDC|1.5.1.2|2|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|13||69902^MDC_DEV_METER_TEMP_VMD^MDC|1.26.0.0|||||||X\r"+
		"OBX|14||69903^MDC_DEV_METER_TEMP_CHAN^MDC|1.26.1.0|||||||X\r"+
		"OBX|15|NM|150344^MDC_TEMP^MDC|1.26.1.1|36.8|268192^MDC_DIM_DEGC^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|16||69903^MDC_DEV_METER_TEMP_CHAN^MDC|1.26.2.0|||||||X\r"+
		"OBX|17|NM|150344^MDC_TEMP^MDC|1.26.2.1|36.9|268192^MDC_DIM_DEGC^MDC|||||R|||||||%s^B1X5_GE",
		deviceID, timestamp, deviceID+now.Format("20060102150405"),
		deviceID, now.Format("20060102150405"), deviceID, deviceID, now.Format("20060102150405"), deviceID,
		now.Format("20060102150405"),
		deviceID, deviceID, deviceID, deviceID, deviceID, deviceID, deviceID, deviceID, deviceID)
	
	return s.addMLLPWrapper(message)
}

// GetSpO2Message returns a sample SpO2 ORU message
func (s *SampleHL7Messages) GetSpO2Message() string {
	// ORU^R01 - SpO2 Monitoring
	now := time.Now()
	timestamp := now.Format("20060102150405-0700")
	deviceID := "080019FFFE134535"
	
	message := fmt.Sprintf("MSH|^~\\&|VSP^%s^EUI-64|GE Healthcare|||%s||ORU^R01^ORU_R01|%s|P|2.6|||NE|AL||UNICODE UTF-8|||PCD_DEC_001^IHE PCD^1.3.6.1.4.1.19376.1.6.1.1.1^ISO\r"+
		"PID|||HED12^^^PID^MR||LAZY^KITTY^^^^^L|||\r"+
		"PV1||E|ICU^^79874\r"+
		"OBR|1|%s%s^VSP^%s^EUI-64|%s%s^VSP^%s^EUI-64|182777000^monitoring ofpatient^SCT|||%s\r"+
		"OBX|1||69965^MDC_DEV_MON_PHYSIO_MULTI_PARAM_MDS^MDC|1.0.0.0|||||||X\r"+
		"OBX|2||69798^MDC_DEV_PULS_OXIM_VMD^MDC|1.6.0.0|||||||X\r"+
		"OBX|3||69799^MDC_DEV_PULS_OXIM_CHAN^MDC|1.6.1.0|||||||X\r"+
		"OBX|4|NM|150456^MDC_PULS_OXIM_SAT_O2^MDC|1.6.1.1|98|262688^MDC_DIM_PERCENT^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|5|NM|150457^MDC_PULS_OXIM_PULS_RATE^MDC|1.6.1.2|76|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|6|NM|150458^MDC_PULS_OXIM_PERF_INDEX^MDC|1.6.1.3|2.1|262688^MDC_DIM_PERCENT^MDC|||||R|||||||%s^B1X5_GE",
		deviceID, timestamp, deviceID+now.Format("20060102150405"),
		deviceID, now.Format("20060102150405"), deviceID, deviceID, now.Format("20060102150405"), deviceID,
		now.Format("20060102150405"),
		deviceID, deviceID, deviceID)
	
	return s.addMLLPWrapper(message)
}

// GetECGMessage returns a sample ECG ORU message
func (s *SampleHL7Messages) GetECGMessage() string {
	// ORU^R01 - ECG Monitoring
	now := time.Now()
	timestamp := now.Format("20060102150405-0700")
	deviceID := "080019FFFE134535"
	
	message := fmt.Sprintf("MSH|^~\\&|VSP^%s^EUI-64|GE Healthcare|||%s||ORU^R01^ORU_R01|%s|P|2.6|||NE|AL||UNICODE UTF-8|||PCD_DEC_001^IHE PCD^1.3.6.1.4.1.19376.1.6.1.1.1^ISO\r"+
		"PID|||HED12^^^PID^MR||LAZY^KITTY^^^^^L|||\r"+
		"PV1||E|ICU^^79874\r"+
		"OBR|1|%s%s^VSP^%s^EUI-64|%s%s^VSP^%s^EUI-64|182777000^monitoring ofpatient^SCT|||%s\r"+
		"OBX|1||69965^MDC_DEV_MON_PHYSIO_MULTI_PARAM_MDS^MDC|1.0.0.0|||||||X\r"+
		"OBX|2||69798^MDC_DEV_ECG_VMD^MDC|1.5.0.0|||||||X\r"+
		"OBX|3||69799^MDC_DEV_ECG_CHAN^MDC|1.5.1.0|||||||X\r"+
		"OBX|4|NM|147842^MDC_ECG_HEART_RATE^MDC|1.5.1.1|72|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|5|NM|148066^MDC_ECG_V_P_C_RATE^MDC|1.5.1.2|0|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|6|NM|147844^MDC_ECG_RESP_RATE^MDC|1.5.1.3|16|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|7|NM|147845^MDC_ECG_RESP_RATE_SPONT^MDC|1.5.1.4|16|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|8|NM|147846^MDC_ECG_RESP_RATE_MECH^MDC|1.5.1.5|0|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE",
		deviceID, timestamp, deviceID+now.Format("20060102150405"),
		deviceID, now.Format("20060102150405"), deviceID, deviceID, now.Format("20060102150405"), deviceID,
		now.Format("20060102150405"),
		deviceID, deviceID, deviceID, deviceID, deviceID)
	
	return s.addMLLPWrapper(message)
}

// GetCO2Message returns a sample CO2 ORU message
func (s *SampleHL7Messages) GetCO2Message() string {
	// ORU^R01 - CO2 Monitoring
	now := time.Now()
	timestamp := now.Format("20060102150405-0700")
	deviceID := "080019FFFE134535"
	
	message := fmt.Sprintf("MSH|^~\\&|VSP^%s^EUI-64|GE Healthcare|||%s||ORU^R01^ORU_R01|%s|P|2.6|||NE|AL||UNICODE UTF-8|||PCD_DEC_001^IHE PCD^1.3.6.1.4.1.19376.1.6.1.1.1^ISO\r"+
		"PID|||HED12^^^PID^MR||LAZY^KITTY^^^^^L|||\r"+
		"PV1||E|ICU^^79874\r"+
		"OBR|1|%s%s^VSP^%s^EUI-64|%s%s^VSP^%s^EUI-64|182777000^monitoring ofpatient^SCT|||%s\r"+
		"OBX|1||69965^MDC_DEV_MON_PHYSIO_MULTI_PARAM_MDS^MDC|1.0.0.0|||||||X\r"+
		"OBX|2||69800^MDC_DEV_CO2_VMD^MDC|1.7.0.0|||||||X\r"+
		"OBX|3||69801^MDC_DEV_CO2_CHAN^MDC|1.7.1.0|||||||X\r"+
		"OBX|4|NM|150456^MDC_CO2_ET^MDC|1.7.1.1|35|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|5|NM|150457^MDC_CO2_INSP^MDC|1.7.1.2|0|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|6|NM|150458^MDC_CO2_RESP_RATE^MDC|1.7.1.3|12|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE",
		deviceID, timestamp, deviceID+now.Format("20060102150405"),
		deviceID, now.Format("20060102150405"), deviceID, deviceID, now.Format("20060102150405"), deviceID,
		now.Format("20060102150405"),
		deviceID, deviceID, deviceID)
	
	return s.addMLLPWrapper(message)
}

// GetComprehensiveMessage returns Example 2 with comprehensive monitoring data
func (s *SampleHL7Messages) GetComprehensiveMessage() string {
	// ORU^R01 - Comprehensive Monitoring (Example 2 format)
	now := time.Now()
	timestamp := now.Format("20060102150405+0900")
	deviceID := "080019FFFE0B4020"
	
	message := fmt.Sprintf("MSH|^~\\&|VSP^%s^EUI-64|GE Healthcare|||%s||ORU^R01^ORU_R01|000C290B4020|P|2.6|||NE|AL||UNICODE|||PCD_DEC_001^IHE PCD^1.3.6.1.4.1.19376.1.6.1.1.1^ISO\r"+
		"PID|||999999999^^^PID^MR||^^^^^^L|||\r"+
		"PV1||E|ICU^^79874\r"+
		"OBR|1|%s%s^VSP^%s^EUI-64|%s%s^VSP^%s^EUI-64|182777000^monitoring of patient^SCT|||%s\r"+
		"OBX|1||69965^MDC_DEV_MON_PHYSIO_MULTI_PARAM_MDS^MDC|1.0.0.0|||||||X\r"+
		"OBX|2||69854^MDC_DEV_METER_PRESS_BLD_VMD^MDC|1.13.0.0|||||||X\r"+
		"OBX|3||69855^MDC_DEV_METER_PRESS_BLD_CHAN^MDC|1.13.1.0|||||||X\r"+
		"OBX|4|NM|150033^MDC_PRESS_BLD_ART_SYS^MDC|1.13.1.1|112|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|5|NM|150034^MDC_PRESS_BLD_ART_DIA^MDC|1.13.1.2|76|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|6|NM|150035^MDC_PRESS_BLD_ART_MEAN^MDC|1.13.1.3|95|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|7|NM|149522^MDC_BLD_PULS_RATE_INV^MDC|1.13.1.4|80|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|8||69855^MDC_DEV_METER_PRESS_BLD_CHAN^MDC|1.13.2.0|||||||X\r"+
		"OBX|9|NM|150087^MDC_PRESS_BLD_VEN_CENT_MEAN^MDC|1.13.2.1|9|266016^MDC_DIM_MMHG^MDC|||||R|||||||%s^B1X5_GE\r"+
		"OBX|10||69798^MDC_DEV_ECG_VMD^MDC|1.5.0.0|||||||X\r"+
		"OBX|11|NM|147842^MDC_ECG_HEART_RATE^MDC|1.5.1.1|80|264864^MDC_DIM_BEAT_PER_MIN^MDC|||||R|||||||%s^B1X5_GE",
		deviceID, timestamp,
		deviceID, now.Format("20060102150405"), deviceID, deviceID, now.Format("20060102150405"), deviceID,
		now.Format("20060102150405"),
		deviceID, deviceID, deviceID, deviceID, deviceID, deviceID)
	
	return s.addMLLPWrapper(message)
}

// GetADTMessage returns a sample ADT (Admission, Discharge, Transfer) message
func (s *SampleHL7Messages) GetADTMessage() string {
	// ADT^A01 - Patient Admission
	message := fmt.Sprintf("MSH|^~\\&|HIS|HOSPITAL|HL7SERVER|HOSPITAL|%s||ADT^A01|MSG001|P|2.5\r"+
		"PID||12345^^^MRN||SMITH^JOHN^A||19800101|M|||123 MAIN ST^^ANYTOWN^CA^12345||555-1234\r"+
		"PV1||I|2000^2012^01||||123456^SMITH^JOHN^J^^^MD|123456^SMITH^JOHN^J^^^MD|||||||||||I|2000^01|01\r"+
		"DG1|1|I10|I50.9|HEART FAILURE|20240115\r"+
		"AL1|1|DA|PENICILLIN|SEVERE RASH",
		time.Now().Format("20060102150405"))
	
	return s.addMLLPWrapper(message)
}

// GetORUMessage returns a sample ORU (Observation Result) message
func (s *SampleHL7Messages) GetORUMessage() string {
	// ORU^R01 - Observation Result
	message := fmt.Sprintf("MSH|^~\\&|LAB|HOSPITAL|HL7SERVER|HOSPITAL|%s||ORU^R01|MSG002|P|2.5\r"+
		"PID||12345^^^MRN||SMITH^JOHN^A||19800101|M|||123 MAIN ST^^ANYTOWN^CA^12345||555-1234\r"+
		"OBR|1|LAB001||CBC^COMPLETE BLOOD COUNT|R|%s|||||||||||123456^SMITH^JOHN^J^^^MD\r"+
		"OBX|1|NM|WBC^WHITE BLOOD CELLS|1|7.5|K/uL|4.5-11.0|N|||F\r"+
		"OBX|2|NM|RBC^RED BLOOD CELLS|1|4.8|M/uL|4.5-5.9|N|||F\r"+
		"OBX|3|NM|HGB^HEMOGLOBIN|1|14.2|g/dL|13.5-17.5|N|||F\r"+
		"OBX|4|NM|HCT^HEMATOCRIT|1|42.5|%%|41.0-50.0|N|||F\r"+
		"OBX|5|NM|PLT^PLATELETS|1|250|K/uL|150-450|N|||F",
		time.Now().Format("20060102150405"),
		time.Now().Format("20060102150405"))
	
	return s.addMLLPWrapper(message)
}

// GetORMMessage returns a sample ORM (Order Message) message
func (s *SampleHL7Messages) GetORMMessage() string {
	// ORM^O01 - Order Message
	message := fmt.Sprintf("MSH|^~\\&|HIS|HOSPITAL|HL7SERVER|HOSPITAL|%s||ORM^O01|MSG003|P|2.5\r"+
		"PID||12345^^^MRN||SMITH^JOHN^A||19800101|M|||123 MAIN ST^^ANYTOWN^CA^12345||555-1234\r"+
		"ORC|NW|LAB001|||CM|%s|||||123456^SMITH^JOHN^J^^^MD\r"+
		"OBR|1|LAB001||CBC^COMPLETE BLOOD COUNT|R|%s|||||||||||123456^SMITH^JOHN^J^^^MD",
		time.Now().Format("20060102150405"),
		time.Now().Format("20060102150405"),
		time.Now().Format("20060102150405"))
	
	return s.addMLLPWrapper(message)
}

// GetDischargeMessage returns a sample ADT discharge message
func (s *SampleHL7Messages) GetDischargeMessage() string {
	// ADT^A03 - Patient Discharge
	message := fmt.Sprintf("MSH|^~\\&|HIS|HOSPITAL|HL7SERVER|HOSPITAL|%s||ADT^A03|MSG006|P|2.5\r"+
		"PID||12345^^^MRN||SMITH^JOHN^A||19800101|M|||123 MAIN ST^^ANYTOWN^CA^12345||555-1234\r"+
		"PV1||D|2000^2012^01||||123456^SMITH^JOHN^J^^^MD|123456^SMITH^JOHN^J^^^MD|||||||||||D|2000^01|01\r"+
		"DG1|1|I10|I50.9|HEART FAILURE|20240115",
		time.Now().Format("20060102150405"))
	
	return s.addMLLPWrapper(message)
}

// GetTransferMessage returns a sample ADT transfer message
func (s *SampleHL7Messages) GetTransferMessage() string {
	// ADT^A02 - Patient Transfer
	message := fmt.Sprintf("MSH|^~\\&|HIS|HOSPITAL|HL7SERVER|HOSPITAL|%s||ADT^A02|MSG007|P|2.5\r"+
		"PID||12345^^^MRN||SMITH^JOHN^A||19800101|M|||123 MAIN ST^^ANYTOWN^CA^12345||555-1234\r"+
		"PV1||T|2000^2012^02||||123456^SMITH^JOHN^J^^^MD|123456^SMITH^JOHN^J^^^MD|||||||||||T|2000^02|02",
		time.Now().Format("20060102150405"))
	
	return s.addMLLPWrapper(message)
}

// addMLLPWrapper adds MLLP framing to the HL7 message
func (s *SampleHL7Messages) addMLLPWrapper(message string) string {
	// MLLP wrapper: 0x0B (VT) + message + 0x1C (FS) + 0x0D (CR)
	return fmt.Sprintf("%c%s%c%c", 0x0B, message, 0x1C, 0x0D)
}

// GetAllSampleMessages returns all sample messages
func (s *SampleHL7Messages) GetAllSampleMessages() map[string]string {
	return map[string]string{
		"ORU_VitalSigns":      s.GetVitalSignsMessage(),
		"ORU_SpO2":            s.GetSpO2Message(),
		"ORU_ECG":             s.GetECGMessage(),
		"ORU_CO2":             s.GetCO2Message(),
		"ORU_Comprehensive":   s.GetComprehensiveMessage(),
		"ADT_Admission":       s.GetADTMessage(),
		"ORU_LabResults":      s.GetORUMessage(),
		"ORM_Order":           s.GetORMMessage(),
		"ADT_Discharge":       s.GetDischargeMessage(),
		"ADT_Transfer":        s.GetTransferMessage(),
	}
}

// GetMessageDescription returns a description of the message type
func (s *SampleHL7Messages) GetMessageDescription(messageType string) string {
	descriptions := map[string]string{
		"ORU_VitalSigns":    "Vital Signs Monitoring (GE Healthcare format)",
		"ORU_SpO2":          "SpO2 Monitoring (GE Healthcare format)",
		"ORU_ECG":           "ECG Monitoring (GE Healthcare format)",
		"ORU_CO2":           "CO2 Monitoring (GE Healthcare format)",
		"ORU_Comprehensive": "Comprehensive Monitoring with 12-lead ECG, Gas Analysis, EEG (Example 2)",
		"ADT_Admission":     "Patient Admission (ADT^A01)",
		"ORU_LabResults":    "Laboratory Results (ORU^R01)",
		"ORM_Order":         "Medical Order (ORM^O01)",
		"ADT_Discharge":     "Patient Discharge (ADT^A03)",
		"ADT_Transfer":      "Patient Transfer (ADT^A02)",
	}
	
	if desc, exists := descriptions[messageType]; exists {
		return desc
	}
	return "Unknown message type"
}
