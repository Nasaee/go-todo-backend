package todo

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidDateRange = errors.New("date_end cannot be before date_start")
)

// Service คือ business logic layer
type Service interface {
	CreateTodo(ctx context.Context, userID int64, in CreateTodoInput) (*Todo, error)
	GetTodo(ctx context.Context, id, userID int64) (*Todo, error)
	ListTodos(ctx context.Context, userID int64) ([]Todo, error)
	ListTodayTodos(ctx context.Context, userID int64) ([]Todo, error)
	ListTomorrowTodos(ctx context.Context, userID int64) ([]Todo, error)
	ListThisWeekTodos(ctx context.Context, userID int64) ([]Todo, error)
	UpdateTodo(ctx context.Context, id, userID int64, in UpdateTodoInput) (*Todo, error)
	DeleteTodo(ctx context.Context, id, userID int64) error
}

type service struct {
	repo TodoRepository
	// ถ้าอนาคตอยากเช็คว่า todo_group เป็นของ user จริง
	// สามารถ inject groupRepo เพิ่มตรงนี้ได้
}

func NewService(repo TodoRepository) Service {
	return &service{repo: repo}
}

// ===== helper validate =====

func validateDateRange(start time.Time, end *time.Time) error {
	if end == nil {
		return nil
	}
	if end.Before(start) {
		return ErrInvalidDateRange
	}
	return nil
}

// ===== Create =====

func (s *service) CreateTodo(ctx context.Context, userID int64, in CreateTodoInput) (*Todo, error) {
	// title required
	if in.Title == "" {
		return nil, ErrInvalidInput
	}

	// date_start required (ห้าม zero)
	if in.DateStart.IsZero() {
		return nil, ErrInvalidInput
	}

	// todo_group_id required (เพราะใช้ identity เริ่มจาก 1)
	if in.TodoGroupID == 0 {
		return nil, ErrInvalidInput
	}

	// validate date range (date_end optional)
	if err := validateDateRange(in.DateStart, in.DateEnd); err != nil {
		return nil, err
	}

	todo := &Todo{
		Title:       in.Title,
		Description: in.Description,
		DateStart:   in.DateStart,
		DateEnd:     in.DateEnd, // nil OK
		IsSuccess:   false,
		UserID:      userID,
		TodoGroupID: in.TodoGroupID,
	}

	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

// ===== Get / List =====

func (s *service) GetTodo(ctx context.Context, id, userID int64) (*Todo, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *service) ListTodos(ctx context.Context, userID int64) ([]Todo, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *service) ListTodayTodos(ctx context.Context, userID int64) ([]Todo, error) {
	return s.repo.ListToday(ctx, userID)
}

func (s *service) ListTomorrowTodos(ctx context.Context, userID int64) ([]Todo, error) {
	return s.repo.ListTomorrow(ctx, userID)
}

func (s *service) ListThisWeekTodos(ctx context.Context, userID int64) ([]Todo, error) {
	return s.repo.ListThisWeek(ctx, userID)
}

// ===== Update =====

func (s *service) UpdateTodo(ctx context.Context, id, userID int64, in UpdateTodoInput) (*Todo, error) {
	// ดึงข้อมูลเดิมมาก่อน
	existing, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// title (optional แต่ถ้าส่งมา ห้ามค่าว่าง)
	if in.Title != nil {
		if *in.Title == "" {
			return nil, ErrInvalidInput
		}
		existing.Title = *in.Title
	}

	// description
	if in.Description != nil {
		existing.Description = in.Description
	}

	// date_start (ถ้าส่งมา ต้องไม่ zero)
	if in.DateStart != nil {
		if in.DateStart.IsZero() {
			return nil, ErrInvalidInput
		}
		existing.DateStart = *in.DateStart
	}

	// date_end (optional, แต่ validate ตอนท้าย)
	if in.DateEnd != nil {
		existing.DateEnd = in.DateEnd // อนุญาตให้แก้เป็น nil ได้ด้วย
	}

	// is_success
	if in.IsSuccess != nil {
		existing.IsSuccess = *in.IsSuccess
	}

	// todo_group_id (ถ้าส่งมา ห้ามเป็น 0)
	if in.TodoGroupID != nil {
		if *in.TodoGroupID == 0 {
			return nil, ErrInvalidInput
		}
		existing.TodoGroupID = *in.TodoGroupID
	}

	// validate date range หลัง merge
	if err := validateDateRange(existing.DateStart, existing.DateEnd); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// ===== Delete =====

func (s *service) DeleteTodo(ctx context.Context, id, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}
