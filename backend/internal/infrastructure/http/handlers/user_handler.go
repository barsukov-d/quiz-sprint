package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	appUser "github.com/barsukov/quiz-sprint/backend/internal/application/user"
	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/middleware"
)

// UserHandler handles HTTP requests for users
// NOTE: This is a THIN adapter - no business logic here!
type UserHandler struct {
	registerUserUC       *appUser.RegisterUserUseCase
	getUserUC            *appUser.GetUserUseCase
	updateUserProfileUC  *appUser.UpdateUserProfileUseCase
	updateUserLanguageUC *appUser.UpdateUserLanguageUseCase
	listUsersUC          *appUser.ListUsersUseCase
	getUserByUsernameUC  *appUser.GetUserByTelegramUsernameUseCase
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(
	registerUserUC *appUser.RegisterUserUseCase,
	getUserUC *appUser.GetUserUseCase,
	updateUserProfileUC *appUser.UpdateUserProfileUseCase,
	updateUserLanguageUC *appUser.UpdateUserLanguageUseCase,
	listUsersUC *appUser.ListUsersUseCase,
	getUserByUsernameUC *appUser.GetUserByTelegramUsernameUseCase,
) *UserHandler {
	return &UserHandler{
		registerUserUC:       registerUserUC,
		getUserUC:            getUserUC,
		updateUserProfileUC:  updateUserProfileUC,
		updateUserLanguageUC: updateUserLanguageUC,
		listUsersUC:          listUsersUC,
		getUserByUsernameUC:  getUserByUsernameUC,
	}
}

// ========================================
// Handlers (Thin Adapters)
// ========================================

// RegisterUser handles POST /api/v1/user/register
// @Summary Register or update user
// @Description Register a new user from Telegram Mini App or update existing user profile (idempotent). Requires valid Telegram init data in Authorization header.
// @Tags user
// @Accept json
// @Produce json
// @Security TelegramAuth
// @Success 200 {object} RegisterUserResponse "User registered or updated"
// @Failure 401 {object} ErrorResponse "Invalid or missing Telegram authorization"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/register [post]
func (h *UserHandler) RegisterUser(c fiber.Ctx) error {
	fmt.Println("📥 RegisterUser handler called")

	// 1. Extract validated Telegram init data from middleware
	initData := middleware.GetTelegramInitData(c)
	fmt.Printf("🔍 initData from context: %v\n", initData)

	if initData == nil {
		fmt.Println("❌ No init data in context!")
		return fiber.NewError(fiber.StatusUnauthorized, "No validated Telegram init data found")
	}

	// 2. Extract user data from init data
	// Check if User field is empty (ID == 0 means no user)
	if initData.User.ID == 0 {
		fmt.Println("❌ User ID is 0!")
		return fiber.NewError(fiber.StatusBadRequest, "No user data in init data")
	}

	fmt.Printf("✅ Got user from init data: id=%d, username=%s\n", initData.User.ID, initData.User.Username)

	user := &initData.User

	// Build username from first name + last name
	username := user.FirstName
	if user.LastName != "" {
		username = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}

	// 3. Execute use case with data from validated init data
	output, err := h.registerUserUC.Execute(appUser.RegisterUserInput{
		UserID:           fmt.Sprintf("%d", user.ID),
		Username:         username,
		TelegramUsername: user.Username,
		AvatarURL:        user.PhotoURL,
		LanguageCode:     user.LanguageCode,
	})
	if err != nil {
		return mapUserError(err)
	}

	// 4. Return response
	return c.JSON(fiber.Map{
		"data": output,
	})
}

// GetUser handles GET /api/v1/user/:id
// @Summary Get user by ID
// @Description Retrieve user profile by Telegram user ID
// @Tags user
// @Accept json
// @Produce json
// @Param id path string true "User ID (Telegram ID)"
// @Success 200 {object} GetUserResponse "User profile"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/{id} [get]
func (h *UserHandler) GetUser(c fiber.Ctx) error {
	// 1. Extract path parameter
	userID := c.Params("id")

	// 2. Execute use case
	output, err := h.getUserUC.Execute(appUser.GetUserInput{
		UserID: userID,
	})
	if err != nil {
		return mapUserError(err)
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data": output.User,
	})
}

