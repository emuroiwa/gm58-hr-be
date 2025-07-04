// First, create pkg/types/date.go
package types

import (
	"encoding/json"
	"strings"
	"time"
)

// CustomDate handles both "YYYY-MM-DD" and RFC3339 formats
type CustomDate struct {
	time.Time
}

func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")

	if s == "" || s == "null" {
		cd.Time = time.Time{}
		return nil
	}

	// Try date-only format first
	if t, err := time.Parse("2006-01-02", s); err == nil {
		cd.Time = t
		return nil
	}

	// Try RFC3339 format
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		cd.Time = t
		return nil
	}

	return &time.ParseError{Layout: "2006-01-02", Value: s}
}

func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if cd.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(cd.Time.Format("2006-01-02"))
}
