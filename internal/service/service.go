package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strings"

	_ "github.com/lib/pq"
	"github.com/m42-labs/alnitak-events/internal/config"
	"github.com/nats-io/nats.go"
)

type Service struct {
	cfg   config.Config
	log   *slog.Logger
	db    *sql.DB
	store *Store
	nats  *nats.Conn
}

func New(cfg config.Config, log *slog.Logger) (*Service, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	store := NewStore(db)
	if err := store.Init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Warn("nats_connect_failed", "error", err.Error())
	}

	return &Service{cfg: cfg, log: log, db: db, store: store, nats: nc}, nil
}

func (s *Service) Start(ctx context.Context) error {
	if s.nats == nil {
		return nil
	}
	_, err := s.nats.Subscribe("orion.>", func(msg *nats.Msg) {
		var envelope EventEnvelope
		if err := json.Unmarshal(msg.Data, &envelope); err != nil {
			s.log.Warn("event_decode_failed", "error", err.Error())
			return
		}
		if envelope.EventID == "" {
			s.log.Warn("event_missing_id", "subject", msg.Subject)
			return
		}
		if err := s.HandleEvent(ctx, envelope); err != nil {
			s.log.Warn("event_handle_failed", "eventId", envelope.EventID, "error", err.Error())
		}
	})
	return err
}

func (s *Service) HandleEvent(ctx context.Context, event EventEnvelope) error {
	if err := s.store.SaveEvent(ctx, event); err != nil {
		return err
	}
	targetURL := "simulated://default-webhook"
	if raw, ok := event.Payload["webhookUrl"].(string); ok && raw != "" {
		targetURL = raw
	}
	for attempt := 1; attempt <= s.cfg.MaxDeliveryAttempts; attempt++ {
		if strings.Contains(targetURL, "fail") {
			_ = s.store.SaveAttempt(ctx, event.EventID, targetURL, attempt, "failed", "simulated failure")
			continue
		}
		return s.store.SaveAttempt(ctx, event.EventID, targetURL, attempt, "delivered", "")
	}
	return s.store.DeadLetter(ctx, event.EventID, "max delivery attempts exceeded")
}

func (s *Service) Events(ctx context.Context) ([]EventRecord, error) {
	return s.store.Events(ctx)
}

func (s *Service) Deliveries(ctx context.Context) ([]DeliveryAttempt, error) {
	return s.store.Deliveries(ctx)
}

func (s *Service) Close() {
	if s.nats != nil {
		s.nats.Close()
	}
	if s.db != nil {
		_ = s.db.Close()
	}
}
