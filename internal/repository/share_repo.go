package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ShareRepo struct {
	db *pgxpool.Pool
}

func NewShareRepo(db *pgxpool.Pool) *ShareRepo {
	return &ShareRepo{db: db}
}

func (r *ShareRepo) GrantAccess(ctx context.Context, ownerID, sharedWithID string, permission string) error {
	query := `
		INSERT INTO shared_access (owner_id, shared_with_id, permission)
		VALUES ($1, $2, $3)
		ON CONFLICT (owner_id, shared_with_id) 
		DO UPDATE SET permission = EXCLUDED.permission;`

	_, err := r.db.Exec(ctx, query, ownerID, sharedWithID, permission)
	return err
}
