package handlers

import (
	"net/http"
	"profile-svc/internal/messaging"
)

type Server struct {
	profileHandler *ProfileHandler
}

func NewServer(profileService ProfileService, broker messaging.Broker) *Server {
	profileHandler := NewProfileHandler(profileService);
	broker.RegisterConsumer(broker.GetBillingResponseDataSourceName(), profileHandler.HandleBillingSvcResponse)

	return &Server{
		profileHandler: NewProfileHandler(profileService),
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/profile", s.profileHandler.HandleProfile)
	mux.HandleFunc("/profile/list", s.profileHandler.List)

	return mux
}
