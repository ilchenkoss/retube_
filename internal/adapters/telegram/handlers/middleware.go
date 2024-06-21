package handlers

import (
	"birthdayapp/internal/core/domain"
	"birthdayapp/internal/core/port"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Middleware struct {
	ur port.UserRepo
}

func NewMiddleware(ur port.UserRepo) *Middleware {
	return &Middleware{
		ur: ur,
	}
}

func (m *Middleware) UserMiddleware(update tgbotapi.Update) error {
	user := &domain.User{TelegramID: update.SentFrom().ID}
	_, guErr := m.ur.GetUserByTelegramID(user)
	if guErr != nil {
		//domain.ErrNotFound
		return guErr
	}
	return nil
}
