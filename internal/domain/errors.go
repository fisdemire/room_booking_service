package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidRole       = errors.New("invalid role")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidUser       = errors.New("invalid user")

	ErrRoomNotFound    = errors.New("room not found")
	ErrInvalidRoomName = errors.New("invalid room name")
	ErrInvalidCapacity = errors.New("invalid room capacity")

	ErrScheduleAlreadyExists = errors.New("schedule already exists")
	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrInvalidTimeRange      = errors.New("invalid time range")
	ErrDurationTooShort      = errors.New("schedule duration is too short")
	ErrNoDaysSpecified       = errors.New("no days specified")
	ErrInvalidDayOfWeek      = errors.New("invalid day of week")

	ErrSlotNotFound = errors.New("slot not found")
	ErrSlotOccupied = errors.New("slot is occupied")
	ErrSlotInPast   = errors.New("slot is in the past")

	ErrBookingNotFound = errors.New("booking not found")
	ErrForbidden       = errors.New("forbidden")

	ErrRoomIDAndSlotIDRequired = errors.New("room ID and slot ID are required")
	ErrSlotNotInRoom           = errors.New("slot does not belong to the specified room")
)
