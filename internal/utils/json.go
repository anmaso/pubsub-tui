package utils

import (
	"bytes"
	"encoding/json"
)

// FormatJSON formats JSON data with indentation
func FormatJSON(data []byte) (string, error) {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		// If it's not valid JSON, return as-is
		return string(data), nil
	}
	return out.String(), nil
}

// IsValidJSON checks if data is valid JSON
func IsValidJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

// CompactJSON compacts JSON data by removing whitespace
func CompactJSON(data []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Compact(&out, data)
	if err != nil {
		return data, err
	}
	return out.Bytes(), nil
}

// PrettyPrint formats any value as indented JSON
func PrettyPrint(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
