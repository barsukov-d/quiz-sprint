package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// GetMarathonStatusUseCase handles retrieving active marathon game status
type GetMarathonStatusUseCase struct {
	marathonRepo solo_marathon.Repository
}

// NewGetMarathonStatusUseCase creates a new GetMarathonStatusUseCase
func NewGetMarathonStatusUseCase(
	marathonRepo solo_marathon.Repository,
) *GetMarathonStatusUseCase {
	return &GetMarathonStatusUseCase{
		marathonRepo: marathonRepo,
	}
}

// Execute retrieves the active marathon game for a player
func (uc *GetMarathonStatusUseCase) Execute(input GetMarathonStatusInput) (GetMarathonStatusOutput, error) {
	// 1. Validate and convert input to domain types
	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return GetMarathonStatusOutput{}, err
	}

	// 2. Find active game for player
	game, err := uc.marathonRepo.FindActiveByPlayer(playerID)
	if err != nil {
		if err == solo_marathon.ErrGameNotFound {
			// No active game - return empty result
			return GetMarathonStatusOutput{
				HasActiveGame: false,
			}, nil
		}
		return GetMarathonStatusOutput{}, err
	}

	// 3. Game found - build output
	now := time.Now().Unix()
	gameDTO := ToMarathonGameDTOV2(game, now)

	// Calculate time limit for current question
	timeLimit := GetTimeLimit(game.Difficulty(), game.CurrentStreak())

	return GetMarathonStatusOutput{
		HasActiveGame: true,
		Game:          &gameDTO,
		TimeLimit:     &timeLimit,
	}, nil
}
