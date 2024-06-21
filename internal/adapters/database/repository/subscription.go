package repository

import (
	"birthdayapp/internal/adapters/database"
	"birthdayapp/internal/core/domain"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
)

type SubscriptionsRepository struct {
	db *database.DB
}

func NewSubscriptionsRepository(db *database.DB) *SubscriptionsRepository {
	return &SubscriptionsRepository{
		db,
	}
}

func (sr *SubscriptionsRepository) InsertSubscriptionByTelegramID(subscription *domain.Subscriptions) (*domain.Subscriptions, error) {
	query := `WITH subscriber_ids AS (
		SELECT id AS subscriber_id
		FROM users
		WHERE telegram_id = $1
	),
	subscribe_to_ids AS (
		SELECT id AS subscribe_to_id
		FROM users
		WHERE telegram_id = $2
	)
	INSERT INTO subscriptions (subscriber, subscribe_to)
	SELECT subscriber_id, subscribe_to_id
	FROM subscriber_ids, subscribe_to_ids
	RETURNING id;`

	err := sr.db.QueryRow(query, subscription.Subscriber.TelegramID, subscription.SubscribeTo.TelegramID).Scan(&subscription.ID)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		if errors.As(err, &sqliteErr) {
			switch {
			case errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique):
				return nil, domain.ErrAlreadyExist
			default:
				return nil, fmt.Errorf("error creating subscription with subscriber %d and subscribe_to %d: %w", subscription.Subscriber.TelegramID, subscription.SubscribeTo.TelegramID, err)
			}
		}
		return nil, err
	}
	return subscription, nil
}

func (sr *SubscriptionsRepository) DeleteSubscriptionByTelegramID(subscription *domain.Subscriptions) error {
	query := `WITH subscriber_ids AS (
		SELECT id
		FROM users
		WHERE telegram_id = $1
	),
	subscribe_to_ids AS (
		SELECT id
		FROM users
		WHERE telegram_id = $2
	)
	DELETE FROM subscriptions
	WHERE subscriber = (SELECT id FROM subscriber_ids)
	  AND subscribe_to = (SELECT id FROM subscribe_to_ids)`

	result, eErr := sr.db.Exec(query, subscription.Subscriber.TelegramID, subscription.SubscribeTo.TelegramID)
	if eErr != nil {
		return fmt.Errorf("error remove subscription with subscriber %d and subscribe_to %d: %w", subscription.Subscriber.TelegramID, subscription.SubscribeTo.TelegramID, eErr)
	}

	rowsAffected, raErr := result.RowsAffected()
	if raErr != nil {
		return fmt.Errorf("error remove subscription with subscriber %d and subscribe_to %d: %w", subscription.Subscriber.TelegramID, subscription.SubscribeTo.TelegramID, raErr)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription with subscriber %d and subscribe_to %d: %w", subscription.Subscriber.TelegramID, subscription.SubscribeTo.TelegramID, domain.ErrNotFound)
	}

	return nil
}
