package domain_test

import (
	"errors"
	"testing"

	"github.com/fisdemire/room-booking/internal/domain"
)

func TestRoom_Validate(t *testing.T) {
	validRoom := domain.Room{
		ID:          "room-id",
		Name:        "Meeting room",
		Description: "Room for meetings",
		Capacity:    1,
	}

	tests := []struct {
		name    string
		mutate  func(*domain.Room)
		wantErr error
	}{
		{
			name:    "valid room",
			wantErr: nil,
		},
		{
			name: "empty name",
			mutate: func(r *domain.Room) {
				r.Name = ""
			},
			wantErr: domain.ErrInvalidRoomName,
		},
		{
			name: "whitespace name",
			mutate: func(r *domain.Room) {
				r.Name = " \t\n "
			},
			wantErr: domain.ErrInvalidRoomName,
		},
		{
			name: "zero capacity",
			mutate: func(r *domain.Room) {
				r.Capacity = 0
			},
			wantErr: domain.ErrInvalidCapacity,
		},
		{
			name: "capacity one",
			mutate: func(r *domain.Room) {
				r.Capacity = 1
			},
			wantErr: nil,
		},
		{
			name: "negative capacity",
			mutate: func(r *domain.Room) {
				r.Capacity = -1
			},
			wantErr: domain.ErrInvalidCapacity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room := validRoom

			if tt.mutate != nil {
				tt.mutate(&room)
			}

			err := room.Validate()

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf(
						"Validate() error = %v, want nil",
						err,
					)
				}
				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf(
					"Validate() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
