# alnitak-events

Event storage and simulated webhook delivery service for Orion Platform V1.

## Behavior

- Subscribes to `orion.>` NATS subjects.
- Stores canonical event envelopes.
- Creates simulated webhook delivery attempts.
- Dead-letters events after max failed attempts.

## Endpoints

- `GET /health`
- `GET /metrics`
- `GET /events`
- `GET /deliveries`
