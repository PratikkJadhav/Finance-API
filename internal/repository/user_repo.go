package repository

import (
	"context"

	"github.com/PratikkJadhav/Finance-API/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (email , password , name , role)
	VALUES ($1 , $2 , $3 , $4)
	RETURNING id ,is_active , created_at , updated_at`

	err := r.db.QueryRow(ctx, query, user.Email, user.Password, user.Name, user.Role).Scan(&user.ID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	return err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id , email , password , name , role , is_active , created_at , updated_at
		FROM users
		WHERE email = $1
	`

	user := &model.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id , email , password , name , role , is_active , created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &model.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
