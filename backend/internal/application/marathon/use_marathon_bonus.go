package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// UseMarathonBonusUseCase handles using a bonus in marathon mode
type UseMarathonBonusUseCase struct {
	marathonRepo solo_marathon.Repository
	questionRepo quiz.QuestionRepository
	eventBus     EventBus
}

// NewUseMarathonBonusUseCase creates a new UseMarathonBonusUseCase
func NewUseMarathonBonusUseCase(
	marathonRepo solo_marathon.Repository,
	questionRepo quiz.QuestionRepository,
	eventBus EventBus,
) *UseMarathonBonusUseCase {
	return &UseMarathonBonusUseCase{
		marathonRepo: marathonRepo,
		questionRepo: questionRepo,
		eventBus:     eventBus,
	}
}

// Execute uses a bonus in a marathon game
func (uc *UseMarathonBonusUseCase) Execute(input UseMarathonBonusInput) (UseMarathonBonusOutput, error) {
	// 1. Validate and convert input to domain types
	gameID := solo_marathon.NewGameIDFromString(input.GameID)
	if gameID.IsZero() {
		return UseMarathonBonusOutput{}, solo_marathon.ErrInvalidGameID
	}

	questionID, err := quiz.NewQuestionIDFromString(input.QuestionID)
	if err != nil {
		return UseMarathonBonusOutput{}, err
	}

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return UseMarathonBonusOutput{}, err
	}

	// Validate bonus type
	var bonusType solo_marathon.BonusType
	switch input.BonusType {
	case "shield":
		bonusType = solo_marathon.BonusShield
	case "fifty_fifty":
		bonusType = solo_marathon.BonusFiftyFifty
	case "skip":
		bonusType = solo_marathon.BonusSkip
	case "freeze":
		bonusType = solo_marathon.BonusFreeze
	default:
		return UseMarathonBonusOutput{}, solo_marathon.ErrInvalidBonusType
	}

	// 2. Load game aggregate
	game, err := uc.marathonRepo.FindByID(gameID)
	if err != nil {
		return UseMarathonBonusOutput{}, err
	}

	// 3. Validate game belongs to player
	if !game.PlayerID().Equals(playerID) {
		return UseMarathonBonusOutput{}, quiz.ErrUnauthorized
	}

	// 4. Get current question for bonus application
	currentQuestion, err := game.GetCurrentQuestion()
	if err != nil {
		return UseMarathonBonusOutput{}, err
	}

	// 5. Use bonus (domain business logic)
	now := time.Now().Unix()
	if err := game.UseBonus(questionID, bonusType, now); err != nil {
		return UseMarathonBonusOutput{}, err
	}

	// 6. Build bonus result based on type
	bonusResult := BonusResultDTO{}

	switch bonusType {
	case solo_marathon.BonusShield:
		// Shield activated — return active state
		active := true
		bonusResult.ShieldActive = &active

	case solo_marathon.BonusFiftyFifty:
		// Return IDs of 2 incorrect answers to hide
		hiddenAnswers := selectTwoIncorrectAnswers(currentQuestion)
		bonusResult.HiddenAnswerIDs = hiddenAnswers

	case solo_marathon.BonusFreeze:
		// Return new time limit (+10 seconds)
		questionIndex := game.QuestionNumber()
		currentTimeLimit := GetTimeLimit(game.Difficulty(), questionIndex)
		newTimeLimit := currentTimeLimit + 10
		bonusResult.NewTimeLimit = &newTimeLimit

	case solo_marathon.BonusSkip:
		// Skip moves to next question — load it
		questionSelector := solo_marathon.NewQuestionSelector(uc.questionRepo)
		if err := game.LoadNextQuestion(questionSelector); err != nil {
			return UseMarathonBonusOutput{}, err
		}

		nextQuestion, err := game.GetCurrentQuestion()
		if err == nil {
			nextQuestionDTO := ToQuestionDTO(nextQuestion)
			bonusResult.NextQuestion = &nextQuestionDTO

			nextTimeLimit := GetTimeLimit(game.Difficulty(), game.QuestionNumber())
			bonusResult.NextTimeLimit = &nextTimeLimit
		}
	}

	// 7. Get remaining count for this bonus type
	remainingCount := game.BonusInventory().Count(bonusType)

	// 8. Persist game
	if err := uc.marathonRepo.Save(game); err != nil {
		return UseMarathonBonusOutput{}, err
	}

	// 9. Publish domain events
	if uc.eventBus != nil {
		events := game.Events()
		for _, event := range events {
			uc.eventBus.Publish(event)
		}
	}

	// 10. Build output
	return UseMarathonBonusOutput{
		BonusType:      input.BonusType,
		RemainingCount: remainingCount,
		BonusInventory: ToBonusInventoryDTO(game.BonusInventory()),
		BonusResult:    bonusResult,
	}, nil
}

// selectTwoIncorrectAnswers selects 2 incorrect answer IDs to hide
func selectTwoIncorrectAnswers(question *quiz.Question) []string {
	incorrectAnswers := make([]string, 0, 3) // Max 3 incorrect answers

	for _, answer := range question.Answers() {
		if !answer.IsCorrect() {
			incorrectAnswers = append(incorrectAnswers, answer.ID().String())
		}
	}

	// Return first 2 incorrect answers
	if len(incorrectAnswers) >= 2 {
		return incorrectAnswers[:2]
	}

	return incorrectAnswers
}
