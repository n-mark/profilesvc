package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"profile-service/internal/models"
	"profile-service/internal/store"
)

type ProfileHandler struct {
	service ProfileService
}

func NewProfileHandler(service ProfileService) *ProfileHandler {
	return &ProfileHandler{service: service}
}

func (h *ProfileHandler) HandleBillingSvcResponse(body []byte) (bool, error) {
	payload := models.BillingResponse{}
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Error("failed to unmarshal", "error", err)
		return false, err
	}

	_, err := h.service.UpdateStatus(context.Background(), payload.UserId, "CREATED")
	if err != nil {
		slog.Error("failed to update user profile status", "error", err)
		return false, err
	}

	return true, nil
}

// HandleProfile dispatches GET/PUT /profile based on the authenticated user (X-User-Id).
// The endpoint is protected by Traefik ForwardAuth: by the time the request lands here,
// the auth-service has already validated the JWT and injected X-User-Id.
func (h *ProfileHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := ownerIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		profile, err := h.service.GetByOwner(r.Context(), ownerID)
		if err != nil {
			if errors.Is(err, store.ErrProfileNotFound) {
				http.Error(w, "profile not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, profile)
	case http.MethodPut:
		var dto models.ProfileDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		profile, err := h.service.UpsertByOwner(r.Context(), ownerID, dto)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, profile)
	case http.MethodPost:
		var dto models.ProfileDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		profile, err := h.service.Create(r.Context(), ownerID, dto)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, profile)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ownerIDFromRequest extracts the user id injected by Traefik ForwardAuth.
// Traefik forwards the X-User-Id header from auth-service /validate response.
func ownerIDFromRequest(r *http.Request) (int64, bool) {
	raw := r.Header.Get("X-User-Id")
	if raw == "" {
		raw = r.Header.Get("X-User-ID")
	}
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	profileIDStr := r.URL.Query().Get("profile_id")
	if profileIDStr == "" {
		http.Error(w, "profile_id is required", http.StatusBadRequest)
		return
	}

	profileID, err := (strconv.Atoi(profileIDStr))
	if err != nil {
		http.Error(w, "invalid profile_id", http.StatusBadRequest)
		return
	}

	profile, err := h.service.GetOne(r.Context(), int64(profileID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

func (h *ProfileHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var query models.QueryDTO
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	profiles, err := h.service.List(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, profiles)
}

func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var dto models.ProfileDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	profile, err := h.service.Create(r.Context(), int64(ownerID), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, profile)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	profileIDStr := r.URL.Query().Get("profile_id")
	if profileIDStr == "" {
		http.Error(w, "profile_id is required", http.StatusBadRequest)
		return
	}

	profileID, err := strconv.Atoi(profileIDStr)
	if err != nil {
		http.Error(w, "invalid profile_id", http.StatusBadRequest)
		return
	}

	var dto models.ProfileDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	profile, err := h.service.Update(r.Context(), int64(ownerID), int64(profileID), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}
