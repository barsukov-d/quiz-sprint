package routes

import (
	"database/sql"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"

	appQuiz "github.com/barsukov/quiz-sprint/backend/internal/application/quiz"
	appUser "github.com/barsukov/quiz-sprint/backend/internal/application/user"
	appMarathon "github.com/barsukov/quiz-sprint/backend/internal/application/marathon"
	appDaily "github.com/barsukov/quiz-sprint/backend/internal/application/daily_challenge"
	appDuel "github.com/barsukov/quiz-sprint/backend/internal/application/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	domainUser "github.com/barsukov/quiz-sprint/backend/internal/domain/user"
	domainMarathon "github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
	domainDaily "github.com/barsukov/quiz-sprint/backend/internal/domain/daily_challenge"
	domainDuel "github.com/barsukov/quiz-sprint/backend/internal/domain/quick_duel"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/handlers"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/http/middleware"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/messaging"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/persistence/memory"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/persistence/postgres"
	redisStore "github.com/barsukov/quiz-sprint/backend/internal/infrastructure/persistence/redis"
	"github.com/barsukov/quiz-sprint/backend/internal/infrastructure/telegram"

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

	// Session and Leaderboard: use PostgreSQL if available, fallback to memory
	var sessionRepo quiz.SessionRepository
	var leaderboardRepo interface {
		quiz.LeaderboardRepository
		quiz.GlobalLeaderboardRepository
	}

	if db != nil {
		sessionRepo = postgres.NewSessionRepository(db)
		leaderboardRepo = postgres.NewLeaderboardRepository(db)
	} else {
		memSessionRepo := memory.NewSessionRepository()
		sessionRepo = memSessionRepo
		leaderboardRepo = memory.NewLeaderboardRepository(memSessionRepo)
	}

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

	// Marathon repositories: only available with PostgreSQL
	var (
		questionRepo      quiz.QuestionRepository
		marathonRepo      domainMarathon.Repository
		personalBestRepo  domainMarathon.PersonalBestRepository
		bonusWalletRepo   domainMarathon.BonusWalletRepository
	)
	if db != nil {
		questionRepo = postgres.NewQuestionRepository(db)
		marathonRepo = postgres.NewMarathonRepository(db, questionRepo)
		personalBestRepo = postgres.NewPersonalBestRepository(db)
		bonusWalletRepo = postgres.NewBonusWalletRepository(db)
	}

	// Daily Challenge repositories: only available with PostgreSQL
	var (
		dailyQuizRepo domainDaily.DailyQuizRepository
		dailyGameRepo domainDaily.DailyGameRepository
	)
	if db != nil && quizRepo != nil && questionRepo != nil {
		dailyQuizRepo = postgres.NewDailyQuizRepository(db)
		dailyGameRepo = postgres.NewDailyGameRepository(db, quizRepo, questionRepo, dailyQuizRepo)
	}

	// Duel (PvP) repositories: only available with PostgreSQL
	var (
		duelGameRepo     domainDuel.DuelGameRepository
		playerRatingRepo domainDuel.PlayerRatingRepository
		challengeRepo    domainDuel.ChallengeRepository
		referralRepo     domainDuel.ReferralRepository
		seasonRepo       domainDuel.SeasonRepository
		matchmakingQueue domainDuel.MatchmakingQueue
		duelRoundCache   appDuel.DuelRoundCache
	)
	if db != nil {
		duelGameRepo = postgres.NewDuelGameRepository(db)
		playerRatingRepo = postgres.NewPlayerRatingRepository(db)
		challengeRepo = postgres.NewChallengeRepository(db)
		referralRepo = postgres.NewReferralRepository(db)
		seasonRepo = postgres.NewSeasonRepository(db)

		// Initialize Redis for matchmaking queue and round-answer cache
		redisClient, err := redisStore.NewClient()
		if err != nil {
			log.Printf("⚠️ Redis not available, using in-memory fallbacks: %v", err)
			duelRoundCache = appDuel.NewMemoryRoundCache()
		} else {
			log.Println("✅ Connected to Redis for matchmaking queue and round-answer cache")
			matchmakingQueue = redisStore.NewMatchmakingQueue(redisClient)
			duelRoundCache = redisStore.NewDuelRoundCache(redisClient)
		}
	}

	// ========================================
	// Infrastructure Layer: Event Bus
	// ========================================
	eventBus := messaging.NewInMemoryEventBus()
	loggingEventBus := messaging.NewLoggingEventBus(eventBus)

	// Marathon event bus (separate from quiz events)
	marathonEventBus := messaging.NewMarathonEventBus(true) // Enable logging

	// Daily Challenge event bus
	dailyChallengeEventBus := messaging.NewDailyChallengeEventBus(true) // Enable logging

	// Subscribe to ChestEarnedEvent: credit marathon bonuses to player's wallet
	if bonusWalletRepo != nil {
		dailyChallengeEventBus.Subscribe("chest_earned", func(event domainDaily.Event) {
			chestEvent, ok := event.(domainDaily.ChestEarnedEvent)
			if !ok {
				return
			}

			playerID := chestEvent.PlayerID()
			wallet, err := bonusWalletRepo.FindByPlayer(playerID)
			if err != nil || wallet == nil {
				wallet = domainMarathon.NewBonusWallet(playerID)
			}

			for _, bonus := range chestEvent.MarathonBonuses() {
				wallet.AddBonus(domainMarathon.BonusType(string(bonus)))
			}

			if err := bonusWalletRepo.Save(wallet); err != nil {
				log.Printf("[DAILY→MARATHON] Failed to credit bonuses: %v", err)
			} else {
				log.Printf("[DAILY→MARATHON] Credited %d bonuses to player %s",
					len(chestEvent.MarathonBonuses()), playerID.String())
			}
		})
	}

	// ========================================
	// Infrastructure Layer: WebSocket Hub
	// ========================================
	wsHub := handlers.NewWebSocketHub(leaderboardRepo)

	// ========================================
	// Infrastructure Layer: Event Handlers
	// ========================================
	// Register QuizCompletedEvent handler to broadcast leaderboard updates
	loggingEventBus.Subscribe("quiz.completed", func(event quiz.Event) {
		completedEvent, ok := event.(quiz.QuizCompletedEvent)
		if !ok {
			return
		}

		// Broadcast to per-quiz leaderboard WebSocket
		wsHub.BroadcastLeaderboardUpdate(completedEvent.QuizID().String())

		// Broadcast to global leaderboard WebSocket
		wsHub.BroadcastGlobalLeaderboardUpdate()
	})

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
	getGlobalLeaderboardUC := appQuiz.NewGetGlobalLeaderboardUseCase(leaderboardRepo)
	getActiveSessionUC := appQuiz.NewGetActiveSessionUseCase(quizRepo, sessionRepo)
	abandonSessionUC := appQuiz.NewAbandonSessionUseCase(sessionRepo)
	getSessionResultsUC := appQuiz.NewGetSessionResultsUseCase(sessionRepo, quizRepo)
	getDailyQuizUC := appQuiz.NewGetDailyQuizUseCase(quizRepo, sessionRepo, leaderboardRepo)
	getRandomQuizUC := appQuiz.NewGetRandomQuizUseCase(quizRepo)
	getUserActiveSessionsUC := appQuiz.NewGetUserActiveSessionsUseCase(quizRepo, sessionRepo)

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

	// Marathon use cases (only if database is available)
	var (
		startMarathonUC           *appMarathon.StartMarathonUseCase
		submitMarathonAnswerUC    *appMarathon.SubmitMarathonAnswerUseCase
		useMarathonBonusUC        *appMarathon.UseMarathonBonusUseCase
		continueMarathonUC        *appMarathon.ContinueMarathonUseCase
		abandonMarathonUC         *appMarathon.AbandonMarathonUseCase
		getMarathonStatusUC       *appMarathon.GetMarathonStatusUseCase
		getPersonalBestsUC        *appMarathon.GetPersonalBestsUseCase
		getMarathonLeaderboardUC  *appMarathon.GetMarathonLeaderboardUseCase
	)

	if marathonRepo != nil && personalBestRepo != nil && questionRepo != nil && categoryRepo != nil && userRepo != nil {
		startMarathonUC = appMarathon.NewStartMarathonUseCase(
			marathonRepo,
			personalBestRepo,
			questionRepo,
			categoryRepo,
			marathonEventBus,
			bonusWalletRepo,
		)
		submitMarathonAnswerUC = appMarathon.NewSubmitMarathonAnswerUseCase(
			marathonRepo,
			personalBestRepo,
			questionRepo,
			marathonEventBus,
		)
		useMarathonBonusUC = appMarathon.NewUseMarathonBonusUseCase(
			marathonRepo,
			questionRepo,
			marathonEventBus,
		)
		continueMarathonUC = appMarathon.NewContinueMarathonUseCase(
			marathonRepo,
			questionRepo,
			marathonEventBus,
		)
		abandonMarathonUC = appMarathon.NewAbandonMarathonUseCase(
			marathonRepo,
			personalBestRepo,
			marathonEventBus,
		)
		getMarathonStatusUC = appMarathon.NewGetMarathonStatusUseCase(
			marathonRepo,
			bonusWalletRepo,
		)
		getPersonalBestsUC = appMarathon.NewGetPersonalBestsUseCase(
			personalBestRepo,
		)
		getMarathonLeaderboardUC = appMarathon.NewGetMarathonLeaderboardUseCase(
			personalBestRepo,
			categoryRepo,
			userRepo,
		)
	}

	// Daily Challenge use cases (only if database is available)
	var (
		getOrCreateDailyQuizUC *appDaily.GetOrCreateDailyQuizUseCase
		startDailyChallengeUC  *appDaily.StartDailyChallengeUseCase
		submitDailyAnswerUC    *appDaily.SubmitDailyAnswerUseCase
		getDailyGameStatusUC   *appDaily.GetDailyGameStatusUseCase
		getDailyLeaderboardUC  *appDaily.GetDailyLeaderboardUseCase
		getPlayerStreakUC      *appDaily.GetPlayerStreakUseCase
		openChestUC            *appDaily.OpenChestUseCase
		retryUC                *appDaily.RetryChallengeUseCase
	)

	if dailyQuizRepo != nil && dailyGameRepo != nil && questionRepo != nil && quizRepo != nil && userRepo != nil {
		// Create ChestRewardCalculator with RNG
		// Using time-based seed for randomness per docs/game_modes/daily_challenge/04_rewards.md
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		chestRewardCalc := domainDaily.NewChestRewardCalculator(rng)

		getOrCreateDailyQuizUC = appDaily.NewGetOrCreateDailyQuizUseCase(
			dailyQuizRepo,
			dailyGameRepo,
			questionRepo,
			dailyChallengeEventBus,
		)
		startDailyChallengeUC = appDaily.NewStartDailyChallengeUseCase(
			dailyQuizRepo,
			dailyGameRepo,
			questionRepo,
			quizRepo,
			dailyChallengeEventBus,
			getOrCreateDailyQuizUC,
		)
		getDailyLeaderboardUC = appDaily.NewGetDailyLeaderboardUseCase(
			dailyGameRepo,
			userRepo,
		)
		submitDailyAnswerUC = appDaily.NewSubmitDailyAnswerUseCase(
			dailyGameRepo,
			dailyChallengeEventBus,
			getDailyLeaderboardUC,
			chestRewardCalc,
		)
		getDailyGameStatusUC = appDaily.NewGetDailyGameStatusUseCase(
			dailyQuizRepo,
			dailyGameRepo,
			getDailyLeaderboardUC,
		)
		getPlayerStreakUC = appDaily.NewGetPlayerStreakUseCase(
			dailyGameRepo,
		)
		openChestUC = appDaily.NewOpenChestUseCase(
			dailyGameRepo,
		)
		retryUC = appDaily.NewRetryChallengeUseCase(
			dailyGameRepo,
			dailyQuizRepo,
			questionRepo,
			dailyChallengeEventBus,
		)
	}

	// Duel (PvP) use cases (only if database is available)
	var (
		getDuelStatusUC        *appDuel.GetDuelStatusUseCase
		joinQueueUC            *appDuel.JoinQueueUseCase
		leaveQueueUC           *appDuel.LeaveQueueUseCase
		sendChallengeUC        *appDuel.SendChallengeUseCase
		respondChallengeUC     *appDuel.RespondChallengeUseCase
		acceptByLinkCodeUC     *appDuel.AcceptByLinkCodeUseCase
		startChallengeUC       *appDuel.StartChallengeUseCase
		createChallengeLinkUC  *appDuel.CreateChallengeLinkUseCase
		getGameHistoryUC       *appDuel.GetGameHistoryUseCase
		getDuelLeaderboardUC   *appDuel.GetLeaderboardUseCase
		requestRematchUC       *appDuel.RequestRematchUseCase
		startGameUC            *appDuel.StartGameUseCase
		submitDuelAnswerUC     *appDuel.SubmitDuelAnswerUseCase
		getGameResultUC        *appDuel.GetGameResultUseCase
	)

	if duelGameRepo != nil && playerRatingRepo != nil && challengeRepo != nil && referralRepo != nil && seasonRepo != nil && userRepo != nil {
		duelEventBus := appDuel.NewNoOpEventBus() // Simple event bus for now

		var telegramNotifier appDuel.TelegramNotifier = telegram.NewNoOpNotifier()
		if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
			telegramNotifier = telegram.NewHTTPNotifier(token)
		}

		getDuelStatusUC = appDuel.NewGetDuelStatusUseCase(
			playerRatingRepo,
			duelGameRepo,
			challengeRepo,
			seasonRepo,
			userRepo,
		)
		if matchmakingQueue != nil {
			joinQueueUC = appDuel.NewJoinQueueUseCase(
				matchmakingQueue,
				playerRatingRepo,
				duelGameRepo,
				seasonRepo,
			)
			leaveQueueUC = appDuel.NewLeaveQueueUseCase(
				matchmakingQueue,
			)
		}
		sendChallengeUC = appDuel.NewSendChallengeUseCase(
			challengeRepo,
			duelGameRepo,
			duelEventBus,
		)
		respondChallengeUC = appDuel.NewRespondChallengeUseCase(
			challengeRepo,
			duelGameRepo,
			playerRatingRepo,
			seasonRepo,
			duelEventBus,
		)
		acceptByLinkCodeUC = appDuel.NewAcceptByLinkCodeUseCase(
			challengeRepo,
			userRepo,
			telegramNotifier,
			duelEventBus,
		)
		// Build duel question adapter if questionRepo is available
		if questionRepo != nil {
			duelQuestionRepo := postgres.NewDuelQuestionRepositoryAdapter(questionRepo)
			startGameUC = appDuel.NewStartGameUseCase(
				duelGameRepo,
				playerRatingRepo,
				duelQuestionRepo,
				seasonRepo,
				duelEventBus,
			)
			startChallengeUC = appDuel.NewStartChallengeUseCase(
				challengeRepo,
				duelGameRepo,
				playerRatingRepo,
				seasonRepo,
				duelQuestionRepo,
				userRepo,
				duelEventBus,
			)
			submitDuelAnswerUC = appDuel.NewSubmitDuelAnswerUseCase(
				duelGameRepo,
				playerRatingRepo,
				duelQuestionRepo,
				seasonRepo,
				duelEventBus,
				duelRoundCache,
			)
		}
		createChallengeLinkUC = appDuel.NewCreateChallengeLinkUseCase(
			challengeRepo,
			duelEventBus,
		)
		getGameHistoryUC = appDuel.NewGetGameHistoryUseCase(
			duelGameRepo,
			userRepo,
		)
		getDuelLeaderboardUC = appDuel.NewGetLeaderboardUseCase(
			playerRatingRepo,
			referralRepo,
			seasonRepo,
			userRepo,
		)
		requestRematchUC = appDuel.NewRequestRematchUseCase(
			duelGameRepo,
			challengeRepo,
			duelEventBus,
		)
		getGameResultUC = appDuel.NewGetGameResultUseCase(
			duelGameRepo,
			playerRatingRepo,
			seasonRepo,
			userRepo,
		)

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
		getGlobalLeaderboardUC,
		getActiveSessionUC,
		abandonSessionUC,
		getSessionResultsUC,
		getDailyQuizUC,
		getRandomQuizUC,
		getUserActiveSessionsUC,
	)

	// Category handler (only if database is available)
	var categoryHandler *handlers.CategoryHandler
	if categoryRepo != nil {
		categoryHandler = handlers.NewCategoryHandler(createCategoryUC, listCategoriesUC)
	}

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

	// Marathon handler (only if database is available)
	var marathonHandler *handlers.MarathonHandler
	if startMarathonUC != nil {
		marathonHandler = handlers.NewMarathonHandler(
			startMarathonUC,
			submitMarathonAnswerUC,
			useMarathonBonusUC,
			continueMarathonUC,
			abandonMarathonUC,
			getMarathonStatusUC,
			getPersonalBestsUC,
			getMarathonLeaderboardUC,
		)
	}

	// Daily Challenge handler (only if database is available)
	var dailyChallengeHandler *handlers.DailyChallengeHandler
	if startDailyChallengeUC != nil {
		dailyChallengeHandler = handlers.NewDailyChallengeHandler(
			getOrCreateDailyQuizUC,
			startDailyChallengeUC,
			submitDailyAnswerUC,
			getDailyGameStatusUC,
			getDailyLeaderboardUC,
			getPlayerStreakUC,
			openChestUC,
			retryUC,
		)
	}

	// Duel (PvP) handler (only if database is available)
	var duelHandler *handlers.DuelHandler
	if getDuelStatusUC != nil {
		duelHandler = handlers.NewDuelHandler(
			getDuelStatusUC,
			joinQueueUC,
			leaveQueueUC,
			sendChallengeUC,
			respondChallengeUC,
			acceptByLinkCodeUC,
			startChallengeUC,
			createChallengeLinkUC,
			getGameHistoryUC,
			getDuelLeaderboardUC,
			requestRematchUC,
			getGameResultUC,
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
	quiz.Get("/daily", middleware.TelegramAuthMiddleware(), quizHandler.GetDailyQuiz) // Daily quiz with auth (before /:id)
	quiz.Get("/random", quizHandler.GetRandomQuiz)                                    // Random quiz (before /:id)
	quiz.Get("/:id", quizHandler.GetQuizByID)
	quiz.Post("/:id/start", quizHandler.StartQuiz)
	quiz.Get("/:id/active-session", quizHandler.GetActiveSession)
	quiz.Get("/:id/leaderboard", quizHandler.GetLeaderboard)

	// Global leaderboard route
	v1.Get("/leaderboard", quizHandler.GetGlobalLeaderboard)

	// Session routes
	session := v1.Group("/quiz/session")
	// IMPORTANT: More specific routes MUST come before generic /:sessionId
	session.Post("/:sessionId/answer", quizHandler.SubmitAnswer)
	session.Get("/:sessionId", quizHandler.GetSessionResults)
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
	ws.Get("/leaderboard/global", websocket.New(wsHub.HandleGlobalLeaderboardWebSocket))

	// Duel WebSocket (if database available)
	if duelGameRepo != nil {
		// startGameUC and submitDuelAnswerUC require a duel-specific QuestionRepository
		// adapter (appDuel.QuestionRepository) that does not exist yet. They will be nil
		// until that adapter is implemented; the hub handles nil use cases gracefully.
		duelWsHub := handlers.NewDuelWebSocketHub(startGameUC, submitDuelAnswerUC)
		ws.Get("/duel/:gameId", websocket.New(duelWsHub.HandleDuelWebSocket))
	}

	// User routes (only if database is available)
	if userHandler != nil {
		user := v1.Group("/user")

		// Protected routes - require Telegram authentication
		user.Post("/register", middleware.TelegramAuthMiddleware(), userHandler.RegisterUser)

		// Public routes (for now - can add auth later)
		user.Get("/:id", userHandler.GetUser)
		user.Get("/:userId/sessions/active", quizHandler.GetUserActiveSessions) // Active sessions
		user.Put("/:id", userHandler.UpdateUserProfile)
		user.Get("/by-username/:username", userHandler.GetUserByTelegramUsername)

		// Admin routes
		users := v1.Group("/users")
		users.Get("/", userHandler.ListUsers)
	}

	// Marathon routes (only if database is available)
	if marathonHandler != nil {
		marathon := v1.Group("/marathon")
		marathon.Post("/start", marathonHandler.StartMarathon)
		marathon.Post("/:gameId/answer", marathonHandler.SubmitMarathonAnswer)
		marathon.Post("/:gameId/bonus", marathonHandler.UseMarathonBonus)
		marathon.Post("/:gameId/continue", marathonHandler.ContinueMarathon)
		marathon.Delete("/:gameId", marathonHandler.AbandonMarathon)
		marathon.Get("/status", marathonHandler.GetMarathonStatus)
		marathon.Get("/personal-bests", marathonHandler.GetPersonalBests)
		marathon.Get("/leaderboard", marathonHandler.GetMarathonLeaderboard)
	}

	// Daily Challenge routes (only if database is available)
	if dailyChallengeHandler != nil {
		daily := v1.Group("/daily-challenge")
		daily.Post("/start", dailyChallengeHandler.StartDailyChallenge)
		daily.Post("/:gameId/answer", dailyChallengeHandler.SubmitDailyAnswer)
		daily.Post("/:gameId/chest/open", dailyChallengeHandler.OpenChest)
		daily.Post("/:gameId/retry", dailyChallengeHandler.RetryChallenge)
		daily.Get("/status", dailyChallengeHandler.GetDailyStatus)
		daily.Get("/leaderboard", dailyChallengeHandler.GetDailyLeaderboard)
		daily.Get("/streak", dailyChallengeHandler.GetPlayerStreak)
	}

	// Duel (PvP) routes (only if database is available)
	if duelHandler != nil {
		duel := v1.Group("/duel")
		duel.Get("/status", duelHandler.GetDuelStatus)
		duel.Post("/queue/join", duelHandler.JoinQueue)
		duel.Delete("/queue/leave", duelHandler.LeaveQueue)
		duel.Post("/challenge", duelHandler.SendChallenge)
		duel.Post("/challenge/link", duelHandler.CreateChallengeLink)
		duel.Post("/challenge/accept-by-code", duelHandler.AcceptByLinkCode)
		duel.Post("/challenge/:challengeId/start", duelHandler.StartChallenge)
		duel.Post("/challenge/:challengeId/respond", duelHandler.RespondChallenge)
		duel.Get("/history", duelHandler.GetGameHistory)
		duel.Get("/leaderboard", duelHandler.GetDuelLeaderboard)
		duel.Get("/game/:gameId", duelHandler.GetGameResult)
		duel.Post("/game/:gameId/rematch", duelHandler.RequestRematch)
	}

	// Admin routes (debug/testing, protected by API key)
	if db != nil {
		adminHandler := handlers.NewAdminHandler(db)
		admin := v1.Group("/admin", handlers.AdminKeyMiddleware())
		adminDaily := admin.Group("/daily-challenge")
		adminDaily.Patch("/streak", adminHandler.UpdateStreak)
		adminDaily.Delete("/games", adminHandler.DeleteGames)
		adminDaily.Get("/games", adminHandler.ListGames)
		adminDaily.Post("/simulate-streak", adminHandler.SimulateStreak)

		// Player-wide admin
		admin.Delete("/player/reset", adminHandler.ResetPlayer)

		// Marathon admin
		adminMarathon := admin.Group("/marathon")
		adminMarathon.Patch("/game", adminHandler.UpdateMarathonGame)
		adminMarathon.Get("/games", adminHandler.ListMarathonGames)
		adminMarathon.Delete("/games", adminHandler.DeleteMarathonGames)
	}

	// Swagger documentation
	app.Get("/swagger/*", swaggo.New(swaggo.Config{
		URL: "/swagger/doc.json",
	}))
}
