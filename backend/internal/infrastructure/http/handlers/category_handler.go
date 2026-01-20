package handlers

import (
	"github.com/gofiber/fiber/v3"

	appQuiz "github.com/barsukov/quiz-sprint/backend/internal/application/quiz"
)

// CategoryHandler handles HTTP requests for categories.
type CategoryHandler struct {
	createCategoryUC *appQuiz.CreateCategoryUseCase
	listCategoriesUC *appQuiz.ListCategoriesUseCase
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(
	createCategoryUC *appQuiz.CreateCategoryUseCase,
	listCategoriesUC *appQuiz.ListCategoriesUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		createCategoryUC: createCategoryUC,
		listCategoriesUC: listCategoriesUC,
	}
}

// CreateCategory handles POST /api/v1/categories
// @Summary Create a new category
// @Description Create a new quiz category
// @Tags category
// @Accept json
// @Produce json
// @Param request body handlers.CreateCategoryRequest true "Category data"
// @Success 201 {object} handlers.CreateCategoryResponse "Category created"
// @Failure 400 {object} handlers.ErrorResponse "Invalid request"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c fiber.Ctx) error {
	var req appQuiz.CreateCategoryInput
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	output, err := h.createCategoryUC.Execute(req)
	if err != nil {
		return mapError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": output.Category})
}

// GetAllCategories handles GET /api/v1/categories
// @Summary List all categories
// @Description Get a list of all available quiz categories
// @Tags category
// @Accept json
// @Produce json
// @Success 200 {object} handlers.ListCategoriesResponse "List of categories"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Router /categories [get]
func (h *CategoryHandler) GetAllCategories(c fiber.Ctx) error {
	output, err := h.listCategoriesUC.Execute(appQuiz.ListCategoriesInput{})
	if err != nil {
		return mapError(err)
	}

	return c.JSON(fiber.Map{"data": output.Categories})
}
