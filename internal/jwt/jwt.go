package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidConfig = errors.New("invalid JWT configuration")
	ErrInvalidToken  = errors.New("invalid JWT token")
)

const issuer = "room-booking"

type Service struct {
	secret []byte
	ttl    time.Duration
}

type claims struct {
	Role domain.Role `json:"role"`
	jwtlib.RegisteredClaims
}

func NewService(secret string, ttl time.Duration) (*Service, error) {
	if len([]byte(secret)) < 32 {
		return nil, fmt.Errorf(
			"%w: JWT secret must contain at least 32 bytes",
			ErrInvalidConfig,
		)
	}

	if ttl <= 0 {
		return nil, fmt.Errorf(
			"%w: JWT TTL must be positive",
			ErrInvalidConfig,
		)
	}

	return &Service{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

func (s *Service) Generate(
	ctx context.Context,
	userID string,
	role domain.Role,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fmt.Errorf("%w: empty user ID", ErrInvalidToken)
	}

	if role != domain.RoleUser && role != domain.RoleAdmin {
		return "", fmt.Errorf("%w: invalid role", ErrInvalidToken)
	}

	now := time.Now().UTC()

	tokenClaims := claims{
		Role: role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID,
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(s.ttl)),
		},
	}

	token := jwtlib.NewWithClaims(
		jwtlib.SigningMethodHS256,
		tokenClaims,
	)

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signed, nil
}

func (s *Service) Parse(
	ctx context.Context,
	tokenString string,
) (string, domain.Role, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}

	if strings.TrimSpace(tokenString) == "" {
		return "", "", ErrInvalidToken
	}

	tokenClaims := &claims{}

	token, err := jwtlib.ParseWithClaims(
		tokenString,
		tokenClaims,
		func(token *jwtlib.Token) (any, error) {
			if token.Method != jwtlib.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return s.secret, nil
		},
		jwtlib.WithValidMethods(
			[]string{jwtlib.SigningMethodHS256.Alg()},
		),
		jwtlib.WithIssuer(issuer),
	)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid ||
		tokenClaims.Subject == "" ||
		tokenClaims.ExpiresAt == nil ||
		(tokenClaims.Role != domain.RoleUser &&
			tokenClaims.Role != domain.RoleAdmin) {
		return "", "", ErrInvalidToken
	}

	return tokenClaims.Subject, tokenClaims.Role, nil
}
