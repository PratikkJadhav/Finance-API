package db

import (
	"context"
	"log"

	"github.com/PratikkJadhav/Finance-API/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Conn *pgxpool.Pool
}

func NewDatabase(cfg *config.Config) *Database {
	pool, err := NewPgConnectionPool(cfg)
	if err != nil {
		log.Fatalf("failed to create connection pool: %v", err)
	}
	return &Database{
		Conn: pool,
	}
}

func NewPgConnectionPool(cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Println("database connected successfully")
	return pool, nil
}
