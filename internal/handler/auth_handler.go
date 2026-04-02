// internal/handler/auth_handler.go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/PratikkJadhav/Finance-API/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Email == "" || input.Password == "" || input.Name == "" {
		writeError(w, http.StatusBadRequest, "email, password and name are required")
		return
	}

	user, err := h.authService.Register(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "user registered successfully",
		"user":    user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Email == "" || input.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	resp, err := h.authService.Login(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
