package user

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// GetUserUseCase handles the business logic for getting a user by ID
type GetUserUseCase struct {
	userRepo user.UserRepository
}

// NewGetUserUseCase creates a new GetUserUseCase
func NewGetUserUseCase(userRepo user.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves a user by ID
func (uc *GetUserUseCase) Execute(input GetUserInput) (GetUserOutput, error) {
	// 1. Validate input
	userID, err := user.NewUserID(input.UserID)
	if err != nil {
		return GetUserOutput{}, err
	}

	// 2. Load user from repository
	userEntity, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return GetUserOutput{}, err
	}

	// 3. Return DTO
	return GetUserOutput{
		User: ToUserDTO(userEntity),
	}, nil
}

// ========================================
// GetUserByTelegramUsername Use Case
// ========================================

// GetUserByTelegramUsernameUseCase handles the business logic for getting a user by Telegram username
type GetUserByTelegramUsernameUseCase struct {
	userRepo user.UserRepository
}

// NewGetUserByTelegramUsernameUseCase creates a new GetUserByTelegramUsernameUseCase
func NewGetUserByTelegramUsernameUseCase(userRepo user.UserRepository) *GetUserByTelegramUsernameUseCase {
	return &GetUserByTelegramUsernameUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves a user by Telegram username
func (uc *GetUserByTelegramUsernameUseCase) Execute(input GetUserByTelegramUsernameInput) (GetUserByTelegramUsernameOutput, error) {
	// 1. Validate input
	telegramUsername, err := user.NewTelegramUsername(input.TelegramUsername)
	if err != nil {
		return GetUserByTelegramUsernameOutput{}, err
	}

	// 2. Load user from repository
	userEntity, err := uc.userRepo.FindByTelegramUsername(telegramUsername)
	if err != nil {
		return GetUserByTelegramUsernameOutput{}, err
	}

	// 3. Return DTO
	return GetUserByTelegramUsernameOutput{
		User: ToUserDTO(userEntity),
	}, nil
}

// ========================================
// ListUsers Use Case (Admin)
// ========================================

// ListUsersUseCase handles the business logic for listing users
type ListUsersUseCase struct {
	userRepo user.UserRepository
}

// NewListUsersUseCase creates a new ListUsersUseCase
func NewListUsersUseCase(userRepo user.UserRepository) *ListUsersUseCase {
	return &ListUsersUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves a list of users with pagination
func (uc *ListUsersUseCase) Execute(input ListUsersInput) (ListUsersOutput, error) {
	// 1. Set default pagination
	limit := input.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	// 2. Load users from repository
	users, err := uc.userRepo.FindAll(limit, offset)
	if err != nil {
		return ListUsersOutput{}, err
	}

	// 3. Return DTOs
	return ListUsersOutput{
		Users: ToUserListDTO(users),
		Total: len(users), // TODO: Add total count query to repository
	}, nil
}
