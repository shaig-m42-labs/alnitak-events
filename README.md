# alnitak-events
Event-driven messaging service for async workflows and integrations.

**Language:** `Go`
**Stack:** `Go, NATS JetStream, PostgreSQL optional, Redis optional.`

**Responsibilities:**
```
Consume domain events from NATS
Store delivery attempts
Deliver webhooks
Retry failed webhook deliveries
Move failed events to dead-letter
Expose event delivery history API
```
**Entites:**
```
EventEnvelope
WebhookEndpoint
DeliveryAttempt
RetryPolicy
DeadLetterEvent
```
*topics:*
```
retry
dead-letter
idempotency
at-least-once delivery
event envelope
correlation id
```
