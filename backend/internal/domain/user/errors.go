package user

import "errors"

var (
	// Value object errors
	ErrUsernameTooLong         = errors.New("username is too long (max 100 characters)")
	ErrInvalidTelegramUsername = errors.New("invalid Telegram username format")
	ErrInvalidEmail            = errors.New("invalid email format")
	ErrEmailTooLong            = errors.New("email is too long (max 255 characters)")
	ErrInvalidAvatarURL        = errors.New("invalid avatar URL format")
	ErrAvatarURLTooLong        = errors.New("avatar URL is too long (max 500 characters)")
	ErrInvalidLanguageCode     = errors.New("invalid language code (must be 2-letter ISO 639-1)")

	// Entity errors
	ErrInvalidUserID     = errors.New("user ID cannot be empty")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserBlocked       = errors.New("user is blocked")
)
