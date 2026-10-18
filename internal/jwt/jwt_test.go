package jwt_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	jwtservice "github.com/fisdemire/room-booking/internal/jwt"
)

func TestService_GenerateAndParse(t *testing.T) {
	service, err := jwtservice.NewService(
		"test-secret-that-is-at-least-32-bytes",
		time.Hour,
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	token, err := service.Generate(
		ctx,
		"user-123",
		domain.RoleUser,
	)
	if err != nil {
		t.Fatal(err)
	}

	userID, role, err := service.Parse(ctx, token)
	if err != nil {
		t.Fatal(err)
	}

	if userID != "user-123" {
		t.Fatalf("unexpected user ID: %s", userID)
	}

	if role != domain.RoleUser {
		t.Fatalf("unexpected role: %s", role)
	}
}

func TestService_ParseInvalidToken(t *testing.T) {
	service, err := jwtservice.NewService(
		"test-secret-that-is-at-least-32-bytes",
		time.Hour,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = service.Parse(context.Background(), "not-a-token")

	if !errors.Is(err, jwtservice.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}
