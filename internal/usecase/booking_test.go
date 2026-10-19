package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/usecase"
	"github.com/fisdemire/room-booking/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type bookingTestTxRunner struct{}

func (bookingTestTxRunner) WithinTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	return fn(ctx)
}

func TestBookingUseCase_CreateSuccess(t *testing.T) {
	ctx := context.Background()
	userID := "00000000-0000-0000-0000-000000000001"
	slotID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"

	bookingRepo := mocks.NewMockBookingRepository(t)
	slotRepo := mocks.NewMockSlotRepository(t)

	slot := &domain.Slot{
		ID:          slotID,
		RoomID:      "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		StartTime:   time.Now().UTC().Add(time.Hour),
		EndTime:     time.Now().UTC().Add(90 * time.Minute),
		IsAvailable: true,
	}

	slotRepo.EXPECT().
		GetByID(mock.Anything, slotID).
		Return(slot, nil).
		Once()

	slotRepo.EXPECT().
		Update(mock.Anything, mock.MatchedBy(func(s *domain.Slot) bool {
			return s.ID == slotID && !s.IsAvailable
		})).
		Return(nil).
		Once()

	bookingRepo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(b *domain.Booking) bool {
			return b != nil &&
				b.SlotID == slotID &&
				b.UserID == userID &&
				b.Status == domain.BookingStatusActive &&
				b.CreatedAt.Location() == time.UTC
		})).
		Return(nil).
		Once()

	uc := usecase.NewBookingUseCase(
		bookingRepo,
		slotRepo,
		bookingTestTxRunner{},
		nil,
	)

	booking, err := uc.CreateBooking(ctx, userID, slotID)

	require.NoError(t, err)
	require.NotNil(t, booking)
	require.Equal(t, slotID, booking.SlotID)
	require.Equal(t, userID, booking.UserID)
	require.Equal(t, domain.BookingStatusActive, booking.Status)
	require.Empty(t, booking.ConferenceLink)
}

func TestBookingUseCase_CreateOccupiedSlot(t *testing.T) {
	slotID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"

	bookingRepo := mocks.NewMockBookingRepository(t)
	slotRepo := mocks.NewMockSlotRepository(t)

	slotRepo.EXPECT().
		GetByID(mock.Anything, slotID).
		Return(&domain.Slot{
			ID:          slotID,
			StartTime:   time.Now().UTC().Add(time.Hour),
			IsAvailable: false,
		}, nil).
		Once()

	uc := usecase.NewBookingUseCase(
		bookingRepo,
		slotRepo,
		bookingTestTxRunner{},
		nil,
	)

	booking, err := uc.CreateBooking(
		context.Background(),
		"00000000-0000-0000-0000-000000000001",
		slotID,
	)

	require.Nil(t, booking)
	require.ErrorIs(t, err, domain.ErrSlotOccupied)
}

func TestBookingUseCase_CannotCancelAnotherUsersBooking(t *testing.T) {
	bookingRepo := mocks.NewMockBookingRepository(t)
	slotRepo := mocks.NewMockSlotRepository(t)

	bookingRepo.EXPECT().
		GetByID(mock.Anything, "cccccccc-cccc-4ccc-8ccc-cccccccccccc").
		Return(&domain.Booking{
			ID:     "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
			UserID: "another-user",
			Status: domain.BookingStatusActive,
		}, nil).
		Once()

	uc := usecase.NewBookingUseCase(
		bookingRepo,
		slotRepo,
		bookingTestTxRunner{},
		nil,
	)

	booking, err := uc.CancelBooking(
		context.Background(),
		"current-user",
		"cccccccc-cccc-4ccc-8ccc-cccccccccccc",
	)

	require.Nil(t, booking)
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestBookingUseCase_CancelIsIdempotent(t *testing.T) {
	bookingRepo := mocks.NewMockBookingRepository(t)
	slotRepo := mocks.NewMockSlotRepository(t)

	existing := &domain.Booking{
		ID:     "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
		SlotID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		UserID: "current-user",
		Status: domain.BookingStatusCancelled,
	}

	bookingRepo.EXPECT().
		GetByID(mock.Anything, existing.ID).
		Return(existing, nil).
		Once()

	uc := usecase.NewBookingUseCase(
		bookingRepo,
		slotRepo,
		bookingTestTxRunner{},
		nil,
	)

	booking, err := uc.CancelBooking(
		context.Background(),
		existing.UserID,
		existing.ID,
	)

	require.NoError(t, err)
	require.NotNil(t, booking)
	require.Equal(t, domain.BookingStatusCancelled, booking.Status)
}
