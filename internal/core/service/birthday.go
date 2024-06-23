package service

import (
	"birthdayapp/internal/config"
	"birthdayapp/internal/core/domain"
	"birthdayapp/internal/core/port"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type BirthdayService struct {
	log *slog.Logger
	cfg *config.Config
	ur  port.UserRepo
	tg  port.Telegram
}

func NewBirthdayService(log *slog.Logger, ur port.UserRepo, tg port.Telegram, cfg *config.Config) *BirthdayService {
	return &BirthdayService{
		log: log,
		cfg: cfg,
		ur:  ur,
		tg:  tg,
	}
}

func (bs *BirthdayService) BirthdayNotify(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	op := "birthdayService.BirthdayNotify"
	bs.log.With(slog.String("op", op))

	birthdayUsers, btErr := bs.ur.GetUsersWithBirthdayToday()
	if btErr != nil {
		bs.log.Error("GetUsersWithBirthdayToday error: ", "error", btErr.Error())
		return
	}
	if len(*birthdayUsers) == 0 {
		//without birthday today
		return
	}

	subscribers, gsuErr := bs.ur.GetUsersSubscribedToUsers(birthdayUsers)
	if gsuErr != nil {
		bs.log.Error("GetUsersSubscribedToUsers error from birthdayUsers: ", "error", birthdayUsers, gsuErr)
		return
	}
	if len(*subscribers) == 0 && len(*birthdayUsers) == 1 {
		//if no one to wish happy birthday
		return
	}

	allUsers := append(*birthdayUsers, *subscribers...)

	var birthdayUsernamesBuffer bytes.Buffer

	for i, birthdayUser := range *birthdayUsers {
		birthdayUsernamesBuffer.WriteString("@")
		birthdayUsernamesBuffer.WriteString(birthdayUser.Username)
		if i != len(*birthdayUsers)-1 {
			birthdayUsernamesBuffer.WriteString(", ")
		}
	}
	birthdayUsernamesString := birthdayUsernamesBuffer.String()
	bs.sendInviteForUsers(&allUsers, birthdayUsernamesString)
	bs.tg.SendMessage(bs.cfg.BirthdayGroupID, fmt.Sprintf("happy birthday %s", birthdayUsernamesString))

	bs.kickUsers(ctx, &allUsers)
}

func (bs *BirthdayService) kickUsers(ctx context.Context, usersToKick *[]domain.User) {
	op := "birthdayService.kickUsers"
	bs.log.With(slog.String("op", op))

	timeToKick := time.NewTimer(bs.cfg.TimeToKick)

	defer func() {
		for _, user := range *usersToKick {
			if user.TelegramID == bs.cfg.GroupOwnerID {
				continue
			}
			kErr := bs.tg.KickUser(bs.cfg.BirthdayGroupID, user.TelegramID)
			if kErr != nil {
				bs.log.Error("error kick user with telegram_id: ", "error", user.TelegramID, kErr)
				bs.tg.SendMessage(user.TelegramID, "please, leave from group. We'll wait for next birthday")
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timeToKick.C:
			timeToKick.Stop()
			return
		}
	}
}

func (bs *BirthdayService) sendInviteForUsers(usersForSendInvite *[]domain.User, birthdayUsers string) {
	op := "birthdayService.sendInviteForUsers"
	bs.log.With(slog.String("op", op))

	inviteLink, ilErr := bs.tg.GetInviteLink(bs.cfg.BirthdayGroupID, birthdayUsers)
	if ilErr != nil {
		bs.log.Error("error generate invite link", "error", ilErr)
		inviteLink = fmt.Sprintf("to invite birthday group with users celebrating: %s, contact support", birthdayUsers)
	}

	for _, userForNotify := range *usersForSendInvite {
		if ubErr := bs.tg.UnBanUser(bs.cfg.BirthdayGroupID, userForNotify.TelegramID); ubErr != nil && userForNotify.TelegramID != bs.cfg.GroupOwnerID {
			bs.tg.SendMessage(userForNotify.TelegramID, fmt.Sprintf("to invite birthday group with users celebrating: %s, contact support", birthdayUsers))
			continue
		}
		bs.tg.SendMessage(userForNotify.TelegramID, inviteLink)
	}
}
