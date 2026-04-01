package repository

import (
	"context"

	"github.com/PratikkJadhav/Finance-API/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepo(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{
		db: db,
	}
}

func (r *TransactionRepo) Create(ctx context.Context, txn *model.Transaction) error {
	query := `
		INSERT INTO transactions (user_id , amount , type , category , description , date) 
		VALUES ($1 , $2 ,$3 , $4 , $5 , $6)
		RETURNING id , created_at , updated_at
	`

	return r.db.QueryRow(ctx, query, txn.UserID, txn.Amount, txn.Type, txn.Category, txn.Description, txn.Date).Scan(&txn.ID, &txn.CreatedAt, &txn.UpdatedAt)
}

func (r *TransactionRepo) GetDashboardSummary(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	query := `
	SELECT
		COALESCE(SUM(CASE WHEN type='income' THEN amount ELSE 0 END) , 0 ) AS total_income,
		COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END) , 0) AS net_balance,
		COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) , 0) AS total_expense
	FROM transactions
	WHERE user_id = $1 AND deleted_at is NULL
	`

	var income, expense, balance float64
	err := r.db.QueryRow(ctx, query, userID).Scan(&income, &balance, &expense)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_income":  income,
		"total_expense": expense,
		"net_balance":   balance,
	}, nil
}
