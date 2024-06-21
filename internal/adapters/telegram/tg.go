package telegram

import (
	"birthdayapp/internal/core/domain"
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
)

type Telegram struct {
	log *slog.Logger
	bot *tgbotapi.BotAPI
}

func NewTelegramBot(log *slog.Logger, token string) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Telegram{
		bot: bot,
		log: log,
	}, nil
}

func (t *Telegram) GetInviteLink(chatID int64, birthdayUsernames string) (string, error) {
	inviteLink, err := t.bot.GetInviteLink(tgbotapi.ChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID:             chatID,
			SuperGroupUsername: "Primary Link",
		},
	})
	if err != nil {
		t.log.Debug("error of create join link", "error", err)
		return inviteLink, fmt.Errorf("error of create invite link: %w", err)
	}

	inviteLink = fmt.Sprintf("Join the group to congratulate the birthday for users: %s. Link: %s", birthdayUsernames, inviteLink)
	return inviteLink, nil
}

func (t *Telegram) SendLinkForJoinBirthdayGroup(userID int64, chatID int64, birthdayUsers *[]domain.User) error {
	op := "Telegram.SendLinkForJoinBirthdayGroup"
	t.log.With(slog.String("op", op))

	inviteLink, err := t.bot.GetInviteLink(tgbotapi.ChatInviteLinkConfig{
		tgbotapi.ChatConfig{
			ChatID:             chatID,
			SuperGroupUsername: "Invite Link",
		},
	})
	if err != nil {
		t.log.Debug("error of create join link", "error", err)
		return fmt.Errorf("error of create join link: %w", err)
	}

	var birthdayUsernames bytes.Buffer
	for i, birthdayUser := range *birthdayUsers {
		birthdayUsernames.WriteString(fmt.Sprintf("@%s", birthdayUser.Username))
		if i != len(*birthdayUsers)-1 {
			birthdayUsernames.WriteString(", ")
		}
	}

	t.SendMessage(userID, fmt.Sprintf("Join the group to congratulate the birthday for users: %s. Link: %s", birthdayUsernames.String(), inviteLink))
	return nil
}

func (t *Telegram) KickUser(chatID int64, userID int64) error {
	kickConfig := tgbotapi.KickChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		UntilDate: 0,
	}

	if _, err := t.bot.Request(kickConfig); err != nil {
		return fmt.Errorf("error kick user:%w", err)
	}
	return nil
}

func (t *Telegram) UnBanUser(chatID int64, userID int64) error {
	op := "Telegram.UnBanUser"
	t.log.With(slog.String("op", op))

	kickConfig := tgbotapi.UnbanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
	}

	if _, err := t.bot.Request(kickConfig); err != nil {
		return fmt.Errorf("error unBan user: %w", err)
	}
	return nil
}

func (t *Telegram) SendMessage(chatID int64, text string) {
	op := "Telegram.SendMessage"
	t.log.With(slog.String("op", op))

	msg := tgbotapi.NewMessage(chatID, text)

	_, err := t.bot.Send(msg)
	if err != nil {
		err = fmt.Errorf(fmt.Sprintf("error send message to chatID %d with text: '%s'", chatID, text), err)
		t.log.Debug("", "error", err)
	}
}
