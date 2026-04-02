// internal/repository/transaction_repo.go  (complete version)
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PratikkJadhav/Finance-API/internal/dto"
	"github.com/PratikkJadhav/Finance-API/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepo(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) Create(ctx context.Context, input dto.CreateTransactionInput) (*model.Transaction, error) {
	query := `
		INSERT INTO transactions (user_id, amount, type, category, description, date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, amount, type, category, description, date, created_at, updated_at
	`
	txn := &model.Transaction{}
	err := r.db.QueryRow(ctx, query,
		input.UserID, input.Amount, input.Type,
		input.Category, input.Description, input.Date,
	).Scan(
		&txn.ID, &txn.UserID, &txn.Amount, &txn.Type,
		&txn.Category, &txn.Description, &txn.Date,
		&txn.CreatedAt, &txn.UpdatedAt,
	)
	return txn, err
}

func (r *TransactionRepo) GetAll(ctx context.Context, filter dto.TransactionFilter) ([]*model.Transaction, int, error) {
	args := []interface{}{}
	conditions := []string{"deleted_at IS NULL"}
	argIdx := 1

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.From != "" {
		conditions = append(conditions, fmt.Sprintf("date >= $%d", argIdx))
		args = append(args, filter.From)
		argIdx++
	}
	if filter.To != "" {
		conditions = append(conditions, fmt.Sprintf("date <= $%d", argIdx))
		args = append(args, filter.To)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM transactions %s", where)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// pagination
	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, amount, type, category, description, date, created_at, updated_at
		FROM transactions
		%s
		ORDER BY date DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txns []*model.Transaction
	for rows.Next() {
		txn := &model.Transaction{}
		err := rows.Scan(
			&txn.ID, &txn.UserID, &txn.Amount, &txn.Type,
			&txn.Category, &txn.Description, &txn.Date,
			&txn.CreatedAt, &txn.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		txns = append(txns, txn)
	}

	return txns, total, nil
}

func (r *TransactionRepo) GetByID(ctx context.Context, id string) (*model.Transaction, error) {
	query := `
		SELECT id, user_id, amount, type, category, description, date, created_at, updated_at
		FROM transactions
		WHERE id = $1 AND deleted_at IS NULL
	`
	txn := &model.Transaction{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&txn.ID, &txn.UserID, &txn.Amount, &txn.Type,
		&txn.Category, &txn.Description, &txn.Date,
		&txn.CreatedAt, &txn.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

func (r *TransactionRepo) Update(ctx context.Context, id string, input dto.UpdateTransactionInput) (*model.Transaction, error) {
	query := `
		UPDATE transactions
		SET amount = $1, type = $2, category = $3, description = $4, date = $5, updated_at = now()
		WHERE id = $6 AND deleted_at IS NULL
		RETURNING id, user_id, amount, type, category, description, date, created_at, updated_at
	`
	txn := &model.Transaction{}
	err := r.db.QueryRow(ctx, query,
		input.Amount, input.Type, input.Category, input.Description, input.Date, id,
	).Scan(
		&txn.ID, &txn.UserID, &txn.Amount, &txn.Type,
		&txn.Category, &txn.Description, &txn.Date,
		&txn.CreatedAt, &txn.UpdatedAt,
	)
	return txn, err
}

func (r *TransactionRepo) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE transactions SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	return err
}

func (r *TransactionRepo) GetSummaryByUser(ctx context.Context, userID string) (map[string]interface{}, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income'  THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS total_expense,
			COALESCE(SUM(CASE WHEN type = 'income'  THEN amount ELSE -amount END), 0) AS net_balance
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	var income, expense, balance float64
	err := r.db.QueryRow(ctx, query, userID).Scan(&income, &expense, &balance)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_income":  income,
		"total_expense": expense,
		"net_balance":   balance,
	}, nil
}

func (r *TransactionRepo) GetGlobalSummary(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income'  THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS total_expense,
			COALESCE(SUM(CASE WHEN type = 'income'  THEN amount ELSE -amount END), 0) AS net_balance
		FROM transactions
		WHERE deleted_at IS NULL
	`
	var income, expense, balance float64
	err := r.db.QueryRow(ctx, query).Scan(&income, &expense, &balance)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_income":  income,
		"total_expense": expense,
		"net_balance":   balance,
	}, nil
}

func (r *TransactionRepo) GetRecent(ctx context.Context, userID string, role string) ([]*model.Transaction, error) {
	var query string
	var args []interface{}

	if role == "admin" {
		query = `
			SELECT id, user_id, amount, type, category, description, date, created_at, updated_at
			FROM transactions
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 10
		`
	} else {
		query = `
			SELECT id, user_id, amount, type, category, description, date, created_at, updated_at
			FROM transactions
			WHERE deleted_at IS NULL AND user_id = $1
			ORDER BY created_at DESC
			LIMIT 10
		`
		args = append(args, userID)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []*model.Transaction
	for rows.Next() {
		txn := &model.Transaction{}
		err := rows.Scan(
			&txn.ID, &txn.UserID, &txn.Amount, &txn.Type,
			&txn.Category, &txn.Description, &txn.Date,
			&txn.CreatedAt, &txn.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}
	return txns, nil
}

func (r *TransactionRepo) GetMonthlyTrends(ctx context.Context, userID string) (interface{}, error) {
	args := []interface{}{}
	userFilter := ""
	if userID != "" {
		userFilter = "AND user_id = $1"
		args = append(args, userID)
	}

	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(DATE_TRUNC('month', date), 'YYYY-MM') AS month,
			COALESCE(SUM(CASE WHEN type = 'income'  THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS expense
		FROM transactions
		WHERE deleted_at IS NULL %s
		GROUP BY DATE_TRUNC('month', date)
		ORDER BY DATE_TRUNC('month', date) DESC
		LIMIT 12
	`, userFilter)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type MonthlyTrend struct {
		Month   string  `json:"month"`
		Income  float64 `json:"income"`
		Expense float64 `json:"expense"`
	}

	var trends []MonthlyTrend
	for rows.Next() {
		var t MonthlyTrend
		if err := rows.Scan(&t.Month, &t.Income, &t.Expense); err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}
	return trends, nil
}

func (r *TransactionRepo) GetCategoryTotals(ctx context.Context, userID string) (interface{}, error) {
	args := []interface{}{}
	userFilter := ""
	if userID != "" {
		userFilter = "AND user_id = $1"
		args = append(args, userID)
	}

	query := fmt.Sprintf(`
		SELECT
			category,
			type,
			COALESCE(SUM(amount), 0) AS total
		FROM transactions
		WHERE deleted_at IS NULL %s
		GROUP BY category, type
		ORDER BY total DESC
	`, userFilter)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type CategoryTotal struct {
		Category string  `json:"category"`
		Type     string  `json:"type"`
		Total    float64 `json:"total"`
	}

	var totals []CategoryTotal
	for rows.Next() {
		var c CategoryTotal
		if err := rows.Scan(&c.Category, &c.Type, &c.Total); err != nil {
			return nil, err
		}
		totals = append(totals, c)
	}
	return totals, nil
}
