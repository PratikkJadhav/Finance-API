// internal/middleware/role.go
package middleware

import (
	"net/http"
)

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(string)
			if !ok || role == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error":"forbidden: insufficient permissions"}`, http.StatusForbidden)
		})
	}
}

// helper — call these in handlers to extract values from context
func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(UserIDKey).(string)
	return id
}

func GetRole(r *http.Request) string {
	role, _ := r.Context().Value(RoleKey).(string)
	return role
}
