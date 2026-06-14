package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/m42-labs/alnitak-events/internal/config"
)

func TestEventEnvelopeShape(t *testing.T) {
	event := EventEnvelope{EventID: "evt-1", EventType: "incident.opened", Source: "test"}
	if event.EventID != "evt-1" || event.EventType == "" || event.Source == "" {
		t.Fatal("unexpected event envelope")
	}
}

func TestHandleEventRecordsDeliveredAttempt(t *testing.T) {
	store := &fakeStore{inserted: true}
	svc := testService(store, 3)

	err := svc.HandleEvent(context.Background(), EventEnvelope{
		EventID:    "evt-1",
		EventType:  "incident.opened",
		Source:     "test",
		OccurredAt: time.Now(),
		Payload:    map[string]any{},
	})
	if err != nil {
		t.Fatalf("handle event failed: %v", err)
	}
	if store.attempts != 1 || store.lastStatus != "delivered" {
		t.Fatalf("expected delivered attempt, attempts=%d status=%s", store.attempts, store.lastStatus)
	}
}

func TestHandleEventSkipsDuplicateEvent(t *testing.T) {
	store := &fakeStore{inserted: false}
	svc := testService(store, 3)

	err := svc.HandleEvent(context.Background(), EventEnvelope{EventID: "evt-duplicate", Payload: map[string]any{}})
	if err != nil {
		t.Fatalf("handle event failed: %v", err)
	}
	if store.attempts != 0 || store.deadLetters != 0 {
		t.Fatalf("duplicate event should not deliver, attempts=%d deadLetters=%d", store.attempts, store.deadLetters)
	}
}

func TestHandleEventDeadLettersAfterMaxRetries(t *testing.T) {
	store := &fakeStore{inserted: true}
	svc := testService(store, 2)

	err := svc.HandleEvent(context.Background(), EventEnvelope{
		EventID: "evt-fail",
		Payload: map[string]any{
			"webhookUrl": "https://example.test/fail",
		},
	})
	if err != nil {
		t.Fatalf("handle event failed: %v", err)
	}
	if store.attempts != 2 || store.deadLetters != 1 {
		t.Fatalf("expected retries and dead letter, attempts=%d deadLetters=%d", store.attempts, store.deadLetters)
	}
}

func testService(store eventStore, maxAttempts int) *Service {
	return &Service{
		cfg:   config.Config{MaxDeliveryAttempts: maxAttempts},
		log:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
		store: store,
	}
}

type fakeStore struct {
	inserted    bool
	attempts    int
	deadLetters int
	lastStatus  string
}

func (f *fakeStore) SaveEvent(context.Context, EventEnvelope) (bool, error) {
	return f.inserted, nil
}

func (f *fakeStore) SaveAttempt(_ context.Context, _, _ string, _ int, status, _ string) error {
	f.attempts++
	f.lastStatus = status
	return nil
}

func (f *fakeStore) DeadLetter(context.Context, string, string) error {
	f.deadLetters++
	return nil
}

func (f *fakeStore) Events(context.Context) ([]EventRecord, error) {
	return nil, nil
}

func (f *fakeStore) Deliveries(context.Context) ([]DeliveryAttempt, error) {
	return nil, nil
}
