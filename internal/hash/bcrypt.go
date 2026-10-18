package hash

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmptyPassword = errors.New("password must not be empty")

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) (*BcryptHasher, error) {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf(
			"bcrypt cost must be between %d and %d",
			bcrypt.MinCost,
			bcrypt.MaxCost,
		)
	}

	return &BcryptHasher{
		cost: cost,
	}, nil
}

func (h *BcryptHasher) Hash(
	ctx context.Context,
	password string,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	if password == "" {
		return "", ErrEmptyPassword
	}

	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		h.cost,
	)
	if err != nil {
		return "", fmt.Errorf("generate password hash: %w", err)
	}

	return string(hashed), nil
}

func (h *BcryptHasher) Compare(
	ctx context.Context,
	hashedPassword string,
	password string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if password == "" {
		return ErrEmptyPassword
	}

	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
}
