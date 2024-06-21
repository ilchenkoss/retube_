package handlers

import (
	"birthdayapp/internal/core/domain"
	"birthdayapp/internal/core/port"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"strconv"
	"strings"
)

type SubscriptionsHandler struct {
	ss port.SubscriptionsService
	us port.UserService
}

func NewSubscriptionsHandler(ss port.SubscriptionsService, us port.UserService) *SubscriptionsHandler {
	return &SubscriptionsHandler{
		ss: ss,
		us: us,
	}
}

func isInt(str string) bool {
	_, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return false
	}
	return true
}

func (sh *SubscriptionsHandler) newSubscription(log *slog.Logger, update tgbotapi.Update, tg port.Telegram) *domain.Subscriptions {

	var subscription domain.Subscriptions
	switch {

	case update.Message.CommandArguments() == "":
		tg.SendMessage(update.Message.Chat.ID, "username must be @username or 0000(telegram_id)")
		return nil

	case strings.Contains(update.Message.CommandArguments(), "@"):

		usernameSlice := strings.Split(update.Message.CommandArguments(), "@")
		if len(usernameSlice) != 2 {
			tg.SendMessage(update.Message.Chat.ID, "username must be @username")
			return nil
		}

		username := usernameSlice[1]
		telegramID, guErr := sh.us.GetTelegramIDByUsername(username)
		if guErr != nil {
			switch {
			case errors.Is(guErr, domain.ErrNotFound):
				tg.SendMessage(update.Message.Chat.ID, fmt.Sprintf("user @%s not register in service", username))
				return nil
			default:
				log.Debug("error of get user by username", "error", guErr)
				tg.SendMessage(update.Message.Chat.ID, "internal server error")
				return nil
			}
		}
		subscription = domain.Subscriptions{
			Subscriber:  &domain.User{TelegramID: update.SentFrom().ID},
			SubscribeTo: &domain.User{TelegramID: telegramID},
		}

	case isInt(update.Message.CommandArguments()):
		userID, _ := strconv.ParseInt(update.Message.CommandArguments(), 10, 64)
		subscription = domain.Subscriptions{
			Subscriber:  &domain.User{TelegramID: update.SentFrom().ID},
			SubscribeTo: &domain.User{TelegramID: userID},
		}

	default:
		tg.SendMessage(update.Message.Chat.ID, "couldn't get user ID or @username")
		return nil
	}
	return &subscription
}

func (sh *SubscriptionsHandler) SubscribeTo(log *slog.Logger, update tgbotapi.Update, tg port.Telegram) {
	op := "handlers.SubscribeTo"
	log.With(slog.String("op", op))

	subscription := sh.newSubscription(log, update, tg)
	if subscription == nil {
		return
	}

	_, nsErr := sh.ss.NewSubscription(subscription)

	if nsErr != nil {
		switch {
		case errors.Is(nsErr, domain.ErrAlreadyExist):
			tg.SendMessage(update.Message.Chat.ID, "you are already subscribed to user")
			return
		case errors.Is(nsErr, domain.ErrUserRecursion):
			tg.SendMessage(update.Message.Chat.ID, "you can't subscribe to yourself")
			return
		case errors.Is(nsErr, domain.ErrNotFound):
			tg.SendMessage(update.Message.Chat.ID, "user not register in service")
			return
		default:
			log.Debug("error new subscription", "error", nsErr)
			tg.SendMessage(update.Message.Chat.ID, "internal server error")
			return
		}
	}
	tg.SendMessage(update.Message.Chat.ID, "success, you are subscribed")
}

func (sh *SubscriptionsHandler) UnSubscribeFrom(log *slog.Logger, update tgbotapi.Update, tg port.Telegram) {
	op := "handlers.UnSubscribeFrom"
	log.With(slog.String("op", op))

	subscription := sh.newSubscription(log, update, tg)
	if subscription == nil {
		return
	}

	nsErr := sh.ss.RemoveSubscription(subscription)
	if nsErr != nil {
		switch {
		case errors.Is(nsErr, domain.ErrNotFound):
			tg.SendMessage(update.Message.Chat.ID, "you are not subscribe for this user")
			return
		default:
			log.Debug("error remove subscription", "error", nsErr)
			tg.SendMessage(update.Message.Chat.ID, "internal server error")
			return
		}

	}
	tg.SendMessage(update.Message.Chat.ID, "success, subscription removed")
}

func (sh *SubscriptionsHandler) SubscribeToNotifications(log *slog.Logger, update tgbotapi.Update, tg port.Telegram) {
	op := "handlers.SubscribeToNotifications"
	log.With(slog.String("op", op))

	user := &domain.User{TelegramID: update.SentFrom().ID}

	var notify bool

	switch update.Message.CommandArguments() {
	case "true":
		notify = true
		user.NotifyBirthday = notify
	case "false":
		notify = false
		user.NotifyBirthday = notify
	default:
		tg.SendMessage(update.Message.Chat.ID, "arg must be 'true' or 'false' ")
		return
	}

	guErr := sh.us.ChangeNotify(user)
	if guErr != nil {
		log.Debug("error change notification", "error", guErr)
		tg.SendMessage(update.Message.Chat.ID, "couldn't change notification, try again later")
		return
	}

	tg.SendMessage(update.Message.Chat.ID, fmt.Sprintf("success, notification change to %v", notify))
}
