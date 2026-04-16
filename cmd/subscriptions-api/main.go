package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"subscriptions-service/internal"
	"subscriptions-service/internal/application"
	"syscall"
	"time"

	_ "subscriptions-service/docs"
)

// @title Subscriptions Service API
// @version 1.0
// @description REST API for managing user subscriptions.
// @BasePath /api/v1
// @schemes http
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := internal.NewConfig[internal.AppConfig](".env")
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	container, err := application.NewContainer(ctx, cfg).Init(ctx)
	if err != nil {
		panic(fmt.Errorf("init container: %w", err))
	}

	publicServer, err := application.NewPublicServer(cfg, container.Logger).Configure(container)
	if err != nil {
		panic(fmt.Errorf("configure server: %w", err))
	}

	app := application.NewApp(publicServer, container)

	if err := app.Run(ctx); err != nil {
		container.Logger.Error("run app", "error", err)
	}

	container.Logger.Info("server started", "port", cfg.PublicServerConfig.Port)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutDown(shutdownCtx); err != nil {
		container.Logger.Error("shutdown app", "error", err)
	}
}
