package models

import "time"

type DummyLoginRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
}

type RoomResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Capacity    int       `json:"capacity"`
	CreatedAt   time.Time `json:"createdAt"`
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
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CreateScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek" binding:"required,min=1,dive,min=1,max=7"`
	StartTime  string `json:"startTime" binding:"required"`
	EndTime    string `json:"endTime" binding:"required"`
}

type ScheduleResponse struct {
	ID         string `json:"id"`
	RoomID     string `json:"roomId"`
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type CreateScheduleResponse struct {
	Schedule ScheduleResponse `json:"schedule"`
}
