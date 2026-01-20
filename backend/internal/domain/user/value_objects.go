package user

import (
	"regexp"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
)

// UserID wraps the shared UserID (Telegram user ID)
type UserID = shared.UserID

// NewUserID creates a new UserID from Telegram user ID string
var NewUserID = shared.NewUserID

// Username is a value object for user display name
// Can be empty for anonymous users (as per DOMAIN.md)
type Username struct {
	value string
}

// NewUsername creates a new Username
// Empty username is allowed for anonymous users
func NewUsername(value string) (Username, error) {
	if len(value) > 100 {
		return Username{}, ErrUsernameTooLong
	}
	return Username{value: value}, nil
}

func (u Username) String() string {
	if u.value == "" {
		return "anonymous"
	}
	return u.value
}

func (u Username) Value() string {
	return u.value
}

func (u Username) IsEmpty() bool {
	return u.value == ""
}

func (u Username) IsAnonymous() bool {
	return u.value == ""
}

// TelegramUsername is a value object for Telegram @username
type TelegramUsername struct {
	value string // without @ prefix
}

// NewTelegramUsername creates a new TelegramUsername
func NewTelegramUsername(value string) (TelegramUsername, error) {
	if value == "" {
		return TelegramUsername{}, nil // Telegram username is optional
	}

	// Remove @ prefix if present
	if value[0] == '@' {
		value = value[1:]
	}

	// Validate username format (alphanumeric and underscores, 5-32 chars)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{5,32}$`, value)
	if !matched {
		return TelegramUsername{}, ErrInvalidTelegramUsername
	}

	return TelegramUsername{value: value}, nil
}

func (tu TelegramUsername) String() string {
	if tu.value == "" {
		return ""
	}
	return "@" + tu.value
}

func (tu TelegramUsername) Value() string {
	return tu.value
}

func (tu TelegramUsername) IsEmpty() bool {
	return tu.value == ""
}

// Email is a value object for user email
type Email struct {
	value string
}

// NewEmail creates a new Email
func NewEmail(value string) (Email, error) {
	if value == "" {
		return Email{}, nil // Email is optional
	}

	// Basic email validation
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, value)
	if !matched {
		return Email{}, ErrInvalidEmail
	}

	if len(value) > 255 {
		return Email{}, ErrEmailTooLong
	}

	return Email{value: value}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) IsEmpty() bool {
	return e.value == ""
}

// AvatarURL is a value object for user avatar URL
type AvatarURL struct {
	value string
}

// NewAvatarURL creates a new AvatarURL
func NewAvatarURL(value string) (AvatarURL, error) {
	if value == "" {
		return AvatarURL{}, nil // Avatar is optional
	}

	// Basic URL validation
	matched, _ := regexp.MatchString(`^https?://`, value)
	if !matched {
		return AvatarURL{}, ErrInvalidAvatarURL
	}

	if len(value) > 500 {
		return AvatarURL{}, ErrAvatarURLTooLong
	}

	return AvatarURL{value: value}, nil
}

func (a AvatarURL) String() string {
	return a.value
}

func (a AvatarURL) IsEmpty() bool {
	return a.value == ""
}

// LanguageCode is a value object for user language preference
type LanguageCode struct {
	value string
}

// NewLanguageCode creates a new LanguageCode
func NewLanguageCode(value string) (LanguageCode, error) {
	if value == "" {
		return LanguageCode{value: "en"}, nil // Default to English
	}

	// ISO 639-1 (2-letter codes)
	matched, _ := regexp.MatchString(`^[a-z]{2}$`, value)
	if !matched {
		return LanguageCode{}, ErrInvalidLanguageCode
	}

	return LanguageCode{value: value}, nil
}

func (lc LanguageCode) String() string {
	return lc.value
}

func (lc LanguageCode) IsDefault() bool {
	return lc.value == "en"
}
