package server_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/server"
	"github.com/fisdemire/room-booking/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

func setupRouter() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_Success(t *testing.T) {
	setupRouter()

	parser := mocks.NewMockTokenParser(t)

	parser.EXPECT().
		Parse(mock.Anything, "valid-token").
		Return("user-123", domain.RoleUser, nil).
		Once()

	router := gin.New()

	router.GET(
		"/protected",
		server.AuthMiddleware(parser),
		func(c *gin.Context) {
			userID, userOK := server.UserIDFromContext(c)
			role, roleOK := server.RoleFromContext(c)

			if !userOK || !roleOK {
				c.Status(http.StatusInternalServerError)
				return
			}

			if userID != "user-123" || role != domain.RoleUser {
				c.Status(http.StatusInternalServerError)
				return
			}

			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestAuthMiddleware_RejectsInvalidAuthorizationHeader(t *testing.T) {
	setupRouter()

	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "missing header",
			header: "",
		},
		{
			name:   "wrong scheme",
			header: "Basic valid-token",
		},
		{
			name:   "missing token",
			header: "Bearer",
		},
		{
			name:   "too many parts",
			header: "Bearer token extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := mocks.NewMockTokenParser(t)

			router := gin.New()
			handlerCalled := false

			router.GET(
				"/protected",
				server.AuthMiddleware(parser),
				func(c *gin.Context) {
					handlerCalled = true
					c.Status(http.StatusNoContent)
				},
			)

			req := httptest.NewRequest(
				http.MethodGet,
				"/protected",
				nil,
			)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf(
					"expected status 401, got %d",
					rec.Code,
				)
			}

			if handlerCalled {
				t.Error("protected handler must not be called")
			}
		})
	}
}

func TestAuthMiddleware_RejectsInvalidToken(t *testing.T) {
	setupRouter()

	parser := mocks.NewMockTokenParser(t)

	parser.EXPECT().
		Parse(mock.Anything, "invalid-token").
		Return("", "", errors.New("invalid token")).
		Once()

	router := gin.New()
	handlerCalled := false

	router.GET(
		"/protected",
		server.AuthMiddleware(parser),
		func(c *gin.Context) {
			handlerCalled = true
			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	if handlerCalled {
		t.Error("protected handler must not be called")
	}
}

func TestRequireAdmin(t *testing.T) {
	setupRouter()

	tests := []struct {
		name       string
		role       domain.Role
		wantStatus int
	}{
		{
			name:       "admin is allowed",
			role:       domain.RoleAdmin,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "regular user is forbidden",
			role:       domain.RoleUser,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := mocks.NewMockTokenParser(t)

			parser.EXPECT().
				Parse(mock.Anything, "valid-token").
				Return("user-123", tt.role, nil).
				Once()

			router := gin.New()

			router.GET(
				"/admin",
				server.AuthMiddleware(parser),
				server.RequireAdmin(),
				func(c *gin.Context) {
					c.Status(http.StatusNoContent)
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req.Header.Set("Authorization", "Bearer valid-token")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}
		})
	}
}
