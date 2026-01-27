package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// AbandonMarathonUseCase handles abandoning a marathon game
type AbandonMarathonUseCase struct {
	marathonRepo     solo_marathon.Repository
	personalBestRepo solo_marathon.PersonalBestRepository
	eventBus         EventBus
}

// NewAbandonMarathonUseCase creates a new AbandonMarathonUseCase
func NewAbandonMarathonUseCase(
	marathonRepo solo_marathon.Repository,
	personalBestRepo solo_marathon.PersonalBestRepository,
	eventBus EventBus,
) *AbandonMarathonUseCase {
	return &AbandonMarathonUseCase{
		marathonRepo:     marathonRepo,
		personalBestRepo: personalBestRepo,
		eventBus:         eventBus,
	}
}

// Execute abandons a marathon game (player quits voluntarily)
func (uc *AbandonMarathonUseCase) Execute(input AbandonMarathonInput) (AbandonMarathonOutput, error) {
	// 1. Validate and convert input to domain types
	gameID := solo_marathon.NewGameIDFromString(input.GameID)
	if gameID.IsZero() {
		return AbandonMarathonOutput{}, solo_marathon.ErrInvalidGameID
	}

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return AbandonMarathonOutput{}, err
	}

	// 2. Load game aggregate
	game, err := uc.marathonRepo.FindByID(gameID)
	if err != nil {
		return AbandonMarathonOutput{}, err
	}

	// 3. Validate game belongs to player
	if !game.PlayerID().Equals(playerID) {
		return AbandonMarathonOutput{}, quiz.ErrUnauthorized
	}

	// 4. Abandon game (domain business logic)
	now := time.Now().Unix()
	if err := game.Abandon(now); err != nil {
		return AbandonMarathonOutput{}, err
	}

	// 5. Update PersonalBest if new record
	if game.IsNewPersonalBest() {
		personalBest, err := uc.personalBestRepo.FindByPlayerAndCategory(game.PlayerID(), game.Category())
		if err != nil && err != solo_marathon.ErrPersonalBestNotFound {
			// Log error but don't fail the request
			// TODO: Add logging
		}

		if personalBest == nil {
			// First time playing this category - create new record
			personalBest, err = solo_marathon.NewPersonalBest(
				game.PlayerID(),
				game.Category(),
				game.MaxStreak(),
				game.BaseScore(),
				now,
			)
			if err == nil {
				_ = uc.personalBestRepo.Save(personalBest)
			}
		} else {
			// Update existing record
			updated := personalBest.UpdateIfBetter(
				game.MaxStreak(),
				game.BaseScore(),
				now,
			)
			if updated {
				_ = uc.personalBestRepo.Save(personalBest)
			}
		}
	}

	// 6. Persist game
	if err := uc.marathonRepo.Save(game); err != nil {
		return AbandonMarathonOutput{}, err
	}

	// 7. Publish domain events
	if uc.eventBus != nil {
		events := game.Events()
		for _, event := range events {
			uc.eventBus.Publish(event)
		}
	}

	// 8. Build output
	// TODO: Get global rank from leaderboard repository
	var globalRank *int = nil

	return AbandonMarathonOutput{
		GameOverResult: BuildGameOverResultV2(game, globalRank),
	}, nil
}
