package usecase_test

import (
	"context"
	"testing"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/usecase"
	"github.com/fisdemire/room-booking/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserUseCase_Register(t *testing.T) {
	ctx := context.Background()

	repo := mocks.NewMockUserRepository(t)
	hasher := mocks.NewMockPasswordHasher(t)
	tokens := mocks.NewMockTokenGenerator(t)

	repo.EXPECT().
		GetByEmail(mock.Anything, "user@example.com").
		Return(nil, domain.ErrUserNotFound).
		Once()

	hasher.EXPECT().
		Hash(mock.Anything, "password123").
		Return("hashed-password", nil).
		Once()

	repo.EXPECT().
		Create(
			mock.Anything,
			mock.MatchedBy(func(user *domain.User) bool {
				return user.Email == "user@example.com" &&
					user.PasswordHash == "hashed-password" &&
					user.Role == domain.RoleUser
			}),
		).
		Run(func(_ context.Context, user *domain.User) {
			user.ID = "user-1"
		}).
		Return(nil).
		Once()

	service := usecase.NewUserUseCase(repo, tokens, hasher)

	user, err := service.Register(
		ctx,
		" USER@example.com ",
		"password123",
		"user",
	)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "user-1", user.ID)
	require.Equal(t, "user@example.com", user.Email)
	require.Equal(t, "hashed-password", user.PasswordHash)
	require.Equal(t, domain.RoleUser, user.Role)
}
