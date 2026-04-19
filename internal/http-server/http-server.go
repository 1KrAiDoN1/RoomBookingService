package httpserver

import (
	"context"
	"errors"
	"fmt"
	"internship/internal/config"
	"internship/internal/http-server/handler"
	"internship/internal/http-server/middleware"
	"internship/internal/http-server/routes"
	"net/http"

	_ "internship/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type Server struct {
	server     *http.Server
	router     *gin.Engine
	handlers   *handler.Handlers
	jwtMANAGER middleware.TokenParser
	logger     *zap.Logger
	config     config.ServiceConfig
}

func NewServer(logger *zap.Logger, cfg config.ServiceConfig, jwtMANAGER middleware.TokenParser, handlers *handler.Handlers) *Server {
	router := gin.New()

	srv := &http.Server{
		Addr:         cfg.ServerConfig.Address,
		Handler:      router,
		ReadTimeout:  cfg.ServerConfig.ReadTimeout,
		WriteTimeout: cfg.ServerConfig.WriteTimeout,
		IdleTimeout:  cfg.ServerConfig.IdleTimeout,
	}

	return &Server{
		server:     srv,
		router:     router,
		handlers:   handlers,
		logger:     logger,
		config:     cfg,
		jwtMANAGER: jwtMANAGER,
	}
}

func (s *Server) Run() error {
	s.setupRoutes()

	s.logger.Info("starting HTTP server", zap.String("address", s.config.ServerConfig.Address))

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx_shutdown, cancel := context.WithTimeout(ctx, s.config.ServerConfig.ShutdownTimeout)
	defer cancel()

	s.logger.Info("shutting down HTTP server...")
	if err := s.server.Shutdown(ctx_shutdown); err != nil {
		s.logger.Error("server forced to shutdown", zap.Error(err))
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	s.logger.Info("http server gracefully shut down")
	return nil
}

func (s *Server) setupRoutes() {
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := s.router.Group("")
	api.Use(middleware.Logger(s.logger))
	api.GET("/_info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	routes.SetupAuthRoutes(api, s.handlers.AuthHandler)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(s.jwtMANAGER, s.logger))
	{
		routes.SetupBookingServiceRoutes(protected, s.handlers)
	}
}
