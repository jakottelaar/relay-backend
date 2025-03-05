package users

import (
	"context"

	"github.com/alexedwards/argon2id"
	"github.com/jakottelaar/relay-backend/config"
	"github.com/jakottelaar/relay-backend/internal"
)

type UserService interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	LoginUser(ctx context.Context, email, password string) (*LoginResponse, error)
}

type userService struct {
	repo UserRepo
	cfg  config.Config
}

func NewUserService(repo UserRepo, cfg config.Config) UserService {
	return &userService{
		repo: repo,
		cfg:  cfg,
	}
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

func (s *userService) LoginUser(ctx context.Context, email, password string) (*LoginResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, internal.NewNotFoundError("user not found")
	}

	match, err := argon2id.ComparePasswordAndHash(password, user.Password)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, internal.NewUnauthorizedError("invalid credentials")
	}

	token, err := internal.GenerateJWT(user.ID.String(), s.cfg.JwtSecret, s.cfg.JwtExpirationSecond)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		UserID:      user.ID,
		UserName:    user.Username,
		AccessToken: token,
	}, nil
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
