package mocks

import (
	"context"
	"internship/internal/domain/entity"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockAuthRepository мок для AuthRepositoryInterface
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockAuthRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// MockRoomsRepository мок для RoomsRepositoryInterface
type MockRoomsRepository struct {
	mock.Mock
}

func (m *MockRoomsRepository) CreateRoom(ctx context.Context, room *entity.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomsRepository) ListRooms(ctx context.Context) ([]*entity.Room, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Room), args.Error(1)
}

func (m *MockRoomsRepository) GetRoomByID(ctx context.Context, id string) (*entity.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Room), args.Error(1)
}

func (m *MockRoomsRepository) CreateSchedule(ctx context.Context, schedule *entity.Schedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockRoomsRepository) GetScheduleByRoomID(ctx context.Context, roomID string) (*entity.Schedule, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Schedule), args.Error(1)
}

func (m *MockRoomsRepository) IsExistsSchedule(ctx context.Context, roomID string) (bool, error) {
	args := m.Called(ctx, roomID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoomsRepository) GetAvailableSlots(ctx context.Context, roomID, date string) ([]*entity.Slot, error) {
	args := m.Called(ctx, roomID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Slot), args.Error(1)
}

func (m *MockRoomsRepository) GetBookedSlots(ctx context.Context, roomID, date string) ([]*entity.Slot, error) {
	args := m.Called(ctx, roomID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Slot), args.Error(1)
}

func (m *MockRoomsRepository) GetSlotByID(ctx context.Context, id string) (*entity.Slot, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Slot), args.Error(1)
}

func (m *MockRoomsRepository) IsAvailableSlot(ctx context.Context, slotID string) (bool, error) {
	args := m.Called(ctx, slotID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoomsRepository) UpsertSlots(ctx context.Context, slots []*entity.Slot) error {
	args := m.Called(ctx, slots)
	return args.Error(0)
}

// MockBookingRepository мок для BookingRepositoryInterface
type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) CreateBooking(ctx context.Context, booking *entity.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingRepository) GetBookingByID(ctx context.Context, id string) (*entity.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Booking), args.Error(1)
}

func (m *MockBookingRepository) ListAllBookings(ctx context.Context, page, pageSize int) ([]*entity.Booking, int, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entity.Booking), args.Int(1), args.Error(2)
}

func (m *MockBookingRepository) ListMyBookings(ctx context.Context, userID string, nowStr time.Time) ([]*entity.Booking, error) {
	args := m.Called(ctx, userID, nowStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Booking), args.Error(1)
}

func (m *MockBookingRepository) CancelBooking(ctx context.Context, id string) (*entity.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Booking), args.Error(1)
}

// MockJWTManager мок для JWTManagerInterface
type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateToken(userID, role string) (string, error) {
	args := m.Called(userID, role)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ParseToken(token string) (string, string, error) {
	args := m.Called(token)
	return args.String(0), args.String(1), args.Error(2)
}

// MockConferenceClient мок для ConferenceClientInterface
type MockConferenceClient struct {
	mock.Mock
}

func (m *MockConferenceClient) CreateLink(ctx context.Context, bookingID string) (string, error) {
	args := m.Called(ctx, bookingID)
	return args.String(0), args.Error(1)
}

// MockTransactionManager мок для TransactionManagerInterface
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)

	// Проверяем, есть ли возвращаемая ошибка
	if args.Error(0) != nil {
		return args.Error(0)
	}

	// Если функция передана и нужно ее выполнить
	if fn != nil {
		// Создаем контекст для транзакции (можно использовать переданный)
		txCtx := ctx
		return fn(txCtx)
	}

	return nil
}
