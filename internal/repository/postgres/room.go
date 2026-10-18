package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	getter func(context.Context) DBTX
}

func NewRoomRepository(
	pool *pgxpool.Pool,
	getter func(context.Context) DBTX,
) *RoomRepository {
	if getter == nil {
		getter = func(context.Context) DBTX {
			return pool
		}
	}

	return &RoomRepository{
		getter: getter,
	}
}

func (r *RoomRepository) Create(
	ctx context.Context,
	room *domain.Room,
) error {
	if room == nil {
		return fmt.Errorf("create room: room is nil")
	}

	_, err := r.getter(ctx).Exec(
		ctx,
		`INSERT INTO rooms (
			id,
			name,
			description,
			capacity,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)`,
		room.ID,
		room.Name,
		room.Description,
		room.Capacity,
		room.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert room: %w", err)
	}

	return nil
}

func (r *RoomRepository) GetAll(
	ctx context.Context,
) ([]*domain.Room, error) {
	rows, err := r.getter(ctx).Query(
		ctx,
		`SELECT
			id::text,
			name,
			description,
			capacity,
			created_at
		FROM rooms
		ORDER BY created_at, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("select rooms: %w", err)
	}
	defer rows.Close()

	rooms := make([]*domain.Room, 0)

	for rows.Next() {
		var room domain.Room

		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.Capacity,
			&room.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}

		rooms = append(rooms, &room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rooms: %w", err)
	}

	return rooms, nil
}

func (r *RoomRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Room, error) {
	var room domain.Room

	err := r.getter(ctx).QueryRow(
		ctx,
		`SELECT id::text, name, description, capacity, created_at
         FROM rooms
         WHERE id = $1`,
		id,
	).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.Capacity,
		&room.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrRoomNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get room by ID: %w", err)
	}

	return &room, nil
}
