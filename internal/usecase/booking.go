package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/google/uuid"
)

type bookingUseCase struct {
	bookingRepo       domain.BookingRepository
	slotRepo          domain.SlotRepository
	txManager         TransactionRunner
	conferenceService domain.ConferenceService
}

func NewBookingUseCase(
	bookingRepo domain.BookingRepository,
	slotRepo domain.SlotRepository,
	txManager TransactionRunner,
	conferenceService domain.ConferenceService,
) domain.BookingUseCase {
	return &bookingUseCase{
		bookingRepo:       bookingRepo,
		slotRepo:          slotRepo,
		txManager:         txManager,
		conferenceService: conferenceService,
	}
}

func (uc *bookingUseCase) CreateBooking(
	ctx context.Context,
	userID, slotID string,
) (*domain.Booking, error) {
	return uc.createBooking(ctx, userID, slotID, false)
}

func (uc *bookingUseCase) CreateBookingWithConference(
	ctx context.Context,
	userID, slotID string,
) (*domain.Booking, error) {
	return uc.createBooking(ctx, userID, slotID, true)
}

func (uc *bookingUseCase) createBooking(
	ctx context.Context,
	userID, slotID string,
	withConference bool,
) (*domain.Booking, error) {
	if slotID == "" {
		return nil, domain.ErrSlotIDRequired
	}
	if userID == "" {
		return nil, domain.ErrForbidden
	}

	var created *domain.Booking

	err := uc.txManager.WithinTransaction(
		ctx,
		func(txCtx context.Context) error {
			slot, err := uc.slotRepo.GetByID(txCtx, slotID)
			if err != nil {
				return fmt.Errorf("get slot: %w", err)
			}

			if slot.IsPast(time.Now().UTC()) {
				return domain.ErrSlotInPast
			}

			if err := slot.Book(); err != nil {
				return err
			}

			if err := uc.slotRepo.Update(txCtx, slot); err != nil {
				return fmt.Errorf("mark slot as booked: %w", err)
			}

			booking := &domain.Booking{
				ID:        uuid.NewString(),
				SlotID:    slot.ID,
				UserID:    userID,
				Status:    domain.BookingStatusActive,
				CreatedAt: time.Now().UTC(),
			}

			if withConference {
				if uc.conferenceService == nil {
					return domain.ErrConferenceServiceInvalid
				}

				link, err := uc.conferenceService.CreateConference(
					txCtx,
					booking.ID,
				)
				if err != nil {
					return fmt.Errorf(
						"create conference link: %w",
						err,
					)
				}

				booking.ConferenceLink = link
			}

			if err := uc.bookingRepo.Create(txCtx, booking); err != nil {
				return fmt.Errorf("create booking record: %w", err)
			}

			created = booking
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create booking: %w", err)
	}

	return created, nil
}

func (uc *bookingUseCase) CancelBooking(
	ctx context.Context,
	userID, bookingID string,
) (*domain.Booking, error) {
	if bookingID == "" {
		return nil, domain.ErrBookingIDRequired
	}
	if userID == "" {
		return nil, domain.ErrForbidden
	}

	var result *domain.Booking

	err := uc.txManager.WithinTransaction(
		ctx,
		func(txCtx context.Context) error {
			booking, err := uc.bookingRepo.GetByID(txCtx, bookingID)
			if err != nil {
				return fmt.Errorf("get booking: %w", err)
			}

			if !booking.IsOwnedBy(userID) {
				return domain.ErrForbidden
			}

			// Повторная отмена идемпотентна.
			if !booking.Cancel() {
				result = booking
				return nil
			}

			if err := uc.bookingRepo.Update(txCtx, booking); err != nil {
				return fmt.Errorf("update booking: %w", err)
			}

			slot, err := uc.slotRepo.GetByID(txCtx, booking.SlotID)
			if err != nil {
				return fmt.Errorf("get slot: %w", err)
			}

			slot.Cancel()

			if err := uc.slotRepo.Update(txCtx, slot); err != nil {
				return fmt.Errorf("release slot: %w", err)
			}

			result = booking
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cancel booking: %w", err)
	}

	return result, nil
}

func (uc *bookingUseCase) GetMyBookings(
	ctx context.Context,
	userID string,
) ([]*domain.Booking, error) {
	bookings, err := uc.bookingRepo.GetFutureByUserID(
		ctx,
		userID,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("get my bookings: %w", err)
	}

	if bookings == nil {
		bookings = make([]*domain.Booking, 0)
	}

	return bookings, nil
}

func (uc *bookingUseCase) GetAllBookings(
	ctx context.Context,
	limit, offset int,
) ([]*domain.Booking, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	bookings, err := uc.bookingRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get all bookings: %w", err)
	}

	if bookings == nil {
		bookings = make([]*domain.Booking, 0)
	}

	return bookings, nil
}

func (uc *bookingUseCase) GetTotalCount(
	ctx context.Context,
) (int, error) {
	count, err := uc.bookingRepo.GetTotalCount(ctx)
	if err != nil {
		return 0, fmt.Errorf("get total bookings count: %w", err)
	}

	return count, nil
}
