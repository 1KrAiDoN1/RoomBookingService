package service

import (
	"context"
	"internship/internal/domain"
	"internship/internal/domain/entity"
	"internship/internal/service/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestBookingService_CreateBooking(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name                 string
		userID               string
		slotID               string
		createConferenceLink bool
		setupMocks           func(*mocks.MockBookingRepository, *mocks.MockRoomsRepository, *mocks.MockTransactionManager, *mocks.MockConferenceClient)
		expectedErr          error
	}{
		{
			name:                 "Успешное создание бронирования",
			userID:               uuid.New().String(),
			slotID:               uuid.New().String(),
			createConferenceLink: false,
			setupMocks: func(bookingRepo *mocks.MockBookingRepository, roomsRepo *mocks.MockRoomsRepository, trManager *mocks.MockTransactionManager, confClient *mocks.MockConferenceClient) {
				futureTime := time.Now().UTC().Add(24 * time.Hour)
				slot := &entity.Slot{
					ID:    uuid.New().String(),
					Start: futureTime.Format(time.RFC3339),
					End:   futureTime.Add(30 * time.Minute).Format(time.RFC3339),
				}
				roomsRepo.On("GetSlotByID", mock.Anything, mock.AnythingOfType("string")).
					Return(slot, nil)
				roomsRepo.On("IsAvailableSlot", mock.Anything, mock.AnythingOfType("string")).
					Return(true, nil)
				bookingRepo.On("CreateBooking", mock.Anything, mock.AnythingOfType("*entity.Booking")).
					Return(nil)
				trManager.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:                 "Слот не найден",
			userID:               uuid.New().String(),
			slotID:               uuid.New().String(),
			createConferenceLink: false,
			setupMocks: func(bookingRepo *mocks.MockBookingRepository, roomsRepo *mocks.MockRoomsRepository, trManager *mocks.MockTransactionManager, confClient *mocks.MockConferenceClient) {
				roomsRepo.On("GetSlotByID", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, domain.ErrorSlotNotFound)
			},
			expectedErr: domain.ErrorSlotNotFound,
		},
		{
			name:                 "Слот уже забронирован",
			userID:               uuid.New().String(),
			slotID:               uuid.New().String(),
			createConferenceLink: false,
			setupMocks: func(bookingRepo *mocks.MockBookingRepository, roomsRepo *mocks.MockRoomsRepository, trManager *mocks.MockTransactionManager, confClient *mocks.MockConferenceClient) {
				futureTime := time.Now().UTC().Add(24 * time.Hour)
				slot := &entity.Slot{
					ID:    uuid.New().String(),
					Start: futureTime.Format(time.RFC3339),
					End:   futureTime.Add(30 * time.Minute).Format(time.RFC3339),
				}
				roomsRepo.On("GetSlotByID", mock.Anything, mock.AnythingOfType("string")).
					Return(slot, nil)
				roomsRepo.On("IsAvailableSlot", mock.Anything, mock.AnythingOfType("string")).
					Return(false, nil)
			},
			expectedErr: domain.ErrorSlotAlreadyBooked,
		},
		{
			name:                 "Попытка бронировать прошедший слот",
			userID:               uuid.New().String(),
			slotID:               uuid.New().String(),
			createConferenceLink: false,
			setupMocks: func(bookingRepo *mocks.MockBookingRepository, roomsRepo *mocks.MockRoomsRepository, trManager *mocks.MockTransactionManager, confClient *mocks.MockConferenceClient) {
				pastTime := time.Now().UTC().Add(-24 * time.Hour)
				slot := &entity.Slot{
					ID:    uuid.New().String(),
					Start: pastTime.Format(time.RFC3339),
					End:   pastTime.Add(30 * time.Minute).Format(time.RFC3339),
				}
				roomsRepo.On("GetSlotByID", mock.Anything, mock.AnythingOfType("string")).
					Return(slot, nil)
			},
			expectedErr: domain.ErrorInvalidRequest,
		},
		{
			name:                 "Создание бронирования с конференц-ссылкой",
			userID:               uuid.New().String(),
			slotID:               uuid.New().String(),
			createConferenceLink: true,
			setupMocks: func(bookingRepo *mocks.MockBookingRepository, roomsRepo *mocks.MockRoomsRepository, trManager *mocks.MockTransactionManager, confClient *mocks.MockConferenceClient) {
				futureTime := time.Now().UTC().Add(24 * time.Hour)
				slot := &entity.Slot{
					ID:    uuid.New().String(),
					Start: futureTime.Format(time.RFC3339),
					End:   futureTime.Add(30 * time.Minute).Format(time.RFC3339),
				}
				roomsRepo.On("GetSlotByID", mock.Anything, mock.AnythingOfType("string")).
					Return(slot, nil)
				roomsRepo.On("IsAvailableSlot", mock.Anything, mock.AnythingOfType("string")).
					Return(true, nil)
				bookingRepo.On("CreateBooking", mock.Anything, mock.AnythingOfType("*entity.Booking")).
					Return(nil)
				confClient.On("CreateLink", mock.Anything, mock.AnythingOfType("string")).
					Return("https://meet.example.com/123", nil)
				trManager.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBookingRepo := new(mocks.MockBookingRepository)
			mockRoomsRepo := new(mocks.MockRoomsRepository)
			mockTrManager := new(mocks.MockTransactionManager)
			mockConfClient := new(mocks.MockConferenceClient)

			tt.setupMocks(mockBookingRepo, mockRoomsRepo, mockTrManager, mockConfClient)

			service := NewBookingService(mockBookingRepo, mockRoomsRepo, mockTrManager, mockConfClient, logger)

			booking, err := service.CreateBooking(context.Background(), tt.userID, tt.slotID, tt.createConferenceLink)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, booking)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, booking)
				assert.Equal(t, tt.userID, booking.UserID)
				assert.Equal(t, tt.slotID, booking.SlotID)
				assert.Equal(t, entity.BookingStatusActive, booking.Status)

				if tt.createConferenceLink {
					assert.NotEmpty(t, booking.ConferenceLink)
				}
			}

			mockBookingRepo.AssertExpectations(t)
			mockRoomsRepo.AssertExpectations(t)
		})
	}
}

