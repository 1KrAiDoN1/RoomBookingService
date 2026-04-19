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

func TestAuthService_Register(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name        string
		email       string
		password    string
		role        string
		setupMocks  func(*mocks.MockAuthRepository)
		expectedErr error
	}{
		{
			name:     "Успешная регистрация",
			email:    "test@example.com",
			password: "password123",
			role:     entity.RoleUser,
			setupMocks: func(authRepo *mocks.MockAuthRepository) {
				authRepo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(nil, domain.ErrorNotFound)
				authRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*entity.User")).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:     "Пользователь уже существует",
			email:    "existing@example.com",
			password: "password123",
			role:     entity.RoleUser,
			setupMocks: func(authRepo *mocks.MockAuthRepository) {
				existingUser := &entity.User{
					ID:    uuid.New().String(),
					Email: "existing@example.com",
					Role:  entity.RoleUser,
				}
				authRepo.On("GetUserByEmail", mock.Anything, "existing@example.com").
					Return(existingUser, nil)
			},
			expectedErr: domain.ErrorUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := new(mocks.MockAuthRepository)
			mockJWTManager := new(mocks.MockJWTManager)

			tt.setupMocks(mockAuthRepo)

			service := NewAuthService(mockAuthRepo, mockJWTManager, logger)

			user, err := service.Register(context.Background(), tt.email, tt.password, tt.role)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.role, user.Role)
				assert.NotEmpty(t, user.ID)
			}

			mockAuthRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_DummyLogin(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockAuthRepo := new(mocks.MockAuthRepository)
	mockJWTManager := new(mocks.MockJWTManager)

	service := NewAuthService(mockAuthRepo, mockJWTManager, logger)

	tests := []struct {
		name        string
		role        string
		expectedErr error
	}{
		{
			name:        "Admin login",
			role:        entity.RoleAdmin,
			expectedErr: nil,
		},
		{
			name:        "User login",
			role:        entity.RoleUser,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.DummyLogin(context.Background(), tt.role)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}
