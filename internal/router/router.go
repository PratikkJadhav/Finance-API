// internal/router/router.go
package router

import (
	"github.com/PratikkJadhav/Finance-API/internal/handler"
	customMiddleware "github.com/PratikkJadhav/Finance-API/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	txnHandler *handler.TransactionHandler,
	jwtSecret string,
) *chi.Mux {
	r := chi.NewRouter()

	// global middleware
	r.Use(customMiddleware.Logger) // our custom logger
	r.Use(chiMiddleware.Recoverer) // panic recovery
	r.Use(chiMiddleware.RequestID) // attach request ID

	// public routes
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(jwtSecret))

		// user management — admin only
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.RequireRole("admin"))
			r.Get("/users", userHandler.ListUsers)
			r.Patch("/users/{id}/role", userHandler.UpdateRole)
			r.Patch("/users/{id}/status", userHandler.UpdateStatus)
		})

		// transactions
		r.Group(func(r chi.Router) {
			r.With(customMiddleware.RequireRole("viewer", "analyst", "admin")).Get("/transactions", txnHandler.List)
			r.With(customMiddleware.RequireRole("viewer", "analyst", "admin")).Get("/transactions/{id}", txnHandler.GetByID)
			r.With(customMiddleware.RequireRole("analyst", "admin")).Post("/transactions", txnHandler.Create)
			r.With(customMiddleware.RequireRole("analyst", "admin")).Put("/transactions/{id}", txnHandler.Update)
			r.With(customMiddleware.RequireRole("admin")).Delete("/transactions/{id}", txnHandler.Delete)
		})

		// dashboard
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.RequireRole("viewer", "analyst", "admin"))
			r.Get("/dashboard/summary", txnHandler.GetSummary)
			r.Get("/dashboard/recent", txnHandler.GetRecent)
			r.With(customMiddleware.RequireRole("analyst", "admin")).Get("/dashboard/trends", txnHandler.GetTrends)
			r.With(customMiddleware.RequireRole("analyst", "admin")).Get("/dashboard/categories", txnHandler.GetCategoryTotals)
		})
	})

	return r
}
