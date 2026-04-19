package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"internship/internal/domain"
	"internship/internal/domain/entity"
	"internship/internal/service/mocks"
)

func TestAuthService_DummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockAuthRepositoryInterface(ctrl)
	jwtMock := mocks.NewMockJWTManagerInterface(ctrl)
	svc := NewAuthService(repo, jwtMock, zap.NewNop())

	t.Run("admin token", func(t *testing.T) {
		token, err := svc.DummyLogin(context.Background(), entity.RoleAdmin)
		require.NoError(t, err)
		require.Equal(t, testing_admin_role, token)
	})
	t.Run("user token", func(t *testing.T) {
		token, err := svc.DummyLogin(context.Background(), entity.RoleUser)
		require.NoError(t, err)
		require.Equal(t, testing_user_role, token)
	})
	t.Run("unknown role returns empty token", func(t *testing.T) {
		token, err := svc.DummyLogin(context.Background(), "guest")
		require.NoError(t, err)
		require.Empty(t, token)
	})
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()
	email := "user@example.com"
	password := "secret-password"
	role := entity.RoleUser

	t.Run("user already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockAuthRepositoryInterface(ctrl)
		jwtMock := mocks.NewMockJWTManagerInterface(ctrl)
		svc := NewAuthService(repo, jwtMock, zap.NewNop())

		repo.EXPECT().GetUserByEmail(ctx, email).Return(&entity.User{Email: email}, nil)

		user, err := svc.Register(ctx, email, password, role)
		require.ErrorIs(t, err, domain.ErrorUserAlreadyExists)
		require.Nil(t, user)
	})

	t.Run("create user fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockAuthRepositoryInterface(ctrl)
		jwtMock := mocks.NewMockJWTManagerInterface(ctrl)
		svc := NewAuthService(repo, jwtMock, zap.NewNop())

		repo.EXPECT().GetUserByEmail(ctx, email).Return(nil, domain.ErrorNotFound)
		repo.EXPECT().CreateUser(ctx, gomock.Any()).Return(errors.New("db down"))

		user, err := svc.Register(ctx, email, password, role)
		require.Error(t, err)
		require.Nil(t, user)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockAuthRepositoryInterface(ctrl)
		jwtMock := mocks.NewMockJWTManagerInterface(ctrl)
		svc := NewAuthService(repo, jwtMock, zap.NewNop())

		repo.EXPECT().GetUserByEmail(ctx, email).Return(nil, domain.ErrorNotFound)
		repo.EXPECT().CreateUser(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, u *entity.User) error {
			require.Equal(t, email, u.Email)
			require.Equal(t, role, u.Role)
			require.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)))
			return nil
		})

		user, err := svc.Register(ctx, email, password, role)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, email, user.Email)
		require.Equal(t, role, user.Role)
	})
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	email := "login@example.com"
	password := "correct-horse"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	storedUser := &entity.User{
		ID:           "user-id-1",
		Email:        email,
		PasswordHash: string(hash),
		Role:         entity.RoleUser,
	}

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockAuthRepositoryInterface(ctrl)
		jwtMock := mocks.NewMockJWTManagerInterface(ctrl)
		svc := NewAuthService(repo, jwtMock, zap.NewNop())

		repo.EXPECT().GetUserByEmail(ctx, email).Return(nil, errors.New("db error"))

		token, err := svc.Login(ctx, email, password)
		require.Error(t, err)
		require.Empty(t, token)
	})

	t.Run("wrong password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockAuthRepositoryInterface(ctrl)
		jwtMock := mocks.NewMockJWTManagerInterface(ctrl)
		svc := NewAuthService(repo, jwtMock, zap.NewNop())

		repo.EXPECT().GetUserByEmail(ctx, email).Return(storedUser, nil)

		token, err := svc.Login(ctx, email, "wrong")
		require.ErrorIs(t, err, domain.ErrorInvalidRequest)
		require.Empty(t, token)
	})

	t.Run("jwt error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockAuthRepositoryInterface(ctrl)
		jwtMock := mocks.NewMockJWTManagerInterface(ctrl)
		svc := NewAuthService(repo, jwtMock, zap.NewNop())

		repo.EXPECT().GetUserByEmail(ctx, email).Return(storedUser, nil)
		jwtMock.EXPECT().GenerateToken(storedUser.ID, storedUser.Role).Return("", errors.New("jwt failed"))

		token, err := svc.Login(ctx, email, password)
		require.Error(t, err)
		require.Empty(t, token)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockAuthRepositoryInterface(ctrl)
		jwtMock := mocks.NewMockJWTManagerInterface(ctrl)
		svc := NewAuthService(repo, jwtMock, zap.NewNop())

		repo.EXPECT().GetUserByEmail(ctx, email).Return(storedUser, nil)
		jwtMock.EXPECT().GenerateToken(storedUser.ID, storedUser.Role).Return("signed-jwt", nil)

		token, err := svc.Login(ctx, email, password)
		require.NoError(t, err)
		require.Equal(t, "signed-jwt", token)
	})
}