func TestBookingService_CancelBooking(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name        string
		bookingID   string
		userID      string
		setupMocks  func(*mocks.MockBookingRepository)
		expectedErr error
	}{
		{
			name:      "Успешная отмена бронирования",
			bookingID: uuid.New().String(),
			userID:    uuid.New().String(),
			setupMocks: func(bookingRepo *mocks.MockBookingRepository) {
				booking := &entity.Booking{
					ID:     uuid.New().String(),
					UserID: uuid.New().String(),
					Status: entity.BookingStatusActive,
				}
				bookingRepo.On("GetBookingByID", mock.Anything, mock.AnythingOfType("string")).
					Return(booking, nil)

				cancelledBooking := &entity.Booking{
					ID:     booking.ID,
					UserID: booking.UserID,
					Status: entity.BookingStatusCancelled,
				}
				bookingRepo.On("CancelBooking", mock.Anything, mock.AnythingOfType("string")).
					Return(cancelledBooking, nil)
			},
			expectedErr: nil,
		},
		{
			name:      "Отмена уже отмененной брони (идемпотентность)",
			bookingID: uuid.New().String(),
			userID:    uuid.New().String(),
			setupMocks: func(bookingRepo *mocks.MockBookingRepository) {
				booking := &entity.Booking{
					ID:     uuid.New().String(),
					UserID: uuid.New().String(),
					Status: entity.BookingStatusCancelled,
				}
				bookingRepo.On("GetBookingByID", mock.Anything, mock.AnythingOfType("string")).
					Return(booking, nil)
			},
			expectedErr: nil,
		},
		{
			name:      "Бронирование не найдено",
			bookingID: uuid.New().String(),
			userID:    uuid.New().String(),
			setupMocks: func(bookingRepo *mocks.MockBookingRepository) {
				bookingRepo.On("GetBookingByID", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, domain.ErrorBookingNotFound)
			},
			expectedErr: domain.ErrorBookingNotFound,
		},
		{
			name:      "Попытка отменить чужую бронь",
			bookingID: uuid.New().String(),
			userID:    "other-user-id",
			setupMocks: func(bookingRepo *mocks.MockBookingRepository) {
				booking := &entity.Booking{
					ID:     uuid.New().String(),
					UserID: "owner-user-id",
					Status: entity.BookingStatusActive,
				}
				bookingRepo.On("GetBookingByID", mock.Anything, mock.AnythingOfType("string")).
					Return(booking, nil)
			},
			expectedErr: domain.ErrorForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBookingRepo := new(mocks.MockBookingRepository)
			mockRoomsRepo := new(mocks.MockRoomsRepository)
			mockTrManager := new(mocks.MockTransactionManager)
			mockConfClient := new(mocks.MockConferenceClient)

			tt.setupMocks(mockBookingRepo)

			service := NewBookingService(mockBookingRepo, mockRoomsRepo, mockTrManager, mockConfClient, logger)

			booking, err := service.CancelBooking(context.Background(), tt.bookingID, tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, booking)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, booking)
				assert.Equal(t, entity.BookingStatusCancelled, booking.Status)
			}

			mockBookingRepo.AssertExpectations(t)
		})
	}
}

func TestBookingService_ListMyBookings(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name        string
		userID      string
		setupMocks  func(*mocks.MockBookingRepository)
		expectedLen int
	}{
		{
			name:   "Получение своих броней",
			userID: uuid.New().String(),
			setupMocks: func(bookingRepo *mocks.MockBookingRepository) {
				bookings := []*entity.Booking{
					{
						ID:     uuid.New().String(),
						UserID: uuid.New().String(),
						Status: entity.BookingStatusActive,
					},
				}
				bookingRepo.On("ListMyBookings", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(bookings, nil)
			},
			expectedLen: 1,
		},
		{
			name:   "Пустой список броней",
			userID: uuid.New().String(),
			setupMocks: func(bookingRepo *mocks.MockBookingRepository) {
				bookingRepo.On("ListMyBookings", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return([]*entity.Booking{}, nil)
			},
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBookingRepo := new(mocks.MockBookingRepository)
			mockRoomsRepo := new(mocks.MockRoomsRepository)
			mockTrManager := new(mocks.MockTransactionManager)
			mockConfClient := new(mocks.MockConferenceClient)

			tt.setupMocks(mockBookingRepo)

			service := NewBookingService(mockBookingRepo, mockRoomsRepo, mockTrManager, mockConfClient, logger)

			bookings, err := service.ListMyBookings(context.Background(), tt.userID)

			assert.NoError(t, err)
			assert.Len(t, bookings, tt.expectedLen)

			mockBookingRepo.AssertExpectations(t)
		})
	}
}
