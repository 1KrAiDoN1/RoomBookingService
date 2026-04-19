package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

func LoadServiceConfig() (ServiceConfig, error) {
	var serviceConfig ServiceConfig

	if err := cleanenv.ReadConfig(".env", &serviceConfig); err != nil {
		if err := cleanenv.ReadEnv(&serviceConfig); err != nil {
			return ServiceConfig{}, fmt.Errorf("failed to read environment variables: %w", err)
		}
	}
	if err := validator.New().Struct(&serviceConfig); err != nil {
		return serviceConfig, fmt.Errorf("config validation failed: %w", err)
	}
	dbConnStr := DSN(serviceConfig)
	serviceConfig.DBConfig.DBConn = dbConnStr

	return serviceConfig, nil
}

func DSN(cfg ServiceConfig) string {
	dsn := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBConfig.Driver, cfg.DBConfig.User, cfg.DBConfig.Password, cfg.DBConfig.Host, cfg.DBConfig.Port, cfg.DBConfig.DBName)

	return dsn
}

type ServiceConfig struct {
	ServerConfig struct {
		Address         string        `env:"SERVER_ADDRESS" env-default:":8080" validate:"required"`
		ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" env-default:"15s" validate:"required"`
		WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" env-default:"15s" validate:"required"`
		IdleTimeout     time.Duration `env:"SERVER_IDLE_TIMEOUT" env-default:"120s" validate:"required"`
		ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" env-default:"30s" validate:"required"`
	}

	DBConfig struct {
		Driver   string `env:"DB_DRIVER" env-default:"postgres" validate:"required"`
		Host     string `env:"DB_HOST" env-default:"postgres" validate:"required"`
		Port     int    `env:"DB_PORT" env-default:"5432" validate:"required"`
		User     string `env:"DB_USER" env-default:"postgres" validate:"required"`
		Password string `env:"DB_PASSWORD" env-default:"admin" validate:"required"`
		DBName   string `env:"DB_NAME" env-default:"room_booking" validate:"required"`
		DBConn   string
	}
	AuthConfig struct {
		SecretKey string        `env:"JWT_SECRET_KEY" env-default:"ekerjvjhmzleptktysqpkfznzfwtewmff" validate:"required"`
		TokenTTL  time.Duration `env:"TOKEN_TTL" env-default:"24h" validate:"required"`
	}
}
