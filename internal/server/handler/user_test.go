package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fisdemire/room-booking/internal/server/handler"
	"github.com/fisdemire/room-booking/internal/usecase"
	"github.com/fisdemire/room-booking/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDummyLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	users := mocks.NewMockUserUseCase(t)
	users.EXPECT().
		DummyLogin(mock.Anything, "user").
		Return("test-jwt", nil).
		Once()

	router := gin.New()
	handler.RegisterRoutes(router, users)

	req := httptest.NewRequest(
		http.MethodPost,
		"/dummyLogin",
		bytes.NewBufferString(`{"role":"user"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"token":"test-jwt"}`, response.Body.String())
}

func TestDummyLoginInvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	users := mocks.NewMockUserUseCase(t)
	router := gin.New()
	handler.RegisterRoutes(router, users)

	req := httptest.NewRequest(
		http.MethodPost,
		"/dummyLogin",
		bytes.NewBufferString(`{"role":"superuser"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(
		t,
		`{"error":{"code":"INVALID_REQUEST","message":"invalid request"}}`,
		response.Body.String(),
	)
}

func TestLoginInvalidCredentialsUseSameResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	users := mocks.NewMockUserUseCase(t)
	users.EXPECT().
		Login(mock.Anything, "user@example.com", "wrong-password").
		Return("", usecase.ErrInvalidCredentials).
		Once()

	router := gin.New()
	handler.RegisterRoutes(router, users)

	requestBody, err := json.Marshal(map[string]string{
		"email":    "user@example.com",
		"password": "wrong-password",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.JSONEq(
		t,
		`{"error":{"code":"UNAUTHORIZED","message":"invalid email or password"}}`,
		response.Body.String(),
	)
}
