package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/jakottelaar/relay-backend/internal"
)

type UserRepo interface {
	SaveUser(ctx context.Context, user *User) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) SaveUser(ctx context.Context, user *User) (*User, error) {
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id, created_at`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		switch err.Error() {
		case "pq: duplicate key value violates unique constraint \"users_email_key\"":
			return nil, internal.NewDuplicateError("email already exists")
		case "pq: duplicate key value violates unique constraint \"users_username_key\"":
			return nil, internal.NewDuplicateError("username already exists")
		default:
			return nil, err
		}
	}

	return user, nil
}

func (r *userRepo) FindUserByID(ctx context.Context, id string) (*User, error) {

	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var user User

	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &user, nil

}

func (r *userRepo) FindUserByEmail(ctx context.Context, email string) (*User, error) {

	query := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = $1`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user User

	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &user, nil

}

func (r *userRepo) FindUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE username = $1`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user User

	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &user, nil

}
