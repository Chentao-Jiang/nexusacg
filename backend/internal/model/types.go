package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringArray is a custom type for storing string slices in PostgreSQL JSONB columns.
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan StringArray")
	}
	return json.Unmarshal(bytes, s)
}
