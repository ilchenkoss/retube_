package port

import "birthdayapp/internal/core/domain"

type UserRepo interface {
	InsertUser(user *domain.User) (*domain.User, error)
	InsertUsers(users *[]domain.User) error
	ChangeNotifyBirthdayByTelegramID(user *domain.User) (*domain.User, error)
	GetUserByTelegramID(user *domain.User) (*domain.User, error)
	GetUserByUsername(user *domain.User) (*domain.User, error)
	GetUsersToSubscribeByTelegramID(user *domain.User) (*[]domain.User, error)
	GetUsersWithBirthdayToday() (*[]domain.User, error)
	GetUsersSubscribedToUsers(birthdayUsers *[]domain.User) (*[]domain.User, error)
}

type UserService interface {
	UpdateUsers() error
	GetUsers(user *domain.User) (*[]domain.User, error)
	GetTelegramIDByUsername(username string) (int64, error)
	ChangeNotify(user *domain.User) error
}
