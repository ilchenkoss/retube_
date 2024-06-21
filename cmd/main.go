package main

import (
	"birthdayapp/internal/app"
	"birthdayapp/internal/config"
	"context"
	"log/slog"
	"os"
	"os/signal"
)

const (
	envDev  = "dev"
	envProd = "prod"
)

func main() {

	//main context for interrupt
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	cfg, lcErr := config.LoadConfig()
	if lcErr != nil {
		panic(lcErr)
	}
	log := initLogger(cfg.Env)
	log.Info("Config loaded")

	app.MustNew(ctx, log, *cfg)
}

func initLogger(envType string) *slog.Logger {
	var logger *slog.Logger
	switch envType {
	case envDev:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	case envProd:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	}
	return logger
}
