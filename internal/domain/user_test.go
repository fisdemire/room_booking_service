package domain_test

import (
	"errors"
	"testing"

	"github.com/fisdemire/room-booking/internal/domain"
)

func TestUser_Validate(t *testing.T) {
	validUser := domain.User{
		ID:           "1",
		Email:        "user@example.com",
		PasswordHash: "hashed-password",
		Role:         domain.RoleUser,
	}

	tests := []struct {
		name    string
		mutate  func(*domain.User)
		wantErr error
	}{
		{
			name:    "valid user",
			wantErr: nil,
		},
		{
			name: "empty email",
			mutate: func(u *domain.User) {
				u.Email = ""
			},
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name: "empty password hash",
			mutate: func(u *domain.User) {
				u.PasswordHash = ""
			},
			wantErr: domain.ErrInvalidPassword,
		},
		{
			name: "unknown role",
			mutate: func(u *domain.User) {
				u.Role = domain.Role("superuser")
			},
			wantErr: domain.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := validUser

			if tt.mutate != nil {
				tt.mutate(&user)
			}

			err := user.Validate()

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf(
					"Validate() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role domain.Role
		want bool
	}{
		{
			name: "admin role",
			role: domain.RoleAdmin,
			want: true,
		},
		{
			name: "user role",
			role: domain.RoleUser,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := domain.User{
				Role: tt.role,
			}

			if got := user.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}
