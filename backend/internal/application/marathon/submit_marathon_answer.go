package marathon

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/shared"
	"github.com/barsukov/quiz-sprint/backend/internal/domain/solo_marathon"
)

// SubmitMarathonAnswerUseCase handles submitting an answer in marathon mode
type SubmitMarathonAnswerUseCase struct {
	marathonRepo     solo_marathon.Repository
	personalBestRepo solo_marathon.PersonalBestRepository
	questionRepo     quiz.QuestionRepository
	eventBus         EventBus
}

// NewSubmitMarathonAnswerUseCase creates a new SubmitMarathonAnswerUseCase
func NewSubmitMarathonAnswerUseCase(
	marathonRepo solo_marathon.Repository,
	personalBestRepo solo_marathon.PersonalBestRepository,
	questionRepo quiz.QuestionRepository,
	eventBus EventBus,
) *SubmitMarathonAnswerUseCase {
	return &SubmitMarathonAnswerUseCase{
		marathonRepo:     marathonRepo,
		personalBestRepo: personalBestRepo,
		questionRepo:     questionRepo,
		eventBus:         eventBus,
	}
}

// Execute submits an answer for a marathon game question
func (uc *SubmitMarathonAnswerUseCase) Execute(input SubmitMarathonAnswerInput) (SubmitMarathonAnswerOutput, error) {
	// 1. Validate and convert input to domain types
	gameID := solo_marathon.NewGameIDFromString(input.GameID)
	if gameID.IsZero() {
		return SubmitMarathonAnswerOutput{}, solo_marathon.ErrInvalidGameID
	}

	questionID, err := quiz.NewQuestionIDFromString(input.QuestionID)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	answerID, err := quiz.NewAnswerIDFromString(input.AnswerID)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	playerID, err := shared.NewUserID(input.PlayerID)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	// Validate timeTaken
	if input.TimeTaken < 0 {
		return SubmitMarathonAnswerOutput{}, quiz.ErrInvalidTimeTaken
	}
	if input.TimeTaken > 3600000 { // Max 1 hour in milliseconds
		return SubmitMarathonAnswerOutput{}, quiz.ErrTimeTakenTooLong
	}

	// 2. Load game aggregate
	game, err := uc.marathonRepo.FindByID(gameID)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	// 3. Validate game belongs to player
	if !game.PlayerID().Equals(playerID) {
		return SubmitMarathonAnswerOutput{}, quiz.ErrUnauthorized
	}

	// 4. Get the current question for finding correct answer later
	currentQuestion, err := game.GetCurrentQuestion()
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	// 5. Submit answer (domain business logic)
	now := time.Now().Unix()
	result, err := game.AnswerQuestion(questionID, answerID, input.TimeTaken, now)
	if err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	// 5b. If game continues, load next question
	if !result.IsGameOver {
		questionSelector := solo_marathon.NewQuestionSelector(uc.questionRepo)
		if err := game.LoadNextQuestion(questionSelector); err != nil {
			// Log error but don't fail - game can continue
			// TODO: Add proper logging
			_ = err
		}
	}

	// 6. Build output
	output := SubmitMarathonAnswerOutput{
		IsCorrect:       result.IsCorrect,
		CorrectAnswerID: FindCorrectAnswerID(currentQuestion),
		BasePoints:      result.BasePoints,
		TimeTaken:       result.TimeTaken,
		CurrentStreak:   result.CurrentStreak,
		MaxStreak:       result.MaxStreak,
		DifficultyLevel: string(result.DifficultyLevel),
		LifeLost:        result.LifeLost,
		RemainingLives:  result.RemainingLives,
		IsGameOver:      result.IsGameOver,
	}

	// 7. Handle game over scenario
	if result.IsGameOver {
		// Update PersonalBest if new record
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

		// TODO: Get global rank from leaderboard repository
		var globalRank *int = nil

		output.GameOverResult = &GameOverResultDTO{
			FinalStreak:       game.MaxStreak(),
			IsNewPersonalBest: game.IsNewPersonalBest(),
			PreviousRecord:    game.PersonalBestStreak(),
			TotalBaseScore:    game.BaseScore(),
			GlobalRank:        globalRank,
		}
	} else {
		// Game continues - get next question
		nextQuestion, err := game.GetCurrentQuestion()
		if err == nil {
			nextQuestionDTO := ToQuestionDTO(nextQuestion)
			output.NextQuestion = &nextQuestionDTO

			// Calculate time limit for next question
			nextTimeLimit := GetTimeLimit(game.Difficulty(), game.CurrentStreak())
			output.NextTimeLimit = &nextTimeLimit
		}
	}

	// 8. Persist game
	if err := uc.marathonRepo.Save(game); err != nil {
		return SubmitMarathonAnswerOutput{}, err
	}

	// 9. Publish domain events
	if uc.eventBus != nil {
		events := game.Events()
		for _, event := range events {
			uc.eventBus.Publish(event)
		}
	}

	return output, nil
}
