package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/server/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SlotHandler struct {
	slots domain.SlotUseCase
}

func NewSlotHandler(slots domain.SlotUseCase) *SlotHandler {
	return &SlotHandler{slots: slots}
}

func (h *SlotHandler) RegisterRoutes(
	router gin.IRouter,
	authMiddleware gin.HandlerFunc,
) {
	group := router.Group("/rooms", authMiddleware)
	group.GET("/:roomId/slots/list", h.GetAvailableSlots)
}

func (h *SlotHandler) GetAvailableSlots(c *gin.Context) {
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

	dateValue := c.Query("date")
	if dateValue == "" {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request",
		)
		return
	}

	date, err := time.Parse(time.DateOnly, dateValue)
	if err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request",
		)
		return
	}

	slots, err := h.slots.GetAvailableSlots(
		c.Request.Context(),
		roomID,
		date.UTC(),
	)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			writeError(
				c,
				http.StatusNotFound,
				"ROOM_NOT_FOUND",
				"room not found",
			)
			return
		}

		writeError(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	response := models.ListSlotsResponse{
		Slots: make([]models.SlotResponse, 0, len(slots)),
	}

	for _, slot := range slots {
		if slot == nil {
			continue
		}

		response.Slots = append(
			response.Slots,
			models.SlotResponse{
				ID:     slot.ID,
				RoomID: slot.RoomID,
				Start:  slot.StartTime.UTC(),
				End:    slot.EndTime.UTC(),
			},
		)
	}

	c.JSON(http.StatusOK, response)
}
