package handlers

import (
	"context"
	"net/http"

	"profile-service/internal/models"
)

type ProfileService interface {
	GetOne(ctx context.Context, id int64) (models.GetProfileDTO, error)
	List(ctx context.Context, query models.QueryDTO) ([]models.GetProfileDTO, error)
	Create(ctx context.Context, ownerID int64, dto models.ProfileDTO) (models.GetProfileDTO, error)
	Update(ctx context.Context, ownerID int64, profileID int64, dto models.ProfileDTO) (models.GetProfileDTO, error)
	UpdateStatus(ctx context.Context, ownerID int64, status string) (bool, error)
	GetByOwner(ctx context.Context, ownerID int64) (models.GetProfileDTO, error)
	UpsertByOwner(ctx context.Context, ownerID int64, dto models.ProfileDTO) (models.GetProfileDTO, error)
}


type AuthMiddleware interface {
	RequireAuth(next http.Handler) http.Handler
}
