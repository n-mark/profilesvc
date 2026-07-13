package handlers

import (
	"net/http"
)

type Server struct {
	profileHandler *ProfileHandler
}

func NewServer(profileService ProfileService) *Server {
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
