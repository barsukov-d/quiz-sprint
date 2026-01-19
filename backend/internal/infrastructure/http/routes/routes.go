package routes

import (
	"database/sql"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"

	appQuiz "github.com/barsukov/quiz-sprint/backend/internal/application/quiz"
	appUser "github.com/barsukov/quiz-sprint/backend/internal/application/user"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/handlers"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/middleware"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/messaging"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/persistence/memory"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/persistence/postgres"

	"github.com/gofiber/contrib/v3/swaggo"
)

// SetupRoutes configures all application routes
// db can be nil if PostgreSQL is not available (user endpoints will be disabled)
func SetupRoutes(app *fiber.App, db *sql.DB) {
	// ========================================
	// Infrastructure Layer: Repositories
	// ========================================

	// Quiz repository: use PostgreSQL if available, fallback to memory
	var quizRepo quiz.QuizRepository
	if db != nil {
		quizRepo = postgres.NewQuizRepository(db)
	} else {
		// quizRepo = memory.NewQuizRepository()
	}

	// Session and Leaderboard: still use memory for now
	sessionRepo := memory.NewSessionRepository()
	leaderboardRepo := memory.NewLeaderboardRepository(sessionRepo)

	// User repository: use PostgreSQL if available
	var userRepo domainUser.UserRepository
	if db != nil {
		userRepo = postgres.NewUserRepository(db)
	}

	// Category repository: only available with PostgreSQL
	var categoryRepo quiz.CategoryRepository
	if db != nil {
		categoryRepo = postgres.NewCategoryRepository(db)
	}

	// ========================================
	// Infrastructure Layer: Event Bus
	// ========================================
	eventBus := messaging.NewInMemoryEventBus()
	loggingEventBus := messaging.NewLoggingEventBus(eventBus)

	// ========================================
	// Application Layer: Use Cases
	// ========================================

	// Quiz use cases
	listQuizzesUC := appQuiz.NewListQuizzesUseCase(quizRepo)
	getQuizUC := appQuiz.NewGetQuizUseCase(quizRepo)
	getQuizDetailsUC := appQuiz.NewGetQuizDetailsUseCase(quizRepo, leaderboardRepo)
	startQuizUC := appQuiz.NewStartQuizUseCase(quizRepo, sessionRepo, loggingEventBus)
	submitAnswerUC := appQuiz.NewSubmitAnswerUseCase(quizRepo, sessionRepo, loggingEventBus)
	getLeaderboardUC := appQuiz.NewGetLeaderboardUseCase(leaderboardRepo)
	getActiveSessionUC := appQuiz.NewGetActiveSessionUseCase(quizRepo, sessionRepo)
	abandonSessionUC := appQuiz.NewAbandonSessionUseCase(sessionRepo)

	// Category use cases (only if database is available)
	var (
		listCategoriesUC *appQuiz.ListCategoriesUseCase
		createCategoryUC *appQuiz.CreateCategoryUseCase
	)
	if categoryRepo != nil {
		listCategoriesUC = appQuiz.NewListCategoriesUseCase(categoryRepo)
		createCategoryUC = appQuiz.NewCreateCategoryUseCase(categoryRepo)
	}

	// User use cases (only if database is available)
	var (
		registerUserUC       *appUser.RegisterUserUseCase
		getUserUC            *appUser.GetUserUseCase
		updateUserProfileUC  *appUser.UpdateUserProfileUseCase
		updateUserLanguageUC *appUser.UpdateUserLanguageUseCase
		listUsersUC          *appUser.ListUsersUseCase
		getUserByUsernameUC  *appUser.GetUserByTelegramUsernameUseCase
	)

	if userRepo != nil {
		registerUserUC = appUser.NewRegisterUserUseCase(userRepo)
		getUserUC = appUser.NewGetUserUseCase(userRepo)
		updateUserProfileUC = appUser.NewUpdateUserProfileUseCase(userRepo)
		updateUserLanguageUC = appUser.NewUpdateUserLanguageUseCase(userRepo)
		listUsersUC = appUser.NewListUsersUseCase(userRepo)
		getUserByUsernameUC = appUser.NewGetUserByTelegramUsernameUseCase(userRepo)
	}

	// ========================================
	// Infrastructure Layer: HTTP Handlers
	// ========================================
	quizHandler := handlers.NewQuizHandler(
		listQuizzesUC,
		getQuizUC,
		getQuizDetailsUC,
		startQuizUC,
		submitAnswerUC,
		getLeaderboardUC,
		getActiveSessionUC,
		abandonSessionUC,
	)

	// Category handler (only if database is available)
	var categoryHandler *handlers.CategoryHandler
	if categoryRepo != nil {
		categoryHandler = handlers.NewCategoryHandler(createCategoryUC, listCategoriesUC)
	}

	wsHub := handlers.NewWebSocketHub(leaderboardRepo)

	// User handler (only if database is available)
	var userHandler *handlers.UserHandler
	if userRepo != nil {
		userHandler = handlers.NewUserHandler(
			registerUserUC,
			getUserUC,
			updateUserProfileUC,
			updateUserLanguageUC,
			listUsersUC,
			getUserByUsernameUC,
		)
	}

	// ========================================
	// Routes
	// ========================================

	// API v1 routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Quiz routes
	quiz := v1.Group("/quiz")
	quiz.Get("/", quizHandler.GetAllQuizzes)
	quiz.Get("/:id", quizHandler.GetQuizByID)
	quiz.Post("/:id/start", quizHandler.StartQuiz)
	quiz.Get("/:id/active-session", quizHandler.GetActiveSession)
	quiz.Get("/:id/leaderboard", quizHandler.GetLeaderboard)

	// Session routes
	session := v1.Group("/quiz/session")
	session.Post("/:sessionId/answer", quizHandler.SubmitAnswer)
	session.Delete("/:sessionId", quizHandler.AbandonSession)

	// Category routes
	if categoryHandler != nil {
		categories := v1.Group("/categories")
		categories.Get("/", categoryHandler.GetAllCategories)
		categories.Post("/", categoryHandler.CreateCategory) // Maybe add auth later
	}

	// WebSocket routes
	ws := app.Group("/ws")

	// WebSocket upgrade middleware
	ws.Use("/", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	ws.Get("/leaderboard/:id", websocket.New(wsHub.HandleLeaderboardWebSocket))

	// User routes (only if database is available)
	if userHandler != nil {
		user := v1.Group("/user")

		// Protected routes - require Telegram authentication
		user.Post("/register", middleware.TelegramAuthMiddleware(), userHandler.RegisterUser)

		// Public routes (for now - can add auth later)
		user.Get("/:id", userHandler.GetUser)
		user.Put("/:id", userHandler.UpdateUserProfile)
		user.Get("/by-username/:username", userHandler.GetUserByTelegramUsername)

		// Admin routes
		users := v1.Group("/users")
		users.Get("/", userHandler.ListUsers)
	}

	// Swagger documentation
	app.Get("/swagger/*", swaggo.New(swaggo.Config{
		URL: "/swagger/doc.json",
	}))
}
