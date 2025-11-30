package todo

import "time"

type Todo struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	DateStart   time.Time  `json:"date_start"`
	DateEnd     *time.Time `json:"date_end,omitempty"`
	IsSuccess   bool       `json:"is_success"`

	UserID      int64 `json:"user_id"`
	TodoGroupID int64 `json:"todo_group_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
