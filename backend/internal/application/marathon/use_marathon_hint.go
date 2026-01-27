package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// UseMarathonHintUseCase handles using a hint in marathon mode
type UseMarathonHintUseCase struct {
	marathonRepo solo_marathon.Repository
	eventBus     EventBus
}

// NewUseMarathonHintUseCase creates a new UseMarathonHintUseCase
func NewUseMarathonHintUseCase(
	marathonRepo solo_marathon.Repository,
	eventBus EventBus,
) *UseMarathonHintUseCase {
	return &UseMarathonHintUseCase{
		marathonRepo: marathonRepo,
		eventBus:     eventBus,
	}
}

// Execute uses a hint in a marathon game
func (uc *UseMarathonHintUseCase) Execute(input UseMarathonHintInput) (UseMarathonHintOutput, error) {
	// 1. Validate and convert input to domain types
	gameID := solo_marathon.NewGameIDFromString(input.GameID)
	if gameID.IsZero() {
		return UseMarathonHintOutput{}, solo_marathon.ErrInvalidGameID
	}

	questionID, err := quiz.NewQuestionIDFromString(input.QuestionID)
	if err != nil {
		return UseMarathonHintOutput{}, err
	}

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return UseMarathonHintOutput{}, err
	}

	// Validate hint type
	var hintType solo_marathon.HintType
	switch input.HintType {
	case "fifty_fifty":
		hintType = solo_marathon.HintFiftyFifty
	case "extra_time":
		hintType = solo_marathon.HintExtraTime
	case "skip":
		hintType = solo_marathon.HintSkip
	default:
		return UseMarathonHintOutput{}, solo_marathon.ErrInvalidHintType
	}

	// 2. Load game aggregate
	game, err := uc.marathonRepo.FindByID(gameID)
	if err != nil {
		return UseMarathonHintOutput{}, err
	}

	// 3. Validate game belongs to player
	if !game.PlayerID().Equals(playerID) {
		return UseMarathonHintOutput{}, quiz.ErrUnauthorized
	}

	// 4. Get current question for hint application
	currentQuestion, err := game.GetCurrentQuestion()
	if err != nil {
		return UseMarathonHintOutput{}, err
	}

	// 5. Use hint (domain business logic)
	now := time.Now().Unix()
	if err := game.UseHint(questionID, hintType, now); err != nil {
		return UseMarathonHintOutput{}, err
	}

	// 6. Build hint result based on type
	hintResult := HintResultDTO{}

	switch hintType {
	case solo_marathon.HintFiftyFifty:
		// Return IDs of 2 incorrect answers to hide
		hiddenAnswers := selectTwoIncorrectAnswers(currentQuestion)
		hintResult.HiddenAnswerIDs = hiddenAnswers

	case solo_marathon.HintExtraTime:
		// Return new time limit (+10 seconds)
		currentTimeLimit := GetTimeLimit(game.Difficulty(), game.CurrentStreak())
		newTimeLimit := currentTimeLimit + 10
		hintResult.NewTimeLimit = &newTimeLimit

	case solo_marathon.HintSkip:
		// Skip current question and get next one
		// Note: In domain, skip should move to next question
		// For now, we'll just return the next question
		// TODO: Implement skip logic in domain layer
		nextQuestion, err := game.GetCurrentQuestion()
		if err == nil {
			nextQuestionDTO := ToQuestionDTO(nextQuestion)
			hintResult.NextQuestion = &nextQuestionDTO

			nextTimeLimit := GetTimeLimit(game.Difficulty(), game.CurrentStreak())
			hintResult.NextTimeLimit = &nextTimeLimit
		}
	}

	// 7. Get remaining hints count
	var remainingHints int
	switch hintType {
	case solo_marathon.HintFiftyFifty:
		remainingHints = game.Hints().FiftyFifty()
	case solo_marathon.HintExtraTime:
		remainingHints = game.Hints().ExtraTime()
	case solo_marathon.HintSkip:
		remainingHints = game.Hints().Skip()
	}

	// 8. Persist game
	if err := uc.marathonRepo.Save(game); err != nil {
		return UseMarathonHintOutput{}, err
	}

	// 9. Publish domain events
	if uc.eventBus != nil {
		events := game.Events()
		for _, event := range events {
			uc.eventBus.Publish(event)
		}
	}

	// 10. Build output
	return UseMarathonHintOutput{
		HintType:       input.HintType,
		RemainingHints: remainingHints,
		HintResult:     hintResult,
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
	// In production, this could be randomized
	if len(incorrectAnswers) >= 2 {
		return incorrectAnswers[:2]
	}

	return incorrectAnswers
}
