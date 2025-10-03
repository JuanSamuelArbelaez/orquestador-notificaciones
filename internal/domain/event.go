package domain

import "time"
import "encoding/json"

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Source    string          `json:"source"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// transform el evento en JSON para un formato legible
func (e *Event) DecodePayload(v interface{}) error {
	return json.Unmarshal(e.Payload, v)
}
