package conference

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
)

const conferenceLinkPrefix = "https://meet.example.com/conf-"

type MockConferenceService struct {
	FailureRate float64
	Latency     time.Duration
}

func NewMockConferenceService(
	failureRate float64,
	latency time.Duration,
) *MockConferenceService {
	return &MockConferenceService{
		FailureRate: failureRate,
		Latency:     latency,
	}
}

func (m *MockConferenceService) CreateConference(
	ctx context.Context,
	bookingID string,
) (string, error) {
	if m.Latency > 0 {
		timer := time.NewTimer(m.Latency)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
			return "", fmt.Errorf("conference service: %w", ctx.Err())
		}
	} else if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("conference service: %w", err)
	}

	if m.FailureRate > 0 && rand.Float64() < m.FailureRate {
		return "", domain.ErrConferenceServiceInvalid
	}

	return conferenceLinkPrefix + bookingID, nil
}
