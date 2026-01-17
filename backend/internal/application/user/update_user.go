package user

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// UpdateUserProfileUseCase handles the business logic for updating user profile
type UpdateUserProfileUseCase struct {
	userRepo user.UserRepository
}

// NewUpdateUserProfileUseCase creates a new UpdateUserProfileUseCase
func NewUpdateUserProfileUseCase(userRepo user.UserRepository) *UpdateUserProfileUseCase {
	return &UpdateUserProfileUseCase{
		userRepo: userRepo,
	}
}

// Execute updates user profile
func (uc *UpdateUserProfileUseCase) Execute(input UpdateUserProfileInput) (UpdateUserProfileOutput, error) {
	// 1. Validate input and create value objects
	userID, err := user.NewUserID(input.UserID)
	if err != nil {
		return UpdateUserProfileOutput{}, err
	}

	username, err := user.NewUsername(input.Username)
	if err != nil {
		return UpdateUserProfileOutput{}, err
	}

	telegramUsername, err := user.NewTelegramUsername(input.TelegramUsername)
	if err != nil {
		return UpdateUserProfileOutput{}, err
	}

	email, err := user.NewEmail(input.Email)
	if err != nil {
		return UpdateUserProfileOutput{}, err
	}

	avatarURL, err := user.NewAvatarURL(input.AvatarURL)
	if err != nil {
		return UpdateUserProfileOutput{}, err
	}

	languageCode, err := user.NewLanguageCode(input.LanguageCode)
	if err != nil {
		return UpdateUserProfileOutput{}, err
	}

	// 2. Load user from repository
	userEntity, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return UpdateUserProfileOutput{}, err
	}

	// 3. Update profile
	now := time.Now().Unix()
	err = userEntity.UpdateProfile(username, telegramUsername, email, avatarURL, languageCode, now)
	if err != nil {
		return UpdateUserProfileOutput{}, err
	}

	// 4. Save to repository
	err = uc.userRepo.Save(userEntity)
	if err != nil {
		return UpdateUserProfileOutput{}, err
	}

	// 5. Return DTO
	return UpdateUserProfileOutput{
		User: ToUserDTO(userEntity),
	}, nil
}

// ========================================
// UpdateUserLanguage Use Case
// ========================================

// UpdateUserLanguageUseCase handles the business logic for updating user language preference
type UpdateUserLanguageUseCase struct {
	userRepo user.UserRepository
}

// NewUpdateUserLanguageUseCase creates a new UpdateUserLanguageUseCase
func NewUpdateUserLanguageUseCase(userRepo user.UserRepository) *UpdateUserLanguageUseCase {
	return &UpdateUserLanguageUseCase{
		userRepo: userRepo,
	}
}

// Execute updates user language preference
func (uc *UpdateUserLanguageUseCase) Execute(input UpdateUserLanguageInput) (UpdateUserLanguageOutput, error) {
	// 1. Validate input and create value objects
	userID, err := user.NewUserID(input.UserID)
	if err != nil {
		return UpdateUserLanguageOutput{}, err
	}

	languageCode, err := user.NewLanguageCode(input.LanguageCode)
	if err != nil {
		return UpdateUserLanguageOutput{}, err
	}

	// 2. Load user from repository
	userEntity, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return UpdateUserLanguageOutput{}, err
	}

	// 3. Update language
	now := time.Now().Unix()
	userEntity.UpdateLanguage(languageCode, now)

	// 4. Save to repository
	err = uc.userRepo.Save(userEntity)
	if err != nil {
		return UpdateUserLanguageOutput{}, err
	}

	// 5. Return DTO
	return UpdateUserLanguageOutput{
		User: ToUserDTO(userEntity),
	}, nil
}
