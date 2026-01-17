package user

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// RegisterUserUseCase handles the business logic for registering a user
// This is typically called when a user first opens the Telegram Mini App
// It creates a new user or returns an existing one (idempotent)
type RegisterUserUseCase struct {
	userRepo user.UserRepository
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase
func NewRegisterUserUseCase(userRepo user.UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo: userRepo,
	}
}

// Execute registers a new user or returns an existing one
func (uc *RegisterUserUseCase) Execute(input RegisterUserInput) (RegisterUserOutput, error) {
	// 1. Validate and create value objects
	userID, err := user.NewUserID(input.UserID)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	username, err := user.NewUsername(input.Username)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	telegramUsername, err := user.NewTelegramUsername(input.TelegramUsername)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	avatarURL, err := user.NewAvatarURL(input.AvatarURL)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	languageCode, err := user.NewLanguageCode(input.LanguageCode)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	// 2. Check if user already exists
	existingUser, err := uc.userRepo.FindByID(userID)
	if err == nil {
		// User already exists - update profile and return
		now := time.Now().Unix()
		err = existingUser.UpdateProfile(username, telegramUsername, user.Email{}, avatarURL, languageCode, now)
		if err != nil {
			return RegisterUserOutput{}, err
		}

		err = uc.userRepo.Save(existingUser)
		if err != nil {
			return RegisterUserOutput{}, err
		}

		return RegisterUserOutput{
			User:      ToUserDTO(existingUser),
			IsNewUser: false,
		}, nil
	}

	// 3. User doesn't exist - create new user
	now := time.Now().Unix()
	newUser, err := user.NewUser(userID, username, now)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	// Update with additional profile data
	err = newUser.UpdateProfile(username, telegramUsername, user.Email{}, avatarURL, languageCode, now)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	// 4. Save to repository
	err = uc.userRepo.Save(newUser)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	// 5. Return DTO
	return RegisterUserOutput{
		User:      ToUserDTO(newUser),
		IsNewUser: true,
	}, nil
}
