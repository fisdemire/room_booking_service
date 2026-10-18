package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct {
	getter func(context.Context) DBTX
}

func NewScheduleRepository(
	pool *pgxpool.Pool,
	getter func(context.Context) DBTX,
) *ScheduleRepository {
	if getter == nil {
		getter = func(context.Context) DBTX {
			return pool
		}
	}

	return &ScheduleRepository{getter: getter}
}

type scheduleScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(row scheduleScanner) (*domain.Schedule, error) {
	var schedule domain.Schedule
	var days []int16

	err := row.Scan(
		&schedule.ID,
		&schedule.RoomID,
		&days,
		&schedule.StartTime,
		&schedule.EndTime,
	)
	if err != nil {
		return nil, err
	}

	schedule.DaysOfWeek = make([]int, len(days))
	for i, day := range days {
		schedule.DaysOfWeek[i] = int(day)
	}

	return &schedule, nil
}

func mapScheduleDBError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return domain.ErrScheduleAlreadyExists
		case "23503":
			return domain.ErrRoomNotFound
		}
	}

	return err
}

func (r *ScheduleRepository) Create(
	ctx context.Context,
	schedule *domain.Schedule,
) (*domain.Schedule, error) {
	if schedule == nil {
		return nil, fmt.Errorf("create schedule: nil schedule")
	}

	days := make([]int16, len(schedule.DaysOfWeek))
	for i, day := range schedule.DaysOfWeek {
		days[i] = int16(day)
	}

	query := `
		INSERT INTO schedules (
			id,
			room_id,
			days_of_week,
			start_time,
			end_time
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id::text,
			room_id::text,
			days_of_week,
			start_time,
			end_time
	`

	created, err := scanSchedule(r.getter(ctx).QueryRow(
		ctx,
		query,
		schedule.ID,
		schedule.RoomID,
		days,
		schedule.StartTime,
		schedule.EndTime,
	))
	if err != nil {
		return nil, fmt.Errorf(
			"insert schedule: %w",
			mapScheduleDBError(err),
		)
	}

	return created, nil
}

func (r *ScheduleRepository) GetByRoomID(
	ctx context.Context,
	roomID string,
) (*domain.Schedule, error) {
	query := `
		SELECT
			id::text,
			room_id::text,
			days_of_week,
			start_time,
			end_time
		FROM schedules
		WHERE room_id = $1
	`

	schedule, err := scanSchedule(
		r.getter(ctx).QueryRow(ctx, query, roomID),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrScheduleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get schedule by room ID: %w", err)
	}

	return schedule, nil
}

func (r *ScheduleRepository) GetAll(
	ctx context.Context,
) ([]*domain.Schedule, error) {
	query := `
		SELECT
			id::text,
			room_id::text,
			days_of_week,
			start_time,
			end_time
		FROM schedules
		ORDER BY room_id
	`

	rows, err := r.getter(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all schedules: %w", err)
	}
	defer rows.Close()

	schedules := make([]*domain.Schedule, 0)

	for rows.Next() {
		schedule, err := scanSchedule(rows)
		if err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}

		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schedules: %w", err)
	}

	return schedules, nil
}
