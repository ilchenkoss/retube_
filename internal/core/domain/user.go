package domain

import "time"

type User struct {
	ID             int
	Username       string
	TelegramID     int64
	Birthday       time.Time
	NotifyBirthday bool
}
