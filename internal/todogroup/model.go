package todogroup

import "time"

type TodoGroup struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// รับจาก body ตอน client สร้าง group
type CreateTodoGroupInput struct {
	Name string `json:"name"`
}
