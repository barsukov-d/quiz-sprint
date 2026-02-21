package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// StartMarathonUseCase handles starting a new marathon game
type StartMarathonUseCase struct {
	marathonRepo     solo_marathon.Repository
	personalBestRepo solo_marathon.PersonalBestRepository
	questionRepo     quiz.QuestionRepository
	categoryRepo     quiz.CategoryRepository
	eventBus         EventBus
	bonusWalletRepo  solo_marathon.BonusWalletRepository
}

// NewStartMarathonUseCase creates a new StartMarathonUseCase
func NewStartMarathonUseCase(
	marathonRepo solo_marathon.Repository,
	personalBestRepo solo_marathon.PersonalBestRepository,
	questionRepo quiz.QuestionRepository,
	categoryRepo quiz.CategoryRepository,
	eventBus EventBus,
	bonusWalletRepo solo_marathon.BonusWalletRepository,
) *StartMarathonUseCase {
	return &StartMarathonUseCase{
		marathonRepo:     marathonRepo,
		personalBestRepo: personalBestRepo,
		questionRepo:     questionRepo,
		categoryRepo:     categoryRepo,
		eventBus:         eventBus,
		bonusWalletRepo:  bonusWalletRepo,
	}
}

// Execute starts a new marathon game
func (uc *StartMarathonUseCase) Execute(input StartMarathonInput) (StartMarathonOutput, error) {
	// 1. Validate and convert input to domain types
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return StartMarathonOutput{}, err
	}

	// 2. Check if player already has an active game
	existingGame, err := uc.marathonRepo.FindActiveByPlayer(playerID)
	if err == nil && existingGame != nil {
		if existingGame.IsWaitingForContinue() {
			// game_over is an intermediate state — auto-complete it so a new run can start immediately
			now := time.Now().Unix()
			if err := existingGame.CompleteGame(now); err != nil {
				return StartMarathonOutput{}, err
			}
			if err := uc.marathonRepo.Save(existingGame); err != nil {
				return StartMarathonOutput{}, err
			}
		} else {
			// in_progress — player still has an active game
			return StartMarathonOutput{}, solo_marathon.ErrActiveGameExists
		}
	}

	// 3. Determine category
	var category solo_marathon.MarathonCategory
	if input.CategoryID == nil || *input.CategoryID == "" || *input.CategoryID == "all" {
		// "All categories" mode
		category = solo_marathon.NewMarathonCategoryAll()
	} else {
		// Specific category
		categoryID, err := quiz.NewCategoryIDFromString(*input.CategoryID)
		if err != nil {
			return StartMarathonOutput{}, err
		}

		// Validate category exists
		categoryAggregate, err := uc.categoryRepo.FindByID(categoryID)
		if err != nil {
			return StartMarathonOutput{}, err
		}

		category = solo_marathon.NewMarathonCategory(categoryID, categoryAggregate.Name().String())
	}

	// 4. Load PersonalBest for this category (if exists)
	personalBest, err := uc.personalBestRepo.FindByPlayerAndCategory(playerID, category)
	if err != nil && err != solo_marathon.ErrPersonalBestNotFound {
		return StartMarathonOutput{}, err
	}
	// personalBest can be nil - that's okay for first-time players

	// 5. Create MarathonGame aggregate (V2 with defaults + wallet bonuses)
	now := time.Now().Unix()
	defaultBonuses := solo_marathon.NewBonusInventory()
	var walletBonuses solo_marathon.BonusInventory
	if uc.bonusWalletRepo != nil {
		wallet, err := uc.bonusWalletRepo.FindByPlayer(playerID)
		if err == nil && wallet != nil {
			walletBonuses = wallet.ConsumeAll()
			_ = uc.bonusWalletRepo.Save(wallet) // Save zeroed wallet
		}
	}
	bonuses := defaultBonuses.Add(walletBonuses)
	game, err := solo_marathon.NewMarathonGameV2(
		playerID,
		category,
		personalBest,
		bonuses,
		now,
	)
	if err != nil {
		return StartMarathonOutput{}, err
	}

	// 6. Load first question using QuestionSelector Domain Service
	questionSelector := solo_marathon.NewQuestionSelector(uc.questionRepo)
	if err := game.LoadNextQuestion(questionSelector); err != nil {
		return StartMarathonOutput{}, err
	}

	// 7. Persist game
	if err := uc.marathonRepo.Save(game); err != nil {
		return StartMarathonOutput{}, err
	}

	// 8. Publish domain events
	if uc.eventBus != nil {
		events := game.Events()
		for _, event := range events {
			uc.eventBus.Publish(event)
		}
	}

	// 9. Build output DTO
	return StartMarathonOutput{
		Game:            ToMarathonGameDTOV2(game, now),
		HasPersonalBest: personalBest != nil,
	}, nil
}
