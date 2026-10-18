package domain

import (
	"context"
	"strings"
	"time"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) Validate() error {
	if u == nil {
		return ErrInvalidUser
	}

	if strings.TrimSpace(u.Email) == "" {
		return ErrInvalidEmail
	}

	if u.PasswordHash == "" {
		return ErrInvalidPassword
	}

	if u.Role != RoleAdmin && u.Role != RoleUser {
		return ErrInvalidRole
	}

	return nil
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type UserUseCase interface {
	Register(ctx context.Context, email, password, role string) (*User, error)
	Login(ctx context.Context, email, password string) (string, error)
	DummyLogin(ctx context.Context, role string) (string, error)
}
