package services

import "profile-service/internal/models"

type ResponseWriter interface {
	ReportProfileCreated(p models.ProfileCreatedEvent) error
}