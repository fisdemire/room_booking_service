package worker_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/worker"
	"github.com/stretchr/testify/require"
)

type fakeSlotGenerator struct {
	calls chan int
}

func (f *fakeSlotGenerator) GenerateSlots(
	ctx context.Context,
	daysAhead int,
) error {
	select {
	case f.calls <- daysAhead:
	default:
	}
	return nil
}

func (f *fakeSlotGenerator) GenerateSlotsForRoom(
	context.Context,
	string,
	int,
) error {
	return nil
}

func TestSlotWorkerGeneratesImmediatelyAndPeriodically(t *testing.T) {
	generator := &fakeSlotGenerator{
		calls: make(chan int, 5),
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	w := worker.NewSlotWorker(
		generator,
		20*time.Millisecond,
		7,
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		w.Start(ctx)
	}()

	select {
	case days := <-generator.calls:
		require.Equal(t, 7, days)
	case <-time.After(time.Second):
		t.Fatal("initial slot generation did not run")
	}

	select {
	case days := <-generator.calls:
		require.Equal(t, 7, days)
	case <-time.After(time.Second):
		t.Fatal("periodic slot generation did not run")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancellation")
	}

	var _ domain.SlotGenerator = generator
}
