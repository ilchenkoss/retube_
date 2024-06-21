package service

import (
	"birthdayapp/internal/core/domain"
	"birthdayapp/internal/core/port"
)

type SubscriptionService struct {
	sr port.SubscriptionsRepo
}

func NewSubscriptionService(sr port.SubscriptionsRepo) *SubscriptionService {
	return &SubscriptionService{
		sr: sr,
	}
}

func (ss *SubscriptionService) NewSubscription(subscription *domain.Subscriptions) (*domain.Subscriptions, error) {
	if subscription.Subscriber.TelegramID == subscription.SubscribeTo.TelegramID {
		return nil, domain.ErrUserRecursion
	}
	return ss.sr.InsertSubscriptionByTelegramID(subscription)

}

func (ss *SubscriptionService) RemoveSubscription(subscription *domain.Subscriptions) error {
	return ss.sr.DeleteSubscriptionByTelegramID(subscription)
}
