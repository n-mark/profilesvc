package models

import "github.com/google/uuid"

// ProfileCreatedEvent is published to billing-svc so it can create a billing
// account for the freshly created profile.
type ProfileCreatedEvent struct {
	EventId   uuid.UUID `json:"event_id"`
	EventType string    `json:"event_type"`
	UserId    int64     `json:"user_id"`
}
