package todogroup

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmptyName = errors.New("todo group name is required")

type TodoGroupRepository interface {
	Create(ctx context.Context, g *TodoGroup) error
	GetAllByUser(ctx context.Context, userID int64) ([]TodoGroup, error)
}

type postgresRepo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) TodoGroupRepository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(ctx context.Context, g *TodoGroup) error {
	if g.Name == "" {
		return ErrEmptyName
	}

	query := `
		INSERT INTO todo_groups (name, user_id, color)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	/*
		ม่ต้องปิด(defer rows.Close()) ถ้าใช้:
		db.QueryRow()
		pool.QueryRow()
	*/
	row := r.db.QueryRow(ctx, query, g.Name, g.UserID, g.Color)

	return row.Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)
}

func (r *postgresRepo) GetAllByUser(ctx context.Context, userID int64) ([]TodoGroup, error) {
	query := `
		SELECT id, name, color, user_id, created_at, updated_at
		FROM todo_groups
		WHERE user_id = $1
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []TodoGroup
	for rows.Next() {
		var g TodoGroup
		// .Scan() เพื่อ อ่านค่าของแถวนี้ แล้ว ใส่ลงในตัวแปร ที่เตรียมไว้
		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.Color,
			&g.UserID,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, err
		}

		groups = append(groups, g)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return groups, nil
}
