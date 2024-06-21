package app

import (
	"birthdayapp/internal/adapters"
	"birthdayapp/internal/adapters/database"
	"birthdayapp/internal/adapters/database/repository"
	"birthdayapp/internal/adapters/telegram"
	"birthdayapp/internal/adapters/telegram/handlers"
	"birthdayapp/internal/config"
	"birthdayapp/internal/core/service"
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

type App struct {
	slog *slog.Logger
	cfg  *config.Config
}

func MustNew(ctx context.Context, log *slog.Logger, cfg config.Config) {

	op := "App.New"
	log.With(slog.String("op", op))

	//init db connection
	dbConnection, cErr := database.NewConnection(&cfg)
	if cErr != nil {
		log.Debug("error db connection", "error", cErr)
		panic(cErr)
	}
	pErr := dbConnection.Ping()
	if pErr != nil {
		log.Debug("error ping db", "error", pErr)
		panic(pErr)
	}
	log.Info("DB connection OK")
	mErr := dbConnection.MakeMigrations()
	if mErr != nil {
		log.Debug("error make migrations", "error", mErr)
		panic(mErr)
	}
	log.Info("Migrations OK")

	tg, tgErr := telegram.NewTelegramBot(log, os.Getenv("TELEGRAM_TOKEN"))
	if tgErr != nil {
		log.Debug("error init telegram bot", "error", tgErr)
		panic(tgErr)
	}
	log.Info("Telegram bot running...")

	//dependencies injection
	userRepo := repository.NewUserRepository(dbConnection)
	subRepo := repository.NewSubscriptionsRepository(dbConnection)

	extApi := adapters.NewExternalAPI()
	birthdayService := service.NewBirthdayService(log, userRepo, tg, &cfg)
	userService := service.NewUserService(userRepo, extApi)
	subService := service.NewSubscriptionService(subRepo)

	subHandler := handlers.NewSubscriptionsHandler(subService, userService)
	middleware := handlers.NewMiddleware(userRepo)
	tgHandlers := telegram.Handlers{
		SubscribeHandler: subHandler,
		Middleware:       middleware,
	}

	//update fake users
	if uErr := userService.UpdateUsers(); uErr != nil {
		log.Debug("error update users", "error", cErr)
		panic(uErr)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go telegram.NewRouter(ctx, &wg, log, &tgHandlers, tg)
	go func() {

		//auto check birthdays every day in 8:00AM
		now := time.Now()
		nextUpdate := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
		if now.After(nextUpdate) {
			nextUpdate = nextUpdate.Add(24 * time.Hour)
		}
		timeToNextUpdate := nextUpdate.Sub(now)
		time.Sleep(timeToNextUpdate)

		//add ticker
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				wg.Add(1)
				go birthdayService.BirthdayNotify(ctx, &wg)
			}
		}
	}()

	<-ctx.Done()
	log.Info("server shutting down...")
	wg.Wait()
	if ccErr := dbConnection.CloseConnection(); ccErr != nil {
		log.Error("error close connection", "error", ccErr)
	}
}
