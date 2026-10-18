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

type roomTokenParserStub struct {
	role domain.Role
	err  error
}

func (s roomTokenParserStub) Parse(
	_ context.Context,
	token string,
) (string, domain.Role, error) {
	if s.err != nil {
		return "", "", s.err
	}
	if token != "test-token" {
		return "", "", errors.New("invalid token")
	}

	return domain.DummyUserID, s.role, nil
}

func newRoomRouter(
	rooms domain.RoomUseCase,
	role domain.Role,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	h := handler.NewRoomHandler(rooms)

	h.RegisterRoutes(
		router,
		server.AuthMiddleware(roomTokenParserStub{role: role}),
		server.RequireAdmin(),
	)

	return router
}

func performRoomRequest(
	router http.Handler,
	method string,
	path string,
	body string,
	token string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(
		method,
		path,
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

func TestCreateRoomAdminSuccess(t *testing.T) {
	rooms := mocks.NewMockRoomUseCase(t)

	createdAt := time.Date(
		2026, 10, 19, 0, 0, 0, 0, time.UTC,
	)

	rooms.EXPECT().
		CreateRoom(mock.Anything, "Loft", "Small room", 4).
		Return(&domain.Room{
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "Loft",
			Description: "Small room",
			Capacity:    4,
			CreatedAt:   createdAt,
		}, nil).
		Once()

	router := newRoomRouter(rooms, domain.RoleAdmin)

	rec := performRoomRequest(
		router,
		http.MethodPost,
		"/rooms/create",
		`{"name":"Loft","description":"Small room","capacity":4}`,
		"test-token",
	)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.JSONEq(t, `{
		"room": {
			"id": "550e8400-e29b-41d4-a716-446655440000",
			"name": "Loft",
			"description": "Small room",
			"capacity": 4,
			"createdAt": "2026-10-19T00:00:00Z"
		}
	}`, rec.Body.String())
}

func TestCreateRoomUserForbidden(t *testing.T) {
	rooms := mocks.NewMockRoomUseCase(t)
	router := newRoomRouter(rooms, domain.RoleUser)

	rec := performRoomRequest(
		router,
		http.MethodPost,
		"/rooms/create",
		`{"name":"Loft","description":"","capacity":4}`,
		"test-token",
	)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.JSONEq(t, `{
		"error": {
			"code": "FORBIDDEN",
			"message": "admin access required"
		}
	}`, rec.Body.String())
}

func TestListRoomsEmptyReturnsArray(t *testing.T) {
	rooms := mocks.NewMockRoomUseCase(t)

	rooms.EXPECT().
		GetAllRooms(mock.Anything).
		Return(nil, nil).
		Once()

	router := newRoomRouter(rooms, domain.RoleUser)

	rec := performRoomRequest(
		router,
		http.MethodGet,
		"/rooms/list",
		"",
		"test-token",
	)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"rooms":[]}`, rec.Body.String())
}

func TestCreateRoomInvalidBody(t *testing.T) {
	rooms := mocks.NewMockRoomUseCase(t)
	router := newRoomRouter(rooms, domain.RoleAdmin)

	rec := performRoomRequest(
		router,
		http.MethodPost,
		"/rooms/create",
		`{"description":"No name","capacity":4}`,
		"test-token",
	)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.JSONEq(t, `{
		"error": {
			"code": "INVALID_REQUEST",
			"message": "invalid request"
		}
	}`, rec.Body.String())
}

func TestListRoomsRequiresAuthentication(t *testing.T) {
	rooms := mocks.NewMockRoomUseCase(t)
	router := newRoomRouter(rooms, domain.RoleUser)

	rec := performRoomRequest(
		router,
		http.MethodGet,
		"/rooms/list",
		"",
		"",
	)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
