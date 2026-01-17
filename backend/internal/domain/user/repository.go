package user

// UserRepository defines the interface for user persistence
// NOTE: No context.Context - domain layer is pure
// Infrastructure implementations add context internally
type UserRepository interface {
	// FindByID retrieves a user by their ID
	FindByID(id UserID) (*User, error)

	// FindByTelegramUsername retrieves a user by their Telegram username
	FindByTelegramUsername(username TelegramUsername) (*User, error)

	// FindAll retrieves all users (for admin purposes)
	FindAll(limit, offset int) ([]User, error)

	// Save persists a user (create or update)
	Save(user *User) error

	// Delete removes a user by ID (soft delete recommended)
	Delete(id UserID) error

	// Exists checks if a user exists by ID
	Exists(id UserID) (bool, error)
}
