package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// CompleteMarathonUseCase handles completing a marathon game (player declines continue)
type CompleteMarathonUseCase struct {
	marathonRepo     solo_marathon.Repository
	personalBestRepo solo_marathon.PersonalBestRepository
	eventBus         EventBus
}

// NewCompleteMarathonUseCase creates a new CompleteMarathonUseCase
func NewCompleteMarathonUseCase(
	marathonRepo solo_marathon.Repository,
	personalBestRepo solo_marathon.PersonalBestRepository,
	eventBus EventBus,
) *CompleteMarathonUseCase {
	return &CompleteMarathonUseCase{
		marathonRepo:     marathonRepo,
		personalBestRepo: personalBestRepo,
		eventBus:         eventBus,
	}
}

// Execute completes a marathon game that is in game_over state (player declined continue)
func (uc *CompleteMarathonUseCase) Execute(input CompleteMarathonInput) (CompleteMarathonOutput, error) {
	// 1. Validate and convert input to domain types
	gameID := solo_marathon.NewGameIDFromString(input.GameID)
	if gameID.IsZero() {
		return CompleteMarathonOutput{}, solo_marathon.ErrInvalidGameID
	}

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return CompleteMarathonOutput{}, err
	}

	// 2. Load game aggregate
	game, err := uc.marathonRepo.FindByID(gameID)
	if err != nil {
		return CompleteMarathonOutput{}, err
	}

	// 3. Validate game belongs to player
	if !game.PlayerID().Equals(playerID) {
		return CompleteMarathonOutput{}, quiz.ErrUnauthorized
	}

	// 4. Complete game (domain business logic — requires game_over status)
	now := time.Now().Unix()
	if err := game.CompleteGame(now); err != nil {
		return CompleteMarathonOutput{}, err
	}

	// 5. Update personal best if this is a new record
	uc.updatePersonalBestIfNeeded(game, now)

	// 6. Persist game
	if err := uc.marathonRepo.Save(game); err != nil {
		return CompleteMarathonOutput{}, err
	}

	// 7. Publish domain events
	if uc.eventBus != nil {
		events := game.Events()
		for _, event := range events {
			uc.eventBus.Publish(event)
		}
	}

	// 8. Build output
	return CompleteMarathonOutput{
		GameOverResult: BuildGameOverResultV2(game),
	}, nil
}

// updatePersonalBestIfNeeded updates the personal best record if the game score is better
func (uc *CompleteMarathonUseCase) updatePersonalBestIfNeeded(game *solo_marathon.MarathonGameV2, now int64) {
	if !game.IsNewPersonalBest() {
		return
	}

	personalBest, err := uc.personalBestRepo.FindByPlayerAndCategory(game.PlayerID(), game.Category())
	if err != nil && err != solo_marathon.ErrPersonalBestNotFound {
		return
	}

	if personalBest == nil {
		// First time playing this category — create new record
		personalBest, err = solo_marathon.NewPersonalBest(
			game.PlayerID(),
			game.Category(),
			game.Score(),
			game.Score(),
			now,
		)
		if err == nil {
			_ = uc.personalBestRepo.Save(personalBest)
		}
	} else {
		// Update existing record
		updated := personalBest.UpdateIfBetter(
			game.Score(),
			game.Score(),
			now,
		)
		if updated {
			_ = uc.personalBestRepo.Save(personalBest)
		}
	}
}
