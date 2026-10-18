package handler

import (
	"errors"
	"net/http"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/server/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	schedules domain.ScheduleUseCase
}

func NewScheduleHandler(
	schedules domain.ScheduleUseCase,
) *ScheduleHandler {
	return &ScheduleHandler{schedules: schedules}
}

func (h *ScheduleHandler) RegisterRoutes(
	router gin.IRouter,
	authMiddleware gin.HandlerFunc,
	adminMiddleware gin.HandlerFunc,
) {
	group := router.Group("/rooms", authMiddleware)
	group.POST(
		"/:roomId/schedule/create",
		adminMiddleware,
		h.CreateSchedule,
	)
}

func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	roomID := c.Param("roomId")

	if _, err := uuid.Parse(roomID); err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request",
		)
		return
	}

	var req models.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request",
		)
		return
	}

	schedule, err := h.schedules.CreateSchedule(
		c.Request.Context(),
		roomID,
		req.DaysOfWeek,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrScheduleAlreadyExists):
			writeError(
				c,
				http.StatusConflict,
				"SCHEDULE_EXISTS",
				"schedule for this room already exists and cannot be changed",
			)

		case errors.Is(err, domain.ErrRoomNotFound):
			writeError(
				c,
				http.StatusNotFound,
				"ROOM_NOT_FOUND",
				"room not found",
			)

		case errors.Is(err, domain.ErrInvalidTimeRange),
			errors.Is(err, domain.ErrDurationTooShort),
			errors.Is(err, domain.ErrNoDaysSpecified),
			errors.Is(err, domain.ErrInvalidDayOfWeek):
			writeError(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				"invalid request",
			)

		default:
			writeError(
				c,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"internal server error",
			)
		}

		return
	}

	c.JSON(http.StatusCreated, models.CreateScheduleResponse{
		Schedule: models.ScheduleResponse{
			ID:         schedule.ID,
			RoomID:     schedule.RoomID,
			DaysOfWeek: schedule.DaysOfWeek,
			StartTime:  schedule.StartTime.Format("15:04"),
			EndTime:    schedule.EndTime.Format("15:04"),
		},
	})
}
