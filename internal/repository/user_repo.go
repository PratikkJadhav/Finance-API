// internal/repository/user_repo.go  (complete version)
package repository

import (
	"context"

	"github.com/PratikkJadhav/Finance-API/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, password, name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_active, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		user.Email, user.Password, user.Name, user.Role,
	).Scan(&user.ID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password, name, role, is_active, created_at, updated_at
		FROM users WHERE email = $1
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

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, email, password, name, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
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

func (r *UserRepo) GetAll(ctx context.Context) ([]*model.User, error) {
	query := `
		SELECT id, email, name, role, is_active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name,
			&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepo) UpdateRole(ctx context.Context, id string, role model.Role) error {
	query := `UPDATE users SET role = $1, updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, role, id)
	return err
}

func (r *UserRepo) UpdateStatus(ctx context.Context, id string, isActive bool) error {
	query := `UPDATE users SET is_active = $1, updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, isActive, id)
	return err
}
