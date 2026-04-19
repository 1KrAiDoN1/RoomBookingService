package app

import (
	"context"
	"fmt"
	"internship/internal/config"
	"internship/internal/service"
	"internship/internal/service/jwt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	httpserver "internship/internal/http-server"
	"internship/internal/http-server/handler"
	"internship/internal/repository/postgres"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func Run(ctx context.Context, log *zap.Logger, config config.ServiceConfig) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	storage, err := postgres.NewDatabase(ctx, config.DBConfig.DBConn)
	if err != nil {
		log.Error("failed to connect to database", zap.Error(err))
		return fmt.Errorf("database connection failed: %w", err)
	}
	log.Info("connected to database", zap.String("dsn", config.DBConfig.DBConn))

	dbpool := storage.GetPool()
	defer func() {
		log.Info("closing database connection...")
		dbpool.Close()
		log.Info("database connection closed")
	}()

	getter := trmpgx.DefaultCtxGetter

	trManager := manager.Must(
		trmpgx.NewDefaultFactory(dbpool),
	)

	autRepo := postgres.NewAuthRepository(dbpool, getter)
	roomsRepo := postgres.NewRoomsRepository(dbpool, getter)
	bookingRepo := postgres.NewBookingRepository(dbpool, getter)

	jwtManager := jwt.NewJWTManager(config.AuthConfig.SecretKey, config.AuthConfig.TokenTTL)

	authService := service.NewAuthService(autRepo, jwtManager, log)
	conferenceService := service.NewMockConferenceClient()
	bookingService := service.NewBookingService(bookingRepo, roomsRepo, trManager, conferenceService, log)
	roomsService := service.NewRoomsService(roomsRepo, trManager, log)

	handlers := handler.NewHandlers(authService, bookingService, roomsService, log)

	server := httpserver.NewServer(log, config, jwtManager, handlers)

	serverDone := make(chan error, 1)
	go func() {
		log.Info("starting HTTP server...")
		if err := server.Run(); err != nil {
			serverDone <- err
		}
		close(serverDone)
	}()

	select {
	case sig := <-sigChan:
		log.Info("received shutdown signal", zap.String("signal", sig.String()))
		cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error("failed to shutdown HTTP server", zap.Error(err))
		}

		log.Info("waiting for goroutines to finish...")

		log.Info("application gracefully shut down")
		return nil

	case err := <-serverDone:
		if err != nil {
			log.Error("http server stopped with error", zap.Error(err))
			cancel()
			return err
		}
		return nil
	}
}
