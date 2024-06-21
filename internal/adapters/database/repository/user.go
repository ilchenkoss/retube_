package repository

import (
	"birthdayapp/internal/adapters/database"
	"birthdayapp/internal/core/domain"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"strings"
	"time"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{
		db,
	}
}

func (u *UserRepository) InsertUser(user *domain.User) (*domain.User, error) {
	query := `
        INSERT INTO users (username, telegram_id, birthday, notify_birthday) 
        VALUES ( $1, $2, $3, $4)
        ON CONFLICT DO NOTHING
        RETURNING id
    `
	err := u.db.QueryRow(query, user.Username, user.TelegramID, user.Birthday, user.NotifyBirthday).Scan(&user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user %s: %w", user.Username, domain.ErrAlreadyExist)
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	return user, nil
}

func (u *UserRepository) InsertUsers(users *[]domain.User) error {

	tx, txErr := u.db.Begin()
	if txErr != nil {
		return fmt.Errorf("error begin transaction: %v", txErr)
	}

	defer func() {
		if txErr != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()

	stmt, pErr := tx.Prepare(`
	INSERT INTO users (username, telegram_id, birthday, notify_birthday) 
	VALUES (?,?,?,?)
	`)
	defer stmt.Close()
	if pErr != nil {
		return fmt.Errorf("error prepare statement: %v", pErr)
	}

	for _, user := range *users {
		_, eErr := stmt.Exec(user.Username, user.TelegramID, user.Birthday, user.NotifyBirthday)
		if eErr != nil {
			switch {
			case errors.Is(eErr, sql.ErrNoRows):
				return domain.ErrAlreadyExist
			case isUniqueConstraintError(eErr):
				return domain.ErrAlreadyExist
			default:
				return fmt.Errorf("error execute statement for user %v: %v", user.Username, eErr)
			}
		}
	}

	return nil
}

func isUniqueConstraintError(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return errors.Is(sqliteErr.Code, sqlite3.ErrConstraint)
	}
	return false
}

func (u *UserRepository) ChangeNotifyBirthdayByTelegramID(user *domain.User) (*domain.User, error) {

	query := `
        UPDATE users
        SET notify_birthday = ?
        WHERE telegram_id = ?
    `

	result, err := u.db.Exec(query, user.NotifyBirthday, user.TelegramID)
	if err != nil {
		return nil, fmt.Errorf("error updating notify_birthday: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("telegram_id %d: %w", user.TelegramID, domain.ErrNotFound)
	}

	return user, nil
}

func (u *UserRepository) GetUserByTelegramID(user *domain.User) (*domain.User, error) {

	query := `
        SELECT id, username, telegram_id, birthday, notify_birthday
        FROM users
        WHERE telegram_id = ?
    `

	row := u.db.QueryRow(query, user.TelegramID)

	var uUser domain.User
	err := row.Scan(&uUser.ID, &uUser.Username, &uUser.TelegramID, &uUser.Birthday, &uUser.NotifyBirthday)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("telegram_id %d: %w", user.TelegramID, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("error get user by telegram_id: %w", err)
	}

	return &uUser, nil
}

func (u *UserRepository) GetUserByUsername(user *domain.User) (*domain.User, error) {

	query := `
        SELECT id, username, telegram_id, birthday, notify_birthday
        FROM users
        WHERE username = ?
    `

	row := u.db.QueryRow(query, user.Username)

	var uUser domain.User
	err := row.Scan(&uUser.ID, &uUser.Username, &uUser.TelegramID, &uUser.Birthday, &uUser.NotifyBirthday)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("username %d: %w", user.Username, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("error get user by username: %w", err)
	}

	return &uUser, nil
}

func (u *UserRepository) GetUsersToSubscribeByTelegramID(user *domain.User) (*[]domain.User, error) {

	query := `
        SELECT id, username, telegram_id, birthday, notify_birthday
		FROM users
		WHERE telegram_id != ?
		ORDER BY birthday
		LIMIT ?
    `

	rows, err := u.db.Query(query, user.TelegramID, 10)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("error querying users by excluding telegram_id: %w", err)
	}

	var users []domain.User
	for rows.Next() {
		var uUser domain.User
		err := rows.Scan(&uUser.ID, &uUser.Username, &uUser.TelegramID, &uUser.Birthday, &uUser.NotifyBirthday)
		if err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}
		users = append(users, uUser)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return &users, nil
}

func (u *UserRepository) GetUsersWithBirthdayToday() (*[]domain.User, error) {
	now := time.Now()
	const stdNumMonth = "01"
	const stdNumDay = "02"
	today := now.Format(fmt.Sprintf("%s-%s", stdNumMonth, stdNumDay))

	query := `
        SELECT id, username, telegram_id, birthday, notify_birthday
        FROM users
		WHERE strftime('%m-%d', birthday) = ?
    `

	rows, qErr := u.db.Query(query, today)
	if qErr != nil {
		return nil, fmt.Errorf("error query: %w", qErr)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if sErr := rows.Scan(&user.ID, &user.Username, &user.TelegramID, &user.Birthday, &user.NotifyBirthday); sErr != nil {
			return nil, fmt.Errorf("error scan user : %w", sErr)
		}
		users = append(users, user)
	}

	if rErr := rows.Err(); rErr != nil {
		return nil, fmt.Errorf("error rows : %w", rErr)
	}

	return &users, nil
}

func (u *UserRepository) GetUsersSubscribedToUsers(birthdayUsers *[]domain.User) (*[]domain.User, error) {
	var placeholders []string
	for range *birthdayUsers {
		placeholders = append(placeholders, "?")
	}
	placeholderStr := strings.Join(placeholders, ",")

	query := fmt.Sprintf(`
        SELECT DISTINCT u.id, u.username, u.telegram_id, u.birthday, u.notify_birthday
        FROM users u
        INNER JOIN subscriptions s ON u.id = s.subscriber
        WHERE s.subscribe_to IN (%s)
    `, placeholderStr)

	args := make([]interface{}, len(*birthdayUsers))
	for i, user := range *birthdayUsers {
		args[i] = user.ID
	}

	rows, qErr := u.db.Query(query, args...)
	if qErr != nil {
		return nil, fmt.Errorf("error executing query: %w", qErr)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if sErr := rows.Scan(&user.ID, &user.Username, &user.TelegramID, &user.Birthday, &user.NotifyBirthday); sErr != nil {
			return nil, fmt.Errorf("error scanning user: %w", sErr)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error rows: %w", err)
	}

	return &users, nil
}
