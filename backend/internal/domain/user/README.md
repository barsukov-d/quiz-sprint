# User Domain

This package contains the User domain model following Domain-Driven Design (DDD) principles.

## Structure

- `entity.go` - User aggregate root with business logic
- `value_objects.go` - Value objects (Username, Email, TelegramUsername, etc.)
- `repository.go` - Repository interface for persistence
- `errors.go` - Domain-specific errors

## Value Objects

### UserID
Telegram user ID (reuses `shared.UserID`). Example: `"123456789"`

### Username
User's display name. Required, max 100 characters.

### TelegramUsername
Telegram @username (optional). Validated against Telegram username rules (5-32 alphanumeric + underscores).

### Email
User's email (optional). Basic email format validation.

### AvatarURL
URL to user's avatar image (optional). Must be valid HTTP/HTTPS URL.

### LanguageCode
User's language preference. ISO 639-1 two-letter code (e.g., "en", "ru"). Defaults to "en".

## Entity

### User
Aggregate root representing a user in the system.

**Business Rules:**
- User must have a valid UserID and Username
- User can be blocked/unblocked
- Profile updates modify username, email, avatar, language, and Telegram username
- All modifications update the `updatedAt` timestamp
- Timestamps are Unix timestamps (int64) to keep domain pure

**Methods:**
- `NewUser(id, username, createdAt)` - Create new user
- `ReconstructUser(...)` - Reconstruct from database (no validation)
- `UpdateProfile(...)` - Update all profile fields
- `UpdateUsername(username, updatedAt)` - Update only username
- `UpdateLanguage(languageCode, updatedAt)` - Update language preference
- `Block(updatedAt)` - Block user
- `Unblock(updatedAt)` - Unblock user
- `IsActive()` - Check if user is not blocked

## Repository Interface

### UserRepository
Defines persistence operations (implementations in infrastructure layer):

- `FindByID(id)` - Get user by Telegram ID
- `FindByTelegramUsername(username)` - Get user by @username
- `FindAll(limit, offset)` - List users (admin)
- `Save(user)` - Create or update user
- `Delete(id)` - Delete user (recommend soft delete)
- `Exists(id)` - Check if user exists

## Usage Example

```go
import (
    "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
    "time"
)

// Create new user
userID, _ := user.NewUserID("123456789")
username, _ := user.NewUsername("John Doe")
now := time.Now().Unix()

u, err := user.NewUser(userID, username, now)
if err != nil {
    // handle error
}

// Update profile
email, _ := user.NewEmail("john@example.com")
telegramUsername, _ := user.NewTelegramUsername("johndoe")
langCode, _ := user.NewLanguageCode("en")
avatarURL, _ := user.NewAvatarURL("https://example.com/avatar.jpg")

err = u.UpdateProfile(username, telegramUsername, email, avatarURL, langCode, time.Now().Unix())

// Block user
err = u.Block(time.Now().Unix())

// Check if active
if u.IsActive() {
    // user can access the system
}
```

## Next Steps

To integrate this domain into the application:

1. **Application Layer** - Create use cases in `internal/application/user/`:
   - `RegisterUser` - Register new user from Telegram data
   - `UpdateUserProfile` - Update user profile
   - `GetUser` - Retrieve user details
   - `BlockUser` / `UnblockUser` - Admin operations

2. **Infrastructure Layer** - Implement repository in `internal/infrastructure/persistence/`:
   - PostgreSQL implementation of `UserRepository`
   - Database migrations for `users` table

3. **HTTP Handlers** - Create handlers in `internal/infrastructure/http/handlers/`:
   - `GET /api/v1/user/:id` - Get user profile
   - `PUT /api/v1/user/:id` - Update user profile
   - `GET /api/v1/users` - List users (admin)
   - Add Swagger annotations for API documentation

4. **Integration** - Connect with quiz domain:
   - Update `QuizSession` to reference `user.UserID`
   - Update leaderboard queries to join with users table
   - Add user info to WebSocket leaderboard updates
