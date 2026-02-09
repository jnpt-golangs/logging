package logger

import (
	"strings"
)

// MaskingUtil provides utilities for masking sensitive data
type MaskingUtil struct {
	maskedHeaders []string
	maskedFields  []string
	maskValue     string
}

// NewMaskingUtil creates a new MaskingUtil
func NewMaskingUtil(config *Config) *MaskingUtil {
	return &MaskingUtil{
		maskedHeaders: config.MaskedHeaders,
		maskedFields:  config.MaskedFields,
		maskValue:     config.MaskValue,
	}
}

// MaskHeaders masks sensitive headers
func (m *MaskingUtil) MaskHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	
	masked := make(map[string]string)
	for key, value := range headers {
		if m.shouldMaskHeader(key) {
			masked[key] = m.maskValue
		} else {
			masked[key] = value
		}
	}
	return masked
}

// shouldMaskHeader checks if header should be masked
func (m *MaskingUtil) shouldMaskHeader(headerName string) bool {
	for _, masked := range m.maskedHeaders {
		if strings.EqualFold(masked, headerName) {
			return true
		}
	}
	return false
}

// MaskBody masks sensitive fields in body
func (m *MaskingUtil) MaskBody(body interface{}) interface{} {
	if body == nil {
		return nil
	}

	switch v := body.(type) {
	case map[string]interface{}:
		return m.maskMap(v)
	case []interface{}:
		return m.maskSlice(v)
	default:
		return body
	}
}

func (m *MaskingUtil) maskMap(data map[string]interface{}) map[string]interface{} {
	masked := make(map[string]interface{})
	for key, value := range data {
		if m.shouldMaskField(key) {
			masked[key] = m.maskValue
		} else {
			switch v := value.(type) {
			case map[string]interface{}:
				masked[key] = m.maskMap(v)
			case []interface{}:
				masked[key] = m.maskSlice(v)
			default:
				masked[key] = value
			}
		}
	}
	return masked
}

func (m *MaskingUtil) maskSlice(data []interface{}) []interface{} {
	masked := make([]interface{}, len(data))
	for i, item := range data {
		switch v := item.(type) {
		case map[string]interface{}:
			masked[i] = m.maskMap(v)
		case []interface{}:
			masked[i] = m.maskSlice(v)
		default:
			masked[i] = item
		}
	}
	return masked
}

func (m *MaskingUtil) shouldMaskField(fieldName string) bool {
	for _, masked := range m.maskedFields {
		if strings.EqualFold(masked, fieldName) {
			return true
		}
	}
	return false
}
