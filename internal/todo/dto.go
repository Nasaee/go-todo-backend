// dto.go
package todo

import "time"

// ใช้ตอนสร้าง
type CreateTodoInput struct {
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	DateStart   time.Time  `json:"date_start"`
	DateEnd     *time.Time `json:"date_end"`
	TodoGroupID int64      `json:"todo_group_id"`
}

// ใช้ตอนแก้ไข
type UpdateTodoInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	DateStart   *time.Time `json:"date_start"`
	DateEnd     *time.Time `json:"date_end"`
	IsSuccess   *bool      `json:"is_success"`
	TodoGroupID *int64     `json:"todo_group_id"`
}
