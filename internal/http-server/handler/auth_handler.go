package handler

import (
	"context"
	"internship/internal/domain"
	"internship/internal/domain/entity"
	dto "internship/internal/models/dto/handler"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService AuthServiceInterface
	logger      *zap.Logger
}

func NewAuthHandler(authService AuthServiceInterface,
	logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// DummyLogin godoc
// @Summary      Тестовый логин
// @Description  Создает токен на основе переданной роли без пароля
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.DummyLoginRequest true "Роль пользователя (admin, user)"
// @Success      200  {object}  dto.TokenResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  InternalErrorResponse
// @Router       /auth/dummyLogin [post]
func (a *AuthHandler) DummyLogin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	var req dto.DummyLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("dummyLogin: bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "invalid request body",
		})
		return
	}

	if req.Role != entity.RoleAdmin && req.Role != entity.RoleUser {
		a.logger.Error("dummyLogin: invalid role", zap.String("role", req.Role))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "role must be 'admin' or 'user'",
		})
		return
	}

	token, err := a.authService.DummyLogin(ctx, req.Role)
	if err != nil {
		a.logger.Error("dummyLogin: service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, InternalErrorResponse{
			Code:    domain.ErrInternalError,
			Message: "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{Token: token})
}

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Создает нового пользователя в системе
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Данные для регистрации"
// @Success      201  {object}  dto.RegisterResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  InternalErrorResponse
// @Router       /auth/register [post]
func (a *AuthHandler) Register(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("register: bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "invalid request body",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "request body is required",
		})
		return
	}
	if req.Role != entity.RoleAdmin && req.Role != entity.RoleUser {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "invalid request body",
		})
		return
	}

	user, err := a.authService.Register(ctx, req.Email, req.Password, req.Role)
	if err != nil {
		a.logger.Error("register: service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, InternalErrorResponse{
			Code:    domain.ErrInternalError,
			Message: "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.RegisterResponse{
		User: dto.UserData{
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Login godoc
// @Summary      Авторизация
// @Description  Вход по email и паролю для получения токена
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Креды пользователя"
// @Success      200  {object}  dto.TokenResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /auth/login [post]
func (a *AuthHandler) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), ContextTimeout)
	defer cancel()

	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("login: bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "invalid request body",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		a.logger.Error("login: missing fields", zap.String("email", req.Email))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    domain.ErrInvalidRequest,
			Message: "email and password are required",
		})
		return
	}

	token, err := a.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		a.logger.Error("login: failed", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    domain.ErrUnauthorized,
			Message: "invalid email or password",
		})
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{Token: token})
}
