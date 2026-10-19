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

// CreateSchedule godoc
// @Summary Создать расписание комнаты
// @Description Настраивает еженедельное расписание комнаты и создаёт слоты согласно расписанию. Для комнаты допускается одно расписание. Требуется роль admin.
// @Tags schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "UUID комнаты" format(uuid)
// @Param request body models.CreateScheduleRequest true "Расписание комнаты"
// @Success 201 {object} models.CreateScheduleResponse "Расписание создано"
// @Failure 400 {object} models.ErrorResponse "Некорректный UUID, дни или временной диапазон"
// @Failure 401 {object} models.ErrorResponse "Отсутствует или недействителен JWT"
// @Failure 403 {object} models.ErrorResponse "Требуется роль admin"
// @Failure 404 {object} models.ErrorResponse "Комната не найдена"
// @Failure 409 {object} models.ErrorResponse "Расписание для комнаты уже существует"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка"
// @Router /rooms/{roomId}/schedule/create [post]
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
