package handler

import (
	"errors"
	"net/http"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/server/models"
	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	rooms domain.RoomUseCase
}

func NewRoomHandler(rooms domain.RoomUseCase) *RoomHandler {
	return &RoomHandler{
		rooms: rooms,
	}
}

func (h *RoomHandler) RegisterRoutes(
	router gin.IRouter,
	authMiddleware gin.HandlerFunc,
	adminMiddleware gin.HandlerFunc,
) {
	group := router.Group("/rooms", authMiddleware)

	group.GET("/list", h.ListRooms)
	group.POST("/create", adminMiddleware, h.CreateRoom)
}

// CreateRoom godoc
// @Summary Создать комнату
// @Description Создаёт переговорную комнату. Требуется JWT пользователя с ролью admin.
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateRoomRequest true "Данные комнаты"
// @Success 201 {object} models.CreateRoomResponse "Комната создана"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные комнаты"
// @Failure 401 {object} models.ErrorResponse "Отсутствует или недействителен JWT"
// @Failure 403 {object} models.ErrorResponse "Требуется роль admin"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка"
// @Router /rooms/create [post]
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req models.CreateRoomRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request",
		)
		return
	}

	room, err := h.rooms.CreateRoom(
		c.Request.Context(),
		req.Name,
		req.Description,
		req.Capacity,
	)
	if err != nil {
		writeRoomError(c, err)
		return
	}

	if room == nil {
		writeError(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	c.JSON(http.StatusCreated, models.CreateRoomResponse{
		Room: models.RoomResponse{
			ID:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			Capacity:    room.Capacity,
			CreatedAt:   room.CreatedAt.UTC(),
		},
	})
}

// ListRooms godoc
// @Summary Получить список комнат
// @Description Возвращает список всех переговорных комнат. Требуется JWT.
// @Tags rooms
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ListRoomsResponse "Список комнат"
// @Failure 401 {object} models.ErrorResponse "Отсутствует или недействителен JWT"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка"
// @Router /rooms/list [get]
func (h *RoomHandler) ListRooms(c *gin.Context) {
	rooms, err := h.rooms.GetAllRooms(c.Request.Context())
	if err != nil {
		writeError(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	response := models.ListRoomsResponse{
		Rooms: make([]models.RoomResponse, 0, len(rooms)),
	}

	for _, room := range rooms {
		if room == nil {
			continue
		}

		response.Rooms = append(
			response.Rooms,
			models.RoomResponse{
				ID:          room.ID,
				Name:        room.Name,
				Description: room.Description,
				Capacity:    room.Capacity,
				CreatedAt:   room.CreatedAt.UTC(),
			},
		)
	}

	c.JSON(http.StatusOK, response)
}

func writeRoomError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRoomName),
		errors.Is(err, domain.ErrInvalidCapacity):
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
}
