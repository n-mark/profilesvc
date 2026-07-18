package services

import "profile-svc/internal/models"

type ResponseWriter interface {
	ReportProfileCreated(p models.ProfileCreatedEvent) error
}