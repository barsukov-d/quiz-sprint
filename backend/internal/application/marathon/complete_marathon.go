package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// CompleteMarathonUseCase handles completing a marathon game (player declines continue)
type CompleteMarathonUseCase struct {
	marathonRepo      solo_marathon.Repository
	personalBestRepo  solo_marathon.PersonalBestRepository
	eventBus          EventBus
	inventoryService  InventoryService
	milestoneClaimsRepo MilestoneClaimsRepository
}

// NewCompleteMarathonUseCase creates a new CompleteMarathonUseCase
func NewCompleteMarathonUseCase(
	marathonRepo solo_marathon.Repository,
	personalBestRepo solo_marathon.PersonalBestRepository,
	eventBus EventBus,
	inventoryService InventoryService,
) *CompleteMarathonUseCase {
	return &CompleteMarathonUseCase{
		marathonRepo:     marathonRepo,
		personalBestRepo: personalBestRepo,
		eventBus:         eventBus,
		inventoryService: inventoryService,
	}
}

// WithMilestoneClaimsRepository attaches the deduplication repository for milestone rewards.
func (uc *CompleteMarathonUseCase) WithMilestoneClaimsRepository(repo MilestoneClaimsRepository) *CompleteMarathonUseCase {
	uc.milestoneClaimsRepo = repo
	return uc
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

	// 5b. Credit milestone rewards for the final score
	uc.creditMilestoneRewards(input.PlayerID, game.Score())

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

// personalBestMilestoneRewards defines coins awarded when a new personal best
// reaches specific milestone thresholds for the first time.
var personalBestMilestoneRewards = map[int]int{
	25:  100,
	50:  250,
	100: 500,
	200: 1000,
	500: 5000,
}

// creditMilestoneRewards credits coins for each milestone the player's score reaches.
// Only called on CompleteGame (not Abandon), matching the personal-best save logic.
// Deduplication via MilestoneClaimsRepository ensures each milestone is credited only once.
func (uc *CompleteMarathonUseCase) creditMilestoneRewards(playerID string, score int) {
	if uc.inventoryService == nil {
		return
	}
	for threshold, coins := range personalBestMilestoneRewards {
		if score < threshold {
			continue
		}
		// Skip if already claimed (deduplication guard)
		if uc.milestoneClaimsRepo != nil {
			already, err := uc.milestoneClaimsRepo.HasClaimed(playerID, threshold)
			if err != nil || already {
				continue
			}
		}
		if err := uc.inventoryService.Credit(playerID, "marathon_milestone_personal_best", map[string]int{"coins": coins}); err != nil {
			continue
		}
		if uc.milestoneClaimsRepo != nil {
			_ = uc.milestoneClaimsRepo.MarkClaimed(playerID, threshold)
		}
	}
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
