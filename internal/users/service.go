package users

import (
	"context"

	"github.com/alexedwards/argon2id"
	"github.com/jakottelaar/relay-backend/internal"
)

type UserService interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type userService struct {
	repo UserRepo
}

func NewUserService(repo UserRepo) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *User) (*User, error) {
	passwordHash, err := argon2id.CreateHash(user.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	user.Password = passwordHash

	savedUser, err := s.repo.SaveUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return savedUser, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, internal.NewNotFoundError("user not found")
	}

	return user, nil
}