// UpdateUserProfile handles PUT /api/v1/user/:id
// @Summary Update user profile
// @Description Update user profile information
// @Tags user
// @Accept json
// @Produce json
// @Param id path string true "User ID (Telegram ID)"
// @Param request body UpdateUserProfileRequest true "Updated profile data"
// @Success 200 {object} UpdateUserProfileResponse "Updated user profile"
// @Failure 400 {object} ErrorResponse "Invalid request body or user ID"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/{id} [put]
func (h *UserHandler) UpdateUserProfile(c fiber.Ctx) error {
	// 1. Parse request body
	var req UpdateUserProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// 2. Extract path parameter
	userID := c.Params("id")

	// 3. Execute use case
	output, err := h.updateUserProfileUC.Execute(appUser.UpdateUserProfileInput{
		UserID:           userID,
		Username:         req.Username,
		TelegramUsername: req.TelegramUsername,
		Email:            req.Email,
		AvatarURL:        req.AvatarURL,
		LanguageCode:     req.LanguageCode,
	})
	if err != nil {
		return mapUserError(err)
	}

	// 4. Return response
	return c.JSON(fiber.Map{
		"data": output.User,
	})
}

// ListUsers handles GET /api/v1/users
// @Summary List all users (Admin)
// @Description Get a paginated list of all users
// @Tags user
// @Accept json
// @Produce json
// @Param limit query int false "Limit (default 50, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} ListUsersResponse "List of users"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users [get]
func (h *UserHandler) ListUsers(c fiber.Ctx) error {
	// 1. Parse query parameters
	limit := fiber.Query[int](c, "limit", 50)
	offset := fiber.Query[int](c, "offset", 0)

	// 2. Execute use case
	output, err := h.listUsersUC.Execute(appUser.ListUsersInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list users")
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data":  output.Users,
		"total": output.Total,
	})
}

// GetUserByTelegramUsername handles GET /api/v1/user/by-username/:username
// @Summary Get user by Telegram username
// @Description Retrieve user profile by Telegram @username
// @Id GetUserByUsername
// @Tags user
// @Accept json
// @Produce json
// @Param username path string true "Telegram username (without @)"
// @Success 200 {object} GetUserResponse "User profile"
// @Failure 400 {object} ErrorResponse "Invalid username"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/by-username/{username} [get]
func (h *UserHandler) GetUserByTelegramUsername(c fiber.Ctx) error {
	// 1. Extract path parameter
	username := c.Params("username")

	// 2. Execute use case
	output, err := h.getUserByUsernameUC.Execute(appUser.GetUserByTelegramUsernameInput{
		TelegramUsername: username,
	})
	if err != nil {
		return mapUserError(err)
	}

	// 3. Return response
	return c.JSON(fiber.Map{
		"data": output.User,
	})
}

// ========================================
// Error Mapping
// ========================================

// mapUserError maps domain errors to HTTP errors
func mapUserError(err error) error {
	switch err {
	case domainUser.ErrUserNotFound:
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	case domainUser.ErrInvalidUserID:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	case domainUser.ErrUsernameTooLong:
		return fiber.NewError(fiber.StatusBadRequest, "Username too long")
	case domainUser.ErrInvalidTelegramUsername:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid Telegram username")
	case domainUser.ErrInvalidEmail, domainUser.ErrEmailTooLong:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid email")
	case domainUser.ErrInvalidAvatarURL, domainUser.ErrAvatarURLTooLong:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid avatar URL")
	case domainUser.ErrInvalidLanguageCode:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid language code")
	case domainUser.ErrUserBlocked:
		return fiber.NewError(fiber.StatusForbidden, "User is blocked")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
}
