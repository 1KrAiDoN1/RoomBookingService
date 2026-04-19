package service

import (
	"context"
	"fmt"
	"internship/internal/domain"
	"internship/internal/domain/entity"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	testing_admin_role = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJ1c2VyX2lkIjoiMDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAxIn0.2fNzBaLd9x7RAk6eEaS5Ie7YFlvUEixfBYX4Sbd1U6M"
	testing_user_role  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoidXNlciIsInVzZXJfaWQiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDIifQ.cgLevY14XS_kf42m1TYrFZ8fczj9lGdYhdnKQIJmK3Q"
)

type authService struct {
	authRepository AuthRepositoryInterface
	jwtManager     JWTManagerInterface
	logger         *zap.Logger
}

func NewAuthService(authRepository AuthRepositoryInterface, jwtManager JWTManagerInterface, logger *zap.Logger) *authService {
	return &authService{
		authRepository: authRepository,
		jwtManager:     jwtManager,
		logger:         logger,
	}
}

func (s *authService) DummyLogin(_ context.Context, role string) (string, error) {
	var token string
	switch role {
	case entity.RoleAdmin:
		token = testing_admin_role
	case entity.RoleUser:
		token = testing_user_role
	}
	return token, nil

}

func (s *authService) Register(ctx context.Context, email, password, role string) (*entity.User, error) {
	userdata, err := s.authRepository.GetUserByEmail(ctx, email)
	if err == nil && userdata != nil {
		return nil, domain.ErrorUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &entity.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.authRepository.CreateUser(ctx, user); err != nil {
		s.logger.Error("register: create user", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.authRepository.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error("login: find user", zap.String("email", email), zap.Error(err))
		return "", fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrorInvalidRequest
	}

	token, err := s.jwtManager.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}
