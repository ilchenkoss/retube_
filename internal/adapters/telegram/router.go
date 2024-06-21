package telegram

import (
	"birthdayapp/internal/adapters/telegram/handlers"
	"birthdayapp/internal/core/domain"
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"sync"
)

type Handlers struct {
	SubscribeHandler *handlers.SubscriptionsHandler
	Middleware       *handlers.Middleware
}

func NewRouter(ctx context.Context, wg *sync.WaitGroup, log *slog.Logger, h *Handlers, tg *Telegram) {
	defer wg.Done()
	op := "Telegram.Router"
	log.With(slog.String("op", op))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tg.bot.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return

		case update := <-updates:

			go func() {
				if update.Message == nil { // ignore any non-Message updates
					return
				}
				if !update.Message.IsCommand() { // ignore any non-command Messages
					return
				}
				if mErr := h.Middleware.UserMiddleware(update); mErr != nil {
					switch {
					case errors.Is(mErr, domain.ErrNotFound):
						tg.SendMessage(update.Message.Chat.ID, "You are not register in this service")
						return
					default:
						log.Debug("error userMiddleware", "error", mErr)
						tg.SendMessage(update.Message.Chat.ID, "internal server error")
						return
					}
				}

				switch update.Message.Command() {
				case "help":
					Help(update, tg)
				case "subscribeTo":
					h.SubscribeHandler.SubscribeTo(log, update, tg)
				case "unSubscribeFrom":
					h.SubscribeHandler.UnSubscribeFrom(log, update, tg)
				case "subscribeToNotifications":
					h.SubscribeHandler.SubscribeToNotifications(log, update, tg)
				default:
					tg.SendMessage(update.Message.Chat.ID, "unknown command, please send /help to get a list of commands")
				}
			}()

		}
	}
}

func Help(update tgbotapi.Update, tg *Telegram) {
	helpMessage := `u can use commands: 
	/subscribeToNotifications "true" for turn on and "false" for turn off notifications
	/subscribeTo "telegram_id" or "@username" for subscribe to user birthday,
	/unSubscribeFrom "telegram_id" or "@username" for unsubscribe from user
	`
	tg.SendMessage(update.Message.Chat.ID, helpMessage)
}
