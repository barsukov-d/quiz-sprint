package user

import (
	"github.com/barsukov/quiz-sprint/backend/internal/domain/user"
)

// ========================================
// Domain → DTO Mappers
// ========================================

// ToUserDTO converts a User entity to UserDTO
func ToUserDTO(u *user.User) UserDTO {
	return UserDTO{
		ID:               u.ID().String(),
		Username:         u.Username().String(),
		TelegramUsername: u.TelegramUsername().String(),
		Email:            u.Email().String(),
		AvatarURL:        u.AvatarURL().String(),
		LanguageCode:     u.LanguageCode().String(),
		IsBlocked:        u.IsBlocked(),
		CreatedAt:        u.CreatedAt(),
		UpdatedAt:        u.UpdatedAt(),
	}
}

// ToUserProfileDTO converts a User entity to UserProfileDTO
func ToUserProfileDTO(u *user.User) UserProfileDTO {
	return UserProfileDTO{
		ID:               u.ID().String(),
		Username:         u.Username().String(),
		TelegramUsername: u.TelegramUsername().String(),
		AvatarURL:        u.AvatarURL().String(),
	}
}

// ToUserListDTO converts a slice of User entities to DTOs
func ToUserListDTO(users []user.User) []UserDTO {
	dtos := make([]UserDTO, 0, len(users))
	for i := range users {
		dtos = append(dtos, ToUserDTO(&users[i]))
	}
	return dtos
}
