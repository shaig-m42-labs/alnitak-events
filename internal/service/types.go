package service

import "time"

type EventEnvelope struct {
	EventID       string         `json:"eventId"`
	EventType     string         `json:"eventType"`
	Source        string         `json:"source"`
	OccurredAt    time.Time      `json:"occurredAt"`
	CorrelationID string         `json:"correlationId"`
	Payload       map[string]any `json:"payload"`
}

type EventRecord struct {
	EventID       string    `json:"eventId"`
	EventType     string    `json:"eventType"`
	Source        string    `json:"source"`
	OccurredAt    time.Time `json:"occurredAt"`
	CorrelationID string    `json:"correlationId"`
	Payload       string    `json:"payload"`
}

type DeliveryAttempt struct {
	ID        int64     `json:"id"`
	EventID   string    `json:"eventId"`
	TargetURL string    `json:"targetUrl"`
	Attempt   int       `json:"attempt"`
	Status    string    `json:"status"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"createdAt"`
}
