package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) UserRepository {
	return &repo{db: db}
}

func (r *repo) Create(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (first_name, last_name, email, password)
		VALUES ($1, $2, LOWER($3), $4)
		RETURNING id, created_at, updated_at
	`
	row := r.db.QueryRow(ctx, query, u.FirstName, u.LastName, u.Email, u.Password)
	return row.Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *repo) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, first_name, last_name, email, password, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1
	`
	var u User
	row := r.db.QueryRow(ctx, query, email)
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
