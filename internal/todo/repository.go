package todo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ================== Error กลาง ==================

var ErrNotFound = errors.New("todo not found")

// ใช้ร่วมกับ UPDATE / DELETE เพื่อตรวจว่าโดนแก้ไขจริงกี่แถว
func checkRowsAffectedOne(cmdTag pgconn.CommandTag) error {
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ================== Interface ==================
type TodoRepository interface {
	Create(ctx context.Context, t *Todo) error
	GetByID(ctx context.Context, id, userID int64) (*Todo, error)
	ListByUser(ctx context.Context, userID int64) ([]Todo, error)
	Update(ctx context.Context, t *Todo) error
	Delete(ctx context.Context, id, userID int64) error
}

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) TodoRepository {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) Create(ctx context.Context, t *Todo) error {
	query := `
		INSERT INTO todos (
			title,
			description,
			date_start,
			date_end,
			is_success,
			user_id,
			todo_group_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	// กันเคสลืมเซ็ต date_start (ถึง DB บังคับ NOT NULL แล้ว แต่ช่วย set ให้ตรงนี้ด้วย)
	if t.DateStart.IsZero() {
		t.DateStart = time.Now().UTC()
	}

	return r.db.QueryRow(
		ctx,
		query,
		t.Title,
		t.Description,
		t.DateStart,
		t.DateEnd,
		t.IsSuccess,
		t.UserID,
		t.TodoGroupID,
	).Scan(
		&t.ID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
}

func (r *PostgresRepo) GetByID(ctx context.Context, id, userID int64) (*Todo, error) {
	query := `
		SELECT
			id,
			title,
			description,
			date_start,
			date_end,
			is_success,
			user_id,
			todo_group_id,
			created_at,
			updated_at
		FROM todos
		WHERE id = $1 AND user_id = $2
	`

	var t Todo
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.DateStart,
		&t.DateEnd,
		&t.IsSuccess,
		&t.UserID,
		&t.TodoGroupID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (r *PostgresRepo) ListByUser(ctx context.Context, userID int64) ([]Todo, error) {
	query := `
		SELECT
			id,
			title,
			description,
			date_start,
			date_end,
			is_success,
			user_id,
			todo_group_id,
			created_at,
			updated_at
		FROM todos
		WHERE user_id = $1
		ORDER BY date_start, id
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.DateStart,
			&t.DateEnd,
			&t.IsSuccess,
			&t.UserID,
			&t.TodoGroupID,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *PostgresRepo) Update(ctx context.Context, t *Todo) error {
	query := `
		UPDATE todos
		SET 
			title = $1,
			description = $2,
			date_start = $3,
			date_end = $4,
			is_success = $5,
			todo_group_id = $6
		WHERE id = $7 AND user_id = $8	
	`
	// cmdTag -> สรุปผลลัพธ์ของคำสั่ง SQL ที่รันไป” เช่น: UPDATE ไปกี่แถว, DELETE ไปกี่แถว, INSERT แต่ไม่ RETURNING ได้ผลอะไรมั้ย, คำสั่ง SQL ประเภทไหน
	cmdTag, err := r.db.Exec(
		ctx,
		query,
		t.Title,
		t.Description,
		t.DateStart,
		t.DateEnd,
		t.IsSuccess,
		t.TodoGroupID,
		t.ID,
		t.UserID,
	)
	if err != nil {
		return err
	}

	return checkRowsAffectedOne(cmdTag)
}

func (r *PostgresRepo) Delete(ctx context.Context, id, userID int64) error {
	query := `
		DELETE FROM todos
		WHERE id = $1 AND user_id = $2
	`

	cmdTag, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	return checkRowsAffectedOne(cmdTag)
}
