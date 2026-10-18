package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/fisdemire/room-booking/internal/domain"
	"github.com/fisdemire/room-booking/internal/server/handler"
	"github.com/gin-gonic/gin"
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
	tokens TokenParser,
) *Server {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/_info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

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
