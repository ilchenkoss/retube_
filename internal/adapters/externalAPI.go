package adapters

import (
	"birthdayapp/internal/core/domain"
	"time"
)

type ExternalAPI struct {
}

func NewExternalAPI() *ExternalAPI {
	return &ExternalAPI{}
}

func (extAPI *ExternalAPI) GetUsers() (*[]domain.User, error) {

	userOne := domain.User{
		Username:   "fakeUser1",
		TelegramID: 111,
		Birthday:   time.Date(2000, 06, 21, 0, 0, 0, 0, time.UTC)}

	userTwo := domain.User{
		Username:   "fakeUser2",
		TelegramID: 222,
		Birthday:   time.Date(2001, 07, 21, 0, 0, 0, 0, time.UTC)}

	return &[]domain.User{
		userOne,
		userTwo,
	}, nil
}
