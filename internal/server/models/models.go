package models

import "time"

type InfoResponse struct {
	Status string `json:"status" example:"ok"`
}

type DummyLoginRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user" example:"user"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"StrongPassword123"`
	Role     string `json:"role" binding:"required,oneof=admin user" example:"user"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"StrongPassword123"`
}

type UserResponse struct {
	ID        string    `json:"id" example:"e090bb4e-9508-4c4b-bd18-03f64f56e170"`
	Email     string    `json:"email" example:"user@example.com"`
	Role      string    `json:"role" example:"user"`
	CreatedAt time.Time `json:"createdAt" example:"2026-10-20T10:00:00Z"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}

type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required" example:"Большая переговорная"`
	Description string `json:"description" example:"Переговорная на 8 человек"`
	Capacity    int    `json:"capacity" example:"8"`
}

type RoomResponse struct {
	ID          string    `json:"id" example:"b7a3a880-62c9-47e1-99f7-6f0cc9c5e111"`
	Name        string    `json:"name" example:"Большая переговорная"`
	Description string    `json:"description" example:"Переговорная на 8 человек"`
	Capacity    int       `json:"capacity" example:"8"`
	CreatedAt   time.Time `json:"createdAt" example:"2026-10-20T10:00:00Z"`
}

type CreateRoomResponse struct {
	Room RoomResponse `json:"room"`
}

type ListRoomsResponse struct {
	Rooms []RoomResponse `json:"rooms"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"invalid request"`
}

type CreateScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek" binding:"required,min=1,dive,min=1,max=7" example:"1,2,3,4,5"`
	StartTime  string `json:"startTime" binding:"required" example:"09:00"`
	EndTime    string `json:"endTime" binding:"required" example:"18:00"`
}

type ScheduleResponse struct {
	ID         string `json:"id" example:"2ccfc08f-cc8d-4b8c-b949-55a70a9b8511"`
	RoomID     string `json:"roomId" example:"b7a3a880-62c9-47e1-99f7-6f0cc9c5e111"`
	DaysOfWeek []int  `json:"daysOfWeek" example:"1,2,3,4,5"`
	StartTime  string `json:"startTime" example:"09:00"`
	EndTime    string `json:"endTime" example:"18:00"`
}

type CreateScheduleResponse struct {
	Schedule ScheduleResponse `json:"schedule"`
}

type SlotResponse struct {
	ID     string    `json:"id" example:"6f9f749e-8f28-41e7-8d8f-bc725c58ac24"`
	RoomID string    `json:"roomId" example:"b7a3a880-62c9-47e1-99f7-6f0cc9c5e111"`
	Start  time.Time `json:"start" example:"2026-10-21T09:00:00Z"`
	End    time.Time `json:"end" example:"2026-10-21T09:30:00Z"`
}

type ListSlotsResponse struct {
	Slots []SlotResponse `json:"slots"`
}

type CreateBookingRequest struct {
	SlotID               string `json:"slotId" binding:"required" example:"6f9f749e-8f28-41e7-8d8f-bc725c58ac24"`
	CreateConferenceLink bool   `json:"createConferenceLink" example:"false"`
}

type BookingResponse struct {
	ID             string    `json:"id" example:"89f8a5c2-1e4d-48b6-9d90-32e7c12d3e34"`
	SlotID         string    `json:"slotId" example:"6f9f749e-8f28-41e7-8d8f-bc725c58ac24"`
	UserID         string    `json:"userId" example:"e090bb4e-9508-4c4b-bd18-03f64f56e170"`
	Status         string    `json:"status" example:"active"`
	ConferenceLink *string   `json:"conferenceLink" example:"https://conference.example/room/123"`
	CreatedAt      time.Time `json:"createdAt" example:"2026-10-20T10:00:00Z"`
}

type CreateBookingResponse struct {
	Booking BookingResponse `json:"booking"`
}

type CancelBookingResponse struct {
	Booking BookingResponse `json:"booking"`
}

type ListMyBookingsResponse struct {
	Bookings []BookingResponse `json:"bookings"`
}

type PaginationInfo struct {
	Page     int `json:"page" example:"1"`
	PageSize int `json:"pageSize" example:"20"`
	Total    int `json:"total" example:"42"`
}

type ListBookingsResponse struct {
	Bookings   []BookingResponse `json:"bookings"`
	Pagination PaginationInfo    `json:"pagination"`
}
