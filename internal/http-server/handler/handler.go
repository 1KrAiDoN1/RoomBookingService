package handler

import (
	"time"

	"go.uber.org/zap"
)

const (
	ContextTimeout = 2 * time.Second
)

type Handlers struct {
	AuthHandler    *AuthHandler
	BookingHandler *BookingHandler
	RoomsHandler   *RoomsHandler
}

func NewHandlers(authService AuthServiceInterface, bookingService BookingServiceInterface, roomsService RoomsServiceInterface, log *zap.Logger) *Handlers {
	return &Handlers{
		AuthHandler:    NewAuthHandler(authService, log),
		BookingHandler: NewBookingHandler(bookingService, log),
		RoomsHandler:   NewRoomsHandler(roomsService, log),
	}
}
