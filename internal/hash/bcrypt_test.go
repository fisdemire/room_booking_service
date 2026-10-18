package hash_test

import (
	"context"
	"testing"

	"github.com/fisdemire/room-booking/internal/hash"
)

func TestBcryptHasher_HashAndCompare(t *testing.T) {
	hasher, err := hash.NewBcryptHasher(4)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	hashed, err := hasher.Hash(ctx, "password123")
	if err != nil {
		t.Fatal(err)
	}

	if hashed == "password123" {
		t.Fatal("password was stored without hashing")
	}

	if err := hasher.Compare(ctx, hashed, "password123"); err != nil {
		t.Fatalf("expected password to match: %v", err)
	}

	if err := hasher.Compare(ctx, hashed, "wrong-password"); err == nil {
		t.Fatal("expected an error for wrong password")
	}
}

func TestBcryptHasher_EmptyPassword(t *testing.T) {
	hasher, err := hash.NewBcryptHasher(4)
	if err != nil {
		t.Fatal(err)
	}

	_, err = hasher.Hash(context.Background(), "")
	if err == nil {
		t.Fatal("expected an error for empty password")
	}
}
