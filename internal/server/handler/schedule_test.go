package handler_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/server"
	"github.com/fisdemire/room-booking/internal/server/handler"
	"github.com/fisdemire/room-booking/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type scheduleTokenParserStub struct {
	role domain.Role
}

func (s scheduleTokenParserStub) Parse(
	_ context.Context,
	token string,
) (string, domain.Role, error) {
	if token != "test-token" {
		return "", "", errors.New("invalid token")
	}

	return domain.DummyUserID, s.role, nil
}

func newScheduleRouter(
	schedules domain.ScheduleUseCase,
	role domain.Role,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	h := handler.NewScheduleHandler(schedules)

	h.RegisterRoutes(
		router,
		server.AuthMiddleware(scheduleTokenParserStub{role: role}),
		server.RequireAdmin(),
	)

	return router
}

func performScheduleRequest(
	router http.Handler,
	body string,
	token string,
	roomID string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(
		http.MethodPost,
		"/rooms/"+roomID+"/schedule/create",
		bytes.NewBufferString(body),
	)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestCreateScheduleSuccess(t *testing.T) {
	roomID := "550e8400-e29b-41d4-a716-446655440000"

	schedules := mocks.NewMockScheduleUseCase(t)
	schedules.EXPECT().
		CreateSchedule(
			mock.Anything,
			roomID,
			[]int{1, 2, 3, 4, 5},
			"09:00",
			"18:00",
		).
		Return(&domain.Schedule{
			ID:         "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			RoomID:     roomID,
			DaysOfWeek: []int{1, 2, 3, 4, 5},
			StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
			EndTime:    time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC),
		}, nil).
		Once()

	router := newScheduleRouter(schedules, domain.RoleAdmin)

	rec := performScheduleRequest(
		router,
		`{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}`,
		"test-token",
		roomID,
	)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.JSONEq(t, `{
		"schedule": {
			"id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			"roomId": "550e8400-e29b-41d4-a716-446655440000",
			"daysOfWeek": [1,2,3,4,5],
			"startTime": "09:00",
			"endTime": "18:00"
		}
	}`, rec.Body.String())
}

func TestCreateScheduleUserForbidden(t *testing.T) {
	schedules := mocks.NewMockScheduleUseCase(t)
	router := newScheduleRouter(schedules, domain.RoleUser)

	rec := performScheduleRequest(
		router,
		`{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}`,
		"test-token",
		"550e8400-e29b-41d4-a716-446655440000",
	)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreateScheduleAlreadyExists(t *testing.T) {
	roomID := "550e8400-e29b-41d4-a716-446655440000"

	schedules := mocks.NewMockScheduleUseCase(t)
	schedules.EXPECT().
		CreateSchedule(
			mock.Anything,
			roomID,
			[]int{1, 2, 3, 4, 5},
			"09:00",
			"18:00",
		).
		Return(nil, domain.ErrScheduleAlreadyExists).
		Once()

	router := newScheduleRouter(schedules, domain.RoleAdmin)

	rec := performScheduleRequest(
		router,
		`{"daysOfWeek":[1,2,3,4,5],"startTime":"09:00","endTime":"18:00"}`,
		"test-token",
		roomID,
	)

	require.Equal(t, http.StatusConflict, rec.Code)
	require.JSONEq(t, `{
		"error": {
			"code": "SCHEDULE_EXISTS",
			"message": "schedule for this room already exists and cannot be changed"
		}
	}`, rec.Body.String())
}

func TestCreateScheduleInvalidBody(t *testing.T) {
	schedules := mocks.NewMockScheduleUseCase(t)
	router := newScheduleRouter(schedules, domain.RoleAdmin)

	rec := performScheduleRequest(
		router,
		`{"daysOfWeek":[8],"startTime":"09:00","endTime":"18:00"}`,
		"test-token",
		"550e8400-e29b-41d4-a716-446655440000",
	)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.JSONEq(t, `{
		"error": {
			"code": "INVALID_REQUEST",
			"message": "invalid request"
		}
	}`, rec.Body.String())
}
