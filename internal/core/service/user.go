package service

import (
	"birthdayapp/internal/core/domain"
	"birthdayapp/internal/core/port"
	"errors"
)

type UserService struct {
	ur     port.UserRepo
	extAPI port.ExternalAPI
}

func NewUserService(ur port.UserRepo, extAPI port.ExternalAPI) *UserService {
	return &UserService{
		ur:     ur,
		extAPI: extAPI,
	}
}

func (us *UserService) UpdateUsers() error {

	users, guErr := us.extAPI.GetUsers()
	if guErr != nil {
		return guErr
	}

	iuErr := us.ur.InsertUsers(users)
	if iuErr != nil && !errors.Is(iuErr, domain.ErrAlreadyExist) {
		return iuErr
	}

	return nil
}

func (us *UserService) GetUsers(user *domain.User) (*[]domain.User, error) {
	users, guErr := us.ur.GetUsersToSubscribeByTelegramID(user)
	if guErr != nil {
		return nil, guErr
	}
	return users, nil
}
func (us *UserService) ChangeNotify(user *domain.User) error {
	_, guErr := us.ur.ChangeNotifyBirthdayByTelegramID(user)
	if guErr != nil {
		return guErr
	}
	return nil
}

func (us *UserService) GetTelegramIDByUsername(username string) (int64, error) {
	user := &domain.User{Username: username}
	uUser, guErr := us.ur.GetUserByUsername(user)
	if guErr != nil {
		return 0, guErr
	}
	return uUser.TelegramID, nil
}
