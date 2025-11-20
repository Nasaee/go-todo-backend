package user

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken      = errors.New("email already in use")
	ErrInvalidPassword = errors.New("invalid email or password")
)

type UserService interface {
	Register(ctx context.Context, firstName, lastName, email, password string) (*User, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
}

type service struct {
	repo UserRepository
}

func NewService(repo UserRepository) UserService {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, firstName, lastName, email, password string) (*User, error) {
	// check duplicate email
	_, err := s.repo.FindByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  string(hash),
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return u, nil
}
