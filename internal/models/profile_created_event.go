package models

import "github.com/google/uuid"

// OrderCreatedEvent is published to billing-svc to request payment.
type ProfileCreatedEvent struct {
	EventId    uuid.UUID `json:"event_id"`
	EventType  string    `json:"event_type"`
	OrderId    uuid.UUID `json:"order_id"`
	UserId     int64     `json:"user_id"`
	ToWithdraw float64   `json:"to_withdraw"`
}