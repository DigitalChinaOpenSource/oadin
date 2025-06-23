package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"unsafe"
)

type StringList []string

func (s StringList) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type: %T", value)
	}
	return json.Unmarshal(bytes, s)
}

type Float32List []float32

// Value converts a Float32List to a value for SQL storage
func (f Float32List) Value() (driver.Value, error) {
	if len(f) == 0 {
		return "[]", nil // Return empty JSON array string instead of nil
	}

	// Use JSON for GORM compatibility - this is more reliable for database operations
	return json.Marshal(f)
}

// Scan converts a database value to a Float32List
func (f *Float32List) Scan(value interface{}) error {
	if value == nil {
		*f = make([]float32, 0)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type: %T", value)
	}

	if len(bytes) == 0 {
		*f = make([]float32, 0)
		return nil
	}

	// Try to unmarshal as JSON first (which is how GORM stores it)
	if err := json.Unmarshal(bytes, f); err == nil {
		return nil
	}

	// If JSON unmarshal failed, try to interpret as BLOB of raw float32 values
	// This is for compatibility with sqlite-vec extension searches
	if len(bytes)%4 == 0 {
		count := len(bytes) / 4
		result := make([]float32, count)
		for i := 0; i < count; i++ {
			// Convert 4 bytes to float32 in little-endian format
			var val float32
			bits := (*[4]byte)(unsafe.Pointer(&val))
			bits[0] = bytes[i*4]
			bits[1] = bytes[i*4+1]
			bits[2] = bytes[i*4+2]
			bits[3] = bytes[i*4+3]
			result[i] = val
		}
		*f = result
		return nil
	}

	return fmt.Errorf("cannot parse Float32List from value: %v", string(bytes))
}

type Int64List []int64

func (i Int64List) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (i *Int64List) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type: %T", value)
	}
	return json.Unmarshal(bytes, i)
}
