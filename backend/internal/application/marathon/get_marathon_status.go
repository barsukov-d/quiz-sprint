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

	// 2. Default bonus inventory (what player starts with in a new game)
	defaultBonuses := solo_marathon.NewBonusInventory()
	defaultBonusDTO := BonusInventoryDTO{
		Shield:     defaultBonuses.Shield(),
		FiftyFifty: defaultBonuses.FiftyFifty(),
		Skip:       defaultBonuses.Skip(),
		Freeze:     defaultBonuses.Freeze(),
	}

	// 3. Find active game for player
	game, err := uc.marathonRepo.FindActiveByPlayer(playerID)
	if err != nil {
		if err == solo_marathon.ErrGameNotFound {
			// No active game - return default bonuses so UI can display them
			return GetMarathonStatusOutput{
				HasActiveGame:  false,
				BonusInventory: &defaultBonusDTO,
			}, nil
		}
		return GetMarathonStatusOutput{}, err
	}

	// 4. Game found - build output
	now := time.Now().Unix()
	gameDTO := ToMarathonGameDTOV2(game, now)
	gameBonusDTO := gameDTO.BonusInventory

	return GetMarathonStatusOutput{
		HasActiveGame:  true,
		Game:           &gameDTO,
		BonusInventory: &gameBonusDTO,
	}, nil
}
