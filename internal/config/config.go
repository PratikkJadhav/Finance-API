package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

func getENV(key, fallback string) string {
	val := os.Getenv(key)

	if val == "" {
		if fallback == "" {
			log.Fatalf("env variable %s not set", key)
		}

		return fallback

	}

	return val
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("no .env file found")

	}

	return &Config{
		DatabaseURL: getENV("DATABASE_URL", ""),
		JWTSecret:   getENV("JWT_SECRET", ""),
		Port:        getENV("PORT", "8080"),
	}

}
