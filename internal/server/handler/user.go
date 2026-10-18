package handler

import (
	"errors"
	"net/http"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/server/models"
	"github.com/fisdemire/room-booking/internal/usecase"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users domain.UserUseCase
}

func NewUserHandler(users domain.UserUseCase) *UserHandler {
	return &UserHandler{users: users}
}

func RegisterRoutes(router gin.IRoutes, users domain.UserUseCase) {
	h := NewUserHandler(users)
	router.POST("/dummyLogin", h.DummyLogin)
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)
}

func (h *UserHandler) DummyLogin(c *gin.Context) {
	var req models.DummyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	token, err := h.users.DummyLogin(c.Request.Context(), req.Role)
	if err != nil {
		writeUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{Token: token})
}

func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	user, err := h.users.Register(
		c.Request.Context(),
		req.Email,
		req.Password,
		req.Role,
	)
	if err != nil {
		writeUserError(c, err)
		return
	}
	if user == nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	c.JSON(http.StatusCreated, models.RegisterResponse{
		User: models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Role:      string(user.Role),
			CreatedAt: user.CreatedAt,
		},
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	token, err := h.users.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeUserError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{Token: token})
}

func writeUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidPassword),
		errors.Is(err, domain.ErrInvalidRole),
		errors.Is(err, domain.ErrUserAlreadyExists):
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")

	case errors.Is(err, usecase.ErrInvalidCredentials),
		errors.Is(err, domain.ErrUserNotFound):
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid email or password")

	default:
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
