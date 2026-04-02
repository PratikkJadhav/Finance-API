package router

import (
	"github.com/PratikkJadhav/Finance-API/internal/handler"
	customMiddleware "github.com/PratikkJadhav/Finance-API/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	txnHandler *handler.TransactionHandler,
	shareHandler *handler.ShareHandler, // <--- 1. ADD THIS HERE
	jwtSecret string,
) *chi.Mux {
	r := chi.NewRouter()

	// Handle CORS so your HTML file doesn't get blocked
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

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

		// --- 2. ADD THIS NEW SHARE ROUTE ---
		// We let any logged-in user share their own data
		r.Post("/share", shareHandler.ShareData)
		// -----------------------------------

		// user management — admin only
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.RequireRole("admin"))
			r.Get("/users", userHandler.ListUsers)
			r.Patch("/users/{id}/role", userHandler.UpdateRole)
			r.Patch("/users/{id}/status", userHandler.UpdateStatus)
		})

		// transactions
		r.Group(func(r chi.Router) {
			r.With(customMiddleware.RequireRole("viewers", "analyst", "admin")).Get("/transactions", txnHandler.List)
			r.With(customMiddleware.RequireRole("viewers", "analyst", "admin")).Get("/transactions/{id}", txnHandler.GetByID)
			r.With(customMiddleware.RequireRole("analyst", "admin")).Post("/transactions", txnHandler.Create)
			r.With(customMiddleware.RequireRole("analyst", "admin")).Put("/transactions/{id}", txnHandler.Update)
			r.With(customMiddleware.RequireRole("admin")).Delete("/transactions/{id}", txnHandler.Delete)
		})

		// dashboard
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.RequireRole("viewers", "analyst", "admin"))
			r.Get("/dashboard/summary", txnHandler.GetSummary)
			r.Get("/dashboard/recent", txnHandler.GetRecent)
			r.With(customMiddleware.RequireRole("analyst", "admin")).Get("/dashboard/trends", txnHandler.GetTrends)
			r.With(customMiddleware.RequireRole("analyst", "admin")).Get("/dashboard/categories", txnHandler.GetCategoryTotals)
		})
	})

	return r
}
