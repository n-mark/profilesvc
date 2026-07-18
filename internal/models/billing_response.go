package models

import "github.com/google/uuid"

// BillingResponse is what billing-svc publishes after processing a payment.
type BillingResponse struct {
	EventId   uuid.UUID `json:"event_id,omitempty"`
	EventType string `json:"event_type,omitempty"`
	UserId    int64 `json:"user_id"`
	Status    string `json:"status,omitempty"`
}