package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// GetMarathonStatusUseCase handles retrieving active marathon game status
type GetMarathonStatusUseCase struct {
	marathonRepo    solo_marathon.Repository
	bonusWalletRepo solo_marathon.BonusWalletRepository
}

// NewGetMarathonStatusUseCase creates a new GetMarathonStatusUseCase
func NewGetMarathonStatusUseCase(
	marathonRepo solo_marathon.Repository,
	bonusWalletRepo solo_marathon.BonusWalletRepository,
) *GetMarathonStatusUseCase {
	return &GetMarathonStatusUseCase{
		marathonRepo:    marathonRepo,
		bonusWalletRepo: bonusWalletRepo,
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
			// No active game - return defaults + wallet bonuses so UI can display them
			combined := solo_marathon.NewBonusInventory()
			if uc.bonusWalletRepo != nil {
				wallet, walletErr := uc.bonusWalletRepo.FindByPlayer(playerID)
				if walletErr == nil && wallet != nil {
					combined = combined.Add(wallet.ToBonusInventory())
				}
			}
			bonusDTO := BonusInventoryDTO{
				Shield:     combined.Shield(),
				FiftyFifty: combined.FiftyFifty(),
				Skip:       combined.Skip(),
				Freeze:     combined.Freeze(),
			}
			return GetMarathonStatusOutput{
				HasActiveGame:  false,
				BonusInventory: &bonusDTO,
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
