package port

import "birthdayapp/internal/core/domain"

type SubscriptionsRepo interface {
	InsertSubscriptionByTelegramID(subscription *domain.Subscriptions) (*domain.Subscriptions, error)
	DeleteSubscriptionByTelegramID(subscription *domain.Subscriptions) error
}

type SubscriptionsService interface {
	NewSubscription(subscription *domain.Subscriptions) (*domain.Subscriptions, error)
	RemoveSubscription(subscription *domain.Subscriptions) error
}
