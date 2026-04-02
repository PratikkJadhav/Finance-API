// internal/handler/helpers.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func parseIntQuery(r *http.Request, key string, fallback int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func handleRepoError(w http.ResponseWriter, err error) {
	if err == pgx.ErrNoRows {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "internal server error")
}
