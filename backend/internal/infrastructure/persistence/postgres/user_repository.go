package postgres

import (
	"database/sql"
	"fmt"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// UserRepository is a PostgreSQL implementation of user.UserRepository
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID retrieves a user by Telegram ID
func (r *UserRepository) FindByID(id user.UserID) (*user.User, error) {
	query := `
		SELECT id, username, telegram_username, email, avatar_url, language_code, is_blocked, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var (
		userID           string
		username         string
		telegramUsername sql.NullString
		email            sql.NullString
		avatarURL        sql.NullString
		languageCode     string
		isBlocked        bool
		createdAt        int64
		updatedAt        int64
	)

	err := r.db.QueryRow(query, id.String()).Scan(
		&userID,
		&username,
		&telegramUsername,
		&email,
		&avatarURL,
		&languageCode,
		&isBlocked,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return r.scanUser(userID, username, telegramUsername, email, avatarURL, languageCode, isBlocked, createdAt, updatedAt)
}

// FindByTelegramUsername retrieves a user by Telegram @username
func (r *UserRepository) FindByTelegramUsername(username user.TelegramUsername) (*user.User, error) {
	query := `
		SELECT id, username, telegram_username, email, avatar_url, language_code, is_blocked, created_at, updated_at
		FROM users
		WHERE telegram_username = $1
	`

	var (
		userID           string
		userName         string
		telegramUsername sql.NullString
		email            sql.NullString
		avatarURL        sql.NullString
		languageCode     string
		isBlocked        bool
		createdAt        int64
		updatedAt        int64
	)

	err := r.db.QueryRow(query, username.String()).Scan(
		&userID,
		&userName,
		&telegramUsername,
		&email,
		&avatarURL,
		&languageCode,
		&isBlocked,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, user.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user by telegram username: %w", err)
	}

	return r.scanUser(userID, userName, telegramUsername, email, avatarURL, languageCode, isBlocked, createdAt, updatedAt)
}

// FindAll retrieves all users with pagination
func (r *UserRepository) FindAll(limit, offset int) ([]user.User, error) {
	query := `
		SELECT id, username, telegram_username, email, avatar_url, language_code, is_blocked, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []user.User

	for rows.Next() {
		var (
			userID           string
			username         string
			telegramUsername sql.NullString
			email            sql.NullString
			avatarURL        sql.NullString
			languageCode     string
			isBlocked        bool
			createdAt        int64
			updatedAt        int64
		)

		err := rows.Scan(
			&userID,
			&username,
			&telegramUsername,
			&email,
			&avatarURL,
			&languageCode,
			&isBlocked,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		u, err := r.scanUser(userID, username, telegramUsername, email, avatarURL, languageCode, isBlocked, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}

		users = append(users, *u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

// Save stores a user (create or update)
func (r *UserRepository) Save(u *user.User) error {
	query := `
		INSERT INTO users (id, username, telegram_username, email, avatar_url, language_code, is_blocked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			telegram_username = EXCLUDED.telegram_username,
			email = EXCLUDED.email,
			avatar_url = EXCLUDED.avatar_url,
			language_code = EXCLUDED.language_code,
			is_blocked = EXCLUDED.is_blocked,
			updated_at = EXCLUDED.updated_at
	`

	telegramUsername := toNullString(u.TelegramUsername().String())
	email := toNullString(u.Email().String())
	avatarURL := toNullString(u.AvatarURL().String())

	_, err := r.db.Exec(
		query,
		u.ID().String(),
		u.Username().String(),
		telegramUsername,
		email,
		avatarURL,
		u.LanguageCode().String(),
		u.IsBlocked(),
		u.CreatedAt(),
		u.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

// Delete removes a user
func (r *UserRepository) Delete(id user.UserID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// Exists checks if a user exists
func (r *UserRepository) Exists(id user.UserID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`

	var exists bool
	err := r.db.QueryRow(query, id.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// scanUser reconstructs a User entity from database values
func (r *UserRepository) scanUser(
	userID string,
	username string,
	telegramUsername sql.NullString,
	email sql.NullString,
	avatarURL sql.NullString,
	languageCode string,
	isBlocked bool,
	createdAt int64,
	updatedAt int64,
) (*user.User, error) {
	// Reconstruct value objects
	id, err := user.NewUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	uname, err := user.NewUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	tgUsername, err := user.NewTelegramUsername(fromNullString(telegramUsername))
	if err != nil {
		return nil, fmt.Errorf("invalid telegram username: %w", err)
	}

	userEmail, err := user.NewEmail(fromNullString(email))
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	avatar, err := user.NewAvatarURL(fromNullString(avatarURL))
	if err != nil {
		return nil, fmt.Errorf("invalid avatar URL: %w", err)
	}

	langCode, err := user.NewLanguageCode(languageCode)
	if err != nil {
		return nil, fmt.Errorf("invalid language code: %w", err)
	}

	// Reconstruct user entity (no validation for database reads)
	return user.ReconstructUser(id, uname, tgUsername, userEmail, avatar, langCode, isBlocked, createdAt, updatedAt), nil
}

// Helper functions for nullable fields
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNullString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}
