package handler_test

import (
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

type slotTestTokenParser struct{}

func (slotTestTokenParser) Parse(
	_ context.Context,
	token string,
) (string, domain.Role, error) {
	if token != "valid-token" {
		return "", "", errors.New("invalid token")
	}

	return "test-user-id", domain.RoleUser, nil
}

func newSlotTestRouter(slots domain.SlotUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler.NewSlotHandler(slots).RegisterRoutes(
		router,
		server.AuthMiddleware(slotTestTokenParser{}),
	)

	return router
}

func TestGetAvailableSlotsSuccess(t *testing.T) {
	roomID := "550e8400-e29b-41d4-a716-446655440000"

	slots := mocks.NewMockSlotUseCase(t)

	date := time.Date(2030, 1, 2, 0, 0, 0, 0, time.UTC)
	start := time.Date(2030, 1, 2, 9, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)

	slots.EXPECT().
		GetAvailableSlots(
			mock.Anything,
			roomID,
			date,
		).
		Return([]*domain.Slot{
			{
				ID:          "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				RoomID:      roomID,
				StartTime:   start,
				EndTime:     end,
				IsAvailable: true,
			},
		}, nil).
		Once()

	router := newSlotTestRouter(slots)

	req := httptest.NewRequest(
		http.MethodGet,
		"/rooms/"+roomID+"/slots/list?date=2030-01-02",
		nil,
	)
	req.Header.Set("Authorization", "Bearer valid-token")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{
		"slots": [{
			"id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			"roomId": "550e8400-e29b-41d4-a716-446655440000",
			"start": "2030-01-02T09:00:00Z",
			"end": "2030-01-02T09:30:00Z"
		}]
	}`, rec.Body.String())
}

func TestGetAvailableSlotsMissingDate(t *testing.T) {
	slots := mocks.NewMockSlotUseCase(t)
	router := newSlotTestRouter(slots)

	req := httptest.NewRequest(
		http.MethodGet,
		"/rooms/550e8400-e29b-41d4-a716-446655440000/slots/list",
		nil,
	)
	req.Header.Set("Authorization", "Bearer valid-token")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.JSONEq(t, `{
		"error": {
			"code": "INVALID_REQUEST",
			"message": "invalid request"
		}
	}`, rec.Body.String())
}

func TestGetAvailableSlotsInvalidDate(t *testing.T) {
	slots := mocks.NewMockSlotUseCase(t)
	router := newSlotTestRouter(slots)

	req := httptest.NewRequest(
		http.MethodGet,
		"/rooms/550e8400-e29b-41d4-a716-446655440000/slots/list?date=not-a-date",
		nil,
	)
	req.Header.Set("Authorization", "Bearer valid-token")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetAvailableSlotsUnauthorized(t *testing.T) {
	slots := mocks.NewMockSlotUseCase(t)
	router := newSlotTestRouter(slots)

	req := httptest.NewRequest(
		http.MethodGet,
		"/rooms/550e8400-e29b-41d4-a716-446655440000/slots/list?date=2030-01-02",
		nil,
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
