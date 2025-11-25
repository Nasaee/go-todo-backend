package todogroup

import (
	"context"
	"strings"
)

type TodoGroupService interface {
	Create(ctx context.Context, userID int64, input CreateTodoGroupInput) (*TodoGroup, error)
	GetAll(ctx context.Context, userID int64) ([]TodoGroup, error)
}

type service struct {
	repo TodoGroupRepository
}

func NewService(repo TodoGroupRepository) TodoGroupService {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID int64, input CreateTodoGroupInput) (*TodoGroup, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrEmptyName
	}

	g := &TodoGroup{
		Name:   name,
		UserID: userID,
		Color:  input.Color,
	}

	if err := s.repo.Create(ctx, g); err != nil {
		return nil, err
	}

	return g, nil
}

func (s *service) GetAll(ctx context.Context, userID int64) ([]TodoGroup, error) {
	return s.repo.GetAllByUser(ctx, userID)
}
