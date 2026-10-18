package domain

import (
	"context"
	"errors"
)

var ErrConferenceServiceInvalid = errors.New("conference service invalid")

type ConferenceService interface {
	CreateConference(ctx context.Context, bookingID string) (string, error)
}
