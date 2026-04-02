// cmd/seed/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/PratikkJadhav/Finance-API/internal/config"
	"github.com/PratikkJadhav/Finance-API/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var categories = []string{
	"salary", "freelance", "investment",
	"food", "transport", "utilities",
	"entertainment", "healthcare", "shopping", "rent",
}

var txnTypes = []string{"income", "expense"}

func main() {
	cfg := config.Load()
	database := db.NewDatabase(cfg)
	defer database.Conn.Close()

	ctx := context.Background()

	// create seed users
	users := []struct {
		email string
		name  string
		role  string
	}{
		{"admin@finance.com", "Admin User", "admin"},
		{"analyst@finance.com", "Analyst User", "analyst"},
		{"viewer@finance.com", "Viewer User", "viewer"},
	}

	userIDs := []string{}

	for _, u := range users {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		var id string
		err := database.Conn.QueryRow(ctx, `
			INSERT INTO users (email, password, name, role)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
			RETURNING id
		`, u.email, string(hashed), u.name, u.role).Scan(&id)
		if err != nil {
			log.Printf("failed to seed user %s: %v", u.email, err)
			continue
		}
		userIDs = append(userIDs, id)
		log.Printf("seeded user: %s (role=%s)", u.email, u.role)
	}

	if len(userIDs) == 0 {
		log.Fatal("no users seeded, aborting")
	}

	// seed 1000 transactions
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := 0

	for i := 0; i < 1000; i++ {
		userID := userIDs[rng.Intn(len(userIDs))]
		txnType := txnTypes[rng.Intn(len(txnTypes))]
		category := categories[rng.Intn(len(categories))]

		// random amount between 10 and 5000
		amount := float64(rng.Intn(4990)+10) + rng.Float64()

		// random date in the last 12 months
		daysAgo := rng.Intn(365)
		date := time.Now().AddDate(0, 0, -daysAgo).Format("2006-01-02")

		description := fmt.Sprintf("%s transaction #%d", category, i+1)

		_, err := database.Conn.Exec(ctx, `
			INSERT INTO transactions (user_id, amount, type, category, description, date)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, userID, amount, txnType, category, description, date)

		if err != nil {
			log.Printf("failed to seed transaction %d: %v", i, err)
			continue
		}
		count++
	}

	log.Printf("seeded %d transactions successfully", count)
	log.Println("seed credentials:")
	log.Println("  admin@finance.com   / password123 (admin)")
	log.Println("  analyst@finance.com / password123 (analyst)")
	log.Println("  viewer@finance.com  / password123 (viewer)")
}
