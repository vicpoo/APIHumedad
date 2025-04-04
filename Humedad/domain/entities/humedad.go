// humedad.go
package entities

import "time"

type Temperature struct {
	ID        int       `json:"id"`
	Value     float64   `json:"temp"`
	Unit      string    `json:"unit"`
	DeviceID  string    `json:"device"`
	Timestamp int64     `json:"ts"`
	CreatedAt time.Time `json:"created_at"`
}

func NewTemperature(value float64, unit, deviceID string, timestamp int64) *Temperature {
	return &Temperature{
		Value:     value,
		Unit:      unit,
		DeviceID:  deviceID,
		Timestamp: timestamp,
		CreatedAt: time.Now(),
	}
}
