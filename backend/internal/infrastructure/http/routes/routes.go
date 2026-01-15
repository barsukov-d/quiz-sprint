package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	appQuiz "github.com/barsukov/quiz-sprint/backend/internal/application/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/handlers"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/messaging"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/persistence/memory"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App) {
	// ========================================
	// Infrastructure Layer: Repositories
	// ========================================
	quizRepo := memory.NewQuizRepository()
	sessionRepo := memory.NewSessionRepository()
	leaderboardRepo := memory.NewLeaderboardRepository(sessionRepo)

	// ========================================
	// Infrastructure Layer: Event Bus
	// ========================================
	eventBus := messaging.NewInMemoryEventBus()
	loggingEventBus := messaging.NewLoggingEventBus(eventBus)

	// ========================================
	// Application Layer: Use Cases
	// ========================================
	listQuizzesUC := appQuiz.NewListQuizzesUseCase(quizRepo)
	getQuizUC := appQuiz.NewGetQuizUseCase(quizRepo)
	getQuizDetailsUC := appQuiz.NewGetQuizDetailsUseCase(quizRepo, leaderboardRepo)
	startQuizUC := appQuiz.NewStartQuizUseCase(quizRepo, sessionRepo, loggingEventBus)
	submitAnswerUC := appQuiz.NewSubmitAnswerUseCase(quizRepo, sessionRepo, loggingEventBus)
	getLeaderboardUC := appQuiz.NewGetLeaderboardUseCase(leaderboardRepo)

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
	)

	wsHub := handlers.NewWebSocketHub(leaderboardRepo)

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
	quiz.Get("/:id/leaderboard", quizHandler.GetLeaderboard)

	// Session routes
	session := v1.Group("/quiz/session")
	session.Post("/:sessionId/answer", quizHandler.SubmitAnswer)

	// WebSocket routes
	ws := app.Group("/ws")

	// WebSocket upgrade middleware
	ws.Use("/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	ws.Get("/leaderboard/:id", websocket.New(wsHub.HandleLeaderboardWebSocket))
}
