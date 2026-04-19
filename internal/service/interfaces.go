package service

import (
	"context"
	"internship/internal/domain/entity"
	"time"
)

type AuthRepositoryInterface interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
}

type RoomsRepositoryInterface interface {
	CreateRoom(ctx context.Context, room *entity.Room) error
	ListRooms(ctx context.Context) ([]*entity.Room, error)
	GetRoomByID(ctx context.Context, id string) (*entity.Room, error)
	CreateSchedule(ctx context.Context, schedule *entity.Schedule) error
	GetScheduleByRoomID(ctx context.Context, roomID string) (*entity.Schedule, error)
	IsExistsSchedule(ctx context.Context, roomID string) (bool, error)
	UpsertSlots(ctx context.Context, slots []*entity.Slot) error
	GetAvailableSlots(ctx context.Context, roomID, date string) ([]*entity.Slot, error)
	GetBookedSlots(ctx context.Context, roomID, date string) ([]*entity.Slot, error)
	GetSlotByID(ctx context.Context, id string) (*entity.Slot, error)
	IsAvailableSlot(ctx context.Context, slotID string) (bool, error)
}

type BookingRepositoryInterface interface {
	CreateBooking(ctx context.Context, booking *entity.Booking) error
	GetBookingByID(ctx context.Context, id string) (*entity.Booking, error)
	ListAllBookings(ctx context.Context, page, pageSize int) ([]*entity.Booking, int, error)
	ListMyBookings(ctx context.Context, userID string, time time.Time) ([]*entity.Booking, error)
	CancelBooking(ctx context.Context, id string) (*entity.Booking, error)
}

type JWTManagerInterface interface {
	GenerateToken(userID string, role string) (string, error)
	ParseToken(tokenStr string) (string, string, error)
}

type ConferenceClientInterface interface {
	CreateLink(ctx context.Context, bookingID string) (string, error)
}

type TransactionManagerInterface interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=mocks/auth_mocks.go -package=mocks internship/internal/service AuthRepositoryInterface,JWTManagerInterface
