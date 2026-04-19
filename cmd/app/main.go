package main

import (
	"context"
	"fmt"

	_ "internship/docs"
	"internship/internal/app"
	"internship/internal/config"
	zapplogger "internship/pkg/lib/logger/zaplogger"
	"os"

	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()
	log := zapplogger.SetupLoggerWithLevel(zapcore.DebugLevel)
	log.Info("Service started")
	config, err := config.LoadServiceConfig()
	fmt.Println(config)
	if err != nil {
		log.Error("failed to load service config", zapplogger.Err(err))
		os.Exit(1)
	}

	if err := app.Run(ctx, log, config); err != nil {
		log.Error("failed to run application service", zapplogger.Err(err))
		os.Exit(1)
	}

}
