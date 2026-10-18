package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(
		ctx context.Context,
		sql string,
		args ...any,
	) (pgconn.CommandTag, error)

	Query(
		ctx context.Context,
		sql string,
		args ...any,
	) (pgx.Rows, error)

	QueryRow(
		ctx context.Context,
		sql string,
		args ...any,
	) pgx.Row
}

type txContextKey struct{}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) WithinTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	if fn == nil {
		return fmt.Errorf("transaction callback is nil")
	}

	if _, ok := ctx.Value(txContextKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	if err := fn(txCtx); err != nil {
		rollbackCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			5*time.Second,
		)
		defer cancel()

		rollbackErr := tx.Rollback(rollbackCtx)
		if rollbackErr != nil &&
			!errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return errors.Join(
				err,
				fmt.Errorf("rollback transaction: %w", rollbackErr),
			)
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (m *TxManager) Querier(ctx context.Context) DBTX {
	if tx, ok := ctx.Value(txContextKey{}).(pgx.Tx); ok {
		return tx
	}

	return m.pool
}

type Repositories struct {
	pool   *pgxpool.Pool
	getter func(context.Context) DBTX
}

// New creates repositories that use the pool directly.
func New(pool *pgxpool.Pool) *Repositories {
	return NewWithGetter(pool, nil)
}

// NewWithGetter creates repositories that can use transactions stored in ctx.
func NewWithGetter(
	pool *pgxpool.Pool,
	getter func(context.Context) DBTX,
) *Repositories {
	if getter == nil {
		getter = func(context.Context) DBTX { return pool }
	}

	return &Repositories{
		pool:   pool,
		getter: getter,
	}
}

func (r *Repositories) User() domain.UserRepository {
	return NewUserRepository(r.pool, r.getter)
}

func (r *Repositories) Room() domain.RoomRepository {
	return NewRoomRepository(r.pool, r.getter)
}

func (r *Repositories) Close() {
	if r == nil || r.pool == nil {
		return
	}

	r.pool.Close()
}

func (r *Repositories) Schedule() *ScheduleRepository {
	return NewScheduleRepository(r.pool, r.getter)
}

func (r *Repositories) Slot() *SlotRepository {
	return NewSlotRepository(r.pool, r.getter)
}
