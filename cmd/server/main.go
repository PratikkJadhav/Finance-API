package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PratikkJadhav/Finance-API/internal/config"
	"github.com/PratikkJadhav/Finance-API/internal/db"
	"github.com/PratikkJadhav/Finance-API/internal/handler"
	"github.com/PratikkJadhav/Finance-API/internal/repository"
	"github.com/PratikkJadhav/Finance-API/internal/router"
	"github.com/PratikkJadhav/Finance-API/internal/service"
)

func main() {
	// 1. load config
	cfg := config.Load()

	// 2. connect db
	database := db.NewDatabase(cfg)
	defer database.Conn.Close()

	// 3. repositories
	userRepo := repository.NewUserRepo(database.Conn)
	txnRepo := repository.NewTransactionRepo(database.Conn)
	shareRepo := repository.NewShareRepo(database.Conn)

	// 4. services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	txnService := service.NewTransactionService(txnRepo)

	// 5. handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	txnHandler := handler.NewTransactionHandler(txnService)
	shareHandler := handler.NewShareHandler(shareRepo, userRepo)

	// 6. router
	r := router.NewRouter(authHandler, userHandler, txnHandler, shareHandler, cfg.JWTSecret)

	// 7. start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("server running on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
