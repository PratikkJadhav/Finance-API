// internal/handler/user_handler.go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/PratikkJadhav/Finance-API/internal/service"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Role == "" {
		writeError(w, http.StatusBadRequest, "role is required")
		return
	}

	if err := h.userService.UpdateRole(r.Context(), id, body.Role); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "role updated successfully",
	})
}

func (h *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.userService.UpdateStatus(r.Context(), id, body.IsActive); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "status updated successfully",
	})
}
