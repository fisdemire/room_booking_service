package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	getter func(context.Context) DBTX
}

func NewBookingRepository(
	pool *pgxpool.Pool,
	getter func(context.Context) DBTX,
) *BookingRepository {
	if getter == nil {
		getter = func(context.Context) DBTX {
			return pool
		}
	}

	return &BookingRepository{getter: getter}
}

func (r *BookingRepository) Create(
	ctx context.Context,
	booking *domain.Booking,
) error {
	if booking == nil {
		return fmt.Errorf("create booking: nil booking")
	}

	var conferenceLink any
	if booking.ConferenceLink != "" {
		conferenceLink = booking.ConferenceLink
	}

	_, err := r.getter(ctx).Exec(
		ctx,
		`INSERT INTO bookings (
			id, slot_id, user_id, status, conference_link, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		booking.ID,
		booking.SlotID,
		booking.UserID,
		string(booking.Status),
		conferenceLink,
		booking.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "uq_bookings_active_slot" {
			return domain.ErrSlotOccupied
		}

		return fmt.Errorf("insert booking: %w", err)
	}

	return nil
}

func (r *BookingRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Booking, error) {
	query := `
		SELECT
			id::text,
			slot_id::text,
			user_id::text,
			status,
			conference_link,
			created_at
		FROM bookings
		WHERE id = $1
	`

	booking, err := scanBooking(r.getter(ctx).QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBookingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get booking by ID: %w", err)
	}

	return booking, nil
}

func (r *BookingRepository) Update(
	ctx context.Context,
	booking *domain.Booking,
) error {
	if booking == nil {
		return fmt.Errorf("update booking: nil booking")
	}

	tag, err := r.getter(ctx).Exec(
		ctx,
		`UPDATE bookings SET status = $1 WHERE id = $2`,
		string(booking.Status),
		booking.ID,
	)
	if err != nil {
		return fmt.Errorf("update booking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBookingNotFound
	}

	return nil
}

func (r *BookingRepository) GetFutureByUserID(
	ctx context.Context,
	userID string,
	now time.Time,
) ([]*domain.Booking, error) {
	query := `
		SELECT
			b.id::text,
			b.slot_id::text,
			b.user_id::text,
			b.status,
			b.conference_link,
			b.created_at
		FROM bookings AS b
		JOIN slots AS s ON s.id = b.slot_id
		WHERE b.user_id = $1 AND s.start_time >= $2
		ORDER BY s.start_time ASC
	`

	rows, err := r.getter(ctx).Query(ctx, query, userID, now.UTC())
	if err != nil {
		return nil, fmt.Errorf("get future bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]*domain.Booking, 0)

	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}

	return bookings, nil
}

func (r *BookingRepository) GetAll(
	ctx context.Context,
	limit, offset int,
) ([]*domain.Booking, error) {
	query := `
		SELECT
			id::text,
			slot_id::text,
			user_id::text,
			status,
			conference_link,
			created_at
		FROM bookings
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.getter(ctx).Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get all bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]*domain.Booking, 0)

	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}

	return bookings, nil
}

func (r *BookingRepository) GetTotalCount(
	ctx context.Context,
) (int, error) {
	var count int

	err := r.getter(ctx).QueryRow(
		ctx,
		`SELECT COUNT(*) FROM bookings`,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count bookings: %w", err)
	}

	return count, nil
}

type bookingScanner interface {
	Scan(dest ...any) error
}

func scanBooking(row bookingScanner) (*domain.Booking, error) {
	var booking domain.Booking
	var conferenceLink pgtype.Text

	err := row.Scan(
		&booking.ID,
		&booking.SlotID,
		&booking.UserID,
		&booking.Status,
		&conferenceLink,
		&booking.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if conferenceLink.Valid {
		booking.ConferenceLink = conferenceLink.String
	}

	return &booking, nil
}
