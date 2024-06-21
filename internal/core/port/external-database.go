package port

import "birthdayapp/internal/core/domain"

type ExternalAPI interface {
	GetUsers() (*[]domain.User, error)
}
