package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// ContinueMarathonUseCase handles continuing a marathon game after game over
type ContinueMarathonUseCase struct {
	marathonRepo solo_marathon.Repository
	questionRepo quiz.QuestionRepository
	eventBus     EventBus
}

// NewContinueMarathonUseCase creates a new ContinueMarathonUseCase
func NewContinueMarathonUseCase(
	marathonRepo solo_marathon.Repository,
	questionRepo quiz.QuestionRepository,
	eventBus EventBus,
) *ContinueMarathonUseCase {
	return &ContinueMarathonUseCase{
		marathonRepo: marathonRepo,
		questionRepo: questionRepo,
		eventBus:     eventBus,
	}
}

// Execute continues a marathon game after game over (player pays to resume)
func (uc *ContinueMarathonUseCase) Execute(input ContinueMarathonInput) (ContinueMarathonOutput, error) {
	// 1. Validate and convert input to domain types
	gameID := solo_marathon.NewGameIDFromString(input.GameID)
	if gameID.IsZero() {
		return ContinueMarathonOutput{}, solo_marathon.ErrInvalidGameID
	}

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return ContinueMarathonOutput{}, err
	}

	// Validate payment method
	var paymentMethod solo_marathon.PaymentMethod
	switch input.PaymentMethod {
	case "coins":
		paymentMethod = solo_marathon.PaymentCoins
	case "ad":
		paymentMethod = solo_marathon.PaymentAd
	default:
		return ContinueMarathonOutput{}, solo_marathon.ErrContinueNotAvailable
	}

	// 2. Load game aggregate
	game, err := uc.marathonRepo.FindByID(gameID)
	if err != nil {
		return ContinueMarathonOutput{}, err
	}

	// 3. Validate game belongs to player
	if !game.PlayerID().Equals(playerID) {
		return ContinueMarathonOutput{}, quiz.ErrUnauthorized
	}

	// 4. Calculate continue cost
	costCalc := solo_marathon.ContinueCostCalculator{}
	costCoins := 0
	if paymentMethod == solo_marathon.PaymentCoins {
		costCoins = costCalc.GetCost(game.ContinueCount())
		// TODO: Validate player has enough coins and deduct from balance
		// This requires UserRepository integration
	}

	// 5. Continue game (domain business logic)
	now := time.Now().Unix()
	if err := game.Continue(paymentMethod, costCoins, now); err != nil {
		return ContinueMarathonOutput{}, err
	}

	// 6. Load next question for the resumed game
	questionSelector := solo_marathon.NewQuestionSelector(uc.questionRepo)
	if err := game.LoadNextQuestion(questionSelector); err != nil {
		return ContinueMarathonOutput{}, err
	}

	// 7. Persist game
	if err := uc.marathonRepo.Save(game); err != nil {
		return ContinueMarathonOutput{}, err
	}

	// 8. Publish domain events
	if uc.eventBus != nil {
		events := game.Events()
		for _, event := range events {
			uc.eventBus.Publish(event)
		}
	}

	// 9. Calculate next continue cost
	nextContinueCost := costCalc.GetCost(game.ContinueCount())

	// 10. Build output
	return ContinueMarathonOutput{
		Game:             ToMarathonGameDTOV2(game, now),
		ContinueCount:   game.ContinueCount(),
		CoinsDeducted:   costCoins,
		NextContinueCost: nextContinueCost,
	}, nil
}
