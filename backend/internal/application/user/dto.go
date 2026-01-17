package user

// ========================================
// Common DTOs
// ========================================

// UserDTO is a data transfer object for User
type UserDTO struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	TelegramUsername string `json:"telegramUsername,omitempty"`
	Email            string `json:"email,omitempty"`
	AvatarURL        string `json:"avatarUrl,omitempty"`
	LanguageCode     string `json:"languageCode"`
	IsBlocked        bool   `json:"isBlocked"`
	CreatedAt        int64  `json:"createdAt"`
	UpdatedAt        int64  `json:"updatedAt"`
}

// UserProfileDTO is a lightweight DTO for user profile (used in leaderboards, etc.)
type UserProfileDTO struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	TelegramUsername string `json:"telegramUsername,omitempty"`
	AvatarURL        string `json:"avatarUrl,omitempty"`
}

// ========================================
// RegisterUser Use Case
// ========================================

// RegisterUserInput is the input DTO for RegisterUser use case
// Typically populated from Telegram WebApp initData
type RegisterUserInput struct {
	UserID           string `json:"userId" validate:"required"`     // Telegram user ID
	Username         string `json:"username,omitempty"`             // Display name (optional, can be empty for anonymous)
	TelegramUsername string `json:"telegramUsername,omitempty"`     // @username (optional)
	AvatarURL        string `json:"avatarUrl,omitempty"`            // Photo URL from Telegram
	LanguageCode     string `json:"languageCode,omitempty"`         // Language preference
}

// RegisterUserOutput is the output DTO for RegisterUser use case
type RegisterUserOutput struct {
	User      UserDTO `json:"user"`
	IsNewUser bool    `json:"isNewUser"` // true if user was created, false if already existed
}

// ========================================
// GetUser Use Case
// ========================================

// GetUserInput is the input DTO for GetUser use case
type GetUserInput struct {
	UserID string `json:"userId" validate:"required"`
}

// GetUserOutput is the output DTO for GetUser use case
type GetUserOutput struct {
	User UserDTO `json:"user"`
}

// ========================================
// UpdateUserProfile Use Case
// ========================================

// UpdateUserProfileInput is the input DTO for UpdateUserProfile use case
type UpdateUserProfileInput struct {
	UserID           string `json:"userId" validate:"required"`
	Username         string `json:"username,omitempty"`        // Can be empty for anonymous users
	TelegramUsername string `json:"telegramUsername,omitempty"`
	Email            string `json:"email,omitempty"`
	AvatarURL        string `json:"avatarUrl,omitempty"`
	LanguageCode     string `json:"languageCode,omitempty"`
}

// UpdateUserProfileOutput is the output DTO for UpdateUserProfile use case
type UpdateUserProfileOutput struct {
	User UserDTO `json:"user"`
}

// ========================================
// UpdateUserLanguage Use Case
// ========================================

// UpdateUserLanguageInput is the input DTO for UpdateUserLanguage use case
type UpdateUserLanguageInput struct {
	UserID       string `json:"userId" validate:"required"`
	LanguageCode string `json:"languageCode" validate:"required"`
}

// UpdateUserLanguageOutput is the output DTO for UpdateUserLanguage use case
type UpdateUserLanguageOutput struct {
	User UserDTO `json:"user"`
}

// ========================================
// ListUsers Use Case (Admin)
// ========================================

// ListUsersInput is the input DTO for ListUsers use case
type ListUsersInput struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ListUsersOutput is the output DTO for ListUsers use case
type ListUsersOutput struct {
	Users []UserDTO `json:"users"`
	Total int       `json:"total"`
}

// ========================================
// GetUserByTelegramUsername Use Case
// ========================================

// GetUserByTelegramUsernameInput is the input DTO
type GetUserByTelegramUsernameInput struct {
	TelegramUsername string `json:"telegramUsername" validate:"required"`
}

// GetUserByTelegramUsernameOutput is the output DTO
type GetUserByTelegramUsernameOutput struct {
	User UserDTO `json:"user"`
}
