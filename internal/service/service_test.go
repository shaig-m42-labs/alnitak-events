package service

import "testing"

func TestEventEnvelopeShape(t *testing.T) {
	event := EventEnvelope{EventID: "evt-1", EventType: "incident.opened", Source: "test"}
	if event.EventID != "evt-1" || event.EventType == "" || event.Source == "" {
		t.Fatal("unexpected event envelope")
	}
}
