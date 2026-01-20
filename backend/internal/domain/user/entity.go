package user

// User is an aggregate root representing a user in the system
type User struct {
	id               UserID
	username         Username
	telegramUsername TelegramUsername
	email            Email
	avatarURL        AvatarURL
	languageCode     LanguageCode
	isBlocked        bool
	createdAt        int64 // Unix timestamp (no time.Time to keep domain pure)
	updatedAt        int64 // Unix timestamp
}

// NewUser creates a new User entity
// Username can be empty for anonymous users (as per DOMAIN.md)
func NewUser(
	id UserID,
	username Username,
	createdAt int64,
) (*User, error) {
	if id.IsZero() {
		return nil, ErrInvalidUserID
	}

	// Username can be empty (anonymous user)

	return &User{
		id:        id,
		username:  username,
		createdAt: createdAt,
		updatedAt: createdAt,
		isBlocked: false,
	}, nil
}

// ReconstructUser reconstructs a User from persistence (no validation)
// Used by repository when loading from database
func ReconstructUser(
	id UserID,
	username Username,
	telegramUsername TelegramUsername,
	email Email,
	avatarURL AvatarURL,
	languageCode LanguageCode,
	isBlocked bool,
	createdAt int64,
	updatedAt int64,
) *User {
	return &User{
		id:               id,
		username:         username,
		telegramUsername: telegramUsername,
		email:            email,
		avatarURL:        avatarURL,
		languageCode:     languageCode,
		isBlocked:        isBlocked,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

// UpdateProfile updates user profile information
// Username can be empty for anonymous users
func (u *User) UpdateProfile(
	username Username,
	telegramUsername TelegramUsername,
	email Email,
	avatarURL AvatarURL,
	languageCode LanguageCode,
	updatedAt int64,
) error {
	// Username can be empty (anonymous user)

	u.username = username
	u.telegramUsername = telegramUsername
	u.email = email
	u.avatarURL = avatarURL
	u.languageCode = languageCode
	u.updatedAt = updatedAt

	return nil
}

// UpdateUsername updates only the username
// Username can be empty for anonymous users
func (u *User) UpdateUsername(username Username, updatedAt int64) error {
	// Username can be empty (anonymous user)

	u.username = username
	u.updatedAt = updatedAt
	return nil
}

// UpdateLanguage updates user's language preference
func (u *User) UpdateLanguage(languageCode LanguageCode, updatedAt int64) {
	u.languageCode = languageCode
	u.updatedAt = updatedAt
}

// Block blocks the user
func (u *User) Block(updatedAt int64) error {
	if u.isBlocked {
		return ErrUserBlocked
	}

	u.isBlocked = true
	u.updatedAt = updatedAt
	return nil
}

// Unblock unblocks the user
func (u *User) Unblock(updatedAt int64) {
	u.isBlocked = false
	u.updatedAt = updatedAt
}

// IsActive checks if user is active (not blocked)
func (u *User) IsActive() bool {
	return !u.isBlocked
}

// Getters (no setters - modifications only through business methods)
func (u *User) ID() UserID                         { return u.id }
func (u *User) Username() Username                 { return u.username }
func (u *User) TelegramUsername() TelegramUsername { return u.telegramUsername }
func (u *User) Email() Email                       { return u.email }
func (u *User) AvatarURL() AvatarURL               { return u.avatarURL }
func (u *User) LanguageCode() LanguageCode         { return u.languageCode }
func (u *User) IsBlocked() bool                    { return u.isBlocked }
func (u *User) CreatedAt() int64                   { return u.createdAt }
func (u *User) UpdatedAt() int64                   { return u.updatedAt }
