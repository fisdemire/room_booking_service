package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository stores and loads users from PostgreSQL. The getter chooses
// the transaction from the context when one exists, otherwise it returns pool.
type UserRepository struct {
	getter func(context.Context) DBTX
}

func NewUserRepository(
	pool *pgxpool.Pool,
	getter func(context.Context) DBTX,
) *UserRepository {
	if getter == nil {
		getter = func(context.Context) DBTX { return pool }
	}

	return &UserRepository{getter: getter}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return domain.ErrInvalidUser
	}

	err := r.getter(ctx).QueryRow(
		ctx,
		`INSERT INTO users (email, password_hash, role)
		 VALUES ($1, $2, $3)
		 RETURNING id::text, created_at`,
		user.Email,
		user.PasswordHash,
		string(user.Role),
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.User, error) {
	return r.getBy(ctx,
		`SELECT id::text, email, password_hash, role, created_at
		 FROM users WHERE id = $1`,
		id,
	)
}

func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	return r.getBy(ctx,
		`SELECT id::text, email, password_hash, role, created_at
		 FROM users WHERE email = $1`,
		email,
	)
}

func (r *UserRepository) getBy(
	ctx context.Context,
	query string,
	arg string,
) (*domain.User, error) {
	var user domain.User
	var role string

	err := r.getter(ctx).QueryRow(ctx, query, arg).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&role,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}

	user.Role = domain.Role(role)
	return &user, nil
}
