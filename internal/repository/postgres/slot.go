package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotRepository struct {
	getter func(context.Context) DBTX
}

func NewSlotRepository(
	pool *pgxpool.Pool,
	getter func(context.Context) DBTX,
) *SlotRepository {
	if getter == nil {
		getter = func(context.Context) DBTX {
			return pool
		}
	}

	return &SlotRepository{getter: getter}
}

func (r *SlotRepository) CreateBulk(
	ctx context.Context,
	slots []*domain.Slot,
) error {
	query := `
		INSERT INTO slots (
			id, room_id, start_time, end_time, is_available
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (room_id, start_time) DO NOTHING
	`

	for _, slot := range slots {
		if slot == nil {
			return fmt.Errorf("create slots: nil slot")
		}

		_, err := r.getter(ctx).Exec(
			ctx,
			query,
			slot.ID,
			slot.RoomID,
			slot.StartTime.UTC(),
			slot.EndTime.UTC(),
			slot.IsAvailable,
		)
		if err != nil {
			return fmt.Errorf("insert slot: %w", err)
		}
	}

	return nil
}

func (r *SlotRepository) GetAvailableByRoomAndTime(
	ctx context.Context,
	roomID string,
	date time.Time,
) ([]*domain.Slot, error) {
	day := date.UTC()
	start := time.Date(
		day.Year(), day.Month(), day.Day(),
		0, 0, 0, 0, time.UTC,
	)
	end := start.AddDate(0, 0, 1)

	query := `
		SELECT id::text, room_id::text, start_time, end_time, is_available
		FROM slots
		WHERE room_id = $1
			AND start_time >= $2
			AND start_time < $3
			AND is_available = TRUE
		ORDER BY start_time
	`

	rows, err := r.getter(ctx).Query(ctx, query, roomID, start, end)
	if err != nil {
		return nil, fmt.Errorf("get available slots: %w", err)
	}
	defer rows.Close()

	slots := make([]*domain.Slot, 0)

	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(
			&slot.ID,
			&slot.RoomID,
			&slot.StartTime,
			&slot.EndTime,
			&slot.IsAvailable,
		); err != nil {
			return nil, fmt.Errorf("scan slot: %w", err)
		}

		slots = append(slots, &slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate slots: %w", err)
	}

	return slots, nil
}

func (r *SlotRepository) GetByID(
	ctx context.Context,
	slotID string,
) (*domain.Slot, error) {
	query := `
		SELECT id::text, room_id::text, start_time, end_time, is_available
		FROM slots
		WHERE id = $1
	`

	var slot domain.Slot
	err := r.getter(ctx).QueryRow(ctx, query, slotID).Scan(
		&slot.ID,
		&slot.RoomID,
		&slot.StartTime,
		&slot.EndTime,
		&slot.IsAvailable,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSlotNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get slot by ID: %w", err)
	}

	return &slot, nil
}

func (r *SlotRepository) Update(
	ctx context.Context,
	slot *domain.Slot,
) error {
	if slot == nil {
		return fmt.Errorf("update slot: nil slot")
	}

	tag, err := r.getter(ctx).Exec(
		ctx,
		`UPDATE slots SET is_available = $1 WHERE id = $2`,
		slot.IsAvailable,
		slot.ID,
	)
	if err != nil {
		return fmt.Errorf("update slot: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSlotNotFound
	}

	return nil
}
