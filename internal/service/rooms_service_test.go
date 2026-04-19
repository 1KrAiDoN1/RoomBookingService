package service

import (
	"context"
	"internship/internal/domain"
	"internship/internal/domain/entity"
	"internship/internal/service/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestRoomsService_CreateRoom(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name        string
		roomName    string
		description string
		capacity    int
		setupMocks  func(*mocks.MockRoomsRepository)
		expectedErr error
	}{
		{
			name:        "Успешное создание комнаты",
			roomName:    "Конференц-зал",
			description: "Большой зал",
			capacity:    50,
			setupMocks: func(roomsRepo *mocks.MockRoomsRepository) {
				roomsRepo.On("CreateRoom", mock.Anything, mock.AnythingOfType("*entity.Room")).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка при создании комнаты",
			roomName:    "Тестовая комната",
			description: "Тест",
			capacity:    10,
			setupMocks: func(roomsRepo *mocks.MockRoomsRepository) {
				roomsRepo.On("CreateRoom", mock.Anything, mock.AnythingOfType("*entity.Room")).
					Return(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoomsRepo := new(mocks.MockRoomsRepository)
			mockTrManager := new(mocks.MockTransactionManager)

			tt.setupMocks(mockRoomsRepo)

			service := NewRoomsService(mockRoomsRepo, mockTrManager, logger)

			room, err := service.CreateRoom(context.Background(), tt.roomName, tt.description, tt.capacity)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, room)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, room)
				assert.Equal(t, tt.roomName, room.Name)
				assert.Equal(t, tt.description, room.Description)
				assert.Equal(t, tt.capacity, room.Capacity)
				assert.NotEmpty(t, room.ID)
			}

			mockRoomsRepo.AssertExpectations(t)
		})
	}
}

func TestRoomsService_CreateSchedule(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name        string
		roomID      string
		daysOfWeek  []int
		startTime   string
		endTime     string
		setupMocks  func(*mocks.MockRoomsRepository, *mocks.MockTransactionManager)
		expectedErr error
	}{
		{
			name:       "Успешное создание расписания",
			roomID:     uuid.New().String(),
			daysOfWeek: []int{1, 2, 3, 4, 5},
			startTime:  "09:00",
			endTime:    "18:00",
			setupMocks: func(roomsRepo *mocks.MockRoomsRepository, trManager *mocks.MockTransactionManager) {
				room := &entity.Room{
					ID:   uuid.New().String(),
					Name: "Тестовая комната",
				}
				roomsRepo.On("GetRoomByID", mock.Anything, mock.AnythingOfType("string")).
					Return(room, nil)
				roomsRepo.On("IsExistsSchedule", mock.Anything, mock.AnythingOfType("string")).
					Return(false, nil)
				roomsRepo.On("CreateSchedule", mock.Anything, mock.AnythingOfType("*entity.Schedule")).
					Return(nil)
				trManager.On("Do", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:       "Расписание уже существует",
			roomID:     uuid.New().String(),
			daysOfWeek: []int{1, 2, 3},
			startTime:  "10:00",
			endTime:    "19:00",
			setupMocks: func(roomsRepo *mocks.MockRoomsRepository, trManager *mocks.MockTransactionManager) {
				room := &entity.Room{
					ID:   uuid.New().String(),
					Name: "Тестовая комната",
				}
				roomsRepo.On("GetRoomByID", mock.Anything, mock.AnythingOfType("string")).
					Return(room, nil)
				roomsRepo.On("IsExistsSchedule", mock.Anything, mock.AnythingOfType("string")).
					Return(true, nil)
			},
			expectedErr: domain.ErrorScheduleExists,
		},
		{
			name:       "Некорректные дни недели",
			roomID:     uuid.New().String(),
			daysOfWeek: []int{0, 8, 9},
			startTime:  "09:00",
			endTime:    "18:00",
			setupMocks: func(roomsRepo *mocks.MockRoomsRepository, trManager *mocks.MockTransactionManager) {
				room := &entity.Room{
					ID:   uuid.New().String(),
					Name: "Тестовая комната",
				}
				roomsRepo.On("GetRoomByID", mock.Anything, mock.AnythingOfType("string")).
					Return(room, nil)
			},
			expectedErr: domain.ErrorInvalidRequest,
		},
		{
			name:       "Некорректное время (start >= end)",
			roomID:     uuid.New().String(),
			daysOfWeek: []int{1, 2, 3},
			startTime:  "18:00",
			endTime:    "09:00",
			setupMocks: func(roomsRepo *mocks.MockRoomsRepository, trManager *mocks.MockTransactionManager) {
				room := &entity.Room{
					ID:   uuid.New().String(),
					Name: "Тестовая комната",
				}
				roomsRepo.On("GetRoomByID", mock.Anything, mock.AnythingOfType("string")).
					Return(room, nil)
			},
			expectedErr: domain.ErrorInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoomsRepo := new(mocks.MockRoomsRepository)
			mockTrManager := new(mocks.MockTransactionManager)

			tt.setupMocks(mockRoomsRepo, mockTrManager)

			service := NewRoomsService(mockRoomsRepo, mockTrManager, logger)

			schedule, err := service.CreateSchedule(context.Background(), tt.roomID, tt.daysOfWeek, tt.startTime, tt.endTime)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, schedule)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, schedule)
				assert.Equal(t, tt.roomID, schedule.RoomID)
				assert.Equal(t, tt.daysOfWeek, schedule.DaysOfWeek)
				assert.Equal(t, tt.startTime, schedule.StartTime)
				assert.Equal(t, tt.endTime, schedule.EndTime)
			}

			mockRoomsRepo.AssertExpectations(t)
		})
	}
}
