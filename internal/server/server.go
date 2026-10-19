package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	_ "github.com/fisdemire/room-booking/docs"
	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/server/handler"
	"github.com/fisdemire/room-booking/internal/server/models"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	server *http.Server
}

func New(
	addr string,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	idleTimeout time.Duration,
	users domain.UserUseCase,
	rooms domain.RoomUseCase,
	schedules domain.ScheduleUseCase,
	slots domain.SlotUseCase,
	bookings domain.BookingUseCase,
	tokens TokenParser,
) *Server {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/_info", Info)

	handler.RegisterRoutes(router, users)

	roomHandler := handler.NewRoomHandler(rooms)
	roomHandler.RegisterRoutes(
		router,
		AuthMiddleware(tokens),
		RequireAdmin(),
	)

	scheduleHandler := handler.NewScheduleHandler(schedules)
	scheduleHandler.RegisterRoutes(
		router,
		AuthMiddleware(tokens),
		RequireAdmin(),
	)

	slotHandler := handler.NewSlotHandler(slots)
	slotHandler.RegisterRoutes(
		router,
		AuthMiddleware(tokens),
	)

	bookingHandler := handler.NewBookingHandler(bookings)
	bookingHandler.RegisterRoutes(
		router,
		AuthMiddleware(tokens),
		RequireUser(),
		RequireAdmin(),
	)

	return &Server{
		server: &http.Server{
			Addr:              addr,
			Handler:           router,
			ReadTimeout:       readTimeout,
			ReadHeaderTimeout: readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}
}

func (s *Server) Run() error {
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) Close() error {
	return s.server.Close()
}

// Info godoc
// @Summary Проверка состояния сервиса
// @Description Возвращает простой статус процесса API.
// @Tags system
// @Produce json
// @Success 200 {object} models.InfoResponse "Сервис доступен"
// @Router /_info [get]
func Info(c *gin.Context) {
	c.JSON(http.StatusOK, models.InfoResponse{Status: "ok"})
}
