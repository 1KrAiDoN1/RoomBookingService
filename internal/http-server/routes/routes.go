package routes

import (
	"internship/internal/http-server/handler"
	"internship/internal/http-server/middleware"

	"github.com/gin-gonic/gin"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

func SetupBookingServiceRoutes(
	router *gin.RouterGroup,
	handlers *handler.Handlers,
) {
	rooms := router.Group("/rooms")
	{
		rooms.GET("/list", handlers.RoomsHandler.ListRooms)
		rooms.GET("/:roomID/slots/list", handlers.RoomsHandler.ListAvailableSlots)

		adminRooms := rooms.Group("", middleware.RequireRole(RoleAdmin))
		{
			adminRooms.POST("/create", handlers.RoomsHandler.CreateRoom)
			adminRooms.POST("/:roomID/schedule/create", handlers.RoomsHandler.CreateSchedule)
		}
	}

	adminBookings := router.Group("/bookings", middleware.RequireRole(RoleAdmin))
	{
		adminBookings.GET("/list", handlers.BookingHandler.ListAllBookings)
	}

	userBookings := router.Group("/bookings", middleware.RequireRole(RoleUser))
	{
		userBookings.POST("/create", handlers.BookingHandler.CreateBooking)
		userBookings.GET("/my", handlers.BookingHandler.ListMyBookings)
		userBookings.POST("/:bookingID/cancel", handlers.BookingHandler.CancelBooking)
	}
}

func SetupAuthRoutes(router *gin.RouterGroup, handlers *handler.AuthHandler) {
	api := router.Group("/auth")
	{
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)
		api.POST("/dummyLogin", handlers.DummyLogin)
	}

}
