package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type TokenGenerator interface {
	Generate(
		ctx context.Context,
		userID string,
		role domain.Role,
	) (string, error)
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)

	Compare(
		ctx context.Context,
		hashedPassword string,
		password string,
	) error
}

type userUseCase struct {
	users          domain.UserRepository
	tokens         TokenGenerator
	passwordHasher PasswordHasher
}

func NewUserUseCase(
	users domain.UserRepository,
	tokens TokenGenerator,
	passwordHasher PasswordHasher,
) domain.UserUseCase {
	return &userUseCase{
		users:          users,
		tokens:         tokens,
		passwordHasher: passwordHasher,
	}
}

func (u *userUseCase) Register(
	ctx context.Context,
	email string,
	password string,
	role string,
) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	if password == "" {
		return nil, domain.ErrInvalidPassword
	}

	userRole := domain.Role(strings.ToLower(strings.TrimSpace(role)))
	if userRole == "" {
		userRole = domain.RoleUser
	}
	if userRole != domain.RoleUser && userRole != domain.RoleAdmin {
		return nil, domain.ErrInvalidRole
	}

	existing, err := u.users.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	passwordHash, err := u.passwordHasher.Hash(ctx, password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         userRole,
		CreatedAt:    time.Now().UTC(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := u.users.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, domain.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (u *userUseCase) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	email, err := normalizeEmail(email)
	if err != nil || password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := u.users.GetByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("find user by email: %w", err)
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err := u.passwordHasher.Compare(
		ctx,
		user.PasswordHash,
		password,
	); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := u.tokens.Generate(ctx, user.ID, user.Role)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (u *userUseCase) DummyLogin(
	ctx context.Context,
	role string,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	userRole := domain.Role(strings.ToLower(strings.TrimSpace(role)))
	if userRole != domain.RoleUser && userRole != domain.RoleAdmin {
		return "", domain.ErrInvalidRole
	}

	var userID string
	switch userRole {
	case domain.RoleUser:
		userID = domain.DummyUserID
	case domain.RoleAdmin:
		userID = domain.DummyAdminID
	default:
		return "", domain.ErrInvalidRole
	}

	token, err := u.tokens.Generate(ctx, userID, userRole)
	if err != nil {
		return "", fmt.Errorf("generate dummy token: %w", err)
	}

	return token, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" {
		return "", domain.ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", domain.ErrInvalidEmail
	}

	return email, nil
}
