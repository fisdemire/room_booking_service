package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/server/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BookingHandler struct {
	bookings domain.BookingUseCase
}

func NewBookingHandler(bookings domain.BookingUseCase) *BookingHandler {
	return &BookingHandler{bookings: bookings}
}

func (h *BookingHandler) RegisterRoutes(
	router gin.IRouter,
	authMiddleware gin.HandlerFunc,
	userMiddleware gin.HandlerFunc,
	adminMiddleware gin.HandlerFunc,
) {
	group := router.Group("/bookings", authMiddleware)

	group.POST("/create", userMiddleware, h.CreateBooking)
	group.GET("/my", userMiddleware, h.GetMyBookings)
	group.GET("/list", adminMiddleware, h.GetAllBookings)
	group.POST("/:bookingId/cancel", userMiddleware, h.CancelBooking)
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID := c.GetString("auth_user_id")
	if userID == "" {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req models.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	if _, err := uuid.Parse(req.SlotID); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	var (
		booking *domain.Booking
		err     error
	)

	if req.CreateConferenceLink {
		booking, err = h.bookings.CreateBookingWithConference(
			c.Request.Context(),
			userID,
			req.SlotID,
		)
	} else {
		booking, err = h.bookings.CreateBooking(
			c.Request.Context(),
			userID,
			req.SlotID,
		)
	}

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSlotNotFound):
			writeError(c, http.StatusNotFound, "SLOT_NOT_FOUND", "slot not found")

		case errors.Is(err, domain.ErrSlotOccupied):
			writeError(c, http.StatusConflict, "SLOT_ALREADY_BOOKED", "slot is already booked")

		case errors.Is(err, domain.ErrSlotInPast),
			errors.Is(err, domain.ErrSlotIDRequired):
			writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")

		default:
			writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	if booking == nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	c.JSON(http.StatusCreated, models.CreateBookingResponse{
		Booking: toBookingResponse(booking),
	})
}

func (h *BookingHandler) GetMyBookings(c *gin.Context) {
	userID := c.GetString("auth_user_id")
	if userID == "" {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	bookings, err := h.bookings.GetMyBookings(c.Request.Context(), userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	response := make([]models.BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		if booking != nil {
			response = append(response, toBookingResponse(booking))
		}
	}

	c.JSON(http.StatusOK, models.ListMyBookingsResponse{
		Bookings: response,
	})
}

func (h *BookingHandler) GetAllBookings(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	offset := (page - 1) * pageSize

	bookings, err := h.bookings.GetAllBookings(c.Request.Context(), pageSize, offset)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	total, err := h.bookings.GetTotalCount(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	response := make([]models.BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		if booking != nil {
			response = append(response, toBookingResponse(booking))
		}
	}

	c.JSON(http.StatusOK, models.ListBookingsResponse{
		Bookings: response,
		Pagination: models.PaginationInfo{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	userID := c.GetString("auth_user_id")
	if userID == "" {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	bookingID := c.Param("bookingId")
	if _, err := uuid.Parse(bookingID); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	booking, err := h.bookings.CancelBooking(
		c.Request.Context(),
		userID,
		bookingID,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBookingNotFound):
			writeError(c, http.StatusNotFound, "BOOKING_NOT_FOUND", "booking not found")

		case errors.Is(err, domain.ErrForbidden):
			writeError(c, http.StatusForbidden, "FORBIDDEN", "cannot cancel another user's booking")

		default:
			writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	if booking == nil {
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	c.JSON(http.StatusOK, models.CancelBookingResponse{
		Booking: toBookingResponse(booking),
	})
}

func toBookingResponse(booking *domain.Booking) models.BookingResponse {
	var conferenceLink *string
	if booking.ConferenceLink != "" {
		link := booking.ConferenceLink
		conferenceLink = &link
	}

	return models.BookingResponse{
		ID:             booking.ID,
		SlotID:         booking.SlotID,
		UserID:         booking.UserID,
		Status:         string(booking.Status),
		ConferenceLink: conferenceLink,
		CreatedAt:      booking.CreatedAt.UTC(),
	}
}
