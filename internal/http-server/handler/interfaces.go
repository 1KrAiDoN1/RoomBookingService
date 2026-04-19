package handler

import (
	"context"
	"internship/internal/domain/entity"
)

type AuthServiceInterface interface {
	DummyLogin(ctx context.Context, role string) (string, error)
	Register(ctx context.Context, email, password, role string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type RoomsServiceInterface interface {
	CreateRoom(ctx context.Context, name, description string, capacity int) (*entity.Room, error)
	ListRooms(ctx context.Context) ([]*entity.Room, error)
	CreateSchedule(ctx context.Context, roomID string, daysOfWeek []int, startTime, endTime string) (*entity.Schedule, error)
	ListAvailableSlots(ctx context.Context, roomID, date string) ([]*entity.Slot, error)
}

type BookingServiceInterface interface {
	CreateBooking(ctx context.Context, userID, slotID string, createConferenceLink bool) (*entity.Booking, error)
	ListAllBookings(ctx context.Context, page, pageSize int) ([]*entity.Booking, *entity.Pagination, error)
	ListMyBookings(ctx context.Context, userID string) ([]*entity.Booking, error)
	CancelBooking(ctx context.Context, bookingID, userID string) (*entity.Booking, error)
}
