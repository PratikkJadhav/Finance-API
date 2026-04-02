package handler

import (
	"encoding/json"
	"net/http"

	"github.com/PratikkJadhav/Finance-API/internal/model"
	"github.com/PratikkJadhav/Finance-API/internal/repository"
)

type ShareHandler struct {
	shareRepo *repository.ShareRepo
	userRepo  *repository.UserRepo
}

func NewShareHandler(sr *repository.ShareRepo, ur *repository.UserRepo) *ShareHandler {
	return &ShareHandler{shareRepo: sr, userRepo: ur}
}

func (h *ShareHandler) ShareData(w http.ResponseWriter, r *http.Request) {
	// Get the logged-in user's ID from context (they are the owner)
	ownerID := r.Context().Value("user_id").(string)

	var req model.ShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Look up the user they want to share with by email
	targetUser, err := h.userRepo.GetByEmail(r.Context(), req.SharedWithEmail)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Save the permission
	err = h.shareRepo.GrantAccess(r.Context(), ownerID, targetUser.ID.String(), req.Permission)
	if err != nil {
		http.Error(w, "Failed to share access", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Access granted successfully"})
}
