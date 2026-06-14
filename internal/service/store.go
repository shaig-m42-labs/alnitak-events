package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
create table if not exists event_records (
    event_id text primary key,
    event_type text not null,
    source text not null,
    occurred_at timestamptz not null,
    correlation_id text,
    payload jsonb not null,
    received_at timestamptz not null default now()
);

create table if not exists delivery_attempts (
    id bigserial primary key,
    event_id text not null references event_records(event_id),
    target_url text not null,
    attempt integer not null,
    status text not null,
    error text,
    created_at timestamptz not null default now()
);

create table if not exists dead_letter_events (
    event_id text primary key references event_records(event_id),
    reason text not null,
    created_at timestamptz not null default now()
);
`)
	return err
}

func (s *Store) SaveEvent(ctx context.Context, event EventEnvelope) (bool, error) {
	payload, _ := json.Marshal(event.Payload)
	result, err := s.db.ExecContext(ctx, `
insert into event_records(event_id, event_type, source, occurred_at, correlation_id, payload)
values ($1, $2, $3, $4, $5, $6)
on conflict (event_id) do nothing`,
		event.EventID, event.EventType, event.Source, event.OccurredAt, event.CorrelationID, string(payload))
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return true, nil
	}
	return rows > 0, nil
}

func (s *Store) SaveAttempt(ctx context.Context, eventID, targetURL string, attempt int, status, message string) error {
	_, err := s.db.ExecContext(ctx, `
insert into delivery_attempts(event_id, target_url, attempt, status, error)
values ($1, $2, $3, $4, $5)`, eventID, targetURL, attempt, status, message)
	return err
}

func (s *Store) DeadLetter(ctx context.Context, eventID, reason string) error {
	_, err := s.db.ExecContext(ctx, `
insert into dead_letter_events(event_id, reason)
values ($1, $2)
on conflict (event_id) do nothing`, eventID, reason)
	return err
}

func (s *Store) Events(ctx context.Context) ([]EventRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
select event_id, event_type, source, occurred_at, correlation_id, payload::text
from event_records
order by received_at desc
limit 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []EventRecord
	for rows.Next() {
		var event EventRecord
		if err := rows.Scan(&event.EventID, &event.EventType, &event.Source, &event.OccurredAt, &event.CorrelationID, &event.Payload); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) Deliveries(ctx context.Context) ([]DeliveryAttempt, error) {
	rows, err := s.db.QueryContext(ctx, `
select id, event_id, target_url, attempt, status, coalesce(error, ''), created_at
from delivery_attempts
order by created_at desc
limit 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var attempts []DeliveryAttempt
	for rows.Next() {
		var attempt DeliveryAttempt
		var created time.Time
		if err := rows.Scan(&attempt.ID, &attempt.EventID, &attempt.TargetURL, &attempt.Attempt, &attempt.Status, &attempt.Error, &created); err != nil {
			return nil, err
		}
		attempt.CreatedAt = created
		attempts = append(attempts, attempt)
	}
	return attempts, rows.Err()
}
