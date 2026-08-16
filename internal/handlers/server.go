package handlers

import (
	"net/http"
	"profile-service/internal/messaging"
)

type Server struct {
	profileHandler *ProfileHandler
}

func NewServer(profileService ProfileService, broker messaging.Broker) *Server {
	profileHandler := NewProfileHandler(profileService);
	broker.RegisterConsumer(broker.GetBillingResponseSource(), profileHandler.HandleBillingSvcResponse)

	return &Server{
		profileHandler: NewProfileHandler(profileService),
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// Public API (versioned)
	mux.HandleFunc("/api/v1/profile", s.profileHandler.HandleProfile)
	mux.HandleFunc("/api/v1/profile/list", s.profileHandler.List)

	// Internal API (versioned) — for BFF / other services inside the cluster.
	mux.HandleFunc("GET /internal/v1/users/{id}", s.profileHandler.GetUserInternal)
	mux.HandleFunc("POST /internal/v1/users/batch", s.profileHandler.GetUsersBatchInternal)

	return mux
}
